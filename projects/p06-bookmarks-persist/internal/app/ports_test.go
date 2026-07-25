package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/app"
	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/bookmark"
	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/memstore"
)

type memoryCache struct {
	items map[string]bookmark.Bookmark
	hits  int
}

func newMemoryCache() *memoryCache {
	return &memoryCache{items: make(map[string]bookmark.Bookmark)}
}

func (c *memoryCache) Get(_ context.Context, id string) (bookmark.Bookmark, bool, error) {
	b, ok := c.items[id]
	if !ok {
		return bookmark.Bookmark{}, false, nil
	}
	c.hits++
	return bookmark.Clone(b), true, nil
}

func (c *memoryCache) Set(_ context.Context, b bookmark.Bookmark) error {
	c.items[b.ID] = bookmark.Clone(b)
	return nil
}

func (c *memoryCache) Delete(_ context.Context, id string) error {
	delete(c.items, id)
	return nil
}

func (c *memoryCache) Ready(context.Context) error { return nil }

func TestCachedStoreGetHitsCache(t *testing.T) {
	ctx := context.Background()
	store := memstore.New()
	cache := newMemoryCache()
	cs := app.CachedStore{Store: store, Cache: cache}

	b, err := bookmark.New("bm_1", "https://example.com/a", "A", nil, time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := cs.Add(ctx, b); err != nil {
		t.Fatalf("Add: %v", err)
	}

	if _, err := cs.Get(ctx, "bm_1"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if cache.hits != 1 {
		t.Fatalf("po Add+Get chci 1 cache hit (Set při Add), hits=%d", cache.hits)
	}
	if _, err := cs.Get(ctx, "bm_1"); err != nil {
		t.Fatalf("Get 2: %v", err)
	}
	if cache.hits != 2 {
		t.Errorf("druhý Get má jít z cache, hits=%d", cache.hits)
	}

	if err := cs.Delete(ctx, "bm_1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := cs.Get(ctx, "bm_1"); !errors.Is(err, bookmark.ErrNotFound) {
		t.Errorf("po Delete chci ErrNotFound, dostal jsem %v", err)
	}
}
