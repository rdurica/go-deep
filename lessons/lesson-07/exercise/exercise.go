// Package exercise obsahuje cvičení lekce 07.
package exercise

// --- Stupeň: jednoduchý ---

// Grow zajistí cap >= n a zachová len i obsah.
// Pokud už cap(s) >= n, vrátí s beze změny (stejný header nad stejným polem).
// Jinak alokuj nové pole a data zkopíruj.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Nové pole ztratí obsah a má špatnou délku.
// Najdi chybu a oprav — testy před opravou padají.
func Grow(s []int, n int) []int {
	if cap(s) >= n {
		return s
	}
	return make([]int, n)
}

// --- Stupeň: střední ---

// Sum vrací součet prvků. Prázdný i nil vstup dá 0.
func Sum(nums []int) int {
	// TODO
	return 0
}

// RemoveAt smaže prvek na indexu i a zachová pořadí zbytku.
// Index mimo rozsah vrací s beze změny. Mutuje backing pole volajícího.
func RemoveAt(s []int, i int) []int {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---

// Clone vrací nezávislou kopii. Nil vstup → nil; prázdný ne-nil → prázdný ne-nil.
func Clone(s []int) []int {
	// TODO
	return nil
}
