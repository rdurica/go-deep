// Package exercise obsahuje cvičení lekce 10.
package exercise

// Stack je zásobník celých čísel. Vnitřní podobu si klidně změň,
// testy sahají jen na metody.
type Stack struct {
	items []int
}

// --- Stupeň: jednoduchý ---
// DeferOrder registruje tři defery, které přidají "first", "second", "third" do výsledku.
// Vrací skutečné pořadí provedení (LIFO): ["third", "second", "first"].
func DeferOrder() (order []string) {
	// TODO
	return
}

// SumWithLog sečte čísla a vede protokol kroků ve tvaru "+3=6".
// Defer před cyklem přidá na konec "total=<součet>" s finální hodnotou.
// Prázdný vstup: 0 a ["total=0"].
func SumWithLog(nums []int) (total int, steps []string) {
	// TODO
	return
}

// --- Stupeň: střední ---
// SafeDivide vydělí a/b celočíselně.
// Dělení nulou panikuje — odchytni to recover v deferu a převeď na error (fmt.Errorf).
// Při chybě result je 0. Nesmíš dělitele testovat předem přes if b == 0.
func SafeDivide(a, b int) (result int, err error) {
	// TODO
	return
}

// CloseAll zavolá všechny closers, i když některá vrátí chybu, a vrátí první chybu.
// Nil položky přeskočí. Prázdný i nil vstup vrací nil.
func CloseAll(closers []func() error) error {
	// TODO
	return nil
}

// Push vloží prvek navrch zásobníku (na konec vnitřního slice).
func (s *Stack) Push(v int) {
	// TODO
}

// Pop odebere a vrátí vrchní prvek.
// Nad prázdným zásobníkem paniká s hodnotou "pop from empty stack" (chyba programátora).
func (s *Stack) Pop() int {
	// TODO
	return 0
}

// --- Stupeň: obtížný ---
// Len vrací počet prvků. Musí fungovat i na nil pointeru (*Stack)(nil) → 0.
func (s *Stack) Len() int {
	// TODO
	return 0
}

// TryPop je bezpečná varianta: zavolá Pop a paniku odchytí přes recover.
// Při panice vrací 0, false. Po neúspěšném TryPop musí Push/Pop dál fungovat.
func TryPop(s *Stack) (v int, ok bool) {
	// TODO
	return
}
