package httpapi_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/httpapi"
	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/memstore"
)

func newSeqServer(t *testing.T) http.Handler {
	t.Helper()
	var n int
	return httpapi.NewServerWith(memstore.New(), httpapi.Options{
		MaxBodyBytes: 1 << 20,
		Timeout:      2 * time.Second,
	}, func() string {
		n++
		return fmt.Sprintf("bm_%d", n)
	}, func() time.Time {
		return time.Date(2026, 7, 25, 12, n, 0, 0, time.UTC)
	})
}

func TestHealthAndReady(t *testing.T) {
	h := newSeqServer(t)
	for _, path := range []string{"/healthz", "/readyz"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Errorf("%s = %d, chci 200", path, rec.Code)
		}
	}
}

func TestCreateGetDeleteHappyPath(t *testing.T) {
	h := newSeqServer(t)

	rec := httptest.NewRecorder()
	body := `{"url":"https://Example.COM/a/?utm_source=x","title":"Alpha","tags":["Go","blog"]}`
	req := httptest.NewRequest(http.MethodPost, "/bookmarks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST = %d, tělo %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get(httpapi.RequestIDHeader); got == "" {
		t.Error("chybí X-Request-ID")
	}

	var created map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created["url"] != "https://example.com/a" {
		t.Errorf("url = %v, chci normalizovanou", created["url"])
	}
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatal("chybí id")
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/bookmarks/"+id, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/bookmarks?tag=go", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("SEARCH = %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/bookmarks/"+id, nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE = %d", rec.Code)
	}
}

func TestCreateBadJSON(t *testing.T) {
	h := newSeqServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/bookmarks", strings.NewReader(`{`))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("špatný JSON = %d, chci 400", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "problem+json") {
		t.Errorf("Content-Type = %q, chci problem+json", ct)
	}
}

func TestCreateInvalidURL(t *testing.T) {
	h := newSeqServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/bookmarks", strings.NewReader(`{"url":"ftp://x"}`))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("neplatná URL = %d, chci 400", rec.Code)
	}
}

func TestGetNotFound(t *testing.T) {
	h := newSeqServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/bookmarks/missing", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET missing = %d, chci 404", rec.Code)
	}
}

func TestCreateConflict(t *testing.T) {
	h := newSeqServer(t)
	body := `{"url":"https://example.com/dup","title":"One"}`
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/bookmarks", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		h.ServeHTTP(rec, req)
		if i == 0 && rec.Code != http.StatusCreated {
			t.Fatalf("první POST = %d", rec.Code)
		}
		if i == 1 && rec.Code != http.StatusConflict {
			t.Fatalf("duplicitní POST = %d, chci 409; tělo %s", rec.Code, rec.Body.String())
		}
	}
}

func TestRecovery(t *testing.T) {
	inner := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})
	h := httpapi.Wrap(inner, httpapi.Options{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("recovery = %d, chci 500", rec.Code)
	}
	_, _ = io.Copy(io.Discard, rec.Body)
}
