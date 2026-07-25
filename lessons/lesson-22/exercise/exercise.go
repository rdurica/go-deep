// Package exercise obsahuje cvičení lekce 22.
package exercise

import "errors"

// ErrUnsupportedType hlásí typ, který Marshal neumí zakódovat.
var ErrUnsupportedType = errors.New("unsupported type")

// Handler zpracuje cestu a vrátí odpověď. Zjednodušená obdoba http.Handler.
type Handler interface {
	Handle(path string) string
}

// HandlerFunc adaptuje obyčejnou funkci na Handler — stejný trik jako http.HandlerFunc.
type HandlerFunc func(path string) string

// Handle volá f(path), a tím HandlerFunc splňuje interface Handler.
func (f HandlerFunc) Handle(path string) string {
	panic("TODO: úkol A")
}

// Mux je zjednodušený router. Jeho zero value je prázdný, použitelný router.
type Mux struct {
	patterns map[string]Handler
	notFound Handler
}

// Register zaregistruje handler pro vzor.
func (m *Mux) Register(pattern string, h Handler) {
	panic("TODO: úkol B")
}

// SetNotFound nastaví handler pro cesty bez shody.
func (m *Mux) SetNotFound(h Handler) {
	panic("TODO: úkol B")
}

// Handle najde nejlepší shodu a předá jí cestu.
func (m *Mux) Handle(path string) string {
	panic("TODO: úkol B")
}

// Marshal zakóduje podporované hodnoty do JSON textu bez použití reflexe.
func Marshal(v any) (string, error) {
	panic("TODO: úkol C")
}
