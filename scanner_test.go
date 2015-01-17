package iosupport_test

import(
  "os"
  "strings"
  "bytes"
  "testing"

  "github.com/mdouchement/iosupport"
)

var data string = "The first line.\nThe second line :)\n\n"

func TestScannerScanLine(t *testing.T) {
  path := generateTmpFile()
  file, err := os.Open(path)
  check(err)
  defer file.Close()

  sc := iosupport.NewScanner(file)

  for i, expected := range([]bool{true, true, true, false}) {
    actual := sc.ScanLine()
    if actual != expected {
      t.Errorf("[Pass %v]: Expected '%v' but got '%v'", i+1, expected, actual)
    }
  }
}

func TestScannerBytes(t *testing.T) {
  path := generateTmpFile()
  file, err := os.Open(path)
  check(err)
  defer file.Close()

  sc := iosupport.NewScanner(file)
  lines := strings.Split(data, "\n")

  for _, bts := range([]string(lines)) {
    sc.ScanLine()
    expected := []byte(bts)
    actual := sc.Bytes()
    if !bytes.Equal(actual, expected) {
      t.Errorf("Expected '%v' but got '%v'", expected, actual)
    }
  }
}

func TestScannerText(t *testing.T) {
  path := generateTmpFile()
  file, err := os.Open(path)
  check(err)
  defer file.Close()

  sc := iosupport.NewScanner(file)
  lines := strings.Split(data, "\n")

  for _, expected := range([]string(lines)) {
    sc.ScanLine()
    actual := sc.Text()
    if actual != expected {
      t.Errorf("Expected '%v' but got '%v'", expected, actual)
    }
  }
}

func TestScannerEachLine(t *testing.T) {
  path := generateTmpFile()
  file, err := os.Open(path)
  check(err)
  defer file.Close()

  sc := iosupport.NewScanner(file)
  lines := strings.Split(data, "\n")
  i := 0

  sc.EachLine(func(actual []byte, err error) {
    expected := []byte(lines[i])
    if !bytes.Equal(actual, expected) {
      t.Errorf("Expected '%v' but got '%v'", expected, actual)
    }
    i++
  })
}

func TestScannerEachString(t *testing.T) {
  path := generateTmpFile()
  file, err := os.Open(path)
  check(err)
  defer file.Close()

  sc := iosupport.NewScanner(file)
  expected := strings.Split(data, "\n")
  i := 0

  sc.EachString(func(actual string, err error) {
    if actual != expected[i] {
      t.Errorf("Expected '%v' but got '%v'", expected[i], actual)
    }
    i++
  })
}

// Re-read file
func TestScannerReset(t *testing.T) {
  path := generateTmpFile()
  file, err := os.Open(path)
  check(err)
  defer file.Close()

  sc := iosupport.NewScanner(file)
  for i := 1; i < 2; i++ {
    for j, expected := range([]bool{true, true, true, false}) {
      actual := sc.ScanLine()
      if actual != expected {
        t.Errorf("[Pass %v:%v]: Expected '%v' but got '%v'", i+1, j+1, expected, actual)
      }
    }
    sc.Reset()
  }
}

func TestScannerIsLineEmpty(t *testing.T) {
  path := generateTmpFile()
  file, err := os.Open(path)
  check(err)
  defer file.Close()

  sc := iosupport.NewScanner(file)

  for i, expected := range([]bool{false, false, true, true}) {
    sc.ScanLine()
    actual := sc.IsLineEmpty()
    if actual != expected {
      t.Errorf("[Pass %v]: Expected '%v' but got '%v'", i+1, expected, actual)
    }
  }
}

func generateTmpFile() string {
  path := "/tmp/iosupport_test.txt"
  file, err := os.Create(path)
  check(err)
  defer file.Close()
  _, err = file.WriteString(data)
  check(err)
  err = file.Sync()
  check(err)

  return path
}

func check(err error) {
  if err != nil {
    panic(err)
  }
}
