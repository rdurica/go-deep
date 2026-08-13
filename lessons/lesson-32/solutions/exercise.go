// Package solutions obsahuje referenční řešení lekce 32.
package solutions

import (
	"github.com/rdurica/go-deep/lessons/lesson-32/solutions/catalog"
	"github.com/rdurica/go-deep/lessons/lesson-32/solutions/internal/idgen"
	"github.com/rdurica/go-deep/lessons/lesson-32/solutions/pricing"
)

type (
	Product = catalog.Product
	Item    = catalog.Item
	Catalog = catalog.Catalog
	IDGen   = idgen.Gen
)

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

func CatalogFixture(products ...Product) (*Catalog, error) {
	return catalog.New(products...)
}

func Snapshot(c *Catalog) *Catalog {
	if c == nil {
		return nil
	}
	out, err := catalog.New(c.All()...)
	if err != nil {
		return nil
	}
	return out
}

// --- Stupeň: střední ---

func BuildCatalog(products ...Product) (*Catalog, error) {
	return catalog.New(products...)
}

func PriceOf(c *Catalog, sku string, qty int) (int64, error) {
	p, err := c.BySKU(sku)
	if err != nil {
		return 0, err
	}
	return pricing.Total([]Item{{Product: p, Qty: qty}})
}

// --- Stupeň: obtížný ---

func TotalOf(items []Item) (int64, error) {
	return pricing.Total(items)
}

func NewIDGen(prefix string) *IDGen {
	return idgen.New(prefix)
}
