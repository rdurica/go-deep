// Package solutions obsahuje referenční řešení lekce 29.
package solutions

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// ErrEmptyID je chyba pro prázdné ID předané službě.
var ErrEmptyID = errors.New("empty id")

// Redacted je náhrada, kterou RedactingHandler dosadí místo tajemství.
const Redacted = "[REDACTED]"

// secretKeys jsou klíče atributů, jejichž hodnota se nikdy nesmí dostat do logu.
var secretKeys = map[string]bool{
	"password": true,
	"token":    true,
	"api_key":  true,
}

// NewLogger vrátí logger s JSON handlerem, který píše do w a filtruje podle level.
func NewLogger(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
}

// LogRequest zaloguje jeden HTTP požadavek na úrovni Info.
func LogRequest(logger *slog.Logger, method, path string, status int, dur time.Duration) {
	logger.LogAttrs(context.Background(), slog.LevelInfo, "http_request",
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", status),
		slog.Duration("duration", dur),
	)
}

// Service je ukázková služba, která dostává logger konstruktorem.
type Service struct {
	log *slog.Logger
}

// NewService vytvoří službu s odvozeným loggerem.
func NewService(logger *slog.Logger) *Service {
	return &Service{log: logger.With("component", "service")}
}

// Process zpracuje záznam s daným id. Prázdné id je chyba.
func (s *Service) Process(id string) error {
	if id == "" {
		err := ErrEmptyID
		s.log.Error("process failed", slog.String("error", err.Error()))
		return err
	}
	s.log.Info("processed", slog.String("id", id))
	return nil
}

// RedactingHandler obaluje jiný slog.Handler a maskuje citlivé atributy.
type RedactingHandler struct {
	next slog.Handler
}

// NewRedactingHandler vytvoří handler maskující atributy s citlivými klíči.
func NewRedactingHandler(next slog.Handler) *RedactingHandler {
	return &RedactingHandler{next: next}
}

// Enabled implementuje slog.Handler.
func (h *RedactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle implementuje slog.Handler.
func (h *RedactingHandler) Handle(ctx context.Context, r slog.Record) error {
	out := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		out.AddAttrs(redact(a))
		return true
	})
	return h.next.Handle(ctx, out)
}

// WithAttrs implementuje slog.Handler.
func (h *RedactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	safe := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		safe[i] = redact(a)
	}
	return &RedactingHandler{next: h.next.WithAttrs(safe)}
}

// WithGroup implementuje slog.Handler.
func (h *RedactingHandler) WithGroup(name string) slog.Handler {
	return &RedactingHandler{next: h.next.WithGroup(name)}
}

// redact nahradí hodnotu citlivého atributu a rekurzivně projde skupiny.
func redact(a slog.Attr) slog.Attr {
	if secretKeys[strings.ToLower(a.Key)] {
		return slog.String(a.Key, Redacted)
	}
	a.Value = a.Value.Resolve()
	if a.Value.Kind() == slog.KindGroup {
		group := a.Value.Group()
		safe := make([]any, 0, len(group))
		for _, inner := range group {
			safe = append(safe, redact(inner))
		}
		return slog.Group(a.Key, safe...)
	}
	return a
}

// statusRecorder si pamatuje status kód, který handler skutečně poslal.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader zachytí status kód a předá ho dál.
func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Write pokryje handler, který WriteHeader nikdy nezavolá — pak platí 200.
func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

// LoggingMiddleware vrací middleware, který strukturovaně loguje každý požadavek.
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(rec, r)

			if rec.status == 0 {
				rec.status = http.StatusOK
			}
			logger.LogAttrs(r.Context(), slog.LevelInfo, "http_request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
