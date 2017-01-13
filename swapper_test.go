package iosupport_test

import (
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/mdouchement/iosupport"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Swapper", func() {
	Describe("NullSwapper", func() {
		mockCtrl := gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()

		var storage *MockStorageService
		var subject *Swapper
		BeforeEach(func() {
			subject = NewNullSwapper()
			storage = NewMockStorageService(mockCtrl)
			subject.Storage = storage
		})
		var lines = TsvLines{}

		Describe("#IsTimeToSwap", func() {
			It("always return false", func() {
				Expect(subject.IsTimeToSwap(TsvLines{})).To(BeFalse())
			})
		})

		Describe("#HasSwapped", func() {
			It("always return false", func() {
				Expect(subject.HasSwapped()).To(BeFalse())
			})
		})

		Describe("#Swap", func() {
			It("does not swap", func() {
				var k = "not received"
				storage.EXPECT().Marshal(gomock.Any(), gomock.Any()).Do(func(key string, _ interface{}) {
					k = key
				})
				subject.Swap(lines)

				Expect(k).To(Equal("not received"))
			})
		})

		Describe("#ReadIterator", func() {
			var lines = TsvLines{
				TsvLine{"0", 0, 0},
				TsvLine{"1", 1, 1},
				TsvLine{"2", 2, 2},
			}

			BeforeEach(func() {
				subject.KeepWithoutSwap(lines)
			})

			It("allow to iterate across the TsvLines", func() {
				it := subject.ReadIterator()
				check(it.Error())

				i := 0
				for it.Next() {
					check(it.Error())
					Expect(it.Value()).To(Equal(lines[i]))
					i++
				}
			})
		})
	})

	Describe("Swapper", func() {
		var backupGetMemoryUsage func() *HeapMemStat

		mockCtrl := gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()

		var limit uint64 = 800 << 20 // ~800MB
		var basepath = tempDir("", "tsv_swapper_test")
		var storage *MockStorageService
		var subject *Swapper
		BeforeEach(func() {
			backupGetMemoryUsage = GetMemoryUsage

			subject = NewSwapper(limit, basepath)
			storage = NewMockStorageService(mockCtrl)
			subject.Storage = storage
		})
		var lines = TsvLines{
			TsvLine{"0", 0, 0},
			TsvLine{"1", 1, 1},
			TsvLine{"2", 2, 2},
		}

		AfterEach(func() {
			GetMemoryUsage = backupGetMemoryUsage
		})

		Describe("#IsTimeToSwap", func() {
			Context("when memory limit is not reached", func() {
				BeforeEach(func() {
					GetMemoryUsage = func() *HeapMemStat {
						return &HeapMemStat{} // Zero values
					}
				})

				It("return false", func() {
					Expect(subject.IsTimeToSwap(lines)).To(BeFalse())
				})
			})

			Context("when memory limit is reached", func() {
				BeforeEach(func() {
					GetMemoryUsage = func() *HeapMemStat {
						return &HeapMemStat{0, limit + 10, 0, 0, 0, 0}
					}
				})

				It("returns true", func() {
					Expect(subject.IsTimeToSwap(lines)).To(BeTrue())
				})
			})
		})

		Describe("#HasSwapped", func() {
			Context("when there is no dump", func() {
				It("returns false", func() {
					Expect(subject.HasSwapped()).To(BeFalse())
				})
			})

			Context("when there is a dump", func() {
				BeforeEach(func() {
					storage.EXPECT().Marshal("0-0.chunk", lines).AnyTimes()
					subject.Swap(lines)
				})

				It("returns true", func() {
					Expect(subject.HasSwapped()).To(BeTrue())
				})
			})
		})

		Describe("#Swap", func() {
			It("swaps", func() {
				var k string
				var v interface{}
				storage.EXPECT().Marshal(gomock.Any(), gomock.Any()).Do(func(key string, value interface{}) {
					k = key
					v = value
				})
				subject.Swap(lines)

				Expect(k).To(Equal("0-0.chunk"))
				Expect(v).To(Equal(lines))
			})
		})

		Describe("#ReadIterator", func() {
			var lines = TsvLines{
				TsvLine{"0", 0, 0},
				TsvLine{"1", 1, 1},
				TsvLine{"2", 2, 2},
			}

			BeforeEach(func() {
				subject.KeepWithoutSwap(lines)
			})

			Context("when there is no dump", func() {
				It("allow to iterate across the TsvLines", func() {
					it := subject.ReadIterator()
					check(it.Error())

					i := 0
					for it.Next() {
						check(it.Error())
						Expect(it.Value()).To(Equal(lines[i]))
						i++
					}
				})
			})

			Context("when there is a dump", func() {
				BeforeEach(func() {
					storage.EXPECT().Marshal("0-0.chunk", lines).AnyTimes()
					subject.Swap(lines) // HasSwapped now returns true
					storage.EXPECT().Unmarshal("0-0.chunk", gomock.Any()).Do(func(_ string, value interface{}) {
						// Mock response
						vv, _ := value.(*TsvLines)
						for _, e := range lines {
							*vv = append(*vv, e)
						}
					})
				})

				It("allow to iterate across the TsvLines", func() {
					it := subject.ReadIterator()
					check(it.Error())

					i := 0
					for it.Next() {
						check(it.Error())
						Expect(it.Value()).To(Equal(lines[i]))
						i++
					}
				})

				Context("when an error occurred", func() {
					var err = errors.New("test_error")

					BeforeEach(func() {
						storage.EXPECT().Marshal(gomock.Any(), gomock.Any()).AnyTimes()
						subject.Swap(lines) // HasSwapped now returns true
						storage.EXPECT().Unmarshal(gomock.Any(), gomock.Any()).AnyTimes().Return(err)
					})

					It("returns an error through the iterators", func() {
						it := subject.ReadIterator()
						Expect(it.Error()).To(Equal(err))
					})
				})
			})
		})
	})
})
