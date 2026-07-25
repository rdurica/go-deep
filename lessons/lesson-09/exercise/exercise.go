// Package exercise obsahuje cvičení lekce 09.
package exercise

// ByteLen vrací počet bajtů řetězce.
func ByteLen(s string) int {
	// TODO: úkol A
	return 0
}

// RuneLen vrací počet run řetězce.
func RuneLen(s string) int {
	// TODO: úkol A
	return 0
}

// ReverseRunes otočí pořadí run. Funguje s diakritikou i emoji.
func ReverseRunes(s string) string {
	// TODO: úkol B
	return ""
}

// Truncate zkrátí text na nejvýš maxRunes run včetně připojeného "…".
func Truncate(s string, maxRunes int) string {
	// TODO: úkol B
	return ""
}

// Initials vrací iniciály: první runu každého slova velkým písmenem.
func Initials(fullName string) string {
	// TODO: úkol B
	return ""
}

// Join spojí kusy oddělovačem. Postav ho na strings.Builder s předalokací
// přes Grow; strings.Join použít nesmíš.
func Join(parts []string, sep string) string {
	// TODO: úkol C
	return ""
}

// CountRunes spočítá výskyty jednotlivých run. Vždy vrací ne-nil mapu.
func CountRunes(s string) map[rune]int {
	// TODO: úkol C
	return nil
}
