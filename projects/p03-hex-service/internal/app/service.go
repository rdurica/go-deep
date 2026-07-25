package app

import (
	"context"
	"fmt"

	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/order"
)

// LineCommand je jedna položka v příkazu k založení objednávky.
type LineCommand struct {
	SKU            string
	Quantity       int
	UnitPriceCents int64
}

// PlaceCommand je vstup use-casu Place. Je to záměrně jiný typ než
// doménová objednávka: příkaz popisuje, co chce volající, ne co platí.
type PlaceCommand struct {
	Customer string
	Currency string
	Lines    []LineCommand
}

// Service je aplikační vrstva: orchestruje doménu a porty. Sama žádné
// doménové pravidlo nezná — od toho je balíček order.
type Service struct {
	repo  Repository
	clock Clock
	ids   IDGen
}

// NewService složí službu ze všech portů, které potřebuje.
func NewService(repo Repository, clock Clock, ids IDGen) *Service {
	return &Service{repo: repo, clock: clock, ids: ids}
}

// Place založí novou objednávku a uloží ji.
func (s *Service) Place(ctx context.Context, cmd PlaceCommand) (order.Order, error) {
	lines := make([]order.Line, 0, len(cmd.Lines))
	for i, l := range cmd.Lines {
		price, err := order.NewMoney(l.UnitPriceCents, cmd.Currency)
		if err != nil {
			return order.Order{}, fmt.Errorf("položka %d: %w", i, err)
		}
		line, err := order.NewLine(l.SKU, l.Quantity, price)
		if err != nil {
			return order.Order{}, fmt.Errorf("položka %d: %w", i, err)
		}
		lines = append(lines, line)
	}

	o, err := order.New(s.ids.NewID(), cmd.Customer, lines, s.clock.Now())
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

// List vrátí všechny objednávky seřazené podle ID.
func (s *Service) List(ctx context.Context) ([]order.Order, error) {
	orders, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("výpis objednávek: %w", err)
	}
	return orders, nil
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
// doménu, ulož. Když doména přechod zamítne, neuloží se nic.
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
