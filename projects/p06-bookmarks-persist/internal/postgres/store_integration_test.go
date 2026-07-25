package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/app"
	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/bookmark"
	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/postgres"
)

func openTestStore(t *testing.T) *postgres.Store {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL není nastavená — přeskočeno (spusť docker compose)")
	}
	ctx := context.Background()
	st, err := postgres.Open(ctx, url)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(st.Close)
	if err := st.ApplyMigrations(ctx); err != nil {
		t.Fatalf("ApplyMigrations: %v", err)
	}
	// izolace: vyčisti tabulku mezi testy
	if err := st.Migrate(ctx, `TRUNCATE bookmarks`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return st
}

func TestPostgresCRUD(t *testing.T) {
	ctx := context.Background()
	st := openTestStore(t)

	b, err := bookmark.New("bm_1", "https://example.com/a", "Alpha", []string{"go"}, time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := st.Add(ctx, b); err != nil {
		t.Fatalf("Add: %v", err)
	}
	got, err := st.Get(ctx, "bm_1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Title != "Alpha" || got.URL != "https://example.com/a" {
		t.Errorf("got = %+v", got)
	}

	dup, err := bookmark.New("bm_2", "https://example.com/a?utm_source=x", "Dup", nil, time.Now().UTC())
	if err != nil {
		t.Fatalf("New dup: %v", err)
	}
	if err := st.Add(ctx, dup); !errors.Is(err, bookmark.ErrDuplicateURL) {
		t.Errorf("duplicitní URL: chci ErrDuplicateURL, dostal jsem %v", err)
	}

	page, err := st.Search(ctx, app.Query{Tag: "go", Limit: 10})
	if err != nil || page.Total != 1 {
		t.Fatalf("Search = %+v, err=%v", page, err)
	}

	if err := st.Delete(ctx, "bm_1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := st.Get(ctx, "bm_1"); !errors.Is(err, bookmark.ErrNotFound) {
		t.Errorf("po Delete chci ErrNotFound, dostal jsem %v", err)
	}
	if err := st.Ready(ctx); err != nil {
		t.Errorf("Ready: %v", err)
	}
}
