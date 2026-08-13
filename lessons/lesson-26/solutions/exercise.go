// Package solutions obsahuje referenční řešení lekce 26.
package solutions

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorResponse je jednotný tvar chybové odpovědi.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Middleware je funkce, která obalí handler jiným handlerem.
type Middleware func(http.Handler) http.Handler

// --- Stupeň: jednoduchý ---

// WriteJSON zapíše v jako JSON odpověď se status kódem status.
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// statusRecorder obaluje http.ResponseWriter a pamatuje si status kód
// a počet zapsaných bajtů, aby je middleware mohl zalogovat.
type statusRecorder struct {
	http.ResponseWriter

	status      int
	bytes       int
	wroteHeader bool
}

// NewStatusRecorder obalí w pro zachycení statusu a velikosti odpovědi.
func NewStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{ResponseWriter: w, status: http.StatusOK}
}

// RecordedStatus vrátí zapamatovaný HTTP status.
func (rec *statusRecorder) RecordedStatus() int {
	return rec.status
}

// RecordedBytes vrátí počet bajtů zapsaných přes Write.
func (rec *statusRecorder) RecordedBytes() int {
	return rec.bytes
}

// WriteHeader zapamatuje status a přepošle ho podkladovému writeru — jen jednou.
func (rec *statusRecorder) WriteHeader(status int) {
	if rec.wroteHeader {
		return
	}
	rec.wroteHeader = true
	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
}

// Write přepošle data a přičte jejich délku k počítadlu.
func (rec *statusRecorder) Write(b []byte) (int, error) {
	if !rec.wroteHeader {
		rec.WriteHeader(http.StatusOK)
	}
	n, err := rec.ResponseWriter.Write(b)
	rec.bytes += n
	return n, err
}

// --- Stupeň: střední ---

// Chain obalí h middlewary tak, že první uvedený je nejvíc vnější.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// Logging vrací middleware, který po dokončení požadavku zaloguje
// metodu, cestu, status a velikost odpovědi.
func Logging(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := NewStatusRecorder(w)

			next.ServeHTTP(rec, r)

			logger.Info("request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.RecordedStatus()),
				slog.Int("bytes", rec.RecordedBytes()),
			)
		})
	}
}

// --- Stupeň: obtížný ---

// Recovery vrací middleware, který z paniky v handleru udělá 500 s JSON tělem.
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}
				if recovered == http.ErrAbortHandler {
					panic(recovered)
				}
				_ = WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
			}()

			next.ServeHTTP(w, r)
		})
	}
}
