// Package solutions obsahuje referenční řešení lekce 09.
package solutions

import (
	"unicode/utf8"
)

// --- Stupeň: jednoduchý ---

// RuneLen vrací počet run řetězce.
func RuneLen(s string) int {
	return utf8.RuneCountInString(s)
}

// --- Stupeň: střední ---

// ByteLen vrací počet bajtů řetězce.
func ByteLen(s string) int {
	return len(s)
}

// ReverseRunes otočí pořadí run. Funguje s diakritikou i emoji.
func ReverseRunes(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// --- Stupeň: obtížný ---

// Truncate zkrátí text na nejvýš maxRunes run včetně připojeného "…".
func Truncate(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	return string([]rune(s)[:maxRunes-1]) + "…"
}
