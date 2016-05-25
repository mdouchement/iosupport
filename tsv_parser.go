package iosupport

import (
	"bytes"
)

// UnescapeSeparator cleans composed separator like \t
func UnescapeSeparator(separator string) byte {
	return []byte(separator)[0]
}

// TrimNewline removes newline characters at the end of line
func TrimNewline(line []byte) []byte {
	line = bytes.TrimRight(line, "\n")
	line = bytes.TrimRight(line, "\r")
	return bytes.TrimRight(line, "\n")
}

// A TsvParser reads records from a TSV-encoded file.
type TsvParser struct {
	*Scanner
	err       error // Sticky error.
	Separator []byte
	QuoteChar byte
	row       [][]byte
}

// NewTsvParser inatanciates a new TsvParser
func NewTsvParser(sc *Scanner, separator byte) *TsvParser {
	sc.Reset()
	return &TsvParser{
		Scanner:   sc,
		Separator: []byte{separator},
		QuoteChar: '"',
	}
}

// Err returns the first non-EOF error that was encountered by the Scanner.
func (tp *TsvParser) Err() error {
	return tp.err
}

// Row returns a slice of []byte with each []byte representing one field
func (tp *TsvParser) Row() [][]byte {
	return tp.row
}

// Reset resets parser and its underliying scanner. It freeing the memory
func (tp *TsvParser) Reset() {
	tp.Scanner.Reset()
	tp.row = make([][]byte, 0)
}

// ScanRow advances the TSV parser to the next row
func (tp *TsvParser) ScanRow() bool {
	b := tp.ScanLine()
	if tp.Scanner.Err() != nil {
		tp.err = tp.Scanner.Err()
		b = !tp.IsLineEmpty()
	}

	if b {
		tp.row = tp.parseFields()
	}

	return b
}

// Simple fields parser
func (tp *TsvParser) parseFields() [][]byte {
	return bytes.Split(TrimNewline(tp.Bytes()), tp.Separator)
}
