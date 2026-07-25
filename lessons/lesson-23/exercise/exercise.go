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

// Describe vrátí čitelný popis položky ve tvaru "Name: Qty ks".
func Describe(l Loader, sku string) (string, error) {
	panic("TODO: úkol A")
}

// LoadRecords načte položky z textového vstupu ve tvaru sku;name;qty.
func LoadRecords(r io.Reader) ([]Record, error) {
	panic("TODO: úkol B")
}

// Store je sklad v paměti. Jeho zero value je prázdný, použitelný sklad.
type Store struct {
	records map[string]Record
}

// Put uloží nebo přepíše položku.
func (s *Store) Put(r Record) error {
	panic("TODO: úkol C")
}

// PutAll uloží všechny platné položky a spojí chyby těch neplatných.
func (s *Store) PutAll(records []Record) error {
	panic("TODO: úkol C")
}

// Load vrací položku podle SKU, a tím Store splňuje Loader.
func (s *Store) Load(sku string) (Record, error) {
	panic("TODO: úkol C")
}

// Remove smaže položku podle SKU.
func (s *Store) Remove(sku string) error {
	panic("TODO: úkol C")
}

// List vrací všechny položky seřazené podle SKU.
func (s *Store) List() []Record {
	panic("TODO: úkol C")
}

// TotalQty vrací součet množství všech položek.
func (s *Store) TotalQty() int {
	panic("TODO: úkol C")
}
