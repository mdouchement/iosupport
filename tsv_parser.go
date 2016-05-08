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
	Separator []byte
	QuoteChar byte
	line      int   // move in Scanner
	offset    int64 // move in Scanner
	limit     int   // move in Scanner
	row       [][]byte
	err       error
}

// NewTsvParser inatanciates a new TsvParser
func NewTsvParser(sc *Scanner, separator byte) *TsvParser {
	sc.Reset()
	// sc.KeepNewlineSequence(true)
	return &TsvParser{
		sc:        sc,
		Separator: []byte{separator},
		QuoteChar: '"',
		line:      -1,
	}
}

// Err returns the first non-EOF error that was encountered by the Scanner.
func (tp *TsvParser) Err() error {
	return tp.err
}

// Line return the index of the current line (unparsed row)
func (tp *TsvParser) Line() int {
	return tp.line
}

// Offset return the byte offset of the current line (unparsed row)
func (tp *TsvParser) Offset() int64 {
	return tp.offset
}

// Limit return the byte length of the current line (unparsed row)
func (tp *TsvParser) Limit() int {
	return tp.limit
}

// Row returns a slice of []byte with each []byte representing one field
func (tp *TsvParser) Row() [][]byte {
	return tp.row
}

// Reset resets parser and its underliying scanner. It freeing the memory
func (tp *TsvParser) Reset() {
	tp.sc.Reset()
	tp.row = make([][]byte, 0)
	tp.line = 0
	tp.limit = 0
	tp.offset = 0
}

// ScanRow advances the TSV parser to the next row
func (tp *TsvParser) ScanRow() bool {
	tp.offset += int64(tp.limit)

	b := tp.sc.ScanLine()
	if tp.sc.Err() != nil {
		if tp.sc.Err() != nil {
			tp.err = tp.sc.Err()
			b = !tp.sc.IsLineEmpty()
		}
	}

	if b {
		tp.limit = len(tp.sc.Bytes())
		tp.line++

		tp.row = tp.parseFields()
	}

	return b
}

// Simple fields parser
func (tp *TsvParser) parseFields() [][]byte {
	return bytes.Split(TrimNewline(tp.sc.Bytes()), tp.Separator)
}
