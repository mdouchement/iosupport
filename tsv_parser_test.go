package iosupport_test

import (
	"fmt"
	"testing"

	"github.com/MakeNowJust/heredoc"
	. "github.com/mdouchement/iosupport"
	"github.com/mdouchement/stringio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TsvParser", func() {
	Describe("#ScanRow", func() {
		Context("with a well formatted TSV", func() {
			var file = stringio.NewFromString(heredoc.Doc(`c1,"c,2",c3
				val45,val2,val3
				val40,"val42 ""the"" best",val6
				v1,,v3
				v4,v5,
				a,b,c`))

			var sc = NewScanner(file)
			var subject = NewTsvParser(sc, ',')

			var actual = [][]string{}
			var expected = [][]string{
				{"c1", "c,2", "c3"},
				{"val45", "val2", "val3"},
				{"val40", "val42 \"the\" best", "val6"},
				{"v1", "", "v3"},
				{"v4", "v5", ""},
				{"a", "b", "c"},
			}

			for subject.ScanRow() {
				check(subject.Err())
				actual = append(actual, toStringSlice(subject.Row()))
			}

			It("parses the TSV", func() {
				Expect(actual).To(Equal(expected))
			})

			It("has the same line's index of its inner scanner", func() {
				Expect(subject.Line()).To(Equal(sc.Line()))
			})

			It("has the same line's offset of its inner scanner", func() {
				Expect(subject.Offset()).To(Equal(sc.Offset()))
			})

			It("has the same line's limit of its inner scanner", func() {
				Expect(subject.Limit()).To(Equal(sc.Limit()))
			})
		})

		Context("when there is a quote error", func() {
			var tsvParserErrQuote = []struct {
				col int
				row string
			}{
				{6, `c1,"c"2",c3`},
				{9, `c1,c2,"c3`},
			}

			It("detects the error", func() {
				for _, input := range tsvParserErrQuote {
					var file = stringio.NewFromString(input.row)

					var subject = NewTsvParser(NewScanner(file), ',')
					subject.ScanRow()

					Expect(subject.Err().Error()).To(Equal(fmt.Sprintf("line 0, column %d: %s", input.col, ErrQuote)))
				}
			})
		})

		Context("when there is a bare quote error", func() {
			var file = stringio.NewFromString(`c1,c2",c3`)

			var subject = NewTsvParser(NewScanner(file), ',')

			It("detects the error", func() {
				subject.ScanRow()
				Expect(subject.Err().Error()).To(Equal("line 0, column 6: " + ErrBareQuote.Error()))
			})
		})
	})
})

// ------------------ //
// Benchmarks         //
// ------------------ //

func BenchmarkParseFields(b *testing.B) {
	sc := NewScanner(stringio.NewFromString(""))

	tp := NewTsvParser(sc, ',')
	SetToken(tp, []byte("c1,c2,c3,c4,c5,c6,c7,c8,c9,10"))
	for i := 0; i < b.N; i++ {
		ParseFields(tp)
	}
}

func BenchmarkParseFieldsWithQuotes(b *testing.B) {
	sc := NewScanner(stringio.NewFromString(""))

	tp := NewTsvParser(sc, ',')
	SetToken(tp, []byte(`c1,c2,c3,c4,c5,c6,"c,7",c8,c9,10`))
	for i := 0; i < b.N; i++ {
		ParseFields(tp)
	}
}

func BenchmarkParseFieldsWithQuotesAndLongFields(b *testing.B) {
	sc := NewScanner(stringio.NewFromString(""))

	tp := NewTsvParser(sc, ',')
	SetToken(tp, []byte(`c1cccsfdergvergtv,cerfrwtgertgerygertg2,"c,7",c8`))
	for i := 0; i < b.N; i++ {
		ParseFields(tp)
	}
}

func BenchmarkParseFieldsWithQuotesAndLongQuotedFields(b *testing.B) {
	sc := NewScanner(stringio.NewFromString(""))

	tp := NewTsvParser(sc, ',')
	SetToken(tp, []byte(`"c1cccsfdergvergtv","cerfrwtgertgerygertg2","c,7",c8`))
	for i := 0; i < b.N; i++ {
		ParseFields(tp)
	}
}

func BenchmarkParseFieldsWithDoubleQuotes(b *testing.B) {
	sc := NewScanner(stringio.NewFromString(""))

	tp := NewTsvParser(sc, ',')
	SetToken(tp, []byte(`c1,c2,c3,c4,c5,c6,"c ""is"" 7",c8,c9,10`))
	for i := 0; i < b.N; i++ {
		ParseFields(tp)
	}
}
