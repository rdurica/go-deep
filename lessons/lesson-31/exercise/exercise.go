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

// Note je jedna poznámka uložená v paměti.
type Note struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// LoadConfig sestaví Config z getenv a posbírá všechny chyby najednou.
func LoadConfig(getenv func(string) string) (Config, error) {
	// TODO: úkol A
	return *new(Config), nil
}

// Chain složí middleware kolem handleru; první uvedená je nejvíc vně.
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	// TODO: úkol B
	return *new(http.Handler)
}

// RequestIDMiddleware zajistí, že každý požadavek má identifikátor v kontextu i hlavičce.
func RequestIDMiddleware(next http.Handler) http.Handler {
	// TODO: úkol B
	return *new(http.Handler)
}

// RequestIDFromContext vytáhne identifikátor požadavku z kontextu.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	// TODO: úkol B
	return "", false
}

// NewServer sestaví router s middleware chainem a in-memory úložištěm poznámek.
func NewServer(logger *slog.Logger) http.Handler {
	// TODO: úkol B
	return *new(http.Handler)
}

// RecoveryMiddleware zachytí paniku v handleru, zaloguje ji a vrátí 500 v JSONu.
func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	// TODO: úkol C
	return nil
}

// Run obsluhuje ln, dokud se nezruší ctx, pak server elegantně ukončí.
func Run(ctx context.Context, cfg Config, h http.Handler, ln net.Listener) error {
	// TODO: úkol C
	return nil
}
