// Package exercise obsahuje cvičení lekce 04.
package exercise

// Sum sečte libovolný počet čísel. Bez argumentů vrací 0.
func Sum(nums ...int) int {
	// TODO: úkol A
	return 0
}

// MinMax vrátí nejmenší a největší prvek slice.
// Pro prázdný nebo nil vstup vrací (0, 0, false).
func MinMax(nums []int) (min, max int, ok bool) {
	// TODO: úkol A
	return
}

// Counter vrací funkci, která při každém zavolání vrátí o jedna víc.
// První volání vrátí 1.
func Counter() func() int {
	// TODO: úkol B
	return nil
}

// Apply vrátí nový slice, ve kterém je na každý prvek použita funkce f.
// Vstupní slice se nemění.
func Apply(nums []int, f func(int) int) []int {
	// TODO: úkol B
	return nil
}

// Compose složí funkce zleva doprava: Compose(f, g)(x) == g(f(x)).
// Bez argumentů vrací identitu.
func Compose(fs ...func(int) int) func(int) int {
	// TODO: úkol B
	return nil
}

// Memoize vrátí memoizovanou variantu funkce f a funkci, která hlásí,
// kolikrát byla f skutečně zavolána.
func Memoize(f func(int) int) (func(int) int, func() int) {
	// TODO: úkol C
	return nil, nil
}
