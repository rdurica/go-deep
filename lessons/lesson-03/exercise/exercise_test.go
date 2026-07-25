package exercise_test

import (
	"fmt"
	"math"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-03/exercise"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{-1, "negative"},
		{math.MinInt, "negative"},
		{0, "zero"},
		{1, "positive"},
		{math.MaxInt, "positive"},
	}
	for _, tt := range tests {
		if got := exercise.Classify(tt.in); got != tt.want {
			t.Errorf("Classify(%d) = %q, chci %q", tt.in, got, tt.want)
		}
	}
}

func TestZeroValueOf(t *testing.T) {
	tests := map[string]string{
		"int":       "0",
		"float64":   "0",
		"string":    "",
		"bool":      "false",
		"slice":     "nil",
		"map":       "nil",
		"pointer":   "nil",
		"chan":      "nil",
		"interface": "nil",
		"struct":    "unknown",
		"":          "unknown",
	}
	for kind, want := range tests {
		if got := exercise.ZeroValueOf(kind); got != want {
			t.Errorf("ZeroValueOf(%q) = %q, chci %q", kind, got, want)
		}
	}
}

func TestCentsToPrice(t *testing.T) {
	tests := []struct {
		in   int
		want float64
	}{
		{1999, 19.99},
		{0, 0},
		{5, 0.05},
		{-250, -2.5},
		{100, 1},
	}
	for _, tt := range tests {
		got := exercise.CentsToPrice(tt.in)
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("CentsToPrice(%d) = %v, chci %v", tt.in, got, tt.want)
		}
	}
}

func TestToInt8(t *testing.T) {
	tests := []struct {
		in     int
		want   int8
		wantOK bool
	}{
		{0, 0, true},
		{127, 127, true},
		{-128, -128, true},
		{128, 0, false},
		{-129, 0, false},
		{300, 0, false},
	}
	for _, tt := range tests {
		got, ok := exercise.ToInt8(tt.in)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("ToInt8(%d) = (%d, %v), chci (%d, %v)", tt.in, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		in   exercise.Level
		want string
	}{
		{exercise.LevelUnknown, "UNKNOWN"},
		{exercise.LevelDebug, "DEBUG"},
		{exercise.LevelInfo, "INFO"},
		{exercise.LevelWarn, "WARN"},
		{exercise.LevelError, "ERROR"},
		{exercise.Level(99), "UNKNOWN"},
		{exercise.Level(-1), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.in.String(); got != tt.want {
			t.Errorf("Level(%d).String() = %q, chci %q", int(tt.in), got, tt.want)
		}
	}
}

func TestLevelIsStringer(t *testing.T) {
	var s fmt.Stringer = exercise.LevelWarn
	if got := fmt.Sprintf("%v", s); got != "WARN" {
		t.Errorf("fmt.Sprintf(%%v, LevelWarn) = %q, chci %q", got, "WARN")
	}
}

func TestParseLevel(t *testing.T) {
	tests := map[string]exercise.Level{
		"debug":   exercise.LevelDebug,
		"DEBUG":   exercise.LevelDebug,
		"  info ": exercise.LevelInfo,
		"Warn":    exercise.LevelWarn,
		"ERROR":   exercise.LevelError,
		"fatal":   exercise.LevelUnknown,
		"":        exercise.LevelUnknown,
	}
	for in, want := range tests {
		if got := exercise.ParseLevel(in); got != want {
			t.Errorf("ParseLevel(%q) = %v, chci %v", in, got, want)
		}
	}
}

func TestLevelEnabled(t *testing.T) {
	if !exercise.LevelError.Enabled(exercise.LevelInfo) {
		t.Error("Error má být povolený při minimu Info")
	}
	if exercise.LevelDebug.Enabled(exercise.LevelWarn) {
		t.Error("Debug nemá být povolený při minimu Warn")
	}
	if !exercise.LevelInfo.Enabled(exercise.LevelInfo) {
		t.Error("stejná úroveň má být povolená")
	}
}
