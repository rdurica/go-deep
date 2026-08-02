// Package solutions obsahuje referenční řešení lekce 03.
package solutions

import (
	"math"
	"strings"
)

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
// Classify vrací "negative", "zero" nebo "positive".
func Classify(n int) string {
	switch {
	case n < 0:
		return "negative"
	case n == 0:
		return "zero"
	default:
		return "positive"
	}
}

// ZeroValueOf vrací textovou podobu zero value pro název typu.
func ZeroValueOf(kind string) string {
	switch kind {
	case "int", "float64":
		return "0"
	case "string":
		return ""
	case "bool":
		return "false"
	case "slice", "map", "pointer", "chan", "interface":
		return "nil"
	default:
		return "unknown"
	}
}

// --- Stupeň: střední ---
// CentsToPrice převede celé centy na desetinnou cenu (1999 -> 19.99).
func CentsToPrice(cents int) float64 {
	return float64(cents) / 100
}

// ToInt8 vrátí n jako int8 a true, pokud se do rozsahu vejde.
func ToInt8(n int) (int8, bool) {
	if n < math.MinInt8 || n > math.MaxInt8 {
		return 0, false
	}
	return int8(n), true
}

// --- Stupeň: obtížný ---
// String implementuje fmt.Stringer.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel převede jméno úrovně (case-insensitive) na Level.
func ParseLevel(s string) Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelUnknown
	}
}

// Enabled vrací true, pokud je l alespoň na úrovni min.
func (l Level) Enabled(min Level) bool {
	return l >= min
}
