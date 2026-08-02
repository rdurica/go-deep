// Package memstore je adaptér portu app.Repository držící data v paměti.
package memstore

import (
	"context"
	"fmt"
	"sync"

	"github.com/rdurica/go-deep/lessons/lesson-38/solutions/order"
)

// Repository je in-memory úložiště objednávek bezpečné pro souběžné použití.
type Repository struct {
	mu     sync.RWMutex
	orders map[string]order.Order
}

// --- Stupeň: jednoduchý ---
// New vytvoří prázdné úložiště.
func New() *Repository {
	return &Repository{orders: make(map[string]order.Order)}
}

// --- Stupeň: střední ---
// Save uloží nebo přepíše objednávku.
func (r *Repository) Save(ctx context.Context, o order.Order) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[o.ID] = clone(o)
	return nil
}

// --- Stupeň: obtížný ---
// Find vrátí objednávku podle ID, jinak chybu obalující order.ErrNotFound.
func (r *Repository) Find(ctx context.Context, id string) (order.Order, error) {
	if err := ctx.Err(); err != nil {
		return order.Order{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.orders[id]
	if !ok {
		return order.Order{}, fmt.Errorf("%w: %s", order.ErrNotFound, id)
	}
	return clone(o), nil
}

// clone odděluje slice položek. Skutečná databáze data serializuje, takže
// nic nesdílí; in-memory fake to musí dělat sám, jinak by se choval jinak
// než produkční adaptér a testy by lhaly.
func clone(o order.Order) order.Order {
	lines := make([]order.Line, len(o.Lines))
	copy(lines, o.Lines)
	o.Lines = lines
	return o
}
