package iosupport

import (
	"bufio"
	"errors"
	"regexp"
	"sort"
	"strconv"
)

// TsvLine describes the line's details from a TSV
type TsvLine struct {
	Comparables []string
	Offset      uint64
	Limit       uint32
}
type tsvLines []TsvLine

// Internal use
type seeker struct {
	sc     *Scanner
	offset uint64
}

// TsvIndexer contains all stuff for indexing columns from a TSV
type TsvIndexer struct {
	parser        *TsvParser // directly embedded in TsvParser?
	Header        bool
	Separator     byte
	Fields        []string
	FieldsIndex   map[string]int
	Lines         tsvLines
	seekers       []seeker
	scannerFunc   func() *Scanner
	LineThreshold int
}

// NewTsvIndexer instanciates a new TsvIndexer
func NewTsvIndexer(scannerFunc func() *Scanner, header bool, separator string, fields []string) *TsvIndexer {
	sc := scannerFunc()
	sc.Reset()
	sc.KeepNewlineSequence(true)
	return &TsvIndexer{
		parser:        NewTsvParser(sc, UnescapeSeparator(separator)),
		Header:        header,
		Separator:     UnescapeSeparator(separator),
		Fields:        fields,
		FieldsIndex:   make(map[string]int),
		scannerFunc:   scannerFunc,
		LineThreshold: 2500000,
		seekers:       []seeker{seeker{sc, 0}},
	}
}

// CloseIO closes all opened IO
func (ti *TsvIndexer) CloseIO() {
	ti.parser.sc.f.Close()
	ti.releaseSeekers()
}

// Analyze parses the TSV and generates the indexes
func (ti *TsvIndexer) Analyze() error {
	if !ti.Header {
		re := regexp.MustCompile(`var(\d+)`)
		for _, field := range ti.Fields {
			if len(re.FindStringSubmatch(field)) < 2 {
				return errors.New("Field " + field + " do not match with pattern /var\\d+/")
			}
		}
	}
	for ti.parser.ScanRow() {
		if ti.parser.Err() != nil {
			return ti.parser.Err()
		}
		ti.tsvLineAppender(ti.parser.Row(), ti.parser.Line(), ti.parser.Offset(), ti.parser.Limit())
	}
	ti.parser.Reset()
	ti.createSeekers()
	return nil
}

// Sort sorts TsvLine on its comparables
func (ti *TsvIndexer) Sort() {
	sort.Sort(ti.Lines)
}

// Transfer writes sorted TSV into a new file
func (ti *TsvIndexer) Transfer(output FileWriter) error {
	w := bufio.NewWriter(output)

	// For all sorted lines contained in the TSV
	for _, line := range ti.Lines {
		token, err := ti.selectSeeker(line.Offset).sc.ReadAt(int64(line.Offset), int(line.Limit)) // Reads the current line
		if err != nil {
			return err
		}
		if _, err := w.Write(token); err != nil { // writes the current line into the sorted TSV output
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	ti.releaseSeekers()
	return nil
}

// ------------------ //
// Sort stuff         //
// ------------------ //

func (slice tsvLines) Len() int {
	return len(slice)
}

func (slice tsvLines) Less(i, j int) bool {
	b := false
	for x, comparable := range slice[i].Comparables {
		b = b || comparable < slice[j].Comparables[x]
	}
	return b
}

func (slice tsvLines) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// ------------------ //
// Transfer stuff     //
// ------------------ //

// createSeekers creates seekers for increase the speed of random accesses of the read file
func (ti *TsvIndexer) createSeekers() {
	nol := len(ti.Lines)
	if nol < ti.LineThreshold {
		return
	}

	nbOfThresholds := nol / ti.LineThreshold
	lineOffset := nol / (nbOfThresholds + 1)
	lineIndex := 0
	for i := 0; i < nbOfThresholds; i++ {
		lineIndex += lineOffset
		ti.appendSeeker(ti.Lines[lineIndex].Offset)
	}
}

// appendSeeker appends a new seeker based on the given offset. Seekers must be appened ordering by the offset
func (ti *TsvIndexer) appendSeeker(offset uint64) {
	sc := ti.scannerFunc()
	ti.seekers = append(ti.seekers, seeker{sc, offset})
}

// releaseSeekers closes internal opened file
func (ti *TsvIndexer) releaseSeekers() {
	for _, seeker := range ti.seekers {
		seeker.sc.f.Close()
	}
}

// selectSeeker returns the nearest inferior seeker
// e.g. A file with 10,000,000 lines
//    s0 -> offset 0
//    s1 -> offset 2,500,000
//    s2 -> offset 5,000,000
//    s3 -> offset 7,500,000
// offset 666 returns seeker s0
// offset 9,999,999 returns seeker s3
func (ti *TsvIndexer) selectSeeker(offset uint64) seeker {
	for i, seeker := range ti.seekers {
		if seeker.offset > offset {
			return ti.seekers[i-1]
		}
	}
	return ti.seekers[len(ti.seekers)-1]
}

// ------------------ //
// Analyze stuff      //
// ------------------ //

func (ti *TsvIndexer) tsvLineAppender(row [][]byte, index int, offset uint64, limit uint32) error {
	ti.Lines = append(ti.Lines, TsvLine{make([]string, len(ti.Fields), len(ti.Fields)), offset, limit})
	if index == 0 && ti.Header {
		if err := ti.findFieldsIndex(row); err != nil {
			return err
		}
		// Build empty comparable
		// When comparables are sorted, this one (the header) remains the first line
		for i := 0; i < len(ti.Fields); i++ {
			ti.appendComparable([]byte{}, index)
		}
	} else if index == 0 {
		// Without header, fields are named like the following pattern /var\d+/
		// \d+ is used for the index of the variable
		//
		// e.g. `var1,var2,var3` with `var1` had the index 0
		re := regexp.MustCompile(`var(\d+)`)
		for _, field := range ti.Fields {
			i, err := strconv.Atoi(re.FindStringSubmatch(field)[1])
			if err != nil {
				return err
			}
			ti.FieldsIndex[field] = i - 1
			ti.appendComparable(row[i-1], index) // The first row contains data (/!\ it is not an header)
		}
	} else {
		for _, field := range ti.Fields {
			ti.appendComparable(row[ti.FieldsIndex[field]], index)
		}
	}
	return nil
}

func (ti *TsvIndexer) appendComparable(comparable []byte, index int) {
	cp := make([]byte, len(comparable), len(comparable))
	copy(cp, comparable) // Freeing the underlying array (https://blog.golang.org/go-slices-usage-and-internals - chapter: A possible "gotcha")
	ti.Lines[index].Comparables = append(ti.Lines[index].Comparables, string(cp))
}

// Append to TsvIndexer.FieldsIndex the index in the row of all TsvIndexer.Fields
func (ti *TsvIndexer) findFieldsIndex(row [][]byte) error {
	for i, head := range row {
		for _, field := range ti.Fields {
			if string(head) == field {
				ti.FieldsIndex[field] = i
				break
			}
		}
	}
	if len(ti.Fields) != len(ti.FieldsIndex) {
		return errors.New("Invalid separator or sort fields")
	}
	return nil
}
