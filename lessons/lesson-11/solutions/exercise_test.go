package solutions_test

import (
	"fmt"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-11/solutions"
)

func TestNewAndCents(t *testing.T) {
	tests := []int64{0, 1, -1, 1999, -250, 1 << 40}
	for _, want := range tests {
		if got := exercise.NewAmount(want).Cents(); got != want {
			t.Errorf("New(%d).Cents() = %d, chci %d", want, got, want)
		}
	}
}

func TestAmountZeroValue(t *testing.T) {
	var a exercise.Amount
	if got := a.Cents(); got != 0 {
		t.Errorf("zero value Amount.Cents() = %d, chci 0", got)
	}
}

func TestAmountString(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0.00"},
		{5, "0.05"},
		{99, "0.99"},
		{100, "1.00"},
		{1999, "19.99"},
		{-1, "-0.01"},
		{-250, "-2.50"},
	}
	for _, tt := range tests {
		if got := exercise.NewAmount(tt.in).String(); got != tt.want {
			t.Errorf("New(%d).String() = %q, chci %q", tt.in, got, tt.want)
		}
	}
}

func TestAmountIsStringer(t *testing.T) {
	var s fmt.Stringer = exercise.NewAmount(1999)
	if got := fmt.Sprintf("%v", s); got != "19.99" {
		t.Errorf("fmt.Sprintf(%%v, New(1999)) = %q, chci %q", got, "19.99")
	}
}

func TestTotalOf(t *testing.T) {
	tests := []struct {
		name string
		in   []int64
		want int64
	}{
		{"empty input", nil, 0},
		{"one item", []int64{1999}, 1999},
		{"multiple items", []int64{1999, 1, 100}, 2100},
		{"negative items", []int64{500, -250, -250}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amounts := make([]exercise.Amount, 0, len(tt.in))
			for _, c := range tt.in {
				amounts = append(amounts, exercise.NewAmount(c))
			}
			if got := exercise.TotalOf(amounts).Cents(); got != tt.want {
				t.Errorf("TotalOf(%v) = %d, chci %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	tests := map[string]int64{
		"0":     0,
		"7":     700,
		"19.99": 1999,
		"-2.5":  -250,
		"+3.05": 305,
		"-0.01": -1,
		"12.00": 1200,
	}
	for in, want := range tests {
		if got := exercise.MustParse(in).Cents(); got != want {
			t.Errorf("MustParse(%q).Cents() = %d, chci %d", in, got, want)
		}
	}
}

func TestMustParsePanics(t *testing.T) {
	bad := []string{"", "abc", "1.234", "1.", ".5", "1.2.3", "--1", "1,5", " 1", "1 "}
	for _, in := range bad {
		t.Run(fmt.Sprintf("%q", in), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("MustParse(%q) měl panikovat, ale neudělal to", in)
				}
			}()
			_ = exercise.MustParse(in)
		})
	}
}

// TestSumCentsMatchesPublicAPI ověřuje, že vnitřní SumCents (které sahá přímo
// na neexportované pole cents) dává stejný výsledek jako sčítání přes veřejné
// API. Test sám na pole cents nesáhne — je mimo balíček money, takže by se
// ani nezkompiloval.
func TestSumCentsMatchesPublicAPI(t *testing.T) {
	amounts := []exercise.Amount{
		exercise.NewAmount(1999),
		exercise.NewAmount(-250),
		exercise.NewAmount(0),
		exercise.NewAmount(1),
	}

	var viaPublicAPI int64
	for _, a := range amounts {
		viaPublicAPI += a.Cents()
	}

	if got := exercise.SumCents(amounts); got != viaPublicAPI {
		t.Errorf("SumCents(...) = %d, chci %d", got, viaPublicAPI)
	}
	if got := exercise.SumCents(nil); got != 0 {
		t.Errorf("SumCents(nil) = %d, chci 0", got)
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		name  string
		cents int64
		n     int
		want  []int64
	}{
		{"beze zbytku", 900, 3, []int64{300, 300, 300}},
		{"se zbytkem", 1000, 3, []int64{334, 333, 333}},
		{"into one part", 1999, 1, []int64{1999}},
		{"negative amount", -250, 3, []int64{-84, -83, -83}},
		{"nula", 0, 4, []int64{0, 0, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts, ok := exercise.Split(exercise.NewAmount(tt.cents), tt.n)
			if !ok {
				t.Fatalf("Split(%d, %d) vrátil ok=false, chci true", tt.cents, tt.n)
			}
			if len(parts) != len(tt.want) {
				t.Fatalf("Split(%d, %d) má %d dílů, chci %d", tt.cents, tt.n, len(parts), len(tt.want))
			}
			var sum int64
			for i, p := range parts {
				if p.Cents() != tt.want[i] {
					t.Errorf("díl %d = %d, chci %d", i, p.Cents(), tt.want[i])
				}
				sum += p.Cents()
			}
			if sum != tt.cents {
				t.Errorf("součet dílů = %d, chci %d — cent se nesmí ztratit", sum, tt.cents)
			}
		})
	}
}

func TestSplitInvalidN(t *testing.T) {
	for _, n := range []int{0, -1, -10} {
		parts, ok := exercise.Split(exercise.NewAmount(100), n)
		if ok || parts != nil {
			t.Errorf("Split(100, %d) = (%v, %v), chci (nil, false)", n, parts, ok)
		}
	}
}
