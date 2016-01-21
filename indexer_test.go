package iosupport_test

import (
	"os"
	"testing"

	"github.com/mdouchement/iosupport"
)

var indexerInput = "The first line.\nThe second line :)\nThe third\n\n"
var emptyIndexerInput = ""

func TestIndexerAnalyze(t *testing.T) {
	path := generateTmpFile(indexerInput)
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	sc := iosupport.NewScanner(file)
	actual := iosupport.NewIndexer(sc)

	actual.Analyze()

	expected := iosupport.NewIndexer(sc)
	expected.NbOfLines = 4
	expected.Lines = []iosupport.Line{iosupport.Line{0, 0, 16}, iosupport.Line{1, 16, 19}, iosupport.Line{2, 35, 10}, iosupport.Line{3, 45, 1}}

	t.Logf("expected: %v", expected)
	t.Logf("actual: %v", actual)
	if actual.NbOfLines != expected.NbOfLines {
		t.Errorf("Expected '%v' but got '%v' number of lines", expected.NbOfLines, actual.NbOfLines)
	}

	for i, expectedLine := range expected.Lines {
		if actual.Lines[i] != expectedLine {
			t.Errorf("Expected indexed line '%v' but got '%v' at index %v", expectedLine, actual.Lines[i], i)
		}
	}
}

func TestIndexerAnalyzeIsEmpty(t *testing.T) {
	path := generateTmpFile(emptyIndexerInput)
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	sc := iosupport.NewScanner(file)
	actual := iosupport.NewIndexer(sc)

	actual.Analyze()

	expected := iosupport.NewIndexer(sc)
	expected.NbOfLines = 0
	expected.Lines = []iosupport.Line{}

	if actual.NbOfLines != expected.NbOfLines {
		t.Errorf("Expected '%v' but got '%v' number of lines", expected.NbOfLines, actual.NbOfLines)
	}

	for i, expectedLine := range expected.Lines {
		if actual.Lines[i] != expectedLine {
			t.Errorf("Expected indexed line '%v' but got '%v' at index %v", expectedLine, actual.Lines[i], i)
		}
	}
}
