// Package memstore je in-memory adaptér portu app.Repository.
//
// Chová se schválně jako skutečná databáze: data při zápisu i čtení
// kopíruje. Fake, který sdílí slice s volajícím, by v testech ukazoval
// úspěch tam, kde produkční adaptér selže.
package memstore

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/order"
)

// Repository drží objednávky v mapě a je bezpečná pro souběžné použití.
type Repository struct {
	mu     sync.RWMutex
	orders map[string]order.Order
}

// New vytvoří prázdné úložiště.
func New() *Repository {
	return &Repository{orders: make(map[string]order.Order)}
}

// Save uloží nebo přepíše objednávku.
func (r *Repository) Save(ctx context.Context, o order.Order) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if o.ID == "" {
		return order.ErrMissingID
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[o.ID] = clone(o)
	return nil
}

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

// List vrátí všechny objednávky seřazené podle ID.
//
// Pořadí je součást kontraktu portu: mapa v Go iteruje náhodně, takže
// bez seřazení by testy nad výpisem byly flaky.
func (r *Repository) List(ctx context.Context) ([]order.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]order.Order, 0, len(r.orders))
	for _, o := range r.orders {
		out = append(out, clone(o))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func clone(o order.Order) order.Order {
	lines := make([]order.Line, len(o.Lines))
	copy(lines, o.Lines)
	o.Lines = lines
	return o
}
