# iosupport
[![Build Status](https://travis-ci.org/mdouchement/iosupport.svg?branch=master)](https://travis-ci.org/mdouchement/iosupport)

It provides some io supports for GoLang.

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
  indexer = iosupport.NewTsvIndexer(sc, true, ",", []string{"col2", "col1"}) // scanner, headerIsPresent, separator, fieldsForSorting
  indexer.LineThreshold = 25000 // Number of lines between each scanner's seek (see TsvIndexer#selectSeeker)
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
