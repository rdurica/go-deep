// Package exercise obsahuje cvičení lekce 03.
package exercise

// Level je úroveň logování.
type Level int

// Úrovně logování. LevelUnknown je zero value, tedy "nenastaveno".
const (
	LevelUnknown Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
)

// --- Stupeň: jednoduchý ---

// CentsToPrice převede celé centy na desetinnou cenu (1999 → 19.99).
// Pozor na pořadí konverze a dělení.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Obsahuje typickou chybu s celočíselným dělením
// před konverzí na float64. Najdi ji a oprav — testy před opravou padají.
func CentsToPrice(cents int) float64 {
	return float64(cents / 100)
}

// --- Stupeň: střední ---

// Classify vrací "negative", "zero" nebo "positive" podle znaménka n.
// Použij switch bez výrazu (switch { case n < 0: ... }), ne řetězec if/else.
func Classify(n int) string {
	// TODO
	return ""
}

// ZeroValueOf vrací textovou podobu zero value pro název typu.
// int/float64 → "0", string → "", bool → "false", slice/map/pointer/chan/interface → "nil".
// Cokoli jiného → "unknown".
func ZeroValueOf(kind string) string {
	// TODO
	return ""
}

// ToInt8 vrátí n jako int8 a true, pokud se do rozsahu int8 vejde; jinak 0, false.
func ToInt8(n int) (int8, bool) {
	// TODO
	return 0, false
}

// --- Stupeň: obtížný ---

// String implementuje fmt.Stringer: UNKNOWN, DEBUG, INFO, WARN, ERROR.
// Hodnota mimo rozsah také vrací "UNKNOWN".
func (l Level) String() string {
	// TODO
	return ""
}

// ParseLevel převede jméno úrovně (case-insensitive) na Level.
// Neznámý vstup dá LevelUnknown.
func ParseLevel(s string) Level {
	// TODO
	return *new(Level)
}

// Enabled vrací true, pokud je l alespoň na úrovni min.
// LevelUnknown není povolený pro žádné min > Unknown.
func (l Level) Enabled(min Level) bool {
	// TODO
	return false
}
