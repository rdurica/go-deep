// Package memstore je in-memory implementace app.BookmarkStore pro testy.
package memstore

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/app"
	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/bookmark"
)

// Store je úložiště záložek s indexem podle tagu.
type Store struct {
	mu    sync.RWMutex
	items map[string]bookmark.Bookmark
	byURL map[string]string
	byTag map[string]map[string]struct{}
}

// New vytvoří prázdné úložiště.
func New() *Store {
	return &Store{
		items: make(map[string]bookmark.Bookmark),
		byURL: make(map[string]string),
		byTag: make(map[string]map[string]struct{}),
	}
}

// Add uloží novou záložku. Duplicitní ID nebo URL je chyba.
func (s *Store) Add(_ context.Context, b bookmark.Bookmark) error {
	if err := b.Validate(); err != nil {
		return err
	}
	stored := bookmark.Clone(b)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.items[stored.ID]; exists {
		return bookmark.ErrDuplicateID
	}
	if _, exists := s.byURL[stored.URL]; exists {
		return bookmark.ErrDuplicateURL
	}

	s.items[stored.ID] = stored
	s.byURL[stored.URL] = stored.ID
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
func (s *Store) Get(_ context.Context, id string) (bookmark.Bookmark, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, ok := s.items[id]
	if !ok {
		return bookmark.Bookmark{}, bookmark.ErrNotFound
	}
	return bookmark.Clone(b), nil
}

// Delete smaže záložku podle ID.
func (s *Store) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.items[id]
	if !ok {
		return bookmark.ErrNotFound
	}
	delete(s.items, id)
	delete(s.byURL, b.URL)
	for _, tag := range b.Tags {
		ids := s.byTag[tag]
		delete(ids, id)
		if len(ids) == 0 {
			delete(s.byTag, tag)
		}
	}
	return nil
}

// Ready vrací nil, pokud je úložiště použitelné.
func (s *Store) Ready(context.Context) error {
	if s == nil {
		return errors.New("memstore: nil")
	}
	return nil
}

func sortNewest(items []bookmark.Bookmark) {
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].CreatedAt.After(items[j].CreatedAt)
		}
		return items[i].ID < items[j].ID
	})
}

// Search vrátí stránku záložek podle dotazu.
func (s *Store) Search(_ context.Context, q app.Query) (app.Page, error) {
	if q.Limit < 0 || q.Limit > bookmark.MaxLimit {
		return app.Page{}, bookmark.ErrInvalidQuery
	}
	limit := q.Limit
	if limit == 0 {
		limit = bookmark.DefaultLimit
	}

	text := strings.ToLower(strings.TrimSpace(q.Text))
	tag := strings.ToLower(strings.TrimSpace(q.Tag))
	if tag != "" {
		norm, err := bookmark.NormalizeTags([]string{tag})
		if err != nil || len(norm) != 1 {
			return app.Page{}, bookmark.ErrInvalidQuery
		}
		tag = norm[0]
	}

	s.mu.RLock()
	matched := make([]bookmark.Bookmark, 0, len(s.items))
	if tag != "" {
		for id := range s.byTag[tag] {
			b := s.items[id]
			if text != "" && !strings.Contains(strings.ToLower(b.Title), text) &&
				!strings.Contains(strings.ToLower(b.URL), text) {
				continue
			}
			matched = append(matched, bookmark.Clone(b))
		}
	} else {
		for _, b := range s.items {
			if text != "" && !strings.Contains(strings.ToLower(b.Title), text) &&
				!strings.Contains(strings.ToLower(b.URL), text) {
				continue
			}
			matched = append(matched, bookmark.Clone(b))
		}
	}
	s.mu.RUnlock()

	sortNewest(matched)

	page := app.Page{Total: len(matched)}
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
			return app.Page{}, bookmark.ErrInvalidCursor
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
