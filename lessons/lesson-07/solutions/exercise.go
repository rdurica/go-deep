// Package solutions obsahuje referenční řešení lekce 07.
package solutions

// Sum vrací součet prvků. Prázdný i nil vstup dá 0.
func Sum(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Reverse otočí pořadí prvků in-place.
func Reverse(nums []int) {
	for i, j := 0, len(nums)-1; i < j; i, j = i+1, j-1 {
		nums[i], nums[j] = nums[j], nums[i]
	}
}

// Grow zajistí, že výsledek má cap >= n, a zachová len i obsah.
// Pokud už cap(s) >= n, vrací s beze změny.
func Grow(s []int, n int) []int {
	if cap(s) >= n {
		return s
	}
	grown := make([]int, len(s), n)
	copy(grown, s)
	return grown
}

// RemoveAt smaže prvek na indexu i a zachová pořadí zbytku.
// Index mimo rozsah vrací s beze změny.
func RemoveAt(s []int, i int) []int {
	if i < 0 || i >= len(s) {
		return s
	}
	return append(s[:i], s[i+1:]...)
}

// RemoveFast smaže prvek na indexu i v O(1) bez zachování pořadí.
// Index mimo rozsah vrací s beze změny.
func RemoveFast(s []int, i int) []int {
	if i < 0 || i >= len(s) {
		return s
	}
	last := len(s) - 1
	s[i] = s[last]
	return s[:last]
}

// Clone vrací nezávislou kopii s. Nil vstup vrací nil.
func Clone(s []int) []int {
	if s == nil {
		return nil
	}
	out := make([]int, len(s))
	copy(out, s)
	return out
}

// Chunk rozdělí s na nezávislé kopie délky size; poslední kus může být kratší.
// Pro size <= 0 nebo prázdný vstup vrací výsledek nulové délky.
func Chunk(s []int, size int) [][]int {
	if size <= 0 || len(s) == 0 {
		return nil
	}
	out := make([][]int, 0, (len(s)+size-1)/size)
	for start := 0; start < len(s); start += size {
		end := start + size
		if end > len(s) {
			end = len(s)
		}
		// Bez kopie by chunky sdílely backing pole se vstupem.
		out = append(out, Clone(s[start:end]))
	}
	return out
}

// Filter vrací prvky, pro které keep vrátí true. Používá trik s[:0],
// takže nealokuje a přepíše vstupní slice.
func Filter(s []int, keep func(int) bool) []int {
	out := s[:0]
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}
