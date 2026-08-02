package solutions_test

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-15/solutions"
)

// Celsius a UserID jsou vlastní pojmenované typy. Do constraintu Number se
// vejdou jen díky vlnovce (~float64, ~int).
type (
	Celsius float64
	UserID  int
)

func TestMap(t *testing.T) {
	t.Run("int to string", func(t *testing.T) {
		got := exercise.Map([]int{1, 2, 3}, strconv.Itoa)
		want := []string{"1", "2", "3"}
		if !slices.Equal(got, want) {
			t.Errorf("Map() = %v, chci %v", got, want)
		}
	})

	t.Run("string to length", func(t *testing.T) {
		got := exercise.Map([]string{"a", "bb", ""}, func(s string) int { return len(s) })
		want := []int{1, 2, 0}
		if !slices.Equal(got, want) {
			t.Errorf("Map() = %v, chci %v", got, want)
		}
	})

	t.Run("custom type", func(t *testing.T) {
		got := exercise.Map([]UserID{1, 2}, func(id UserID) string {
			return "u" + strconv.Itoa(int(id))
		})
		want := []string{"u1", "u2"}
		if !slices.Equal(got, want) {
			t.Errorf("Map() = %v, chci %v", got, want)
		}
	})

	t.Run("empty and nil input", func(t *testing.T) {
		if got := exercise.Map([]int{}, strconv.Itoa); len(got) != 0 {
			t.Errorf("Map([]) = %v, chci prázdné", got)
		}
		if got := exercise.Map(nil, strconv.Itoa); len(got) != 0 {
			t.Errorf("Map(nil) = %v, chci prázdné", got)
		}
	})
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

// TestMapFilterComposition ukazuje, proč Map potřebuje dva type parametry.
func TestMapFilterComposition(t *testing.T) {
	words := []string{"go", "php", "golang", "c"}
	long := exercise.Filter(words, func(s string) bool { return len(s) > 2 })
	lengths := exercise.Map(long, func(s string) int { return len(s) })

	want := []int{3, 6}
	if !slices.Equal(lengths, want) {
		t.Errorf("složení Filter+Map = %v, chci %v", lengths, want)
	}
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

// TestSumWithCustomType projde jen díky vlnovce v constraintu Number.
// Bez ní by kompilátor odmítl Celsius i UserID.
func TestSumWithCustomType(t *testing.T) {
	temps := []Celsius{21.5, -3.5, 2}
	if got := exercise.Sum(temps); got != Celsius(20) {
		t.Errorf("Sum([]Celsius) = %v, chci 20", got)
	}

	ids := []UserID{1, 2, 3}
	if got := exercise.Sum(ids); got != UserID(6) {
		t.Errorf("Sum([]UserID) = %v, chci 6", got)
	}

	// Návratový typ si drží pojmenovaný typ, ne jen podkladový.
	var _ Celsius = exercise.Sum(temps)
	var _ UserID = exercise.Sum(ids)
}

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

func TestKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	got := exercise.Keys(m)
	slices.Sort(got) // pořadí iterace mapy není definované

	if want := []string{"a", "b", "c"}; !slices.Equal(got, want) {
		t.Errorf("Keys() = %v, chci %v", got, want)
	}

	ids := exercise.Keys(map[UserID]string{7: "radek"})
	if !slices.Equal(ids, []UserID{7}) {
		t.Errorf("Keys(map[UserID]string) = %v, chci [7]", ids)
	}

	if got := exercise.Keys(map[string]int{}); len(got) != 0 {
		t.Errorf("Keys(prázdná mapa) = %v, chci prázdné", got)
	}
	if got := exercise.Keys(map[string]int(nil)); len(got) != 0 {
		t.Errorf("Keys(nil) = %v, chci prázdné", got)
	}
}

func TestStackInt(t *testing.T) {
	var s exercise.Stack[int] // zero value musí být použitelná

	if s.Len() != 0 {
		t.Errorf("Len() na prázdném zásobníku = %d, chci 0", s.Len())
	}
	if _, ok := s.Pop(); ok {
		t.Error("Pop() na prázdném zásobníku vrátil ok=true")
	}
	if _, ok := s.Peek(); ok {
		t.Error("Peek() na prázdném zásobníku vrátil ok=true")
	}

	s.Push(1)
	s.Push(2)
	s.Push(3)
	if s.Len() != 3 {
		t.Errorf("Len() = %d, chci 3", s.Len())
	}

	if got, ok := s.Peek(); got != 3 || !ok {
		t.Errorf("Peek() = (%d, %v), chci (3, true)", got, ok)
	}
	if s.Len() != 3 {
		t.Errorf("Peek() změnil délku na %d, chci 3", s.Len())
	}

	for _, want := range []int{3, 2, 1} {
		got, ok := s.Pop()
		if !ok || got != want {
			t.Errorf("Pop() = (%d, %v), chci (%d, true)", got, ok, want)
		}
	}
	if s.Len() != 0 {
		t.Errorf("Len() po vyprázdnění = %d, chci 0", s.Len())
	}
}

// TestStackOtherType instancuje ten samý typ podruhé, tentokrát strukturou.
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
	if got, ok := strs.Pop(); ok || got != "" {
		t.Errorf("Pop() po vyprázdnění = (%q, %v), chci (\"\", false)", got, ok)
	}
}

func TestCache(t *testing.T) {
	c := exercise.NewCache[string, int](2)

	if c.Len() != 0 {
		t.Errorf("Len() = %d, chci 0", c.Len())
	}
	if _, ok := c.Get("chybí"); ok {
		t.Error("Get() na prázdné cache vrátil ok=true")
	}

	c.Set("a", 1)
	c.Set("b", 2)
	if c.Len() != 2 {
		t.Errorf("Len() = %d, chci 2", c.Len())
	}

	c.Set("c", 3) // "a" musí vypadnout
	if c.Len() != 2 {
		t.Errorf("Len() = %d, chci 2 — limit se nesmí překročit", c.Len())
	}
	if _, ok := c.Get("a"); ok {
		t.Error("nejstarší záznam \"a\" měl vypadnout, ale je pořád v cache")
	}
	for key, want := range map[string]int{"b": 2, "c": 3} {
		got, ok := c.Get(key)
		if !ok || got != want {
			t.Errorf("Get(%q) = (%d, %v), chci (%d, true)", key, got, ok, want)
		}
	}
}

func TestCacheOverwritePreservesOrder(t *testing.T) {
	c := exercise.NewCache[string, string](2)
	c.Set("a", "first")
	c.Set("b", "druhá")
	c.Set("a", "aktualizovaná") // jen přepis, "a" zůstává nejstarší

	if c.Len() != 2 {
		t.Fatalf("Len() = %d, chci 2", c.Len())
	}
	if got, ok := c.Get("a"); !ok || got != "aktualizovaná" {
		t.Errorf("Get(\"a\") = (%q, %v), chci (\"aktualizovaná\", true)", got, ok)
	}

	c.Set("c", "třetí") // vypadnout má "a", ne "b"
	if _, ok := c.Get("a"); ok {
		t.Error("po přeplnění měla vypadnout \"a\"")
	}
	if _, ok := c.Get("b"); !ok {
		t.Error("\"b\" neměla vypadnout")
	}
}

// TestCacheWithCustomKey instancuje Cache podruhé, jiným klíčem i hodnotou.
func TestCacheWithCustomKey(t *testing.T) {
	type profile struct{ Name string }

	c := exercise.NewCache[UserID, profile](1)
	c.Set(1, profile{Name: "radek"})

	got, ok := c.Get(1)
	if !ok || got.Name != "radek" {
		t.Errorf("Get(1) = (%v, %v), chci ({radek}, true)", got, ok)
	}

	c.Set(2, profile{Name: "jana"})
	if _, ok := c.Get(1); ok {
		t.Error("při limitu 1 měl první záznam vypadnout")
	}
	if c.Len() != 1 {
		t.Errorf("Len() = %d, chci 1", c.Len())
	}
}

func TestCacheMinimalSize(t *testing.T) {
	for _, max := range []int{0, -5} {
		c := exercise.NewCache[string, int](max)
		c.Set("a", 1)
		c.Set("b", 2)

		if c.Len() != 1 {
			t.Errorf("NewCache(%d): Len() = %d, chci 1", max, c.Len())
		}
		if _, ok := c.Get("b"); !ok {
			t.Errorf("NewCache(%d): poslední vložený záznam chybí", max)
		}
	}
}
