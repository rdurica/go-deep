// Package exercise obsahuje cvičení lekce 09.
package exercise

// --- Stupeň: jednoduchý ---

// RuneLen vrací počet run (např. "kůň" → 3, "go" → 2).
// Použij utf8.RuneCountInString — nesmíš převádět celý text na []rune.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Počítá bajty místo run — typická chyba po PHP.
// Najdi ji a oprav; testy před opravou padají.
func RuneLen(s string) int {
	return len(s)
}

// --- Stupeň: střední ---

// ByteLen vrací počet bajtů řetězce (např. "kůň" → 5).
func ByteLen(s string) int {
	// TODO
	return 0
}

// ReverseRunes otočí pořadí run. Zvládne diakritiku ("kůň" → "ňůk") i emoji ("a🐹b" → "b🐹a").
// Prázdný vstup vrací prázdný string.
func ReverseRunes(s string) string {
	// TODO
	return ""
}

// --- Stupeň: obtížný ---

// Truncate zkrátí text na nejvýš maxRunes run včetně znaku … (U+2026).
// maxRunes <= 0 → "". Text krátký nebo přesně dlouhý maxRunes run → beze změny, bez ….
// Jinak prvních maxRunes-1 run plus … (Truncate("příliš", 4) → "pří…").
func Truncate(s string, maxRunes int) string {
	// TODO
	return ""
}
