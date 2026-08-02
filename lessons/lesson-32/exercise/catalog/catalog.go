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

// Validate ověří jeden produkt v pořadí SKU, jméno, cena.
// Prázdné SKU/jméno (i jen bílé znaky) → ErrEmptySKU/ErrEmptyName;
// záporná cena → ErrNegativeCents; nula je platná. Chybu obal SKU pro errors.Is.
func Validate(p Product) error {
	// TODO
	return nil
}

// New sestaví katalog ze zadaných produktů. Každý projde Validate,
// duplicitní SKU je ErrDuplicateSKU. Prázdný vstup dává prázdný katalog.
func New(products ...Product) (*Catalog, error) {
	// TODO
	return nil, nil
}

// BySKU vrací produkt podle SKU. Neznámé SKU i nil katalog vrací chybu
// obalující ErrNotFound (metoda na nil pointeru je legální).
func (c *Catalog) BySKU(sku string) (Product, error) {
	// TODO
	return *new(Product), nil
}

// All vrací všechny produkty seřazené vzestupně podle SKU.
// Iterace mapy je náhodná — řazení je součást kontraktu, ne volitelný detail.
func (c *Catalog) All() []Product {
	// TODO
	return nil
}

// norm je pomocník pro ořez bílých znaků. Malé písmeno = neviditelné mimo
// balíček; přesně tohle je v Go jednotka zapouzdření.
func norm(s string) string { return strings.TrimSpace(s) }

// errFor obalí sentinel identifikátorem produktu, aby chyba nesla kontext.
func errFor(sku string, err error) error {
	return fmt.Errorf("catalog: %q: %w", sku, err)
}
