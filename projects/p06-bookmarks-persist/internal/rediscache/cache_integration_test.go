package rediscache_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/bookmark"
	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/rediscache"
)

func TestRedisCacheRoundTrip(t *testing.T) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		t.Skip("REDIS_URL není nastavená — přeskočeno")
	}
	ctx := context.Background()
	c, err := rediscache.Open(ctx, url, time.Minute)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })

	b, err := bookmark.New("bm_cache", "https://example.com/cache", "Cache", []string{"go"}, time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_ = c.Delete(ctx, b.ID)
	if _, ok, err := c.Get(ctx, b.ID); err != nil || ok {
		t.Fatalf("před Set chci miss, ok=%v err=%v", ok, err)
	}
	if err := c.Set(ctx, b); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, ok, err := c.Get(ctx, b.ID)
	if err != nil || !ok || got.Title != "Cache" {
		t.Fatalf("Get = %+v ok=%v err=%v", got, ok, err)
	}
	if err := c.Delete(ctx, b.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := c.Ready(ctx); err != nil {
		t.Errorf("Ready: %v", err)
	}
}
