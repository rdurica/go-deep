// Package app drží use-casy služby a porty, které k nim potřebuje.
//
// Porty jsou definované tady, u konzumenta. Adaptéry (memstore, httpapi,
// budoucí pgstore) o tomhle balíčku vědět nemusí — Go interfacy se
// implementují implicitně.
package app

import (
	"context"
	"time"

	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/order"
)

// Repository je port pro perzistenci objednávek.
// Při neexistující objednávce vrací chybu obalující order.ErrNotFound.
type Repository interface {
	Save(ctx context.Context, o order.Order) error
	Find(ctx context.Context, id string) (order.Order, error)
	List(ctx context.Context) ([]order.Order, error)
}

// Clock je port pro čtení aktuálního času. Bez něj by byl každý test
// o časových razítkách hádankou.
type Clock interface {
	Now() time.Time
}

// IDGen je port pro generování identifikátorů objednávek.
type IDGen interface {
	NewID() string
}

// ClockFunc dělá z funkce port Clock.
type ClockFunc func() time.Time

// Now implementuje Clock.
func (f ClockFunc) Now() time.Time { return f() }

// IDGenFunc dělá z funkce port IDGen.
type IDGenFunc func() string

// NewID implementuje IDGen.
func (f IDGenFunc) NewID() string { return f() }
