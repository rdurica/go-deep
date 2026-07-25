// Package exercise obsahuje cvičení lekce 15.
package exercise

import "cmp"

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
	panic("TODO: úkol A")
}

// Filter vrací prvky, pro které keep vrátí true.
func Filter[T any](s []T, keep func(T) bool) []T {
	panic("TODO: úkol A")
}

// Sum sečte prvky slice. Prázdný vstup dá nulu daného typu.
func Sum[T Number](s []T) T {
	panic("TODO: úkol B")
}

// Max vrací největší prvek a true. Pro prázdný vstup vrací zero value a false.
func Max[T cmp.Ordered](s []T) (T, bool) {
	panic("TODO: úkol B")
}

// Keys vrací klíče mapy v nespecifikovaném pořadí.
func Keys[K comparable, V any](m map[K]V) []K {
	panic("TODO: úkol B")
}

// Push vloží hodnotu na vrchol zásobníku.
func (s *Stack[T]) Push(v T) {
	panic("TODO: úkol C")
}

// Pop odebere a vrátí vrchol. Na prázdném zásobníku vrací zero value a false.
func (s *Stack[T]) Pop() (T, bool) {
	panic("TODO: úkol C")
}

// Peek vrátí vrchol bez odebrání. Na prázdném zásobníku vrací zero value a false.
func (s *Stack[T]) Peek() (T, bool) {
	panic("TODO: úkol C")
}

// Len vrací počet prvků v zásobníku.
func (s *Stack[T]) Len() int {
	panic("TODO: úkol C")
}

// NewCache vytvoří cache pro nejvýš max záznamů. Hodnoty menší než 1 se berou jako 1.
func NewCache[K comparable, V any](max int) *Cache[K, V] {
	panic("TODO: úkol C")
}

// Get vrací hodnotu a true, pokud klíč existuje.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	panic("TODO: úkol C")
}

// Set uloží hodnotu. Při překročení limitu vypadne nejdéle uložený záznam.
func (c *Cache[K, V]) Set(key K, value V) {
	panic("TODO: úkol C")
}

// Len vrací počet uložených záznamů.
func (c *Cache[K, V]) Len() int {
	panic("TODO: úkol C")
}
