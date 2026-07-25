// Package solutions obsahuje referenční řešení lekce 10.
package solutions

import "fmt"

// Stack je zásobník celých čísel.
type Stack struct {
	items []int
}

// DeferOrder registruje tři defery zapisující do výsledku a vrací jejich
// skutečné pořadí provedení.
func DeferOrder() (order []string) {
	// Defery se provedou v opačném pořadí registrace (LIFO) a protože je
	// návratová hodnota pojmenovaná, můžou ji ještě po returnu doplnit.
	defer func() { order = append(order, "first") }()
	defer func() { order = append(order, "second") }()
	defer func() { order = append(order, "third") }()
	return order
}

// SumWithLog sečte čísla a vede o tom protokol. Poslední krok doplní defer.
func SumWithLog(nums []int) (total int, steps []string) {
	// Uzávěr čte total až v okamžiku provedení, takže vidí konečný součet.
	defer func() {
		steps = append(steps, fmt.Sprintf("total=%d", total))
	}()

	for _, n := range nums {
		total += n
		steps = append(steps, fmt.Sprintf("+%d=%d", n, total))
	}
	return total, steps
}

// SafeDivide vydělí a/b a paniku z dělení nulou převede přes recover na error.
func SafeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = 0
			err = fmt.Errorf("dělení %d/%d selhalo: %v", a, b, r)
		}
	}()
	return a / b, nil
}

// CloseAll zavolá všechny funkce a vrátí první vzniklou chybu.
func CloseAll(closers []func() error) error {
	var first error
	for _, closeFn := range closers {
		if closeFn == nil {
			continue
		}
		// Zavíráme všechno, i po chybě — jinak by zbytek zdrojů zůstal viset.
		if err := closeFn(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// Push vloží prvek navrch zásobníku.
func (s *Stack) Push(v int) {
	s.items = append(s.items, v)
}

// Pop odebere a vrátí vrchní prvek. Nad prázdným zásobníkem paniká.
func (s *Stack) Pop() int {
	if len(s.items) == 0 {
		// Chyba programátora, ne provozní stav — proto panika, ne error.
		panic("pop from empty stack")
	}
	last := len(s.items) - 1
	v := s.items[last]
	s.items = s.items[:last]
	return v
}

// Len vrací počet prvků. Funguje i na nil pointeru.
func (s *Stack) Len() int {
	if s == nil {
		return 0
	}
	return len(s.items)
}

// TryPop je bezpečná varianta Pop: paniku odchytí a vrátí (0, false).
func TryPop(s *Stack) (v int, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			v, ok = 0, false
		}
	}()
	return s.Pop(), true
}
