// Package exercise obsahuje cvičení lekce 04.
package exercise

// --- Stupeň: jednoduchý ---
// Sum sečte libovolný počet čísel. Bez argumentů vrací 0.
func Sum(nums ...int) int {
	// TODO
	return 0
}

// MinMax vrátí nejmenší a největší prvek slice.
// Pro prázdný nebo nil vstup vrací 0, 0, false. Vstup nemění.
// Signatura má pojmenované návratové hodnoty — v těle piš explicitní return.
func MinMax(nums []int) (min, max int, ok bool) {
	// TODO
	return
}

// --- Stupeň: střední ---
// Counter vrací funkci, která při každém zavolání vrátí o jedna víc (první volání → 1).
// Dva čítače z dvou volání Counter() jsou nezávislé.
func Counter() func() int {
	// TODO
	return nil
}

// Apply vrátí nový slice se stejnou délkou, kde je na každý prvek použita f.
// Vstup se nemění; výsledek nesdílí podkladové pole. Nil/prázdný vstup → slice délky 0.
func Apply(nums []int, f func(int) int) []int {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---
// Compose složí funkce zleva doprava: Compose(f, g)(x) == g(f(x)).
// Bez argumentů vrací identitu.
func Compose(fs ...func(int) int) func(int) int {
	// TODO
	return nil
}

// Memoize vrátí memoizovanou variantu f a funkci, která hlásí počet skutečných volání f.
// Obě vrácené funkce sdílí stejný stav. Každé volání Memoize vytvoří nezávislou instanci.
func Memoize(f func(int) int) (func(int) int, func() int) {
	// TODO
	return nil, nil
}
