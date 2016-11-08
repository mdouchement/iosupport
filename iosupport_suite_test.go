package iosupport_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/mdouchement/iosupport"
	"github.com/mdouchement/stringio"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"

	"testing"
)

func TestIosupport(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Iosupport Suite")
}

func check(err error) {
	if err != nil {
		Fail(err.Error())
	}
}

func toStringSlice(values [][]byte) []string {
	var sl []string
	for _, value := range values {
		sl = append(sl, string(value))
	}
	return sl
}

func scanner(data string) func() *iosupport.Scanner {
	return func() *iosupport.Scanner {
		return iosupport.NewScanner(stringio.NewFromString(data))
	}
}

func cs(cols ...string) string {
	var buf bytes.Buffer
	for _, col := range cols {
		buf.WriteString(col)
		buf.WriteString(iosupport.COMPARABLE_SEPARATOR)
	}
	return buf.String()
}

// ------------------ //
// Custom matchers    //
// ------------------ //

// == Uint64ConsistOf ==

type uint64ConsistOf struct {
	Expected []uint64
}

func Uint64ConsistOf(expected ...uint64) types.GomegaMatcher {
	return &uint64ConsistOf{
		Expected: expected,
	}
}

func (matcher *uint64ConsistOf) Match(actual interface{}) (success bool, err error) {
	v, ok := actual.([]uint64)
	if !ok {
		return false, fmt.Errorf("Expected %s type but got %#v", reflect.TypeOf(matcher.Expected), actual)
	}

	return reflect.DeepEqual(v, matcher.Expected), nil
}

func (matcher *uint64ConsistOf) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to equal", matcher.Expected)
}

func (matcher *uint64ConsistOf) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to equal", matcher.Expected)
}

// == Uint32ConsistOf ==

type uint32ConsistOf struct {
	Expected []uint32
}

func Uint32ConsistOf(expected ...uint32) types.GomegaMatcher {
	return &uint32ConsistOf{
		Expected: expected,
	}
}

func (matcher *uint32ConsistOf) Match(actual interface{}) (success bool, err error) {
	v, ok := actual.([]uint32)
	if !ok {
		return false, fmt.Errorf("Expected %s type but got %#v", reflect.TypeOf(matcher.Expected), actual)
	}

	return reflect.DeepEqual(v, matcher.Expected), nil
}

func (matcher *uint32ConsistOf) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to equal", matcher.Expected)
}

func (matcher *uint32ConsistOf) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to equal", matcher.Expected)
}

// == TlConsistOf ==

type tl iosupport.TsvLine

type tlConsistOf struct {
	Expected []tl
}

func TlConsistOf(expected ...tl) types.GomegaMatcher {
	return &tlConsistOf{
		Expected: expected,
	}
}

func (matcher *tlConsistOf) Match(actual interface{}) (success bool, err error) {
	raw, err := json.Marshal(actual)
	if err != nil {
		return false, fmt.Errorf("TlConsistOf marshalling: %s", err)
	}

	var v []tl
	if err = json.Unmarshal(raw, &v); err != nil {
		return false, fmt.Errorf("TlConsistOf unmarshalling: %s", err)
	}

	return reflect.DeepEqual(v, matcher.Expected), nil
}

func (matcher *tlConsistOf) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to equal", matcher.Expected)
}

func (matcher *tlConsistOf) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to equal", matcher.Expected)
}
