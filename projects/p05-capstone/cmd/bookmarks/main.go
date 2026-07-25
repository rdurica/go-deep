// Command bookmarks spouští capstone službu záložek URL.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rdurica/go-deep/projects/p05-capstone/internal/config"
	"github.com/rdurica/go-deep/projects/p05-capstone/internal/httpapi"
	"github.com/rdurica/go-deep/projects/p05-capstone/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger, os.Getenv); err != nil {
		logger.Error("služba skončila chybou", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger, getenv func(string) string) error {
	cfg, err := config.Load(getenv)
	if err != nil {
		return fmt.Errorf("konfigurace: %w", err)
	}

	st := store.New()
	handler := httpapi.NewServer(st, httpapi.Options{
		MaxBodyBytes: cfg.MaxBodyBytes,
		Timeout:      cfg.RequestTimeout,
	})

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() {
		logger.Info("server startuje", slog.String("addr", cfg.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errc <- err
			return
		}
		errc <- nil
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		logger.Info("dostal jsem signál, ukončuji", slog.Duration("timeout", cfg.ShutdownTimeout))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	if err := <-errc; err != nil {
		return err
	}
	logger.Info("server ukončen")
	return nil
}
