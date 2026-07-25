// Package exercise obsahuje cvičení lekce 07.
package exercise

// Sum vrací součet prvků. Prázdný i nil vstup dá 0.
func Sum(nums []int) int {
	panic("TODO: úkol A")
}

// Reverse otočí pořadí prvků in-place.
func Reverse(nums []int) {
	panic("TODO: úkol A")
}

// Grow zajistí, že výsledek má cap >= n, a zachová len i obsah.
// Pokud už cap(s) >= n, vrací s beze změny.
func Grow(s []int, n int) []int {
	panic("TODO: úkol B")
}

// RemoveAt smaže prvek na indexu i a zachová pořadí zbytku.
// Index mimo rozsah vrací s beze změny.
func RemoveAt(s []int, i int) []int {
	panic("TODO: úkol B")
}

// RemoveFast smaže prvek na indexu i v O(1) bez zachování pořadí.
// Index mimo rozsah vrací s beze změny.
func RemoveFast(s []int, i int) []int {
	panic("TODO: úkol B")
}

// Clone vrací nezávislou kopii s. Nil vstup vrací nil.
func Clone(s []int) []int {
	panic("TODO: úkol B")
}

// Chunk rozdělí s na nezávislé kopie délky size; poslední kus může být kratší.
// Pro size <= 0 nebo prázdný vstup vrací výsledek nulové délky.
func Chunk(s []int, size int) [][]int {
	panic("TODO: úkol C")
}

// Filter vrací prvky, pro které keep vrátí true. Implementuj trikem s[:0],
// tedy bez alokace a s vědomím, že přepíšeš vstupní slice.
func Filter(s []int, keep func(int) bool) []int {
	panic("TODO: úkol C")
}
