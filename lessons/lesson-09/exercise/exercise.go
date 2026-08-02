// Package exercise obsahuje cvičení lekce 09.
package exercise

// --- Stupeň: jednoduchý ---
// ByteLen vrací počet bajtů řetězce (např. "kůň" → 5).
func ByteLen(s string) int {
	// TODO
	return 0
}

// RuneLen vrací počet run (např. "kůň" → 3, "go" → 2).
// Použij utf8.RuneCountInString — nesmíš převádět celý text na []rune.
func RuneLen(s string) int {
	// TODO
	return 0
}

// --- Stupeň: střední ---
// ReverseRunes otočí pořadí run. Zvládne diakritiku ("kůň" → "ňůk") i emoji ("a🐹b" → "b🐹a").
// Prázdný vstup vrací prázdný string.
func ReverseRunes(s string) string {
	// TODO
	return ""
}

// Truncate zkrátí text na nejvýš maxRunes run včetně znaku … (U+2026).
// maxRunes <= 0 → "". Text krátký nebo přesně dlouhý maxRunes run → beze změny, bez ….
// Jinak prvních maxRunes-1 run plus … (Truncate("příliš", 4) → "pří…").
func Truncate(s string, maxRunes int) string {
	// TODO
	return ""
}

// --- Stupeň: obtížný ---
// Initials vrací iniciály: první runu každého slova velkým písmenem.
// Ošetři vícenásobné mezery a okrajové bílé znaky. Prázdný vstup → "".
// Např. "Radek Ďurica" → "RĎ", "  jan   novák " → "JN".
func Initials(fullName string) string {
	// TODO
	return ""
}

// Join spojí kusy oddělovačem sep. Nesmíš použít strings.Join.
// Postav to na strings.Builder; délku v bajtech spočítej dopředu a předalokuj.
// Prázdný vstup → ""; jeden prvek → ten prvek bez oddělovače.
func Join(parts []string, sep string) string {
	// TODO
	return ""
}

// CountRunes spočítá výskyty jednotlivých run. Vždy vrací ne-nil mapu.
// Projdi text po runách bez mezilehlého []rune.
func CountRunes(s string) map[rune]int {
	// TODO
	return nil
}
