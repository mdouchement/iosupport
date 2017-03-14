package iosupport_test

import (
	"time"

	. "github.com/mdouchement/iosupport"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
)

var _ = Describe("MemoryService", func() {
	Describe(".GetMemoryUsage", func() {
		GetMemoryUsage()                   // init
		time.Sleep(500 * time.Millisecond) // init

		It("returns memory usage", func() {
			Expect(*GetMemoryUsage()).To(gstruct.MatchAllFields(gstruct.Fields{
				"TimeMsAgo":      BeNumerically("<", 0),
				"SysKb":          BeNumerically(">", 0),
				"HeapSysKb":      BeNumerically(">", 0),
				"HeapAllocKb":    BeNumerically(">", 0),
				"HeapIdleKb":     BeNumerically(">", 0),
				"HeapReleasedKb": BeNumerically(">=", 0),
			}))
		})
	})
})
