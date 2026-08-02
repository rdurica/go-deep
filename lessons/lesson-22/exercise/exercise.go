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

// --- Stupeň: jednoduchý ---
// Handle volá f(path), a tím HandlerFunc splňuje interface Handler.
func (f HandlerFunc) Handle(path string) string {
	// TODO
	return ""
}

// Mux je zjednodušený router. Jeho zero value je prázdný, použitelný router.
type Mux struct {
	patterns map[string]Handler
	notFound Handler
}

// --- Stupeň: střední ---
// Register zaregistruje handler pro vzor.
// Panika "mux: empty pattern", "mux: nil handler" nebo "mux: multiple registrations for …".
// Mapu inicializuj líně; zero value muxu je použitelná.
func (m *Mux) Register(pattern string, h Handler) {
	// TODO
}

// SetNotFound nastaví handler pro cesty bez shody.
// nil handler panikuje.
func (m *Mux) SetNotFound(h Handler) {
	// TODO
}

// Handle najde nejlepší shodu a předá jí cestu.
// Vzor bez / jen přesná shoda; s / prefix; vyhrává nejdelší; bez shody notFound nebo výchozí 404.
func (m *Mux) Handle(path string) string {
	// TODO
	return ""
}

// --- Stupeň: obtížný ---
// Marshal zakóduje podporované hodnoty do JSON textu bez použití reflexe.
// nil slice/map → null; prázdné → [] nebo {}; klíče mapy vzestupně; escapuj ", \, \n, \t.
// Nepodporovaný typ → fmt.Errorf("%w: %T", ErrUnsupportedType, v).
func Marshal(v any) (string, error) {
	// TODO
	return "", nil
}
