package solutions_test

import (
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-60/solutions"
)

// clock je deterministický zdroj času bezpečný pro souběžné použití.
type clock struct {
	mu sync.Mutex
	t  time.Time
}

func newClock() *clock {
	return &clock{t: time.Date(2024, time.May, 1, 12, 0, 0, 0, time.UTC)}
}

func (c *clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *clock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

func TestRateLimiterBucket(t *testing.T) {
	c := newClock()
	rl := exercise.NewRateLimiter(3, time.Second, c.Now)

	for i := 0; i < 3; i++ {
		if !rl.Allow("a") {
			t.Fatalf("Allow(a) #%d = false, chci true (kapacita 3)", i+1)
		}
	}
	if rl.Allow("a") {
		t.Error("Allow(a) #4 = true, chci false (kbelík je prázdný)")
	}

	c.Advance(time.Second)
	if !rl.Allow("a") {
		t.Error("Allow(a) po sekundě = false, chci true (doplnil se token)")
	}
	if rl.Allow("a") {
		t.Error("Allow(a) = true, chci false (doplnil se jen jeden token)")
	}

	c.Advance(time.Hour)
	for i := 0; i < 3; i++ {
		if !rl.Allow("a") {
			t.Fatalf("Allow(a) #%d po hodině = false, chci true", i+1)
		}
	}
	if rl.Allow("a") {
		t.Error("Allow(a) = true, chci false — kbelík se nesmí naplnit nad kapacitu")
	}
}

func TestRateLimiterKeysAreIndependent(t *testing.T) {
	c := newClock()
	rl := exercise.NewRateLimiter(1, time.Second, c.Now)

	if !rl.Allow("a") || !rl.Allow("b") {
		t.Fatal("první požadavek každého klíče má projít")
	}
	if rl.Allow("a") || rl.Allow("b") {
		t.Error("druhý požadavek stejného klíče má být odmítnut")
	}
	if got := rl.Len(); got != 2 {
		t.Errorf("Len() = %d, chci 2", got)
	}
}

func TestRateLimiterDefaults(t *testing.T) {
	c := newClock()
	rl := exercise.NewRateLimiter(0, 0, c.Now)
	if !rl.Allow("a") {
		t.Error("Allow() při kapacitě 0 = false, chci aspoň jeden token")
	}
	if rl.Allow("a") {
		t.Error("Allow() podruhé = true, chci false")
	}

	if rl := exercise.NewRateLimiter(1, time.Second, nil); !rl.Allow("a") {
		t.Error("Allow() s výchozími hodinami = false, chci true")
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	c := newClock()
	rl := exercise.NewRateLimiter(5, time.Second, c.Now)

	rl.Allow("stary")
	c.Advance(10 * time.Minute)
	rl.Allow("novy")

	if got := rl.Len(); got != 2 {
		t.Fatalf("Len() = %d, chci 2", got)
	}
	if got := rl.Cleanup(5 * time.Minute); got != 1 {
		t.Errorf("Cleanup(5m) = %d, chci 1", got)
	}
	if got := rl.Len(); got != 1 {
		t.Errorf("Len() po úklidu = %d, chci 1", got)
	}
	if got := rl.Cleanup(5 * time.Minute); got != 0 {
		t.Errorf("druhý Cleanup(5m) = %d, chci 0", got)
	}

	c.Advance(time.Hour)
	if got := rl.Cleanup(time.Minute); got != 1 {
		t.Errorf("Cleanup po hodině = %d, chci 1", got)
	}
	if got := rl.Len(); got != 0 {
		t.Errorf("Len() = %d, chci 0", got)
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	c := newClock()
	const (
		capacity   = 500
		goroutines = 8
		perG       = 100
	)
	rl := exercise.NewRateLimiter(capacity, time.Hour, c.Now)

	var (
		mu      sync.Mutex
		allowed int
		wg      sync.WaitGroup
	)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := 0
			for j := 0; j < perG; j++ {
				if rl.Allow("shared") {
					local++
				}
			}
			mu.Lock()
			allowed += local
			mu.Unlock()
		}()
	}
	wg.Wait()

	if allowed != capacity {
		t.Errorf("prošlo %d požadavků, chci přesně %d (žádný token se nesmí ztratit ani zdvojit)", allowed, capacity)
	}
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ok")
	})
}

func TestHardenPassthrough(t *testing.T) {
	h := exercise.Harden(okHandler(), exercise.HardenOptions{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("kód = %d, chci 200", rec.Code)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Errorf("tělo = %q, chci %q", got, "ok")
	}
}

func TestHardenSecurityHeaders(t *testing.T) {
	h := exercise.Harden(okHandler(), exercise.HardenOptions{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	for name, want := range exercise.SecurityHeaders {
		if got := rec.Header().Get(name); got != want {
			t.Errorf("hlavička %s = %q, chci %q", name, got, want)
		}
	}
}

func TestHardenRateLimit(t *testing.T) {
	c := newClock()
	limiter := exercise.NewRateLimiter(2, time.Second, c.Now)
	h := exercise.Harden(okHandler(), exercise.HardenOptions{Limiter: limiter})

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("požadavek %d = %d, chci 200", i+1, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("třetí požadavek = %d, chci 429", rec.Code)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("odmítnutá odpověď nemá bezpečnostní hlavičky: %q", got)
	}

	c.Advance(time.Second)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("po doplnění tokenu = %d, chci 200", rec.Code)
	}
}

func TestHardenRateLimitKeyFunc(t *testing.T) {
	c := newClock()
	limiter := exercise.NewRateLimiter(1, time.Minute, c.Now)
	h := exercise.Harden(okHandler(), exercise.HardenOptions{
		Limiter: limiter,
		KeyFunc: func(r *http.Request) string { return r.Header.Get("X-Api-Key") },
	})

	call := func(key string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Api-Key", key)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	if got := call("alice"); got != http.StatusOK {
		t.Errorf("alice #1 = %d, chci 200", got)
	}
	if got := call("bob"); got != http.StatusOK {
		t.Errorf("bob #1 = %d, chci 200", got)
	}
	if got := call("alice"); got != http.StatusTooManyRequests {
		t.Errorf("alice #2 = %d, chci 429", got)
	}
}

func TestHardenBodyLimit(t *testing.T) {
	h := exercise.Harden(okHandler(), exercise.HardenOptions{MaxBodyBytes: 16})

	t.Run("body within limit passes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", 16)))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("kód = %d, chci 200", rec.Code)
		}
	})

	t.Run("oversized body returns 413", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", 17)))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("kód = %d, chci 413", rec.Code)
		}
	})

	t.Run("unknown length truncated on read", func(t *testing.T) {
		var readErr error
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, readErr = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		})
		limited := exercise.Harden(inner, exercise.HardenOptions{MaxBodyBytes: 16})

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", 1000)))
		req.ContentLength = -1
		limited.ServeHTTP(httptest.NewRecorder(), req)

		if readErr == nil {
			t.Error("čtení těla nad limit skončilo bez chyby, chci chybu z MaxBytesReader")
		}
	})
}

func TestHardenTimeout(t *testing.T) {
	blocked := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})
	h := exercise.Harden(blocked, exercise.HardenOptions{Timeout: 10 * time.Millisecond})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("kód = %d, chci 503", rec.Code)
	}
	if got := rec.Body.String(); !strings.Contains(got, "timeout") {
		t.Errorf("tělo = %q, chci zprávu o timeoutu", got)
	}
}

func TestHardenRecovery(t *testing.T) {
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("něco se pokazilo")
	})

	options := map[string]exercise.HardenOptions{
		"bez timeoutu": {},
		"s timeoutem":  {Timeout: time.Second, MaxBodyBytes: 1024},
	}
	for name, opts := range options {
		t.Run(name, func(t *testing.T) {
			h := exercise.Harden(panicking, opts)
			rec := httptest.NewRecorder()

			defer func() {
				if p := recover(); p != nil {
					t.Fatalf("panika unikla z Harden: %v", p)
				}
			}()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

			if rec.Code != http.StatusInternalServerError {
				t.Errorf("kód = %d, chci 500", rec.Code)
			}
		})
	}
}

func productionChecks() []exercise.Check {
	return []exercise.Check{
		{ID: "timeouts", Area: "http", Done: true, Critical: true},
		{ID: "graceful-shutdown", Area: "http", Done: true, Critical: true},
		{ID: "rate-limit", Area: "security", Done: true, Critical: false},
		{ID: "body-limit", Area: "security", Done: true, Critical: true},
		{ID: "security-headers", Area: "security", Done: true, Critical: false},
		{ID: "structured-logs", Area: "observability", Done: true, Critical: false},
		{ID: "healthz", Area: "observability", Done: true, Critical: false},
		{ID: "metrics", Area: "observability", Done: true, Critical: false},
		{ID: "no-pii-in-logs", Area: "security", Done: true, Critical: true},
		{ID: "runbook", Area: "provoz", Done: true, Critical: false},
	}
}

func TestAuditReportReady(t *testing.T) {
	got, err := exercise.AuditReport(productionChecks())
	if err != nil {
		t.Fatalf("AuditReport() = chyba %v", err)
	}
	if got.Total != 10 || got.Passed != 10 || got.CriticalFailed != 0 {
		t.Errorf("Report = %+v, chci 10/10 bez kritických nedodělků", got)
	}
	if math.Abs(got.Score-1) > 1e-9 {
		t.Errorf("Score = %v, chci 1", got.Score)
	}
	if !got.Ready {
		t.Error("Ready = false, chci true")
	}
	if len(got.Missing) != 0 {
		t.Errorf("Missing = %v, chci prázdné", got.Missing)
	}
}

func TestAuditReportNotReady(t *testing.T) {
	checks := productionChecks()
	checks[0].Done = false // timeouts, kritické
	checks[6].Done = false // healthz

	got, err := exercise.AuditReport(checks)
	if err != nil {
		t.Fatalf("AuditReport() = chyba %v", err)
	}
	if got.Passed != 8 || got.CriticalFailed != 1 {
		t.Errorf("Report = %+v, chci 8 splněných a 1 kritický nedodělek", got)
	}
	if math.Abs(got.Score-0.8) > 1e-9 {
		t.Errorf("Score = %v, chci 0.8", got.Score)
	}
	if got.Ready {
		t.Error("Ready = true, chci false — chybí kritická kontrola")
	}
	if want := []string{"healthz", "timeouts"}; !reflect.DeepEqual(got.Missing, want) {
		t.Errorf("Missing = %v, chci %v (seřazené)", got.Missing, want)
	}
}

func TestAuditReportScoreThreshold(t *testing.T) {
	checks := productionChecks()
	for i := range checks {
		checks[i].Critical = false
	}
	checks[0].Done = false
	checks[1].Done = false
	checks[2].Done = false

	got, err := exercise.AuditReport(checks)
	if err != nil {
		t.Fatalf("AuditReport() = chyba %v", err)
	}
	if got.Ready {
		t.Errorf("Ready = true při skóre %v, chci false (práh je 0.8)", got.Score)
	}
}

func TestAuditReportErrors(t *testing.T) {
	if _, err := exercise.AuditReport(nil); !errors.Is(err, exercise.ErrNoChecks) {
		t.Errorf("AuditReport(nil) = %v, chci ErrNoChecks", err)
	}
	dup := []exercise.Check{{ID: "a", Done: true}, {ID: "a"}}
	if _, err := exercise.AuditReport(dup); !errors.Is(err, exercise.ErrDuplicateCheck) {
		t.Errorf("AuditReport(duplicita) = %v, chci ErrDuplicateCheck", err)
	}
	empty := []exercise.Check{{ID: "  ", Done: true}}
	if _, err := exercise.AuditReport(empty); !errors.Is(err, exercise.ErrInvalidCheck) {
		t.Errorf("AuditReport(bez ID) = %v, chci ErrInvalidCheck", err)
	}
}

func TestCoverage(t *testing.T) {
	lessons := []exercise.LessonResult{
		{Lesson: 1, Phase: 0, Passed: true},
		{Lesson: 2, Phase: 0, Passed: true},
		{Lesson: 3, Phase: 1, Passed: true},
		{Lesson: 4, Phase: 1, Passed: false},
		{Lesson: 5, Phase: 2, Passed: false},
		{Lesson: 6, Phase: 2, Passed: false},
		{Lesson: 7, Phase: 3, Passed: true},
	}

	got := exercise.Coverage(lessons)
	if got.Total != 7 || got.Passed != 4 {
		t.Errorf("Coverage() = %d/%d, chci 4/7", got.Passed, got.Total)
	}
	if math.Abs(got.Percent-400.0/7.0) > 1e-9 {
		t.Errorf("Percent = %v, chci %v", got.Percent, 400.0/7.0)
	}
	wantPhases := map[int]exercise.PhaseStat{
		0: {Total: 2, Passed: 2},
		1: {Total: 2, Passed: 1},
		2: {Total: 2, Passed: 0},
		3: {Total: 1, Passed: 1},
	}
	if !reflect.DeepEqual(got.ByPhase, wantPhases) {
		t.Errorf("ByPhase = %+v, chci %+v", got.ByPhase, wantPhases)
	}
	if got.WeakestPhase != 2 {
		t.Errorf("WeakestPhase = %d, chci 2", got.WeakestPhase)
	}
}

func TestCoverageEdgeCases(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		got := exercise.Coverage(nil)
		if got.Total != 0 || got.Passed != 0 || got.Percent != 0 || got.WeakestPhase != 0 {
			t.Errorf("Coverage(nil) = %+v, chci nuly", got)
		}
		if got.ByPhase == nil {
			t.Error("ByPhase = nil, chci inicializovanou mapu")
		}
	})

	t.Run("weakest phase match wins lower number", func(t *testing.T) {
		got := exercise.Coverage([]exercise.LessonResult{
			{Lesson: 1, Phase: 5, Passed: false},
			{Lesson: 2, Phase: 3, Passed: false},
			{Lesson: 3, Phase: 1, Passed: true},
		})
		if got.WeakestPhase != 3 {
			t.Errorf("WeakestPhase = %d, chci 3", got.WeakestPhase)
		}
	})

	t.Run("all done", func(t *testing.T) {
		got := exercise.Coverage([]exercise.LessonResult{
			{Lesson: 1, Phase: 1, Passed: true},
			{Lesson: 2, Phase: 2, Passed: true},
		})
		if math.Abs(got.Percent-100) > 1e-9 {
			t.Errorf("Percent = %v, chci 100", got.Percent)
		}
		if got.WeakestPhase != 1 {
			t.Errorf("WeakestPhase = %d, chci 1 (při shodě nejnižší fáze)", got.WeakestPhase)
		}
	})
}
