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

// Classify vrací "negative", "zero" nebo "positive".
func Classify(n int) string {
	// TODO: úkol A
	return ""
}

// ZeroValueOf vrací textovou podobu zero value pro název typu.
func ZeroValueOf(kind string) string {
	// TODO: úkol B
	return ""
}

// CentsToPrice převede celé centy na desetinnou cenu (1999 -> 19.99).
func CentsToPrice(cents int) float64 {
	// TODO: úkol B
	return 0
}

// ToInt8 vrátí n jako int8 a true, pokud se do rozsahu vejde.
func ToInt8(n int) (int8, bool) {
	// TODO: úkol B
	return 0, false
}

// String implementuje fmt.Stringer.
func (l Level) String() string {
	// TODO: úkol C
	return ""
}

// ParseLevel převede jméno úrovně (case-insensitive) na Level.
func ParseLevel(s string) Level {
	// TODO: úkol C
	return *new(Level)
}

// Enabled vrací true, pokud je l alespoň na úrovni min.
func (l Level) Enabled(min Level) bool {
	// TODO: úkol C
	return false
}
