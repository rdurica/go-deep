// Package solutions obsahuje referenční řešení lekce 59.
package solutions

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
// NormalizeURL převede URL na kanonický tvar: malé schéma a host, bez výchozího portu,
// bez utm_ parametrů, bez fragmentu, s seřazeným query a bez koncového lomítka.
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

	if u.RawQuery != "" {
		values := u.Query()
		for key := range values {
			if strings.HasPrefix(strings.ToLower(key), "utm_") {
				values.Del(key)
			}
		}
		u.RawQuery = values.Encode() // Encode řadí podle klíče, výsledek je deterministický
	}

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

// Validate zkontroluje, že je záložka kompletní a normalizovaná.
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
// New sestaví normalizovanou a ověřenou záložku.
func New(id, rawURL, title string, tags []string, createdAt time.Time) (Bookmark, error) {
	normURL, err := NormalizeURL(rawURL)
	if err != nil {
		return Bookmark{}, err
	}
	normTags, err := NormalizeTags(tags)
	if err != nil {
		return Bookmark{}, err
	}
	b := Bookmark{
		ID:        strings.TrimSpace(id),
		URL:       normURL,
		Title:     strings.TrimSpace(title),
		Tags:      normTags,
		CreatedAt: createdAt,
	}
	if err := b.Validate(); err != nil {
		return Bookmark{}, err
	}
	return b, nil
}

// clone vrátí hlubokou kopii záložky, aby volající nemohl přepsat obsah store.
func clone(b Bookmark) Bookmark {
	out := b
	if b.Tags != nil {
		out.Tags = make([]string, len(b.Tags))
		copy(out.Tags, b.Tags)
	}
	return out
}

// Store je in-memory úložiště záložek s indexem podle tagu, bezpečné pro souběžné použití.
type Store struct {
	mu    sync.RWMutex
	items map[string]Bookmark
	byTag map[string]map[string]struct{}
}

// NewStore vytvoří prázdné úložiště.
func NewStore() *Store {
	return &Store{
		items: make(map[string]Bookmark),
		byTag: make(map[string]map[string]struct{}),
	}
}

// Add uloží novou záložku. Duplicitní ID vrací ErrDuplicateID.
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

// Get vrátí záložku podle ID.
func (s *Store) Get(id string) (Bookmark, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, ok := s.items[id]
	if !ok {
		return Bookmark{}, ErrNotFound
	}
	return clone(b), nil
}

// --- Stupeň: obtížný ---
// Delete smaže záložku podle ID i její záznamy v indexu.
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

// ByTag vrátí záložky s daným tagem od nejnovější.
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

// matchesTags vrací true, pokud záložka splňuje tagovou část dotazu.
func matchesTags(b Bookmark, tags []string, all bool) bool {
	if len(tags) == 0 {
		return true
	}
	has := make(map[string]struct{}, len(b.Tags))
	for _, tag := range b.Tags {
		has[tag] = struct{}{}
	}
	for _, want := range tags {
		if _, ok := has[want]; ok {
			if !all {
				return true
			}
			continue
		}
		if all {
			return false
		}
	}
	return all
}

// Search vrátí stránku záložek podle dotazu.
func (s *Store) Search(q Query) (Page, error) {
	if q.Limit < 0 || q.Limit > MaxLimit {
		return Page{}, ErrInvalidQuery
	}
	if q.Sort != SortNewest && q.Sort != SortTitle {
		return Page{}, ErrInvalidQuery
	}
	tags, err := NormalizeTags(q.Tags)
	if err != nil {
		return Page{}, ErrInvalidQuery
	}
	limit := q.Limit
	if limit == 0 {
		limit = DefaultLimit
	}
	text := strings.ToLower(strings.TrimSpace(q.Text))

	s.mu.RLock()
	matched := make([]Bookmark, 0, len(s.items))
	for _, b := range s.items {
		if !matchesTags(b, tags, q.MatchAll) {
			continue
		}
		if text != "" && !strings.Contains(strings.ToLower(b.Title), text) {
			continue
		}
		matched = append(matched, clone(b))
	}
	s.mu.RUnlock()

	switch q.Sort {
	case SortTitle:
		sort.SliceStable(matched, func(i, j int) bool {
			ti, tj := strings.ToLower(matched[i].Title), strings.ToLower(matched[j].Title)
			if ti != tj {
				return ti < tj
			}
			return matched[i].ID < matched[j].ID
		})
	default:
		sortNewest(matched)
	}

	page := Page{Total: len(matched)}
	start := 0
	if q.Cursor != "" {
		idx := -1
		for i, b := range matched {
			if b.ID == q.Cursor {
				idx = i
				break
			}
		}
		if idx < 0 {
			return Page{}, ErrInvalidCursor
		}
		start = idx + 1
	}

	end := start + limit
	if end > len(matched) {
		end = len(matched)
	}
	page.Items = matched[start:end]
	if end < len(matched) && end > start {
		page.NextCursor = matched[end-1].ID
	}
	return page, nil
}
