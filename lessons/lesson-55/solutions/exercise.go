// Package solutions obsahuje referenční řešení lekce 55.
package solutions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Hodnoty vkládané při buildu přes -ldflags "-X ...". Bez nich platí tyhle výchozí.
var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

// ErrNoStepFunc označuje krok ukončovací sekvence bez funkce.
var ErrNoStepFunc = errors.New("krok nemá funkci")

// BuildInfo popisuje sestavenou binárku.
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

// Check je jedna pojmenovaná kontrola připravenosti.
type Check func(ctx context.Context) error

// ReadyResponse je tělo odpovědi readiness probe.
type ReadyResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// HealthChecker drží registrované kontroly připravenosti.
type HealthChecker struct {
	timeout time.Duration
	mu      sync.Mutex
	order   []string
	checks  map[string]Check
}

// ShutdownStep je jeden pojmenovaný krok ukončovací sekvence.
type ShutdownStep struct {
	Name string
	Fn   func(ctx context.Context) error
}

// statusOK je hodnota, kterou v odpovědi nese úspěšná kontrola.
const statusOK = "ok"

// String vrací "verze (commit, čas)"; prázdná pole nahradí výchozími hodnotami.
func (b BuildInfo) String() string {
	return fmt.Sprintf("%s (%s, %s)",
		orDefault(b.Version, "dev"),
		orDefault(b.Commit, "none"),
		orDefault(b.BuildTime, "unknown"),
	)
}

// Current vrátí BuildInfo poskládané z ldflags proměnných.
func Current() BuildInfo {
	return BuildInfo{Version: version, Commit: commit, BuildTime: buildTime}
}

// VersionHandler vrací JSON s informacemi o buildu.
func VersionHandler(info BuildInfo) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, info)
	})
}

// NewHealthChecker vytvoří checker s daným timeoutem pro readiness kontroly.
func NewHealthChecker(timeout time.Duration) *HealthChecker {
	if timeout <= 0 {
		timeout = time.Second
	}
	return &HealthChecker{timeout: timeout, checks: make(map[string]Check)}
}

// Register přidá nebo přepíše pojmenovanou kontrolu.
func (h *HealthChecker) Register(name string, check Check) {
	if check == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.checks[name]; !ok {
		h.order = append(h.order, name)
	}
	h.checks[name] = check
}

// LiveHandler odpovídá 200, dokud proces žije.
func (h *HealthChecker) LiveHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(statusOK))
	})
}

// ReadyHandler spustí kontroly souběžně a vrátí 200, nebo 503 s detailem.
func (h *HealthChecker) ReadyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
		defer cancel()

		names, checks := h.snapshot()
		resp := ReadyResponse{Status: statusOK, Checks: make(map[string]string, len(names))}

		type result struct{ name, msg string }
		// Buffer na všechny výsledky: goroutina zaseknuté kontroly se nikdy
		// nezablokuje na zápisu, takže po jejím doběhnutí zmizí.
		results := make(chan result, len(names))
		for _, name := range names {
			go func(name string, check Check) {
				if err := check(ctx); err != nil {
					results <- result{name, err.Error()}
					return
				}
				results <- result{name, statusOK}
			}(name, checks[name])
		}

		for remaining := len(names); remaining > 0; {
			select {
			case res := <-results:
				resp.Checks[res.name] = res.msg
				if res.msg != statusOK {
					resp.Status = "fail"
				}
				remaining--
			case <-ctx.Done():
				for _, name := range names {
					if _, ok := resp.Checks[name]; !ok {
						resp.Checks[name] = ctx.Err().Error()
					}
				}
				resp.Status = "fail"
				remaining = 0
			}
		}

		code := http.StatusOK
		if resp.Status != statusOK {
			code = http.StatusServiceUnavailable
		}
		writeJSON(w, code, resp)
	})
}

// ShutdownSequence provede kroky v pořadí a vrátí souhrnnou chybu.
func ShutdownSequence(ctx context.Context, timeout time.Duration, steps []ShutdownStep) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var errs []error
	for _, step := range steps {
		if err := ctx.Err(); err != nil {
			errs = append(errs, fmt.Errorf("krok %q nespuštěn: %w", step.Name, err))
			continue
		}
		if step.Fn == nil {
			errs = append(errs, fmt.Errorf("krok %q: %w", step.Name, ErrNoStepFunc))
			continue
		}
		done := make(chan error, 1)
		go func(fn func(context.Context) error) { done <- fn(ctx) }(step.Fn)
		select {
		case err := <-done:
			if err != nil {
				errs = append(errs, fmt.Errorf("krok %q: %w", step.Name, err))
			}
		case <-ctx.Done():
			errs = append(errs, fmt.Errorf("krok %q nedoběhl: %w", step.Name, ctx.Err()))
		}
	}
	return errors.Join(errs...)
}

// snapshot vrátí kopii registrovaných kontrol, aby se handler nedržel zámku.
func (h *HealthChecker) snapshot() ([]string, map[string]Check) {
	h.mu.Lock()
	defer h.mu.Unlock()
	names := make([]string, len(h.order))
	copy(names, h.order)
	checks := make(map[string]Check, len(h.checks))
	for k, v := range h.checks {
		checks[k] = v
	}
	return names, checks
}

// orDefault vrátí s, nebo náhradu, pokud je s prázdné.
func orDefault(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// writeJSON zapíše hodnotu jako JSON s daným stavovým kódem.
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
