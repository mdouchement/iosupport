// Copyright 2015-2016 mdouchement. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package iosupport reads and writes large text files.
// It can count lines, bytes, characters and words contained in the give file
// and it can parse and sort CSV files.
//
// There are many kinds of CSV files; this package supports the format
// described in RFC 4180 except newlines in quoted fields.
//
// A CSV file contains zero or more records of one or more fields per record.
// Each record is separated by the newline character. The final record may
// optionally be followed by a newline character.
//	field1,field2,field3
//
// White space is considered part of a field.
//
// Carriage returns before newline characters are silently removed.
//
// Fields which start and stop with the quote character " are called
// quoted-fields. The beginning and ending quote are not part of the
// field.
//
// The source:
//
//	normal string,"quoted-field"
//
// results in the fields
//
//	{`normal string`, `quoted-field`}
//
// Within a quoted-field a quote character followed by a second quote
// character is considered a single quote.
//
//	"the ""word"" is true","a ""quoted-field"""
//
// results in
//
//	{`the "word" is true`, `a "quoted-field"`}
package iosupport
