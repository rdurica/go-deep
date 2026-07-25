// Package csvstats parsuje CSV s útratami a počítá nad nimi statistiky.
//
// Balíček nic nevypisuje na standardní výstup a nečte argumenty příkazové řádky —
// to je práce pro package main v cmd/csvstats.
package csvstats

import (
	"cmp"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

// Record je jedna útrata načtená z CSV.
type Record struct {
	Name     string
	Amount   float64
	Category string
}

// CategoryStat shrnuje jednu kategorii.
type CategoryStat struct {
	Category string
	Count    int
	Total    float64
	Average  float64
}

// Summary je agregovaný pohled na všechny záznamy.
// Categories jsou seřazené sestupně podle Total, při shodě abecedně.
type Summary struct {
	Records    int
	Total      float64
	Categories []CategoryStat
}

// Header je povinná hlavička vstupního CSV.
var Header = []string{"name", "amount", "category"}

// ErrEmptyInput signalizuje vstup bez hlavičky.
var ErrEmptyInput = errors.New("csvstats: prázdný vstup, chybí hlavička")

// ParseRecords načte CSV s hlavičkou "name,amount,category".
// Vrací chybu s číslem řádku, pokud je řádek neúplný nebo částka není číslo.
func ParseRecords(r io.Reader) ([]Record, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = len(Header)

	header, err := cr.Read()
	if errors.Is(err, io.EOF) {
		return nil, ErrEmptyInput
	}
	if err != nil {
		return nil, fmt.Errorf("csvstats: hlavička: %w", err)
	}
	for i, want := range Header {
		if !strings.EqualFold(strings.TrimSpace(header[i]), want) {
			return nil, fmt.Errorf("csvstats: hlavička %v, chci %v", header, Header)
		}
	}

	records := []Record{}
	for line := 2; ; line++ {
		row, err := cr.Read()
		if errors.Is(err, io.EOF) {
			return records, nil
		}
		if err != nil {
			return nil, fmt.Errorf("csvstats: řádek %d: %w", line, err)
		}

		rec, err := parseRow(row)
		if err != nil {
			return nil, fmt.Errorf("csvstats: řádek %d: %w", line, err)
		}
		records = append(records, rec)
	}
}

func parseRow(row []string) (Record, error) {
	name := strings.TrimSpace(row[0])
	if name == "" {
		return Record{}, errors.New("prázdné jméno")
	}
	category := strings.TrimSpace(row[2])
	if category == "" {
		return Record{}, errors.New("prázdná kategorie")
	}
	amount, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
	if err != nil {
		return Record{}, fmt.Errorf("částka %q není číslo", row[1])
	}
	return Record{Name: name, Amount: amount, Category: category}, nil
}

// LoadFile načte záznamy ze souboru na dané cestě.
func LoadFile(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("csvstats: %w", err)
	}
	defer f.Close()

	recs, err := ParseRecords(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return recs, nil
}

// Summarize spočítá souhrn přes kategorie.
func Summarize(recs []Record) Summary {
	sums := make(map[string]*CategoryStat, len(recs))
	s := Summary{Records: len(recs), Categories: []CategoryStat{}}

	for _, rec := range recs {
		s.Total += rec.Amount
		stat, ok := sums[rec.Category]
		if !ok {
			stat = &CategoryStat{Category: rec.Category}
			sums[rec.Category] = stat
		}
		stat.Count++
		stat.Total += rec.Amount
	}

	for _, stat := range sums {
		stat.Average = stat.Total / float64(stat.Count)
		s.Categories = append(s.Categories, *stat)
	}
	slices.SortFunc(s.Categories, func(a, b CategoryStat) int {
		if c := cmp.Compare(b.Total, a.Total); c != 0 {
			return c
		}
		return cmp.Compare(a.Category, b.Category)
	})
	return s
}

// TopN vrací n záznamů s nejvyšší částkou, při shodě v původním pořadí.
// Vstupní slice nemění.
func TopN(recs []Record, n int) []Record {
	if n <= 0 || len(recs) == 0 {
		return []Record{}
	}
	sorted := slices.Clone(recs)
	slices.SortStableFunc(sorted, func(a, b Record) int {
		return cmp.Compare(b.Amount, a.Amount)
	})
	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}
