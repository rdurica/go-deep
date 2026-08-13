// Package exercise obsahuje cvičení lekce 10.
package exercise

import (
	"fmt"
	"io"
)

// --- Stupeň: jednoduchý ---

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
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Defer nepřiřadí výsledek appendu do pojmenovaného
// návratu steps — používá `_ = append`. Najdi chybu a oprav.
func SumWithLog(nums []int) (total int, steps []string) {
	defer func() {
		_ = append(steps, fmt.Sprintf("total=%d", total))
	}()

	for _, n := range nums {
		total += n
		steps = append(steps, fmt.Sprintf("+%d=%d", n, total))
	}
	return total, steps
}

// --- Stupeň: střední ---

// DeferOrder zaregistruje tři defery, které do pojmenovaného order přidají
// "first", "second", "third" (v tomto pořadí registrace).
// Defery běží LIFO, takže výsledek je ["third", "second", "first"].
// Append musí být uvnitř defer func(){ … }() — pojmenovaný návrat order
// defery ještě po returnu doplní.
func DeferOrder() (order []string) {
	// TODO
	return
}

// SafeDivide vydělí a/b celočíselně (truncace k nule: 7/2 → 3).
// Dělení nulou panikuje — odchytni to recover v deferu a převeď na error
// (fmt.Errorf). Při chybě result je 0. Nesmíš dělitele testovat předem
// přes if b == 0 — paniku musí způsobit samotné a/b.
func SafeDivide(a, b int) (result int, err error) {
	// TODO
	return
}

// --- Stupeň: obtížný ---

// WriteAndClose zapíše data přes w.Write a vždy zavolá w.Close.
// Použij pojmenovaný návrat (err error) a defer: když Close selže a Write
// uspěl (err == nil), přiřaď chybu Close do err. Nesmíš defer w.Close()
// bez kontroly návratu — chyba z Close by se ztratila.
// Když Write selže, Close se stejně zavolá, ale návratová chyba zůstane
// z Write. w == nil → vrať fmt.Errorf (Close nevolej).
func WriteAndClose(w io.WriteCloser, data []byte) (err error) {
	// TODO
	return
}
