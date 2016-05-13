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
	sc        *Scanner // directly embedded in TsvParser?
	err       error    // Sticky error.
	Separator []byte
	QuoteChar byte
	row       [][]byte
}

// NewTsvParser inatanciates a new TsvParser
func NewTsvParser(sc *Scanner, separator byte) *TsvParser {
	sc.Reset()
	// sc.KeepNewlineSequence(true)
	return &TsvParser{
		sc:        sc,
		Separator: []byte{separator},
		QuoteChar: '"',
	}
}

// Err returns the first non-EOF error that was encountered by the Scanner.
func (tp *TsvParser) Err() error {
	return tp.err
}

// Line return the index of the current line (unparsed row)
func (tp *TsvParser) Line() int {
	return tp.sc.Line()
}

// Offset return the byte offset of the current line (unparsed row)
func (tp *TsvParser) Offset() uint64 {
	return tp.sc.Offset()
}

// Limit return the byte length of the current line (unparsed row)
func (tp *TsvParser) Limit() uint32 {
	return tp.sc.Limit()
}

// Row returns a slice of []byte with each []byte representing one field
func (tp *TsvParser) Row() [][]byte {
	return tp.row
}

// Reset resets parser and its underliying scanner. It freeing the memory
func (tp *TsvParser) Reset() {
	tp.sc.Reset()
	tp.row = make([][]byte, 0)
}

// ScanRow advances the TSV parser to the next row
func (tp *TsvParser) ScanRow() bool {
	b := tp.sc.ScanLine()
	if tp.sc.Err() != nil {
		if tp.sc.Err() != nil {
			tp.err = tp.sc.Err()
			b = !tp.sc.IsLineEmpty()
		}
	}

	if b {
		tp.row = tp.parseFields()
	}

	return b
}

// Simple fields parser
func (tp *TsvParser) parseFields() [][]byte {
	return bytes.Split(TrimNewline(tp.sc.Bytes()), tp.Separator)
}
