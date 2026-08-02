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

// --- Stupeň: jednoduchý ---
// Now vrací aktuální systémový čas přes time.Now().
// Driven adaptér portu Clock — doména nesmí volat time.Now() přímo, test to pozná.
func (SystemClock) Now() time.Time {
	// TODO
	return *new(time.Time)
}

// RandomIDGen je driven adaptér portu IDGen nad crypto/rand.
type RandomIDGen struct{}

// NewID vrací 32 hexadecimálních znaků (16 bajtů z crypto/rand přes encoding/hex).
// Test ověřuje tvar ^[0-9a-f]{32}$ i to, že se pět set ID neopakuje.
func (RandomIDGen) NewID() string {
	// TODO
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

// NewService sestaví službu ze dvou portů Clock a IDGen.
// Chybějící kterýkoli port (i oba) vrací ErrMissingDependency.
func NewService(clock Clock, ids IDGen) (*Service, error) {
	// TODO
	return nil, nil
}

// NewOrder sestaví novou objednávku. Jméno zákazníka ořízne o bílé znaky,
// prázdné jméno → ErrEmptyCustomer, nekladná částka → ErrInvalidTotal.
// ID z IDGen, PlacedAt z Clock — nikde nevolej time.Now() přímo.
func (s *Service) NewOrder(customer string, totalCents int64) (Order, error) {
	// TODO
	return *new(Order), nil
}

// MemoryStore je in-memory adaptér portu OrderStore. Bezpečný pro souběžné
// použití (sync.RWMutex); uložení existujícího ID hodnotu přepíše.
type MemoryStore struct {
	mu     sync.RWMutex
	orders map[string]Order
}

// --- Stupeň: střední ---
// NewMemoryStore vytvoří prázdné úložiště s inicializovanou mapou.
// Bezpečné pro souběžné použití; test běží s -race. Save přepíše existující ID.
func NewMemoryStore() *MemoryStore {
	// TODO
	return nil
}

// Save uloží objednávku do mapy. Existující se stejným ID přepíše.
// Musí být bezpečné pro souběžné volání z více goroutin.
func (s *MemoryStore) Save(o Order) error {
	// TODO
	return nil
}

// Get vrací objednávku podle ID a příznak, zda existuje.
// Neznámé ID vrátí nulovou Order a false.
func (s *MemoryStore) Get(id string) (Order, bool) {
	// TODO
	return *new(Order), false
}

// FailingStore je adaptér, který při zápisu vždy selže. Slouží k testům
// chybové cesty — fake místo mockovacího frameworku.
type FailingStore struct {
	Err    error
	Orders map[string]Order
}

// Save vždy vrací pole Err bez ohledu na předanou objednávku.
// Fake adaptér pro testování chybové cesty úložiště.
func (s FailingStore) Save(o Order) error {
	// TODO
	return nil
}

// Get čte objednávku z mapy Orders. Funguje i když je Orders nil.
func (s FailingStore) Get(id string) (Order, bool) {
	// TODO
	return *new(Order), false
}

// OrderService je doménová služba nad úložištěm. Skládá tři porty a nezná
// žádnou jejich konkrétní implementaci.
type OrderService struct {
	store OrderStore
	svc   *Service
}

// --- Stupeň: obtížný ---
// NewOrderService sestaví službu ze tří portů: store, clock a ids.
// Chybějící kterákoli závislost → ErrMissingDependency.
func NewOrderService(store OrderStore, clock Clock, ids IDGen) (*OrderService, error) {
	// TODO
	return nil, nil
}

// Place vytvoří objednávku přes Service.NewOrder a uloží ji.
// Neplatná objednávka se nesmí uložit. Selhání úložiště obal ErrStore
// i původní chybu (%w), aby platilo errors.Is pro obě.
func (s *OrderService) Place(customer string, totalCents int64) (Order, error) {
	// TODO
	return *new(Order), nil
}

// Find vrací objednávku podle ID z úložiště.
// Neznámé ID → chyba obalující ErrNotFound (errors.Is).
func (s *OrderService) Find(id string) (Order, error) {
	// TODO
	return *new(Order), nil
}

// Cancel stornuje objednávku: nastaví Canceled a uloží.
// Neznámé ID → ErrNotFound, už stornovaná → ErrAlreadyCanceled.
// Selhání zápisu obal jako u Place. Storno musí být vidět při dalším Find.
func (s *OrderService) Cancel(id string) (Order, error) {
	// TODO
	return *new(Order), nil
}

// Wire sestaví OrderService z produkčních adaptérů: MemoryStore, SystemClock, RandomIDGen.
// Jeden výraz, žádná logika — celý DI kontejner.
func Wire() (*OrderService, error) {
	// TODO
	return nil, nil
}
