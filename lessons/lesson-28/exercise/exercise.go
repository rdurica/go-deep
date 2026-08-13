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

// --- Stupeň: jednoduchý ---

// LookupString vrátí hodnotu klíče key z getenv, nebo def, pokud je prázdná.
// Chybějící klíč i prázdný string z getenv → def.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Prázdný řetězec nepadá na výchozí hodnotu.
// Najdi chybu a oprav — testy před opravou padají.
func LookupString(getenv func(string) string, key, def string) string {
	return getenv(key)
}

// --- Stupeň: střední ---

// LookupInt vrátí hodnotu klíče key z getenv převedenou na int, nebo def, pokud je prázdná.
// Nečíselná hodnota → 0 a chyba obalující ErrInvalid s názvem klíče v textu.
func LookupInt(getenv func(string) string, key string, def int) (int, error) {
	// TODO
	return 0, nil
}

// Load sestaví Config z getenv a posbírá všechny chyby najednou.
// errors.Join; errors.Is pro ErrMissing i ErrInvalid. PORT 1–65535; READ_TIMEOUT kladný;
// DATABASE_URL a API_KEY povinné. ADDR výchozí 0.0.0.0, PORT 8080, READ_TIMEOUT 5s, DEBUG false.
func Load(getenv func(string) string) (Config, error) {
	// TODO
	return *new(Config), nil
}

// --- Stupeň: obtížný ---

// String implementuje fmt.Stringer a maskuje tajemství.
// APIKey a heslo z DatabaseURL se do výstupu nedostanou; port a host ano.
func (c Config) String() string {
	// TODO
	return ""
}
