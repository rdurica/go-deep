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
	panic("TODO: úkol A")
}

// ZeroValueOf vrací textovou podobu zero value pro název typu.
func ZeroValueOf(kind string) string {
	panic("TODO: úkol B")
}

// CentsToPrice převede celé centy na desetinnou cenu (1999 -> 19.99).
func CentsToPrice(cents int) float64 {
	panic("TODO: úkol B")
}

// ToInt8 vrátí n jako int8 a true, pokud se do rozsahu vejde.
func ToInt8(n int) (int8, bool) {
	panic("TODO: úkol B")
}

// String implementuje fmt.Stringer.
func (l Level) String() string {
	panic("TODO: úkol C")
}

// ParseLevel převede jméno úrovně (case-insensitive) na Level.
func ParseLevel(s string) Level {
	panic("TODO: úkol C")
}

// Enabled vrací true, pokud je l alespoň na úrovni min.
func (l Level) Enabled(min Level) bool {
	panic("TODO: úkol C")
}
