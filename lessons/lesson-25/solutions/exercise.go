// Package solutions obsahuje referenční řešení lekce 25.
package solutions

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// writeError je zkratka pro chybovou odpověď v jednotném tvaru.
func writeError(w http.ResponseWriter, status int, msg string) {
	_ = WriteJSON(w, status, ErrorResponse{Error: msg})
}

// ItemsRouter vrací router nad úložištěm položek.
func ItemsRouter(store *Store) http.Handler {
	mux := http.NewServeMux()

	// {$} znamená "přesně kořen" — bez něj by vzor "/" chytal úplně všechno.
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		_ = WriteJSON(w, http.StatusOK, ServiceInfo{Service: "items"})
	})

	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		item, ok := store.Get(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "item not found")
			return
		}
		_ = WriteJSON(w, http.StatusOK, item)
	})

	mux.HandleFunc("DELETE /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !store.Delete(r.PathValue("id")) {
			writeError(w, http.StatusNotFound, "item not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		limit, q, err := ParseListQuery(r.URL.Query())
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		out := make([]Item, 0)
		for _, item := range store.List() {
			if q != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(q)) {
				continue
			}
			out = append(out, item)
			if limit > 0 && len(out) == limit {
				break
			}
		}
		_ = WriteJSON(w, http.StatusOK, out)
	})

	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		var req CreateItemRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		name := strings.TrimSpace(req.Name)
		if name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}

		item := store.Add(name)
		w.Header().Set("Location", "/items/"+item.ID)
		_ = WriteJSON(w, http.StatusCreated, item)
	})

	return mux
}

// ParseListQuery přečte a zvaliduje query parametry limit a q.
func ParseListQuery(values url.Values) (limit int, q string, err error) {
	q = strings.TrimSpace(values.Get("q"))

	raw := strings.TrimSpace(values.Get("limit"))
	if raw == "" {
		return 0, q, nil
	}

	n, convErr := strconv.Atoi(raw)
	if convErr != nil {
		return 0, "", ErrInvalidQuery
	}
	if n < 1 {
		return 0, "", ErrInvalidQuery
	}
	return n, q, nil
}

// SafeJoin složí cestu k souboru pod kořenem root a odmítne pokus o únik z něj.
func SafeJoin(root, rel string) (string, error) {
	if rel == "" || strings.HasPrefix(rel, "/") || filepath.IsAbs(rel) {
		return "", ErrInvalidPath
	}
	for _, segment := range strings.Split(rel, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", ErrInvalidPath
		}
	}

	full := filepath.Join(root, filepath.FromSlash(rel))

	// Pás a šle: i po kontrole segmentů ověř, že výsledek opravdu leží pod kořenem.
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", ErrInvalidPath
	}
	absFull, err := filepath.Abs(full)
	if err != nil {
		return "", ErrInvalidPath
	}
	if !strings.HasPrefix(absFull, absRoot+string(filepath.Separator)) {
		return "", ErrInvalidPath
	}
	return full, nil
}

// FilesHandler servíruje soubory pod adresářem root podle wildcardu {path...}.
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
