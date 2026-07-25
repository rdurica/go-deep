package store_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p05-capstone/internal/bookmark"
	"github.com/rdurica/go-deep/projects/p05-capstone/internal/store"
)

func mustBookmark(t *testing.T, id, rawURL, title string, tags ...string) bookmark.Bookmark {
	t.Helper()
	b, err := bookmark.New(id, rawURL, title, tags, time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("bookmark.New = %v", err)
	}
	return b
}

func TestStoreCRUD(t *testing.T) {
	s := store.New()
	b := mustBookmark(t, "bm_1", "https://example.com/a", "Alpha", "go")
	if err := s.Add(b); err != nil {
		t.Fatalf("Add = %v", err)
	}
	got, err := s.Get("bm_1")
	if err != nil {
		t.Fatalf("Get = %v", err)
	}
	if got.Title != "Alpha" {
		t.Errorf("Title = %q, chci Alpha", got.Title)
	}
	if err := s.Add(mustBookmark(t, "bm_2", "https://example.com/a?utm_source=x", "Dup")); !errors.Is(err, bookmark.ErrDuplicateURL) {
		t.Errorf("duplicitní URL: chci ErrDuplicateURL, dostal jsem %v", err)
	}
	if err := s.Delete("bm_1"); err != nil {
		t.Fatalf("Delete = %v", err)
	}
	if _, err := s.Get("bm_1"); !errors.Is(err, bookmark.ErrNotFound) {
		t.Errorf("po Delete chci ErrNotFound, dostal jsem %v", err)
	}
}

func TestStoreSearch(t *testing.T) {
	s := store.New()
	items := []bookmark.Bookmark{
		mustBookmark(t, "a", "https://a.example/1", "Go Blog", "go"),
		mustBookmark(t, "b", "https://b.example/2", "Rust Notes", "rust"),
		mustBookmark(t, "c", "https://c.example/3", "Go Tips", "go", "tips"),
	}
	for i, b := range items {
		b.CreatedAt = b.CreatedAt.Add(time.Duration(i) * time.Second)
		if err := s.Add(b); err != nil {
			t.Fatalf("Add = %v", err)
		}
	}

	page, err := s.Search(store.Query{Tag: "go", Limit: 1})
	if err != nil {
		t.Fatalf("Search = %v", err)
	}
	if page.Total != 2 || len(page.Items) != 1 || page.NextCursor == "" {
		t.Fatalf("Search page = %+v, chci 2 celkem, 1 položku a cursor", page)
	}

	page2, err := s.Search(store.Query{Tag: "go", Limit: 1, Cursor: page.NextCursor})
	if err != nil {
		t.Fatalf("Search page2 = %v", err)
	}
	if len(page2.Items) != 1 || page2.Items[0].ID == page.Items[0].ID {
		t.Errorf("druhá stránka má být jiná záložka: %+v vs %+v", page.Items, page2.Items)
	}

	qpage, err := s.Search(store.Query{Text: "rust"})
	if err != nil || len(qpage.Items) != 1 || qpage.Items[0].ID != "b" {
		t.Errorf("text search = %+v, chci jen b", qpage)
	}
}

func TestStoreRace(t *testing.T) {
	s := store.New()
	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			id := fmt.Sprintf("bm_%d", i)
			b, err := bookmark.New(id, fmt.Sprintf("https://example.com/%d", i), "t", []string{"go"}, time.Now())
			if err != nil {
				errs <- err
				return
			}
			if err := s.Add(b); err != nil {
				errs <- err
				return
			}
			if _, err := s.Get(id); err != nil {
				errs <- err
				return
			}
			if _, err := s.Search(store.Query{Tag: "go", Limit: 10}); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("souběžná operace: %v", err)
	}
	if s.Len() != n {
		t.Errorf("Len = %d, chci %d", s.Len(), n)
	}
}
