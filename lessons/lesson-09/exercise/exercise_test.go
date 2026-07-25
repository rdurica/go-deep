package exercise_test

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	exercise "github.com/rdurica/go-deep/lessons/lesson-09/exercise"
)

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

func TestByteLenARuneLenSeLisi(t *testing.T) {
	const s = "žluťoučký"
	b, r := exercise.ByteLen(s), exercise.RuneLen(s)
	if b <= r {
		t.Errorf("ByteLen(%q) = %d, RuneLen = %d — u diakritiky musí být bajtů víc", s, b, r)
	}
}

func TestReverseRunes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"prázdný", "", ""},
		{"jeden znak", "a", "a"},
		{"ascii", "gopher", "rehpog"},
		{"čeština", "kůň", "ňůk"},
		{"celá věta", "příliš žluťoučký kůň", "ňůk ýkčuoťulž šilířp"},
		{"emoji uprostřed", "a🐹b", "b🐹a"},
		{"jen emoji", "🐹🐢", "🐢🐹"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.ReverseRunes(tt.in); got != tt.want {
				t.Errorf("ReverseRunes(%q) = %q, chci %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestReverseRunesNerozbijeUTF8(t *testing.T) {
	// Otočení po bajtech by vyrobilo neplatné UTF-8 sekvence.
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
		{"kratší než limit", "kůň", 10, "kůň"},
		{"přesně na limit", "příliš", 6, "příliš"},
		{"zkracuje", "příliš", 4, "pří…"},
		{"zkracuje na jednu runu", "příliš", 1, "…"},
		{"limit nula", "příliš", 0, ""},
		{"záporný limit", "příliš", -3, ""},
		{"prázdný vstup", "", 5, ""},
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

func TestTruncateNikdyNeprekrociLimit(t *testing.T) {
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

func TestInitials(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"dvě slova s diakritikou", "Radek Ďurica", "RĎ"},
		{"malá písmena", "jan novák", "JN"},
		{"tři slova", "Jan Amos Komenský", "JAK"},
		{"vícenásobné mezery", "  jan   novák ", "JN"},
		{"tabulátory a nový řádek", "jan\t\tnovák\n", "JN"},
		{"jedno slovo", "Ďurica", "Ď"},
		{"prázdný vstup", "", ""},
		{"jen bílé znaky", "   \t ", ""},
		{"emoji jako slovo", "🐹 gopher", "🐹G"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.Initials(tt.in); got != tt.want {
				t.Errorf("Initials(%q) = %q, chci %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		sep   string
		want  string
	}{
		{"nil", nil, ",", ""},
		{"prázdný slice", []string{}, ",", ""},
		{"jeden prvek", []string{"go"}, ",", "go"},
		{"dva prvky", []string{"go", "php"}, ", ", "go, php"},
		{"prázdný oddělovač", []string{"a", "b", "c"}, "", "abc"},
		{"prázdné prvky", []string{"", "", ""}, "-", "--"},
		{"diakritika", []string{"kůň", "žluť"}, " a ", "kůň a žluť"},
		{"vícebajtový oddělovač", []string{"a", "b"}, "→", "a→b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.Join(tt.parts, tt.sep); got != tt.want {
				t.Errorf("Join(%q, %q) = %q, chci %q", tt.parts, tt.sep, got, tt.want)
			}
		})
	}
}

func TestJoinOdpovidaStringsJoin(t *testing.T) {
	// Referenční chování bereme ze stdlib, ale implementovat ho musíš sám.
	parts := []string{"příliš", "žluťoučký", "kůň", "", "úpěl"}
	for _, sep := range []string{"", ",", ", ", "→", "\n"} {
		want := strings.Join(parts, sep)
		if got := exercise.Join(parts, sep); got != want {
			t.Errorf("Join(parts, %q) = %q, chci %q", sep, got, want)
		}
	}
}

func TestJoinPredalokuje(t *testing.T) {
	// S Grow na přesnou délku stačí Builderu jediná alokace.
	parts := make([]string, 200)
	for i := range parts {
		parts[i] = "kus-" + strconv.Itoa(i)
	}

	allocs := testing.AllocsPerRun(100, func() {
		_ = exercise.Join(parts, ", ")
	})
	if allocs > 1 {
		t.Errorf("Join udělal %.0f alokací, chci nejvýš 1 — použij sb.Grow(size)", allocs)
	}
}

func TestCountRunes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want map[rune]int
	}{
		{"prázdný", "", map[rune]int{}},
		{"bez opakování", "go", map[rune]int{'g': 1, 'o': 1}},
		{"s opakováním", "gopher go", map[rune]int{'g': 2, 'o': 2, 'p': 1, 'h': 1, 'e': 1, 'r': 1, ' ': 1}},
		{"diakritika", "ůů", map[rune]int{'ů': 2}},
		{"emoji", "🐹a🐹", map[rune]int{'🐹': 2, 'a': 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exercise.CountRunes(tt.in)
			if got == nil {
				t.Fatal("CountRunes vrátil nil mapu, chci ne-nil")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CountRunes(%q) = %v, chci %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestCountRunesPocitaRunyNeBajty(t *testing.T) {
	// "kůň" má 4 bajty, ale 3 různé runy po jednom výskytu.
	got := exercise.CountRunes("kůň")
	if len(got) != 3 {
		t.Fatalf("CountRunes(\"kůň\") má %d klíčů, chci 3 (počítáš bajty?)", len(got))
	}
	total := 0
	for _, n := range got {
		total += n
	}
	if total != 3 {
		t.Errorf("součet výskytů = %d, chci 3", total)
	}
}

// benchParts je společný vstup obou benchmarků.
var benchParts = func() []string {
	parts := make([]string, 1000)
	for i := range parts {
		parts[i] = "část-" + strconv.Itoa(i)
	}
	return parts
}()

// BenchmarkBuilder měří tvůj Join postavený na strings.Builder.
func BenchmarkBuilder(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = exercise.Join(benchParts, ", ")
	}
}

// BenchmarkConcat měří naivní skládání přes += v cyklu.
func BenchmarkConcat(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		out := ""
		for j, p := range benchParts {
			if j > 0 {
				out += ", "
			}
			out += p
		}
		_ = out
	}
}
