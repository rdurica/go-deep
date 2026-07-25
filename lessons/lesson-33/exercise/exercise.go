// Package exercise obsahuje cvičení lekce 33.
package exercise

import (
	"errors"
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

// Now vrací aktuální systémový čas.
func (SystemClock) Now() time.Time {
	// TODO: úkol C
	return *new(time.Time)
}

// RandomIDGen je driven adaptér portu IDGen nad crypto/rand.
type RandomIDGen struct{}

// NewID vrací 32 hexadecimálních znaků, tedy 16 náhodných bajtů.
func (RandomIDGen) NewID() string {
	// TODO: úkol C
	return ""
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
	// TODO: úkol A
	return nil, nil
}

// NewOrder sestaví novou objednávku. Jméno zákazníka ořízne o bílé znaky,
// prázdné jméno je ErrEmptyCustomer, nekladná částka ErrInvalidTotal.
func (s *Service) NewOrder(customer string, totalCents int64) (Order, error) {
	// TODO: úkol A
	return *new(Order), nil
}

// MemoryStore je in-memory adaptér portu OrderStore. Je bezpečný pro
// souběžné použití.
type MemoryStore struct {
	mu     sync.RWMutex
	orders map[string]Order
}

// NewMemoryStore vytvoří prázdné úložiště.
func NewMemoryStore() *MemoryStore {
	// TODO: úkol B
	return nil
}

// Save uloží objednávku; existující se stejným ID přepíše.
func (s *MemoryStore) Save(o Order) error {
	// TODO: úkol B
	return nil
}

// Get vrací objednávku podle ID a příznak, jestli existuje.
func (s *MemoryStore) Get(id string) (Order, bool) {
	// TODO: úkol B
	return *new(Order), false
}

// FailingStore je adaptér, který při zápisu vždy selže. Slouží k testům
// chybové cesty — fake místo mockovacího frameworku.
type FailingStore struct {
	Err    error
	Orders map[string]Order
}

// Save vždy vrací nastavenou chybu.
func (s FailingStore) Save(o Order) error {
	// TODO: úkol B
	return nil
}

// Get vrací objednávku z předvyplněné mapy.
func (s FailingStore) Get(id string) (Order, bool) {
	// TODO: úkol B
	return *new(Order), false
}

// OrderService je doménová služba nad úložištěm. Skládá tři porty a nezná
// žádnou jejich konkrétní implementaci.
type OrderService struct {
	store OrderStore
	svc   *Service
}

// NewOrderService sestaví službu ze tří portů. Chybějící port vrací
// ErrMissingDependency.
func NewOrderService(store OrderStore, clock Clock, ids IDGen) (*OrderService, error) {
	// TODO: úkol B
	return nil, nil
}

// Place vytvoří objednávku a uloží ji. Chybu úložiště obalí ErrStore
// a zachová původní příčinu.
func (s *OrderService) Place(customer string, totalCents int64) (Order, error) {
	// TODO: úkol B
	return *new(Order), nil
}

// Find vrací objednávku podle ID, jinak chybu obalující ErrNotFound.
func (s *OrderService) Find(id string) (Order, error) {
	// TODO: úkol B
	return *new(Order), nil
}

// Cancel stornuje objednávku. Neznámé ID je ErrNotFound, opakované storno
// ErrAlreadyCanceled, selhání zápisu ErrStore s původní příčinou.
func (s *OrderService) Cancel(id string) (Order, error) {
	// TODO: úkol B
	return *new(Order), nil
}

// Wire sestaví službu z produkčních adaptérů. Tohle je celý DI kontejner:
// jeden výraz, ve kterém je vidět, co na čem visí.
func Wire() (*OrderService, error) {
	// TODO: úkol C
	return nil, nil
}
