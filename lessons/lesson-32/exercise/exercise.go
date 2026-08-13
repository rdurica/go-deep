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

// CatalogFixture sestaví katalog pro testy (neměň) — obchází studentův BuildCatalog v PART1.
func CatalogFixture(products ...Product) (*Catalog, error) {
	return catalog.New(products...)
}

// --- Stupeň: jednoduchý ---

// Snapshot vrátí kopii katalogu. Zápis do výsledku nesmí měnit originál.
// Pro nil vstup vrať nil.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Vrací stejný pointer místo kopie mapy.
// Najdi chybu a oprav — testy před opravou padají.
func Snapshot(c *Catalog) *Catalog {
	return c
}

// --- Stupeň: střední ---

// BuildCatalog sestaví katalog ze zadaných produktů přes catalog.New.
// Duplicitní SKU propaguje ErrDuplicateSKU. Prázdný vstup → prázdný katalog.
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

// NewIDGen vytvoří generátor identifikátorů s daným prefixem přes idgen.New.
// Prázdný prefix se nahradí "id". NewID vrací "<prefix>-000001", … bezpečně pro souběžné volání.
func NewIDGen(prefix string) *IDGen {
	// TODO
	return nil
}
