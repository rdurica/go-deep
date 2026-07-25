// Příkaz csvstats počítá statistiky nad CSV se sloupci name,amount,category.
//
// Použití:
//
//	csvstats -file data.csv
//	cat data.csv | csvstats -top 3
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/rdurica/go-deep/projects/p01-csv-cli/csvstats"
)

const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run je testovatelné jádro příkazu: žádné globální stavy, návratem je exit kód.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("csvstats", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "csvstats — statistiky nad CSV (name,amount,category)")
		fmt.Fprintln(stderr, "\nPoužití: csvstats [-file cesta] [-top N]")
		fs.PrintDefaults()
	}

	file := fs.String("file", "", "cesta k CSV souboru (prázdné = čti ze stdin)")
	top := fs.Int("top", 0, "vypiš navíc N největších útrat")

	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "csvstats: nečekaný argument %q\n", fs.Arg(0))
		fs.Usage()
		return exitUsage
	}
	if *top < 0 {
		fmt.Fprintln(stderr, "csvstats: -top nesmí být záporné")
		return exitUsage
	}

	records, err := load(*file, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "csvstats: %v\n", err)
		if errors.Is(err, csvstats.ErrEmptyInput) {
			fmt.Fprintln(stderr, "nápověda: první řádek musí být name,amount,category")
		}
		return exitFailure
	}

	if err := csvstats.RenderSummary(stdout, csvstats.Summarize(records)); err != nil {
		fmt.Fprintf(stderr, "csvstats: %v\n", err)
		return exitFailure
	}
	if *top > 0 {
		fmt.Fprintln(stdout)
		if err := csvstats.RenderTop(stdout, csvstats.TopN(records, *top)); err != nil {
			fmt.Fprintf(stderr, "csvstats: %v\n", err)
			return exitFailure
		}
	}
	return exitOK
}

func load(path string, stdin io.Reader) ([]csvstats.Record, error) {
	if path == "" {
		return csvstats.ParseRecords(stdin)
	}
	return csvstats.LoadFile(path)
}
