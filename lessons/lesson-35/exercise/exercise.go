// Package exercise obsahuje cvičení lekce 35 — persistence.
package exercise

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrNotFound         = errors.New("store: user not found")
	ErrInvalidUser      = errors.New("store: invalid user")
	ErrUnknownTable     = errors.New("query: unknown table")
	ErrUnknownColumn    = errors.New("query: unknown column")
	ErrNoColumns        = errors.New("query: no columns selected")
	ErrDuplicateVersion = errors.New("migrate: duplicate version")
	ErrDrift            = errors.New("migrate: applied migration missing")
	ErrInvalidMigration = errors.New("migrate: invalid migration")
)

// User je doménový typ bez db tagů.
type User struct {
	ID     string
	Email  string
	Name   string
	Active bool
}

// UserRepo je port u konzumenta.
type UserRepo interface {
	Get(ctx context.Context, id string) (User, error)
	Save(ctx context.Context, u User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]User, error)
}

// MemoryRepo je in-memory adaptér.
type MemoryRepo struct {
	mu    sync.RWMutex
	users map[string]User
}

// NewMemoryRepo vytvoří prázdný repozitář.
func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{users: make(map[string]User)}
}

// --- Stupeň: jednoduchý ---

// Get vrátí uživatele podle ID. Neznámé ID → chyba obalující ErrNotFound.
// Nejdřív zkontroluj ctx.Err().
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. U chybějícího ID vrací nulového uživatele bez chyby.
func (r *MemoryRepo) Get(ctx context.Context, id string) (User, error) {
	if err := ctx.Err(); err != nil {
		return User{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return User{}, nil
	}
	return u, nil
}

// --- Stupeň: střední ---

// Save uloží uživatele (upsert). Prázdné ID nebo prázdný e-mail po trim → ErrInvalidUser.
// Respektuj ctx.Err().
func (r *MemoryRepo) Save(ctx context.Context, u User) error {
	// TODO
	return nil
}

// Delete smaže uživatele. Neznámé ID → chyba obalující ErrNotFound.
func (r *MemoryRepo) Delete(ctx context.Context, id string) error {
	// TODO
	return nil
}

// List vrátí všechny uživatele seřazené podle ID vzestupně.
func (r *MemoryRepo) List(ctx context.Context) ([]User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]User, 0, len(r.users))
	for _, u := range r.users {
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// --- Stupeň: obtížný ---

var schema = map[string][]string{
	"users":  {"id", "email", "name", "active"},
	"orders": {"id", "user_id", "total_cents", "status"},
}

// BuildSelect sestaví bezpečný SELECT z whitelistu tabulek a sloupců.
// Hodnoty filtru jen do args; klíče filtru abecedně.
func BuildSelect(table string, cols []string, filters map[string]any) (query string, args []any, err error) {
	// TODO
	return
}

// Migration je jedna dopředná migrace.
type Migration struct {
	Version int
	Name    string
	Up      string
}

// Plan spočítá migrace k doběhnutí. Kompletní implementace — neměň.
func Plan(applied []int, all []Migration) ([]Migration, error) {
	seen := make(map[int]struct{}, len(all))
	for _, m := range all {
		if m.Version <= 0 || strings.TrimSpace(m.Name) == "" {
			return nil, ErrInvalidMigration
		}
		if _, dup := seen[m.Version]; dup {
			return nil, ErrDuplicateVersion
		}
		seen[m.Version] = struct{}{}
	}
	appliedSet := make(map[int]struct{})
	for _, v := range applied {
		appliedSet[v] = struct{}{}
	}
	for v := range appliedSet {
		if _, ok := seen[v]; !ok {
			return nil, ErrDrift
		}
	}
	var pending []Migration
	for _, m := range all {
		if _, done := appliedSet[m.Version]; !done {
			pending = append(pending, m)
		}
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].Version < pending[j].Version })
	return pending, nil
}

func allowedColumn(table, col string) bool {
	for _, c := range schema[table] {
		if c == col {
			return true
		}
	}
	return false
}

func sortedFilterKeys(filters map[string]any) []string {
	keys := make([]string, 0, len(filters))
	for k := range filters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func placeholder(n int) string { return "$" + strconv.Itoa(n) }

func notFound(id string) error {
	return fmt.Errorf("store: %q: %w", id, ErrNotFound)
}
