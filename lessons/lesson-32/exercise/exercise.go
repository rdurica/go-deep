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

// --- Stupeň: jednoduchý ---
// Validate ověří jeden produkt doménovými pravidly katalogu (průchozí volání do catalog.Validate).
// Kontroluje SKU, jméno a cenu; chyby propaguje beze změny sentinelu.
func Validate(p Product) error {
	// TODO
	return nil
}

// --- Stupeň: střední ---
// BuildCatalog sestaví katalog ze zadaných produktů (průchozí volání do catalog.New).
// Každý produkt projde Validate; duplicitní SKU je ErrDuplicateSKU. Prázdný vstup → prázdný katalog.
func BuildCatalog(products ...Product) (*Catalog, error) {
	// TODO
	return nil, nil
}

// PriceOf vrátí cenu qty kusů produktu se zadaným SKU v katalogu.
// Neznámé SKU propaguje ErrNotFound; nekladné množství ErrInvalidQty z pricing.
func PriceOf(c *Catalog, sku string, qty int) (int64, error) {
	// TODO
	return 0, nil
}

// --- Stupeň: obtížný ---
// TotalOf spočítá cenu celého košíku přes pricing.Total.
// Prázdný košík → 0, nil. Chyby validace a přetečení propaguje beze změny.
func TotalOf(items []Item) (int64, error) {
	// TODO
	return 0, nil
}

// NewIDGen vytvoří generátor identifikátorů s daným prefixem (průchozí volání do idgen.New).
// Prázdný prefix se nahradí "id". NewID vrací "<prefix>-000001", … bezpečně pro souběžné volání.
func NewIDGen(prefix string) *IDGen {
	// TODO
	return nil
}
