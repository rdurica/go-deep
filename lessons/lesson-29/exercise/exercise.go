// Package exercise obsahuje cvičení lekce 29.
package exercise

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// ErrEmptyID je chyba pro prázdné ID předané službě.
var ErrEmptyID = errors.New("empty id")

// Redacted je náhrada, kterou RedactingHandler dosadí místo tajemství.
const Redacted = "[REDACTED]"

// --- Stupeň: jednoduchý ---

// LogRequest zaloguje jeden HTTP požadavek na úrovni Info.
// Zpráva "http_request"; atributy method, path, status (int), duration (slog.Duration) přes LogAttrs.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Proměnné hodnoty jsou ve zprávě místo v atributech.
// Najdi chybu a oprav — testy před opravou padají.
func LogRequest(logger *slog.Logger, method, path string, status int, dur time.Duration) {
	logger.Info(fmt.Sprintf("http_request %s %s status=%d", method, path, status))
}

// --- Stupeň: střední ---

// NewLogger vrátí logger s JSON handlerem, který píše do w a filtruje podle level.
func NewLogger(w io.Writer, level slog.Level) *slog.Logger {
	// TODO
	return nil
}

// Service je ukázková služba, která dostává logger konstruktorem.
type Service struct {
	log *slog.Logger
}

// NewService vytvoří službu s odvozeným loggerem.
// Logger s atributem component=service přes With.
func NewService(logger *slog.Logger) *Service {
	// TODO
	return nil
}

// Process zpracuje záznam s daným id. Prázdné id je chyba.
// Prázdné id → Error "process failed" + atribut error, vrátí ErrEmptyID.
// Neprázdné → Info "processed" + atribut id, vrátí nil.
func (s *Service) Process(id string) error {
	// TODO
	return nil
}

// RedactingHandler obaluje jiný slog.Handler a maskuje citlivé atributy.
type RedactingHandler struct {
	next slog.Handler
}

// NewRedactingHandler vytvoří handler maskující atributy s citlivými klíči.
// Maskuj password, token, api_key (case-insensitive) konstantou Redacted.
func NewRedactingHandler(next slog.Handler) *RedactingHandler {
	// TODO
	return nil
}

// Enabled implementuje slog.Handler.
// Deleguje na obalovaný handler.
func (h *RedactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// TODO
	return false
}

// --- Stupeň: obtížný ---

// Handle implementuje slog.Handler.
// Maskuj klíče password, token, api_key (case-insensitive) včetně skupin rekurzivně.
func (h *RedactingHandler) Handle(ctx context.Context, r slog.Record) error {
	// TODO
	return nil
}

// WithAttrs implementuje slog.Handler.
// Vrací nový handler; nemutuje původní.
// Attrs před předáním dál zaredactuj; výsledek zůstane RedactingHandler (maskování platí dál).
func (h *RedactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// TODO
	return *new(slog.Handler)
}

// WithGroup implementuje slog.Handler.
// Deleguje WithGroup na next, ale výsledek znovu zabal — maskování platí i ve skupině.
func (h *RedactingHandler) WithGroup(name string) slog.Handler {
	// TODO
	return *new(slog.Handler)
}

// LoggingMiddleware vrací middleware, který strukturovaně loguje každý požadavek.
// Po dokončení "http_request" s method, path, status, duration; skutečný odeslaný status.
// Odpověď nesmí změnit.
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	// TODO
	return nil
}
