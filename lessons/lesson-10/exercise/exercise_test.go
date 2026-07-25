package exercise_test

import (
	"errors"
	"reflect"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-10/exercise"
)

func TestDeferOrderJeLIFO(t *testing.T) {
	got := exercise.DeferOrder()
	want := []string{"third", "second", "first"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DeferOrder() = %v, chci %v (defery běží v opačném pořadí registrace)", got, want)
	}
}

func TestDeferOrderJeStabilni(t *testing.T) {
	first := exercise.DeferOrder()
	for i := 0; i < 10; i++ {
		if got := exercise.DeferOrder(); !reflect.DeepEqual(got, first) {
			t.Fatalf("DeferOrder() vrátil %v, předtím %v", got, first)
		}
	}
}

func TestSumWithLog(t *testing.T) {
	tests := []struct {
		name      string
		nums      []int
		wantTotal int
		wantSteps []string
	}{
		{"prázdný", nil, 0, []string{"total=0"}},
		{"jedno číslo", []int{5}, 5, []string{"+5=5", "total=5"}},
		{"tři čísla", []int{1, 2, 3}, 6, []string{"+1=1", "+2=3", "+3=6", "total=6"}},
		{"se zápornými", []int{10, -4}, 6, []string{"+10=10", "+-4=6", "total=6"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, steps := exercise.SumWithLog(tt.nums)
			if total != tt.wantTotal {
				t.Errorf("SumWithLog(%v) total = %d, chci %d", tt.nums, total, tt.wantTotal)
			}
			if !reflect.DeepEqual(steps, tt.wantSteps) {
				t.Errorf("SumWithLog(%v) steps = %v, chci %v", tt.nums, steps, tt.wantSteps)
			}
		})
	}
}

func TestSumWithLogPosledniKrokPridavaDefer(t *testing.T) {
	// Kdyby defer neupravoval pojmenovanou návratovou hodnotu, poslední
	// krok by se do výsledku vůbec nedostal.
	_, steps := exercise.SumWithLog([]int{4, 4})
	if len(steps) == 0 {
		t.Fatal("SumWithLog nevrátil žádné kroky")
	}
	if last := steps[len(steps)-1]; last != "total=8" {
		t.Errorf("poslední krok = %q, chci %q", last, "total=8")
	}
}

func TestSafeDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    int
		want    int
		wantErr bool
	}{
		{"běžné dělení", 10, 2, 5, false},
		{"celočíselné zaokrouhlení dolů", 7, 2, 3, false},
		{"záporný dělitel", -9, 3, -3, false},
		{"nula v čitateli", 0, 5, 0, false},
		{"dělení nulou", 10, 0, 0, true},
		{"nula dělená nulou", 0, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exercise.SafeDivide(tt.a, tt.b)
			if tt.wantErr && err == nil {
				t.Fatalf("SafeDivide(%d, %d) vrátil err = nil, chci chybu", tt.a, tt.b)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("SafeDivide(%d, %d) vrátil err = %v, chci nil", tt.a, tt.b, err)
			}
			if got != tt.want {
				t.Errorf("SafeDivide(%d, %d) = %d, chci %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSafeDivideNepropustiPaniku(t *testing.T) {
	// Kdyby recover chyběl, tenhle test shodí celý běh testů.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SafeDivide propustila paniku: %v", r)
		}
	}()
	if _, err := exercise.SafeDivide(1, 0); err == nil {
		t.Error("SafeDivide(1, 0) nevrátila chybu")
	}
}

func TestCloseAll(t *testing.T) {
	t.Run("nil vstup", func(t *testing.T) {
		if err := exercise.CloseAll(nil); err != nil {
			t.Errorf("CloseAll(nil) = %v, chci nil", err)
		}
	})

	t.Run("prázdný vstup", func(t *testing.T) {
		if err := exercise.CloseAll([]func() error{}); err != nil {
			t.Errorf("CloseAll([]) = %v, chci nil", err)
		}
	})

	t.Run("všechno projde", func(t *testing.T) {
		calls := 0
		ok := func() error { calls++; return nil }
		if err := exercise.CloseAll([]func() error{ok, ok, ok}); err != nil {
			t.Errorf("CloseAll = %v, chci nil", err)
		}
		if calls != 3 {
			t.Errorf("zavoláno %d zavíračů, chci 3", calls)
		}
	})

	t.Run("vrací první chybu a zavře všechno", func(t *testing.T) {
		errFirst := errors.New("první chyba")
		errSecond := errors.New("druhá chyba")
		calls := 0

		closers := []func() error{
			func() error { calls++; return nil },
			func() error { calls++; return errFirst },
			func() error { calls++; return errSecond },
			func() error { calls++; return nil },
		}

		err := exercise.CloseAll(closers)
		if !errors.Is(err, errFirst) {
			t.Errorf("CloseAll vrátil %v, chci %v", err, errFirst)
		}
		if calls != 4 {
			t.Errorf("zavoláno %d zavíračů, chci 4 — zavírat se musí i po chybě", calls)
		}
	})

	t.Run("přeskočí nil položky", func(t *testing.T) {
		calls := 0
		closers := []func() error{
			nil,
			func() error { calls++; return nil },
			nil,
		}
		if err := exercise.CloseAll(closers); err != nil {
			t.Errorf("CloseAll = %v, chci nil", err)
		}
		if calls != 1 {
			t.Errorf("zavoláno %d zavíračů, chci 1", calls)
		}
	})
}

func TestStackPushPop(t *testing.T) {
	var s exercise.Stack
	if got := s.Len(); got != 0 {
		t.Errorf("Len() nového zásobníku = %d, chci 0", got)
	}

	s.Push(1)
	s.Push(2)
	s.Push(3)
	if got := s.Len(); got != 3 {
		t.Errorf("Len() = %d, chci 3", got)
	}

	for _, want := range []int{3, 2, 1} {
		if got := s.Pop(); got != want {
			t.Errorf("Pop() = %d, chci %d", got, want)
		}
	}
	if got := s.Len(); got != 0 {
		t.Errorf("Len() po vyprázdnění = %d, chci 0", got)
	}
}

func TestStackLenNaNilPointeru(t *testing.T) {
	var s *exercise.Stack
	if got := s.Len(); got != 0 {
		t.Errorf("Len() na nil pointeru = %d, chci 0", got)
	}
}

func TestStackPopNaPrazdnemPanikuje(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Pop() na prázdném zásobníku nepanikoval")
		}
		if got, ok := r.(string); !ok || got != "pop from empty stack" {
			t.Errorf("panika s hodnotou %v, chci %q", r, "pop from empty stack")
		}
	}()

	var s exercise.Stack
	s.Pop()
}

func TestTryPop(t *testing.T) {
	var s exercise.Stack
	s.Push(42)

	v, ok := exercise.TryPop(&s)
	if !ok || v != 42 {
		t.Errorf("TryPop = (%d, %v), chci (42, true)", v, ok)
	}
}

func TestTryPopNaPrazdnem(t *testing.T) {
	var s exercise.Stack
	v, ok := exercise.TryPop(&s)
	if ok {
		t.Error("TryPop na prázdném zásobníku vrátil ok = true, chci false")
	}
	if v != 0 {
		t.Errorf("TryPop na prázdném zásobníku vrátil v = %d, chci 0", v)
	}
}

func TestTryPopNaNilPointeru(t *testing.T) {
	// Nil dereference je taky panika a recover ji musí pobrat.
	v, ok := exercise.TryPop(nil)
	if ok || v != 0 {
		t.Errorf("TryPop(nil) = (%d, %v), chci (0, false)", v, ok)
	}
}

func TestZasobnikJePoRecoveruPouzitelny(t *testing.T) {
	var s exercise.Stack

	if _, ok := exercise.TryPop(&s); ok {
		t.Fatal("TryPop na prázdném zásobníku vrátil ok = true")
	}

	s.Push(7)
	s.Push(8)
	if got := s.Len(); got != 2 {
		t.Fatalf("Len() po zotavení = %d, chci 2", got)
	}

	if v, ok := exercise.TryPop(&s); !ok || v != 8 {
		t.Errorf("TryPop po zotavení = (%d, %v), chci (8, true)", v, ok)
	}
	if got := s.Pop(); got != 7 {
		t.Errorf("Pop() po zotavení = %d, chci 7", got)
	}
	if got := s.Len(); got != 0 {
		t.Errorf("Len() = %d, chci 0", got)
	}
}

func TestStackStridaniOperaci(t *testing.T) {
	var s exercise.Stack
	for i := 0; i < 100; i++ {
		s.Push(i)
	}
	for i := 0; i < 50; i++ {
		if _, ok := exercise.TryPop(&s); !ok {
			t.Fatalf("TryPop selhal na %d. iteraci, ačkoli zásobník není prázdný", i)
		}
	}
	if got := s.Len(); got != 50 {
		t.Fatalf("Len() = %d, chci 50", got)
	}
	if got := s.Pop(); got != 49 {
		t.Errorf("Pop() = %d, chci 49", got)
	}
}
