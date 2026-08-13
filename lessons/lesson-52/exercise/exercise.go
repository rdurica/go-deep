// Package exercise obsahuje cvičení lekce 52.
package exercise

import (
	"errors"
	"strings"
)

// ErrFormat označuje vstup, který neodpovídá formátu záznamů.
var ErrFormat = errors.New("neplatný formát záznamu")

// Record je jeden záznam textového formátu.
type Record struct {
	ID    string
	Name  string
	Score int
}

// --- Stupeň: jednoduchý ---
// Normalize ořízne okraje, vnitřní bílé znaky (unicode.IsSpace) sjednotí na jednu mezeru
// a převede na malá písmena. Už normalizovaný vstup vrací beze změny a bez alokace.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Chybí převod na malá písmena.
// Najdi chybu a oprav — testy před opravou padají.
func Normalize(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return strings.Join(fields, " ")
}

// --- Stupeň: střední ---
// Encode zapíše záznamy jako id|name|score po řádcích bez koncového LF.
// ID a Name escapují \\, \|, \n, \r; prázdný vstup dá prázdný řetězec.
func Encode(recs []Record) string {
	// TODO
	return ""
}

// Decode je inverze Encode; prázdný vstup vrací prázdný nenilový slice.
// Chyby obalují ErrFormat; nesmí panikovat na žádném vstupu.
func Decode(s string) ([]Record, error) {
	// TODO
	return nil, nil
}

// --- Stupeň: obtížný ---
// RenderTable vykreslí report o pevné šířce sloupců (8|20|5); řádky se neořezávají.
// Výstup ověřuje golden test proti testdata/table.golden.
func RenderTable(recs []Record) string {
	// TODO
	return ""
}

// RenderTableFast vrací bajt po bajtu stejný výstup jako RenderTable přes strings.Builder,
// Grow, strconv.AppendInt a utf8.RuneCountInString pro šířku sloupců.
func RenderTableFast(recs []Record) string {
	// TODO
	return ""
}
