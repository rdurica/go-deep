// Package exercise obsahuje cvičení lekce 33.
package exercise

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

// Chyby domény objednávek.
var (
	ErrMissingDependency = errors.New("ordering: missing dependency")
	ErrEmptyCustomer     = errors.New("ordering: empty customer")
	ErrInvalidTotal      = errors.New("ordering: total must be positive")
	ErrNotFound          = errors.New("ordering: order not found")
	ErrStore             = errors.New("ordering: store failed")
)

// Clock je port u konzumenta.
type Clock interface {
	Now() time.Time
}

// IDGen je port u konzumenta.
type IDGen interface {
	NewID() string
}

// OrderStore je driven port — dvě metody stačí.
type OrderStore interface {
	Save(o Order) error
	Get(id string) (Order, bool)
}

// SystemClock je adaptér portu Clock.
type SystemClock struct{}

// Now vrací systémový čas.
func (SystemClock) Now() time.Time { return time.Now() }

// RandomIDGen je adaptér portu IDGen.
type RandomIDGen struct{}

// NewID vrací 32 hex znaků z crypto/rand.
func (RandomIDGen) NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("ordering: rand.Read: " + err.Error())
	}
	return hex.EncodeToString(b[:])
}

// Order je hodnotový typ objednávky.
type Order struct {
	ID         string
	Customer   string
	TotalCents int64
	PlacedAt   time.Time
}

// Service sestavuje objednávky přes porty Clock a IDGen.
type Service struct {
	clock Clock
	ids   IDGen
}

// NewService sestaví službu. Chybějící port → ErrMissingDependency.
func NewService(clock Clock, ids IDGen) (*Service, error) {
	if clock == nil || ids == nil {
		return nil, ErrMissingDependency
	}
	return &Service{clock: clock, ids: ids}, nil
}

// --- Stupeň: jednoduchý ---

// NewOrder sestaví objednávku. Jméno ořízni, prázdné → ErrEmptyCustomer,
// nekladná částka → ErrInvalidTotal. ID z IDGen, čas z Clock.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Volá time.Now() místo portu Clock.
// Testy s fake hodinami před opravou padají.
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
		PlacedAt:   time.Now(),
	}, nil
}

// --- Stupeň: střední ---

// MemoryStore je in-memory adaptér portu OrderStore.
type MemoryStore struct {
	mu     sync.RWMutex
	orders map[string]Order
}

// NewMemoryStore vytvoří prázdné úložiště. Bezpečné pro souběžné použití (-race).
func NewMemoryStore() *MemoryStore {
	// TODO
	return nil
}

// Save uloží objednávku. Existující ID přepíše. Bezpečné pro souběh.
func (s *MemoryStore) Save(o Order) error {
	// TODO
	return nil
}

// Get vrátí objednávku a příznak existence. Neznámé ID → false.
func (s *MemoryStore) Get(id string) (Order, bool) {
	// TODO
	return *new(Order), false
}

// FailingStore je fake adaptér pro test chybové cesty.
type FailingStore struct {
	Err error
}

// Save vždy vrací Err.
func (s FailingStore) Save(o Order) error { return s.Err }

// Get vždy vrací false.
func (s FailingStore) Get(id string) (Order, bool) { return Order{}, false }

// OrderService je use-case vrstva nad úložištěm.
type OrderService struct {
	store OrderStore
	svc   *Service
}

// NewOrderService sestaví službu. Chybějící store nebo svc → ErrMissingDependency.
func NewOrderService(store OrderStore, svc *Service) (*OrderService, error) {
	if store == nil || svc == nil {
		return nil, ErrMissingDependency
	}
	return &OrderService{store: store, svc: svc}, nil
}

// --- Stupeň: obtížný ---

// Place vytvoří objednávku přes Service.NewOrder a uloží ji.
// Neplatná objednávka se nesmí uložit. Selhání úložiště obal ErrStore
// i původní chybu (%w), aby platilo errors.Is pro obě.
func (s *OrderService) Place(customer string, totalCents int64) (Order, error) {
	// TODO
	return *new(Order), nil
}

// Wire sestaví OrderService z MemoryStore a Service se SystemClock a RandomIDGen.
func Wire() (*OrderService, error) {
	svc, err := NewService(SystemClock{}, RandomIDGen{})
	if err != nil {
		return nil, err
	}
	return NewOrderService(NewMemoryStore(), svc)
}
