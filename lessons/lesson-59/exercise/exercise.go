// Package exercise obsahuje cvičení lekce 59.
package exercise

import (
	"errors"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
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

// --- Stupeň: jednoduchý ---
// NormalizeURL sjednotí URL podle pravidel: trim; jen http/https (jinak ErrInvalidURL);
// malý host; pryč výchozí port; prázdný host je chyba; pryč fragment; pryč parametry utm_*
// (case-insensitive); zbytek query seřazený podle klíče; useknuté koncové lomítko.
// Musí být idempotentní: NormalizeURL(NormalizeURL(x)) == NormalizeURL(x).
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ — chybí odstranění utm_* parametrů.
// Najdi chybu a oprav — testy před opravou padají.
func NormalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidURL
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", ErrInvalidURL
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", ErrInvalidURL
	}
	u.Scheme = scheme

	host := strings.ToLower(u.Hostname())
	if host == "" {
		return "", ErrInvalidURL
	}
	port := u.Port()
	if (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
		port = ""
	}
	u.Host = host
	if port != "" {
		u.Host = host + ":" + port
	}

	u.Fragment = ""
	u.RawFragment = ""

	if u.Path != "" {
		u.Path = strings.TrimRight(u.Path, "/")
		u.RawPath = ""
	}
	return u.String(), nil
}

// validTag vrací true pro tag složený z malých písmen, číslic a pomlček.
func validTag(tag string) bool {
	if tag == "" {
		return false
	}
	for _, r := range tag {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
		default:
			return false
		}
	}
	return true
}

// NormalizeTags ořeže, převede na malá písmena, odstraní duplicity a seřadí tagy.
func NormalizeTags(tags []string) ([]string, error) {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		if !validTag(tag) {
			return nil, ErrInvalidTag
		}
		if _, dup := seen[tag]; dup {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	if len(out) > MaxTags {
		return nil, ErrTooManyTags
	}
	sort.Strings(out)
	return out, nil
}

// Validate je hodnotový receiver — nic nemění. Kontroluje: neprázdné ID (ErrEmptyID),
// URL v normalizovaném tvaru (ErrInvalidURL), neprázdný titulek (ErrEmptyTitle)
// do MaxTitleLen run (ErrTitleTooLong), tagy (ErrInvalidTag), bez duplicit (ErrDuplicateTag)
// a nanejvýš MaxTags (ErrTooManyTags).
func (b Bookmark) Validate() error {
	if strings.TrimSpace(b.ID) == "" {
		return ErrEmptyID
	}
	norm, err := NormalizeURL(b.URL)
	if err != nil {
		return err
	}
	if norm != b.URL {
		return ErrInvalidURL
	}
	if strings.TrimSpace(b.Title) == "" {
		return ErrEmptyTitle
	}
	if utf8.RuneCountInString(b.Title) > MaxTitleLen {
		return ErrTitleTooLong
	}
	if len(b.Tags) > MaxTags {
		return ErrTooManyTags
	}
	seen := make(map[string]struct{}, len(b.Tags))
	for _, tag := range b.Tags {
		if !validTag(tag) {
			return ErrInvalidTag
		}
		if _, dup := seen[tag]; dup {
			return ErrDuplicateTag
		}
		seen[tag] = struct{}{}
	}
	return nil
}

// --- Stupeň: střední ---
// New normalizuje ID (trim), URL, titulek (trim) a tagy, pak zavolá Validate.
// Jediné místo, kde vzniká platná záložka.
func New(id, rawURL, title string, tags []string, createdAt time.Time) (Bookmark, error) {
	// TODO
	return Bookmark{}, nil
}

// Store je in-memory úložiště se indexem podle tagu, bezpečné pro souběžné použití (-race).
// Nesmí sdílet Tags s volajícím ani na vstupu, ani na výstupu.
type Store struct {
	mu    sync.RWMutex
	items map[string]Bookmark
	byTag map[string]map[string]struct{}
}

// clone vrátí hlubokou kopii záložky.
func clone(b Bookmark) Bookmark {
	out := b
	if b.Tags != nil {
		out.Tags = make([]string, len(b.Tags))
		copy(out.Tags, b.Tags)
	}
	return out
}

// NewStore vytvoří úložiště s připravenými (ne-nil) mapami.
func NewStore() *Store {
	return &Store{
		items: make(map[string]Bookmark),
		byTag: make(map[string]map[string]struct{}),
	}
}

// Add nejdřív Validate — neplatná záložka se neuloží. Při kolizi ID → ErrDuplicateID.
func (s *Store) Add(b Bookmark) error {
	if err := b.Validate(); err != nil {
		return err
	}
	stored := clone(b)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.items[stored.ID]; exists {
		return ErrDuplicateID
	}
	s.items[stored.ID] = stored
	for _, tag := range stored.Tags {
		ids, ok := s.byTag[tag]
		if !ok {
			ids = make(map[string]struct{})
			s.byTag[tag] = ids
		}
		ids[stored.ID] = struct{}{}
	}
	return nil
}

// Get vrátí kopii záložky (včetně kopie Tags). Neznámé ID → ErrNotFound.
func (s *Store) Get(id string) (Bookmark, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, ok := s.items[id]
	if !ok {
		return Bookmark{}, ErrNotFound
	}
	return clone(b), nil
}

// Delete smaže položku i všechny její záznamy v indexu.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.items[id]
	if !ok {
		return ErrNotFound
	}
	delete(s.items, id)
	for _, tag := range b.Tags {
		ids := s.byTag[tag]
		delete(ids, id)
		if len(ids) == 0 {
			delete(s.byTag, tag)
		}
	}
	return nil
}

// Len vrací počet uložených záložek.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// sortNewest seřadí záložky od nejnovější, při shodě podle ID.
func sortNewest(items []Bookmark) {
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].CreatedAt.After(items[j].CreatedAt)
		}
		return items[i].ID < items[j].ID
	})
}

// ByTag normalizuje tag stejně jako při ukládání (trim, malá písmena).
func (s *Store) ByTag(tag string) []Bookmark {
	tag = strings.ToLower(strings.TrimSpace(tag))

	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := s.byTag[tag]
	out := make([]Bookmark, 0, len(ids))
	for id := range ids {
		out = append(out, clone(s.items[id]))
	}
	sortNewest(out)
	return out
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

// --- Stupeň: obtížný ---
// Search filtruje, řadí a stránkuje.
// Validace: Limit < 0 nebo > MaxLimit / neznámé Sort / neplatný tag → ErrInvalidQuery;
// Limit == 0 znamená DefaultLimit.
// Filtr: prázdné Tags = bez omezení; MatchAll false = OR, true = AND; Text = case-insensitive
// podřetězec v titulku.
// Řazení: SortNewest = CreatedAt sestupně (tiebreak ID); SortTitle = titulek vzestupně
// case-insensitive (tiebreak ID).
// Stránkování: Cursor = ID poslední položky předchozí stránky (není ve výsledku → ErrInvalidCursor);
// NextCursor = ID poslední vrácené, pokud ještě něco zbývá, jinak ""; Total = počet před stránkováním.
func (s *Store) Search(q Query) (Page, error) {
	// TODO
	return Page{}, nil
}
