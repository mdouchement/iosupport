package iosupport_test

import (
	. "github.com/mdouchement/iosupport"
	"github.com/mdouchement/stringio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("WC", func() {
	Describe("#Perform", func() {
		var file = stringio.NewFromString("The first line.\nThe secönd line :)\n\n")

		Context("with default options", func() {
			var subject = NewWordCount(file)

			file.Seek(0, 0)
			err := subject.Perform()
			check(err)

			It("does not count bytes", func() {
				Expect(subject.Bytes).To(Equal(0))
			})

			It("counts characters", func() {
				Expect(subject.Chars).To(Equal(36))
			})

			It("counts words", func() {
				Expect(subject.Words).To(Equal(7))
			})

			It("counts lines", func() {
				Expect(subject.Lines).To(Equal(3))
			})
		})

		Context("with custom options", func() {
			var subject = NewWordCount(file)

			file.Seek(0, 0)
			opts := NewWordCountOptions()
			opts.CountByte = true
			subject.Opts = opts
			err := subject.Perform()
			check(err)

			It("counts bytes", func() {
				Expect(subject.Bytes).To(Equal(37))
			})

			It("does not count characters", func() {
				Expect(subject.Chars).To(Equal(0))
			})

			It("does not count words", func() {
				Expect(subject.Words).To(Equal(0))
			})

			It("does not count lines", func() {
				Expect(subject.Words).To(Equal(0))
			})
		})
	})
})
