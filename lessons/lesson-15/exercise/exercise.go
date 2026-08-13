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

// --- Stupeň: jednoduchý ---

// Max vrací největší prvek a true. Prázdný vstup → (zero value, false).
// Prázdný vstup nesmí panikovat.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Na prázdném vstupu panikuje.
// Najdi chybu a oprav — testy před opravou padají.
func Max[T cmp.Ordered](s []T) (T, bool) {
	return s[0], true
}

// --- Stupeň: střední ---

// Filter vrací prvky, pro které keep vrátí true, v původním pořadí.
// Nil vstup dá prázdný výsledek (ne nil slice). Pořadí odpovídá vstupu.
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

// --- Stupeň: obtížný ---

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
