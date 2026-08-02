// Package exercise obsahuje kumulativní cvičení checkpointu fáze 2.
package exercise

import (
	"errors"
	"io"
)

// Sentinelové chyby skladu.
var (
	ErrMissingLoader = errors.New("missing loader")
	ErrEmptySKU      = errors.New("empty sku")
	ErrInvalidQty    = errors.New("invalid quantity")
	ErrMalformedLine = errors.New("malformed line")
	ErrDuplicateSKU  = errors.New("duplicate sku")
	ErrNotFound      = errors.New("not found")
)

// Record je jedna skladová položka.
type Record struct {
	SKU  string
	Name string
	Qty  int
}

// Loader je minimální port definovaný u konzumenta — jedna metoda, ne celý sklad.
type Loader interface {
	Load(sku string) (Record, error)
}

// --- Stupeň: jednoduchý ---
// Describe vrátí čitelný popis položky ve tvaru "Name: Qty ks".
// nil loader → ErrMissingLoader; prázdné sku → ErrEmptySKU (loader se nevolá).
// Chyba Load → fmt.Errorf("describe %q: %w", sku, err).
func Describe(l Loader, sku string) (string, error) {
	// TODO
	return "", nil
}

// LoadRecords načte položky z textového vstupu ve tvaru sku;name;qty.
// Prázdné a # řádky přeskoč (počítají se); 3 sloupce; duplicitní SKU je chyba.
// Sentinely přes errors.Is; při chybě nil slice.
func LoadRecords(r io.Reader) ([]Record, error) {
	// TODO
	return nil, nil
}

// Store je sklad v paměti. Jeho zero value je prázdný, použitelný sklad.
type Store struct {
	records map[string]Record
}

// --- Stupeň: střední ---
// Put uloží nebo přepíše položku.
// Prázdné SKU → ErrEmptySKU; Qty < 0 → chyba obalující ErrInvalidQty. Neplatný záznam se neuloží.
func (s *Store) Put(r Record) error {
	// TODO
	return nil
}

// PutAll uloží všechny platné položky a spojí chyby těch neplatných.
// Chyby přes errors.Join s fmt.Errorf("record %d: %w", i, err).
func (s *Store) PutAll(records []Record) error {
	// TODO
	return nil
}

// Load vrací položku podle SKU, a tím Store splňuje Loader.
// Chybějící SKU → Record{} a fmt.Errorf("%w: %q", ErrNotFound, sku).
func (s *Store) Load(sku string) (Record, error) {
	// TODO
	return *new(Record), nil
}

// --- Stupeň: obtížný ---
// Remove smaže položku podle SKU.
// Chybějící SKU → stejná chyba jako u Load.
func (s *Store) Remove(sku string) error {
	// TODO
	return nil
}

// List vrací všechny položky seřazené podle SKU.
// Prázdný sklad → prázdný slice (kopie, ne nil panika).
func (s *Store) List() []Record {
	// TODO
	return nil
}

// TotalQty vrací součet množství všech položek.
func (s *Store) TotalQty() int {
	// TODO
	return 0
}
