package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// RequestIDHeader je hlavička, ve které se přenáší identifikátor požadavku.
const RequestIDHeader = "X-Request-ID"

// Middleware je obalovač handleru.
type Middleware func(http.Handler) http.Handler

// Chain složí middleware kolem handleru; první uvedená je nejvíc vně.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// ctxKey je privátní typ klíče, aby hodnotu v kontextu nemohl přepsat cizí balíček.
type ctxKey struct{}

var requestIDKey ctxKey

// RequestIDFromContext vrátí identifikátor požadavku uložený middlewarem.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok && id != ""
}

// RequestID doplní každému požadavku identifikátor do kontextu i do odpovědi.
// Příchozí hlavičku respektuje, aby šlo trasovat napříč službami.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set(RequestIDHeader, id)

		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// newRequestID vyrobí náhodný identifikátor požadavku.
func newRequestID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b[:])
}

// statusRecorder si pamatuje status kód, který handler skutečně poslal.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Logging strukturovaně zaloguje každý dokončený požadavek.
func Logging(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(rec, r)

			if rec.status == 0 {
				rec.status = http.StatusOK
			}
			requestID, _ := RequestIDFromContext(r.Context())
			logger.LogAttrs(r.Context(), levelFor(rec.status), "http_request",
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Int("bytes", rec.bytes),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

// levelFor vybere úroveň logu podle status kódu.
func levelFor(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

// Recovery zachytí paniku v handleru, zaloguje ji a vrátí 500 v JSONu.
func Recovery(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				requestID, _ := RequestIDFromContext(r.Context())
				logger.LogAttrs(r.Context(), slog.LevelError, "panic recovered",
					slog.String("request_id", requestID),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("panic", fmt.Sprint(rec)),
				)
				writeError(w, http.StatusInternalServerError, CodeInternalError, "internal server error")
			}()
			next.ServeHTTP(w, r)
		})
	}
}
