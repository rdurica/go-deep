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

// --- Stupeň: jednoduchý ---
// ApplyDiscount vrátí cenu po slevě percent procent, zaokrouhlenou na celé centy
// (půlka nahoru: 5 centů se slevou 50 % je 3). Percent mimo 0–100 se ořízne.
// priceCents <= 0 vrací 0.
func ApplyDiscount(priceCents int, percent int) int {
	// TODO
	return 0
}

// TotalCents sečte PriceCents * Qty přes všechny položky.
// Položky s Qty <= 0 přeskočí. Nil i prázdný slice dají 0. Vstup nemění.
func TotalCents(items []Item) int {
	// TODO
	return 0
}

// --- Stupeň: střední ---
// Cheapest vrátí položku s nejnižší jednotkovou cenou (PriceCents) a true.
// Při shodě ceny vyhrává první výskyt. Prázdný nebo nil vstup: Item{}, false.
func Cheapest(items []Item) (Item, bool) {
	// TODO
	return *new(Item), false
}

// NewCatalog vytvoří ceník z položek. Vstupní slice okopíruje,
// aby pozdější změna u volajícího ceník neovlivnila.
func NewCatalog(items []Item) *Catalog {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---
// Price vrátí jednotkovou cenu položky podle jména a true, pokud existuje.
// Nenalezeno vrací 0, false.
func (c *Catalog) Price(name string) (int, bool) {
	// TODO
	return 0, false
}

// Checkout spočítá cenu objednávky ze jmen (opakované jméno se počítá vícekrát),
// aplikuje ApplyDiscount a vrátí true. Chybějící jméno: 0, false.
// Prázdný seznam jmen je platná objednávka za 0.
func (c *Catalog) Checkout(names []string, percent int) (int, bool) {
	// TODO
	return 0, false
}
