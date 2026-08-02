package solutions_test

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-34/solutions"
)

func mustMoney(t *testing.T, cents int64, c exercise.Currency) exercise.Money {
	t.Helper()
	m, err := exercise.NewMoney(cents, c)
	if err != nil {
		t.Fatalf("NewMoney(%d, %q) = %v", cents, string(c), err)
	}
	return m
}

func TestNewMoney(t *testing.T) {
	tests := []struct {
		in      exercise.Currency
		wantErr bool
	}{
		{"EUR", false},
		{"CZK", false},
		{"USD", false},
		{"eur", true},
		{"EU", true},
		{"EURO", true},
		{"", true},
		{"E1R", true},
		{"E R", true},
	}
	for _, tt := range tests {
		t.Run(string(tt.in), func(t *testing.T) {
			m, err := exercise.NewMoney(1999, tt.in)
			if tt.wantErr {
				if !errors.Is(err, exercise.ErrInvalidCurrency) {
					t.Fatalf("NewMoney(_, %q) = %v, chci ErrInvalidCurrency", string(tt.in), err)
				}
				if m != (exercise.Money{}) {
					t.Errorf("při chybě chci nulovou hodnotu, mám %+v", m)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewMoney(_, %q) = %v, chci nil", string(tt.in), err)
			}
			if m.Cents() != 1999 || m.Currency() != tt.in {
				t.Errorf("NewMoney = (%d, %q), chci (1999, %q)", m.Cents(), string(m.Currency()), string(tt.in))
			}
		})
	}
}

func TestMoneyString(t *testing.T) {
	tests := []struct {
		cents int64
		want  string
	}{
		{1999, "19.99 EUR"},
		{100, "1.00 EUR"},
		{5, "0.05 EUR"},
		{0, "0.00 EUR"},
		{-1999, "-19.99 EUR"},
		{-5, "-0.05 EUR"},
		{123456789, "1234567.89 EUR"},
	}
	for _, tt := range tests {
		m := mustMoney(t, tt.cents, "EUR")
		if got := m.String(); got != tt.want {
			t.Errorf("Money(%d).String() = %q, chci %q", tt.cents, got, tt.want)
		}
	}

	var zero exercise.Money
	if got := zero.String(); got != "0.00" {
		t.Errorf("nulová hodnota String() = %q, chci %q", got, "0.00")
	}
}

func TestMoneyIsStringer(t *testing.T) {
	var s fmt.Stringer = mustMoney(t, 1999, "EUR")
	if got := fmt.Sprintf("%v", s); got != "19.99 EUR" {
		t.Errorf("Sprintf(%%v) = %q, chci %q", got, "19.99 EUR")
	}
}

func TestFloatNaPenizeSelhava(t *testing.T) {
	// Motivace celé lekce: tohle je v float64 nepravda.
	if 0.1+0.2 == 0.3 {
		t.Skip("tenhle stroj počítá jinak než IEEE 754")
	}

	a := mustMoney(t, 10, "EUR")
	b := mustMoney(t, 20, "EUR")
	want := mustMoney(t, 30, "EUR")

	got, err := a.Add(b)
	if err != nil {
		t.Fatalf("Add = %v", err)
	}
	if got != want {
		t.Errorf("0.10 + 0.20 = %v, chci %v — celá čísla musí sedět přesně", got, want)
	}
}

func TestAddSub(t *testing.T) {
	a := mustMoney(t, 1999, "EUR")
	b := mustMoney(t, 250, "EUR")

	sum, err := a.Add(b)
	if err != nil {
		t.Fatalf("Add = %v", err)
	}
	if sum.Cents() != 2249 {
		t.Errorf("Add = %d, chci 2249", sum.Cents())
	}

	diff, err := a.Sub(b)
	if err != nil {
		t.Fatalf("Sub = %v", err)
	}
	if diff.Cents() != 1749 {
		t.Errorf("Sub = %d, chci 1749", diff.Cents())
	}

	// Operace nesmí měnit původní hodnoty.
	if a.Cents() != 1999 || b.Cents() != 250 {
		t.Errorf("operace zmutovaly operandy: a=%d b=%d", a.Cents(), b.Cents())
	}
}

func TestMichaniMen(t *testing.T) {
	eur := mustMoney(t, 100, "EUR")
	czk := mustMoney(t, 100, "CZK")

	if _, err := eur.Add(czk); !errors.Is(err, exercise.ErrCurrencyMismatch) {
		t.Errorf("Add jiné měny = %v, chci ErrCurrencyMismatch", err)
	}
	if _, err := eur.Sub(czk); !errors.Is(err, exercise.ErrCurrencyMismatch) {
		t.Errorf("Sub jiné měny = %v, chci ErrCurrencyMismatch", err)
	}
	var zero exercise.Money
	if _, err := eur.Add(zero); !errors.Is(err, exercise.ErrCurrencyMismatch) {
		t.Errorf("Add nulové hodnoty bez měny = %v, chci ErrCurrencyMismatch", err)
	}
}

func TestMulNegIsZero(t *testing.T) {
	m := mustMoney(t, 250, "EUR")

	if got := m.Mul(4); got.Cents() != 1000 || got.Currency() != "EUR" {
		t.Errorf("Mul(4) = %v, chci 10.00 EUR", got)
	}
	if got := m.Mul(0); !got.IsZero() {
		t.Errorf("Mul(0) = %v, chci nulu", got)
	}
	if got := m.Mul(-2); got.Cents() != -500 {
		t.Errorf("Mul(-2) = %d, chci -500", got.Cents())
	}
	if got := m.Neg(); got.Cents() != -250 || got.Currency() != "EUR" {
		t.Errorf("Neg() = %v, chci -2.50 EUR", got)
	}
	if got := m.Neg().Neg(); got != m {
		t.Errorf("dvojitá negace = %v, chci %v", got, m)
	}
	if m.IsZero() {
		t.Error("2.50 EUR není nula")
	}
	if !mustMoney(t, 0, "EUR").IsZero() {
		t.Error("0.00 EUR má být nula")
	}
	if m.Cents() != 250 {
		t.Errorf("Mul/Neg zmutovaly příjemce: %d", m.Cents())
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b int64
		want int
	}{
		{100, 200, -1},
		{200, 100, 1},
		{100, 100, 0},
		{-100, 0, -1},
		{0, -100, 1},
	}
	for _, tt := range tests {
		got, err := mustMoney(t, tt.a, "EUR").Compare(mustMoney(t, tt.b, "EUR"))
		if err != nil {
			t.Fatalf("Compare = %v", err)
		}
		if got != tt.want {
			t.Errorf("Compare(%d, %d) = %d, chci %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMoneyIsComparableAndMapKey(t *testing.T) {
	a := mustMoney(t, 1999, "EUR")
	b := mustMoney(t, 1999, "EUR")
	c := mustMoney(t, 1999, "CZK")

	if a != b {
		t.Error("stejné částky ve stejné měně se musí rovnat přes ==")
	}
	if a == c {
		t.Error("stejné číslo v jiné měně se nesmí rovnat")
	}

	counts := map[exercise.Money]int{}
	counts[a]++
	counts[b]++
	counts[c]++
	if counts[a] != 2 || counts[c] != 1 {
		t.Errorf("Money jako klíč mapy = %v, chci 2 a 1", counts)
	}
}

func TestAllocate(t *testing.T) {
	tests := []struct {
		name  string
		cents int64
		n     int
		want  []int64
	}{
		{"beze zbytku", 100, 2, []int64{50, 50}},
		{"not divisible by three", 100, 3, []int64{34, 33, 33}},
		{"one part", 1999, 1, []int64{1999}},
		{"more parts than cents", 3, 5, []int64{1, 1, 1, 0, 0}},
		{"nula", 0, 4, []int64{0, 0, 0, 0}},
		{"negative amount", -100, 3, []int64{-34, -33, -33}},
		{"zbytek dva", 5, 3, []int64{2, 2, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts, err := mustMoney(t, tt.cents, "EUR").Allocate(tt.n)
			if err != nil {
				t.Fatalf("Allocate(%d) = %v", tt.n, err)
			}
			if len(parts) != tt.n {
				t.Fatalf("len(parts) = %d, chci %d", len(parts), tt.n)
			}
			for i, p := range parts {
				if p.Cents() != tt.want[i] {
					t.Errorf("parts = %v, chci %v", centsOf(parts), tt.want)
					break
				}
				if p.Currency() != "EUR" {
					t.Errorf("parts[%d].Currency() = %q, chci EUR", i, string(p.Currency()))
				}
			}
		})
	}
}

func TestAllocateInvalidCount(t *testing.T) {
	m := mustMoney(t, 100, "EUR")
	for _, n := range []int{0, -1, -100} {
		if _, err := m.Allocate(n); !errors.Is(err, exercise.ErrInvalidSplit) {
			t.Errorf("Allocate(%d) = %v, chci ErrInvalidSplit", n, err)
		}
	}
}

func TestAllocateDoesNotLoseCents(t *testing.T) {
	// Generovaná data, aby nešlo projít zadrátovaným výsledkem.
	rnd := rand.New(rand.NewSource(20240307))
	for i := 0; i < 2000; i++ {
		cents := int64(rnd.Intn(2_000_001) - 1_000_000)
		n := rnd.Intn(17) + 1

		m := mustMoney(t, cents, "EUR")
		parts, err := m.Allocate(n)
		if err != nil {
			t.Fatalf("Allocate(%d) na %d = %v", n, cents, err)
		}

		var sum int64
		for _, p := range parts {
			sum += p.Cents()
		}
		if sum != cents {
			t.Fatalf("součet dílů %d != originál %d (n=%d, díly=%v)", sum, cents, n, centsOf(parts))
		}

		// Rozdíl mezi největším a nejmenším dílem smí být nejvýš jedna jednotka.
		min, max := parts[0].Cents(), parts[0].Cents()
		for _, p := range parts {
			if p.Cents() < min {
				min = p.Cents()
			}
			if p.Cents() > max {
				max = p.Cents()
			}
		}
		if max-min > 1 {
			t.Fatalf("díly %v se liší o víc než jednu jednotku (cents=%d, n=%d)", centsOf(parts), cents, n)
		}
	}
}

func TestAllocateRatio(t *testing.T) {
	tests := []struct {
		name   string
		cents  int64
		ratios []int
		want   []int64
	}{
		{"classic Fowler", 5, []int{3, 7}, []int64{2, 3}},
		{"beze zbytku", 500, []int{3, 7}, []int64{150, 350}},
		{"equal shares", 100, []int{1, 1, 1}, []int64{34, 33, 33}},
		{"zero share", 100, []int{0, 1}, []int64{0, 100}},
		{"single share", 999, []int{5}, []int64{999}},
		{"negative amount", -5, []int{3, 7}, []int64{-2, -3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts, err := mustMoney(t, tt.cents, "EUR").AllocateRatio(tt.ratios)
			if err != nil {
				t.Fatalf("AllocateRatio(%v) = %v", tt.ratios, err)
			}
			got := centsOf(parts)
			if len(got) != len(tt.want) {
				t.Fatalf("AllocateRatio(%v) = %v, chci %v", tt.ratios, got, tt.want)
			}
			var sum int64
			for i := range got {
				sum += got[i]
				if got[i] != tt.want[i] {
					t.Fatalf("AllocateRatio(%v) = %v, chci %v", tt.ratios, got, tt.want)
				}
			}
			if sum != tt.cents {
				t.Errorf("součet = %d, chci %d", sum, tt.cents)
			}
		})
	}
}

func TestAllocateRatioInvalidInputs(t *testing.T) {
	m := mustMoney(t, 100, "EUR")
	tests := []struct {
		name   string
		ratios []int
	}{
		{"nil", nil},
		{"empty", []int{}},
		{"negative ratio", []int{1, -1}},
		{"zero sum", []int{0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := m.AllocateRatio(tt.ratios); !errors.Is(err, exercise.ErrInvalidRatios) {
				t.Errorf("AllocateRatio(%v) = %v, chci ErrInvalidRatios", tt.ratios, err)
			}
		})
	}
}

func TestParseMoney(t *testing.T) {
	tests := []struct {
		in    string
		cents int64
		cur   exercise.Currency
	}{
		{"19.99 EUR", 1999, "EUR"},
		{"0.05 CZK", 5, "CZK"},
		{"0.00 USD", 0, "USD"},
		{"-19.99 EUR", -1999, "EUR"},
		{"-0.05 EUR", -5, "EUR"},
		{"  19.99 EUR  ", 1999, "EUR"},
		{"1234567.89 EUR", 123456789, "EUR"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := exercise.ParseMoney(tt.in)
			if err != nil {
				t.Fatalf("ParseMoney(%q) = %v", tt.in, err)
			}
			want := mustMoney(t, tt.cents, tt.cur)
			if got != want {
				t.Errorf("ParseMoney(%q) = %v, chci %v", tt.in, got, want)
			}
		})
	}
}

func TestParseMoneyErrors(t *testing.T) {
	bad := []string{
		"",
		"19.99",
		"EUR",
		"19,99 EUR",
		"19.9 EUR",
		"19.999 EUR",
		"19 EUR",
		"+19.99 EUR",
		"19.99 eur",
		"19.99 EURO",
		"abc EUR",
		"19.99 EUR extra",
		".99 EUR",
	}
	for _, in := range bad {
		t.Run(in, func(t *testing.T) {
			if _, err := exercise.ParseMoney(in); err == nil {
				t.Fatalf("ParseMoney(%q) = nil, chci chybu", in)
			} else if !errors.Is(err, exercise.ErrInvalidFormat) && !errors.Is(err, exercise.ErrInvalidCurrency) {
				t.Errorf("ParseMoney(%q) = %v, chci ErrInvalidFormat nebo ErrInvalidCurrency", in, err)
			}
		})
	}
}

func TestParseMoneyStringRoundTrip(t *testing.T) {
	rnd := rand.New(rand.NewSource(7))
	currencies := []exercise.Currency{"EUR", "CZK", "USD"}
	for i := 0; i < 1000; i++ {
		want := mustMoney(t, int64(rnd.Intn(4_000_001)-2_000_000), currencies[rnd.Intn(len(currencies))])
		got, err := exercise.ParseMoney(want.String())
		if err != nil {
			t.Fatalf("ParseMoney(%q) = %v", want.String(), err)
		}
		if got != want {
			t.Fatalf("round trip %q dal %v, chci %v", want.String(), got, want)
		}
	}
}

// centsOf vytáhne z dílů holá čísla, aby se daly pohodlně porovnat v hlášce.
func centsOf(parts []exercise.Money) []int64 {
	out := make([]int64, len(parts))
	for i, p := range parts {
		out[i] = p.Cents()
	}
	return out
}
