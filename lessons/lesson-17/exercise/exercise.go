// Package exercise obsahuje cvičení lekce 17.
package exercise

import "io"

// Record je jeden řádek CSV s útratou.
type Record struct {
	Name     string
	Amount   float64
	Category string
}

// Median vrací medián hodnot a false pro prázdný vstup.
// Vstupní slice nemění.
func Median(nums []float64) (float64, bool) {
	// TODO: úkol A
	return 0, false
}

// ParseRecords načte CSV s hlavičkou "name,amount,category" a vrátí datové řádky.
func ParseRecords(r io.Reader) ([]Record, error) {
	// TODO: úkol B
	return nil, nil
}

// SumByCategory sečte částky podle kategorie.
func SumByCategory(recs []Record) map[string]float64 {
	// TODO: úkol B
	return nil
}

// TopN vrací n záznamů s nejvyšší částkou, při shodě v původním pořadí.
func TopN(recs []Record, n int) []Record {
	// TODO: úkol C
	return nil
}

// LoadFile načte záznamy ze souboru na dané cestě.
func LoadFile(path string) ([]Record, error) {
	// TODO: úkol C
	return nil, nil
}
