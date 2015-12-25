package iosupport_test

import (
	"os"
	"testing"

	"github.com/mdouchement/iosupport"
)

var tsvIndexerInput = "c1,c2,c3\nval1,val2,val3\nval4,val5,val6\n"
var tsvIndexerInputFields = []string{"c2", "c1"}

var tsvIndexerInputWithoutHeader = "val1,val2,val3\nval4,val5,val6\nval7,val8,val9\n"
var tsvIndexerInputFieldsWithoutHeader = []string{"var2"}

var emptyTsvIndexerInput = ""

func TestTsvIndexerAnalyze(t *testing.T) {
	file, actual, expected := prepareTsvIndexer(tsvIndexerInput)
	defer file.Close()

	actual.Analyze()

	t.Logf("expected.Lines: %v", expected.Lines)
	t.Logf("actual.Lines: %v", actual.Lines)
	for i, expectedLine := range expected.Lines {

		if actual.Lines[i].Index != expectedLine.Index {
			t.Errorf("Expected '%v' but got '%v' at index %v", actual.Lines[i].Index, expectedLine.Index, i)
		}

		for j, exepectedComparable := range expectedLine.Comparables {
			if actual.Lines[i].Comparables[j] != exepectedComparable {
				t.Errorf("Expected '%v' but got '%v' at index %v", actual.Lines[i].Comparables[j], exepectedComparable, i)
			}
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
	expected.Lines = []iosupport.TsvLine{iosupport.TsvLine{0, []string{"val2"}}, iosupport.TsvLine{1, []string{"val5"}}, iosupport.TsvLine{2, []string{"val8"}}}

	t.Logf("expected.Lines: %v", expected.Lines)
	t.Logf("actual.Lines: %v", actual.Lines)
	for i, expectedLine := range expected.Lines {
		if actual.Lines[i].Index != expectedLine.Index {
			t.Errorf("Expected '%v' but got '%v' at index %v", actual.Lines[i].Index, expectedLine.Index, i)
		}

		for j, exepectedComparable := range expectedLine.Comparables {
			if actual.Lines[i].Comparables[j] != exepectedComparable {
				t.Errorf("Expected '%v' but got '%v' at index %v", actual.Lines[i].Comparables[j], exepectedComparable, i)
			}
		}
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

func TestTsvSortAnalyze(t *testing.T) {
	file, actual, expected := prepareTsvIndexer(tsvIndexerInput)
	defer file.Close()

	actual.Analyze()
	actual.Sort()

	t.Logf("expected.Lines: %v", expected.Lines)
	t.Logf("actual.Lines: %v", actual.Lines)
	for i, expectedLine := range expected.Lines {
		if actual.Lines[i].Index != expectedLine.Index {
			t.Errorf("Expected '%v' but got '%v' at index %v", actual.Lines[i].Index, expectedLine.Index, i)
		}

		for j, exepectedComparable := range expectedLine.Comparables {
			if actual.Lines[i].Comparables[j] != exepectedComparable {
				t.Errorf("Expected '%v' but got '%v' at index %v", actual.Lines[i].Comparables[j], exepectedComparable, i)
			}
		}
	}
}

// ----------------------- //
// Helpers                 //
// ----------------------- //

func prepareTsvIndexer(input string) (file *os.File, actual *iosupport.TsvIndexer, expected *iosupport.TsvIndexer) {
	path := generateTmpFile(input)
	var err error
	file, err = os.Open(path)
	check(err)

	sc := iosupport.NewScanner(file)
	actual = iosupport.NewTsvIndexer(sc, true, ",", tsvIndexerInputFields)

	expected = iosupport.NewTsvIndexer(sc, true, ",", tsvIndexerInputFields)
	expected.I = iosupport.NewIndexer(sc)
	expected.I.NbOfLines = 3
	expected.I.Lines = []iosupport.Line{iosupport.Line{0, 0, 9}, iosupport.Line{1, 10, 15}, iosupport.Line{2, 16, 15}}

	expected.Lines = []iosupport.TsvLine{iosupport.TsvLine{0, []string{"", ""}}, iosupport.TsvLine{1, []string{"val2", "val1"}}, iosupport.TsvLine{2, []string{"val5", "val4"}}}
	return
}
