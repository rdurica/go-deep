// Package solutions obsahuje referenční řešení lekce 08.
package solutions

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
func CloneMap(in map[string]int) map[string]int {
	if in == nil {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// --- Stupeň: střední ---

// WordCount spočítá výskyty každého slova. Vždy vrací ne-nil mapu.
func WordCount(words []string) map[string]int {
	counts := make(map[string]int, len(words))
	for _, w := range words {
		// Chybějící klíč se čte jako 0, comma-ok tu není potřeba.
		counts[w]++
	}
	return counts
}

// NewSet vytvoří ne-nil množinu z předaných prvků.
func NewSet(items ...string) Set {
	s := make(Set, len(items))
	for _, item := range items {
		s.Add(item)
	}
	return s
}

// Add vloží prvek do množiny.
func (s Set) Add(item string) {
	s[item] = struct{}{}
}

// Has vrací true, pokud je prvek v množině. Funguje i na nil množině.
func (s Set) Has(item string) bool {
	_, ok := s[item]
	return ok
}

// --- Stupeň: obtížný ---

// AddStock přičte n kusů k položce sku. Chybějící položku založí.
// Nil inventář a n <= 0 jsou no-op.
func AddStock(inv Inventory, sku string, n int) {
	if inv == nil || n <= 0 {
		return
	}
	if item, ok := inv[sku]; ok {
		// Mapa vrací pointer, ten je adresovatelný — mutujeme přímo.
		item.Qty += n
		return
	}
	inv[sku] = &Item{SKU: sku, Qty: n}
}
