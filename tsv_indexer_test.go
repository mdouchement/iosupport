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
	t.Logf("actual.Lines: %v", actual.Lines)
	for i, expectedLine := range expected.Lines {

		if actual.Lines[i].Index != expectedLine.Index {
			t.Errorf("Expected '%v' but got '%v' at index %v", expectedLine.Index, actual.Lines[i].Index, i)
		}

		for j, exepectedComparable := range expectedLine.Comparables {
			if actual.Lines[i].Comparables[j] != exepectedComparable {
				t.Errorf("Expected '%v' but got '%v' at index %v", exepectedComparable, actual.Lines[i].Comparables[j], i)
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
	expected.Lines = []iosupport.TsvLine{iosupport.TsvLine{0, []string{"val2"}, 0, 0}, iosupport.TsvLine{1, []string{"val5"}, 0, 0}, iosupport.TsvLine{2, []string{"val8"}, 0, 0}}

	t.Logf("expected.Lines: %v", expected.Lines)
	t.Logf("actual.Lines: %v", actual.Lines)
	for i, expectedLine := range expected.Lines {
		if actual.Lines[i].Index != expectedLine.Index {
			t.Errorf("Expected '%v' but got '%v' at index %v", expectedLine.Index, actual.Lines[i].Index, i)
		}

		for j, exepectedComparable := range expectedLine.Comparables {
			if actual.Lines[i].Comparables[j] != exepectedComparable {
				t.Errorf("Expected '%v' but got '%v' at index %v", exepectedComparable, actual.Lines[i].Comparables[j], i)
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

func TestTsvSort(t *testing.T) {
	file, actual, expected := prepareTsvIndexer(tsvIndexerInput)
	defer file.Close()

	actual.Analyze()
	actual.Sort()

	expected.Lines = []iosupport.TsvLine{iosupport.TsvLine{0, []string{"", ""}, 0, 9}, iosupport.TsvLine{2, []string{"val2", "val40"}, 16, 15}, iosupport.TsvLine{1, []string{"val2", "val45"}, 10, 15}}

	t.Logf("expected.Lines: %v", expected.Lines)
	t.Logf("actual.Lines: %v", actual.Lines)
	for i, expectedLine := range expected.Lines {
		if actual.Lines[i].Index != expectedLine.Index {
			t.Errorf("Expected '%v' but got '%v' at index %v", expectedLine.Index, actual.Lines[i].Index, i)
		}

		for j, expectedComparable := range expectedLine.Comparables {
			if actual.Lines[i].Comparables[j] != expectedComparable {
				t.Errorf("Expected '%v' but got '%v' at index %v", expectedComparable, actual.Lines[i].Comparables[j], i)
			}
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
	file, err = os.Open(path)
	check(err)

	sc := iosupport.NewScanner(file)
	actual = iosupport.NewTsvIndexer(sc, true, ",", tsvIndexerInputFields)

	expected = iosupport.NewTsvIndexer(sc, true, ",", tsvIndexerInputFields)
	expected.I = iosupport.NewIndexer(sc)
	expected.I.NbOfLines = 3
	// expected.I.Lines = []iosupport.Line{iosupport.Line{0, 0, 9}, iosupport.Line{1, 10, 15}, iosupport.Line{2, 16, 15}}

	expected.Lines = []iosupport.TsvLine{iosupport.TsvLine{0, []string{"", ""}, 0, 9}, iosupport.TsvLine{1, []string{"val2", "val45"}, 10, 15}, iosupport.TsvLine{2, []string{"val2", "val40"}, 16, 15}}
	return
}
