package iosupport

import "io"

// FileReader is an interface for supported File by Scanner
type FileReader interface {
	ReadAt(b []byte, off int64) (n int, err error)
	Seek(offset int64, whence int) (ret int64, err error)
	Name() string
	Close() error

	io.Reader
}

// FileWriter is an interface for supported File by Scanner
type FileWriter interface {
	Close() error

	io.Writer
}
