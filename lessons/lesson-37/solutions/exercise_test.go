package solutions_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-37/solutions"
)

func TestParseBearerOK(t *testing.T) {
	tests := []struct {
		header string
		want   string
	}{
		{"Bearer abc123", "abc123"},
		{"bearer abc123", "abc123"},
		{"BEARER abc123", "abc123"},
		{"BeArEr abc123", "abc123"},
		{"  Bearer   abc123  ", "abc123"},
		{"Bearer\tabc123", "abc123"},
	}
	for _, tt := range tests {
		got, err := exercise.ParseBearer(tt.header)
		if err != nil {
			t.Errorf("ParseBearer(%q) = chyba %v, chci token %q", tt.header, err, tt.want)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseBearer(%q) = %q, chci %q", tt.header, got, tt.want)
		}
	}
}

func TestParseBearerErrors(t *testing.T) {
	tests := []struct {
		header  string
		wantErr error
	}{
		{"", exercise.ErrMissingAuthorization},
		{"   ", exercise.ErrMissingAuthorization},
		{"\t\n", exercise.ErrMissingAuthorization},
		{"Basic YWxhZGRpbjpvcGVuc2VzYW1l", exercise.ErrUnsupportedScheme},
		{"Token abc123", exercise.ErrUnsupportedScheme},
		{"abc123", exercise.ErrUnsupportedScheme},
		{"Bearer", exercise.ErrMissingToken},
		{"bearer   ", exercise.ErrMissingToken},
		{"Bearer abc 123", exercise.ErrMalformedAuthorization},
	}
	for _, tt := range tests {
		got, err := exercise.ParseBearer(tt.header)
		if !errors.Is(err, tt.wantErr) {
			t.Errorf("ParseBearer(%q) = (%q, %v), chci chybu %v", tt.header, got, err, tt.wantErr)
		}
		if got != "" {
			t.Errorf("ParseBearer(%q) vrátil při chybě token %q, chci prázdný", tt.header, got)
		}
	}
}

func TestHashPassword(t *testing.T) {
	const password = "tajneheslo"
	hash := exercise.HashPassword(password, "sul")

	if hash == "" {
		t.Fatal("HashPassword vrátil prázdný řetězec")
	}
	if strings.Contains(hash, password) {
		t.Errorf("otisk %q obsahuje heslo v čitelné podobě", hash)
	}
	if len(hash) != 64 {
		t.Errorf("len(hash) = %d, chci 64 (SHA-256 v hexu)", len(hash))
	}
	if again := exercise.HashPassword(password, "sul"); again != hash {
		t.Errorf("HashPassword není deterministická: %q vs %q", again, hash)
	}
	if other := exercise.HashPassword(password, "jinasul"); other == hash {
		t.Error("jiná sůl musí dát jiný otisk")
	}
	if other := exercise.HashPassword("jineheslo", "sul"); other == hash {
		t.Error("jiné heslo musí dát jiný otisk")
	}
}

func TestVerifyPassword(t *testing.T) {
	hash := exercise.HashPassword("tajneheslo", "sul")

	if !exercise.VerifyPassword(hash, "tajneheslo", "sul") {
		t.Error("VerifyPassword se správným heslem má vrátit true")
	}
	if exercise.VerifyPassword(hash, "tajneheslo", "jinasul") {
		t.Error("jiná sůl nesmí projít")
	}
	if exercise.VerifyPassword(hash, "tajneheslX", "sul") {
		t.Error("jiné heslo nesmí projít")
	}
	if exercise.VerifyPassword(hash, "tajneheslo!", "sul") {
		t.Error("delší heslo nesmí projít")
	}
	if exercise.VerifyPassword("", "tajneheslo", "sul") {
		t.Error("prázdný otisk nesmí projít")
	}
}

// echoUser odpoví jménem uživatele z kontextu a poznamená, že byl zavolán.
func echoUser(called *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		user, ok := exercise.UserFrom(r.Context())
		if !ok {
			http.Error(w, "v kontextu chybí uživatel", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, user)
	})
}

func TestAuthenticateAllowsValidToken(t *testing.T) {
	tokens := map[string]string{"tok-alice": "alice", "tok-bob": "bob"}
	called := false
	handler := exercise.Authenticate(tokens)(echoUser(&called))

	for header, want := range map[string]string{
		"Bearer tok-alice": "alice",
		"bearer tok-bob":   "bob",
	} {
		called = false
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", header)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%q: status = %d, chci 200 (tělo: %s)", header, rec.Code, rec.Body.String())
		}
		if !called {
			t.Fatalf("%q: handler nebyl zavolaný", header)
		}
		if got := rec.Body.String(); got != want {
			t.Errorf("%q: uživatel v kontextu = %q, chci %q", header, got, want)
		}
	}
}

func TestAuthenticateRejects(t *testing.T) {
	tokens := map[string]string{"tok-alice": "alice"}
	tests := map[string]string{
		"bez hlavičky":     "",
		"prázdná hlavička": "   ",
		"other scheme":     "Basic dG9rLWFsaWNl",
		"unknown token":    "Bearer tok-mallory",
		"prefix tokenu":    "Bearer tok-alic",
		"token s příponou": "Bearer tok-alice2",
		"chybí token":      "Bearer",
	}
	for name, header := range tests {
		t.Run(name, func(t *testing.T) {
			called := false
			handler := exercise.Authenticate(tokens)(echoUser(&called))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if header != "" {
				req.Header.Set("Authorization", header)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, chci 401", rec.Code)
			}
			if called {
				t.Error("chráněný handler se nesmí zavolat")
			}
			if got := rec.Header().Get("WWW-Authenticate"); got == "" {
				t.Error("401 musí mít hlavičku WWW-Authenticate")
			} else if !strings.Contains(strings.ToLower(got), "bearer") {
				t.Errorf("WWW-Authenticate = %q, chci schéma Bearer", got)
			}
		})
	}
}

func TestUserFromEmptyContext(t *testing.T) {
	user, ok := exercise.UserFrom(context.Background())
	if ok {
		t.Errorf("UserFrom(prázdný kontext) = (%q, true), chci false", user)
	}
	if user != "" {
		t.Errorf("UserFrom(prázdný kontext) = %q, chci prázdný řetězec", user)
	}
}

func TestSeriesKey(t *testing.T) {
	tests := []struct {
		name   string
		labels map[string]string
		want   string
	}{
		{"http_requests_total", nil, "http_requests_total"},
		{"http_requests_total", map[string]string{}, "http_requests_total"},
		{"http_requests_total", map[string]string{"method": "GET"},
			`http_requests_total{method="GET"}`},
		{"http_requests_total", map[string]string{"status": "200", "method": "GET", "route": "/items/{id}"},
			`http_requests_total{method="GET",route="/items/{id}",status="200"}`},
	}
	for _, tt := range tests {
		got := exercise.SeriesKey(tt.name, tt.labels)
		if got != tt.want {
			t.Errorf("SeriesKey(%q, %v) = %q, chci %q", tt.name, tt.labels, got, tt.want)
		}
	}

	labels := map[string]string{"a": "1", "z": "26", "m": "13", "b": "2"}
	first := exercise.SeriesKey("x", labels)
	for i := 0; i < 50; i++ {
		if got := exercise.SeriesKey("x", labels); got != first {
			t.Fatalf("SeriesKey není deterministický: %q vs %q", got, first)
		}
	}
}

func TestMetricsIncObserve(t *testing.T) {
	m := exercise.NewMetrics()
	labels := map[string]string{"route": "/items"}
	m.Inc("http_requests_total", labels)
	m.Inc("http_requests_total", labels)
	m.Inc("http_requests_total", map[string]string{"route": "/health"})
	m.Observe("queue_depth", 3)
	m.Observe("queue_depth", 5)
	m.Observe("queue_depth", -2)

	snap := m.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("Snapshot() má %d řad (%v), chci 3", len(snap), snap)
	}

	items := snap[`http_requests_total{route="/items"}`]
	if items.Count != 2 || items.Sum != 2 {
		t.Errorf("čítač /items = %+v, chci Count 2 a Sum 2", items)
	}
	health := snap[`http_requests_total{route="/health"}`]
	if health.Count != 1 {
		t.Errorf("čítač /health = %+v, chci Count 1", health)
	}
	depth := snap["queue_depth"]
	want := exercise.Stat{Count: 3, Sum: 6, Min: -2, Max: 5}
	if depth != want {
		t.Errorf("queue_depth = %+v, chci %+v", depth, want)
	}
}

func TestMetricsFirstObservationSetsMinMax(t *testing.T) {
	m := exercise.NewMetrics()
	m.Observe("latency", 7)
	got := m.Snapshot()["latency"]
	want := exercise.Stat{Count: 1, Sum: 7, Min: 7, Max: 7}
	if got != want {
		t.Errorf("latency = %+v, chci %+v (min ani max nesmí zůstat na nule)", got, want)
	}
}

func TestSnapshotIsCopy(t *testing.T) {
	m := exercise.NewMetrics()
	m.Observe("latency", 1)

	snap := m.Snapshot()
	snap["latency"] = exercise.Stat{Count: 999}
	delete(snap, "neexistuje")
	snap["podvrzeno"] = exercise.Stat{Count: 1}

	again := m.Snapshot()
	if again["latency"].Count != 1 {
		t.Errorf("Snapshot() vrací vnitřní mapu — zápis do ní změnil registr: %+v", again["latency"])
	}
	if _, ok := again["podvrzeno"]; ok {
		t.Error("Snapshot() vrací vnitřní mapu — přibyla podvržená řada")
	}
}

func TestMetricsText(t *testing.T) {
	m := exercise.NewMetrics()
	m.Inc("http_requests_total", map[string]string{"method": "GET"})
	m.Inc("http_requests_total", map[string]string{"method": "GET"})
	m.Observe("queue_depth", 3)
	m.Observe("queue_depth", 5)

	want := "http_requests_total{method=\"GET\"} count=2 sum=2 min=1 max=1\n" +
		"queue_depth count=2 sum=8 min=3 max=5\n"
	got := m.Text()
	if got != want {
		t.Errorf("Text() =\n%q\nchci\n%q", got, want)
	}
	for i := 0; i < 20; i++ {
		if again := m.Text(); again != got {
			t.Fatalf("Text() není deterministický:\n%q\nvs\n%q", again, got)
		}
	}
}

func TestMetricsConcurrent(t *testing.T) {
	const (
		goroutines = 8
		perG       = 500
	)
	m := exercise.NewMetrics()

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				m.Inc("http_requests_total", map[string]string{"method": "GET"})
				m.Observe("http_request_duration_seconds", float64(i%10))
				m.Snapshot()
			}
		}(g)
	}
	wg.Wait()

	snap := m.Snapshot()
	total := snap[`http_requests_total{method="GET"}`]
	if total.Count != goroutines*perG {
		t.Errorf("čítač = %d, chci %d — přišly zápisy", total.Count, goroutines*perG)
	}
	dur := snap["http_request_duration_seconds"]
	if dur.Count != goroutines*perG {
		t.Errorf("pozorování = %d, chci %d", dur.Count, goroutines*perG)
	}
	if dur.Min != 0 || dur.Max != 9 {
		t.Errorf("min/max = %v/%v, chci 0/9", dur.Min, dur.Max)
	}
}

func TestRouteFromEmptyContext(t *testing.T) {
	if got := exercise.RouteFrom(context.Background()); got != exercise.RouteUnknown {
		t.Errorf("RouteFrom(prázdný kontext) = %q, chci %q", got, exercise.RouteUnknown)
	}
}

func TestWithRoutePassesPattern(t *testing.T) {
	var seen string
	h := exercise.WithRoute("/items/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = exercise.RouteFrom(r.Context())
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/items/42", nil))

	if seen != "/items/{id}" {
		t.Errorf("RouteFrom v handleru = %q, chci %q", seen, "/items/{id}")
	}
}

func TestInstrumentCardinality(t *testing.T) {
	m := exercise.NewMetrics()
	item := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.PathValue("id"))
	})

	mux := http.NewServeMux()
	mux.Handle("GET /items/{id}", exercise.WithRoute("/items/{id}", exercise.Instrument(m)(item)))

	for _, path := range []string{"/items/1", "/items/2", "/items/9999"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, chci 200", path, rec.Code)
		}
	}

	snap := m.Snapshot()
	var counters []string
	for key := range snap {
		if strings.HasPrefix(key, "http_requests_total") {
			counters = append(counters, key)
		}
	}
	if len(counters) != 1 {
		t.Fatalf("čítač má %d řad (%v), chci 1 — label route musí být vzor cesty, ne URL", len(counters), counters)
	}

	wantKey := `http_requests_total{method="GET",route="/items/{id}",status="200"}`
	if counters[0] != wantKey {
		t.Errorf("řada = %q, chci %q", counters[0], wantKey)
	}
	if snap[wantKey].Count != 3 {
		t.Errorf("Count = %d, chci 3", snap[wantKey].Count)
	}
	for key := range snap {
		for _, path := range []string{"/items/1", "/items/2", "/items/9999"} {
			if strings.Contains(key, `"`+path+`"`) {
				t.Errorf("řada %q obsahuje konkrétní URL %q — to je exploze kardinality", key, path)
			}
		}
	}

	dur := snap["http_request_duration_seconds"]
	if dur.Count != 3 {
		t.Errorf("http_request_duration_seconds Count = %d, chci 3", dur.Count)
	}
	if dur.Sum < 0 {
		t.Errorf("http_request_duration_seconds Sum = %v, chci nezáporné trvání", dur.Sum)
	}
}

func TestInstrumentStatusLabel(t *testing.T) {
	m := exercise.NewMetrics()
	fail := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusTeapot)
	})
	h := exercise.WithRoute("/brew", exercise.Instrument(m)(fail))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/brew", nil))

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, chci 418 — middleware nesmí odpověď měnit", rec.Code)
	}
	wantKey := `http_requests_total{method="POST",route="/brew",status="418"}`
	if got := m.Snapshot()[wantKey].Count; got != 1 {
		t.Errorf("řada %q má Count %d, chci 1 (mám %v)", wantKey, got, m.Snapshot())
	}
}

func TestInstrumentWithoutPathPattern(t *testing.T) {
	m := exercise.NewMetrics()
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h := exercise.Instrument(m)(ok)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/kdovico", nil))

	wantKey := fmt.Sprintf(`http_requests_total{method="GET",route=%q,status="200"}`, exercise.RouteUnknown)
	if got := m.Snapshot()[wantKey].Count; got != 1 {
		t.Errorf("bez WithRoute chci řadu %q, mám %v", wantKey, m.Snapshot())
	}
}

func TestInstrumentPreservesBody(t *testing.T) {
	m := exercise.NewMetrics()
	h := exercise.Instrument(m)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ahoj")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Body.String(); got != "ahoj" {
		t.Errorf("tělo = %q, chci %q", got, "ahoj")
	}
}
