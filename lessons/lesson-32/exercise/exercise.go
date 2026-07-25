// Package exercise je aplikační vrstva lekce 32.
//
// Nemá vlastní pravidla. Jen skládá dohromady doménu (catalog), výpočet
// (pricing) a interní nástroj (internal/idgen). Přesně tuhle roli má
// v reálné službě funkce main v cmd/.
package exercise

import (
	"github.com/rdurica/go-deep/lessons/lesson-32/exercise/catalog"
	"github.com/rdurica/go-deep/lessons/lesson-32/exercise/internal/idgen"
	"github.com/rdurica/go-deep/lessons/lesson-32/exercise/pricing"
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
	panic("TODO: úkol A")
}

// BuildCatalog sestaví katalog ze zadaných produktů.
func BuildCatalog(products ...Product) (*Catalog, error) {
	panic("TODO: úkol A")
}

// PriceOf vrátí cenu qty kusů produktu se zadaným SKU.
// Neznámé SKU propaguje ErrNotFound, nekladné množství ErrInvalidQty.
func PriceOf(c *Catalog, sku string, qty int) (int64, error) {
	panic("TODO: úkol B")
}

// TotalOf spočítá cenu celého košíku.
func TotalOf(items []Item) (int64, error) {
	panic("TODO: úkol B")
}

// NewIDGen vytvoří generátor identifikátorů s daným prefixem.
func NewIDGen(prefix string) *IDGen {
	panic("TODO: úkol C")
}
