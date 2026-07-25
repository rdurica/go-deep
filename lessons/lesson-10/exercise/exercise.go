// Package exercise obsahuje cvičení lekce 10.
package exercise

// Stack je zásobník celých čísel. Vnitřní podobu si klidně změň,
// testy sahají jen na metody.
type Stack struct {
	items []int
}

// DeferOrder registruje tři defery zapisující do výsledku a vrací jejich
// skutečné pořadí provedení.
func DeferOrder() (order []string) {
	panic("TODO: úkol A")
}

// SumWithLog sečte čísla a vede o tom protokol. Poslední krok doplní defer.
func SumWithLog(nums []int) (total int, steps []string) {
	panic("TODO: úkol B")
}

// SafeDivide vydělí a/b a paniku z dělení nulou převede přes recover na error.
func SafeDivide(a, b int) (result int, err error) {
	panic("TODO: úkol B")
}

// CloseAll zavolá všechny funkce a vrátí první vzniklou chybu.
func CloseAll(closers []func() error) error {
	panic("TODO: úkol B")
}

// Push vloží prvek navrch zásobníku.
func (s *Stack) Push(v int) {
	panic("TODO: úkol C")
}

// Pop odebere a vrátí vrchní prvek. Nad prázdným zásobníkem paniká.
func (s *Stack) Pop() int {
	panic("TODO: úkol C")
}

// Len vrací počet prvků. Funguje i na nil pointeru.
func (s *Stack) Len() int {
	panic("TODO: úkol C")
}

// TryPop je bezpečná varianta Pop: paniku odchytí a vrátí (0, false).
func TryPop(s *Stack) (v int, ok bool) {
	panic("TODO: úkol C")
}
