# iosupport
[![CircleCI](https://circleci.com/gh/mdouchement/iosupport.svg?style=shield)](https://circleci.com/gh/mdouchement/iosupport)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/mdouchement/iosupport)
[![Go Report Card](https://goreportcard.com/badge/github.com/mdouchement/iosupport)](https://goreportcard.com/report/github.com/mdouchement/iosupport)
[![License](https://img.shields.io/github/license/mdouchement/iosupport.svg)](http://opensource.org/licenses/MIT)

It provides some io supports for GoLang:
- Read large files (line length and large amount of lines)
- Parse CSV files according the RFC4180, but:
  - It does not support `\r\n` in quoted field
  - It does not support comment
- Sort CSV on one or several columns

## Usage

In order to start, go get this repository:

```bash
$ go get github.com/mdouchement/iosupport
```

### Example

- Scanner & wc

```go
package main

import(
  "os"

  "github.com/mdouchement/iosupport"
)

func main() {
  // With local filesystem
  file, _ := os.Open("my_file.txt")
  defer file.Close()

  // Or with HDFS "github.com/colinmarc/hdfs"
  // client, _ := hdfs.New("localhost:9000")
  // file, _ := client.Open("/iris.csv")

  // See scanner.go for more examples
  sc := iosupport.NewScanner(file)
  sc.EachString(func(line string, err error) {
    check(err)
    println(line)
  })

  // See wc.go for more examples
  wc := iosupport.NewWordCount(file)
  wc.Perform()
  println(wc.Chars)
  println(wc.Words)
  println(wc.Lines)
}

func check(err error) {
  if err != nil {
    panic(err)
  }
}
```

- TSV sort

```go
package main

import(
  "os"

  "github.com/mdouchement/iosupport"
)

func main() {
  sc := func() *iosupport.Scanner {
    file, _ := os.Open("iris.csv")
    // Or with HDFS "github.com/colinmarc/hdfs"
    // client, _ := hdfs.New("localhost:9000")
    // file, _ := client.Open("/iris.csv")
    return iosupport.NewScanner(file)
  }

  // See tsv_indexer.go for more examples
  indexer = iosupport.NewTsvIndexer(sc, iosupport.HasHeader(), iosupport.Separator(","), iosupport.Fields("col2", "col1")) // scanner, headerIsPresent, separator, fieldsForSorting
  defer indexer.CloseIO()
  err := indexer.Analyze() // creates lines index
  check(err)
  indexer.Sort() // sorts indexed lines
  ofile, _ := os.Open("my_sorted.tsv")
  defer ofile.Close()
  indexer.Transfer(ofile) // transfers the input TSV in sorted output TSV
}

func check(err error) {
  if err != nil {
    panic(err)
  }
}
```

## Tests

- Installation

```sh
$ go get github.com/onsi/ginkgo/ginkgo
$ go get github.com/onsi/gomega
$ go get github.com/golang/mock/gomock
```

_ Run tests

```sh
# One shot
$ ginko

# With watch
$ ginkgo watch
```

- Generate package test file

```sh
$ ginkgo bootstrap # set up a new ginkgo suite
$ ginkgo generate my_file.go # will create a sample test file.  edit this file and add your tests then...
```

- Benchmarks

```sh
$ go test -run=NONE -bench=ParseFields
```

- Generate mocks

```sh
# go get github.com/golang/mock/mockgen
$ mockgen -package=iosupport_test -source=storage_service.go -destination=storage_service_mock_test.go
```

## License

**MIT**

## Contributing

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request
