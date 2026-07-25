// Package config načítá konfiguraci služby z prostředí.
package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Config je provozní konfigurace HTTP služby s Postgres a Redis.
type Config struct {
	Addr            string
	DatabaseURL     string
	RedisURL        string
	CacheTTL        time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	RequestTimeout  time.Duration
	MaxBodyBytes    int64
}

// Load sestaví konfiguraci z getenv. Prázdné hodnoty nahradí výchozími.
// DATABASE_URL je povinná.
func Load(getenv func(string) string) (Config, error) {
	if getenv == nil {
		return Config{}, errors.New("config: getenv je nil")
	}

	cfg := Config{
		Addr:            ":8080",
		RedisURL:        "redis://127.0.0.1:6379/0",
		CacheTTL:        5 * time.Minute,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		RequestTimeout:  3 * time.Second,
		MaxBodyBytes:    1 << 20,
	}

	var errs []error
	if v := strings.TrimSpace(getenv("BOOKMARKS_ADDR")); v != "" {
		cfg.Addr = v
	}
	cfg.DatabaseURL = strings.TrimSpace(getenv("DATABASE_URL"))
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL je povinná"))
	}
	if v := strings.TrimSpace(getenv("REDIS_URL")); v != "" {
		cfg.RedisURL = v
	}
	cfg.CacheTTL = duration(getenv, "BOOKMARKS_CACHE_TTL", cfg.CacheTTL, &errs)
	cfg.ReadTimeout = duration(getenv, "BOOKMARKS_READ_TIMEOUT", cfg.ReadTimeout, &errs)
	cfg.WriteTimeout = duration(getenv, "BOOKMARKS_WRITE_TIMEOUT", cfg.WriteTimeout, &errs)
	cfg.IdleTimeout = duration(getenv, "BOOKMARKS_IDLE_TIMEOUT", cfg.IdleTimeout, &errs)
	cfg.ShutdownTimeout = duration(getenv, "BOOKMARKS_SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout, &errs)
	cfg.RequestTimeout = duration(getenv, "BOOKMARKS_REQUEST_TIMEOUT", cfg.RequestTimeout, &errs)
	cfg.MaxBodyBytes = int64Value(getenv, "BOOKMARKS_MAX_BODY_BYTES", cfg.MaxBodyBytes, &errs)

	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}
	return cfg, nil
}

func duration(getenv func(string) string, key string, def time.Duration, errs *[]error) time.Duration {
	raw := strings.TrimSpace(getenv(key))
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	switch {
	case err != nil:
		*errs = append(*errs, fmt.Errorf("%s=%q: %w", key, raw, err))
	case d <= 0:
		*errs = append(*errs, fmt.Errorf("%s=%q: musí být kladné", key, raw))
	default:
		return d
	}
	return def
}

func int64Value(getenv func(string) string, key string, def int64, errs *[]error) int64 {
	raw := strings.TrimSpace(getenv(key))
	if raw == "" {
		return def
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	switch {
	case err != nil:
		*errs = append(*errs, fmt.Errorf("%s=%q: %w", key, raw, err))
	case n <= 0:
		*errs = append(*errs, fmt.Errorf("%s=%q: musí být kladné", key, raw))
	default:
		return n
	}
	return def
}
