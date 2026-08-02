package exercise_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-24/exercise"
)

// decodeJSON dekóduje tělo odpovědi do v a selže, pokud to není platný JSON.
func decodeJSON(t *testing.T, body []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("tělo není platný JSON: %v (tělo = %q)", err, string(body))
	}
}

// checkJSONContentType ověří, že odpověď má hlavičku Content-Type application/json.
func checkJSONContentType(t *testing.T, got string) {
	t.Helper()
	if !strings.HasPrefix(got, "application/json") {
		t.Errorf("Content-Type = %q, chci application/json", got)
	}
}

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()

	err := exercise.WriteJSON(rec, http.StatusCreated, exercise.HealthResponse{Status: "ok"})
	if err != nil {
		t.Fatalf("WriteJSON vrátil chybu: %v", err)
	}

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, chci %d", res.StatusCode, http.StatusCreated)
	}
	checkJSONContentType(t, res.Header.Get("Content-Type"))

	body, _ := io.ReadAll(res.Body)
	var got exercise.HealthResponse
	decodeJSON(t, body, &got)
	if got.Status != "ok" {
		t.Errorf("status v těle = %q, chci %q", got.Status, "ok")
	}
}

func TestWriteJSONSetsHeaderBeforeWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	if err := exercise.WriteJSON(rec, http.StatusTeapot, map[string]int{"n": 1}); err != nil {
		t.Fatalf("WriteJSON vrátil chybu: %v", err)
	}
	// httptest.ResponseRecorder si při WriteHeader udělá snapshot hlaviček do
	// Result().Header — pokud se Content-Type nastavil až po WriteHeader, tady chybí.
	res := rec.Result()
	defer res.Body.Close()
	checkJSONContentType(t, res.Header.Get("Content-Type"))
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	exercise.HealthHandler().ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, chci %d", res.StatusCode, http.StatusOK)
	}
	checkJSONContentType(t, res.Header.Get("Content-Type"))

	body, _ := io.ReadAll(res.Body)
	var got exercise.HealthResponse
	decodeJSON(t, body, &got)
	if got.Status != "ok" {
		t.Errorf("HealthHandler tělo = %q, chci status \"ok\"", string(body))
	}
}

func TestEchoHandlerSuccess(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantEcho  string
		wantCount int
	}{
		{"without repeat defaults to 1", `{"message":"ahoj"}`, "ahoj", 1},
		{"repeat 3", `{"message":"ab","repeat":3}`, "ababab", 3},
		{"repeat 10 is upper bound", `{"message":"x","repeat":10}`, strings.Repeat("x", 10), 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			exercise.EchoHandler().ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, chci %d", res.StatusCode, http.StatusOK)
			}
			checkJSONContentType(t, res.Header.Get("Content-Type"))

			raw, _ := io.ReadAll(res.Body)
			var got exercise.EchoResponse
			decodeJSON(t, raw, &got)
			if got.Echo != tt.wantEcho {
				t.Errorf("echo = %q, chci %q", got.Echo, tt.wantEcho)
			}
			if got.Count != tt.wantCount {
				t.Errorf("count = %d, chci %d", got.Count, tt.wantCount)
			}
		})
	}
}

func TestEchoHandlerAcceptsContentTypeWithParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"message":"a"}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()

	exercise.EchoHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, chci %d (charset parametr je platný)", rec.Code, http.StatusOK)
	}
}

func TestEchoHandlerErrors(t *testing.T) {
	velkeTelo := `{"message":"` + strings.Repeat("a", 4*exercise.MaxBodyBytes) + `"}`

	tests := []struct {
		name        string
		method      string
		contentType string
		body        string
		wantStatus  int
	}{
		{"GET instead of POST", http.MethodGet, "application/json", "", http.StatusMethodNotAllowed},
		{"PUT instead of POST", http.MethodPut, "application/json", `{"message":"a"}`, http.StatusMethodNotAllowed},
		{"missing Content-Type", http.MethodPost, "", `{"message":"a"}`, http.StatusUnsupportedMediaType},
		{"text/plain", http.MethodPost, "text/plain", `{"message":"a"}`, http.StatusUnsupportedMediaType},
		{"broken JSON", http.MethodPost, "application/json", `{"message":`, http.StatusBadRequest},
		{"unknown field", http.MethodPost, "application/json", `{"message":"a","nope":1}`, http.StatusBadRequest},
		{"empty message", http.MethodPost, "application/json", `{"message":"   "}`, http.StatusBadRequest},
		{"negative repeat", http.MethodPost, "application/json", `{"message":"a","repeat":-2}`, http.StatusBadRequest},
		{"repeat too large", http.MethodPost, "application/json", `{"message":"a","repeat":11}`, http.StatusBadRequest},
		{"body too large", http.MethodPost, "application/json", velkeTelo, http.StatusRequestEntityTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/echo", strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rec := httptest.NewRecorder()

			exercise.EchoHandler().ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, chci %d", res.StatusCode, tt.wantStatus)
			}
			checkJSONContentType(t, res.Header.Get("Content-Type"))

			raw, _ := io.ReadAll(res.Body)
			var got exercise.ErrorResponse
			decodeJSON(t, raw, &got)
			if got.Error == "" {
				t.Errorf("chybová odpověď má prázdné pole error (tělo = %q)", string(raw))
			}
		})
	}
}

func TestEchoHandlerSendsAllowOn405(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/echo", nil)
	rec := httptest.NewRecorder()

	exercise.EchoHandler().ServeHTTP(rec, req)

	if got := rec.Result().Header.Get("Allow"); !strings.Contains(got, http.MethodPost) {
		t.Errorf("hlavička Allow = %q, chci aby obsahovala %q", got, http.MethodPost)
	}
}

func TestNotFoundHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/nekde", nil)
	rec := httptest.NewRecorder()

	exercise.NotFoundHandler().ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, chci %d", res.StatusCode, http.StatusNotFound)
	}
	checkJSONContentType(t, res.Header.Get("Content-Type"))
}

func TestNewRouterRouting(t *testing.T) {
	router := exercise.NewRouter()

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{"health", http.MethodGet, "/healthz", "", http.StatusOK},
		{"echo POST", http.MethodPost, "/echo", `{"message":"a"}`, http.StatusOK},
		{"echo GET je 405", http.MethodGet, "/echo", "", http.StatusMethodNotAllowed},
		{"unknown path", http.MethodGet, "/tohle-neexistuje", "", http.StatusNotFound},
		{"root", http.MethodGet, "/", "", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.wantStatus {
				t.Errorf("%s %s = %d, chci %d", tt.method, tt.path, res.StatusCode, tt.wantStatus)
			}
			checkJSONContentType(t, res.Header.Get("Content-Type"))
		})
	}
}

func TestNewRouterViaRealServer(t *testing.T) {
	srv := httptest.NewServer(exercise.NewRouter())
	defer srv.Close()

	res, err := srv.Client().Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz selhal: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, chci %d", res.StatusCode, http.StatusOK)
	}

	raw, _ := io.ReadAll(res.Body)
	var got exercise.HealthResponse
	decodeJSON(t, raw, &got)
	if got.Status != "ok" {
		t.Errorf("status v těle = %q, chci %q", got.Status, "ok")
	}

	res2, err := srv.Client().Post(srv.URL+"/echo", "application/json", strings.NewReader(`{"message":"hej","repeat":2}`))
	if err != nil {
		t.Fatalf("POST /echo selhal: %v", err)
	}
	defer res2.Body.Close()

	raw2, _ := io.ReadAll(res2.Body)
	var echo exercise.EchoResponse
	decodeJSON(t, raw2, &echo)
	if echo.Echo != "hejhej" {
		t.Errorf("echo = %q, chci %q", echo.Echo, "hejhej")
	}
}

func TestNewServerHasTimeouts(t *testing.T) {
	router := exercise.NewRouter()
	srv := exercise.NewServer(":8080", router)

	if srv == nil {
		t.Fatal("NewServer vrátil nil")
	}
	if srv.Addr != ":8080" {
		t.Errorf("Addr = %q, chci %q", srv.Addr, ":8080")
	}
	if srv.Handler == nil {
		t.Error("Handler je nil, chci předaný router")
	}

	if srv.ReadHeaderTimeout <= 0 {
		t.Errorf("ReadHeaderTimeout = %v, chci kladnou hodnotu", srv.ReadHeaderTimeout)
	}
	if srv.ReadTimeout <= 0 {
		t.Errorf("ReadTimeout = %v, chci kladnou hodnotu", srv.ReadTimeout)
	}
	if srv.WriteTimeout <= 0 {
		t.Errorf("WriteTimeout = %v, chci kladnou hodnotu", srv.WriteTimeout)
	}
	if srv.IdleTimeout <= 0 {
		t.Errorf("IdleTimeout = %v, chci kladnou hodnotu", srv.IdleTimeout)
	}
	if srv.ReadHeaderTimeout > srv.ReadTimeout {
		t.Errorf("ReadHeaderTimeout (%v) nemá být delší než ReadTimeout (%v)", srv.ReadHeaderTimeout, srv.ReadTimeout)
	}
}
