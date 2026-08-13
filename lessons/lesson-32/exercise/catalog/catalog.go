// Package catalog je doménový balíček lekce 32.
//
// Je pojmenovaný podle domény, ne podle vrstvy. Nezná HTTP, SQL, ceny ani
// generování ID — a hlavně neimportuje nic z tohohle modulu. Je to list
// grafu závislostí, takže do něj smí ukazovat kdokoli a on nikam.
package catalog

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Chyby domény katalogu. Jsou to sentinely, porovnávej je přes errors.Is.
var (
	// ErrEmptySKU hlásí produkt bez SKU.
	ErrEmptySKU = errors.New("catalog: empty SKU")
	// ErrEmptyName hlásí produkt bez jména.
	ErrEmptyName = errors.New("catalog: empty name")
	// ErrNegativeCents hlásí zápornou cenu.
	ErrNegativeCents = errors.New("catalog: negative price")
	// ErrDuplicateSKU hlásí dva produkty se stejným SKU.
	ErrDuplicateSKU = errors.New("catalog: duplicate SKU")
	// ErrNotFound hlásí, že SKU v katalogu není.
	ErrNotFound = errors.New("catalog: product not found")
)

// Product je položka katalogu. Cena je vždy v celých centech, nikdy float64.
type Product struct {
	SKU   string
	Name  string
	Cents int64
}

// Item je produkt v požadovaném množství. Bydlí v catalog, protože je to
// doménový typ — balíček pricing si ho jen půjčuje.
type Item struct {
	Product Product
	Qty     int
}

// Catalog je kolekce produktů indexovaná podle SKU. Vytváří se přes New,
// protože nulová hodnota by neměla naplněnou mapu.
type Catalog struct {
	bySKU map[string]Product
}

// --- Stupeň: jednoduchý ---
// Validate ověří jeden produkt. Kontroluje v pořadí SKU, jméno, cenu
// a vrací odpovídající sentinel obalený jménem produktu.
func Validate(p Product) error {
	switch {
	case norm(p.SKU) == "":
		return errFor(p.SKU, ErrEmptySKU)
	case norm(p.Name) == "":
		return errFor(p.SKU, ErrEmptyName)
	case p.Cents < 0:
		return errFor(p.SKU, ErrNegativeCents)
	default:
		return nil
	}
}

// --- Stupeň: střední ---
// New sestaví katalog ze zadaných produktů. Každý produkt projde Validate,
// duplicitní SKU je chyba. Prázdný vstup dává prázdný katalog, ne chybu.
func New(products ...Product) (*Catalog, error) {
	c := &Catalog{bySKU: make(map[string]Product, len(products))}
	for _, p := range products {
		if err := Validate(p); err != nil {
			return nil, err
		}
		if _, dup := c.bySKU[p.SKU]; dup {
			return nil, errFor(p.SKU, ErrDuplicateSKU)
		}
		c.bySKU[p.SKU] = p
	}
	return c, nil
}

// --- Stupeň: obtížný ---
// BySKU vrací produkt podle SKU. Neznámé SKU (i nil katalog) vrací chybu
// obalující ErrNotFound.
func (c *Catalog) BySKU(sku string) (Product, error) {
	if c == nil {
		return Product{}, errFor(sku, ErrNotFound)
	}
	p, ok := c.bySKU[sku]
	if !ok {
		return Product{}, errFor(sku, ErrNotFound)
	}
	return p, nil
}

// All vrací všechny produkty seřazené vzestupně podle SKU. Mapa nemá pořadí,
// takže řazení je součástí kontraktu.
func (c *Catalog) All() []Product {
	if c == nil || len(c.bySKU) == 0 {
		return nil
	}
	out := make([]Product, 0, len(c.bySKU))
	for _, p := range c.bySKU {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SKU < out[j].SKU })
	return out
}

// norm je pomocník pro ořez bílých znaků. Malé písmeno = neviditelné mimo
// balíček; přesně tohle je v Go jednotka zapouzdření.
func norm(s string) string { return strings.TrimSpace(s) }

// errFor obalí sentinel identifikátorem produktu, aby chyba nesla kontext.
func errFor(sku string, err error) error {
	return fmt.Errorf("catalog: %q: %w", sku, err)
}
