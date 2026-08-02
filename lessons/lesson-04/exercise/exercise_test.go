package exercise_test

import (
	"math"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-04/exercise"
)

func TestSum(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"no args", nil, 0},
		{"jeden", []int{7}, 7},
		{"more", []int{1, 2, 3, 4}, 10},
		{"negative", []int{-5, 5, -3}, -3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.Sum(tt.in...); got != tt.want {
				t.Errorf("Sum(%v...) = %d, chci %d", tt.in, got, tt.want)
			}
		})
	}

	if got := exercise.Sum(); got != 0 {
		t.Errorf("Sum() = %d, chci 0", got)
	}
	if got := exercise.Sum(2, 3); got != 5 {
		t.Errorf("Sum(2, 3) = %d, chci 5", got)
	}
}

func TestMinMax(t *testing.T) {
	tests := []struct {
		name    string
		in      []int
		wantMin int
		wantMax int
		wantOK  bool
	}{
		{"nil", nil, 0, 0, false},
		{"empty", []int{}, 0, 0, false},
		{"one element", []int{42}, 42, 42, true},
		{"sorted", []int{1, 2, 3}, 1, 3, true},
		{"unsorted", []int{5, -2, 9, 0}, -2, 9, true},
		{"all same", []int{3, 3, 3}, 3, 3, true},
		{"edge values", []int{math.MaxInt, math.MinInt, 0}, math.MinInt, math.MaxInt, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotMax, ok := exercise.MinMax(tt.in)
			if gotMin != tt.wantMin || gotMax != tt.wantMax || ok != tt.wantOK {
				t.Errorf("MinMax(%v) = (%d, %d, %v), chci (%d, %d, %v)",
					tt.in, gotMin, gotMax, ok, tt.wantMin, tt.wantMax, tt.wantOK)
			}
		})
	}
}

func TestMinMaxDoesNotMutateInput(t *testing.T) {
	in := []int{5, -2, 9, 0}
	before := append([]int(nil), in...)
	exercise.MinMax(in)
	for i := range in {
		if in[i] != before[i] {
			t.Fatalf("MinMax změnila vstup: %v, chci %v", in, before)
		}
	}
}

func TestCounter(t *testing.T) {
	next := exercise.Counter()
	for want := 1; want <= 5; want++ {
		if got := next(); got != want {
			t.Fatalf("%d. volání čítače = %d, chci %d", want, got, want)
		}
	}
}

func TestCounterIndependentState(t *testing.T) {
	a := exercise.Counter()
	b := exercise.Counter()

	a()
	a()
	a()

	if got := b(); got != 1 {
		t.Errorf("druhý čítač vrátil %d, chci 1 — každý má mít vlastní stav", got)
	}
	if got := a(); got != 4 {
		t.Errorf("první čítač vrátil %d, chci 4", got)
	}
}

func TestApply(t *testing.T) {
	double := func(x int) int { return x * 2 }

	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"nil", nil, []int{}},
		{"empty", []int{}, []int{}},
		{"common", []int{1, 2, 3}, []int{2, 4, 6}},
		{"negative", []int{-1, 0, 5}, []int{-2, 0, 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exercise.Apply(tt.in, double)
			if len(got) != len(tt.want) {
				t.Fatalf("Apply(%v, double) má délku %d, chci %d", tt.in, len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("Apply(%v, double) = %v, chci %v", tt.in, got, tt.want)
				}
			}
		})
	}
}

func TestApplyDoesNotMutateInput(t *testing.T) {
	in := []int{1, 2, 3}
	got := exercise.Apply(in, func(x int) int { return x * 10 })

	for i, want := range []int{1, 2, 3} {
		if in[i] != want {
			t.Fatalf("Apply změnila vstup: %v, chci [1 2 3]", in)
		}
	}
	if len(got) > 0 && &got[0] == &in[0] {
		t.Error("Apply vrátila stejné podkladové pole, chci nový slice")
	}
}

func TestCompose(t *testing.T) {
	inc := func(x int) int { return x + 1 }
	double := func(x int) int { return x * 2 }

	if got := exercise.Compose()(7); got != 7 {
		t.Errorf("Compose()(7) = %d, chci 7 — bez argumentů má vzniknout identita", got)
	}
	if got := exercise.Compose(inc)(7); got != 8 {
		t.Errorf("Compose(inc)(7) = %d, chci 8", got)
	}
	// zleva doprava: nejdřív inc, pak double
	if got := exercise.Compose(inc, double)(3); got != 8 {
		t.Errorf("Compose(inc, double)(3) = %d, chci 8", got)
	}
	if got := exercise.Compose(double, inc)(3); got != 7 {
		t.Errorf("Compose(double, inc)(3) = %d, chci 7", got)
	}
	if got := exercise.Compose(inc, inc, inc, double)(0); got != 6 {
		t.Errorf("Compose(inc, inc, inc, double)(0) = %d, chci 6", got)
	}
}

func TestComposeCapturesLoopVariable(t *testing.T) {
	// Od Go 1.22 má každá iterace vlastní proměnnou d, takže každá closure
	// zachytí jinou hodnotu. Se starou sémantikou by tenhle test spadl.
	var fs []func(int) int
	for _, d := range []int{1, 2, 3} {
		fs = append(fs, func(x int) int { return x + d })
	}

	if got := exercise.Compose(fs...)(0); got != 6 {
		t.Errorf("Compose(fs...)(0) = %d, chci 6", got)
	}
}

func TestMemoizeReturnsSameResult(t *testing.T) {
	square := func(x int) int { return x * x }
	memo, _ := exercise.Memoize(square)

	for _, x := range []int{0, 1, 5, -3, 5, 1} {
		if got := memo(x); got != x*x {
			t.Errorf("memo(%d) = %d, chci %d", x, got, x*x)
		}
	}
}

func TestMemoizeCountsActualCalls(t *testing.T) {
	square := func(x int) int { return x * x }
	memo, calls := exercise.Memoize(square)

	if got := calls(); got != 0 {
		t.Errorf("před prvním voláním calls() = %d, chci 0", got)
	}

	memo(4)
	memo(4)
	memo(4)
	if got := calls(); got != 1 {
		t.Errorf("po třech voláních memo(4) je calls() = %d, chci 1", got)
	}

	memo(5)
	memo(4)
	if got := calls(); got != 2 {
		t.Errorf("po přidání memo(5) je calls() = %d, chci 2", got)
	}
}

func TestMemoizeActuallyCaches(t *testing.T) {
	// Funkce, která pokaždé vrátí něco jiného — kdyby memoizace necachovala,
	// druhé volání by dalo jinou hodnotu.
	n := 0
	f := func(x int) int {
		n++
		return n * 100
	}
	memo, _ := exercise.Memoize(f)

	first := memo(1)
	second := memo(1)
	if first != second {
		t.Errorf("memo(1) vrátila podruhé %d, chci %d — výsledek se má cachovat", second, first)
	}
}

func TestMemoizeIndependentInstances(t *testing.T) {
	id := func(x int) int { return x }

	memoA, callsA := exercise.Memoize(id)
	memoB, callsB := exercise.Memoize(id)

	memoA(1)
	memoA(2)
	memoB(1)

	if got := callsA(); got != 2 {
		t.Errorf("callsA() = %d, chci 2", got)
	}
	if got := callsB(); got != 1 {
		t.Errorf("callsB() = %d, chci 1 — instance mají mít oddělený stav", got)
	}
}
