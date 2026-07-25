// Package exercise obsahuje cvičení lekce 29.
package exercise

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// ErrEmptyID je chyba pro prázdné ID předané službě.
var ErrEmptyID = errors.New("empty id")

// Redacted je náhrada, kterou RedactingHandler dosadí místo tajemství.
const Redacted = "[REDACTED]"

// NewLogger vrátí logger s JSON handlerem, který píše do w a filtruje podle level.
func NewLogger(w io.Writer, level slog.Level) *slog.Logger {
	// TODO: úkol A
	return nil
}

// LogRequest zaloguje jeden HTTP požadavek na úrovni Info.
func LogRequest(logger *slog.Logger, method, path string, status int, dur time.Duration) {
	// TODO: úkol A
}

// Service je ukázková služba, která dostává logger konstruktorem.
type Service struct {
	log *slog.Logger
}

// NewService vytvoří službu s odvozeným loggerem.
func NewService(logger *slog.Logger) *Service {
	// TODO: úkol B
	return nil
}

// Process zpracuje záznam s daným id. Prázdné id je chyba.
func (s *Service) Process(id string) error {
	// TODO: úkol B
	return nil
}

// RedactingHandler obaluje jiný slog.Handler a maskuje citlivé atributy.
type RedactingHandler struct {
	next slog.Handler
}

// NewRedactingHandler vytvoří handler maskující atributy s citlivými klíči.
func NewRedactingHandler(next slog.Handler) *RedactingHandler {
	// TODO: úkol B
	return nil
}

// Enabled implementuje slog.Handler.
func (h *RedactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// TODO: úkol B
	return false
}

// Handle implementuje slog.Handler.
func (h *RedactingHandler) Handle(ctx context.Context, r slog.Record) error {
	// TODO: úkol B
	return nil
}

// WithAttrs implementuje slog.Handler.
func (h *RedactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// TODO: úkol B
	return *new(slog.Handler)
}

// WithGroup implementuje slog.Handler.
func (h *RedactingHandler) WithGroup(name string) slog.Handler {
	// TODO: úkol B
	return *new(slog.Handler)
}

// LoggingMiddleware vrací middleware, který strukturovaně loguje každý požadavek.
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	// TODO: úkol C
	return nil
}
