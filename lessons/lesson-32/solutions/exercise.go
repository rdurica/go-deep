// Package solutions obsahuje referenční řešení lekce 32.
//
// Nemá vlastní pravidla. Jen skládá dohromady doménu (catalog), výpočet
// (pricing) a interní nástroj (internal/idgen). Přesně tuhle roli má
// v reálné službě funkce main v cmd/.
package solutions

import (
	"github.com/rdurica/go-deep/lessons/lesson-32/solutions/catalog"
	"github.com/rdurica/go-deep/lessons/lesson-32/solutions/internal/idgen"
	"github.com/rdurica/go-deep/lessons/lesson-32/solutions/pricing"
)

// Doménové typy jsou tu jen jako aliasy. Aplikační vrstva je nedefinuje
// znovu, protože jediné místo pravdy je balíček, kde téma bydlí.
type (
	// Product je položka katalogu.
	Product = catalog.Product
	// Item je produkt v požadovaném množství.
	Item = catalog.Item
	// Catalog je kolekce produktů.
	Catalog = catalog.Catalog
	// IDGen je generátor identifikátorů z internal/idgen.
	IDGen = idgen.Gen
)

// Sentinely přeexportované pro pohodlí konzumentů kompozice.
var (
	ErrEmptySKU      = catalog.ErrEmptySKU
	ErrEmptyName     = catalog.ErrEmptyName
	ErrNegativeCents = catalog.ErrNegativeCents
	ErrDuplicateSKU  = catalog.ErrDuplicateSKU
	ErrNotFound      = catalog.ErrNotFound
	ErrInvalidQty    = pricing.ErrInvalidQty
	ErrOverflow      = pricing.ErrOverflow
)

// Validate ověří jeden produkt doménovými pravidly.
func Validate(p Product) error {
	return catalog.Validate(p)
}

// BuildCatalog sestaví katalog ze zadaných produktů.
func BuildCatalog(products ...Product) (*Catalog, error) {
	return catalog.New(products...)
}

// PriceOf vrátí cenu qty kusů produktu se zadaným SKU.
// Neznámé SKU propaguje ErrNotFound, nekladné množství ErrInvalidQty.
func PriceOf(c *Catalog, sku string, qty int) (int64, error) {
	p, err := c.BySKU(sku)
	if err != nil {
		return 0, err
	}
	return pricing.Total([]Item{{Product: p, Qty: qty}})
}

// TotalOf spočítá cenu celého košíku.
func TotalOf(items []Item) (int64, error) {
	return pricing.Total(items)
}

// NewIDGen vytvoří generátor identifikátorů s daným prefixem.
func NewIDGen(prefix string) *IDGen {
	return idgen.New(prefix)
}
