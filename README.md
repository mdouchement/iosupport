# iosupport
[![Build Status](https://travis-ci.org/mdouchement/iosupport.svg?branch=master)](https://travis-ci.org/mdouchement/iosupport)

It provides some io supports for GoLang.

## Usage

In order to start, go get this repository:

```bash
$ go get github.com/mdouchement/iosupport
```

### Example

```go
package main

import(
  "os"

  "github.com/mdouchement/iosupport"
)

func main() {
  file, _ := os.Open("my_file.txt")

  // See scanner.go for more examples
  sc := iosupport.NewScanner(file)
  sc.EachString(func(line string, err error) {
    if err != nil {
      panic(err.Error())
    }
    println(line)
  }

  // See wc.go for more examples
  wc := iosupport.NewWordCount(file)
  wc.Perform()
  println(wc.Chars)
  println(wc.Words)
  println(wc.Lines)

  // See tsv_indexer.go for more examples
  indexer = iosupport.NewTsvIndexer(sc, true, ",", []string{"col2", "col1"}) // scanner, headerIsPresent, separator, fieldsForSorting
  indexer.Analyze() // creates lines index
  indexer.Sort() // sorts indexed lines
  indexer.transfer() // transfers the input TSV in sorted output TSV
}
```
