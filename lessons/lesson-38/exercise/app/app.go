// Package app drží use-casy služby a porty, které k nim potřebuje.
//
// Porty jsou definované tady, u konzumenta — ne u adaptéru, který je plní.
package app

import (
	"context"

	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/order"
)

// Repository je port pro uložení a načtení objednávky.
// Při neexistující objednávce vrací chybu obalující order.ErrNotFound.
type Repository interface {
	Save(ctx context.Context, o order.Order) error
	Find(ctx context.Context, id string) (order.Order, error)
}

// IDGen je port pro generování identifikátorů objednávek.
type IDGen interface {
	NewID() string
}

// Service je aplikační vrstva: orchestruje doménu a porty, sama žádné
// pravidlo nezná.
type Service struct {
	repo Repository
	ids  IDGen
}

// NewService složí službu z portů.
func NewService(repo Repository, ids IDGen) *Service {
	// TODO: úkol B
	return nil
}

// Place založí novou objednávku a uloží ji.
func (s *Service) Place(ctx context.Context, lines []order.Line) (order.Order, error) {
	// TODO: úkol B
	return *new(order.Order), nil
}

// Get vrátí objednávku podle ID.
func (s *Service) Get(ctx context.Context, id string) (order.Order, error) {
	// TODO: úkol B
	return *new(order.Order), nil
}

// Pay zaplatí objednávku.
func (s *Service) Pay(ctx context.Context, id string) (order.Order, error) {
	// TODO: úkol B
	return *new(order.Order), nil
}

// Ship odešle objednávku.
func (s *Service) Ship(ctx context.Context, id string) (order.Order, error) {
	// TODO: úkol B
	return *new(order.Order), nil
}

// Cancel zruší objednávku.
func (s *Service) Cancel(ctx context.Context, id string) (order.Order, error) {
	// TODO: úkol B
	return *new(order.Order), nil
}
