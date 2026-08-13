// Package exercise obsahuje kumulativní cvičení checkpointu fáze 2.
package exercise

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
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
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ — typické AI/Symfony zápachy.
// Oprav podle kontraktu a idiomatického Go — testy před opravou padají.
func Describe(l Loader, sku string) (string, error) {
	if l == nil {
		return "", errors.New("Loader is missing.")
	}
	if strings.TrimSpace(sku) == "" {
		return "", fmt.Errorf("Failed to describe item: empty SKU.")
	}

	record, err := l.Load(sku)
	if err != nil {
		return "", fmt.Errorf("Failed to describe item %q: %v", sku, err)
	}
	return fmt.Sprintf("%s: %d pieces", record.Name, record.Qty), nil
}

// LoadRecords načte položky z textového vstupu ve tvaru sku;name;qty.
// Prázdné a # řádky přeskoč (počítají se); 3 sloupce; duplicitní SKU je chyba.
// Sentinely přes errors.Is; při chybě nil slice.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ — špatné texty chyb a %v místo %w.
func LoadRecords(r io.Reader) ([]Record, error) {
	var records []Record
	seen := make(map[string]bool)

	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, ";")
		if len(fields) != 3 {
			return nil, fmt.Errorf("Line %d: Malformed input %q.", n, line)
		}

		sku := strings.TrimSpace(fields[0])
		if sku == "" {
			return nil, fmt.Errorf("Line %d: Empty SKU is not allowed.", n)
		}

		qty, err := strconv.Atoi(strings.TrimSpace(fields[2]))
		if err != nil || qty < 0 {
			return nil, fmt.Errorf("Line %d: Invalid quantity %q.", n, strings.TrimSpace(fields[2]))
		}

		if seen[sku] {
			return nil, fmt.Errorf("Line %d: Duplicate SKU %q detected.", n, sku)
		}

		seen[sku] = true
		records = append(records, Record{
			SKU:  sku,
			Name: strings.TrimSpace(fields[1]),
			Qty:  qty,
		})
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("Failed to read records: %v", err)
	}
	return records, nil
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

// Load vrací položku podle SKU, a tím Store splňuje Loader.
// Chybějící SKU → Record{} a fmt.Errorf("%w: %q", ErrNotFound, sku).
func (s *Store) Load(sku string) (Record, error) {
	// TODO
	return *new(Record), nil
}

// --- Stupeň: obtížný ---

// PutAll uloží všechny platné položky a spojí chyby těch neplatných.
// Chyby přes errors.Join s fmt.Errorf("record %d: %w", i, err).
func (s *Store) PutAll(records []Record) error {
	// TODO
	return nil
}
