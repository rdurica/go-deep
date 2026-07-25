// Package solutions obsahuje referenční řešení lekce 21.
package solutions

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

// ReadConfig načte konfiguraci ve tvaru key=value a každou chybu opatří kontextem.
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
			return Config{}, fmt.Errorf("line %d: %w: %q", n, ErrMalformedLine, line)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "name":
			cfg.Name = value
		case "port":
			port, err := strconv.Atoi(value)
			if err != nil || port < 1 || port > 65535 {
				return Config{}, fmt.Errorf("line %d: %w: %q", n, ErrInvalidPort, value)
			}
			cfg.Port = port
			hasPort = true
		case "debug":
			debug, err := strconv.ParseBool(value)
			if err != nil {
				return Config{}, fmt.Errorf("line %d: %w: %q", n, ErrInvalidBool, value)
			}
			cfg.Debug = debug
		default:
			return Config{}, fmt.Errorf("line %d: %w: %q", n, ErrUnknownKey, key)
		}
	}
	if err := sc.Err(); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	if cfg.Name == "" {
		return Config{}, fmt.Errorf("%w: %q", ErrMissingKey, "name")
	}
	if !hasPort {
		return Config{}, fmt.Errorf("%w: %q", ErrMissingKey, "port")
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

// NewPipeline vrací pipeline složenou ze zadaných kroků.
func NewPipeline(steps ...Step) *Pipeline {
	return &Pipeline{steps: steps}
}

// Run protáhne vstup všemi kroky.
func (p *Pipeline) Run(input string) (string, error) {
	out := input
	for _, step := range p.steps {
		if step.Fn == nil {
			return "", fmt.Errorf("step %q: %w", step.Name, ErrNilStep)
		}
		next, err := step.Fn(out)
		if err != nil {
			return "", fmt.Errorf("step %q: %w", step.Name, err)
		}
		out = next
	}
	return out, nil
}

// CloseAll zavře všechny closery a spojí jejich chyby.
func CloseAll(closers []io.Closer) error {
	var errs []error
	for i, c := range closers {
		if c == nil {
			continue
		}
		if err := c.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close %d: %w", i, err))
		}
	}
	return errors.Join(errs...)
}

// WithCleanup zavolá f a poté vždy cleanup, přičemž správně zkombinuje obě chyby.
func WithCleanup(f func() error, cleanup func() error) (err error) {
	if f == nil {
		return ErrNilFunc
	}

	defer func() {
		if cleanup == nil {
			return
		}
		if cerr := cleanup(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("cleanup: %w", cerr))
		}
	}()

	return f()
}
