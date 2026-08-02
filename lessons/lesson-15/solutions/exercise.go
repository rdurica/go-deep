// Package solutions obsahuje referenční řešení lekce 15.
package solutions

import (
	"cmp"
	"slices"
)

// Number je constraint pro číselné typy včetně těch pojmenovaných.
// Vlnovka znamená "libovolný typ, jehož podkladovým typem je tenhle".
type Number interface {
	~int | ~int64 | ~float64
}

// Stack je zásobník (LIFO) libovolných hodnot.
// Zero value je prázdný a použitelný zásobník.
type Stack[T any] struct {
	items []T
}

// Cache je mapa s omezenou velikostí. Při přeplnění vypadne nejstarší záznam.
type Cache[K comparable, V any] struct {
	max   int
	order []K
	items map[K]V
}

// Map převede každý prvek s na nový typ pomocí f.
func Map[T, U any](s []T, f func(T) U) []U {
	if s == nil {
		return nil
	}
	out := make([]U, 0, len(s))
	for _, v := range s {
		out = append(out, f(v))
	}
	return out
}

// Filter vrací prvky, pro které keep vrátí true.
func Filter[T any](s []T, keep func(T) bool) []T {
	if s == nil {
		return nil
	}
	out := make([]T, 0, len(s))
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// Sum sečte prvky slice. Prázdný vstup dá nulu daného typu.
func Sum[T Number](s []T) T {
	var total T
	for _, v := range s {
		total += v
	}
	return total
}

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

// Keys vrací klíče mapy v nespecifikovaném pořadí.
func Keys[K comparable, V any](m map[K]V) []K {
	if m == nil {
		return nil
	}
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// --- Stupeň: jednoduchý ---
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
	s.items[last] = zero // nedrž referenci na odebranou hodnotu
	s.items = s.items[:last]

	return v, true
}

// --- Stupeň: střední ---
// Peek vrátí vrchol bez odebrání. Na prázdném zásobníku vrací zero value a false.
func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// --- Stupeň: obtížný ---
// Len vrací počet prvků v zásobníku.
func (s *Stack[T]) Len() int {
	return len(s.items)
}

// NewCache vytvoří cache pro nejvýš max záznamů. Hodnoty menší než 1 se berou jako 1.
func NewCache[K comparable, V any](max int) *Cache[K, V] {
	if max < 1 {
		max = 1
	}
	return &Cache[K, V]{
		max:   max,
		order: make([]K, 0, max),
		items: make(map[K]V, max),
	}
}

// Get vrací hodnotu a true, pokud klíč existuje.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	v, ok := c.items[key]
	return v, ok
}

// Set uloží hodnotu. Při překročení limitu vypadne nejdéle uložený záznam.
func (c *Cache[K, V]) Set(key K, value V) {
	if _, exists := c.items[key]; exists {
		c.items[key] = value // přepis pořadí nemění
		return
	}
	if len(c.order) >= c.max {
		oldest := c.order[0]
		c.order = slices.Delete(c.order, 0, 1)
		delete(c.items, oldest)
	}
	c.order = append(c.order, key)
	c.items[key] = value
}

// Len vrací počet uložených záznamů.
func (c *Cache[K, V]) Len() int {
	return len(c.items)
}
