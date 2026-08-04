// Package solutions obsahuje referenční řešení lekce 10.
package solutions

import (
	"fmt"
	"io"
)

// Stack je zásobník celých čísel.
type Stack struct {
	items []int
}

// --- Stupeň: jednoduchý ---
// DeferOrder zaregistruje tři defery, které do pojmenovaného order přidají
// "first", "second", "third" (v tomto pořadí registrace).
// Defery běží LIFO, takže výsledek je ["third", "second", "first"].
// Append musí být uvnitř defer func(){ … }() — pojmenovaný návrat order
// defery ještě po returnu doplní.
func DeferOrder() (order []string) {
	defer func() { order = append(order, "first") }()
	defer func() { order = append(order, "second") }()
	defer func() { order = append(order, "third") }()
	return order
}

// SumWithLog sečte čísla a vede protokol kroků ve tvaru "+N=součet".
// Poslední krok "total=<součet>" doplň deferem před cyklem (named return).
//
//	SumWithLog(nil)        → 0, ["total=0"]
//	SumWithLog([]int{5})   → 5, ["+5=5", "total=5"]
//	SumWithLog([]int{1,2,3}) → 6, ["+1=1", "+2=3", "+3=6", "total=6"]
//	SumWithLog([]int{10,-4}) → 6, ["+10=10", "+-4=6", "total=6"]
//
// Formát kroku je fmt.Sprintf("+%d=%d", n, total) — u záporných tedy "+-4=6".
// total=… nesmíš appendovat v těle funkce, jen v deferu.
func SumWithLog(nums []int) (total int, steps []string) {
	defer func() {
		steps = append(steps, fmt.Sprintf("total=%d", total))
	}()

	for _, n := range nums {
		total += n
		steps = append(steps, fmt.Sprintf("+%d=%d", n, total))
	}
	return total, steps
}

// --- Stupeň: střední ---
// SafeDivide vydělí a/b celočíselně (truncace k nule: 7/2 → 3).
// Dělení nulou panikuje — odchytni to recover v deferu a převeď na error
// (fmt.Errorf). Při chybě result je 0. Nesmíš dělitele testovat předem
// přes if b == 0 — paniku musí způsobit samotné a/b.
func SafeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = 0
			err = fmt.Errorf("dělení %d/%d selhalo: %v", a, b, r)
		}
	}()
	return a / b, nil
}

// WriteAndClose zapíše data přes w.Write a vždy zavolá w.Close.
// Použij pojmenovaný návrat (err error) a defer: když Close selže a Write
// uspěl (err == nil), přiřaď chybu Close do err. Nesmíš defer w.Close()
// bez kontroly návratu — chyba z Close by se ztratila.
// Když Write selže, Close se stejně zavolá, ale návratová chyba zůstane
// z Write. w == nil → vrať fmt.Errorf (Close nevolej).
func WriteAndClose(w io.WriteCloser, data []byte) (err error) {
	if w == nil {
		return fmt.Errorf("WriteAndClose: nil writer")
	}
	defer func() {
		if cerr := w.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	_, err = w.Write(data)
	return err
}

// Push vloží prvek navrch zásobníku (na konec vnitřního slice).
func (s *Stack) Push(v int) {
	s.items = append(s.items, v)
}

// Pop odebere a vrátí vrchní prvek.
// Nad prázdným zásobníkem paniká s hodnotou "pop from empty stack"
// (chyba programátora, ne error).
func (s *Stack) Pop() int {
	if len(s.items) == 0 {
		panic("pop from empty stack")
	}
	last := len(s.items) - 1
	v := s.items[last]
	s.items = s.items[:last]
	return v
}

// --- Stupeň: obtížný ---
// Len vrací počet prvků. Musí fungovat i na nil pointeru (*Stack)(nil) → 0.
func (s *Stack) Len() int {
	if s == nil {
		return 0
	}
	return len(s.items)
}

// TryPop je bezpečná varianta: zavolá Pop a paniku odchytí přes recover.
// Při panice (prázdný stack i nil pointer) vrací 0, false.
// Po neúspěšném TryPop musí Push/Pop dál fungovat.
func TryPop(s *Stack) (v int, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			v, ok = 0, false
		}
	}()
	return s.Pop(), true
}
