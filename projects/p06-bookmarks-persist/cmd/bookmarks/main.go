// Command bookmarks spouští P06 službu záložek s Postgres a Redis.
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

	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/app"
	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/config"
	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/httpapi"
	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/postgres"
	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/rediscache"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pg, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pg.Close()

	if err := pg.ApplyMigrations(ctx); err != nil {
		return err
	}

	cache, err := rediscache.Open(ctx, cfg.RedisURL, cfg.CacheTTL)
	if err != nil {
		return err
	}
	defer func() { _ = cache.Close() }()

	store := app.CachedStore{Store: pg, Cache: cache}
	handler := httpapi.NewServer(store, httpapi.Options{
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

	errc := make(chan error, 1)
	go func() {
		logger.Info("server startuje",
			slog.String("addr", cfg.Addr),
			slog.String("cache_ttl", cfg.CacheTTL.String()),
		)
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
