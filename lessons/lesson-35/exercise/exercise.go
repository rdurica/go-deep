// Package exercise obsahuje cvičení lekce 35 — persistence.
//
// Doménový typ User tady nemá jediný `db:` tag a neví o SQL. Port UserRepo
// je definovaný u konzumenta, první adaptér je in-memory. SQL se v lekci
// probírá teoreticky, aby cvičení nepotřebovalo databázi ani driver.
package exercise

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"
)

// Chyby repozitáře, stavitele dotazů a migrací.
var (
	// ErrNotFound je doménový překlad "nic nenalezeno" (v SQL sql.ErrNoRows).
	ErrNotFound = errors.New("store: user not found")
	// ErrInvalidUser hlásí uživatele bez ID nebo bez e-mailu.
	ErrInvalidUser = errors.New("store: invalid user")
	// ErrUnknownTable hlásí tabulku mimo whitelist.
	ErrUnknownTable = errors.New("query: unknown table")
	// ErrUnknownColumn hlásí sloupec mimo whitelist.
	ErrUnknownColumn = errors.New("query: unknown column")
	// ErrNoColumns hlásí SELECT bez sloupců.
	ErrNoColumns = errors.New("query: no columns selected")
	// ErrDuplicateVersion hlásí dvě migrace se stejnou verzí.
	ErrDuplicateVersion = errors.New("migrate: duplicate version")
	// ErrDrift hlásí aplikovanou verzi, která v seznamu migrací chybí.
	ErrDrift = errors.New("migrate: applied migration missing")
	// ErrInvalidMigration hlásí migraci s nekladnou verzí nebo bez jména.
	ErrInvalidMigration = errors.New("migrate: invalid migration")
)

// User je doménový typ. Žádné `db:` tagy, žádné odkazy na tabulku — persistence
// je detail adaptéru, ne domény.
type User struct {
	ID     string
	Email  string
	Name   string
	Active bool
}

// UserRepo je port pro uložení uživatelů. Definuje si ho konzument, takže
// obsahuje jen to, co doména opravdu potřebuje. Context je první parametr
// každé metody, protože každý dotaz musí jít zrušit.
type UserRepo interface {
	Get(ctx context.Context, id string) (User, error)
	Save(ctx context.Context, u User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]User, error)
}

// MemoryRepo je in-memory adaptér portu UserRepo. Je bezpečný pro souběžné
// použití a chová se stejně jako SQL adaptér — včetně převodu "nenalezeno"
// na ErrNotFound.
type MemoryRepo struct {
	mu    sync.RWMutex
	users map[string]User
}

// NewMemoryRepo vytvoří prázdný in-memory repozitář.
func NewMemoryRepo() *MemoryRepo {
	// TODO: úkol A
	return nil
}

// Get vrátí uživatele podle ID. Neznámé ID je chyba obalující ErrNotFound.
func (r *MemoryRepo) Get(ctx context.Context, id string) (User, error) {
	// TODO: úkol A
	return *new(User), nil
}

// Save uloží uživatele (upsert). Prázdné ID nebo prázdný e-mail je ErrInvalidUser.
func (r *MemoryRepo) Save(ctx context.Context, u User) error {
	// TODO: úkol A
	return nil
}

// Delete smaže uživatele. Neznámé ID je chyba obalující ErrNotFound.
func (r *MemoryRepo) Delete(ctx context.Context, id string) error {
	// TODO: úkol A
	return nil
}

// List vrátí všechny uživatele seřazené vzestupně podle ID.
func (r *MemoryRepo) List(ctx context.Context) ([]User, error) {
	// TODO: úkol A
	return nil, nil
}

// schema je whitelist tabulek a jejich sloupců. Jméno tabulky ani sloupce
// nejde předat placeholderem, takže jediná bezpečná obrana je seznam povolených
// hodnot na straně serveru.
var schema = map[string][]string{
	"users":  {"id", "email", "name", "active"},
	"orders": {"id", "user_id", "total_cents", "status"},
}

// BuildSelect sestaví bezpečný SELECT.
//
// Jména tabulky a sloupců se ověřují proti whitelistu, hodnoty filtru jdou
// výhradně do placeholderů. Filtry se řadí abecedně podle sloupce, aby byl
// dotaz deterministický.
//
//	BuildSelect("users", []string{"id", "email"}, map[string]any{"active": true})
//	-> "SELECT id, email FROM users WHERE active = $1", []any{true}, nil
func BuildSelect(table string, cols []string, filters map[string]any) (query string, args []any, err error) {
	// TODO: úkol B
	return
}

// Migration je jedna dopředná migrace schématu. Verze je pořadí, Up je SQL.
type Migration struct {
	Version int
	Name    string
	Up      string
}

// Plan spočítá, které migrace ještě dobíhají.
//
// Vrací je seřazené vzestupně podle verze. Duplicitní verze v seznamu je
// ErrDuplicateVersion, aplikovaná verze chybějící v seznamu je ErrDrift
// (někdo smazal migraci, která už na produkci proběhla).
func Plan(applied []int, all []Migration) ([]Migration, error) {
	// TODO: úkol C
	return nil, nil
}

// allowedColumn hlásí, jestli je sloupec ve whitelistu dané tabulky.
func allowedColumn(table, col string) bool {
	for _, c := range schema[table] {
		if c == col {
			return true
		}
	}
	return false
}

// sortedFilterKeys vrací klíče filtru v abecedním pořadí.
func sortedFilterKeys(filters map[string]any) []string {
	keys := make([]string, 0, len(filters))
	for k := range filters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// placeholder vrací zástupný symbol pro n-tý argument, číslováno od jedničky.
func placeholder(n int) string { return "$" + strconv.Itoa(n) }

// notFound obalí ErrNotFound identifikátorem, aby chyba nesla kontext.
func notFound(id string) error {
	return fmt.Errorf("store: %q: %w", id, ErrNotFound)
}
