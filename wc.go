package iosupport

import (
	"os"
	"strings"
)

// Provide tool like GNU wc command
// Usage:
//
//  file, _ := os.Open("my_file.txt")
//  wc := iosupport.NewWordCount(file)
//  wc.Perform()
//  println(wc.Bytes)
//  println(wc.Chars)
//  println(wc.Words)
//  println(wc.Lines)
//
//  - With options
//
//  wc := iosupport.NewWordCount(file)
//  opts := iosupport.NewWordCountOptions()
//  opts.CountByte = true
//  opts.CountChar = true
//  opts.CountWord = true
//  opts.CountLine = true
//  wc.Opts = opts
//  wc.Perform()

type WordCount struct {
	s     *Scanner          // The scanner provided by the client.
	Opts  *WordCountOptions // Defines what is counted
	Bytes int               // Bytes counter
	Chars int               // Chars counter
	Words int               // Words counter
	Lines int               // Lines counter
	err   error             // Sticky error.
}

type WordCountOptions struct {
	CountByte bool
	CountChar bool
	CountWord bool
	CountLine bool
}

func NewWordCount(f *os.File) *WordCount {
	return &WordCount{
		s:     NewScanner(f),
		Opts:  defaultWordCounterOptions(),
		Bytes: 0,
		Chars: 0,
		Words: 0,
		Lines: 0,
	}
}

func NewWordCountOptions() *WordCountOptions {
	return new(WordCountOptions)
}

func defaultWordCounterOptions() *WordCountOptions {
	return &WordCountOptions{
		CountByte: false,
		CountChar: true,
		CountWord: true,
		CountLine: true,
	}
}

func (wc *WordCount) Perform() error {
	for wc.s.ScanLine() {
		if wc.s.Err() != nil {
			return wc.s.Err()
		}

		if wc.Opts.CountByte {
			// +1 for EOL character
			wc.Bytes += len(wc.s.Bytes()) + 1
		}
		if wc.Opts.CountChar {
			// +1 for EOL character
			wc.Chars += len(wc.s.Text()) + 1
		}
		if wc.Opts.CountWord {
			wc.Words += CountWords(wc.s.Text())
		}
		if wc.Opts.CountLine {
			wc.Lines++
		}
	}
	return nil
}

func CountWords(str string) int {
	if len(str) == 0 {
		return 0
	}
	words := len(strings.Split(str, " "))
	words += len(strings.Split(str, string('\t'))) - 1 // -1 because string itself is aready counted
	return words
}
