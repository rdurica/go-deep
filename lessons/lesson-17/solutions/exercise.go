// Package solutions obsahuje referenční řešení lekce 17.
package solutions

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

// Record je jeden řádek CSV s útratou.
type Record struct {
	Name     string
	Amount   float64
	Category string
}

var wantHeader = []string{"name", "amount", "category"}

// --- Stupeň: jednoduchý ---
// Median vrací medián hodnot a false pro prázdný vstup.
// Vstupní slice nemění.
func Median(nums []float64) (float64, bool) {
	if len(nums) == 0 {
		return 0, false
	}
	sorted := slices.Clone(nums)
	slices.Sort(sorted)

	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid], true
	}
	return (sorted[mid-1] + sorted[mid]) / 2, true
}

// --- Stupeň: střední ---
// ParseRecords načte CSV s hlavičkou "name,amount,category" a vrátí datové řádky.
func ParseRecords(r io.Reader) ([]Record, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = len(wantHeader)

	header, err := cr.Read()
	if err == io.EOF {
		return nil, errors.New("csv: prázdný vstup, chybí hlavička")
	}
	if err != nil {
		return nil, fmt.Errorf("csv: hlavička: %w", err)
	}
	for i, want := range wantHeader {
		if !strings.EqualFold(strings.TrimSpace(header[i]), want) {
			return nil, fmt.Errorf("csv: hlavička %v, chci %v", header, wantHeader)
		}
	}

	records := []Record{}
	for line := 2; ; line++ {
		row, err := cr.Read()
		if err == io.EOF {
			return records, nil
		}
		if err != nil {
			return nil, fmt.Errorf("csv: řádek %d: %w", line, err)
		}

		rec, err := parseRow(row)
		if err != nil {
			return nil, fmt.Errorf("csv: řádek %d: %w", line, err)
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
		return Record{}, fmt.Errorf("částka %q: %w", row[1], err)
	}
	return Record{Name: name, Amount: amount, Category: category}, nil
}

// SumByCategory sečte částky podle kategorie.
func SumByCategory(recs []Record) map[string]float64 {
	sums := make(map[string]float64, len(recs))
	for _, rec := range recs {
		sums[rec.Category] += rec.Amount
	}
	return sums
}

// --- Stupeň: obtížný ---
// TopN vrací n záznamů s nejvyšší částkou, při shodě v původním pořadí.
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

// LoadFile načte záznamy ze souboru na dané cestě.
func LoadFile(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", path, err)
	}
	defer f.Close()

	recs, err := ParseRecords(f)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", path, err)
	}
	return recs, nil
}
