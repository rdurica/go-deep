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

// SumDigits sečte desítkové číslice v řetězci bez jediné alokace.
func SumDigits(s string) int {
	// TODO: úkol A
	return 0
}

// CountWords spočítá výskyty slov (case-insensitive) s minimem alokací.
func CountWords(text string) map[string]int {
	// TODO: úkol B
	return nil
}

// JoinIDs spojí čísla čárkou do jednoho řetězce.
func JoinIDs(ids []int) string {
	// TODO: úkol B
	return ""
}

// CaptureCPUProfile spustí f a zapíše CPU profil jejího běhu do w.
func CaptureCPUProfile(w io.Writer, f func()) error {
	// TODO: úkol C
	return nil
}

// CaptureHeapProfile zapíše aktuální heap profil do w.
func CaptureHeapProfile(w io.Writer) error {
	// TODO: úkol C
	return nil
}

// PprofHandler vrátí handler s pprof endpointy na vlastním muxu.
func PprofHandler() http.Handler {
	// TODO: úkol C
	return *new(http.Handler)
}
