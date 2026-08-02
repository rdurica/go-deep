// Package exercise obsahuje cvičení lekce 53.
package exercise

import (
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// wordRE hledá slova jako souvislé úseky písmen a číslic.
var wordRE = regexp.MustCompile(`[\p{L}\p{N}]+`)

// --- Stupeň: jednoduchý ---
// SumDigitsSlow sečte desítkové číslice v řetězci. Referenční, ale pomalá varianta:
// každou runu převede na string a nechá ji rozparsovat strconv.
func SumDigitsSlow(s string) int {
	sum := 0
	for _, r := range s {
		n, err := strconv.Atoi(string(r))
		if err != nil {
			continue
		}
		sum += n
	}
	return sum
}

// CountWordsSlow spočítá výskyty slov. Referenční, ale pomalá varianta:
// lowercase celého textu a regulární výraz v horké cestě.
func CountWordsSlow(text string) map[string]int {
	counts := map[string]int{}
	for _, w := range wordRE.FindAllString(strings.ToLower(text), -1) {
		counts[w]++
	}
	return counts
}

// IsWordRune vrací true pro znaky, které patří do slova.
func IsWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

// --- Stupeň: střední ---
// SumDigits sečte desítkové číslice v řetězci bez alokace.
// Výsledek musí shodovat se SumDigitsSlow; test vyžaduje 0 alokací (AllocsPerRun == 0).
func SumDigits(s string) int {
	// TODO
	return 0
}

// CountWords spočítá slova (IsWordRune, klíč malými písmeny) s max. 12 alokacemi.
// Prázdný text dá prázdnou nenilovou mapu. Předalokuj mapu; nekonvertuj celý text na lowercase.
func CountWords(text string) map[string]int {
	// TODO
	return nil
}

// JoinIDs spojí čísla čárkou bez zbytečných alokací.
// Prázdný vstup prázdný řetězec; max. 2 alokace pro 64 čísel (strings.Builder + Grow).
func JoinIDs(ids []int) string {
	// TODO
	return ""
}

// --- Stupeň: obtížný ---
// CaptureCPUProfile spustí CPU profil, zavolá f a profil ukončí i při panice (defer).
// Chybějící writer nebo f je chyba; chyba StartCPUProfile se obaluje kontextem.
func CaptureCPUProfile(w io.Writer, f func()) error {
	// TODO
	return nil
}

// CaptureHeapProfile vynutí GC a zapíše heap profil ve strojovém formátu (debug 0).
// Chybějící writer je chyba; profil nesmí být prázdný.
func CaptureHeapProfile(w io.Writer) error {
	// TODO
	return nil
}

// PprofHandler vrátí vlastní ServeMux s /debug/pprof/* endpointy, ne DefaultServeMux.
// /metrics na tomto handleru musí vrátit 404 (test to kontroluje).
func PprofHandler() http.Handler {
	// TODO
	return nil
}
