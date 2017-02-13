package iosupport

import (
	"bytes"
	"errors"
	"fmt"
)

// A ParseError is returned for parsing errors.
// The first line is 1.  The first column is 0.
type ParseError struct {
	Line   int   // Line where the error occurred
	Column int   // Column (rune index) where the error occurred
	Err    error // The actual error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d, character %d: %s", e.Line, e.Column, e.Err)
}

var (
	// ErrBareQuote -> bare \" in non-quoted-field
	ErrBareQuote = errors.New("bare \" in non-quoted-field")
	// ErrQuote -> extraneous \" in field
	ErrQuote = errors.New("extraneous \" in field")
)

// UnescapeSeparator cleans composed separator like `\t'.
func UnescapeSeparator(separator string) byte {
	if separator == "\\t" {
		separator = "\t"
	}
	return []byte(separator)[0]
}

// TrimNewline removes newline characters at the end of line.
func TrimNewline(line []byte) []byte {
	line = bytes.TrimRight(line, "\n")
	line = bytes.TrimRight(line, "\r")
	return bytes.TrimRight(line, "\n")
}

// A TsvParser reads records from a TSV-encoded file.
// As returned by NewTsvParser, a TsvParser expects input conforming to RFC 4180 (except the warning section).
//
// If LazyQuotes is true, a quote may appear in an unquoted field and a
// non-doubled quote may appear in a quoted field.
//
// /!\ Warning:
//
// - It does not support `\r\n' in quoted field.
//
// - It does not support comment.
type TsvParser struct {
	*Scanner
	err        error // Sticky error.
	Separator  byte
	QuoteChar  byte
	LazyQuotes bool // allow lazy quotes
	row        [][]byte
	separator  []byte // for internal purpose (see parseFields function)
	quoteChar  []byte // for internal purpose (see parseFields function)
}

// NewTsvParser inatanciates a new TsvParser.
func NewTsvParser(sc *Scanner, separator byte) *TsvParser {
	sc.Reset()
	return &TsvParser{
		Scanner:   sc,
		Separator: separator,
		QuoteChar: '"',
		separator: []byte{separator},
		quoteChar: []byte{'"'},
	}
}

// SyncConfig synchronizes the internal configuration to the Separator and QuoteChar attributes.
func (tp *TsvParser) SyncConfig() {
	tp.separator = []byte{tp.Separator}
	tp.quoteChar = []byte{tp.QuoteChar}
}

// error creates a new ParseError based on err.
func (tp *TsvParser) error(col int, err error) error {
	return &ParseError{
		Line:   tp.Line(),
		Column: col,
		Err:    err,
	}
}

// Err returns the first non-EOF error that was encountered by the Scanner.
func (tp *TsvParser) Err() error {
	return tp.err
}

// Row returns a slice of fields for the current row.
func (tp *TsvParser) Row() [][]byte {
	return tp.row
}

// Reset resets parser and its underliying scanner. It freeing the memory.
func (tp *TsvParser) Reset() {
	tp.Scanner.Reset()
	tp.row = make([][]byte, 0)
}

// ScanRow advances the TSV parser to the next row.
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

// Fields parser for the current read row
func (tp *TsvParser) parseFields() [][]byte {
	row := TrimNewline(tp.Bytes())
	if !bytes.Contains(row, tp.quoteChar) {
		// unquoted line (fast mode)
		return bytes.Split(row, tp.separator)
	}
	// quoted line (normal mode)
	fields := [][]byte{}
	r := newReader(row)
	field := tp.parseField(r)
	for field != nil {
		fields = append(fields, field)
		field = tp.parseField(r)
	}
	return fields
}

func (tp *TsvParser) parseField(r *reader) []byte {
	var field bytes.Buffer

	b, ok := r.readByte()
	if !ok {
		if r.isLastByte(tp.Separator) && !r.eof {
			r.eof = true
			return []byte{} // Empty field
		}
		return nil
	}

	switch b {
	case tp.Separator:
		return []byte{} // Empty field
	case tp.QuoteChar:
		// quoted field
		return tp.quotedField(r, field)

	default:
		// unquoted field
		return tp.unquotedField(r, field, b)
	}
}

func (tp *TsvParser) quotedField(r *reader, field bytes.Buffer) []byte {
	var b byte
	var ok bool

	// Fast mode
	// Enabled when the field is basic (e.g `...,"col 1",...')
	si := r.indexOf(tp.Separator)
	qci := r.indexOf(tp.QuoteChar)
	if qci < si && qci+1 == si {
		// qci < si -> there is no separator occurrence until the end of the field
		// qci+1 == si -> end of quoted field detection
		field.Write(r.readBytesTo(si))
		return field.Bytes()
	}

	// Normal mode
	// Enabled when the quoted field is more complex (e.g `...,"col ""is"" 1",...')
	for {
		b, ok = r.readByte()
		if !ok {
			// End of row reached
			if !tp.LazyQuotes {
				tp.err = tp.error(r.index, ErrQuote)
				return nil
			}
			return field.Bytes()
		}

		// CSV quote escaping case
		if b == tp.QuoteChar {
			b, ok = r.readByte() // read next byte after the double-quote
			if b == tp.Separator || !ok {
				// End of field or end of row reached
				return field.Bytes()
			}
			if b != tp.QuoteChar {
				if !tp.LazyQuotes {
					r.index--
					tp.err = tp.error(r.index, ErrQuote)
					return nil
				}
				// accept the bare quote
				field.WriteRune('"')
			}
		}

		field.WriteByte(b)
	}
}

func (tp *TsvParser) unquotedField(r *reader, field bytes.Buffer, b byte) []byte {
	var ok bool

	// Fast mode
	// Enabled when the field does not contain a double-quote (e.g `..,col1,..')
	si := r.indexOf(tp.Separator)
	qci := r.indexOf(tp.QuoteChar)
	if qci == -1 || si < qci {
		// qci == -1 -> no longer quote char in last part of the row
		// si < qsi -> there is no quote char until the next separator
		field.WriteByte(b)
		field.Write(r.readBytesTo(si))
		return field.Bytes()
	}

	// At this point, a quote char is present in the current unquoted field (e.g `col"5')
	// So we will parse and raise an error if malformatted TSV

	// Normal mode
	// Enabled when the field contains a double-quote or not
	for {
		field.WriteByte(b)
		b, ok = r.readByte()
		if b == tp.Separator || !ok {
			// End of field or end of row reached
			return field.Bytes()
		}

		if !tp.LazyQuotes && b == tp.QuoteChar {
			tp.err = tp.error(r.index, ErrBareQuote)
			return nil
		}
	}
}

// ------------------ //
// Parsing stuff      //
// ------------------ //

type reader struct {
	index int
	row   []byte
	eof   bool
}

func newReader(row []byte) *reader {
	return &reader{
		index: 0,
		row:   row,
	}
}

func (r *reader) readByte() (byte, bool) {
	if r.index >= len(r.row) {
		return '\u0000', false
	}
	defer func() { r.index++ }()
	return r.row[r.index], true
}

func (r *reader) readBytesTo(i int) []byte {
	if i == -1 {
		defer func() { r.index = len(r.row) }()
		return r.row[r.index:]
	}
	defer func() {
		// i is relative to r.index
		// 1 is the separator character (byte)
		r.index = r.index + i + 1
	}()
	return r.row[r.index:(r.index + i)]
}

// returns the index of the next occurrence of b or -1 if b is not present
func (r *reader) indexOf(b byte) int {
	return bytes.IndexByte(r.row[r.index:], b)
}

func (r *reader) isLastByte(b byte) bool {
	return r.row[len(r.row)-1] == b
}
