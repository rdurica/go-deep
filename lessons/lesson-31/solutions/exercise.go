// Package solutions obsahuje referenční řešení checkpointu fáze 3 (lekce 31).
package solutions

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

// RequestIDHeader je hlavička, ve které se přenáší identifikátor požadavku.
const RequestIDHeader = "X-Request-ID"

// ErrInvalid označuje hodnotu konfigurace, kterou se nepodařilo zpracovat.
var ErrInvalid = errors.New("invalid value")

const defaultAddr = "127.0.0.1:8080"

// Config je konfigurace HTTP služby načtená z prostředí.
type Config struct {
	Addr            string
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
}

// LoadConfig sestaví Config z getenv a posbírá všechny chyby přes errors.Join.
func LoadConfig(getenv func(string) string) (Config, error) {
	var errs []error

	cfg := Config{
		Addr:            defaultAddr,
		LogLevel:        slog.LevelInfo,
		ShutdownTimeout: 5 * time.Second,
	}

	if raw := getenv("ADDR"); raw != "" {
		cfg.Addr = raw
	}

	if raw := getenv("LOG_LEVEL"); raw != "" {
		var level slog.Level
		if err := level.UnmarshalText([]byte(raw)); err != nil {
			errs = append(errs, fmt.Errorf("LOG_LEVEL=%q: %w", raw, ErrInvalid))
		} else {
			cfg.LogLevel = level
		}
	}

	if raw := getenv("SHUTDOWN_TIMEOUT"); raw != "" {
		d, err := time.ParseDuration(raw)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("SHUTDOWN_TIMEOUT=%q: %w", raw, ErrInvalid))
		case d <= 0:
			errs = append(errs, fmt.Errorf("SHUTDOWN_TIMEOUT=%q musí být kladný: %w", raw, ErrInvalid))
		default:
			cfg.ShutdownTimeout = d
		}
	}

	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}
	return cfg, nil
}

// Chain složí middleware kolem handleru; první uvedená je nejvíc vně.
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// HealthHandler vrací hotový health endpoint pro testy Run.
func HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
}

type ctxKey struct{}

var requestIDKey ctxKey

// RequestIDMiddleware zajistí identifikátor požadavku v kontextu i hlavičce.
func RequestIDMiddleware(next http.Handler) http.Handler {
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

// RequestIDFromContext vytáhne identifikátor požadavku z kontextu.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok && id != ""
}

func newRequestID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b[:])
}

func writeError(w http.ResponseWriter, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

// RecoveryMiddleware zachytí paniku v handleru, zaloguje ji a vrátí 500 v JSONu.
func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				if rec == http.ErrAbortHandler {
					panic(rec)
				}
				requestID, _ := RequestIDFromContext(r.Context())
				logger.LogAttrs(r.Context(), slog.LevelError, "panic recovered",
					slog.String("request_id", requestID),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("panic", fmt.Sprint(rec)),
				)
				writeError(w, "internal_error", "internal server error")
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Run obsluhuje ln; na zrušení ctx provede Shutdown s cfg.ShutdownTimeout.
func Run(ctx context.Context, cfg Config, h http.Handler, ln net.Listener) error {
	srv := &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ln)
	}()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
