// Package solutions obsahuje referenční řešení lekce 33.
package solutions

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrMissingDependency = errors.New("ordering: missing dependency")
	ErrEmptyCustomer     = errors.New("ordering: empty customer")
	ErrInvalidTotal      = errors.New("ordering: total must be positive")
	ErrNotFound          = errors.New("ordering: order not found")
	ErrStore             = errors.New("ordering: store failed")
)

type Clock interface {
	Now() time.Time
}

type IDGen interface {
	NewID() string
}

type OrderStore interface {
	Save(o Order) error
	Get(id string) (Order, bool)
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }

type RandomIDGen struct{}

func (RandomIDGen) NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("ordering: rand.Read: " + err.Error())
	}
	return hex.EncodeToString(b[:])
}

type Order struct {
	ID         string
	Customer   string
	TotalCents int64
	PlacedAt   time.Time
}

type Service struct {
	clock Clock
	ids   IDGen
}

func NewService(clock Clock, ids IDGen) (*Service, error) {
	if clock == nil || ids == nil {
		return nil, ErrMissingDependency
	}
	return &Service{clock: clock, ids: ids}, nil
}

func (s *Service) NewOrder(customer string, totalCents int64) (Order, error) {
	customer = strings.TrimSpace(customer)
	switch {
	case customer == "":
		return Order{}, ErrEmptyCustomer
	case totalCents <= 0:
		return Order{}, ErrInvalidTotal
	}
	return Order{
		ID:         s.ids.NewID(),
		Customer:   customer,
		TotalCents: totalCents,
		PlacedAt:   s.clock.Now(),
	}, nil
}

type MemoryStore struct {
	mu     sync.RWMutex
	orders map[string]Order
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{orders: make(map[string]Order)}
}

func (s *MemoryStore) Save(o Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[o.ID] = o
	return nil
}

func (s *MemoryStore) Get(id string) (Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	return o, ok
}

type FailingStore struct {
	Err error
}

func (s FailingStore) Save(o Order) error          { return s.Err }
func (s FailingStore) Get(id string) (Order, bool) { return Order{}, false }

type OrderService struct {
	store OrderStore
	svc   *Service
}

func NewOrderService(store OrderStore, svc *Service) (*OrderService, error) {
	if store == nil || svc == nil {
		return nil, ErrMissingDependency
	}
	return &OrderService{store: store, svc: svc}, nil
}

func (s *OrderService) Place(customer string, totalCents int64) (Order, error) {
	o, err := s.svc.NewOrder(customer, totalCents)
	if err != nil {
		return Order{}, err
	}
	if err := s.store.Save(o); err != nil {
		return Order{}, fmt.Errorf("place %s: %w: %w", o.ID, ErrStore, err)
	}
	return o, nil
}

func Wire() (*OrderService, error) {
	svc, err := NewService(SystemClock{}, RandomIDGen{})
	if err != nil {
		return nil, err
	}
	return NewOrderService(NewMemoryStore(), svc)
}
