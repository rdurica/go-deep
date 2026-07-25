// Package solutions obsahuje referenční řešení checkpointu fáze 2.
package solutions

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
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

// Describe vrátí čitelný popis položky ve tvaru "Name: Qty ks".
func Describe(l Loader, sku string) (string, error) {
	if l == nil {
		return "", ErrMissingLoader
	}
	if sku == "" {
		return "", ErrEmptySKU
	}

	r, err := l.Load(sku)
	if err != nil {
		return "", fmt.Errorf("describe %q: %w", sku, err)
	}
	return fmt.Sprintf("%s: %d ks", r.Name, r.Qty), nil
}

// LoadRecords načte položky z textového vstupu ve tvaru sku;name;qty.
func LoadRecords(r io.Reader) ([]Record, error) {
	var records []Record
	seen := make(map[string]bool)

	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		rec, err := parseRecord(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", n, err)
		}
		if seen[rec.SKU] {
			return nil, fmt.Errorf("line %d: %w: %q", n, ErrDuplicateSKU, rec.SKU)
		}

		seen[rec.SKU] = true
		records = append(records, rec)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read records: %w", err)
	}
	return records, nil
}

// parseRecord rozebere jeden řádek na položku.
func parseRecord(line string) (Record, error) {
	fields := strings.Split(line, ";")
	if len(fields) != 3 {
		return Record{}, fmt.Errorf("%w: %q", ErrMalformedLine, line)
	}

	sku := strings.TrimSpace(fields[0])
	if sku == "" {
		return Record{}, ErrEmptySKU
	}

	qty, err := strconv.Atoi(strings.TrimSpace(fields[2]))
	if err != nil || qty < 0 {
		return Record{}, fmt.Errorf("%w: %q", ErrInvalidQty, strings.TrimSpace(fields[2]))
	}

	return Record{SKU: sku, Name: strings.TrimSpace(fields[1]), Qty: qty}, nil
}

// Store je sklad v paměti. Jeho zero value je prázdný, použitelný sklad.
type Store struct {
	records map[string]Record
}

// Put uloží nebo přepíše položku.
func (s *Store) Put(r Record) error {
	if r.SKU == "" {
		return ErrEmptySKU
	}
	if r.Qty < 0 {
		return fmt.Errorf("%w: %d", ErrInvalidQty, r.Qty)
	}

	if s.records == nil {
		s.records = make(map[string]Record)
	}
	s.records[r.SKU] = r
	return nil
}

// PutAll uloží všechny platné položky a spojí chyby těch neplatných.
func (s *Store) PutAll(records []Record) error {
	var errs []error
	for i, r := range records {
		if err := s.Put(r); err != nil {
			errs = append(errs, fmt.Errorf("record %d: %w", i, err))
		}
	}
	return errors.Join(errs...)
}

// Load vrací položku podle SKU, a tím Store splňuje Loader.
func (s *Store) Load(sku string) (Record, error) {
	r, ok := s.records[sku]
	if !ok {
		return Record{}, fmt.Errorf("%w: %q", ErrNotFound, sku)
	}
	return r, nil
}

// Remove smaže položku podle SKU.
func (s *Store) Remove(sku string) error {
	if _, ok := s.records[sku]; !ok {
		return fmt.Errorf("%w: %q", ErrNotFound, sku)
	}
	delete(s.records, sku)
	return nil
}

// List vrací všechny položky seřazené podle SKU.
func (s *Store) List() []Record {
	out := make([]Record, 0, len(s.records))
	for _, r := range s.records {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SKU < out[j].SKU })
	return out
}

// TotalQty vrací součet množství všech položek.
func (s *Store) TotalQty() int {
	total := 0
	for _, r := range s.records {
		total += r.Qty
	}
	return total
}
