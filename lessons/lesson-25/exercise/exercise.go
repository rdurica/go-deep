// Package exercise obsahuje cvičení lekce 25.
package exercise

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

// ErrInvalidQuery signalizuje neplatný query parametr.
var ErrInvalidQuery = errors.New("invalid query parameter")

// ErrInvalidPath signalizuje cestu, která se pokouší uniknout z kořenového adresáře.
var ErrInvalidPath = errors.New("invalid path")

// Item je položka v úložišti.
type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CreateItemRequest je tělo požadavku na vytvoření položky.
type CreateItemRequest struct {
	Name string `json:"name"`
}

// ErrorResponse je jednotný tvar chybové odpovědi.
type ErrorResponse struct {
	Error string `json:"error"`
}

// ServiceInfo je odpověď kořenového endpointu.
type ServiceInfo struct {
	Service string `json:"service"`
}

// Store je jednoduché in-memory úložiště položek.
// Mutex ber zatím jako black box — souběžnosti se věnuje fáze 5.
type Store struct {
	mu    sync.Mutex
	seq   int
	order []string
	items map[string]Item
}

// NewStore vytvoří prázdné úložiště.
func NewStore() *Store {
	return &Store{items: make(map[string]Item)}
}

// Add uloží novou položku, přidělí jí ID a vrátí ji.
func (s *Store) Add(name string) Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.seq++
	item := Item{ID: strconv.Itoa(s.seq), Name: name}
	s.items[item.ID] = item
	s.order = append(s.order, item.ID)
	return item
}

// Get vrátí položku podle ID.
func (s *Store) Get(id string) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	return item, ok
}

// Delete smaže položku podle ID a vrátí true, pokud existovala.
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	for i, existing := range s.order {
		if existing == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return true
}

// List vrátí všechny položky v pořadí vložení.
func (s *Store) List() []Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Item, 0, len(s.order))
	for _, id := range s.order {
		out = append(out, s.items[id])
	}
	return out
}

// WriteJSON zapíše v jako JSON odpověď se status kódem status.
// Hotové z lekce 24, tady se soustřeď na routing.
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// ItemsRouter vrací router nad úložištěm položek.
func ItemsRouter(store *Store) http.Handler {
	// TODO: úkol A
	return *new(http.Handler)
}

// ParseListQuery přečte a zvaliduje query parametry limit a q.
func ParseListQuery(values url.Values) (limit int, q string, err error) {
	// TODO: úkol B
	return
}

// SafeJoin složí cestu k souboru pod kořenem root a odmítne pokus o únik z něj.
func SafeJoin(root, rel string) (string, error) {
	// TODO: úkol C
	return "", nil
}

// FilesHandler servíruje soubory pod adresářem root podle wildcardu {path...}.
func FilesHandler(root string) http.Handler {
	// TODO: úkol C
	return *new(http.Handler)
}

// FilesRouter registruje FilesHandler na vzor "GET /files/{path...}".
func FilesRouter(root string) http.Handler {
	// TODO: úkol C
	return *new(http.Handler)
}
