package bookmark_test

import (
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/bookmark"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{
			"HTTPS://Example.COM/Blog/?utm_source=x&b=2&a=1#frag",
			"https://example.com/Blog?a=1&b=2",
		},
		{
			"http://localhost:80/path/",
			"http://localhost/path",
		},
		{
			"https://go.dev:443/",
			"https://go.dev",
		},
	}
	for _, tt := range tests {
		got, err := bookmark.NormalizeURL(tt.in)
		if err != nil {
			t.Fatalf("NormalizeURL(%q) = chyba %v", tt.in, err)
		}
		if got != tt.want {
			t.Errorf("NormalizeURL(%q) = %q, chci %q", tt.in, got, tt.want)
		}
	}
}

func TestNormalizeURLChyby(t *testing.T) {
	for _, in := range []string{"", "ftp://x", "not a url", "http://"} {
		if _, err := bookmark.NormalizeURL(in); !errors.Is(err, bookmark.ErrInvalidURL) {
			t.Errorf("NormalizeURL(%q) = %v, chci ErrInvalidURL", in, err)
		}
	}
}

func TestNewAValidate(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	b, err := bookmark.New("bm_1", "https://Go.DEV/blog/?utm_campaign=x", "", []string{"Go", "go", "blog"}, now)
	if err != nil {
		t.Fatalf("New = chyba %v", err)
	}
	if b.URL != "https://go.dev/blog" {
		t.Errorf("URL = %q, chci normalizovanou", b.URL)
	}
	if b.Title != "go.dev" {
		t.Errorf("Title = %q, chci host jako výchozí titulek", b.Title)
	}
	if len(b.Tags) != 2 || b.Tags[0] != "blog" || b.Tags[1] != "go" {
		t.Errorf("Tags = %v, chci [blog go]", b.Tags)
	}
}

func TestNewNeplatnyTag(t *testing.T) {
	_, err := bookmark.New("id", "https://example.com", "t", []string{"bad tag"}, time.Now())
	if !errors.Is(err, bookmark.ErrInvalidTag) {
		t.Errorf("chci ErrInvalidTag, dostal jsem %v", err)
	}
}

func TestDomainNesmiImportovatHTTPAniJSON(t *testing.T) {
	forbidden := map[string]bool{
		"net/http":      true,
		"encoding/json": true,
	}

	dir := "."
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("ParseFile(%s): %v", path, err)
		}
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if forbidden[path] {
				t.Errorf("%s importuje zakázaný balíček %q", name, path)
			}
		}
	}
}
