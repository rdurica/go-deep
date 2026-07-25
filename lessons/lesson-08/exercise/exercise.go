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

// WordCount spočítá výskyty každého slova. Vždy vrací ne-nil mapu.
func WordCount(words []string) map[string]int {
	// TODO: úkol A
	return nil
}

// NewSet vytvoří ne-nil množinu z předaných prvků.
func NewSet(items ...string) Set {
	// TODO: úkol B
	return *new(Set)
}

// Add vloží prvek do množiny.
func (s Set) Add(item string) {
	// TODO: úkol B
}

// Has vrací true, pokud je prvek v množině. Funguje i na nil množině.
func (s Set) Has(item string) bool {
	// TODO: úkol B
	return false
}

// Remove odebere prvek z množiny. Neexistující prvek i nil množina jsou no-op.
func (s Set) Remove(item string) {
	// TODO: úkol B
}

// Len vrací počet prvků množiny.
func (s Set) Len() int {
	// TODO: úkol B
	return 0
}

// Sorted vrací prvky množiny vzestupně seřazené.
func (s Set) Sorted() []string {
	// TODO: úkol B
	return nil
}

// Union vrací novou množinu se všemi prvky s i other.
func (s Set) Union(other Set) Set {
	// TODO: úkol B
	return *new(Set)
}

// Intersect vrací novou množinu s prvky, které jsou v s i v other.
func (s Set) Intersect(other Set) Set {
	// TODO: úkol B
	return *new(Set)
}

// AddStock přičte n kusů k položce sku. Chybějící položku založí.
// Nil inventář a n <= 0 jsou no-op.
func AddStock(inv Inventory, sku string, n int) {
	// TODO: úkol C
}

// GroupBy seskupí slova podle klíče vráceného funkcí key.
// Uvnitř skupiny zachovává pořadí vstupu. Vždy vrací ne-nil mapu.
func GroupBy(words []string, key func(string) string) map[string][]string {
	// TODO: úkol C
	return nil
}
