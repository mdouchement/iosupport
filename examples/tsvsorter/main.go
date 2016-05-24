package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mdouchement/iosupport"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "TSV sorter"
	app.Version = "0.0.1"
	app.Author = "mdouchement"
	app.Usage = "Sort a TSV on a composite key"
	app.UsageText = `tsvsorter options

   Example:
     tsvsorter -i iris.csv -s=',' -H -f "Sepal Length" -o S_iris.csv

   IO usage:
     - Local FileSystem: /tmp/iris.csv
     - Hadoop FileSystem: hdfs://tmp/iris.csv
  `
	app.Flags = flags
	app.Action = action

	err := app.Run(os.Args)
	if err != nil {
		fail(err)
	}
}

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "f, fields",
		Usage: "Ordered list of columns name to be sorted (pattern: 'col5,col4')",
	},
	cli.StringFlag{
		Name:  "s, separator",
		Usage: "The TSV delimiter",
	},
	cli.BoolFlag{
		Name:  "H, header",
		Usage: "The TSV has an header",
	},
	cli.StringFlag{
		Name:  "i, input",
		Usage: "Input TSV path",
	},
	cli.StringFlag{
		Name:  "o, output",
		Usage: "Output TSV path",
	},
}

func action(context *cli.Context) error {
	inputPath := context.String("i")
	header := context.Bool("H")
	separator := context.String("s")
	fields := strings.Split(context.String("f"), ",")
	outputPath := context.String("o")

	start := time.Now()

	fmt.Println("Openning file...")
	fmt.Println("Scanner and indexer initialization")
	sc := func() *iosupport.Scanner {
		file, err := open(inputPath)
		if err != nil {
			fail(fmt.Errorf("Scanner: %v", err))
		}
		return iosupport.NewScanner(file)
	}
	indexer := iosupport.NewTsvIndexer(sc, header, separator, fields)
	defer indexer.CloseIO()

	elapsed := time.Since(start)
	fmt.Printf("Initialization took %s\n\n", elapsed)

	fmt.Println("Analyzing...")
	astart := time.Now()
	err := indexer.Analyze()
	if err != nil {
		fail(fmt.Errorf("Analyze: %v", err))
	}
	elapsed = time.Since(astart)
	fmt.Printf("Analyze took %s\n\n", elapsed)

	fmt.Println("Sorting...")
	sstart := time.Now()
	indexer.Sort()
	elapsed = time.Since(sstart)
	fmt.Printf("Sort took %s\n\n", elapsed)

	fmt.Println("Transfering...")
	ofile, err := create(outputPath)
	if err != nil {
		fail(err)
	}
	defer ofile.Close()
	tstart := time.Now()
	err = indexer.Transfer(ofile)
	if err != nil {
		fail(fmt.Errorf("Transfer: %v", err))
	}
	elapsed = time.Since(tstart)
	fmt.Printf("Transfer took %s\n\n", elapsed)

	elapsed = time.Since(start)
	fmt.Printf("Total time %s\n", elapsed)

	return nil
}

func fail(err error) {
	panic(err)
}
