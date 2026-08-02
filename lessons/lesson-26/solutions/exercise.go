// Package solutions obsahuje referenční řešení lekce 26.
package solutions

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// RequestIDHeader je jméno hlavičky, ve které se přenáší ID požadavku.
const RequestIDHeader = "X-Request-ID"

// ErrorResponse je jednotný tvar chybové odpovědi.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Middleware je funkce, která obalí handler jiným handlerem.
type Middleware func(http.Handler) http.Handler

// --- Stupeň: jednoduchý ---
// WriteJSON zapíše v jako JSON odpověď se status kódem status.
// Hotové z lekce 24.
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

// WriteHeader zapamatuje status a přepošle ho podkladovému writeru — jen jednou.
// Druhé a další WriteHeader status nepřepisují (první vyhrává).
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
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()

			next.ServeHTTP(rec, r)

			logger.InfoContext(r.Context(), "request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Int("bytes", rec.bytes),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

// Recovery vrací middleware, který z paniky v handleru udělá 500 s JSON tělem.
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}
				// ErrAbortHandler je smluvený způsob, jak handler ukončí spojení.
				// Nesmíme ho spolknout, patří serveru.
				if recovered == http.ErrAbortHandler {
					panic(recovered)
				}
				_ = WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// requestIDKey je neexportovaný typ klíče, takže do něj nikdo zvenčí nesáhne.
type requestIDKey struct{}

// newRequestID vygeneruje náhodné ID požadavku.
func newRequestID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buf[:])
}

// --- Stupeň: obtížný ---
// RequestID vrací middleware, který doplní ID požadavku do kontextu i do odpovědi.
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := strings.TrimSpace(r.Header.Get(RequestIDHeader))
			if id == "" {
				id = newRequestID()
			}
			w.Header().Set(RequestIDHeader, id)

			ctx := context.WithValue(r.Context(), requestIDKey{}, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDFrom vytáhne ID požadavku z kontextu.
func RequestIDFrom(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey{}).(string)
	return id, ok && id != ""
}

// Timeout vrací middleware, který požadavku nastaví deadline d.
func Timeout(d time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, d, `{"error":"request timeout"}`)
	}
}
