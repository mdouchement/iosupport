package iosupport

import (
	"bufio"
	"io"
)

// This Scanner provides very large file reader (also file with very long lines).
// Main usages:
//
// sc := supports.NewScanner(file)
// for sc.ScanLine() {
//   println(sc.Text())
//   println(sc.Bytes())
// }
//
// sc := supports.NewScanner(file)
// sc.EachString(func(line string, err error) {
//   println(line)
// })
//
// sc := supports.NewScanner(file)
// sc.EachLine(func(line []byte, err error) {
//   println(line)
// })
//
// See other methods for custom usage

var (
	// LF -> linefeed
	LF byte = '\n'
	// CR -> carriage return
	CR       byte = '\r'
	newLines      = []byte{CR, LF}
)

// Scanner conatins all stuff for reading a buffered file
type Scanner struct {
	f       FileReader    // The file provided by the client.
	r       *bufio.Reader // Buffered reader on given file.
	keepnls bool          // Keep the newline sequence in returned strings
	token   []byte        // Last token returned by split (scan).
	err     error         // Sticky error.
	line    int           // index of current read line
	offset  uint64        // Offset of the start of the read line
	limit   uint32        // Length of the read line including newline sequence
}

// NewScanner instanciates a Scanner
func NewScanner(f FileReader) *Scanner {
	return &Scanner{
		f:       f,
		r:       bufio.NewReader(f),
		keepnls: false,
		line:    -1,
	}
}

// KeepNewlineSequence keeps the newline sequence in read lines
func (s *Scanner) KeepNewlineSequence(b bool) {
	s.keepnls = b
}

// Bytes returns the most recent line generated by a call to Scan.
// The underlying array may point to data that will be overwritten
// by a subsequent call to Scan. It does no allocation.
func (s *Scanner) Bytes() []byte {
	return s.token
}

// Text returns the most recent line generated by a call to Scan
// as a newly allocated string holding its bytes.
func (s *Scanner) Text() string {
	return string(s.token)
}

// Err returns the first non-EOF error that was encountered by the Scanner.
func (s *Scanner) Err() error {
	if s.err == nil || s.err == io.EOF {
		return nil
	}
	return s.err
}

// Line return the index of the current line
func (s *Scanner) Line() int {
	return s.line
}

// Offset return the byte offset of the current line
func (s *Scanner) Offset() uint64 {
	return s.offset
}

// Limit return the byte length of the current line including newline sequence
func (s *Scanner) Limit() uint32 {
	return s.limit
}

// ScanLine advances the Scanner to the next line), which will then be
// available through the Bytes or Text method. It returns false when the
// scan stops, either by reaching the end of the input or an error.
// After Scan returns false, the Err method will return any error that
// occurred during scanning, except that if it was io.EOF, Err
// will return nil.
func (s *Scanner) ScanLine() bool {
	// Override token value to new bytes array
	s.token = make([]byte, 0)
	s.offset += uint64(s.limit)
	s.line++

	// Loop until we have a token.
	for {
		b, err := s.r.ReadByte()

		// End-of-file detection or error detection
		if err != nil {
			s.err = err
			return !s.IsLineEmpty()
		}

		switch b {
		case LF:
			s.handleNewLineSequence(LF, CR)
			return true
		case CR:
			s.handleNewLineSequence(CR, LF)
			return true
		default:
			s.token = append(s.token, b)
		}
	}
}

// EachLine iterate on each line and execute the given function
func (s *Scanner) EachLine(fn func([]byte, error)) {
	s.Reset()
	for s.ScanLine() {
		fn(s.Bytes(), s.Err())
	}
}

// EachString iterates on each line as string format and execute the given function
func (s *Scanner) EachString(fn func(string, error)) {
	s.Reset()
	for s.ScanLine() {
		fn(s.Text(), s.Err())
	}
}

// ReadAt reads len(b) bytes from the Scanner starting at byte offset off.
// It returns the number of bytes read and the error, if any.
// ReadAt always returns a non-nil error when n < len(b).
// At end of scanner, that error is io.EOF.
func (s *Scanner) ReadAt(offset int64, limit int) ([]byte, error) {
	token := make([]byte, limit)
	if _, err := s.f.ReadAt(token, offset); err != nil {
		return nil, err
	}
	return token, nil
}

// Records the first error encountered.
func (s *Scanner) setErr(err error) {
	if s.err == nil || s.err == io.EOF {
		s.err = err
	}
}

// IsLineEmpty says if the current line is empty (only when newline character is not keeped)
func (s *Scanner) IsLineEmpty() bool {
	return len(s.token) == 0
}

// Reset seek to top of file and clean buffer
func (s *Scanner) Reset() {
	s.f.Seek(0, 0)
	s.line = -1
	s.limit = 0
	s.offset = 0
	s.token = []byte{}
}

func (s *Scanner) handleNewLineSequence(currentNl, nextNl byte) {
	if s.keepnls {
		// Keep current newline character (relative to the seek)
		s.token = append(s.token, currentNl)
		s.limit = uint32(len(s.token))
	} else {
		s.limit = uint32(len(s.token) + 1)
	}

	for {
		b, err := s.r.Peek(1)
		if err != nil {
			s.err = err
			return
		}

		if b[0] == nextNl {
			if s.keepnls {
				// Keep next newline character (relative to the currentNl)
				s.token = append(s.token, nextNl)
				s.limit = uint32(len(s.token))
			} else {
				s.limit++
			}
			s.r.ReadByte()
			return
		}
		return
	}
}
