package iosupport

import "strings"

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

// CountWords counts all words in the given string.
func CountWords(str string) int {
	str = string(TrimNewline([]byte(str)))
	if len(str) == 0 {
		return 0
	}
	words := len(strings.Split(str, " "))
	words += len(strings.Split(str, string('\t'))) - 1 // -1 because string itself is aready counted
	return words
}

// A WordCount counts bytes, chars, words and lines from a given file.
type WordCount struct {
	s     *Scanner          // The scanner provided by the client.
	Opts  *WordCountOptions // Defines what is counted
	Bytes int               // Bytes counter
	Chars int               // Chars counter
	Words int               // Words counter
	Lines int               // Lines counter
}

// WordCountOptions lets you define which thing you want to count.
type WordCountOptions struct {
	CountByte bool
	CountChar bool
	CountWord bool
	CountLine bool
}

// NewWordCount instanciates a new WordCount for the given file.
func NewWordCount(f FileReader) *WordCount {
	return &WordCount{
		s:     NewScanner(f),
		Opts:  defaultWordCounterOptions(),
		Bytes: 0,
		Chars: 0,
		Words: 0,
		Lines: 0,
	}
}

// NewWordCountOptions instanciates a new WordCountOptions
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

// Perform starts the count
func (wc *WordCount) Perform() error {
	wc.s.KeepNewlineSequence(true)
	for wc.s.ScanLine() {
		if wc.s.Err() != nil {
			return wc.s.Err()
		}

		if wc.Opts.CountByte {
			wc.Bytes += len(wc.s.Bytes())
		}
		if wc.Opts.CountChar {
			wc.Chars += len([]rune(wc.s.Text()))
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
