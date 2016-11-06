package iosupport_test

import (
	. "github.com/mdouchement/iosupport"
	"github.com/rsniezynski/stringio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scanner", func() {
	Describe("#ScanLine", func() {
		Context("with a normal file", func() {
			var file = stringio.NewFromString("The first line.\nThe second line :)\n\n")
			var subject = NewScanner(file)
			var actual = []string{}

			for subject.ScanLine() {
				check(subject.Err())
				actual = append(actual, subject.Text())
			}

			It("reads the file", func() {
				Expect(actual).To(ConsistOf("The first line.", "The second line :)", ""))
			})
		})

		Context("when the file does ends with a newline (EOF error)", func() {
			var file = stringio.NewFromString("The first line.\nThe second line :)")
			var subject = NewScanner(file)
			var actual = []string{}

			for subject.ScanLine() {
				check(subject.Err())
				actual = append(actual, subject.Text())
			}

			It("reads the file", func() {
				Expect(actual).To(ConsistOf("The first line.", "The second line :)"))
			})
		})

		Context("when the file uses CR", func() {
			var file = stringio.NewFromString("The first line.\rThe second line :)\r\r")
			var subject = NewScanner(file)
			var actual = []string{}

			for subject.ScanLine() {
				check(subject.Err())
				actual = append(actual, subject.Text())
			}

			It("reads the file", func() {
				Expect(actual).To(ConsistOf("The first line.", "The second line :)", ""))
			})
		})

		Context("when the file uses CRLF", func() {
			var file = stringio.NewFromString("The first line.\r\nThe second line.\r\n\r\n")
			var subject = NewScanner(file)
			var actual = []string{}

			for subject.ScanLine() {
				check(subject.Err())
				actual = append(actual, subject.Text())
			}

			It("reads the file", func() {
				Expect(actual).To(ConsistOf("The first line.", "The second line.", ""))
			})
		})

		Context("when the file uses LFCR", func() {
			var file = stringio.NewFromString("The first line.\n\rThe second line.\n\r\n\r")
			var subject = NewScanner(file)
			var actual = []string{}

			for subject.ScanLine() {
				check(subject.Err())
				actual = append(actual, subject.Text())
			}

			It("reads the file", func() {
				Expect(actual).To(ConsistOf("The first line.", "The second line.", ""))
			})
		})
	})

	Describe("#KeepNewlineSequence", func() {
		var file = stringio.NewFromString("The first line.\r\nThe second line :)\n\n")
		var subject = NewScanner(file)
		subject.KeepNewlineSequence(true)
		var actual = []string{}

		for subject.ScanLine() {
			actual = append(actual, subject.Text())
		}

		It("keeps newline sequence in read lines", func() {
			Expect(actual).To(ConsistOf("The first line.\r\n", "The second line :)\n", "\n"))
		})
	})

	Describe("#Text", func() {
		var file = stringio.NewFromString("The first line.")
		var subject = NewScanner(file)

		subject.ScanLine()
		check(subject.Err())

		It("reads the first line as bytes slice", func() {
			Expect(subject.Text()).To(Equal("The first line."))
		})
	})

	Describe("#Bytes", func() {
		var file = stringio.NewFromString("The first line.")
		var subject = NewScanner(file)

		subject.ScanLine()
		check(subject.Err())

		It("reads the first line as bytes slice", func() {
			Expect(subject.Bytes()).To(Equal([]byte("The first line.")))
		})
	})

	Describe("#Line", func() {
		var file = stringio.NewFromString("The first line.\nThe second line :)\n\n")
		var subject = NewScanner(file)
		var actual = []int{}

		for subject.ScanLine() {
			check(subject.Err())
			actual = append(actual, subject.Line())
		}

		It("reads the file", func() {
			Expect(actual).To(ConsistOf(0, 1, 2))
		})
	})

	Describe("#Offset", func() {
		var file = stringio.NewFromString("The first line.\nThe second line :)\n\n")
		var subject = NewScanner(file)
		var actual = []uint64{}

		for subject.ScanLine() {
			check(subject.Err())
			actual = append(actual, subject.Offset())
		}

		It("reads the file", func() {
			Expect(actual).To(Uint64ConsistOf(0, 16, 35))
		})
	})

	Describe("#Limit", func() {
		var file = stringio.NewFromString("The first line.\nThe second line :)\n\n")
		var subject = NewScanner(file)
		var actual = []uint32{}

		for subject.ScanLine() {
			check(subject.Err())
			actual = append(actual, subject.Limit())
		}

		It("reads the file", func() {
			Expect(actual).To(Uint32ConsistOf(16, 19, 1))
		})
	})

	Describe("#EachLine", func() {
		var file = stringio.NewFromString("a\nb\n\n")
		var subject = NewScanner(file)
		var actual = [][]byte{}

		subject.EachLine(func(line []byte, err error) {
			check(err)
			actual = append(actual, line)
		})

		It("reads the file", func() {
			Expect(actual).To(ConsistOf([]byte{'a'}, []byte{'b'}, []byte{}))
		})
	})

	Describe("#EachString", func() {
		var file = stringio.NewFromString("The first line.\nThe second line :)\n\n")
		var subject = NewScanner(file)
		var actual = []string{}

		subject.EachString(func(line string, err error) {
			check(err)
			actual = append(actual, line)
		})

		It("reads the file", func() {
			Expect(actual).To(ConsistOf("The first line.", "The second line :)", ""))
		})
	})

	Describe("#IsLineEmpty", func() {
		var file = stringio.NewFromString("The first line.\nThe second line :)\n\n")
		var subject = NewScanner(file)
		var actual = []bool{}

		for subject.ScanLine() {
			check(subject.Err())
			actual = append(actual, subject.IsLineEmpty())
		}

		It("reads the first line as bytes slice", func() {
			Expect(actual).To(ConsistOf(false, false, true))
		})
	})

	Describe("#Reset", func() {
		var file = stringio.NewFromString("line1.\nline2.\n")
		var subject = NewScanner(file)
		var actual = []string{}

		for subject.ScanLine() {
			check(subject.Err())
			actual = append(actual, subject.Text())
		}
		subject.Reset()
		for subject.ScanLine() {
			check(subject.Err())
			actual = append(actual, subject.Text())
		}

		It("allows to re-read the file", func() {
			Expect(actual).To(ConsistOf("line1.", "line2.", "line1.", "line2."))
		})
	})
})
