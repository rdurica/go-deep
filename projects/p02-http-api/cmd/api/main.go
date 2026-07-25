// Command api spouští REST API pro správu úkolů.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rdurica/go-deep/projects/p02-http-api/internal/httpapi"
	"github.com/rdurica/go-deep/projects/p02-http-api/internal/task"
)

func main() {
	if err := run(context.Background(), os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

// config je konfigurace služby načtená z prostředí.
type config struct {
	Addr            string
	LogLevel        slog.Level
	ReadTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// loadConfig sestaví konfiguraci a posbírá všechny chyby najednou, aby operátor
// nemusel opravovat proměnné jednu po druhé.
func loadConfig(getenv func(string) string) (config, error) {
	cfg := config{
		Addr:            "0.0.0.0:8080",
		LogLevel:        slog.LevelInfo,
		ReadTimeout:     5 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}

	var errs []error

	if raw := getenv("ADDR"); raw != "" {
		cfg.Addr = raw
	}
	if raw := getenv("PORT"); raw != "" {
		cfg.Addr = net.JoinHostPort("0.0.0.0", raw)
	}
	if raw := getenv("LOG_LEVEL"); raw != "" {
		var level slog.Level
		if err := level.UnmarshalText([]byte(raw)); err != nil {
			errs = append(errs, fmt.Errorf("LOG_LEVEL=%q: %w", raw, err))
		} else {
			cfg.LogLevel = level
		}
	}
	cfg.ReadTimeout = duration(getenv, "READ_TIMEOUT", cfg.ReadTimeout, &errs)
	cfg.ShutdownTimeout = duration(getenv, "SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout, &errs)

	if len(errs) > 0 {
		return config{}, errors.Join(errs...)
	}
	return cfg, nil
}

// duration načte kladnou dobu trvání, nebo zapíše chybu a vrátí výchozí hodnotu.
func duration(getenv func(string) string, key string, def time.Duration, errs *[]error) time.Duration {
	raw := getenv(key)
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	switch {
	case err != nil:
		*errs = append(*errs, fmt.Errorf("%s=%q: %w", key, raw, err))
	case d <= 0:
		*errs = append(*errs, fmt.Errorf("%s=%q must be positive", key, raw))
	default:
		return d
	}
	return def
}

// run sestaví aplikaci a obsluhuje ji, dokud nepřijde SIGINT nebo SIGTERM.
func run(ctx context.Context, getenv func(string) string, out io.Writer) error {
	cfg, err := loadConfig(getenv)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: cfg.LogLevel}))

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", cfg.Addr, err)
	}

	srv := &http.Server{
		Handler:           httpapi.NewRouter(task.NewStore(), logger),
		ReadHeaderTimeout: cfg.ReadTimeout,
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ln)
	}()
	logger.Info("server started", slog.String("addr", ln.Addr().String()))

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received", slog.Duration("grace", cfg.ShutdownTimeout))
	}

	// Vlastní kontext — ten původní je už zrušený signálem a nedal by
	// rozpracovaným požadavkům ani vteřinu.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}
	logger.Info("server stopped")
	return nil
}
