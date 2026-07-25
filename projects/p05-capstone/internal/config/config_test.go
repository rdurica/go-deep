package config_test

import (
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p05-capstone/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := config.Load(func(string) string { return "" })
	if err != nil {
		t.Fatalf("Load = %v", err)
	}
	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, chci :8080", cfg.Addr)
	}
	if cfg.MaxBodyBytes != 1<<20 {
		t.Errorf("MaxBodyBytes = %d", cfg.MaxBodyBytes)
	}
}

func TestLoadOverrides(t *testing.T) {
	env := map[string]string{
		"BOOKMARKS_ADDR":             ":9090",
		"BOOKMARKS_READ_TIMEOUT":     "2s",
		"BOOKMARKS_MAX_BODY_BYTES":   "4096",
		"BOOKMARKS_REQUEST_TIMEOUT":  "1s",
		"BOOKMARKS_SHUTDOWN_TIMEOUT": "7s",
	}
	cfg, err := config.Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("Load = %v", err)
	}
	if cfg.Addr != ":9090" || cfg.ReadTimeout != 2*time.Second || cfg.MaxBodyBytes != 4096 {
		t.Errorf("neočekávaná konfigurace: %+v", cfg)
	}
}

func TestLoadBadDuration(t *testing.T) {
	_, err := config.Load(func(k string) string {
		if k == "BOOKMARKS_READ_TIMEOUT" {
			return "nope"
		}
		return ""
	})
	if err == nil {
		t.Fatal("chci chybu pro neplatný timeout")
	}
}
