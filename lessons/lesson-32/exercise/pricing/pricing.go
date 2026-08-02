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

	"github.com/rdurica/go-deep/lessons/lesson-32/exercise/catalog"
)

// Chyby výpočtu ceny.
var (
	// ErrInvalidQty hlásí nekladné množství.
	ErrInvalidQty = errors.New("pricing: quantity must be positive")
	// ErrOverflow hlásí přetečení int64 při součtu.
	ErrOverflow = errors.New("pricing: total overflow")
)

// Total spočítá cenu košíku v centech.
// Prázdný nebo nil košík → 0, nil. Každý produkt projde catalog.Validate;
// nekladné množství → ErrInvalidQty. Řádková cena Cents*Qty, součet nesmí
// přetéct int64 — vrať ErrOverflow, nikdy tiše přetečenou hodnotu.
func Total(items []catalog.Item) (money int64, err error) {
	// TODO
	return
}

// mulOverflows hlásí, jestli a*b přeteče int64. Oba operandy jsou nezáporné.
func mulOverflows(a, b int64) bool {
	return a != 0 && b > math.MaxInt64/a
}

// wrap obalí sentinel identifikátorem položky.
func wrap(sku string, err error) error {
	return fmt.Errorf("pricing: %q: %w", sku, err)
}
