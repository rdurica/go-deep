// Package memstore je adaptér portu app.Repository držící data v paměti.
package memstore

import (
	"context"
	"sync"

	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/order"
)

// Repository je in-memory úložiště objednávek bezpečné pro souběžné použití.
type Repository struct {
	mu     sync.RWMutex
	orders map[string]order.Order
}

// New vytvoří prázdné úložiště.
func New() *Repository {
	// TODO: úkol B
	return nil
}

// Save uloží nebo přepíše objednávku.
func (r *Repository) Save(ctx context.Context, o order.Order) error {
	// TODO: úkol B
	return nil
}

// Find vrátí objednávku podle ID, jinak chybu obalující order.ErrNotFound.
func (r *Repository) Find(ctx context.Context, id string) (order.Order, error) {
	// TODO: úkol B
	return *new(order.Order), nil
}
