// Package exercise obsahuje cvičení lekce 07.
package exercise

// Sum vrací součet prvků. Prázdný i nil vstup dá 0.
func Sum(nums []int) int {
	// TODO: úkol A
	return 0
}

// Reverse otočí pořadí prvků in-place.
func Reverse(nums []int) {
	// TODO: úkol A
}

// Grow zajistí, že výsledek má cap >= n, a zachová len i obsah.
// Pokud už cap(s) >= n, vrací s beze změny.
func Grow(s []int, n int) []int {
	// TODO: úkol B
	return nil
}

// RemoveAt smaže prvek na indexu i a zachová pořadí zbytku.
// Index mimo rozsah vrací s beze změny.
func RemoveAt(s []int, i int) []int {
	// TODO: úkol B
	return nil
}

// RemoveFast smaže prvek na indexu i v O(1) bez zachování pořadí.
// Index mimo rozsah vrací s beze změny.
func RemoveFast(s []int, i int) []int {
	// TODO: úkol B
	return nil
}

// Clone vrací nezávislou kopii s. Nil vstup vrací nil.
func Clone(s []int) []int {
	// TODO: úkol B
	return nil
}

// Chunk rozdělí s na nezávislé kopie délky size; poslední kus může být kratší.
// Pro size <= 0 nebo prázdný vstup vrací výsledek nulové délky.
func Chunk(s []int, size int) [][]int {
	// TODO: úkol C
	return nil
}

// Filter vrací prvky, pro které keep vrátí true. Implementuj trikem s[:0],
// tedy bez alokace a s vědomím, že přepíšeš vstupní slice.
func Filter(s []int, keep func(int) bool) []int {
	// TODO: úkol C
	return nil
}
