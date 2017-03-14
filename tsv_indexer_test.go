package iosupport_test

import (
	. "github.com/mdouchement/iosupport"
	"github.com/mdouchement/stringio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TsvIndexer", func() {
	Describe("#Analyze", func() {
		Context("with a well formatted TSV", func() {
			var sc = scanner("c1,c2,c3\nval45,val2,val3\nval40,val2,val6\n")
			var subject = NewTsvIndexer(sc, HasHeader(), Separator(","), Fields("c2", "c1"))

			err := subject.Analyze()
			check(err)

			It("indexes the TSV", func() {
				Expect(subject.Lines).To(TlConsistOf(tl{"", 0, 9}, tl{cs("val2", "val45"), 9, 16}, tl{cs("val2", "val40"), 25, 16}))
			})
		})

		Context("when the TSV has no header", func() {
			Context("and sort fields are invalid", func() {
				var sc = scanner("val1,val2,val3\nval4,val5,val6\nval7,val8,val9\n")
				var subject = NewTsvIndexer(sc, Separator(","), Fields("c2"))
				var err error

				err = subject.Analyze()

				It("does not index the TSV", func() {
					Expect(subject.Lines).To(HaveLen(0))
				})

				It("catches an error", func() {
					if err == nil {
						Fail("error is nil")
					}
					Expect(err.Error()).To(Equal("Field c2 do not match with pattern /var\\d+/"))
				})
			})

			Context("and sort fields are valid", func() {
				var sc = scanner("val1,val2,val3\nval4,val5,val6\nval7,val8,val9\n")
				var subject = NewTsvIndexer(sc, Separator(","), Fields("var2"))

				err := subject.Analyze()
				check(err)

				It("indexes the TSV", func() {
					Expect(subject.Lines).To(TlConsistOf(tl{cs("val2"), 0, 15}, tl{cs("val5"), 15, 15}, tl{cs("val8"), 30, 15}))
				})
			})
		})

		Context("when the TSV has empty cells", func() {
			var sc = scanner("c1,c2,c3\nval45,val2,\nval40,val2,val6\n")
			var subject = NewTsvIndexer(sc, HasHeader(), Separator(","), Fields("c3"))

			err := subject.Analyze()
			check(err)

			It("indexes the TSV", func() {
				Expect(subject.Lines).To(TlConsistOf(tl{"", 0, 9}, tl{cs(""), 9, 12}, tl{cs("val6"), 21, 16}))
			})
		})

		Context("when empty indexed fields are dropped", func() {
			var sc = scanner("c1,c2,c3\n1,0,42\n10,0,42\n,,42\n")
			var subject = NewTsvIndexer(sc, HasHeader(), Separator(","), Fields("c1", "c2"), DropEmptyIndexedFields())

			err := subject.Analyze()
			check(err)

			It("indexes the TSV", func() {
				Expect(subject.Lines).To(TlConsistOf(tl{"", 0, 9}, tl{cs("1", "0"), 9, 7}, tl{cs("10", "0"), 16, 8}))
			})
		})

		Context("when mal-formatted lines are skipped", func() {
			var sc = scanner("c1,c2,c3\n1,0,42\n10,0,42\n42,\n")
			var subject = NewTsvIndexer(sc, HasHeader(), Separator(","), Fields("c1", "c3"), SkipMalformattedLines())

			err := subject.Analyze()
			check(err)

			It("indexes the TSV", func() {
				Expect(subject.Lines).To(TlConsistOf(tl{"", 0, 9}, tl{cs("1", "42"), 9, 7}, tl{cs("10", "42"), 16, 8}))
			})
		})

		Context("when there are bad sort fields", func() {
			var sc = scanner("c1,c2,c3\nval45,val2,val3\nval40,val2,val6\n")
			var subject = NewTsvIndexer(sc, HasHeader(), Separator(","), Fields("___c1", "___c4"))
			var err error

			err = subject.Analyze()

			It("catches an error", func() {
				if err == nil {
					Fail("error is nil")
				}
				Expect(err.Error()).To(Equal("Invalid separator or sort fields"))
			})
		})

		Context("when the file is empty", func() {
			var sc = scanner("")
			var subject = NewTsvIndexer(sc, HasHeader(), Separator(","), Fields("c1", "c4"))

			err := subject.Analyze()
			check(err)

			It("does not index the TSV", func() {
				Expect(subject.Lines).To(HaveLen(0))
			})
		})
	})

	Describe("#Sort", func() {
		Context("with a well formatted TSV", func() {
			var sc = scanner("c1,c2,c3\n1,0,42\n10,0,42\n,,42\n")
			var subject = NewTsvIndexer(sc, HasHeader(), Separator(","), Fields("c1", "c2"))

			err := subject.Analyze()
			check(err)
			subject.Sort()

			It("sorts the index", func() {
				Expect(subject.Lines).To(TlConsistOf(tl{"", 0, 9}, tl{cs("", ""), 24, 5}, tl{cs("1", "0"), 9, 7}, tl{cs("10", "0"), 16, 8}))
			})
		})

		Context("when the file is empty", func() {
			var sc = scanner("")
			var subject = NewTsvIndexer(sc, HasHeader(), Separator(","), Fields("c1", "c4"))

			err := subject.Analyze()
			check(err)
			subject.Sort()

			It("does not index the TSV", func() {
				Expect(subject.Lines).To(HaveLen(0))
			})
		})
	})

	Describe("#Transfer", func() {
		Context("with a well formatted TSV", func() {
			var sc = scanner("c1,c2,c3\n1,0,42\n10,0,42\n,,42\n")
			var subject = NewTsvIndexer(sc, HasHeader(), Separator(","), Fields("c1", "c2"))
			var output = stringio.New()

			err := subject.Analyze()
			check(err)
			subject.Sort()
			err = subject.Transfer(output)
			check(err)

			It("sorts the index", func() {
				Expect(output.GetValueString()).To(Equal("c1,c2,c3\n,,42\n1,0,42\n10,0,42\n"))
			})
		})
	})

	Describe("Integration tests", func() {
		var limit uint64 = 4200 << 20
		var sc = scanner("c1,c2,c3\n1,0,42\n10,0,42\n,,42\na,b,c\ng,h,i\nd,e,f\n")
		var subject = NewTsvIndexer(sc, HasHeader(), Separator(","), Fields("c1", "c2"), SwapperOpts(limit, tempDir("", "tsv_swap_itg")))
		var output = stringio.New()

		backupGetMemoryUsage := GetMemoryUsage
		GetMemoryUsage = func() *HeapMemStat {
			return &HeapMemStat{0, limit + 42, 0, 0, 0, 0}
		}
		defer func() { GetMemoryUsage = backupGetMemoryUsage }()

		subject.Lines = make(TsvLines, 0, 2) // 2 lines per dump

		err := subject.Analyze()
		check(err)
		subject.Sort()
		err = subject.Transfer(output)
		check(err)

		It("succeeds", func() {
			Expect(output.GetValueString()).To(Equal("c1,c2,c3\n,,42\n1,0,42\n10,0,42\na,b,c\nd,e,f\ng,h,i\n"))
		})
	})
})
