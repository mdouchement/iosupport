package iosupport_test

import (
	"os"
	"testing"

	"github.com/mdouchement/iosupport"
)

var wordCountInput = "The first line.\nThe secönd line :)\n\n"

func TestWordCountPerform(t *testing.T) {
	path := generateTmpFile(wordCountInput)
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	wc := iosupport.NewWordCount(file)
	err = wc.Perform()
	check(err)

	if wc.Bytes != 0 {
		// Bytes counter not actived
		t.Errorf("Invalid number of bytes. Expected 0 but got %v", wc.Bytes)
	}

	if wc.Chars != 36 {
		t.Errorf("Invalid number of chars. Expected 36 but got %v", wc.Chars)
	}

	if wc.Words != 7 {
		t.Errorf("Invalid number of words. Expected 7 but got %v", wc.Words)
	}

	if wc.Lines != 3 {
		t.Errorf("Invalid number of lines. Expected 3 but got %v", wc.Lines)
	}
}

func TestWordCountWithCustomOptions(t *testing.T) {
	path := generateTmpFile(wordCountInput)
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	wc := iosupport.NewWordCount(file)
	o := iosupport.NewWordCountOptions()
	o.CountByte = true
	wc.Opts = o
	err = wc.Perform()
	check(err)

	if wc.Bytes != 37 {
		t.Errorf("Invalid number of bytes. Expected 37 but got %v", wc.Bytes)
	}

	if wc.Chars != 0 {
		// Chars counter not actived
		t.Errorf("Invalid number of chars. Expected 0 but got %v", wc.Chars)
	}

	if wc.Words != 0 {
		// Words counter not actived
		t.Errorf("Invalid number of words. Expected 0 but got %v", wc.Words)
	}

	if wc.Lines != 0 {
		// Lines counter not actived
		t.Errorf("Invalid number of lines. Expected 0 but got %v", wc.Lines)
	}
}
