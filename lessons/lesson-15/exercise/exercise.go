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

// Map vrátí nový slice s výsledky f pro každý prvek vstupu.
// Zachová pořadí i délku. Nil vstup dá prázdný výsledek (ne nil slice).
func Map[T, U any](s []T, f func(T) U) []U {
	// TODO
	return nil
}

// Filter vrací prvky, pro které keep vrátí true, v původním pořadí.
// Nil vstup dá prázdný výsledek. Pořadí odpovídá vstupu.
func Filter[T any](s []T, keep func(T) bool) []T {
	// TODO
	return nil
}

// Sum sečte prvky slice typu Number (včetně pojmenovaných typů jako Celsius).
// Prázdný nebo nil vstup dá nulu daného typu, ne jen int 0 u float slice.
func Sum[T Number](s []T) T {
	// TODO
	return *new(T)
}

// Max vrací největší prvek a true. Prázdný vstup → (zero value, false).
// Prázdný vstup nesmí panikovat.
func Max[T cmp.Ordered](s []T) (T, bool) {
	// TODO
	return *new(T), false
}

// Keys vrací klíče mapy v nespecifikovaném pořadí (test si výsledek seřadí).
// Nil mapa dá prázdný výsledek. K musí být comparable, V stačí any.
func Keys[K comparable, V any](m map[K]V) []K {
	// TODO
	return nil
}

// --- Stupeň: jednoduchý ---
// Push vloží hodnotu na vrchol zásobníku (LIFO).
// Zero value var s Stack[int] je použitelný bez konstruktoru.
func (s *Stack[T]) Push(v T) {
	// TODO
}

// Pop odebere a vrátí vrchol. Na prázdném zásobníku vrací zero value a false.
// Uvolněné místo přepiš zero value, aby zásobník nedržel referenci.
func (s *Stack[T]) Pop() (T, bool) {
	// TODO
	return *new(T), false
}

// --- Stupeň: střední ---
// Peek vrátí vrchol bez odebrání. Na prázdném zásobníku zero value a false.
// Zásobník se po Peek nemění (na rozdíl od Pop).
func (s *Stack[T]) Peek() (T, bool) {
	// TODO
	return *new(T), false
}

// --- Stupeň: obtížný ---
// Len vrací počet prvků v zásobníku. Na prázdném zásobníku vrací 0.
func (s *Stack[T]) Len() int {
	// TODO
	return 0
}

// NewCache vytvoří cache pro nejvýš max záznamů.
// Hodnoty max < 1 se berou jako 1. Při přeplnění vypadne nejstarší záznam (FIFO).
func NewCache[K comparable, V any](max int) *Cache[K, V] {
	// TODO
	return nil
}

// Get vrací hodnotu a true, pokud klíč existuje.
// Neexistující klíč vrací zero value a false.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	// TODO
	return *new(V), false
}

// Set uloží hodnotu. Při překročení limitu vypadne nejdéle uložený záznam (FIFO).
// Přepis existujícího klíče jen změní hodnotu, pořadí nemění.
func (c *Cache[K, V]) Set(key K, value V) {
	// TODO
}

// Len vrací počet aktuálně uložených záznamů v cache.
func (c *Cache[K, V]) Len() int {
	// TODO
	return 0
}
