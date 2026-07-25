// Package solutions obsahuje referenční řešení lekce 09.
package solutions

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// ByteLen vrací počet bajtů řetězce.
func ByteLen(s string) int {
	return len(s)
}

// RuneLen vrací počet run řetězce.
func RuneLen(s string) int {
	// Nealokuje, na rozdíl od len([]rune(s)).
	return utf8.RuneCountInString(s)
}

// ReverseRunes otočí pořadí run. Funguje s diakritikou i emoji.
func ReverseRunes(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// Truncate zkrátí text na nejvýš maxRunes run včetně připojeného "…".
func Truncate(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	// Jedna runa z rozpočtu padne na výpustku.
	return string([]rune(s)[:maxRunes-1]) + "…"
}

// Initials vrací iniciály: první runu každého slova velkým písmenem.
func Initials(fullName string) string {
	var sb strings.Builder
	// Fields rozdělí podle libovolného počtu bílých znaků a zahodí prázdné kusy.
	for _, word := range strings.Fields(fullName) {
		r, _ := utf8.DecodeRuneInString(word)
		if r == utf8.RuneError {
			continue
		}
		sb.WriteRune(unicode.ToUpper(r))
	}
	return sb.String()
}

// Join spojí kusy oddělovačem pomocí strings.Builder s předalokací.
func Join(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}

	size := len(sep) * (len(parts) - 1)
	for _, p := range parts {
		size += len(p)
	}

	var sb strings.Builder
	sb.Grow(size) // jediná alokace celé funkce
	for i, p := range parts {
		if i > 0 {
			sb.WriteString(sep)
		}
		sb.WriteString(p)
	}
	return sb.String()
}

// CountRunes spočítá výskyty jednotlivých run. Vždy vrací ne-nil mapu.
func CountRunes(s string) map[rune]int {
	counts := make(map[rune]int)
	// range po stringu dekóduje UTF-8, žádný []rune není potřeba.
	for _, r := range s {
		counts[r]++
	}
	return counts
}
