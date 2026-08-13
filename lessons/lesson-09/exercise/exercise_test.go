package exercise_test

import (
	"testing"
	"unicode/utf8"

	exercise "github.com/rdurica/go-deep/lessons/lesson-09/exercise"
)

func TestRuneLen(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"go", 2},
		{"kůň", 3},
		{"příliš žluťoučký kůň", 20},
		{"🐹", 1},
		{"a🐹b", 3},
	}
	for _, tt := range tests {
		if got := exercise.RuneLen(tt.in); got != tt.want {
			t.Errorf("RuneLen(%q) = %d, chci %d", tt.in, got, tt.want)
		}
	}
}

func TestRuneLenWithDiacritics(t *testing.T) {
	const s = "žluťoučký"
	if got := exercise.RuneLen(s); got != 9 {
		t.Errorf("RuneLen(%q) = %d, chci 9 (počítáš bajty místo run?)", s, got)
	}
}

func TestByteLen(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"go", 2},
		{"kůň", 5},
		{"příliš žluťoučký kůň", 29},
		{"🐹", 4},
	}
	for _, tt := range tests {
		if got := exercise.ByteLen(tt.in); got != tt.want {
			t.Errorf("ByteLen(%q) = %d, chci %d", tt.in, got, tt.want)
		}
	}
}

func TestByteLenWithDiacritics(t *testing.T) {
	const s = "žluťoučký"
	if got := exercise.ByteLen(s); got != 13 {
		t.Errorf("ByteLen(%q) = %d, chci 13 (počítáš runy místo bajtů?)", s, got)
	}
}

func TestReverseRunes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"one character", "a", "a"},
		{"ascii", "gopher", "rehpog"},
		{"czech", "kůň", "ňůk"},
		{"full sentence", "příliš žluťoučký kůň", "ňůk ýkčuoťulž šilířp"},
		{"emoji in middle", "a🐹b", "b🐹a"},
		{"emoji only", "🐹🐢", "🐢🐹"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.ReverseRunes(tt.in); got != tt.want {
				t.Errorf("ReverseRunes(%q) = %q, chci %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestReverseRunesDoesNotBreakUTF8(t *testing.T) {
	const s = "příliš žluťoučký kůň úpěl ďábelské ódy"
	got := exercise.ReverseRunes(s)

	if !utf8.ValidString(got) {
		t.Fatalf("ReverseRunes vrátil neplatné UTF-8: %q", got)
	}
	if utf8.RuneCountInString(got) != utf8.RuneCountInString(s) {
		t.Errorf("výsledek má %d run, vstup %d", utf8.RuneCountInString(got), utf8.RuneCountInString(s))
	}
	if back := exercise.ReverseRunes(got); back != s {
		t.Errorf("dvojí otočení dalo %q, chci %q", back, s)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		maxRunes int
		want     string
	}{
		{"shorter than limit", "kůň", 10, "kůň"},
		{"exactly at limit", "příliš", 6, "příliš"},
		{"zkracuje", "příliš", 4, "pří…"},
		{"truncates to one rune", "příliš", 1, "…"},
		{"limit nula", "příliš", 0, ""},
		{"negative limit", "příliš", -3, ""},
		{"empty input", "", 5, ""},
		{"ascii", "gopher", 3, "go…"},
		{"emoji", "a🐹bcd", 3, "a🐹…"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exercise.Truncate(tt.in, tt.maxRunes)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, chci %q", tt.in, tt.maxRunes, got, tt.want)
			}
			if tt.maxRunes > 0 && utf8.RuneCountInString(got) > tt.maxRunes {
				t.Errorf("výsledek má %d run, limit je %d", utf8.RuneCountInString(got), tt.maxRunes)
			}
		})
	}
}

func TestTruncateNeverExceedsLimit(t *testing.T) {
	const s = "žluťoučký kůň úpěl ďábelské ódy"
	for maxRunes := 1; maxRunes <= 40; maxRunes++ {
		got := exercise.Truncate(s, maxRunes)
		if n := utf8.RuneCountInString(got); n > maxRunes {
			t.Errorf("Truncate(s, %d) má %d run, chci nejvýš %d", maxRunes, n, maxRunes)
		}
		if !utf8.ValidString(got) {
			t.Errorf("Truncate(s, %d) vrátil neplatné UTF-8: %q", maxRunes, got)
		}
	}
}
