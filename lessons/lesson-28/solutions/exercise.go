// Package solutions obsahuje referenční řešení lekce 28.
package solutions

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ErrMissing označuje chybějící povinnou hodnotu.
var ErrMissing = errors.New("missing required value")

// ErrInvalid označuje hodnotu, kterou se nepodařilo zpracovat nebo neprošla validací.
var ErrInvalid = errors.New("invalid value")

// Výchozí hodnoty konfigurace.
const (
	defaultAddr        = "0.0.0.0"
	defaultPort        = 8080
	defaultReadTimeout = 5 * time.Second
)

// mask je náhrada za tajemství ve výstupu Config.String.
const mask = "***"

// Config je kompletní konfigurace aplikace načtená z prostředí.
type Config struct {
	Addr        string
	Port        int
	ReadTimeout time.Duration
	Debug       bool
	DatabaseURL string
	APIKey      string
}

// LookupString vrátí hodnotu klíče key z getenv, nebo def, pokud je prázdná.
func LookupString(getenv func(string) string, key, def string) string {
	if v := getenv(key); v != "" {
		return v
	}
	return def
}

// LookupInt vrátí hodnotu klíče key z getenv převedenou na int, nebo def, pokud je prázdná.
func LookupInt(getenv func(string) string, key string, def int) (int, error) {
	raw := getenv(key)
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s=%q: %w", key, raw, ErrInvalid)
	}
	return n, nil
}

// Load sestaví Config z getenv a posbírá všechny chyby najednou.
func Load(getenv func(string) string) (Config, error) {
	var errs []error

	cfg := Config{
		Addr:        LookupString(getenv, "ADDR", defaultAddr),
		ReadTimeout: defaultReadTimeout,
	}

	port, err := LookupInt(getenv, "PORT", defaultPort)
	if err != nil {
		errs = append(errs, err)
	} else if port < 1 || port > 65535 {
		errs = append(errs, fmt.Errorf("PORT=%d mimo rozsah 1-65535: %w", port, ErrInvalid))
	} else {
		cfg.Port = port
	}

	if raw := getenv("READ_TIMEOUT"); raw != "" {
		d, err := time.ParseDuration(raw)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("READ_TIMEOUT=%q: %w", raw, ErrInvalid))
		case d <= 0:
			errs = append(errs, fmt.Errorf("READ_TIMEOUT=%q musí být kladné: %w", raw, ErrInvalid))
		default:
			cfg.ReadTimeout = d
		}
	}

	if raw := getenv("DEBUG"); raw != "" {
		b, err := strconv.ParseBool(raw)
		if err != nil {
			errs = append(errs, fmt.Errorf("DEBUG=%q: %w", raw, ErrInvalid))
		} else {
			cfg.Debug = b
		}
	}

	if cfg.DatabaseURL = getenv("DATABASE_URL"); cfg.DatabaseURL == "" {
		errs = append(errs, fmt.Errorf("DATABASE_URL: %w", ErrMissing))
	}

	if cfg.APIKey = getenv("API_KEY"); cfg.APIKey == "" {
		errs = append(errs, fmt.Errorf("API_KEY: %w", ErrMissing))
	}

	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}
	return cfg, nil
}

// String implementuje fmt.Stringer a maskuje tajemství.
func (c Config) String() string {
	apiKey := ""
	if c.APIKey != "" {
		apiKey = mask
	}
	return fmt.Sprintf(
		"Config{Addr:%s Port:%d ReadTimeout:%s Debug:%t DatabaseURL:%s APIKey:%s}",
		c.Addr, c.Port, c.ReadTimeout, c.Debug, redactURL(c.DatabaseURL), apiKey,
	)
}

// redactURL nahradí heslo v URL. Nerozparsovatelnou hodnotu zamaskuje celou —
// raději přijdeme o informaci v logu, než abychom vypsali tajemství.
func redactURL(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return mask
	}
	return u.Redacted()
}

// LoadFromEnviron sestaví Config ze slice ve formátu os.Environ() ("KEY=VALUE").
func LoadFromEnviron(environ []string) (Config, error) {
	env := make(map[string]string, len(environ))
	for _, entry := range environ {
		key, value, ok := strings.Cut(entry, "=")
		if !ok || key == "" {
			continue
		}
		env[key] = value
	}
	return Load(func(key string) string { return env[key] })
}
