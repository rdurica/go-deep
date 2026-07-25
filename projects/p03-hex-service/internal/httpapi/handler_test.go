package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/app"
	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/httpapi"
	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/memstore"
	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/order"
)

var fixedTime = time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

// stubIDs vydává předvídatelná ID.
type stubIDs struct {
	mu sync.Mutex
	n  int
}

func (s *stubIDs) NewID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.n++
	return "ord-" + strconv.Itoa(s.n)
}

// brokenRepo je fake, který simuluje nedostupné úložiště.
type brokenRepo struct {
	err error
}

func (b brokenRepo) Save(ctx context.Context, o order.Order) error { return b.err }

func (b brokenRepo) Find(ctx context.Context, id string) (order.Order, error) {
	return order.Order{}, fmt.Errorf("%w: %s", order.ErrNotFound, id)
}

func (b brokenRepo) List(ctx context.Context) ([]order.Order, error) { return nil, b.err }

// Integrační test celé služby: skutečná doména, skutečný in-memory
// adaptér, skutečný router. Jen čas a ID jsou stubnuté, aby byl
// výsledek deterministický.
func newServer() http.Handler {
	svc := app.NewService(memstore.New(), app.ClockFunc(func() time.Time { return fixedTime }), &stubIDs{})
	return httpapi.NewHandler(svc)
}

const validBody = `{"customer":"radek@example.com","currency":"CZK","lines":[
	{"sku":"kniha-go","quantity":2,"unit_price_cents":49900},
	{"sku":"hrnek","quantity":1,"unit_price_cents":19900}
]}`

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("tělo není JSON objekt (%v): %s", err, rec.Body.String())
	}
	return out
}

func TestHealthAReadiness(t *testing.T) {
	h := newServer()
	for _, path := range []string{"/healthz", "/readyz"} {
		rec := do(t, h, http.MethodGet, path, "")
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, chci 200", path, rec.Code)
		}
	}
}

func TestReadinessPriNedostupnemUlozisti(t *testing.T) {
	svc := app.NewService(brokenRepo{err: errors.New("connection refused")},
		app.ClockFunc(func() time.Time { return fixedTime }), &stubIDs{})
	rec := do(t, httpapi.NewHandler(svc), http.MethodGet, "/readyz", "")

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("GET /readyz = %d, chci 503", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != httpapi.ProblemContentType {
		t.Errorf("Content-Type = %q, chci %q", got, httpapi.ProblemContentType)
	}
}

func TestZalozeniObjednavky(t *testing.T) {
	rec := do(t, newServer(), http.MethodPost, "/orders", validBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, chci 201 (tělo: %s)", rec.Code, rec.Body.String())
	}

	body := decode(t, rec)
	if body["id"] != "ord-1" {
		t.Errorf("id = %v, chci ord-1", body["id"])
	}
	if body["status"] != "new" {
		t.Errorf("status = %v, chci new", body["status"])
	}
	if body["currency"] != "CZK" {
		t.Errorf("currency = %v, chci CZK", body["currency"])
	}
	if body["total_cents"] != float64(119700) {
		t.Errorf("total_cents = %v, chci 119700", body["total_cents"])
	}
	if body["total"] != "1197.00 CZK" {
		t.Errorf("total = %v, chci %q", body["total"], "1197.00 CZK")
	}
	lines, _ := body["lines"].([]any)
	if len(lines) != 2 {
		t.Fatalf("lines = %v, chci dvě položky", body["lines"])
	}
	first, _ := lines[0].(map[string]any)
	if first["total_cents"] != float64(99800) {
		t.Errorf("lines[0].total_cents = %v, chci 99800", first["total_cents"])
	}
}

func TestChybneVstupy(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"rozbitý JSON", `{"customer":`, http.StatusBadRequest},
		{"neznámé pole", `{"customer":"radek","currency":"CZK","lines":[],"discount":10}`, http.StatusBadRequest},
		{"bez zákazníka", `{"currency":"CZK","lines":[{"sku":"x","quantity":1,"unit_price_cents":100}]}`, http.StatusUnprocessableEntity},
		{"bez položek", `{"customer":"radek","currency":"CZK","lines":[]}`, http.StatusUnprocessableEntity},
		{"neplatná měna", `{"customer":"radek","currency":"koruna","lines":[{"sku":"x","quantity":1,"unit_price_cents":100}]}`, http.StatusUnprocessableEntity},
		{"nulové množství", `{"customer":"radek","currency":"CZK","lines":[{"sku":"x","quantity":0,"unit_price_cents":100}]}`, http.StatusUnprocessableEntity},
		{"záporná cena", `{"customer":"radek","currency":"CZK","lines":[{"sku":"x","quantity":1,"unit_price_cents":-5}]}`, http.StatusUnprocessableEntity},
		{"prázdné SKU", `{"customer":"radek","currency":"CZK","lines":[{"sku":"  ","quantity":1,"unit_price_cents":100}]}`, http.StatusUnprocessableEntity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, newServer(), http.MethodPost, "/orders", tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, chci %d (tělo: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if got := rec.Header().Get("Content-Type"); got != httpapi.ProblemContentType {
				t.Errorf("Content-Type = %q, chci %q", got, httpapi.ProblemContentType)
			}
			body := decode(t, rec)
			if title, _ := body["title"].(string); title == "" {
				t.Errorf("v problem+json chybí title: %v", body)
			}
			if body["status"] != float64(tt.wantStatus) {
				t.Errorf("status v těle = %v, chci %d", body["status"], tt.wantStatus)
			}
			if body["type"] == nil {
				t.Errorf("v problem+json chybí type: %v", body)
			}
		})
	}
}

func TestZivotniCyklusPresHTTP(t *testing.T) {
	h := newServer()

	created := do(t, h, http.MethodPost, "/orders", validBody)
	if created.Code != http.StatusCreated {
		t.Fatalf("POST /orders = %d, chci 201", created.Code)
	}
	id, _ := decode(t, created)["id"].(string)

	for _, step := range []struct {
		path string
		want string
	}{
		{"/orders/" + id + "/pay", "paid"},
		{"/orders/" + id + "/ship", "shipped"},
	} {
		rec := do(t, h, http.MethodPost, step.path, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("POST %s = %d, chci 200 (tělo: %s)", step.path, rec.Code, rec.Body.String())
		}
		if got := decode(t, rec)["status"]; got != step.want {
			t.Errorf("POST %s: status = %v, chci %q", step.path, got, step.want)
		}
	}

	got := do(t, h, http.MethodGet, "/orders/"+id, "")
	if got.Code != http.StatusOK {
		t.Fatalf("GET /orders/%s = %d, chci 200", id, got.Code)
	}
	if s := decode(t, got)["status"]; s != "shipped" {
		t.Errorf("stav po znovunačtení = %v, chci shipped", s)
	}
}

func TestKonfliktStavu(t *testing.T) {
	h := newServer()
	created := do(t, h, http.MethodPost, "/orders", validBody)
	id, _ := decode(t, created)["id"].(string)

	if rec := do(t, h, http.MethodPost, "/orders/"+id+"/ship", ""); rec.Code != http.StatusConflict {
		t.Errorf("ship nezaplacené = %d, chci 409 (tělo: %s)", rec.Code, rec.Body.String())
	}
	if rec := do(t, h, http.MethodPost, "/orders/"+id+"/pay", ""); rec.Code != http.StatusOK {
		t.Fatalf("pay = %d, chci 200", rec.Code)
	}
	if rec := do(t, h, http.MethodPost, "/orders/"+id+"/pay", ""); rec.Code != http.StatusConflict {
		t.Errorf("druhé pay = %d, chci 409", rec.Code)
	}
	if rec := do(t, h, http.MethodPost, "/orders/"+id+"/ship", ""); rec.Code != http.StatusOK {
		t.Fatalf("ship = %d, chci 200", rec.Code)
	}
	if rec := do(t, h, http.MethodPost, "/orders/"+id+"/cancel", ""); rec.Code != http.StatusConflict {
		t.Errorf("cancel odeslané = %d, chci 409", rec.Code)
	}
}

func TestZruseniObjednavky(t *testing.T) {
	h := newServer()
	created := do(t, h, http.MethodPost, "/orders", validBody)
	id, _ := decode(t, created)["id"].(string)

	rec := do(t, h, http.MethodPost, "/orders/"+id+"/cancel", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("cancel = %d, chci 200 (tělo: %s)", rec.Code, rec.Body.String())
	}
	if got := decode(t, rec)["status"]; got != "cancelled" {
		t.Errorf("status = %v, chci cancelled", got)
	}
}

func TestNenalezeno(t *testing.T) {
	h := newServer()
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/orders/neexistuje"},
		{http.MethodPost, "/orders/neexistuje/pay"},
		{http.MethodPost, "/orders/neexistuje/ship"},
		{http.MethodPost, "/orders/neexistuje/cancel"},
	}
	for _, c := range cases {
		rec := do(t, h, c.method, c.path, "")
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s %s = %d, chci 404 (tělo: %s)", c.method, c.path, rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("Content-Type"); got != httpapi.ProblemContentType {
			t.Errorf("%s: Content-Type = %q, chci %q", c.path, got, httpapi.ProblemContentType)
		}
	}
}

func TestVypisObjednavek(t *testing.T) {
	h := newServer()
	rec := do(t, h, http.MethodGet, "/orders", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /orders = %d, chci 200", rec.Code)
	}
	if orders, _ := decode(t, rec)["orders"].([]any); len(orders) != 0 {
		t.Errorf("prázdná služba vrátila %d objednávek", len(orders))
	}

	for i := 0; i < 3; i++ {
		if r := do(t, h, http.MethodPost, "/orders", validBody); r.Code != http.StatusCreated {
			t.Fatalf("POST /orders = %d, chci 201", r.Code)
		}
	}

	rec = do(t, h, http.MethodGet, "/orders", "")
	orders, _ := decode(t, rec)["orders"].([]any)
	if len(orders) != 3 {
		t.Fatalf("GET /orders = %d objednávek, chci 3", len(orders))
	}
	for i, want := range []string{"ord-1", "ord-2", "ord-3"} {
		got, _ := orders[i].(map[string]any)
		if got["id"] != want {
			t.Errorf("orders[%d].id = %v, chci %q (seřazeno podle ID)", i, got["id"], want)
		}
	}
}

func TestRouting(t *testing.T) {
	h := newServer()
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/healthz"},
		{http.MethodDelete, "/orders/ord-1"},
		{http.MethodPut, "/orders"},
		{http.MethodGet, "/orders/ord-1/pay"},
	}
	for _, c := range cases {
		if rec := do(t, h, c.method, c.path, ""); rec.Code == http.StatusOK {
			t.Errorf("%s %s = 200, chci odmítnutí", c.method, c.path)
		}
	}
}

func TestInterniChybaNeunikne(t *testing.T) {
	secret := "pq: password authentication failed for user \"orders\" at 10.0.0.7"
	svc := app.NewService(brokenRepo{err: errors.New(secret)},
		app.ClockFunc(func() time.Time { return fixedTime }), &stubIDs{})

	rec := do(t, httpapi.NewHandler(svc), http.MethodPost, "/orders", validBody)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, chci 500 (tělo: %s)", rec.Code, rec.Body.String())
	}
	for _, leak := range []string{"10.0.0.7", "password", "pq:"} {
		if strings.Contains(rec.Body.String(), leak) {
			t.Errorf("tělo prozradilo %q: %s", leak, rec.Body.String())
		}
	}
}

func TestPrilisVelkeTelo(t *testing.T) {
	huge := `{"customer":"` + strings.Repeat("a", 2<<20) + `","currency":"CZK","lines":[]}`
	rec := do(t, newServer(), http.MethodPost, "/orders", huge)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, chci 400 — tělo musí být omezené", rec.Code)
	}
}
