// Package exercise obsahuje cvičení lekce 02.
package exercise

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

// ApplyDiscount vrátí cenu po slevě percent procent, zaokrouhlenou
// na celé centy (půlka nahoru). Percent mimo rozsah 0–100 se ořízne.
func ApplyDiscount(priceCents int, percent int) int {
	panic("TODO: úkol A")
}

// TotalCents sečte cenu všech položek včetně množství.
func TotalCents(items []Item) int {
	panic("TODO: úkol B")
}

// Cheapest vrátí nejlevnější položku podle jednotkové ceny.
// Druhá návratová hodnota je false, pokud žádná položka není.
func Cheapest(items []Item) (Item, bool) {
	panic("TODO: úkol B")
}

// NewCatalog vytvoří ceník z položek. Wiring je ruční, žádný kontejner.
func NewCatalog(items []Item) *Catalog {
	panic("TODO: úkol C")
}

// Price vrátí jednotkovou cenu položky podle jména a true, pokud existuje.
func (c *Catalog) Price(name string) (int, bool) {
	panic("TODO: úkol C")
}

// Checkout spočítá cenu objednávky ze jmen položek se slevou percent.
// Vrátí false, pokud kterékoli jméno v ceníku není.
func (c *Catalog) Checkout(names []string, percent int) (int, bool) {
	panic("TODO: úkol C")
}
