package exercise_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-25/exercise"
)

// decodeJSON dekóduje tělo odpovědi do v a selže, pokud to není platný JSON.
func decodeJSON(t *testing.T, body []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("tělo není platný JSON: %v (tělo = %q)", err, string(body))
	}
}

// do pošle požadavek routeru a vrátí zaznamenanou odpověď.
func do(t *testing.T, h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, target, nil)
	} else {
		req = httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// seededStore vrátí úložiště se třemi položkami s ID "1", "2" a "3".
func seededStore(t *testing.T) *exercise.Store {
	t.Helper()
	store := exercise.NewStore()
	store.Add("Apple")
	store.Add("Banana")
	store.Add("Cherry")
	return store
}

func TestItemsRouterGetItem(t *testing.T) {
	router := exercise.ItemsRouter(seededStore(t))

	rec := do(t, router, http.MethodGet, "/items/2", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /items/2 = %d, chci %d", rec.Code, http.StatusOK)
	}
	var got exercise.Item
	decodeJSON(t, rec.Body.Bytes(), &got)
	if got.ID != "2" || got.Name != "Banana" {
		t.Errorf("GET /items/2 = %+v, chci {ID:2 Name:Banana}", got)
	}
}

func TestItemsRouterGetItemNotFound(t *testing.T) {
	router := exercise.ItemsRouter(seededStore(t))

	rec := do(t, router, http.MethodGet, "/items/999", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /items/999 = %d, chci %d", rec.Code, http.StatusNotFound)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, chci application/json", ct)
	}
	var got exercise.ErrorResponse
	decodeJSON(t, rec.Body.Bytes(), &got)
	if got.Error == "" {
		t.Error("404 odpověď má prázdné pole error")
	}
}

func TestItemsRouterRootIsExactMatch(t *testing.T) {
	router := exercise.ItemsRouter(seededStore(t))

	rec := do(t, router, http.MethodGet, "/", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / = %d, chci %d", rec.Code, http.StatusOK)
	}
	var info exercise.ServiceInfo
	decodeJSON(t, rec.Body.Bytes(), &info)
	if info.Service != "items" {
		t.Errorf("GET / service = %q, chci %q", info.Service, "items")
	}

	// vzor "GET /{$}" nesmí chytat nic jiného než přesný kořen
	if rec := do(t, router, http.MethodGet, "/tohle-neexistuje", ""); rec.Code != http.StatusNotFound {
		t.Errorf("GET /tohle-neexistuje = %d, chci %d", rec.Code, http.StatusNotFound)
	}
}

func TestItemsRouterWrongMethodIs405(t *testing.T) {
	router := exercise.ItemsRouter(seededStore(t))

	tests := []struct {
		method      string
		target      string
		wantInAllow []string
	}{
		{http.MethodPut, "/items/1", []string{"GET", "DELETE"}},
		{http.MethodPatch, "/items", []string{"GET", "POST"}},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.target, func(t *testing.T) {
			rec := do(t, router, tt.method, tt.target, "")
			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("%s %s = %d, chci %d (405 dělá ServeMux sám od Go 1.22)",
					tt.method, tt.target, rec.Code, http.StatusMethodNotAllowed)
			}
			allow := rec.Header().Get("Allow")
			for _, method := range tt.wantInAllow {
				if !strings.Contains(allow, method) {
					t.Errorf("Allow = %q, chci aby obsahovala %q", allow, method)
				}
			}
		})
	}
}

func TestParseListQuery(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantLimit int
		wantQ     string
		wantErr   bool
	}{
		{"empty query", "", 0, "", false},
		{"jen q", "q=ban", 0, "ban", false},
		{"limit a q", "limit=5&q=ban", 5, "ban", false},
		{"limit 1", "limit=1", 1, "", false},
		{"empty limit", "limit=", 0, "", false},
		{"non-numeric limit", "limit=abc", 0, "", true},
		{"negative limit", "limit=-3", 0, "", true},
		{"zero limit", "limit=0", 0, "", true},
		{"decimal limit", "limit=1.5", 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := url.ParseQuery(tt.raw)
			if err != nil {
				t.Fatalf("nepovedlo se připravit dotaz: %v", err)
			}

			limit, q, err := exercise.ParseListQuery(values)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseListQuery(%q) = (%d, %q, nil), chci chybu", tt.raw, limit, q)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseListQuery(%q) vrátil chybu %v, chci nil", tt.raw, err)
			}
			if limit != tt.wantLimit || q != tt.wantQ {
				t.Errorf("ParseListQuery(%q) = (%d, %q), chci (%d, %q)", tt.raw, limit, q, tt.wantLimit, tt.wantQ)
			}
		})
	}
}

func TestItemsRouterList(t *testing.T) {
	router := exercise.ItemsRouter(seededStore(t))

	tests := []struct {
		name      string
		target    string
		wantNames []string
	}{
		{"no params", "/items", []string{"Apple", "Banana", "Cherry"}},
		{"limit 2", "/items?limit=2", []string{"Apple", "Banana"}},
		{"filtr q", "/items?q=an", []string{"Banana"}},
		{"filtr q je case-insensitive", "/items?q=AN", []string{"Banana"}},
		{"filtr bez shody", "/items?q=zzz", nil},
		{"limit i q", "/items?q=a&limit=1", []string{"Apple"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, router, http.MethodGet, tt.target, "")
			if rec.Code != http.StatusOK {
				t.Fatalf("GET %s = %d, chci %d", tt.target, rec.Code, http.StatusOK)
			}

			var items []exercise.Item
			decodeJSON(t, rec.Body.Bytes(), &items)
			if len(items) != len(tt.wantNames) {
				t.Fatalf("GET %s vrátilo %d položek, chci %d (%v)", tt.target, len(items), len(tt.wantNames), items)
			}
			for i, want := range tt.wantNames {
				if items[i].Name != want {
					t.Errorf("položka %d = %q, chci %q", i, items[i].Name, want)
				}
			}
		})
	}
}

func TestItemsRouterListInvalidLimit(t *testing.T) {
	router := exercise.ItemsRouter(seededStore(t))

	for _, target := range []string{"/items?limit=abc", "/items?limit=-1", "/items?limit=0"} {
		rec := do(t, router, http.MethodGet, target, "")
		if rec.Code != http.StatusBadRequest {
			t.Errorf("GET %s = %d, chci %d", target, rec.Code, http.StatusBadRequest)
		}
		var got exercise.ErrorResponse
		decodeJSON(t, rec.Body.Bytes(), &got)
		if got.Error == "" {
			t.Errorf("GET %s: chybová odpověď má prázdné pole error", target)
		}
	}
}

func TestItemsRouterCreate(t *testing.T) {
	store := exercise.NewStore()
	router := exercise.ItemsRouter(store)

	rec := do(t, router, http.MethodPost, "/items", `{"name":"Durian"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /items = %d, chci %d", rec.Code, http.StatusCreated)
	}

	var created exercise.Item
	decodeJSON(t, rec.Body.Bytes(), &created)
	if created.Name != "Durian" || created.ID == "" {
		t.Fatalf("POST /items vrátilo %+v, chci neprázdné ID a jméno Durian", created)
	}

	wantLocation := "/items/" + created.ID
	if got := rec.Header().Get("Location"); got != wantLocation {
		t.Errorf("Location = %q, chci %q", got, wantLocation)
	}

	// vytvořená položka musí být dosažitelná na adrese z hlavičky Location
	rec2 := do(t, router, http.MethodGet, wantLocation, "")
	if rec2.Code != http.StatusOK {
		t.Errorf("GET %s = %d, chci %d", wantLocation, rec2.Code, http.StatusOK)
	}
}

func TestItemsRouterCreateErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"broken JSON", `{"name":`},
		{"empty name", `{"name":""}`},
		{"name only spaces", `{"name":"   "}`},
		{"missing name", `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := exercise.ItemsRouter(exercise.NewStore())
			rec := do(t, router, http.MethodPost, "/items", tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("POST /items %s = %d, chci %d", tt.body, rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestItemsRouterDelete(t *testing.T) {
	router := exercise.ItemsRouter(seededStore(t))

	rec := do(t, router, http.MethodDelete, "/items/1", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE /items/1 = %d, chci %d", rec.Code, http.StatusNoContent)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("DELETE /items/1 vrátil tělo %q, chci prázdné", rec.Body.String())
	}

	if rec := do(t, router, http.MethodGet, "/items/1", ""); rec.Code != http.StatusNotFound {
		t.Errorf("GET /items/1 po smazání = %d, chci %d", rec.Code, http.StatusNotFound)
	}
	if rec := do(t, router, http.MethodDelete, "/items/1", ""); rec.Code != http.StatusNotFound {
		t.Errorf("druhý DELETE /items/1 = %d, chci %d", rec.Code, http.StatusNotFound)
	}
}

// tempTree připraví adresář public/ se souborem sub/hello.txt a mimo něj secret.txt.
func tempTree(t *testing.T) (root, secret string) {
	t.Helper()
	base := t.TempDir()
	root = filepath.Join(base, "public")
	if err := os.MkdirAll(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatalf("nepovedlo se vytvořit adresář: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "sub", "hello.txt"), []byte("ahoj ze souboru"), 0o644); err != nil {
		t.Fatalf("nepovedlo se zapsat soubor: %v", err)
	}
	secret = filepath.Join(base, "secret.txt")
	if err := os.WriteFile(secret, []byte("TAJEMSTVI"), 0o644); err != nil {
		t.Fatalf("nepovedlo se zapsat soubor: %v", err)
	}
	return root, secret
}

func TestSafeJoinAllows(t *testing.T) {
	root, _ := tempTree(t)

	got, err := exercise.SafeJoin(root, "sub/hello.txt")
	if err != nil {
		t.Fatalf("SafeJoin(root, %q) vrátil chybu %v, chci nil", "sub/hello.txt", err)
	}
	if want := filepath.Join(root, "sub", "hello.txt"); got != want {
		t.Errorf("SafeJoin(root, %q) = %q, chci %q", "sub/hello.txt", got, want)
	}
}

func TestSafeJoinRejects(t *testing.T) {
	root, _ := tempTree(t)

	bad := []string{
		"",
		"..",
		".",
		"../secret.txt",
		"sub/../../secret.txt",
		"sub/./../../secret.txt",
		"/etc/passwd",
		"/",
		"sub//hello.txt",
	}

	for _, rel := range bad {
		t.Run(rel, func(t *testing.T) {
			got, err := exercise.SafeJoin(root, rel)
			if err == nil {
				t.Fatalf("SafeJoin(root, %q) = (%q, nil), chci chybu", rel, got)
			}
		})
	}
}

func TestFilesHandlerServesFile(t *testing.T) {
	root, _ := tempTree(t)

	req := httptest.NewRequest(http.MethodGet, "/files/sub/hello.txt", nil)
	req.SetPathValue("path", "sub/hello.txt")
	rec := httptest.NewRecorder()

	exercise.FilesHandler(root).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, chci %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); got != "ahoj ze souboru" {
		t.Errorf("tělo = %q, chci %q", got, "ahoj ze souboru")
	}
}

func TestFilesHandlerMissingFile(t *testing.T) {
	root, _ := tempTree(t)

	req := httptest.NewRequest(http.MethodGet, "/files/chybi.txt", nil)
	req.SetPathValue("path", "chybi.txt")
	rec := httptest.NewRecorder()

	exercise.FilesHandler(root).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, chci %d", rec.Code, http.StatusNotFound)
	}
}

func TestFilesHandlerRejectsTraversal(t *testing.T) {
	root, _ := tempTree(t)
	handler := exercise.FilesHandler(root)

	bad := []string{
		"../secret.txt",
		"sub/../../secret.txt",
		"/etc/passwd",
		"",
		"..",
	}

	for _, rel := range bad {
		t.Run(rel, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/files/x", nil)
			req.SetPathValue("path", rel)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest && rec.Code != http.StatusNotFound {
				t.Fatalf("path=%q dalo status %d, chci 400 nebo 404", rel, rec.Code)
			}
			if strings.Contains(rec.Body.String(), "TAJEMSTVI") {
				t.Fatalf("path=%q vyzradilo obsah souboru mimo kořen: %q", rel, rec.Body.String())
			}
		})
	}
}

func TestFilesRouterPresServer(t *testing.T) {
	root, _ := tempTree(t)

	srv := httptest.NewServer(exercise.FilesRouter(root))
	defer srv.Close()

	res, err := srv.Client().Get(srv.URL + "/files/sub/hello.txt")
	if err != nil {
		t.Fatalf("GET selhal: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, chci %d", res.StatusCode, http.StatusOK)
	}
	body, _ := io.ReadAll(res.Body)
	if string(body) != "ahoj ze souboru" {
		t.Errorf("tělo = %q, chci %q", string(body), "ahoj ze souboru")
	}

	res2, err := srv.Client().Get(srv.URL + "/files/nic.txt")
	if err != nil {
		t.Fatalf("GET selhal: %v", err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, chci %d", res2.StatusCode, http.StatusNotFound)
	}
}
