// Package solutions obsahuje referenční řešení lekce 08.
package solutions

import "sort"

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

// Remove odebere prvek z množiny. Neexistující prvek i nil množina jsou no-op.
func (s Set) Remove(item string) {
	delete(s, item)
}

// Len vrací počet prvků množiny.
func (s Set) Len() int {
	return len(s)
}

// Sorted vrací prvky množiny vzestupně seřazené.
func (s Set) Sorted() []string {
	items := make([]string, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	sort.Strings(items)
	return items
}

// Union vrací novou množinu se všemi prvky s i other.
func (s Set) Union(other Set) Set {
	out := make(Set, len(s)+len(other))
	for item := range s {
		out.Add(item)
	}
	for item := range other {
		out.Add(item)
	}
	return out
}

// Intersect vrací novou množinu s prvky, které jsou v s i v other.
func (s Set) Intersect(other Set) Set {
	// Iterujeme přes menší množinu, práce je pak úměrná té menší.
	smaller, larger := s, other
	if len(larger) < len(smaller) {
		smaller, larger = larger, smaller
	}
	out := make(Set, len(smaller))
	for item := range smaller {
		if larger.Has(item) {
			out.Add(item)
		}
	}
	return out
}

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

// GroupBy seskupí slova podle klíče vráceného funkcí key.
// Uvnitř skupiny zachovává pořadí vstupu. Vždy vrací ne-nil mapu.
func GroupBy(words []string, key func(string) string) map[string][]string {
	groups := make(map[string][]string)
	for _, w := range words {
		k := key(w)
		groups[k] = append(groups[k], w)
	}
	return groups
}
