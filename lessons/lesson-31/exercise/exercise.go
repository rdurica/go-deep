// Package exercise obsahuje kumulativní cvičení checkpointu fáze 3 (lekce 31).
package exercise

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// RequestIDHeader je hlavička, ve které se přenáší identifikátor požadavku.
const RequestIDHeader = "X-Request-ID"

// ErrInvalid označuje hodnotu konfigurace, kterou se nepodařilo zpracovat.
var ErrInvalid = errors.New("invalid value")

// Config je konfigurace HTTP služby načtená z prostředí.
type Config struct {
	Addr            string
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
}

// --- Stupeň: jednoduchý ---

// LoadConfig sestaví Config z getenv a posbírá všechny chyby přes errors.Join.
// ADDR výchozí 127.0.0.1:8080, LOG_LEVEL přes slog.Level.UnmarshalText,
// SHUTDOWN_TIMEOUT přes time.ParseDuration (musí být kladný).
// Chyby obal ErrInvalid; text musí obsahovat jména vadných klíčů.
func LoadConfig(getenv func(string) string) (Config, error) {
	// TODO
	return *new(Config), nil
}

// Chain složí middleware kolem handleru; první uvedená je nejvíc vně.
// Bez middleware vrátí h beze změny.
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	// TODO
	return *new(http.Handler)
}

// HealthHandler vrací hotový health endpoint pro testy Run — neměň.
func HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
}

// --- Stupeň: střední ---

// RequestIDMiddleware vezme X-Request-ID z požadavku, chybějící vygeneruje,
// uloží do kontextu (neexportovaný typ klíče) a nastaví do hlavičky odpovědi.
func RequestIDMiddleware(next http.Handler) http.Handler {
	// TODO
	return *new(http.Handler)
}

// RequestIDFromContext vrátí identifikátor z kontextu.
// Pro kontext bez ID vrátí false (bez paniky).
func RequestIDFromContext(ctx context.Context) (string, bool) {
	// TODO
	return "", false
}

// --- Stupeň: obtížný ---

// RecoveryMiddleware zachytí paniku, zaloguje Error s request_id z kontextu
// a odpoví 500 v JSONu {"error":{"code":"internal_error","message":"..."}}.
// http.ErrAbortHandler propusť dál (panic).
func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	// TODO
	return nil
}

// Run obsluhuje ln; na zrušení ctx provede Shutdown s cfg.ShutdownTimeout.
// Při čistém ukončení vrátí nil (ErrServerClosed není chyba).
func Run(ctx context.Context, cfg Config, h http.Handler, ln net.Listener) error {
	// TODO
	return nil
}
