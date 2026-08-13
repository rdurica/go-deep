// Package exercise obsahuje cvičení lekce 25.
package exercise

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	_ = WriteJSON(w, status, ErrorResponse{Error: msg})
}

// --- Stupeň: jednoduchý ---

// ParseListQuery přečte a zvaliduje query parametry limit a q.
// Chybějící limit → 0; jinak celé číslo ≥ 1, jinak ErrInvalidQuery. q ořízni od bílých znaků.
func ParseListQuery(values url.Values) (limit int, q string, err error) {
	// TODO
	return
}

// --- Stupeň: střední ---

// ItemsRouter vrací ServeMux nad úložištěm. Chybové odpovědi jako ErrorResponse JSON.
// GET /{$} → 200 ServiceInfo{Service:"items"} (přesný kořen).
// GET /items/{id} → 200 nebo 404.
// GET /items → pole v pořadí vložení; query přes ParseListQuery (chyba → 400);
// q filtruje jméno case-insensitive, limit ořízne počet; prázdný výsledek je [] ne null.
// POST /items → CreateItemRequest; rozbitý JSON/prázdné jméno → 400; úspěch 201 + Location.
// DELETE /items/{id} → 204 nebo 404. PUT a jiné metody řeší mux (405 + Allow) sám.
func ItemsRouter(store *Store) http.Handler {
	// TODO
	return *new(http.Handler)
}

// --- Stupeň: obtížný ---

// SafeJoin složí cestu k souboru pod kořenem root a odmítne pokus o únik z něj.
// Odmítni prázdný rel, absolutní cestu a segmenty "", ".", ".."; ověř prefix pod root.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. filepath.Join cestu „uklidí" a umožní únik.
// Najdi chybu a oprav — testy před opravou padají.
func SafeJoin(root, rel string) (string, error) {
	return filepath.Join(root, rel), nil
}

// FilesHandler servíruje soubory pod adresářem root podle wildcardu {path...}.
// Hotové — soustřeď se na SafeJoin. SafeJoin chyba → 400; neexistující nebo adresář → 404.
func FilesHandler(root string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		full, err := SafeJoin(root, r.PathValue("path"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid path")
			return
		}

		info, err := os.Stat(full)
		if err != nil || info.IsDir() {
			writeError(w, http.StatusNotFound, "file not found")
			return
		}

		http.ServeFile(w, r, full)
	})
}

// FilesRouter registruje FilesHandler na vzor "GET /files/{path...}".
func FilesRouter(root string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /files/{path...}", FilesHandler(root))
	return mux
}
