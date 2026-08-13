package exercise_test

import (
	"slices"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-15/exercise"
)

// Celsius a UserID jsou vlastní pojmenované typy. Do constraintu Number se
// vejdou jen díky vlnovce (~float64, ~int).
type (
	Celsius float64
	UserID  int
)

func TestMax(t *testing.T) {
	if got, ok := exercise.Max([]int{3, 9, 2}); got != 9 || !ok {
		t.Errorf("Max([]int) = (%d, %v), chci (9, true)", got, ok)
	}
	if got, ok := exercise.Max([]string{"b", "a", "c"}); got != "c" || !ok {
		t.Errorf("Max([]string) = (%q, %v), chci (\"c\", true)", got, ok)
	}
	if got, ok := exercise.Max([]float64{-1.5, -0.5}); got != -0.5 || !ok {
		t.Errorf("Max([]float64) = (%v, %v), chci (-0.5, true)", got, ok)
	}
	if got, ok := exercise.Max([]Celsius{1, 2}); got != Celsius(2) || !ok {
		t.Errorf("Max([]Celsius) = (%v, %v), chci (2, true)", got, ok)
	}
	if got, ok := exercise.Max([]int{}); got != 0 || ok {
		t.Errorf("Max([]) = (%d, %v), chci (0, false)", got, ok)
	}
	if got, ok := exercise.Max([]string(nil)); got != "" || ok {
		t.Errorf("Max(nil) = (%q, %v), chci (\"\", false)", got, ok)
	}
}

func TestFilter(t *testing.T) {
	t.Run("even numbers", func(t *testing.T) {
		got := exercise.Filter([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 })
		want := []int{2, 4, 6}
		if !slices.Equal(got, want) {
			t.Errorf("Filter() = %v, chci %v", got, want)
		}
	})

	t.Run("strings with prefix", func(t *testing.T) {
		in := []string{"go", "php", "golang", "rust"}
		got := exercise.Filter(in, func(s string) bool { return strings.HasPrefix(s, "go") })
		want := []string{"go", "golang"}
		if !slices.Equal(got, want) {
			t.Errorf("Filter() = %v, chci %v", got, want)
		}
	})

	t.Run("nothing passes", func(t *testing.T) {
		got := exercise.Filter([]int{1, 3}, func(n int) bool { return n%2 == 0 })
		if len(got) != 0 {
			t.Errorf("Filter() = %v, chci prázdné", got)
		}
	})

	t.Run("nil input", func(t *testing.T) {
		got := exercise.Filter(nil, func(n int) bool { return true })
		if len(got) != 0 {
			t.Errorf("Filter(nil) = %v, chci prázdné", got)
		}
	})
}

func TestSum(t *testing.T) {
	if got := exercise.Sum([]int{1, 2, 3}); got != 6 {
		t.Errorf("Sum([]int) = %d, chci 6", got)
	}
	if got := exercise.Sum([]int64{1 << 40, 1}); got != (1<<40)+1 {
		t.Errorf("Sum([]int64) = %d, chci %d", got, int64(1<<40)+1)
	}
	if got := exercise.Sum([]float64{0.5, 0.25}); got != 0.75 {
		t.Errorf("Sum([]float64) = %v, chci 0.75", got)
	}
	if got := exercise.Sum([]int(nil)); got != 0 {
		t.Errorf("Sum(nil) = %d, chci 0", got)
	}
}

func TestSumWithCustomType(t *testing.T) {
	temps := []Celsius{21.5, -3.5, 2}
	if got := exercise.Sum(temps); got != Celsius(20) {
		t.Errorf("Sum([]Celsius) = %v, chci 20", got)
	}

	ids := []UserID{1, 2, 3}
	if got := exercise.Sum(ids); got != UserID(6) {
		t.Errorf("Sum([]UserID) = %v, chci 6", got)
	}

	var _ Celsius = exercise.Sum(temps)
	var _ UserID = exercise.Sum(ids)
}

func TestStackInt(t *testing.T) {
	var s exercise.Stack[int]

	if _, ok := s.Pop(); ok {
		t.Error("Pop() na prázdném zásobníku vrátil ok=true")
	}

	s.Push(1)
	s.Push(2)
	s.Push(3)

	for _, want := range []int{3, 2, 1} {
		got, ok := s.Pop()
		if !ok || got != want {
			t.Errorf("Pop() = (%d, %v), chci (%d, true)", got, ok, want)
		}
	}
	if _, ok := s.Pop(); ok {
		t.Error("Pop() po vyprázdnění vrátil ok=true")
	}
}

func TestStackOtherType(t *testing.T) {
	type point struct{ X, Y int }

	var s exercise.Stack[point]
	s.Push(point{1, 2})
	s.Push(point{3, 4})

	if got, ok := s.Pop(); !ok || got != (point{3, 4}) {
		t.Errorf("Pop() = (%v, %v), chci ({3 4}, true)", got, ok)
	}

	var strs exercise.Stack[string]
	strs.Push("a")
	if got, ok := strs.Pop(); !ok || got != "a" {
		t.Errorf("Pop() = (%q, %v), chci (\"a\", true)", got, ok)
	}
}
