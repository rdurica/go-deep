package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p06-bookmarks-persist/internal/config"
)

func TestLoadRequiresDatabaseURL(t *testing.T) {
	_, err := config.Load(func(string) string { return "" })
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("chci chybu o DATABASE_URL, dostal jsem %v", err)
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := config.Load(func(k string) string {
		if k == "DATABASE_URL" {
			return "postgres://bookmarks:bookmarks@localhost:5432/bookmarks?sslmode=disable"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Addr != ":8080" || cfg.CacheTTL != 5*time.Minute {
		t.Errorf("neočekávané defaulty: %+v", cfg)
	}
	if cfg.RedisURL == "" {
		t.Error("RedisURL má mít výchozí hodnotu")
	}
}
