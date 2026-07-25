// Package exercise obsahuje cvičení lekce 09.
package exercise

// ByteLen vrací počet bajtů řetězce.
func ByteLen(s string) int {
	panic("TODO: úkol A")
}

// RuneLen vrací počet run řetězce.
func RuneLen(s string) int {
	panic("TODO: úkol A")
}

// ReverseRunes otočí pořadí run. Funguje s diakritikou i emoji.
func ReverseRunes(s string) string {
	panic("TODO: úkol B")
}

// Truncate zkrátí text na nejvýš maxRunes run včetně připojeného "…".
func Truncate(s string, maxRunes int) string {
	panic("TODO: úkol B")
}

// Initials vrací iniciály: první runu každého slova velkým písmenem.
func Initials(fullName string) string {
	panic("TODO: úkol B")
}

// Join spojí kusy oddělovačem. Postav ho na strings.Builder s předalokací
// přes Grow; strings.Join použít nesmíš.
func Join(parts []string, sep string) string {
	panic("TODO: úkol C")
}

// CountRunes spočítá výskyty jednotlivých run. Vždy vrací ne-nil mapu.
func CountRunes(s string) map[rune]int {
	panic("TODO: úkol C")
}
