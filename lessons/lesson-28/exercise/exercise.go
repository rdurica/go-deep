// Package exercise obsahuje cvičení lekce 28.
package exercise

import (
	"errors"
	"time"
)

// ErrMissing označuje chybějící povinnou hodnotu.
var ErrMissing = errors.New("missing required value")

// ErrInvalid označuje hodnotu, kterou se nepodařilo zpracovat nebo neprošla validací.
var ErrInvalid = errors.New("invalid value")

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
	panic("TODO: úkol A")
}

// LookupInt vrátí hodnotu klíče key z getenv převedenou na int, nebo def, pokud je prázdná.
func LookupInt(getenv func(string) string, key string, def int) (int, error) {
	panic("TODO: úkol A")
}

// Load sestaví Config z getenv a posbírá všechny chyby najednou.
func Load(getenv func(string) string) (Config, error) {
	panic("TODO: úkol B")
}

// String implementuje fmt.Stringer a maskuje tajemství.
func (c Config) String() string {
	panic("TODO: úkol C")
}

// LoadFromEnviron sestaví Config ze slice ve formátu os.Environ() ("KEY=VALUE").
func LoadFromEnviron(environ []string) (Config, error) {
	panic("TODO: úkol C")
}
