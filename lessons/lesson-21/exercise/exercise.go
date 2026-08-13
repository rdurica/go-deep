// Package exercise obsahuje cvičení lekce 21.
package exercise

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Sentinelové chyby konfigurace. Text je malým písmenem, bez tečky a bez "failed to".
var (
	ErrMalformedLine = errors.New("malformed line")
	ErrUnknownKey    = errors.New("unknown key")
	ErrMissingKey    = errors.New("missing key")
	ErrInvalidPort   = errors.New("invalid port")
	ErrInvalidBool   = errors.New("invalid bool")
)

// Sentinelové chyby pipeline a úklidu.
var (
	ErrNilStep = errors.New("nil step function")
	ErrNilFunc = errors.New("nil function")
)

// Config je konfigurace načtená z textového vstupu.
type Config struct {
	Name  string
	Port  int
	Debug bool
}

// --- Stupeň: jednoduchý ---

// ReadConfig načte konfiguraci ve tvaru key=value a každou chybu opatří kontextem.
// Prázdné a # řádky přeskoč (počítají se do čísla řádku); chybějící = je malformed line.
// Klíče name, port, debug (case sensitive); poslední výskyt vyhrává; chybí debug → false.
// Chyby řádku: line N: …; chybějící name před port; sentinely dohledatelné přes errors.Is.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ — typické AI/PHP zápachy v error handlingu.
// Oprav texty chyb, %w místo %v a formát podle kontraktu — testy před opravou padají.
func ReadConfig(r io.Reader) (Config, error) {
	var cfg Config
	hasPort := false

	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return Config{}, fmt.Errorf("Error on line %d: Failed to parse line %q.", n, line)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "name":
			cfg.Name = value
		case "port":
			port, err := strconv.Atoi(value)
			if err != nil || port < 1 || port > 65535 {
				return Config{}, fmt.Errorf("Line %d: Invalid port value %q.", n, value)
			}
			cfg.Port = port
			hasPort = true
		case "debug":
			debug, err := strconv.ParseBool(value)
			if err != nil {
				return Config{}, fmt.Errorf("Line %d: Invalid boolean %q.", n, value)
			}
			cfg.Debug = debug
		default:
			return Config{}, fmt.Errorf("Line %d: Unknown configuration key %q.", n, key)
		}
	}
	if err := sc.Err(); err != nil {
		return Config{}, fmt.Errorf("Failed to read configuration: %v", err)
	}

	if cfg.Name == "" {
		return Config{}, fmt.Errorf("Missing required key %q.", "name")
	}
	if !hasPort {
		return Config{}, fmt.Errorf("Missing required key %q.", "port")
	}
	return cfg, nil
}

// Step je pojmenovaný krok pipeline.
type Step struct {
	Name string
	Fn   func(string) (string, error)
}

// Pipeline spouští kroky za sebou a staví řetěz kontextu v chybách.
type Pipeline struct {
	steps []Step
}

// --- Stupeň: střední ---

// NewPipeline vrací pipeline složenou ze zadaných kroků.
func NewPipeline(steps ...Step) *Pipeline {
	// TODO
	return nil
}

// Run protáhne vstup všemi kroky.
// Prázdná pipeline vrací vstup beze změny. Chyba kroku: fmt.Errorf("step %q: %w", name, err).
// Fn == nil → chyba obalující ErrNilStep. Další kroky se po chybě nespustí.
func (p *Pipeline) Run(input string) (string, error) {
	// TODO
	return "", nil
}

// --- Stupeň: obtížný ---

// WithCleanup zavolá f a poté vždy cleanup, přičemž správně zkombinuje obě chyby.
// f == nil → ErrNilFunc, cleanup se nevolá. cleanup == nil je no-op.
// Chybu cleanup obal: fmt.Errorf("cleanup: %w", err). Chybu z f neobaluj.
// Selžou-li obě, vrať obě (errors.Join). Vyžaduje named return (err error).
func WithCleanup(f func() error, cleanup func() error) (err error) {
	// TODO
	return
}
