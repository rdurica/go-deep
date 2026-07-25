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

// ReadConfig načte konfiguraci ve tvaru key=value a každou chybu opatří kontextem.
func ReadConfig(r io.Reader) (Config, error) {
	panic("TODO: úkol A")
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
	panic("TODO: úkol B")
}

// Run protáhne vstup všemi kroky.
func (p *Pipeline) Run(input string) (string, error) {
	panic("TODO: úkol B")
}

// CloseAll zavře všechny closery a spojí jejich chyby.
func CloseAll(closers []io.Closer) error {
	panic("TODO: úkol C")
}

// WithCleanup zavolá f a poté vždy cleanup, přičemž správně zkombinuje obě chyby.
func WithCleanup(f func() error, cleanup func() error) (err error) {
	panic("TODO: úkol C")
}
