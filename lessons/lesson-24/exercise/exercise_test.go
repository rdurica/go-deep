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

	raw, _ := io.ReadAll(res.Body)
	var got exercise.ErrorResponse
	decodeJSON(t, raw, &got)
	if got.Error == "" {
		t.Errorf("404 odpověď má prázdné pole error (tělo = %q)", string(raw))
	}
}

func TestNewServerHasTimeouts(t *testing.T) {
	h := exercise.HealthHandler()
	srv := exercise.NewServer(":8080", h)

	if srv == nil {
		t.Fatal("NewServer vrátil nil")
	}
	if srv.Addr != ":8080" {
		t.Errorf("Addr = %q, chci %q", srv.Addr, ":8080")
	}
	if srv.Handler == nil {
		t.Error("Handler je nil, chci předaný handler")
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
