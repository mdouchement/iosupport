package iosupport

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// COMPARABLE_SEPARATOR defines the separator added between each indexed fields.
const COMPARABLE_SEPARATOR = "\u0000"

// TsvLine describes the line's details from a TSV.
type (
	TsvLine struct {
		Comparable string
		Offset     uint64
		Limit      uint32
	}
	TsvLines []TsvLine

	// Internal use
	seeker struct {
		*Scanner
		offset uint64
	}

	// TsvIndexer contains all stuff for indexing and sorting columns from a TSV.
	TsvIndexer struct {
		*Options
		parser          *TsvParser
		FieldsIndex     map[string]int
		Lines           TsvLines
		nbOfFields      int
		seekers         []seeker
		scannerFunc     func() *Scanner
		blankComparable string
	}
)

// NewTsvIndexer instanciates a new TsvIndexer.
func NewTsvIndexer(scannerFunc func() *Scanner, setters ...Option) *TsvIndexer {
	sc := scannerFunc()
	sc.Reset()
	sc.KeepNewlineSequence(true)

	options := &Options{
		Separator:     ',',
		LineThreshold: 2500000,
		Swapper:       NewNullSwapper(),
	}
	for _, setter := range setters {
		setter(options)
	}

	return &TsvIndexer{
		parser:          NewTsvParser(sc, options.Separator),
		Options:         options,
		FieldsIndex:     make(map[string]int),
		scannerFunc:     scannerFunc,
		nbOfFields:      -1,
		seekers:         []seeker{{sc, 0}},
		blankComparable: strings.Repeat(COMPARABLE_SEPARATOR, len(options.Fields)),
	}
}

// CloseIO closes all opened IO.
func (ti *TsvIndexer) CloseIO() {
	ti.parser.Scanner.f.Close()
	ti.releaseSeekers()
}

// Analyze parses the TSV and generates the indexes.
func (ti *TsvIndexer) Analyze() error {
	if !ti.Header {
		// Validate provided Fields with the generated header.
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
		err := ti.tsvLineAppender(ti.parser.Row(), len(ti.Lines), ti.parser.Line(), ti.parser.Offset(), ti.parser.Limit())
		if err != nil {
			return err
		}

		ti.tryToSwap(false)
	}
	ti.tryToSwap(true)
	ti.parser.Reset()
	ti.createSeekers()
	return nil
}

// Sort sorts TsvLine on its comparables.
func (ti *TsvIndexer) Sort() {
	sort.Sort(ti.Lines)
}

// Transfer writes sorted TSV into a new file.
func (ti *TsvIndexer) Transfer(output FileWriter) error {
	w := bufio.NewWriter(output)
	ns := ti.parser.NewlineSequence()
	n := ns[len(ns)-1]

	// For all sorted lines contained in the TSV
	it := ti.Swapper.ReadIterator()
	if it.Error() != nil {
		return it.Error()
	}

	for it.Next() {
		if it.Error() != nil {
			return it.Error()
		}
		line := it.Value()

		token, err := ti.selectSeeker(line.Offset).ReadAt(int64(line.Offset), int(line.Limit)) // Reads the current line
		if err != nil {
			return err
		}

		if token[len(token)-1] != n {
			token = append(token, ns...) // Appends newline sequence when missing
		}

		if _, err := w.Write(token); err != nil { // writes the current line into the sorted TSV output
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	ti.releaseSeekers()
	ti.Swapper.EraseAll()
	return nil
}

// ------------------ //
// Sort stuff         //
// ------------------ //

func (slice TsvLines) Len() int {
	return len(slice)
}

func (slice TsvLines) Less(i, j int) bool {
	return CompareFunc(slice[i], slice[j])
}

func (slice TsvLines) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

var CompareFunc = func(i, j TsvLine) bool {
	return i.Comparable < j.Comparable
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
	ti.seekers = append(ti.seekers, seeker{Scanner: sc, offset: offset})
}

// releaseSeekers closes internal opened file
func (ti *TsvIndexer) releaseSeekers() {
	for _, seeker := range ti.seekers {
		seeker.f.Close()
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

func (ti *TsvIndexer) tsvLineAppender(row [][]byte, index int, fileline int, offset uint64, limit uint32) error {
	if !ti.isValidRow(row) {
		// Discard mal-formatted lines
		return nil
	}

	ti.Lines = append(ti.Lines, TsvLine{"", offset, limit})
	if fileline == 0 && ti.Header {
		if err := ti.findFieldsIndex(row); err != nil {
			return err
		}
		// Build empty comparable
		// When comparables are sorted, this one (the header) remains the first line
		for i := 0; i < len(ti.Fields); i++ {
			ti.Lines[index].Comparable = ""
		}
		ti.nbOfFields = len(row)
	} else if fileline == 0 {
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
		ti.dropLastLineIfEmptyComparable()
		ti.nbOfFields = len(row)
	} else {
		for _, field := range ti.Fields {
			ti.appendComparable(row[ti.FieldsIndex[field]], index)
		}
		ti.dropLastLineIfEmptyComparable()
	}
	return nil
}

// It concats the given comparable to the existing comparable
func (ti *TsvIndexer) appendComparable(comparable []byte, index int) {
	cp := make([]byte, len(comparable), len(comparable))
	copy(cp, comparable) // Freeing the underlying array (https://blog.golang.org/go-slices-usage-and-internals - chapter: A possible "gotcha")
	ti.Lines[index].Comparable += fmt.Sprintf("%s%s", cp, COMPARABLE_SEPARATOR)
}

func (ti *TsvIndexer) dropLastLineIfEmptyComparable() {
	if !ti.DropEmptyIndexedFields {
		return
	}

	i := len(ti.Lines) - 1
	if ti.Lines[i].Comparable == ti.blankComparable {
		ti.Lines = ti.Lines[:i]
	}
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

func (ti *TsvIndexer) isValidRow(row [][]byte) bool {
	return !ti.SkipMalformattedLines || ti.nbOfFields == len(row) || ti.nbOfFields == -1
}

func (ti *TsvIndexer) tryToSwap(force bool) error {
	if force && !ti.Swapper.HasSwapped() {
		ti.Swapper.KeepWithoutSwap(ti.Lines)
		return nil
	}

	if force || ti.Swapper.IsTimeToSwap(ti.Lines) {
		ti.Sort()
		if err := ti.Swapper.Swap(ti.Lines); err != nil {
			return err
		}
		if force {
			ti.Lines = nil // Freeing
		} else {
			ti.Lines = ti.Lines[:0] // reuse current allocated array
		}
	}
	return nil
}
