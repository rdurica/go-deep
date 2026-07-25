// Package exercise obsahuje cvičení lekce 52.
package exercise

import "errors"

// ErrFormat označuje vstup, který neodpovídá formátu záznamů.
var ErrFormat = errors.New("neplatný formát záznamu")

// Record je jeden záznam textového formátu.
type Record struct {
	ID    string
	Name  string
	Score int
}

// Normalize ořízne okrajové bílé znaky, sjednotí vnitřní bílé znaky
// na jednu mezeru a převede text na malá písmena.
func Normalize(s string) string {
	panic("TODO: úkol A")
}

// Encode zapíše záznamy do textového formátu "id|name|score" po řádcích.
func Encode(recs []Record) string {
	panic("TODO: úkol B")
}

// Decode přečte formát vyrobený funkcí Encode.
func Decode(s string) ([]Record, error) {
	panic("TODO: úkol B")
}

// RenderTable vykreslí záznamy jako zarovnanou textovou tabulku.
func RenderTable(recs []Record) string {
	panic("TODO: úkol C")
}

// RenderTableFast vrací totéž co RenderTable, ale staví výstup
// v jediném bufferu s předalokací.
func RenderTableFast(recs []Record) string {
	panic("TODO: úkol C")
}
