// Package solutions obsahuje referenční řešení lekce 04.
package solutions

// --- Stupeň: jednoduchý ---
// Sum sečte libovolný počet čísel. Bez argumentů vrací 0.
func Sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// MinMax vrátí nejmenší a největší prvek slice.
// Pro prázdný nebo nil vstup vrací (0, 0, false).
func MinMax(nums []int) (min, max int, ok bool) {
	if len(nums) == 0 {
		return 0, 0, false
	}
	min, max = nums[0], nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return min, max, true
}

// --- Stupeň: střední ---
// Counter vrací funkci, která při každém zavolání vrátí o jedna víc.
// První volání vrátí 1.
func Counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

// Apply vrátí nový slice, ve kterém je na každý prvek použita funkce f.
// Vstupní slice se nemění.
func Apply(nums []int, f func(int) int) []int {
	out := make([]int, len(nums))
	for i, n := range nums {
		out[i] = f(n)
	}
	return out
}

// --- Stupeň: obtížný ---
// Compose složí funkce zleva doprava: Compose(f, g)(x) == g(f(x)).
// Bez argumentů vrací identitu.
func Compose(fs ...func(int) int) func(int) int {
	return func(x int) int {
		for _, f := range fs {
			x = f(x)
		}
		return x
	}
}

// Memoize vrátí memoizovanou variantu funkce f a funkci, která hlásí,
// kolikrát byla f skutečně zavolána.
func Memoize(f func(int) int) (func(int) int, func() int) {
	cache := make(map[int]int)
	calls := 0

	memoized := func(x int) int {
		if v, ok := cache[x]; ok {
			return v
		}
		calls++
		v := f(x)
		cache[x] = v
		return v
	}
	return memoized, func() int { return calls }
}
