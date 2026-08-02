// Package exercise obsahuje cvičení lekce 21.
package exercise

import (
	"errors"
	"io"
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
func ReadConfig(r io.Reader) (Config, error) {
	// TODO
	return *new(Config), nil
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
// CloseAll zavře všechny closery a spojí jejich chyby.
// Zavře všechny i při chybě; nil prvky přeskoč; chyby přes errors.Join a fmt.Errorf("close %d: %w", i, err).
func CloseAll(closers []io.Closer) error {
	// TODO
	return nil
}

// WithCleanup zavolá f a poté vždy cleanup, přičemž správně zkombinuje obě chyby.
// f == nil → ErrNilFunc, cleanup se nevolá. cleanup == nil je no-op.
// Chybu cleanup obal: fmt.Errorf("cleanup: %w", err). Chybu z f neobaluj.
// Selžou-li obě, vrať obě (errors.Join). Vyžaduje named return (err error).
func WithCleanup(f func() error, cleanup func() error) (err error) {
	// TODO
	return
}
