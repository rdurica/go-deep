package solutions_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-31/solutions"
)

// safeBuffer je buffer chráněný mutexem — logy vznikají v goroutině serveru.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func fakeEnv(m map[string]string) func(string) string {
	return func(key string) string { return m[key] }
}

// decodeJSON rozparsuje tělo odpovědi do mapy.
func decodeJSON(t *testing.T, r io.Reader) map[string]any {
	t.Helper()

	var m map[string]any
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		t.Fatalf("odpověď není platný JSON: %v", err)
	}
	return m
}

// wantErrorShape ověří konzistentní tvar chybové odpovědi.
func wantErrorShape(t *testing.T, body map[string]any, wantCode string) {
	t.Helper()

	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("chybová odpověď nemá objekt error; tělo: %v", body)
	}
	code, _ := errObj["code"].(string)
	if code != wantCode {
		t.Errorf("error.code = %q, chci %q", code, wantCode)
	}
	if msg, _ := errObj["message"].(string); msg == "" {
		t.Error("error.message je prázdná")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := exercise.LoadConfig(fakeEnv(nil))
	if err != nil {
		t.Fatalf("LoadConfig = %v, chci nil", err)
	}
	want := exercise.Config{
		Addr:            "127.0.0.1:8080",
		LogLevel:        slog.LevelInfo,
		ShutdownTimeout: 5 * time.Second,
	}
	if cfg != want {
		t.Errorf("LoadConfig = %+v, chci %+v", cfg, want)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	t.Parallel()

	cfg, err := exercise.LoadConfig(fakeEnv(map[string]string{
		"ADDR":             "0.0.0.0:9000",
		"LOG_LEVEL":        "debug",
		"SHUTDOWN_TIMEOUT": "250ms",
	}))
	if err != nil {
		t.Fatalf("LoadConfig = %v, chci nil", err)
	}
	if cfg.Addr != "0.0.0.0:9000" {
		t.Errorf("Addr = %q, chci %q", cfg.Addr, "0.0.0.0:9000")
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, chci %v", cfg.LogLevel, slog.LevelDebug)
	}
	if cfg.ShutdownTimeout != 250*time.Millisecond {
		t.Errorf("ShutdownTimeout = %v, chci %v", cfg.ShutdownTimeout, 250*time.Millisecond)
	}
}

func TestLoadConfigCollectsErrors(t *testing.T) {
	t.Parallel()

	_, err := exercise.LoadConfig(fakeEnv(map[string]string{
		"LOG_LEVEL":        "upovídaně",
		"SHUTDOWN_TIMEOUT": "hned",
	}))
	if err == nil {
		t.Fatal("LoadConfig s rozbitým prostředím vrátil nil, chci chybu")
	}
	if !errors.Is(err, exercise.ErrInvalid) {
		t.Errorf("chyba %v se nedá porovnat přes errors.Is s ErrInvalid", err)
	}
	for _, key := range []string{"LOG_LEVEL", "SHUTDOWN_TIMEOUT"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("chyba neobsahuje %q; celý text:\n%s", key, err.Error())
		}
	}
}

func TestChainOrder(t *testing.T) {
	t.Parallel()

	tag := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, name+">")
				next.ServeHTTP(w, r)
			})
		}
	}
	final := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "handler")
	})

	h := exercise.Chain(final, tag("a"), tag("b"), tag("c"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got, want := rec.Body.String(), "a>b>c>handler"; got != want {
		t.Errorf("pořadí middleware = %q, chci %q (první uvedená je nejvíc vně)", got, want)
	}
}

func TestChainWithoutMiddleware(t *testing.T) {
	t.Parallel()

	final := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "handler")
	})
	rec := httptest.NewRecorder()
	exercise.Chain(final).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Body.String(); got != "handler" {
		t.Errorf("Chain bez middleware = %q, chci %q", got, "handler")
	}
}

func TestRequestIDMiddlewareGenerates(t *testing.T) {
	t.Parallel()

	var seen []string
	h := exercise.RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := exercise.RequestIDFromContext(r.Context())
		if !ok {
			t.Error("v kontextu handleru chybí request ID")
		}
		seen = append(seen, id)
	}))

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		header := rec.Header().Get(exercise.RequestIDHeader)
		if header == "" {
			t.Fatalf("odpověď nemá hlavičku %s", exercise.RequestIDHeader)
		}
		if header != seen[i] {
			t.Errorf("hlavička %q se liší od hodnoty v kontextu %q", header, seen[i])
		}
	}
	if seen[0] == seen[1] {
		t.Errorf("dva požadavky dostaly stejné ID %q", seen[0])
	}
}

func TestRequestIDMiddlewareReusesIncoming(t *testing.T) {
	t.Parallel()

	const incoming = "trace-abc-123"
	var got string
	h := exercise.RequestIDMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got, _ = exercise.RequestIDFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(exercise.RequestIDHeader, incoming)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got != incoming {
		t.Errorf("request ID = %q, chci převzít příchozí %q", got, incoming)
	}
	if h := rec.Header().Get(exercise.RequestIDHeader); h != incoming {
		t.Errorf("hlavička odpovědi = %q, chci %q", h, incoming)
	}
}

func TestRequestIDFromContextEmpty(t *testing.T) {
	t.Parallel()

	if _, ok := exercise.RequestIDFromContext(context.Background()); ok {
		t.Error("prázdný kontext nesmí vracet ok=true")
	}
}

// newTestServer spustí kompletní API na náhodném portu localhostu.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(exercise.NewServer(discardLogger()))
	t.Cleanup(srv.Close)
	return srv
}

// postNote vytvoří poznámku a vrátí její ID.
func postNote(t *testing.T, srv *httptest.Server, text string) string {
	t.Helper()

	body := strings.NewReader(`{"text":` + strconv.Quote(text) + `}`)
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/notes", body)
	if err != nil {
		t.Fatalf("sestavení požadavku selhalo: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("POST /notes selhal: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /notes status = %d, chci %d", resp.StatusCode, http.StatusCreated)
	}
	created := decodeJSON(t, resp.Body)
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatalf("vytvořená poznámka nemá id; tělo: %v", created)
	}
	if got, _ := created["text"].(string); got != text {
		t.Errorf("text = %q, chci %q", got, text)
	}
	if loc := resp.Header.Get("Location"); loc != "/notes/"+id {
		t.Errorf("Location = %q, chci %q", loc, "/notes/"+id)
	}
	return id
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)
	resp, err := srv.Client().Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz selhal: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, chci %d", resp.StatusCode, http.StatusOK)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, chci application/json", ct)
	}
	if body := decodeJSON(t, resp.Body); body["status"] != "ok" {
		t.Errorf("tělo = %v, chci status=ok", body)
	}
}

func TestNotesCRUD(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)
	id := postNote(t, srv, "koupit mléko")

	// GET jedné poznámky
	resp, err := srv.Client().Get(srv.URL + "/notes/" + id)
	if err != nil {
		t.Fatalf("GET /notes/%s selhal: %v", id, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /notes/%s status = %d, chci %d", id, resp.StatusCode, http.StatusOK)
	}
	got := decodeJSON(t, resp.Body)
	_ = resp.Body.Close()
	if got["text"] != "koupit mléko" {
		t.Errorf("text = %v, chci %q", got["text"], "koupit mléko")
	}

	// Výpis obsahuje vytvořenou poznámku
	postNote(t, srv, "zavolat mámě")
	listResp, err := srv.Client().Get(srv.URL + "/notes")
	if err != nil {
		t.Fatalf("GET /notes selhal: %v", err)
	}
	list := decodeJSON(t, listResp.Body)
	_ = listResp.Body.Close()
	notes, ok := list["notes"].([]any)
	if !ok {
		t.Fatalf("odpověď nemá pole notes; tělo: %v", list)
	}
	if len(notes) != 2 {
		t.Errorf("počet poznámek = %d, chci 2", len(notes))
	}

	// DELETE
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/notes/"+id, nil)
	delResp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("DELETE selhal: %v", err)
	}
	_ = delResp.Body.Close()
	if delResp.StatusCode != http.StatusNoContent {
		t.Errorf("DELETE status = %d, chci %d", delResp.StatusCode, http.StatusNoContent)
	}

	// Druhé smazání už je 404
	req2, _ := http.NewRequest(http.MethodDelete, srv.URL+"/notes/"+id, nil)
	delResp2, err := srv.Client().Do(req2)
	if err != nil {
		t.Fatalf("druhý DELETE selhal: %v", err)
	}
	defer func() { _ = delResp2.Body.Close() }()
	if delResp2.StatusCode != http.StatusNotFound {
		t.Errorf("druhý DELETE status = %d, chci %d", delResp2.StatusCode, http.StatusNotFound)
	}
	wantErrorShape(t, decodeJSON(t, delResp2.Body), "not_found")
}

func TestNotesEmptyList(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)
	resp, err := srv.Client().Get(srv.URL + "/notes")
	if err != nil {
		t.Fatalf("GET /notes selhal: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body := decodeJSON(t, resp.Body)
	notes, ok := body["notes"].([]any)
	if !ok {
		t.Fatalf("notes není pole (pozor na nil slice v JSONu); tělo: %v", body)
	}
	if len(notes) != 0 {
		t.Errorf("prázdné úložiště vrátilo %d poznámek", len(notes))
	}
}

func TestErrorResponses(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)

	tests := []struct {
		name        string
		method      string
		path        string
		contentType string
		body        string
		wantStatus  int
		wantCode    string
	}{
		{"unknown path", http.MethodGet, "/nope", "", "", http.StatusNotFound, "not_found"},
		{"missing note", http.MethodGet, "/notes/9999", "", "", http.StatusNotFound, "not_found"},
		{"wrong method on collection", http.MethodPatch, "/notes", "application/json", `{}`, http.StatusMethodNotAllowed, "method_not_allowed"},
		{"wrong method on item", http.MethodPost, "/notes/1", "application/json", `{}`, http.StatusMethodNotAllowed, "method_not_allowed"},
		{"missing Content-Type", http.MethodPost, "/notes", "", `{"text":"x"}`, http.StatusUnsupportedMediaType, "unsupported_media_type"},
		{"wrong Content-Type", http.MethodPost, "/notes", "text/plain", `{"text":"x"}`, http.StatusUnsupportedMediaType, "unsupported_media_type"},
		{"broken JSON", http.MethodPost, "/notes", "application/json", `{"text":`, http.StatusBadRequest, "bad_request"},
		{"empty text", http.MethodPost, "/notes", "application/json", `{"text":"   "}`, http.StatusBadRequest, "validation_failed"},
		{"missing text", http.MethodPost, "/notes", "application/json", `{}`, http.StatusBadRequest, "validation_failed"},
		{"text too long", http.MethodPost, "/notes", "application/json", `{"text":"` + strings.Repeat("a", 501) + `"}`, http.StatusBadRequest, "validation_failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, srv.URL+tt.path, strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("sestavení požadavku selhalo: %v", err)
			}
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			resp, err := srv.Client().Do(req)
			if err != nil {
				t.Fatalf("požadavek selhal: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, chci %d", resp.StatusCode, tt.wantStatus)
			}
			if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Errorf("Content-Type = %q, chci application/json i u chyby", ct)
			}
			if tt.wantStatus == http.StatusMethodNotAllowed && resp.Header.Get("Allow") == "" {
				t.Error("odpověď 405 nemá hlavičku Allow")
			}
			wantErrorShape(t, decodeJSON(t, resp.Body), tt.wantCode)
		})
	}
}

func TestServerSetsRequestID(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)
	resp, err := srv.Client().Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz selhal: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.Header.Get(exercise.RequestIDHeader) == "" {
		t.Errorf("odpověď nemá hlavičku %s — je middleware v chainu?", exercise.RequestIDHeader)
	}
}

func TestServerConcurrentWrites(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)

	const n = 30
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()

			req, err := http.NewRequest(http.MethodPost, srv.URL+"/notes", strings.NewReader(`{"text":"souběžně"}`))
			if err != nil {
				t.Errorf("sestavení požadavku selhalo: %v", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := srv.Client().Do(req)
			if err != nil {
				t.Errorf("POST selhal: %v", err)
				return
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}()
	}
	wg.Wait()

	resp, err := srv.Client().Get(srv.URL + "/notes")
	if err != nil {
		t.Fatalf("GET /notes selhal: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	notes, _ := decodeJSON(t, resp.Body)["notes"].([]any)
	if len(notes) != n {
		t.Errorf("po %d souběžných zápisech je v úložišti %d poznámek", n, len(notes))
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	t.Parallel()

	var buf safeBuffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	panicking := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("něco se rozbilo")
	})
	h := exercise.Chain(panicking, exercise.RequestIDMiddleware, exercise.RecoveryMiddleware(logger))

	srv := httptest.NewServer(h)
	resp, err := srv.Client().Get(srv.URL + "/boom")
	if err != nil {
		t.Fatalf("požadavek selhal: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, chci %d", resp.StatusCode, http.StatusInternalServerError)
	}
	body := decodeJSON(t, resp.Body)
	_ = resp.Body.Close()
	srv.Close()

	wantErrorShape(t, body, "internal_error")

	out := buf.String()
	if out == "" {
		t.Fatal("panika se nezalogovala")
	}
	var rec map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &rec); err != nil {
		t.Fatalf("log není platný JSON: %v\n%s", err, out)
	}
	if rec["level"] != "ERROR" {
		t.Errorf("úroveň logu = %v, chci ERROR", rec["level"])
	}
	if id, _ := rec["request_id"].(string); id == "" {
		t.Error("log paniky nemá request_id z kontextu")
	}
}

func TestRun(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen selhal: %v", err)
	}

	cfg := exercise.Config{Addr: ln.Addr().String(), ShutdownTimeout: 2 * time.Second}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- exercise.Run(ctx, cfg, exercise.NewServer(discardLogger()), ln)
	}()

	baseURL := "http://" + ln.Addr().String()
	resp, err := http.Get(baseURL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz přes Run selhal: %v", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, chci %d", resp.StatusCode, http.StatusOK)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run = %v, chci nil při čistém ukončení", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run se nevrátil do 5 s po zrušení kontextu")
	}

	if _, err := http.Get(baseURL + "/healthz"); err == nil {
		t.Error("server po ukončení pořád odpovídá")
	}
}
