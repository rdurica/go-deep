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

// --- Stupeň: střední ---
// Has vrací true, pokud je prvek v množině.
// Na nil množině vrací false (bez paniky).
func (s Set) Has(item string) bool {
	// TODO
	return false
}

// Remove odebere prvek z množiny.
// Neexistující prvek i nil množina jsou no-op (bez paniky).
func (s Set) Remove(item string) {
	// TODO
}

// Len vrací počet prvků množiny.
// Nil množina vrací 0 bez paniky.
func (s Set) Len() int {
	// TODO
	return 0
}

// Sorted vrací prvky množiny vzestupně seřazené.
// Prázdná množina vrací výsledek nulové délky.
// Tohle je jediný způsob, jak z množiny dostat deterministický výstup.
func (s Set) Sorted() []string {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---
// Union vrací novou množinu se všemi prvky s i other.
// Nesmí měnit ani s, ani other.
func (s Set) Union(other Set) Set {
	// TODO
	return *new(Set)
}

// Intersect vrací novou množinu s prvky, které jsou v s i v other.
// Nesmí měnit ani s, ani other.
func (s Set) Intersect(other Set) Set {
	// TODO
	return *new(Set)
}

// AddStock přičte n kusů k položce sku.
// Když klíč chybí, založí &Item{SKU: sku, Qty: n}.
// Když existuje, zvýší Qty přes pointer (ne přepis celé hodnoty).
// Nil inventář a n <= 0 jsou no-op (bez paniky); při n <= 0 chybějící položku nezakládá.
func AddStock(inv Inventory, sku string, n int) {
	// TODO
}

// GroupBy seskupí slova podle klíče vráceného funkcí key.
// Uvnitř skupiny zachovává pořadí vstupu.
// Vždy vrací ne-nil mapu.
func GroupBy(words []string, key func(string) string) map[string][]string {
	// TODO
	return nil
}
