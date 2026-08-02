// Package solutions obsahuje referenční řešení lekce 02.
package solutions

// Item je jedna položka objednávky. Cena je vždy v celých centech,
// nikdy ve float64.
type Item struct {
	Name       string
	PriceCents int
	Qty        int
}

// Catalog je ceník. Pole items je neexportované, takže se k němu
// zvenčí balíčku nedá dostat — zapouzdření dělá balíček, ne modifikátor.
type Catalog struct {
	items []Item
}

// --- Stupeň: jednoduchý ---
// ApplyDiscount vrátí cenu po slevě percent procent, zaokrouhlenou
// na celé centy (půlka nahoru). Percent mimo rozsah 0–100 se ořízne.
func ApplyDiscount(priceCents int, percent int) int {
	if priceCents <= 0 {
		return 0
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return (priceCents*(100-percent) + 50) / 100
}

// TotalCents sečte cenu všech položek včetně množství.
func TotalCents(items []Item) int {
	total := 0
	for _, it := range items {
		if it.Qty <= 0 {
			continue
		}
		total += it.PriceCents * it.Qty
	}
	return total
}

// --- Stupeň: střední ---
// Cheapest vrátí nejlevnější položku podle jednotkové ceny.
// Druhá návratová hodnota je false, pokud žádná položka není.
func Cheapest(items []Item) (Item, bool) {
	if len(items) == 0 {
		return Item{}, false
	}
	best := items[0]
	for _, it := range items[1:] {
		if it.PriceCents < best.PriceCents {
			best = it
		}
	}
	return best, true
}

// NewCatalog vytvoří ceník z položek. Wiring je ruční, žádný kontejner.
func NewCatalog(items []Item) *Catalog {
	copied := make([]Item, len(items))
	copy(copied, items)
	return &Catalog{items: copied}
}

// --- Stupeň: obtížný ---
// Price vrátí jednotkovou cenu položky podle jména a true, pokud existuje.
func (c *Catalog) Price(name string) (int, bool) {
	if c == nil {
		return 0, false
	}
	for _, it := range c.items {
		if it.Name == name {
			return it.PriceCents, true
		}
	}
	return 0, false
}

// Checkout spočítá cenu objednávky ze jmen položek se slevou percent.
// Vrátí false, pokud kterékoli jméno v ceníku není.
func (c *Catalog) Checkout(names []string, percent int) (int, bool) {
	total := 0
	for _, name := range names {
		price, ok := c.Price(name)
		if !ok {
			return 0, false
		}
		total += price
	}
	return ApplyDiscount(total, percent), true
}
