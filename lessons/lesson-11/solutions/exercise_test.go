package solutions_test

import (
	"fmt"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-11/solutions"
)

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

func TestAmountZeroValue(t *testing.T) {
	var a exercise.Amount
	if got := a.Cents(); got != 0 {
		t.Errorf("zero value Amount.Cents() = %d, chci 0", got)
	}
	if got := a.String(); got != "0.00" {
		t.Errorf("zero value Amount.String() = %q, chci %q", got, "0.00")
	}
}

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
