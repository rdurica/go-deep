// Package catalog je doménový balíček lekce 32.
//
// Je pojmenovaný podle domény, ne podle vrstvy. Nezná HTTP, SQL, ceny ani
// generování ID — a hlavně neimportuje nic z tohohle modulu. Je to list
// grafu závislostí, takže do něj smí ukazovat kdokoli a on nikam.
package catalog

import (
	"errors"
	"fmt"
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

// Validate ověří jeden produkt. Kontroluje v pořadí SKU, jméno, cenu
// a vrací odpovídající sentinel obalený jménem produktu.
func Validate(p Product) error {
	panic("TODO: úkol A")
}

// New sestaví katalog ze zadaných produktů. Každý produkt projde Validate,
// duplicitní SKU je chyba. Prázdný vstup dává prázdný katalog, ne chybu.
func New(products ...Product) (*Catalog, error) {
	panic("TODO: úkol A")
}

// BySKU vrací produkt podle SKU. Neznámé SKU (i nil katalog) vrací chybu
// obalující ErrNotFound.
func (c *Catalog) BySKU(sku string) (Product, error) {
	panic("TODO: úkol A")
}

// All vrací všechny produkty seřazené vzestupně podle SKU. Mapa nemá pořadí,
// takže řazení je součástí kontraktu.
func (c *Catalog) All() []Product {
	panic("TODO: úkol A")
}

// norm je pomocník pro ořez bílých znaků. Malé písmeno = neviditelné mimo
// balíček; přesně tohle je v Go jednotka zapouzdření.
func norm(s string) string { return strings.TrimSpace(s) }

// errFor obalí sentinel identifikátorem produktu, aby chyba nesla kontext.
func errFor(sku string, err error) error {
	return fmt.Errorf("catalog: %q: %w", sku, err)
}
