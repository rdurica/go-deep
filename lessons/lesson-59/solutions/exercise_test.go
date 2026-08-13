package solutions_test

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-59/solutions"
)

var base = time.Date(2024, time.May, 1, 12, 0, 0, 0, time.UTC)

func TestNormalizeURL(t *testing.T) {
	tests := map[string]string{
		"HTTPS://Example.com/a/?utm_source=x":               "https://example.com/a",
		"http://example.com:80/":                            "http://example.com",
		"https://example.com:443/x":                         "https://example.com/x",
		"https://example.com:8443/x":                        "https://example.com:8443/x",
		"https://example.com/a?b=2&a=1":                     "https://example.com/a?a=1&b=2",
		"https://example.com/a#section":                     "https://example.com/a",
		"  https://example.com  ":                           "https://example.com",
		"https://example.com/?utm_source=x&utm_medium=cpc":  "https://example.com",
		"https://example.com/a?utm_campaign=jaro&q=go":      "https://example.com/a?q=go",
		"HTTP://EXAMPLE.COM":                                "http://example.com",
		"https://example.com/blog/2024/":                    "https://example.com/blog/2024",
		"https://example.com/a?UTM_SOURCE=x&keep=1":         "https://example.com/a?keep=1",
		"https://go.dev/doc/effective_go#interfaces":        "https://go.dev/doc/effective_go",
		"https://example.com/hledat?q=go+routine&utm_id=99": "https://example.com/hledat?q=go+routine",
	}
	for in, want := range tests {
		got, err := exercise.NormalizeURL(in)
		if err != nil || got != want {
			t.Errorf("NormalizeURL(%q) = (%q, %v), chci (%q, nil)", in, got, err, want)
		}
	}
}

func TestNormalizeURLErrors(t *testing.T) {
	for _, in := range []string{"", "   ", "ftp://example.com", "example.com/a", "https://", "mailto:a@b.cz"} {
		got, err := exercise.NormalizeURL(in)
		if !errors.Is(err, exercise.ErrInvalidURL) {
			t.Errorf("NormalizeURL(%q) = (%q, %v), chci ErrInvalidURL", in, got, err)
		}
	}
}

func TestNormalizeURLIsIdempotent(t *testing.T) {
	for _, in := range []string{
		"HTTPS://Example.com/a/?utm_source=x",
		"https://example.com/a?b=2&a=1",
		"http://example.com:80/",
	} {
		once, err := exercise.NormalizeURL(in)
		if err != nil {
			t.Fatalf("NormalizeURL(%q) = chyba %v", in, err)
		}
		twice, err := exercise.NormalizeURL(once)
		if err != nil || twice != once {
			t.Errorf("NormalizeURL(%q) = (%q, %v), chci stabilní %q", once, twice, err, once)
		}
	}
}

func TestNew(t *testing.T) {
	got, err := exercise.New(" b1 ", "HTTPS://Go.dev/doc/?utm_source=x", "  Dokumentace  ", []string{"Docs", "go", "go"}, base)
	if err != nil {
		t.Fatalf("New() = chyba %v", err)
	}
	want := exercise.Bookmark{
		ID:        "b1",
		URL:       "https://go.dev/doc",
		Title:     "Dokumentace",
		Tags:      []string{"docs", "go"},
		CreatedAt: base,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("New() = %+v, chci %+v", got, want)
	}

	if _, err := exercise.New("b2", "nope", "T", nil, base); !errors.Is(err, exercise.ErrInvalidURL) {
		t.Errorf("New(neplatná URL) = %v, chci ErrInvalidURL", err)
	}
	if _, err := exercise.New("", "https://go.dev", "T", nil, base); !errors.Is(err, exercise.ErrEmptyID) {
		t.Errorf("New(bez ID) = %v, chci ErrEmptyID", err)
	}
}

func mustNew(t *testing.T, id, rawURL, title string, tags []string, offset time.Duration) exercise.Bookmark {
	t.Helper()
	b, err := exercise.New(id, rawURL, title, tags, base.Add(offset))
	if err != nil {
		t.Fatalf("New(%q) = chyba %v", id, err)
	}
	return b
}

func ids(items []exercise.Bookmark) []string {
	out := make([]string, 0, len(items))
	for _, b := range items {
		out = append(out, b.ID)
	}
	return out
}

func searchStore(t *testing.T) *exercise.Store {
	t.Helper()
	s := exercise.NewStore()
	items := []exercise.Bookmark{
		mustNew(t, "b1", "https://a.example/1", "Go slices a aliasing", []string{"go", "memory"}, 0),
		mustNew(t, "b2", "https://a.example/2", "HTTP middleware v Go", []string{"go", "http"}, time.Hour),
		mustNew(t, "b3", "https://a.example/3", "Aliasing v PHP", []string{"php"}, 2*time.Hour),
		mustNew(t, "b4", "https://a.example/4", "context v request scope", []string{"go", "http", "context"}, 3*time.Hour),
		mustNew(t, "b5", "https://a.example/5", "Balíčky podle domény", []string{"go", "design"}, 4*time.Hour),
	}
	for _, b := range items {
		if err := s.Add(b); err != nil {
			t.Fatalf("Add(%s) = %v", b.ID, err)
		}
	}
	return s
}

func TestSearchFilters(t *testing.T) {
	s := searchStore(t)

	tests := []struct {
		name string
		q    exercise.Query
		want []string
	}{
		{"no filter, newest first", exercise.Query{}, []string{"b5", "b4", "b3", "b2", "b1"}},
		{"one tag", exercise.Query{Tags: []string{"http"}}, []string{"b4", "b2"}},
		{"tags OR", exercise.Query{Tags: []string{"php", "design"}}, []string{"b5", "b3"}},
		{"tags AND", exercise.Query{Tags: []string{"go", "http"}, MatchAll: true}, []string{"b4", "b2"}},
		{"AND bez shody", exercise.Query{Tags: []string{"php", "go"}, MatchAll: true}, nil},
		{"fulltext v titulku", exercise.Query{Text: "aliasing"}, []string{"b3", "b1"}},
		{"fulltext ignoruje velikost", exercise.Query{Text: "GO"}, []string{"b2", "b1"}},
		{"tag and fulltext", exercise.Query{Tags: []string{"go"}, Text: "aliasing"}, []string{"b1"}},
		{"nic nenajde", exercise.Query{Text: "symfony"}, nil},
		{"normalized tag", exercise.Query{Tags: []string{" HTTP "}}, []string{"b4", "b2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, err := s.Search(tt.q)
			if err != nil {
				t.Fatalf("Search() = chyba %v", err)
			}
			if got := ids(page.Items); !reflect.DeepEqual(got, tt.want) && !(len(got) == 0 && len(tt.want) == 0) {
				t.Errorf("Search() = %v, chci %v", got, tt.want)
			}
			if page.Total != len(tt.want) {
				t.Errorf("Total = %d, chci %d", page.Total, len(tt.want))
			}
			if page.NextCursor != "" {
				t.Errorf("NextCursor = %q, chci prázdný (vejde se na jednu stránku)", page.NextCursor)
			}
		})
	}
}

func TestSearchSortByTitle(t *testing.T) {
	s := searchStore(t)
	page, err := s.Search(exercise.Query{Sort: exercise.SortTitle})
	if err != nil {
		t.Fatalf("Search() = %v", err)
	}
	want := []string{"b3", "b5", "b4", "b1", "b2"}
	if got := ids(page.Items); !reflect.DeepEqual(got, want) {
		t.Errorf("Search(SortTitle) = %v, chci %v", got, want)
	}
}

func TestSearchPagination(t *testing.T) {
	s := searchStore(t)

	var seen []string
	cursor := ""
	for page := 0; page < 5; page++ {
		got, err := s.Search(exercise.Query{Limit: 2, Cursor: cursor})
		if err != nil {
			t.Fatalf("Search(strana %d) = %v", page, err)
		}
		if got.Total != 5 {
			t.Errorf("Total na straně %d = %d, chci 5", page, got.Total)
		}
		seen = append(seen, ids(got.Items)...)
		if got.NextCursor == "" {
			break
		}
		if len(got.Items) == 0 {
			t.Fatalf("prázdná strana s NextCursor %q", got.NextCursor)
		}
		if got.NextCursor != got.Items[len(got.Items)-1].ID {
			t.Errorf("NextCursor = %q, chci ID poslední položky %q", got.NextCursor, got.Items[len(got.Items)-1].ID)
		}
		cursor = got.NextCursor
	}

	want := []string{"b5", "b4", "b3", "b2", "b1"}
	if !reflect.DeepEqual(seen, want) {
		t.Errorf("stránkování prošlo %v, chci %v bez duplicit a bez děr", seen, want)
	}
}

func TestSearchStableWithEqualTimestamps(t *testing.T) {
	s := exercise.NewStore()
	for _, id := range []string{"c", "a", "b"} {
		b := mustNew(t, id, "https://a.example/"+id, "Stejný čas "+id, []string{"go"}, 0)
		if err := s.Add(b); err != nil {
			t.Fatalf("Add(%s) = %v", id, err)
		}
	}
	for i := 0; i < 5; i++ {
		page, err := s.Search(exercise.Query{})
		if err != nil {
			t.Fatalf("Search() = %v", err)
		}
		if got, want := ids(page.Items), []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("Search() = %v, chci %v (stabilní řazení podle ID při shodě času)", got, want)
		}
	}
}

func TestSearchValidation(t *testing.T) {
	s := searchStore(t)

	tests := []struct {
		name string
		q    exercise.Query
		want error
	}{
		{"negative limit", exercise.Query{Limit: -1}, exercise.ErrInvalidQuery},
		{"limit nad strop", exercise.Query{Limit: exercise.MaxLimit + 1}, exercise.ErrInvalidQuery},
		{"unknown sort", exercise.Query{Sort: exercise.SortOrder(9)}, exercise.ErrInvalidQuery},
		{"invalid tag", exercise.Query{Tags: []string{"go lang"}}, exercise.ErrInvalidQuery},
		{"unknown cursor", exercise.Query{Cursor: "neexistuje"}, exercise.ErrInvalidCursor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := s.Search(tt.q); !errors.Is(err, tt.want) {
				t.Errorf("Search() = %v, chci %v", err, tt.want)
			}
		})
	}

	t.Run("limit 0 means default", func(t *testing.T) {
		page, err := s.Search(exercise.Query{Limit: 0})
		if err != nil {
			t.Fatalf("Search() = %v", err)
		}
		if len(page.Items) != 5 {
			t.Errorf("Search(limit 0) vrátil %d položek, chci všech 5", len(page.Items))
		}
	})
}

func TestSearchEmptyStore(t *testing.T) {
	s := exercise.NewStore()
	page, err := s.Search(exercise.Query{Tags: []string{"go"}})
	if err != nil {
		t.Fatalf("Search() = %v", err)
	}
	if len(page.Items) != 0 || page.Total != 0 || page.NextCursor != "" {
		t.Errorf("Search(prázdný store) = %+v, chci prázdnou stránku", page)
	}
}

func TestStoreConcurrent(t *testing.T) {
	s := exercise.NewStore()
	const n = 64

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			b := exercise.Bookmark{
				ID:        fmt.Sprintf("b%03d", i),
				URL:       fmt.Sprintf("https://example.com/%d", i),
				Title:     fmt.Sprintf("Záložka %d", i),
				Tags:      []string{"go"},
				CreatedAt: base.Add(time.Duration(i) * time.Minute),
			}
			if err := s.Add(b); err != nil {
				t.Errorf("Add(%s) = %v", b.ID, err)
			}
		}(i)
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.Search(exercise.Query{Tags: []string{"go"}, Limit: 5}); err != nil {
				t.Errorf("Search() = %v", err)
			}
		}()
	}
	wg.Wait()

	if got := s.Len(); got != n {
		t.Errorf("Len() = %d, chci %d", got, n)
	}
}
