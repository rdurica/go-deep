// Package app drží use-casy služby a porty, které k nim potřebuje.
//
// Porty jsou definované tady, u konzumenta — ne u adaptéru, který je plní.
package app

import (
	"context"
	"fmt"

	"github.com/rdurica/go-deep/lessons/lesson-38/solutions/order"
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
	return &Service{repo: repo, ids: ids}
}

// Place založí novou objednávku a uloží ji.
func (s *Service) Place(ctx context.Context, lines []order.Line) (order.Order, error) {
	o, err := order.New(s.ids.NewID(), lines)
	if err != nil {
		return order.Order{}, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return order.Order{}, fmt.Errorf("uložení objednávky: %w", err)
	}
	return o, nil
}

// Get vrátí objednávku podle ID.
func (s *Service) Get(ctx context.Context, id string) (order.Order, error) {
	o, err := s.repo.Find(ctx, id)
	if err != nil {
		return order.Order{}, fmt.Errorf("načtení objednávky %q: %w", id, err)
	}
	return o, nil
}

// Pay zaplatí objednávku.
func (s *Service) Pay(ctx context.Context, id string) (order.Order, error) {
	return s.apply(ctx, id, order.Order.Pay)
}

// Ship odešle objednávku.
func (s *Service) Ship(ctx context.Context, id string) (order.Order, error) {
	return s.apply(ctx, id, order.Order.Ship)
}

// Cancel zruší objednávku.
func (s *Service) Cancel(ctx context.Context, id string) (order.Order, error) {
	return s.apply(ctx, id, order.Order.Cancel)
}

// apply je společné tělo use-casů měnících stav: načti, nech rozhodnout
// doménu, ulož. Rozhodnutí samotné aplikační vrstvě nepatří.
func (s *Service) apply(ctx context.Context, id string, change func(order.Order) (order.Order, error)) (order.Order, error) {
	current, err := s.repo.Find(ctx, id)
	if err != nil {
		return order.Order{}, fmt.Errorf("načtení objednávky %q: %w", id, err)
	}
	next, err := change(current)
	if err != nil {
		return order.Order{}, err
	}
	if err := s.repo.Save(ctx, next); err != nil {
		return order.Order{}, fmt.Errorf("uložení objednávky %q: %w", id, err)
	}
	return next, nil
}
