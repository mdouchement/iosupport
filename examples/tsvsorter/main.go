package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/labstack/gommon/bytes"
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
		panic(err)
	}
}

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "m, memory-limit",
		Usage: "Memory limit (e.g. 8GB or 8G)",
	},
	cli.StringFlag{
		Name:  "f, fields",
		Usage: "Ordered list of columns name to be sorted (pattern: 'col5,col4')",
	},
	cli.StringFlag{
		Name:  "s, separator",
		Usage: "Dataset delimiter",
	},
	cli.BoolFlag{
		Name:  "H, header",
		Usage: "Dataset has an header",
	},
	cli.StringFlag{
		Name:  "u, hadoop-user",
		Usage: "Hadoop user (can use HADOOP_USER_NAME environment variable)",
	},
	cli.StringFlag{
		Name:  "i, input",
		Usage: "Input Dataset path",
	},
	cli.StringFlag{
		Name:  "o, output",
		Usage: "Output Dataset path",
	},
}

func action(context *cli.Context) error {
	memory := context.String("m")
	inputPath := context.String("i")
	header := context.Bool("H")
	separator := context.String("s")
	fields := strings.Split(context.String("f"), ",")
	outputPath := context.String("o")

	if inputPath == "" || separator == "" || context.String("f") == "" || outputPath == "" {
		defer cli.ShowAppHelp(context)
		panic(fmt.Errorf("Invalid command line"))
	}

	if user := context.String("u"); user != "" {
		os.Setenv("HADOOP_USER_NAME", user)
	}

	var limit uint64 = 4 << 30 // ~4GB
	if memory != "" {
		l, err := bytes.Parse(memory)
		if err != nil {
			panic(err)
		}
		limit = uint64(l)
	}
	fmt.Println("Memory limit:", memory)

	start := time.Now()

	fmt.Println("Openning file...")
	fmt.Println("Scanner and indexer initialization")
	sc := func() *iosupport.Scanner {
		file, err := open(inputPath)
		if err != nil {
			panic(fmt.Errorf("Scanner: %v", err))
		}
		return iosupport.NewScanner(file)
	}
	indexer := iosupport.NewTsvIndexer(sc,
		iosupport.Header(header),
		iosupport.Separator(separator),
		iosupport.Fields(fields...),
		iosupport.SkipMalformattedLines(),
		iosupport.DropEmptyIndexedFields(),
		iosupport.SwapperOpts(limit, fmt.Sprintf("/tmp/tsv_swap_%d", time.Now().Nanosecond())))
	defer indexer.CloseIO()

	elapsed := time.Since(start)
	fmt.Printf("Initialization took %s\n\n", elapsed)

	fmt.Println("Analyzing...")
	astart := time.Now()
	err := indexer.Analyze()
	if err != nil {
		panic(fmt.Errorf("Analyze: %v", err))
	}
	elapsed = time.Since(astart)
	fmt.Printf("Analyze took %s\n\n", elapsed)

	fmt.Println("Sorting...")
	sstart := time.Now()
	indexer.Sort()
	elapsed = time.Since(sstart)
	fmt.Printf("Sort took %s\n\n", elapsed)

	fmt.Println("Transferring...")
	ofile, err := create(outputPath)
	if err != nil {
		panic(err)
	}
	defer ofile.Close()
	tstart := time.Now()
	err = indexer.Transfer(ofile)
	if err != nil {
		panic(fmt.Errorf("Transfer: %v", err))
	}
	elapsed = time.Since(tstart)
	fmt.Printf("Transfer took %s\n\n", elapsed)

	elapsed = time.Since(start)
	fmt.Printf("Total time %s\n", elapsed)

	return nil
}
