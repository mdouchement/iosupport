package iosupport_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/mdouchement/iosupport"
)

var tsvParserInput = `c1,"c,2",c3
val45,val2,val3
val40,"val42 ""the"" best",val6
v1,,v3
v4,v5,
a,b,c`

var tsvParserErrQuote = []struct {
	col int
	row string
}{
	{6, `c1,"c"2",c3`},
	{9, `c1,c2,"c3`},
}
var tsvParserErrBareQuote = `c1,c2",c3`

func TestTsvParser(t *testing.T) {
	path := generateTmpFile(tsvParserInput)
	file, err := os.Open(path)
	check(err)
	sc := iosupport.NewScanner(file)
	parser := iosupport.NewTsvParser(sc, ',')

	var i int
	expectedRows := [][]string{
		[]string{"c1", "c,2", "c3"},
		[]string{"val45", "val2", "val3"},
		[]string{"val40", "val42 \"the\" best", "val6"},
		[]string{"v1", "", "v3"},
		[]string{"v4", "v5", ""},
		[]string{"a", "b", "c"},
	}
	for parser.ScanRow() {
		check(parser.Err())
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
				t.Logf("Expected '%v' - actual '%v' at index %d", expectedRows[i], actual, i)
				t.Errorf("  -> Expected '%v' but got '%v' at index %d", expected, actual[j], i)
			}
		}
		i++
	}
}

func TestTsvParserErrQuote(t *testing.T) {
	for _, input := range tsvParserErrQuote {
		path := generateTmpFile(input.row)
		file, err := os.Open(path)
		check(err)
		sc := iosupport.NewScanner(file)
		parser := iosupport.NewTsvParser(sc, ',')

		expected := fmt.Sprintf("line 0, column %d: %s", input.col, iosupport.ErrQuote)
		parser.ScanRow()
		if parser.Err().Error() != expected {
			t.Errorf("Expected '%v' but got '%v'", expected, parser.Err())
		}
	}
}

func TestTsvParserErrBareQuote(t *testing.T) {
	path := generateTmpFile(tsvParserErrBareQuote)
	file, err := os.Open(path)
	check(err)
	sc := iosupport.NewScanner(file)
	parser := iosupport.NewTsvParser(sc, ',')

	expected := "line 0, column 6: " + iosupport.ErrBareQuote.Error()
	parser.ScanRow()
	if parser.Err().Error() != expected {
		t.Errorf("Expected '%v' but got '%v'", expected, parser.Err())
	}
}

func BenchmarkParseFields(b *testing.B) {
	path := generateTmpFile(tsvIndexerInput)
	file, err := os.Open(path)
	check(err)
	sc := iosupport.NewScanner(file)

	tp := iosupport.NewTsvParser(sc, ',')
	iosupport.SetToken(tp, []byte("c1,c2,c3,c4,c5,c6,c7,c8,c9,10"))
	for i := 0; i < b.N; i++ {
		iosupport.ParseFields(tp)
	}
}

func BenchmarkParseFieldsWithQuotes(b *testing.B) {
	path := generateTmpFile(tsvIndexerInput)
	file, err := os.Open(path)
	check(err)
	sc := iosupport.NewScanner(file)

	tp := iosupport.NewTsvParser(sc, ',')
	iosupport.SetToken(tp, []byte(`c1,c2,c3,c4,c5,c6,"c,7",c8,c9,10`))
	for i := 0; i < b.N; i++ {
		iosupport.ParseFields(tp)
	}
}

func BenchmarkParseFieldsWithQuotesAndLongFields(b *testing.B) {
	path := generateTmpFile(tsvIndexerInput)
	file, err := os.Open(path)
	check(err)
	sc := iosupport.NewScanner(file)

	tp := iosupport.NewTsvParser(sc, ',')
	iosupport.SetToken(tp, []byte(`c1cccsfdergvergtv,cerfrwtgertgerygertg2,"c,7",c8`))
	for i := 0; i < b.N; i++ {
		iosupport.ParseFields(tp)
	}
}

func BenchmarkParseFieldsWithQuotesAndLongQuotedFields(b *testing.B) {
	path := generateTmpFile(tsvIndexerInput)
	file, err := os.Open(path)
	check(err)
	sc := iosupport.NewScanner(file)

	tp := iosupport.NewTsvParser(sc, ',')
	iosupport.SetToken(tp, []byte(`"c1cccsfdergvergtv","cerfrwtgertgerygertg2","c,7",c8`))
	for i := 0; i < b.N; i++ {
		iosupport.ParseFields(tp)
	}
}

func BenchmarkParseFieldsWithDoubleQuotes(b *testing.B) {
	path := generateTmpFile(tsvIndexerInput)
	file, err := os.Open(path)
	check(err)
	sc := iosupport.NewScanner(file)

	tp := iosupport.NewTsvParser(sc, ',')
	iosupport.SetToken(tp, []byte(`c1,c2,c3,c4,c5,c6,"c ""is"" 7",c8,c9,10`))
	for i := 0; i < b.N; i++ {
		iosupport.ParseFields(tp)
	}
}
