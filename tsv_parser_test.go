package iosupport_test

import (
	"os"
	"testing"

	"github.com/mdouchement/iosupport"
)

func TestTsvParser(t *testing.T) {
	path := generateTmpFile(tsvIndexerInput)
	file, err := os.Open(path)
	check(err)
	sc := iosupport.NewScanner(file)
	parser := iosupport.NewTsvParser(sc, ',')

	var i int
	expectedRows := [][]string{
		[]string{"c1", "c2", "c3"},
		[]string{"val45", "val2", "val3"},
		[]string{"val40", "val2", "val6"},
	}
	for parser.ScanRow() {
		if parser.Line() != sc.Line() {
			t.Errorf("Expected line index '%v' but got '%v'", sc.Line(), parser.Line())
		}
		if parser.Offset() != sc.Offset() {
			t.Errorf("Expected offset '%v' but got '%v'", sc.Offset(), parser.Offset())
		}
		if parser.Limit() != sc.Limit() {
			t.Errorf("Expected limit '%v' but got '%v'", sc.Limit(), parser.Limit())
		}

		var actual []string
		for _, field := range parser.Row() {
			actual = append(actual, string(field))
		}
		for j, expected := range expectedRows[i] {
			if expected != actual[j] {
				t.Logf("expected '%v' - actual '%v' at index %d", expectedRows[i], actual, i)
				t.Errorf("Expected '%v' but got '%v' at index %d", expected, actual[j], i)
			}
		}
		i++
	}
}
