// Package exercise obsahuje cvičení lekce 55.
package exercise

import (
	"context"
	"errors"
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

// --- Stupeň: jednoduchý ---
// String vrací BuildInfo ve tvaru "verze (commit, čas)".
// Prázdná pole nahradí postupně dev, none a unknown — BuildInfo{}.String() → "dev (none, unknown)".
func (b BuildInfo) String() string {
	// TODO
	return ""
}

// Current vrátí BuildInfo poskládané z package-level proměnných version, commit a buildTime.
// Bez -ldflags platí výchozí hodnoty dev, none a unknown.
func Current() BuildInfo {
	// TODO
	return BuildInfo{}
}

// --- Stupeň: střední ---
// VersionHandler vrací handler s JSON BuildInfo a Content-Type application/json.
// Vždy vrací 200 a serializuje předané info.
func VersionHandler(info BuildInfo) http.Handler {
	// TODO
	return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
}

// NewHealthChecker vytvoří checker s timeoutem pro readiness kontroly.
// Nekladný timeout nahradí jednou vteřinou.
func NewHealthChecker(timeout time.Duration) *HealthChecker {
	// TODO
	return &HealthChecker{checks: make(map[string]Check)}
}

// Register přidá nebo přepíše kontrolu pod daným jménem.
// Nil kontrolu ignoruj; volání je souběžně bezpečné pod mutexem.
func (h *HealthChecker) Register(name string, check Check) {
	// TODO
}

// --- Stupeň: obtížný ---
// LiveHandler vždy vrací 200 a nespouští žádnou registrovanou kontrolu.
// I když je zaregistrovaná kontrola, která vždy selže, liveness vrací 200.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Vrací 503 jako readiness — liveness má být vždy 200.
// Najdi chybu a oprav — testy před opravou padají.
func (h *HealthChecker) LiveHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unhealthy", http.StatusServiceUnavailable)
	})
}

// ReadyHandler spustí kontroly souběžně s timeoutem z r.Context(); visící kontrola
// po vypršení doplní ctx.Err(). Bez kontrol 200 a prázdná mapa; nesmí držet zámek při běhu kontrol.
func (h *HealthChecker) ReadyHandler() http.Handler {
	// TODO
	return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
}

// ShutdownSequence provede kroky v pořadí pod jedním rozpočtem timeout.
// Chyba kroku sekvenci nezastaví; po vypršení rozpočtu další kroky se nespustí.
// Výsledek je errors.Join všech chyb, krok bez Fn je ErrNoStepFunc.
func ShutdownSequence(ctx context.Context, timeout time.Duration, steps []ShutdownStep) error {
	// TODO
	return nil
}
