// Package solutions obsahuje referenční řešení lekce 15.
package solutions

import "cmp"

// Number je constraint pro číselné typy včetně těch pojmenovaných.
type Number interface {
	~int | ~int64 | ~float64
}

// Stack je zásobník (LIFO) libovolných hodnot.
type Stack[T any] struct {
	items []T
}

// --- Stupeň: jednoduchý ---

// Max vrací největší prvek a true. Pro prázdný vstup vrací zero value a false.
func Max[T cmp.Ordered](s []T) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}
	best := s[0]
	for _, v := range s[1:] {
		if v > best {
			best = v
		}
	}
	return best, true
}

// --- Stupeň: střední ---

// Filter vrací prvky, pro které keep vrátí true.
func Filter[T any](s []T, keep func(T) bool) []T {
	if s == nil {
		return []T{}
	}
	out := make([]T, 0, len(s))
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// Sum sečte prvky slice.
func Sum[T Number](s []T) T {
	var total T
	for _, v := range s {
		total += v
	}
	return total
}

// --- Stupeň: obtížný ---

// Push vloží hodnotu na vrchol zásobníku.
func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

// Pop odebere a vrátí vrchol. Na prázdném zásobníku vrací zero value a false.
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	last := len(s.items) - 1
	v := s.items[last]

	var zero T
	s.items[last] = zero
	s.items = s.items[:last]

	return v, true
}
