// Package exercise obsahuje cvičení lekce 08.
package exercise

// Set je množina řetězců postavená na mapě s nulovou hodnotou struct{}.
type Set map[string]struct{}

// Item je položka skladu.
type Item struct {
	SKU string
	Qty int
}

// Inventory je sklad indexovaný podle SKU. Hodnotou je pointer, aby šlo
// položku mutovat přímo v mapě.
type Inventory map[string]*Item

// --- Stupeň: jednoduchý ---

// CloneMap vrátí nezávislou kopii mapy. Zápis do výsledku nesmí měnit originál.
// Pro nil vstup vrať nil. Pro prázdnou ne-nil mapu vrať prázdnou ne-nil mapu.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Obsahuje typickou chybu se sdílením mapy.
// Najdi ji a oprav — testy před opravou padají.
func CloneMap(in map[string]int) map[string]int {
	// Špatně: přiřazení kopíruje jen referenci na tutéž tabulku.
	out := in
	return out
}

// --- Stupeň: střední ---

// WordCount spočítá výskyty každého slova.
// Vždy vrací ne-nil mapu (i pro nil / prázdný vstup).
// Chybějící klíč se čte jako nula — comma-ok netřeba.
func WordCount(words []string) map[string]int {
	// TODO
	return nil
}

// NewSet vytvoří množinu z předaných prvků; duplicity se započítají jednou.
// Bez argumentů vrací prázdnou, ale ne-nil množinu.
func NewSet(items ...string) Set {
	// TODO
	return *new(Set)
}

// Add vloží prvek do množiny.
// Opakované vložení nic nemění.
func (s Set) Add(item string) {
	// TODO
}

// Has vrací true, pokud je prvek v množině.
// Na nil množině vrací false (bez paniky).
func (s Set) Has(item string) bool {
	// TODO
	return false
}

// --- Stupeň: obtížný ---

// AddStock přičte n kusů k položce sku.
// Když klíč chybí, založí &Item{SKU: sku, Qty: n}.
// Když existuje, zvýší Qty přes pointer (ne přepis celé hodnoty).
// Nil inventář a n <= 0 jsou no-op (bez paniky); při n <= 0 chybějící položku nezakládá.
func AddStock(inv Inventory, sku string, n int) {
	// TODO
}
