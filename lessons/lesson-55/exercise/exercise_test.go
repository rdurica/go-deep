package exercise_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-55/exercise"
)

var errKontrola = errors.New("databáze neodpovídá")

func TestBuildInfoString(t *testing.T) {
	tests := []struct {
		in   exercise.BuildInfo
		want string
	}{
		{
			exercise.BuildInfo{Version: "v1.2.3", Commit: "abc1234", BuildTime: "2024-01-15T10:00:00Z"},
			"v1.2.3 (abc1234, 2024-01-15T10:00:00Z)",
		},
		{exercise.BuildInfo{}, "dev (none, unknown)"},
		{exercise.BuildInfo{Version: "v2.0.0"}, "v2.0.0 (none, unknown)"},
	}
	for _, tt := range tests {
		if got := tt.in.String(); got != tt.want {
			t.Errorf("BuildInfo%+v.String() = %q, chci %q", tt.in, got, tt.want)
		}
	}
}

func TestCurrent(t *testing.T) {
	got := exercise.Current()
	want := exercise.BuildInfo{Version: "dev", Commit: "none", BuildTime: "unknown"}
	if got != want {
		t.Errorf("Current() = %+v, chci %+v (bez -ldflags platí výchozí hodnoty)", got, want)
	}
}

func TestVersionHandler(t *testing.T) {
	info := exercise.BuildInfo{Version: "v1.0.0", Commit: "deadbee", BuildTime: "2024-06-01T08:00:00Z"}
	rec := httptest.NewRecorder()
	exercise.VersionHandler(info).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/version", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /version = %d, chci 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, chci application/json", ct)
	}
	var got exercise.BuildInfo
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("tělo není JSON: %v", err)
	}
	if got != info {
		t.Errorf("tělo = %+v, chci %+v", got, info)
	}
}

func TestLiveHandler(t *testing.T) {
	h := exercise.NewHealthChecker(50 * time.Millisecond)
	h.Register("vzdy-selze", func(context.Context) error { return errKontrola })

	rec := httptest.NewRecorder()
	h.LiveHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("GET /healthz = %d, chci 200 — liveness na kontrolách nezávisí", rec.Code)
	}
}

func TestReadyHandlerHealthy(t *testing.T) {
	h := exercise.NewHealthChecker(time.Second)
	var volano sync.Map
	for _, name := range []string{"db", "cache", "queue"} {
		name := name
		h.Register(name, func(ctx context.Context) error {
			volano.Store(name, true)
			return nil
		})
	}

	resp, code := zavolejReady(t, h)
	if code != http.StatusOK {
		t.Fatalf("GET /readyz = %d, chci 200", code)
	}
	if resp.Status != "ok" {
		t.Errorf("status = %q, chci %q", resp.Status, "ok")
	}
	if len(resp.Checks) != 3 {
		t.Fatalf("checks = %v, chci tři položky", resp.Checks)
	}
	for _, name := range []string{"db", "cache", "queue"} {
		if resp.Checks[name] != "ok" {
			t.Errorf("checks[%q] = %q, chci %q", name, resp.Checks[name], "ok")
		}
		if _, ok := volano.Load(name); !ok {
			t.Errorf("kontrola %q se nespustila", name)
		}
	}
}

func TestReadyHandlerNoChecks(t *testing.T) {
	h := exercise.NewHealthChecker(time.Second)
	resp, code := zavolejReady(t, h)
	if code != http.StatusOK {
		t.Errorf("GET /readyz bez kontrol = %d, chci 200", code)
	}
	if len(resp.Checks) != 0 {
		t.Errorf("checks = %v, chci prázdnou mapu", resp.Checks)
	}
}

func TestReadyHandlerOneCheckFails(t *testing.T) {
	h := exercise.NewHealthChecker(time.Second)
	h.Register("db", func(context.Context) error { return errKontrola })
	h.Register("cache", func(context.Context) error { return nil })

	resp, code := zavolejReady(t, h)
	if code != http.StatusServiceUnavailable {
		t.Fatalf("GET /readyz = %d, chci 503", code)
	}
	if resp.Status == "ok" {
		t.Errorf("status = %q, chci jinou hodnotu než ok", resp.Status)
	}
	if resp.Checks["cache"] != "ok" {
		t.Errorf("checks[cache] = %q, chci ok", resp.Checks["cache"])
	}
	if !strings.Contains(resp.Checks["db"], errKontrola.Error()) {
		t.Errorf("checks[db] = %q, chci text chyby %q", resp.Checks["db"], errKontrola)
	}
}

// TestReadyHandlerStuckCheck ověřuje, že se handler nezasekne na kontrole,
// která ignoruje context — musí odpovědět na timeoutu, ne viset.
func TestReadyHandlerStuckCheck(t *testing.T) {
	h := exercise.NewHealthChecker(80 * time.Millisecond)
	uvolnit := make(chan struct{})
	t.Cleanup(func() { close(uvolnit) })

	h.Register("ok", func(context.Context) error { return nil })
	h.Register("zasekla-se", func(context.Context) error {
		<-uvolnit
		return nil
	})

	start := time.Now()
	resp, code := zavolejReady(t, h)
	uplynulo := time.Since(start)

	if uplynulo > 3*time.Second {
		t.Fatalf("handler odpovídal %v, měl spadnout na timeoutu", uplynulo)
	}
	if code != http.StatusServiceUnavailable {
		t.Fatalf("GET /readyz = %d, chci 503", code)
	}
	if resp.Checks["zasekla-se"] == "ok" {
		t.Error("zaseknutá kontrola nesmí být ok")
	}
	if _, ok := resp.Checks["zasekla-se"]; !ok {
		t.Errorf("checks = %v, chci i položku pro zaseknutou kontrolu", resp.Checks)
	}
}

func TestShutdownSequenceOrder(t *testing.T) {
	var mu sync.Mutex
	var poradi []string
	krok := func(name string) exercise.ShutdownStep {
		return exercise.ShutdownStep{Name: name, Fn: func(context.Context) error {
			mu.Lock()
			defer mu.Unlock()
			poradi = append(poradi, name)
			return nil
		}}
	}
	steps := []exercise.ShutdownStep{
		krok("not-ready"),
		krok("drain"),
		krok("server"),
		krok("db"),
	}
	if err := exercise.ShutdownSequence(context.Background(), time.Second, steps); err != nil {
		t.Fatalf("ShutdownSequence = %v, chci nil", err)
	}
	want := []string{"not-ready", "drain", "server", "db"}
	mu.Lock()
	defer mu.Unlock()
	if strings.Join(poradi, ",") != strings.Join(want, ",") {
		t.Errorf("pořadí = %v, chci %v", poradi, want)
	}
}

func TestShutdownSequenceErrorDoesNotStopRest(t *testing.T) {
	var mu sync.Mutex
	var poradi []string
	zapis := func(name string) {
		mu.Lock()
		defer mu.Unlock()
		poradi = append(poradi, name)
	}
	steps := []exercise.ShutdownStep{
		{Name: "server", Fn: func(context.Context) error { zapis("server"); return nil }},
		{Name: "db", Fn: func(context.Context) error { zapis("db"); return errKontrola }},
		{Name: "cache", Fn: func(context.Context) error { zapis("cache"); return nil }},
	}
	err := exercise.ShutdownSequence(context.Background(), time.Second, steps)
	if err == nil {
		t.Fatal("ShutdownSequence = nil, chci chybu z kroku db")
	}
	if !errors.Is(err, errKontrola) {
		t.Errorf("chyba = %v, chci obalený errKontrola", err)
	}
	if !strings.Contains(err.Error(), "db") {
		t.Errorf("chyba = %v, chci v textu jméno kroku", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(poradi) != 3 {
		t.Errorf("proběhlo %v, chci všechny tři kroky i po chybě", poradi)
	}
}

func TestShutdownSequenceWithoutFunc(t *testing.T) {
	err := exercise.ShutdownSequence(context.Background(), time.Second,
		[]exercise.ShutdownStep{{Name: "prazdny"}})
	if !errors.Is(err, exercise.ErrNoStepFunc) {
		t.Errorf("chci ErrNoStepFunc, dostal jsem %v", err)
	}
}

// TestShutdownSequenceLimit ověřuje, že pomalý krok nepřetáhne celkový limit
// a že se zbylé kroky už nespustí.
func TestShutdownSequenceLimit(t *testing.T) {
	var mu sync.Mutex
	spusteno := map[string]bool{}
	zapis := func(name string) {
		mu.Lock()
		defer mu.Unlock()
		spusteno[name] = true
	}
	steps := []exercise.ShutdownStep{
		{Name: "rychly", Fn: func(context.Context) error { zapis("rychly"); return nil }},
		{Name: "pomaly", Fn: func(ctx context.Context) error {
			zapis("pomaly")
			select {
			case <-time.After(5 * time.Second):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}},
		{Name: "nikdy", Fn: func(context.Context) error { zapis("nikdy"); return nil }},
	}

	start := time.Now()
	err := exercise.ShutdownSequence(context.Background(), 100*time.Millisecond, steps)
	uplynulo := time.Since(start)

	if uplynulo > 2*time.Second {
		t.Fatalf("sekvence trvala %v, limit byl 100ms", uplynulo)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("chyba = %v, chci obalený context.DeadlineExceeded", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if !spusteno["rychly"] || !spusteno["pomaly"] {
		t.Errorf("spuštěno %v, chci rychly i pomaly", spusteno)
	}
	if spusteno["nikdy"] {
		t.Error("krok po vypršení limitu se neměl spustit")
	}
}

// zavolejReady zavolá readiness handler a rozparsuje odpověď.
func zavolejReady(t *testing.T, h *exercise.HealthChecker) (exercise.ReadyResponse, int) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ReadyHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	var resp exercise.ReadyResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("tělo /readyz není JSON: %v", err)
	}
	if resp.Checks == nil {
		resp.Checks = map[string]string{}
	}
	return resp, rec.Code
}
