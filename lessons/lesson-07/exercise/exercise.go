// Package exercise obsahuje cvičení lekce 07.
package exercise

// --- Stupeň: jednoduchý ---
// Sum vrací součet prvků. Prázdný i nil vstup dá 0.
// Nezáporná i záporná čísla se sčítají normálně.
func Sum(nums []int) int {
	// TODO
	return 0
}

// Reverse otočí pořadí prvků in-place. Funguje pro sudou i lichou délku;
// prázdný i nil vstup je no-op. Použij dva indexy zleva a zprava — nealokuj pomocný slice.
func Reverse(nums []int) {
	// TODO
}

// --- Stupeň: střední ---
// Grow zajistí cap >= n a zachová len i obsah.
// Pokud už cap(s) >= n, vrátí s beze změny (stejný header nad stejným polem).
// Jinak alokuj nové pole a data zkopíruj.
func Grow(s []int, n int) []int {
	// TODO
	return nil
}

// RemoveAt smaže prvek na indexu i a zachová pořadí zbytku.
// Index mimo rozsah vrací s beze změny. Mutuje backing pole volajícího.
func RemoveAt(s []int, i int) []int {
	// TODO
	return nil
}

// RemoveFast smaže prvek na indexu i v O(1) bez zachování pořadí
// (poslední prvek na místo i, pak zkrať). Index mimo rozsah → s beze změny.
func RemoveFast(s []int, i int) []int {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---
// Clone vrací nezávislou kopii. Nil vstup → nil; prázdný ne-nil → prázdný ne-nil.
func Clone(s []int) []int {
	// TODO
	return nil
}

// Chunk rozdělí s na nezávislé kopie délky size; poslední kus může být kratší.
// size <= 0 nebo prázdný vstup → výsledek nulové délky.
// Každý chunk musí být nezávislá kopie — zápis do jednoho nesmí ovlivnit ostatní ani vstup.
func Chunk(s []int, size int) [][]int {
	// TODO
	return nil
}

// Filter vrací prvky, pro které keep vrátí true, v původním pořadí.
// Implementuj trikem s[:0]: výsledek skládej do stejného backing pole jako vstup,
// bez nové alokace. Vstupní slice se tím přepíše — to je záměr.
func Filter(s []int, keep func(int) bool) []int {
	// TODO
	return nil
}
