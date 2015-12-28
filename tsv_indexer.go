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

// UnescapeSeparator cleans coposed separator like \t
func UnescapeSeparator(separator string) string {
	return strings.Replace(separator, "\\t", string([]rune{9}), -1) // String with '\t' rune
}

// TrimNewline removes newline characters from the end of line
func TrimNewline(line string) string {
	line = strings.TrimRight(line, "\n")
	line = strings.TrimRight(line, "\r")
	return strings.TrimRight(line, "\n")
}

// TsvLine describes the content of a line from a TSV
type TsvLine struct {
	Index       int
	Comparables []string
}
type tsvLines []TsvLine

// TsvIndexer contains all stuff for indexing columns from a TSV
type TsvIndexer struct {
	I           *Indexer
	Header      bool
	Separator   string
	Fields      []string
	FieldsIndex map[string]int
	Lines       tsvLines
}

// NewTsvIndexer instanciates a new TsvIndexer
func NewTsvIndexer(sc *Scanner, header bool, separator string, fields []string) *TsvIndexer {
	sc.Reset()
	sc.KeepNewlineSequence(true)
	return &TsvIndexer{
		I:           NewIndexer(sc),
		Header:      header,
		Separator:   UnescapeSeparator(separator),
		Fields:      fields,
		FieldsIndex: make(map[string]int),
	}
}

// Analyze parses the TSV and generates the indexes
func (ti *TsvIndexer) Analyze() error {
	if !ti.Header {
		re, _ := regexp.Compile(`var(\d+)`)
		for _, field := range ti.Fields {
			if len(re.FindStringSubmatch(field)) < 2 {
				return errors.New("Field " + field + " do not match with pattern /var\\d+/")
			}
		}
	}
	if err := ti.I.Analyze(ti.tsvLineAppender); err != nil {
		return err
	}
	return nil
}

// Sort sorts TsvLine on its comparables
func (ti *TsvIndexer) Sort() {
	sort.Sort(ti.Lines)
}

// Transfer writes sorted TSV into a new file
func (ti *TsvIndexer) Transfer(output io.Writer) error {
	w := bufio.NewWriter(output)

	// For all sorted line contained ine the TSV
	for _, tsvLine := range ti.Lines {
		line := ti.I.Lines[tsvLine.Index]                     // Retreives current line index from input file
		token, err := ti.I.sc.readAt(line.Offset, line.Limit) // Reads the current line
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
// Analyze stuff      //
// ------------------ //

// TODO Handle invalid separator (or not)
func (ti *TsvIndexer) tsvLineAppender(line []byte, index int) {
	str := TrimNewline(string(line))
	row := strings.Split(str, ti.Separator)
	if index == 0 && ti.Header {
		ti.findFieldsIndex(row)
		// Build empty comparable
		// When comparables are sorted, this one (the header) remains the first line
		comparables := []string{}
		for i := 0; i < len(ti.Fields); i++ {
			comparables = append(comparables, "")
		}
		ti.Lines = append(ti.Lines, TsvLine{index, comparables})
	} else if index == 0 {
		// Without header, fields are named like the following pattern /var\d+/
		// \d+ is used for the index of the variable
		//
		// e.g. `var1,var2,var3` with `var1` had the index 0
		re, _ := regexp.Compile(`var(\d+)`)
		for _, field := range ti.Fields {
			i, err := strconv.Atoi(re.FindStringSubmatch(field)[1])
			if err != nil {
				panic(err)
			}
			ti.FieldsIndex[field] = i - 1
			ti.appendComparable(row[i-1], index) // The first row contains data (/!\ it is not an header)
		}
	} else {
		for _, field := range ti.Fields {
			i := ti.FieldsIndex[field]
			ti.appendComparable(row[i], index)
		}
	}
}

func (ti *TsvIndexer) appendComparable(comparable string, index int) {
	if index > len(ti.Lines)-1 {
		ti.Lines = append(ti.Lines, TsvLine{index, []string{comparable}})
	} else {
		ti.Lines[index].Comparables = append(ti.Lines[index].Comparables, comparable)
	}
}

// Append to TsvIndexer.FieldsIndex the index in the row of all TsvIndexer.Fields
func (ti *TsvIndexer) findFieldsIndex(row []string) {
	for i, head := range row {
		for _, field := range ti.Fields {
			if head == field {
				ti.FieldsIndex[field] = i
				break
			}
		}
	}
}
