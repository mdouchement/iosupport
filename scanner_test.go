package iosupport_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/mdouchement/iosupport"
)

var data string = "The first line.\nThe second line :)\n\n"
var dataCR string = "The first line.\rThe second line :)\r\r"
var dataCRLF string = "The first line.\r\nThe second line.\r\n\r\n"
var dataLFCR string = "The first line.\n\rThe second line.\n\r\n\r"

func TestScannerScanLine(t *testing.T) {
	for di, d := range []string{data, dataCR, dataCRLF, dataLFCR} {
		path := generateTmpFile(d)
		file, err := os.Open(path)
		check(err)
		defer file.Close()

		sc := iosupport.NewScanner(file)

		for i, expected := range []bool{true, true, true, false} {
			actual := sc.ScanLine()
			if actual != expected {
				t.Errorf("[Data index: %v - Line %v]: Expected '%v' but got '%v'", di, i+1, expected, actual)
			}
		}
	}
}

func TestScannerBytes(t *testing.T) {
	path := generateTmpFile()
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	sc := iosupport.NewScanner(file)
	lines := strings.Split(data, "\n")
	offsets := []uint64{0, 16, 35, 36}
	limits := []uint32{16, 19, 1, 1}

	for i, bts := range []string(lines) {
		sc.ScanLine()
		expected := []byte(bts)
		actual := sc.Bytes()
		t.Logf("expected: %v", expected)
		t.Logf("actual:   %v", actual)
		if !bytes.Equal(actual, expected) {
			t.Errorf("Expected '%v' but got '%v' at index %d", expected, actual, i)
		}
		if sc.Offset() != offsets[i] {
			t.Errorf("Expected offset '%v' but got '%v' at index %d", offsets[i], sc.Offset(), i)
		}
		if sc.Limit() != limits[i] {
			t.Errorf("Expected limit '%v' but got '%v' at index %d", limits[i], sc.Limit(), i)
		}
		if sc.Line() != i {
			t.Errorf("Expected line index '%v' but got '%v' at index %d", i, sc.Line(), i)
		}
	}
}

func TestScannerBytesKeepNewlineSequence(t *testing.T) {
	path := generateTmpFile()
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	sc := iosupport.NewScanner(file)
	sc.KeepNewlineSequence(true)
	lines := strings.Split(data, "\n")

	for i, bts := range []string(lines) {
		sc.ScanLine()
		eol := ""
		if i+1 < len(lines) {
			eol = "\n"
		}
		expected := []byte(bts + eol)
		actual := sc.Bytes()
		if !bytes.Equal(actual, expected) {
			t.Errorf("Expected '%v' but got '%v'", expected, actual)
		}
	}
}

func TestScannerText(t *testing.T) {
	path := generateTmpFile()
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	sc := iosupport.NewScanner(file)
	lines := strings.Split(data, "\n")

	for _, expected := range []string(lines) {
		sc.ScanLine()
		actual := sc.Text()
		if actual != expected {
			t.Errorf("Expected '%v' but got '%v'", expected, actual)
		}
	}
}

func TestScannerTextKeepNewlineSequence(t *testing.T) {
	path := generateTmpFile()
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	sc := iosupport.NewScanner(file)
	sc.KeepNewlineSequence(true)
	lines := strings.Split(data, "\n")

	for i, expected := range []string(lines) {
		sc.ScanLine()
		if i+1 < len(lines) {
			expected = expected + "\n"
		}
		actual := sc.Text()
		if actual != expected {
			t.Errorf("Expected '%v' but got '%v'", expected, actual)
		}
	}
}

func TestScannerEachLine(t *testing.T) {
	path := generateTmpFile()
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	sc := iosupport.NewScanner(file)
	lines := strings.Split(data, "\n")
	i := 0

	sc.EachLine(func(actual []byte, err error) {
		expected := []byte(lines[i])
		if !bytes.Equal(actual, expected) {
			t.Errorf("Expected '%v' but got '%v'", expected, actual)
		}
		i++
	})
}

func TestScannerEachString(t *testing.T) {
	path := generateTmpFile()
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	sc := iosupport.NewScanner(file)
	expected := strings.Split(data, "\n")
	i := 0

	sc.EachString(func(actual string, err error) {
		if actual != expected[i] {
			t.Errorf("Expected '%v' but got '%v'", expected[i], actual)
		}
		i++
	})
}

// Re-read file
func TestScannerReset(t *testing.T) {
	path := generateTmpFile()
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	sc := iosupport.NewScanner(file)
	for i := 1; i < 2; i++ {
		for j, expected := range []bool{true, true, true, false} {
			actual := sc.ScanLine()
			if actual != expected {
				t.Errorf("[Pass %v:%v]: Expected '%v' but got '%v'", i+1, j+1, expected, actual)
			}
		}
		sc.Reset()
	}
}

func TestScannerIsLineEmpty(t *testing.T) {
	path := generateTmpFile()
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	sc := iosupport.NewScanner(file)

	for i, expected := range []bool{false, false, true, true} {
		sc.ScanLine()
		actual := sc.IsLineEmpty()
		if actual != expected {
			t.Errorf("[Pass %v]: Expected '%v' but got '%v'", i+1, expected, actual)
		}
	}
}

func generateTmpFile(input ...string) string {
	var d string
	switch len(input) {
	case 0:
		d = data
	case 1:
		d = input[0]
	default:
		panic("Too many arguments")
	}

	path := "/tmp/iosupport_test.txt"
	file, err := os.Create(path)
	check(err)
	defer file.Close()
	_, err = file.WriteString(d)
	check(err)
	err = file.Sync()
	check(err)

	return path
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
