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

// String vrací "verze (commit, čas)"; prázdná pole nahradí výchozími hodnotami.
func (b BuildInfo) String() string {
	panic("TODO: úkol A")
}

// Current vrátí BuildInfo poskládané z ldflags proměnných.
func Current() BuildInfo {
	panic("TODO: úkol A")
}

// VersionHandler vrací JSON s informacemi o buildu.
func VersionHandler(info BuildInfo) http.Handler {
	panic("TODO: úkol A")
}

// NewHealthChecker vytvoří checker s daným timeoutem pro readiness kontroly.
func NewHealthChecker(timeout time.Duration) *HealthChecker {
	panic("TODO: úkol B")
}

// Register přidá nebo přepíše pojmenovanou kontrolu.
func (h *HealthChecker) Register(name string, check Check) {
	panic("TODO: úkol B")
}

// LiveHandler odpovídá 200, dokud proces žije.
func (h *HealthChecker) LiveHandler() http.Handler {
	panic("TODO: úkol B")
}

// ReadyHandler spustí kontroly souběžně a vrátí 200, nebo 503 s detailem.
func (h *HealthChecker) ReadyHandler() http.Handler {
	panic("TODO: úkol B")
}

// ShutdownSequence provede kroky v pořadí a vrátí souhrnnou chybu.
func ShutdownSequence(ctx context.Context, timeout time.Duration, steps []ShutdownStep) error {
	panic("TODO: úkol C")
}
