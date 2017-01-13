package iosupport_test

import (
	"path/filepath"

	. "github.com/mdouchement/iosupport"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StorageService", func() {
	Describe("HDDStorageService", func() {
		var basepath = tempDir("", "tsv_swap_test")
		var subject = NewHDDStorageService(basepath)
		var key = "dump0-key1.chunk"
		var data = map[string]string{
			"alpha": "trololo",
			"beta":  "trololo42",
		}

		Describe("#Marshall", func() {
			It("serializes data without error", func() {
				Expect(subject.Marshal(key, data)).To(BeNil())
			})

			It("generates a file on file system", func() {
				Expect(fileExists(filepath.Join(basepath, "dump0", "key1.chunk", "dump0-key1.chunk"))).To(BeTrue())
			})
		})

		Describe("#Unmarshall", func() {
			It("does not return an error", func() {
				subject.Marshal(key, data)
				var actual map[string]string
				Expect(subject.Unmarshal(key, &actual)).To(BeNil())
			})

			It("retreives and deserializes data", func() {
				subject.Marshal(key, data)
				var actual map[string]string
				subject.Unmarshal(key, &actual)
				Expect(actual).To(Equal(data))
			})
		})
	})

	Describe("MemStorageService", func() {
		var basepath = tempDir("", "tsv_swap_test")
		var subject = NewMemStorageService()
		var key = "dump0-key1.chunk"
		var data = map[string]string{
			"alpha": "trololo",
			"beta":  "trololo42",
		}

		Describe("#Marshall", func() {
			It("serializes data without error", func() {
				Expect(subject.Marshal(key, data)).To(BeNil())
			})

			It("does not generate a file on file system", func() {
				Expect(fileExists(filepath.Join(basepath, "dump0", "key1.chunk", "dump0-key1.chunk"))).To(BeFalse())
			})
		})

		Describe("#Unmarshall", func() {
			It("does not return an error", func() {
				subject.Marshal(key, data)
				var actual map[string]string
				Expect(subject.Unmarshal(key, &actual)).To(BeNil())
			})

			It("retreives and deserializes data", func() {
				subject.Marshal(key, data)
				var actual map[string]string
				subject.Unmarshal(key, &actual)
				Expect(actual).To(Equal(data))
			})
		})
	})
})
