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
	panic("TODO: úkol A")
}

// Chain složí middleware kolem handleru; první uvedená je nejvíc vně.
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	panic("TODO: úkol B")
}

// RequestIDMiddleware zajistí, že každý požadavek má identifikátor v kontextu i hlavičce.
func RequestIDMiddleware(next http.Handler) http.Handler {
	panic("TODO: úkol B")
}

// RequestIDFromContext vytáhne identifikátor požadavku z kontextu.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	panic("TODO: úkol B")
}

// NewServer sestaví router s middleware chainem a in-memory úložištěm poznámek.
func NewServer(logger *slog.Logger) http.Handler {
	panic("TODO: úkol B")
}

// RecoveryMiddleware zachytí paniku v handleru, zaloguje ji a vrátí 500 v JSONu.
func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	panic("TODO: úkol C")
}

// Run obsluhuje ln, dokud se nezruší ctx, pak server elegantně ukončí.
func Run(ctx context.Context, cfg Config, h http.Handler, ln net.Listener) error {
	panic("TODO: úkol C")
}
