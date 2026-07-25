// Package app drží porty na hranici aplikační vrstvy.
//
// Interface je u konzumenta: HTTP a use-case závisí na BookmarkStore / Cache,
// ne na Postgres ani Redis. Adaptéry implementují porty zvenku.
package app

import (
	"context"

	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/bookmark"
)

// Query je dotaz na hledání záložek.
type Query struct {
	Text   string
	Tag    string
	Limit  int
	Cursor string
}

// Page je stránka výsledků.
type Page struct {
	Items      []bookmark.Bookmark
	NextCursor string
	Total      int
}

// BookmarkStore je port persistence záložek.
type BookmarkStore interface {
	Add(ctx context.Context, b bookmark.Bookmark) error
	Get(ctx context.Context, id string) (bookmark.Bookmark, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, q Query) (Page, error)
	Ready(ctx context.Context) error
}

// Cache je port pro cache hot path (typicky GET podle ID).
type Cache interface {
	Get(ctx context.Context, id string) (bookmark.Bookmark, bool, error)
	Set(ctx context.Context, b bookmark.Bookmark) error
	Delete(ctx context.Context, id string) error
	Ready(ctx context.Context) error
}

// CachedStore obaluje BookmarkStore cache na Get a invalidací při zápisu.
type CachedStore struct {
	Store BookmarkStore
	Cache Cache
}

// Add uloží záložku a invaliduje případný starý cache klíč.
func (s CachedStore) Add(ctx context.Context, b bookmark.Bookmark) error {
	if err := s.Store.Add(ctx, b); err != nil {
		return err
	}
	if s.Cache != nil {
		_ = s.Cache.Set(ctx, b)
	}
	return nil
}

// Get nejdřív zkusí cache, při miss sahá do store a plní cache.
func (s CachedStore) Get(ctx context.Context, id string) (bookmark.Bookmark, error) {
	if s.Cache != nil {
		if b, ok, err := s.Cache.Get(ctx, id); err != nil {
			return bookmark.Bookmark{}, err
		} else if ok {
			return b, nil
		}
	}
	b, err := s.Store.Get(ctx, id)
	if err != nil {
		return bookmark.Bookmark{}, err
	}
	if s.Cache != nil {
		_ = s.Cache.Set(ctx, b)
	}
	return b, nil
}

// Delete maže ze store i z cache.
func (s CachedStore) Delete(ctx context.Context, id string) error {
	if err := s.Store.Delete(ctx, id); err != nil {
		return err
	}
	if s.Cache != nil {
		_ = s.Cache.Delete(ctx, id)
	}
	return nil
}

// Search jde vždy do store (search se necachuje).
func (s CachedStore) Search(ctx context.Context, q Query) (Page, error) {
	return s.Store.Search(ctx, q)
}

// Ready vyžaduje připravený store; cache je volitelná.
func (s CachedStore) Ready(ctx context.Context) error {
	if err := s.Store.Ready(ctx); err != nil {
		return err
	}
	if s.Cache != nil {
		return s.Cache.Ready(ctx)
	}
	return nil
}
