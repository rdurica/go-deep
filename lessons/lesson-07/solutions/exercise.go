// Package solutions obsahuje referenční řešení lekce 07.
package solutions

// --- Stupeň: jednoduchý ---

// Grow zajistí, že výsledek má cap >= n, a zachová len i obsah.
func Grow(s []int, n int) []int {
	if cap(s) >= n {
		return s
	}
	grown := make([]int, len(s), n)
	copy(grown, s)
	return grown
}

// --- Stupeň: střední ---

// Sum vrací součet prvků. Prázdný i nil vstup dá 0.
func Sum(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// RemoveAt smaže prvek na indexu i a zachová pořadí zbytku.
func RemoveAt(s []int, i int) []int {
	if i < 0 || i >= len(s) {
		return s
	}
	return append(s[:i], s[i+1:]...)
}

// --- Stupeň: obtížný ---

// Clone vrací nezávislou kopii s. Nil vstup vrací nil.
func Clone(s []int) []int {
	if s == nil {
		return nil
	}
	out := make([]int, len(s))
	copy(out, s)
	return out
}
