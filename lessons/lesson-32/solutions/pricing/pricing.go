// Package pricing počítá ceny nad doménovými typy z balíčku catalog.
//
// Směr závislosti je pricing -> catalog. Opačný import by byl cyklus a build
// by spadl na "import cycle not allowed". Kompilátor tím za tebe hlídá, že
// doména nezná své konzumenty.
package pricing

import (
	"errors"
	"fmt"
	"math"

	"github.com/rdurica/go-deep/lessons/lesson-32/solutions/catalog"
)

// Chyby výpočtu ceny.
var (
	// ErrInvalidQty hlásí nekladné množství.
	ErrInvalidQty = errors.New("pricing: quantity must be positive")
	// ErrOverflow hlásí přetečení int64 při součtu.
	ErrOverflow = errors.New("pricing: total overflow")
)

// --- Stupeň: jednoduchý ---
// Total spočítá cenu košíku v centech.
//
// Každý produkt musí projít catalog.Validate, množství musí být kladné
// a součet se nesmí přetéct. Prázdný nebo nil vstup dává 0 bez chyby.
func Total(items []catalog.Item) (money int64, err error) {
	for _, it := range items {
		if err := catalog.Validate(it.Product); err != nil {
			return 0, err
		}
		if it.Qty <= 0 {
			return 0, wrap(it.Product.SKU, ErrInvalidQty)
		}
		qty := int64(it.Qty)
		if mulOverflows(it.Product.Cents, qty) {
			return 0, wrap(it.Product.SKU, ErrOverflow)
		}
		line := it.Product.Cents * qty
		if money > math.MaxInt64-line {
			return 0, wrap(it.Product.SKU, ErrOverflow)
		}
		money += line
	}
	return money, nil
}

// mulOverflows hlásí, jestli a*b přeteče int64. Oba operandy jsou nezáporné.
func mulOverflows(a, b int64) bool {
	return a != 0 && b > math.MaxInt64/a
}

// wrap obalí sentinel identifikátorem položky.
func wrap(sku string, err error) error {
	return fmt.Errorf("pricing: %q: %w", sku, err)
}
