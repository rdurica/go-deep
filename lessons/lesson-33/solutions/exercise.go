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

// Chyby domény objednávek. Jsou to sentinely, porovnávej je přes errors.Is.
var (
	// ErrMissingDependency hlásí chybějící port v konstruktoru.
	ErrMissingDependency = errors.New("ordering: missing dependency")
	// ErrEmptyCustomer hlásí objednávku bez zákazníka.
	ErrEmptyCustomer = errors.New("ordering: empty customer")
	// ErrInvalidTotal hlásí nekladnou částku.
	ErrInvalidTotal = errors.New("ordering: total must be positive")
	// ErrNotFound hlásí, že objednávka v úložišti není.
	ErrNotFound = errors.New("ordering: order not found")
	// ErrAlreadyCanceled hlásí opakované storno.
	ErrAlreadyCanceled = errors.New("ordering: order already canceled")
	// ErrStore hlásí selhání úložiště. Doména jím obaluje chyby adaptéru,
	// aby konzument nemusel znát SQL ani souborový systém.
	ErrStore = errors.New("ordering: store failed")
)

// Clock je port pro čtení času. Je definovaný tady, u konzumenta, protože
// jeho tvar určuje doména — ne implementace.
type Clock interface {
	Now() time.Time
}

// IDGen je port pro generování identifikátorů objednávek.
type IDGen interface {
	NewID() string
}

// OrderStore je port pro ukládání objednávek. Dvě metody stačí; kdyby jich
// bylo čtrnáct, je to znamení, že port patří někomu jinému.
type OrderStore interface {
	Save(o Order) error
	Get(id string) (Order, bool)
}

// SystemClock je driven adaptér portu Clock nad systémovým časem.
type SystemClock struct{}

// --- Stupeň: jednoduchý ---
// Now vrací aktuální systémový čas.
func (SystemClock) Now() time.Time { return time.Now() }

// RandomIDGen je driven adaptér portu IDGen nad crypto/rand.
type RandomIDGen struct{}

// NewID vrací 32 hexadecimálních znaků, tedy 16 náhodných bajtů.
func (RandomIDGen) NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("ordering: rand.Read selhalo: " + err.Error())
	}
	return hex.EncodeToString(b[:])
}

// Order je objednávka. Je to hodnotový typ — porovnatelný přes ==, protože
// všechna pole jsou porovnatelná.
type Order struct {
	ID         string
	Customer   string
	TotalCents int64
	PlacedAt   time.Time
	Canceled   bool
}

// Service je doménová služba, která umí objednávku jen sestavit. Zná dva
// porty a nic víc — žádné úložiště, žádné HTTP.
type Service struct {
	clock Clock
	ids   IDGen
}

// NewService sestaví službu ze dvou portů. Chybějící port vrací
// ErrMissingDependency.
func NewService(clock Clock, ids IDGen) (*Service, error) {
	if clock == nil || ids == nil {
		return nil, ErrMissingDependency
	}
	return &Service{clock: clock, ids: ids}, nil
}

// NewOrder sestaví novou objednávku. Jméno zákazníka ořízne o bílé znaky,
// prázdné jméno je ErrEmptyCustomer, nekladná částka ErrInvalidTotal.
func (s *Service) NewOrder(customer string, totalCents int64) (Order, error) {
	customer = strings.TrimSpace(customer)
	switch {
	case customer == "":
		return Order{}, ErrEmptyCustomer
	case totalCents <= 0:
		return Order{}, fmt.Errorf("%w: %d", ErrInvalidTotal, totalCents)
	}
	return Order{
		ID:         s.ids.NewID(),
		Customer:   customer,
		TotalCents: totalCents,
		PlacedAt:   s.clock.Now(),
	}, nil
}

// MemoryStore je in-memory adaptér portu OrderStore. Je bezpečný pro
// souběžné použití.
type MemoryStore struct {
	mu     sync.RWMutex
	orders map[string]Order
}

// --- Stupeň: střední ---
// NewMemoryStore vytvoří prázdné úložiště.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{orders: make(map[string]Order)}
}

// Save uloží objednávku; existující se stejným ID přepíše.
func (s *MemoryStore) Save(o Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[o.ID] = o
	return nil
}

// Get vrací objednávku podle ID a příznak, jestli existuje.
func (s *MemoryStore) Get(id string) (Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	return o, ok
}

// FailingStore je adaptér, který při zápisu vždy selže. Slouží k testům
// chybové cesty — fake místo mockovacího frameworku.
type FailingStore struct {
	Err    error
	Orders map[string]Order
}

// Save vždy vrací nastavenou chybu.
func (s FailingStore) Save(o Order) error { return s.Err }

// Get vrací objednávku z předvyplněné mapy.
func (s FailingStore) Get(id string) (Order, bool) {
	o, ok := s.Orders[id]
	return o, ok
}

// OrderService je doménová služba nad úložištěm. Skládá tři porty a nezná
// žádnou jejich konkrétní implementaci.
type OrderService struct {
	store OrderStore
	svc   *Service
}

// --- Stupeň: obtížný ---
// NewOrderService sestaví službu ze tří portů. Chybějící port vrací
// ErrMissingDependency.
func NewOrderService(store OrderStore, clock Clock, ids IDGen) (*OrderService, error) {
	if store == nil {
		return nil, ErrMissingDependency
	}
	svc, err := NewService(clock, ids)
	if err != nil {
		return nil, err
	}
	return &OrderService{store: store, svc: svc}, nil
}

// Place vytvoří objednávku a uloží ji. Chybu úložiště obalí ErrStore
// a zachová původní příčinu.
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

// Find vrací objednávku podle ID, jinak chybu obalující ErrNotFound.
func (s *OrderService) Find(id string) (Order, error) {
	o, ok := s.store.Get(id)
	if !ok {
		return Order{}, fmt.Errorf("find %s: %w", id, ErrNotFound)
	}
	return o, nil
}

// Cancel stornuje objednávku. Neznámé ID je ErrNotFound, opakované storno
// ErrAlreadyCanceled, selhání zápisu ErrStore s původní příčinou.
func (s *OrderService) Cancel(id string) (Order, error) {
	o, err := s.Find(id)
	if err != nil {
		return Order{}, err
	}
	if o.Canceled {
		return Order{}, fmt.Errorf("cancel %s: %w", id, ErrAlreadyCanceled)
	}
	o.Canceled = true
	if err := s.store.Save(o); err != nil {
		return Order{}, fmt.Errorf("cancel %s: %w: %w", id, ErrStore, err)
	}
	return o, nil
}

// Wire sestaví službu z produkčních adaptérů. Tohle je celý DI kontejner:
// jeden výraz, ve kterém je vidět, co na čem visí.
func Wire() (*OrderService, error) {
	return NewOrderService(NewMemoryStore(), SystemClock{}, RandomIDGen{})
}
