// Package solutions obsahuje referenční řešení lekce 22.
package solutions

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

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
	return f(path)
}

// Mux je zjednodušený router. Jeho zero value je prázdný, použitelný router.
type Mux struct {
	patterns map[string]Handler
	notFound Handler
}

// --- Stupeň: střední ---

// Register zaregistruje handler pro vzor.
func (m *Mux) Register(pattern string, h Handler) {
	if pattern == "" {
		panic("mux: empty pattern")
	}
	if h == nil {
		panic("mux: nil handler")
	}
	if _, ok := m.patterns[pattern]; ok {
		panic("mux: multiple registrations for " + pattern)
	}
	if m.patterns == nil {
		m.patterns = make(map[string]Handler)
	}
	m.patterns[pattern] = h
}

// Handle najde nejlepší shodu a předá jí cestu.
func (m *Mux) Handle(path string) string {
	if h := m.match(path); h != nil {
		return h.Handle(path)
	}
	if m.notFound != nil {
		return m.notFound.Handle(path)
	}
	return "404 not found: " + path
}

// match vrací handler nejdelšího vyhovujícího vzoru, jinak nil.
func (m *Mux) match(path string) Handler {
	var best string
	var found Handler
	for pattern, h := range m.patterns {
		if !matches(pattern, path) {
			continue
		}
		if found == nil || len(pattern) > len(best) {
			best, found = pattern, h
		}
	}
	return found
}

// matches implementuje pravidlo net/http: vzor s koncovým lomítkem je podstrom.
func matches(pattern, path string) bool {
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(path, pattern)
	}
	return path == pattern
}

// --- Stupeň: obtížný ---

// Marshal zakóduje podporované hodnoty do JSON textu bez použití reflexe.
func Marshal(v any) (string, error) {
	switch val := v.(type) {
	case nil:
		return "null", nil
	case string:
		return quote(val), nil
	case int:
		return strconv.Itoa(val), nil
	case []string:
		return marshalSlice(val), nil
	case map[string]string:
		return marshalMap(val), nil
	default:
		return "", fmt.Errorf("%w: %T", ErrUnsupportedType, v)
	}
}

// marshalSlice zakóduje slice řetězců; nil slice je null, prázdný je [].
func marshalSlice(items []string) string {
	if items == nil {
		return "null"
	}

	var b strings.Builder
	b.WriteByte('[')
	for i, item := range items {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(quote(item))
	}
	b.WriteByte(']')
	return b.String()
}

// marshalMap zakóduje mapu se seřazenými klíči; nil mapa je null, prázdná je {}.
func marshalMap(m map[string]string) string {
	if m == nil {
		return "null"
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(quote(k))
		b.WriteByte(':')
		b.WriteString(quote(m[k]))
	}
	b.WriteByte('}')
	return b.String()
}

// quote obalí řetězec uvozovkami a escapuje znaky, které JSON zakazuje.
func quote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
