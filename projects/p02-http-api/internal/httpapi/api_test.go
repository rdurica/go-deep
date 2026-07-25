package httpapi_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p02-http-api/internal/httpapi"
	"github.com/rdurica/go-deep/projects/p02-http-api/internal/task"
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

// newAPI spustí celé API na náhodném portu localhostu a vrátí adresu a logy.
func newAPI(t *testing.T) (*httptest.Server, *safeBuffer) {
	t.Helper()

	logs := &safeBuffer{}
	logger := slog.New(slog.NewJSONHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	srv := httptest.NewServer(httpapi.NewRouter(task.NewStore(), logger))
	t.Cleanup(srv.Close)
	return srv, logs
}

// do pošle požadavek a vrátí odpověď; tělo zavírá volající.
func do(t *testing.T, srv *httptest.Server, method, path, contentType, body string) *http.Response {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, srv.URL+path, reader)
	if err != nil {
		t.Fatalf("sestavení požadavku selhalo: %v", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("%s %s selhal: %v", method, path, err)
	}
	return resp
}

// decode rozparsuje tělo odpovědi do mapy a zavře ho.
func decode(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()

	var m map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		t.Fatalf("odpověď není platný JSON: %v", err)
	}
	return m
}

// createTask vytvoří úkol a vrátí jeho ID.
func createTask(t *testing.T, srv *httptest.Server, title, status string) string {
	t.Helper()

	body := `{"title":` + strconv.Quote(title) + `,"status":` + strconv.Quote(status) + `}`
	resp := do(t, srv, http.MethodPost, "/tasks", "application/json", body)
	if resp.StatusCode != http.StatusCreated {
		defer func() { _ = resp.Body.Close() }()
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST /tasks status = %d, chci %d; tělo: %s", resp.StatusCode, http.StatusCreated, payload)
	}
	created := decode(t, resp)
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatalf("vytvořený úkol nemá id; tělo: %v", created)
	}
	return id
}

// wantError ověří konzistentní tvar chybové odpovědi.
func wantError(t *testing.T, resp *http.Response, wantStatus int, wantCode string) {
	t.Helper()

	if resp.StatusCode != wantStatus {
		t.Errorf("status = %d, chci %d", resp.StatusCode, wantStatus)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, chci application/json i u chyby", ct)
	}

	body := decode(t, resp)
	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("chybová odpověď nemá objekt error; tělo: %v", body)
	}
	if code, _ := errObj["code"].(string); code != wantCode {
		t.Errorf("error.code = %q, chci %q", code, wantCode)
	}
	if msg, _ := errObj["message"].(string); msg == "" {
		t.Error("error.message je prázdná")
	}
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)
	resp := do(t, srv, http.MethodGet, "/healthz", "", "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, chci %d", resp.StatusCode, http.StatusOK)
	}
	if body := decode(t, resp); body["status"] != "ok" {
		t.Errorf("tělo = %v, chci status=ok", body)
	}
}

func TestCreateTask(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)
	resp := do(t, srv, http.MethodPost, "/tasks", "application/json", `{"title":"napsat testy","status":"doing"}`)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, chci %d", resp.StatusCode, http.StatusCreated)
	}
	location := resp.Header.Get("Location")
	body := decode(t, resp)

	id, _ := body["id"].(string)
	if id == "" {
		t.Fatal("odpověď nemá id")
	}
	if location != "/tasks/"+id {
		t.Errorf("Location = %q, chci %q", location, "/tasks/"+id)
	}
	if body["title"] != "napsat testy" {
		t.Errorf("title = %v, chci %q", body["title"], "napsat testy")
	}
	if body["status"] != "doing" {
		t.Errorf("status = %v, chci %q", body["status"], "doing")
	}
	for _, key := range []string{"created_at", "updated_at"} {
		raw, _ := body[key].(string)
		if _, err := time.Parse(time.RFC3339, raw); err != nil {
			t.Errorf("%s = %q, chci čas v RFC3339: %v", key, raw, err)
		}
	}
}

func TestCreateTaskDefaultStatus(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)
	resp := do(t, srv, http.MethodPost, "/tasks", "application/json", `{"title":"bez stavu"}`)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, chci %d", resp.StatusCode, http.StatusCreated)
	}
	if body := decode(t, resp); body["status"] != "todo" {
		t.Errorf("status = %v, chci výchozí %q", body["status"], "todo")
	}
}

func TestGetTask(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)
	id := createTask(t, srv, "koupit mléko", "todo")

	resp := do(t, srv, http.MethodGet, "/tasks/"+id, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, chci %d", resp.StatusCode, http.StatusOK)
	}
	body := decode(t, resp)
	if body["id"] != id || body["title"] != "koupit mléko" {
		t.Errorf("tělo = %v, chci úkol %q s názvem %q", body, id, "koupit mléko")
	}
}

func TestListTasks(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)

	// Prázdný seznam musí být [] a ne null.
	resp := do(t, srv, http.MethodGet, "/tasks", "", "")
	body := decode(t, resp)
	tasks, ok := body["tasks"].([]any)
	if !ok {
		t.Fatalf("odpověď nemá pole tasks; tělo: %v", body)
	}
	if len(tasks) != 0 {
		t.Errorf("prázdné úložiště vrátilo %d úkolů", len(tasks))
	}

	createTask(t, srv, "první", "todo")
	createTask(t, srv, "druhý", "done")

	resp = do(t, srv, http.MethodGet, "/tasks", "", "")
	body = decode(t, resp)
	tasks, _ = body["tasks"].([]any)
	if len(tasks) != 2 {
		t.Fatalf("počet úkolů = %d, chci 2", len(tasks))
	}
	first, _ := tasks[0].(map[string]any)
	if first["title"] != "první" {
		t.Errorf("pořadí se nezachovalo, první je %v", first["title"])
	}
}

func TestUpdateTask(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)
	id := createTask(t, srv, "původní", "todo")

	resp := do(t, srv, http.MethodPut, "/tasks/"+id, "application/json", `{"title":"změněný","status":"done"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, chci %d", resp.StatusCode, http.StatusOK)
	}
	body := decode(t, resp)
	if body["title"] != "změněný" || body["status"] != "done" {
		t.Errorf("tělo = %v, chci title=změněný status=done", body)
	}

	// Změna musí být vidět i při dalším čtení.
	after := decode(t, do(t, srv, http.MethodGet, "/tasks/"+id, "", ""))
	if after["title"] != "změněný" {
		t.Errorf("po PUT vrací GET %v", after["title"])
	}
}

func TestDeleteTask(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)
	id := createTask(t, srv, "smazat mě", "todo")

	resp := do(t, srv, http.MethodDelete, "/tasks/"+id, "", "")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, chci %d", resp.StatusCode, http.StatusNoContent)
	}
	if n, _ := io.Copy(io.Discard, resp.Body); n != 0 {
		t.Errorf("odpověď 204 má prázdné tělo, dostal jsem %d bajtů", n)
	}

	wantError(t, do(t, srv, http.MethodGet, "/tasks/"+id, "", ""), http.StatusNotFound, httpapi.CodeNotFound)
}

func TestNotFound(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"neznámá cesta", http.MethodGet, "/neexistuje"},
		{"neexistující úkol", http.MethodGet, "/tasks/9999"},
		{"mazání neexistujícího úkolu", http.MethodDelete, "/tasks/9999"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantError(t, do(t, srv, tt.method, tt.path, "", ""), http.StatusNotFound, httpapi.CodeNotFound)
		})
	}
}

func TestUpdateNonexistentTask(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)
	resp := do(t, srv, http.MethodPut, "/tasks/9999", "application/json", `{"title":"nový","status":"todo"}`)
	wantError(t, resp, http.StatusNotFound, httpapi.CodeNotFound)
}

func TestBadRequest(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)

	tests := []struct {
		name     string
		method   string
		path     string
		body     string
		wantCode string
	}{
		{"rozbitý JSON", http.MethodPost, "/tasks", `{"title":`, httpapi.CodeBadRequest},
		{"prázdné tělo", http.MethodPost, "/tasks", ``, httpapi.CodeBadRequest},
		{"neznámé pole", http.MethodPost, "/tasks", `{"title":"x","priority":1}`, httpapi.CodeBadRequest},
		{"dva JSON dokumenty", http.MethodPost, "/tasks", `{"title":"x"}{"title":"y"}`, httpapi.CodeBadRequest},
		{"prázdný název", http.MethodPost, "/tasks", `{"title":"   "}`, httpapi.CodeValidationFailed},
		{"chybějící název", http.MethodPost, "/tasks", `{"status":"todo"}`, httpapi.CodeValidationFailed},
		{"dlouhý název", http.MethodPost, "/tasks", `{"title":"` + strings.Repeat("a", task.MaxTitleLength+1) + `"}`, httpapi.CodeValidationFailed},
		{"neznámý stav", http.MethodPost, "/tasks", `{"title":"x","status":"hotovo"}`, httpapi.CodeValidationFailed},
		{"neplatný PUT", http.MethodPut, "/tasks/1", `{"title":""}`, httpapi.CodeValidationFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := do(t, srv, tt.method, tt.path, "application/json", tt.body)
			wantError(t, resp, http.StatusBadRequest, tt.wantCode)
		})
	}
}

func TestUnsupportedMediaType(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)

	tests := []struct {
		name        string
		contentType string
	}{
		{"chybějící Content-Type", ""},
		{"text/plain", "text/plain"},
		{"formulář", "application/x-www-form-urlencoded"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := do(t, srv, http.MethodPost, "/tasks", tt.contentType, `{"title":"x"}`)
			wantError(t, resp, http.StatusUnsupportedMediaType, httpapi.CodeUnsupportedMediaType)
		})
	}
}

func TestContentTypeWithCharset(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)
	resp := do(t, srv, http.MethodPost, "/tasks", "application/json; charset=utf-8", `{"title":"s parametrem"}`)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, chci %d — parametr charset je legitimní", resp.StatusCode, http.StatusCreated)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"PATCH na kolekci", http.MethodPatch, "/tasks"},
		{"DELETE na kolekci", http.MethodDelete, "/tasks"},
		{"POST na položku", http.MethodPost, "/tasks/1"},
		{"POST na healthz", http.MethodPost, "/healthz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := do(t, srv, tt.method, tt.path, "application/json", `{}`)
			if allow := resp.Header.Get("Allow"); allow == "" {
				t.Error("odpověď 405 nemá hlavičku Allow")
			}
			wantError(t, resp, http.StatusMethodNotAllowed, httpapi.CodeMethodNotAllowed)
		})
	}
}

func TestRequestIDHeader(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)

	resp := do(t, srv, http.MethodGet, "/healthz", "", "")
	generated := resp.Header.Get(httpapi.RequestIDHeader)
	_ = resp.Body.Close()
	if generated == "" {
		t.Fatalf("odpověď nemá hlavičku %s", httpapi.RequestIDHeader)
	}

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/healthz", nil)
	if err != nil {
		t.Fatalf("sestavení požadavku selhalo: %v", err)
	}
	req.Header.Set(httpapi.RequestIDHeader, "trace-123")
	incoming, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("požadavek selhal: %v", err)
	}
	defer func() { _ = incoming.Body.Close() }()

	if got := incoming.Header.Get(httpapi.RequestIDHeader); got != "trace-123" {
		t.Errorf("request ID = %q, chci převzít příchozí %q", got, "trace-123")
	}
}

func TestRequestIsLogged(t *testing.T) {
	t.Parallel()

	srv, logs := newAPI(t)
	resp := do(t, srv, http.MethodGet, "/healthz", "", "")
	_ = resp.Body.Close()
	srv.Close() // počká na dokončení požadavků, takže log je zapsaný

	out := strings.TrimSpace(logs.String())
	if out == "" {
		t.Fatal("požadavek se nezalogoval")
	}
	var rec map[string]any
	if err := json.Unmarshal([]byte(strings.Split(out, "\n")[0]), &rec); err != nil {
		t.Fatalf("log není platný JSON: %v\n%s", err, out)
	}
	if rec["method"] != http.MethodGet || rec["path"] != "/healthz" {
		t.Errorf("log = %v, chci method=GET path=/healthz", rec)
	}
	if status, _ := rec["status"].(float64); status != http.StatusOK {
		t.Errorf("status v logu = %v, chci 200", rec["status"])
	}
	if id, _ := rec["request_id"].(string); id == "" {
		t.Error("log nemá request_id")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	t.Parallel()

	logs := &safeBuffer{}
	logger := slog.New(slog.NewJSONHandler(logs, nil))
	panicking := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("rozbité")
	})

	srv := httptest.NewServer(httpapi.Chain(panicking,
		httpapi.RequestID,
		httpapi.Logging(logger),
		httpapi.Recovery(logger),
	))
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/boom")
	if err != nil {
		t.Fatalf("požadavek selhal: %v", err)
	}
	wantError(t, resp, http.StatusInternalServerError, httpapi.CodeInternalError)

	if !strings.Contains(logs.String(), "panic recovered") {
		t.Errorf("panika se nezalogovala; logy:\n%s", logs.String())
	}
}

func TestConcurrentRequests(t *testing.T) {
	t.Parallel()

	srv, _ := newAPI(t)

	// V goroutině se nesmí volat t.Fatalf, proto tu požadavek posíláme ručně
	// místo přes pomocnou funkci do().
	send := func(method, body string) {
		var reader io.Reader
		if body != "" {
			reader = strings.NewReader(body)
		}
		req, err := http.NewRequest(method, srv.URL+"/tasks", reader)
		if err != nil {
			t.Errorf("sestavení požadavku selhalo: %v", err)
			return
		}
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Errorf("%s /tasks selhal: %v", method, err)
			return
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}

	const n = 25
	var wg sync.WaitGroup
	wg.Add(n * 2)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			send(http.MethodPost, `{"title":"souběžný úkol"}`)
		}()
		go func() {
			defer wg.Done()
			send(http.MethodGet, "")
		}()
	}
	wg.Wait()

	body := decode(t, do(t, srv, http.MethodGet, "/tasks", "", ""))
	tasks, _ := body["tasks"].([]any)
	if len(tasks) != n {
		t.Errorf("po %d souběžných zápisech je v API %d úkolů", n, len(tasks))
	}
}
