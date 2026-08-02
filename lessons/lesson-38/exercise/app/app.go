// Package app drží use-casy služby a porty, které k nim potřebuje.
//
// Porty jsou definované tady, u konzumenta — ne u adaptéru, který je plní.
package app

import (
	"context"

	"github.com/rdurica/go-deep/lessons/lesson-38/exercise/order"
)

// Repository je port pro uložení a načtení objednávky.
// Při neexistující objednávce chyba obalující order.ErrNotFound.
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

// NewService složí službu z portů repo a ids.
// Oba porty musí být nenilové; služba sama nevaliduje doménu.
func NewService(repo Repository, ids IDGen) *Service {
	// TODO
	return nil
}

// Place založí objednávku: ID z portu IDGen, doména ověří invarianty, uloží přes Repository.
// Neplatná objednávka se nesmí uložit. Chyby portů obal %w.
func (s *Service) Place(ctx context.Context, lines []order.Line) (order.Order, error) {
	// TODO
	return *new(order.Order), nil
}

// Get vrátí objednávku podle ID z Repository.
// Neznámé ID propaguje chybu z portu obalující order.ErrNotFound.
func (s *Service) Get(ctx context.Context, id string) (order.Order, error) {
	// TODO
	return *new(order.Order), nil
}

// Pay načte objednávku, zavolá doménový Pay, výsledek uloží.
// Zamítnutý přechod stavu → Save se nezavolá vůbec.
func (s *Service) Pay(ctx context.Context, id string) (order.Order, error) {
	// TODO
	return *new(order.Order), nil
}

// Ship načte objednávku, zavolá doménový Ship, výsledek uloží.
// Zamítnutý přechod stavu → Save se nezavolá vůbec.
func (s *Service) Ship(ctx context.Context, id string) (order.Order, error) {
	// TODO
	return *new(order.Order), nil
}

// Cancel načte objednávku, zavolá doménový Cancel, výsledek uloží.
// Zamítnutý přechod stavu → Save se nezavolá vůbec.
func (s *Service) Cancel(ctx context.Context, id string) (order.Order, error) {
	// TODO
	return *new(order.Order), nil
}
