// Package exercise obsahuje cvičení lekce 59.
package exercise

import (
	"errors"
	"sync"
	"time"
)

// Limity domény záložek.
const (
	MaxTitleLen  = 200
	MaxTags      = 10
	DefaultLimit = 20
	MaxLimit     = 100
)

// Doménové chyby.
var (
	ErrEmptyID       = errors.New("bookmark: prázdné ID")
	ErrInvalidURL    = errors.New("bookmark: neplatná URL")
	ErrEmptyTitle    = errors.New("bookmark: prázdný titulek")
	ErrTitleTooLong  = errors.New("bookmark: příliš dlouhý titulek")
	ErrInvalidTag    = errors.New("bookmark: neplatný tag")
	ErrTooManyTags   = errors.New("bookmark: příliš mnoho tagů")
	ErrDuplicateTag  = errors.New("bookmark: duplicitní tag")
	ErrDuplicateID   = errors.New("store: ID už existuje")
	ErrNotFound      = errors.New("store: záložka nenalezena")
	ErrInvalidQuery  = errors.New("store: neplatný dotaz")
	ErrInvalidCursor = errors.New("store: neplatný cursor")
)

// Bookmark je uložená záložka.
type Bookmark struct {
	ID        string
	URL       string
	Title     string
	Tags      []string
	CreatedAt time.Time
}

// NormalizeURL převede URL na kanonický tvar: malé schéma a host, bez výchozího portu,
// bez utm_ parametrů, bez fragmentu, s seřazeným query a bez koncového lomítka.
func NormalizeURL(raw string) (string, error) {
	panic("TODO: úkol A")
}

// NormalizeTags ořeže, převede na malá písmena, odstraní duplicity a seřadí tagy.
func NormalizeTags(tags []string) ([]string, error) {
	panic("TODO: úkol A")
}

// Validate zkontroluje, že je záložka kompletní a normalizovaná.
func (b Bookmark) Validate() error {
	panic("TODO: úkol A")
}

// New sestaví normalizovanou a ověřenou záložku.
func New(id, rawURL, title string, tags []string, createdAt time.Time) (Bookmark, error) {
	panic("TODO: úkol A")
}

// Store je in-memory úložiště záložek s indexem podle tagu, bezpečné pro souběžné použití.
type Store struct {
	mu    sync.RWMutex
	items map[string]Bookmark
	byTag map[string]map[string]struct{}
}

// NewStore vytvoří prázdné úložiště.
func NewStore() *Store {
	panic("TODO: úkol B")
}

// Add uloží novou záložku. Duplicitní ID vrací ErrDuplicateID.
func (s *Store) Add(b Bookmark) error {
	panic("TODO: úkol B")
}

// Get vrátí záložku podle ID.
func (s *Store) Get(id string) (Bookmark, error) {
	panic("TODO: úkol B")
}

// Delete smaže záložku podle ID i její záznamy v indexu.
func (s *Store) Delete(id string) error {
	panic("TODO: úkol B")
}

// Len vrací počet uložených záložek.
func (s *Store) Len() int {
	panic("TODO: úkol B")
}

// ByTag vrátí záložky s daným tagem od nejnovější.
func (s *Store) ByTag(tag string) []Bookmark {
	panic("TODO: úkol B")
}

// SortOrder je řazení výsledků hledání.
type SortOrder int

// Podporovaná řazení.
const (
	SortNewest SortOrder = iota
	SortTitle
)

// Query je dotaz na hledání záložek.
type Query struct {
	Tags     []string
	MatchAll bool
	Text     string
	Sort     SortOrder
	Limit    int
	Cursor   string
}

// Page je stránka výsledků hledání.
type Page struct {
	Items      []Bookmark
	NextCursor string
	Total      int
}

// Search vrátí stránku záložek podle dotazu.
func (s *Store) Search(q Query) (Page, error) {
	panic("TODO: úkol C")
}
