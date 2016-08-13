package iosupport_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mdouchement/iosupport"
)

var tsvIndexerInput = "c1,c2,c3\nval45,val2,val3\nval40,val2,val6\n"
var tsvIndexerInputFields = []string{"c2", "c1"}

var tsvIndexerInputWithoutHeader = "val1,val2,val3\nval4,val5,val6\nval7,val8,val9\n"
var tsvIndexerInputFieldsWithoutHeader = []string{"var2"}

var emptyTsvIndexerInput = ""

func TestTsvIndexerAnalyze(t *testing.T) {
	file, actual, expected := prepareTsvIndexer(tsvIndexerInput)
	defer file.Close()

	actual.Analyze()

	t.Logf("expected.Lines: %v", expected.Lines)
	t.Logf("actual.Lines:   %v", actual.Lines)
	for i, expectedLine := range expected.Lines {
		if actual.Lines[i].Offset != expectedLine.Offset {
			t.Errorf("Expected offset '%v' but got '%v' at index %v", expectedLine.Offset, actual.Lines[i].Offset, i)
		}
		if actual.Lines[i].Limit != expectedLine.Limit {
			t.Errorf("Expected limit '%v' but got '%v' at index %v", expectedLine.Limit, actual.Lines[i].Limit, i)
		}

		if expectedLine.Comparable != actual.Lines[i].Comparable {
			t.Errorf("Expected '%v' but got '%v' at index %v", expectedLine.Comparable, actual.Lines[i].Comparable, i)
		}
	}
}

func TestTsvIndexerAnalyzeWithoutHeader(t *testing.T) {
	file, actual, expected := prepareTsvIndexer(tsvIndexerInputWithoutHeader)
	defer file.Close()

	actual.Header = false

	// When fields are invalid
	if err := actual.Analyze(); err == nil || err.Error() != "Field c2 do not match with pattern /var\\d+/" {
		t.Error("Expect Analyze to returns an errro")
	}
	actual.Fields = tsvIndexerInputFieldsWithoutHeader

	// When fields have a valid format
	if err := actual.Analyze(); err != nil {
		t.Errorf("Expect Analyze to not return an error but returns: %v", err)
	}

	expected.Fields = tsvIndexerInputFieldsWithoutHeader
	expected.Lines = []iosupport.TsvLine{iosupport.TsvLine{"val2", 0, 0}, iosupport.TsvLine{"val5", 0, 0}, iosupport.TsvLine{"val8", 0, 0}}

	t.Logf("expected.Lines: %v", expected.Lines)
	t.Logf("actual.Lines:   %v", actual.Lines)
	for i, expectedLine := range expected.Lines {
		if actual.Lines[i].Comparable != expectedLine.Comparable {
			t.Errorf("Expected '%v' but got '%v' at index %v", expectedLine.Comparable, actual.Lines[i].Comparable, i)
		}
	}
}

func TestTsvIndexerAnalyzeHasBadFields(t *testing.T) {
	file, actual, _ := prepareTsvIndexer(tsvIndexerInput)
	defer file.Close()

	actual.Fields = []string{"___c2", "___c1"}
	err := actual.Analyze()

	if err.Error() != "Invalid separator or sort fields" {
		t.Errorf("Expected 'Invalid separator or sort fields' but got '%s'", err.Error())
	}
}

func TestTsvIndexerAnalyzeIsEmpty(t *testing.T) {
	file, actual, _ := prepareTsvIndexer(emptyTsvIndexerInput)
	defer file.Close()

	actual.Analyze()

	if len(actual.Lines) > 0 {
		t.Errorf("Expected '%v' to be empty", actual.Lines)
	}
}

func TestTsvIndexerSortIsEmpty(t *testing.T) {
	file, actual, _ := prepareTsvIndexer(emptyTsvIndexerInput)
	defer file.Close()

	actual.Analyze()
	actual.Sort()

	if len(actual.Lines) > 0 {
		t.Errorf("Expected '%v' to be empty", actual.Lines)
	}
}

func TestTsvSort(t *testing.T) {
	file, actual, expected := prepareTsvIndexer(tsvIndexerInput)
	defer file.Close()

	err := actual.Analyze()
	check(err)
	actual.Sort()

	expected.Lines = []iosupport.TsvLine{iosupport.TsvLine{"", 0, 9}, iosupport.TsvLine{"val2val40", 25, 16}, iosupport.TsvLine{"val2val45", 9, 16}}

	t.Logf("expected.Lines: %v", expected.Lines)
	t.Logf("actual.Lines:   %v", actual.Lines)
	for i, expectedLine := range expected.Lines {
		if actual.Lines[i].Comparable != expectedLine.Comparable {
			t.Errorf("Expected '%v' but got '%v' at index %v", expectedLine.Comparable, actual.Lines[i].Comparable, i)
		}
	}
}

func TestTsvTransfer(t *testing.T) {
	ifile, current, _ := prepareTsvIndexer(tsvIndexerInput)
	defer ifile.Close()
	ofile, err := ioutil.TempFile("/tmp", "tsv_transfer")
	check(err)
	defer ofile.Close()

	current.Analyze()
	current.Sort()
	current.Transfer(ofile)

	buff, err := ioutil.ReadFile(ofile.Name())
	check(err)
	actual := string(buff)
	expected := "c1,c2,c3\nval40,val2,val6\nval45,val2,val3\n"
	if actual != expected {
		t.Errorf("Expected:\n%v but got:\n%v", expected, actual)
	}
}

// ----------------------- //
// Helpers                 //
// ----------------------- //

func prepareTsvIndexer(input string) (file *os.File, actual *iosupport.TsvIndexer, expected *iosupport.TsvIndexer) {
	path := generateTmpFile(input)
	var err error

	sc := func() *iosupport.Scanner {
		file, err = os.Open(path)
		check(err)
		return iosupport.NewScanner(file)
	}

	actual = iosupport.NewTsvIndexer(sc, true, ",", tsvIndexerInputFields)

	expected = iosupport.NewTsvIndexer(sc, true, ",", tsvIndexerInputFields)
	// expected.I = iosupport.NewIndexer(sc())
	// expected.I.NbOfLines = 3

	expected.Lines = []iosupport.TsvLine{iosupport.TsvLine{"", 0, 9}, iosupport.TsvLine{"val2val45", 9, 16}, iosupport.TsvLine{"val2val40", 25, 16}}
	return
}
