// Package exercise obsahuje cvičení lekce 17.
package exercise

import "io"

// Record je jeden řádek CSV s útratou.
type Record struct {
	Name     string
	Amount   float64
	Category string
}

// --- Stupeň: jednoduchý ---
// Median vrací medián hodnot a false pro prázdný vstup.
// Lichý počet: prostřední prvek seřazené posloupnosti; sudý: průměr dvou prostředních.
// Vstupní slice nemění — pracuj nad kopií (slices.Clone).
func Median(nums []float64) (float64, bool) {
	// TODO
	return 0, false
}

// --- Stupeň: střední ---
// ParseRecords načte CSV s hlavičkou "name,amount,category" a vrátí datové řádky.
// Hlavička case-insensitive, bílé znaky ořízni; špatná hlavička nebo úplně prázdný vstup jsou chyba.
// Datové řádky mají přesně 3 sloupce; jméno a kategorie nesmí být prázdné po oříznutí, částka musí být číslo.
// Chyba obsahuje číslo řádku (hlavička je 1); vstup jen s hlavičkou → prázdný výsledek.
func ParseRecords(r io.Reader) ([]Record, error) {
	// TODO
	return nil, nil
}

// SumByCategory sečte částky podle kategorie.
// Prázdný vstup vrací prázdnou, ale nenilovou mapu.
func SumByCategory(recs []Record) map[string]float64 {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---
// TopN vrací n záznamů s nejvyšší částkou, při shodě v původním pořadí.
// n <= 0 → prázdný výsledek; n > len(recs) → všechny záznamy. Vstupní slice nemění.
func TopN(recs []Record, n int) []Record {
	// TODO
	return nil
}

// LoadFile načte záznamy ze souboru na dané cestě.
// Soubor zavři přes defer, deleguj na ParseRecords; chyby obal %w a doplň cestu.
func LoadFile(path string) ([]Record, error) {
	// TODO
	return nil, nil
}
