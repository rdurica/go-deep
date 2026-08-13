package solutions_test

import (
	"errors"
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
	m, err := exercise.NewMoney(1999, "EUR")
	if err != nil {
		t.Fatalf("NewMoney = %v", err)
	}
	if m.Cents() != 1999 || m.Currency() != "EUR" {
		t.Errorf("NewMoney = (%d, %q)", m.Cents(), m.Currency())
	}
	if _, err := exercise.NewMoney(100, "eur"); !errors.Is(err, exercise.ErrInvalidCurrency) {
		t.Errorf("NewMoney = %v, chci ErrInvalidCurrency", err)
	}
}

func TestMoneyString(t *testing.T) {
	if got := mustMoney(t, 1999, "EUR").String(); got != "19.99 EUR" {
		t.Errorf("String() = %q, chci %q", got, "19.99 EUR")
	}
	var zero exercise.Money
	if got := zero.String(); got != "0.00" {
		t.Errorf("nulová Money String() = %q, chci 0.00", got)
	}
}

func TestAddIsImmutable(t *testing.T) {
	a := mustMoney(t, 1999, "EUR")
	b := mustMoney(t, 250, "EUR")
	sum, err := a.Add(b)
	if err != nil {
		t.Fatalf("Add = %v", err)
	}
	if sum.Cents() != 2249 {
		t.Errorf("Add = %d, chci 2249", sum.Cents())
	}
	if a.Cents() != 1999 {
		t.Errorf("Add zmutoval operand a: %d, chci 1999", a.Cents())
	}
}

func TestAddCurrencyMismatch(t *testing.T) {
	eur := mustMoney(t, 100, "EUR")
	czk := mustMoney(t, 100, "CZK")
	if _, err := eur.Add(czk); !errors.Is(err, exercise.ErrCurrencyMismatch) {
		t.Errorf("Add jiné měny = %v, chci ErrCurrencyMismatch", err)
	}
}

func TestAllocate(t *testing.T) {
	parts, err := mustMoney(t, 100, "EUR").Allocate(3)
	if err != nil {
		t.Fatalf("Allocate = %v", err)
	}
	want := []int64{34, 33, 33}
	for i, p := range parts {
		if p.Cents() != want[i] {
			t.Fatalf("Allocate = %v, chci %v", centsOf(parts), want)
		}
	}
}

func TestAllocateInvalidCount(t *testing.T) {
	if _, err := mustMoney(t, 100, "EUR").Allocate(0); !errors.Is(err, exercise.ErrInvalidSplit) {
		t.Errorf("Allocate(0) = %v, chci ErrInvalidSplit", err)
	}
}

func TestAllocateDoesNotLoseCents(t *testing.T) {
	rnd := rand.New(rand.NewSource(34))
	for i := 0; i < 500; i++ {
		cents := int64(rnd.Intn(2001) - 1000)
		n := rnd.Intn(10) + 1
		m := mustMoney(t, cents, "EUR")
		parts, err := m.Allocate(n)
		if err != nil {
			t.Fatalf("Allocate = %v", err)
		}
		var sum int64
		for _, p := range parts {
			sum += p.Cents()
		}
		if sum != cents {
			t.Fatalf("součet %d != originál %d", sum, cents)
		}
	}
}

func centsOf(parts []exercise.Money) []int64 {
	out := make([]int64, len(parts))
	for i, p := range parts {
		out[i] = p.Cents()
	}
	return out
}
