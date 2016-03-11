package iosupport

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// UnescapeSeparator cleans composed separator like \t
func UnescapeSeparator(separator string) string {
	return strings.Replace(separator, "\\t", string([]rune{9}), -1) // String with '\t' rune
}

// TrimNewline removes newline characters at the end of line
func TrimNewline(line string) string {
	line = strings.TrimRight(line, "\n")
	line = strings.TrimRight(line, "\r")
	return strings.TrimRight(line, "\n")
}

// TsvLine describes the line's details from a TSV
type TsvLine struct {
	Index       int
	Comparables []string
	Offset      int64
	Limit       int
}
type tsvLines []TsvLine

// Internal use
type seeker struct {
	sc     *Scanner
	offset int64
}

// TsvIndexer contains all stuff for indexing columns from a TSV
type TsvIndexer struct {
	I             *Indexer
	Header        bool
	Separator     string
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
		I:             NewIndexer(sc),
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
	ti.I.sc.f.Close()
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
	if err := ti.I.Analyze(ti.tsvLineAppender); err != nil {
		return err
	}
	ti.createSeekers(len(ti.Lines), func(index int) int64 {
		return ti.Lines[index].Offset
	}) // For transfer part
	return nil
}

// Sort sorts TsvLine on its comparables
func (ti *TsvIndexer) Sort() {
	sort.Sort(ti.Lines)
}

// Transfer writes sorted TSV into a new file
func (ti *TsvIndexer) Transfer(output io.Writer) error {
	w := bufio.NewWriter(output)

	// For all sorted lines contained in the TSV
	for _, line := range ti.Lines {
		token, err := ti.selectSeeker(line.Offset).sc.ReadAt(line.Offset, line.Limit) // Reads the current line
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
func (ti *TsvIndexer) createSeekers(nol int, offset func(int) int64) {
	if nol < ti.LineThreshold {
		return
	}

	nbOfThresholds := nol / ti.LineThreshold
	lineOffset := nol / (nbOfThresholds + 1)
	lineIndex := 0
	for i := 0; i < nbOfThresholds; i++ {
		lineIndex += lineOffset
		ti.appendSeeker(offset(lineIndex))
	}
}

// appendSeeker appends a new seeker based on the given offset. Seekers must be appened ordering by the offset
func (ti *TsvIndexer) appendSeeker(offset int64) {
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
func (ti *TsvIndexer) selectSeeker(offset int64) seeker {
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

type appenderBuffer struct {
	str        string
	row        []string
	fieldIndex int
}

var buf appenderBuffer

func (ti *TsvIndexer) tsvLineAppender(line []byte, index int, offset int64, limit int) error {
	buf.str = TrimNewline(string(line))
	buf.row = strings.Split(buf.str, ti.Separator)
	ti.Lines = append(ti.Lines, TsvLine{index, []string{}, offset, limit})
	if index == 0 && ti.Header {
		if err := ti.findFieldsIndex(buf.row); err != nil {
			return err
		}
		// Build empty comparable
		// When comparables are sorted, this one (the header) remains the first line
		for i := 0; i < len(ti.Fields); i++ {
			ti.appendComparable("", index)
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
			ti.appendComparable(buf.row[i-1], index) // The first row contains data (/!\ it is not an header)
		}
	} else {
		for _, field := range ti.Fields {
			buf.fieldIndex = ti.FieldsIndex[field]
			ti.appendComparable(buf.row[buf.fieldIndex], index)
		}
	}
	buf.str = ""
	buf.row = nil
	return nil
}

func (ti *TsvIndexer) appendComparable(comparable string, index int) {
	cp := make([]byte, len(comparable))
	copy(cp, comparable) // Freeing the underlying array (https://blog.golang.org/go-slices-usage-and-internals - chapter: A possible "gotcha")
	ti.Lines[index].Comparables = append(ti.Lines[index].Comparables, string(cp))
}

// Append to TsvIndexer.FieldsIndex the index in the row of all TsvIndexer.Fields
func (ti *TsvIndexer) findFieldsIndex(row []string) error {
	for i, head := range row {
		for _, field := range ti.Fields {
			if head == field {
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
