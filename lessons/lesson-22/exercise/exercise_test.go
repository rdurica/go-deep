package exercise_test

import (
	"errors"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-22/exercise"
)

func TestHandlerFuncSatisfiesHandler(t *testing.T) {
	// Přesně jako u http.HandlerFunc: obyčejná funkce se stane Handlerem.
	var h exercise.Handler = exercise.HandlerFunc(func(path string) string {
		return "ahoj " + path
	})

	if got, want := h.Handle("/svete"), "ahoj /svete"; got != want {
		t.Errorf("Handle(%q) = %q, chci %q", "/svete", got, want)
	}
}

func TestHandlerFuncPassesPath(t *testing.T) {
	var seen []string
	h := exercise.HandlerFunc(func(path string) string {
		seen = append(seen, path)
		return path
	})

	h.Handle("/a")
	h.Handle("/b")

	if len(seen) != 2 || seen[0] != "/a" || seen[1] != "/b" {
		t.Errorf("handler dostal %v, chci [/a /b]", seen)
	}
}

// echo vrací handler, který odpoví pevným jménem a cestou.
func echo(name string) exercise.Handler {
	return exercise.HandlerFunc(func(path string) string {
		return name + ":" + path
	})
}

func TestMuxZeroValue(t *testing.T) {
	var mux exercise.Mux // žádný konstruktor, jako http.ServeMux

	if got, want := mux.Handle("/cokoli"), "404 not found: /cokoli"; got != want {
		t.Errorf("Handle(%q) = %q, chci %q", "/cokoli", got, want)
	}

	mux.Register("/ping", echo("ping"))
	if got, want := mux.Handle("/ping"), "ping:/ping"; got != want {
		t.Errorf("Handle(%q) = %q, chci %q", "/ping", got, want)
	}
}

func TestMuxPatternSelection(t *testing.T) {
	var mux exercise.Mux
	mux.Register("/", echo("root"))
	mux.Register("/api/", echo("api"))
	mux.Register("/api/users", echo("users"))

	tests := map[string]string{
		"/":             "root:/",
		"/index.html":   "root:/index.html",
		"/api":          "root:/api",
		"/api/":         "api:/api/",
		"/api/orders":   "api:/api/orders",
		"/api/users":    "users:/api/users",
		"/api/users/42": "api:/api/users/42",
	}
	for path, want := range tests {
		if got := mux.Handle(path); got != want {
			t.Errorf("Handle(%q) = %q, chci %q", path, got, want)
		}
	}
}

func TestMuxPanics(t *testing.T) {
	tests := map[string]func(){
		"prázdný vzor": func() {
			var mux exercise.Mux
			mux.Register("", echo("x"))
		},
		"nil handler": func() {
			var mux exercise.Mux
			mux.Register("/x", nil)
		},
		"duplicitní registrace": func() {
			var mux exercise.Mux
			mux.Register("/x", echo("a"))
			mux.Register("/x", echo("b"))
		},
	}

	for name, fn := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("chci paniku, žádná nepřišla")
				}
			}()
			fn()
		})
	}
}

func TestMarshalSupportedTypes(t *testing.T) {
	var nilSlice []string
	var nilMap map[string]string

	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, "null"},
		{"empty string", "", `""`},
		{"string", "ahoj", `"ahoj"`},
		{"string s uvozovkami", `he said "hi"`, `"he said \"hi\""`},
		{"string with backslash", `a\b`, `"a\\b"`},
		{"string with newline", "line1\nline2", `"line1\nline2"`},
		{"string with tab", "a\tb", `"a\tb"`},
		{"nula", 0, "0"},
		{"positive number", 42, "42"},
		{"negative number", -42, "-42"},
		{"slice", []string{"a", "b"}, `["a","b"]`},
		{"slice s escapem", []string{`"`}, `["\""]`},
		{"empty slice", []string{}, "[]"},
		{"nil slice", nilSlice, "null"},
		{"map with sorted keys", map[string]string{"b": "2", "a": "1", "c": "3"}, `{"a":"1","b":"2","c":"3"}`},
		{"empty map", map[string]string{}, "{}"},
		{"nil mapa", nilMap, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exercise.Marshal(tt.in)
			if err != nil {
				t.Fatalf("Marshal(%#v) vrátil chybu %v, chci nil", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Marshal(%#v) = %s, chci %s", tt.in, got, tt.want)
			}
		})
	}
}

func TestMarshalUnsupportedTypes(t *testing.T) {
	tests := []struct {
		in      any
		wantSub string
	}{
		{3.14, "float64"},
		{[]int{1, 2}, "[]int"},
		{true, "bool"},
		{map[string]int{"a": 1}, "map[string]int"},
		{struct{ A int }{1}, "struct"},
	}

	for _, tt := range tests {
		got, err := exercise.Marshal(tt.in)
		if !errors.Is(err, exercise.ErrUnsupportedType) {
			t.Fatalf("Marshal(%#v) chyba = %v, chci obalenou ErrUnsupportedType", tt.in, err)
		}
		if !strings.Contains(err.Error(), tt.wantSub) {
			t.Errorf("chyba %q neobsahuje název typu %q", err.Error(), tt.wantSub)
		}
		if got != "" {
			t.Errorf("při chybě chci prázdný výstup, mám %q", got)
		}
	}
}

func TestMarshalLargeMapIsDeterministic(t *testing.T) {
	m := map[string]string{}
	for _, k := range []string{"zeta", "alfa", "omega", "beta", "gama", "delta"} {
		m[k] = strings.ToUpper(k)
	}
	want := `{"alfa":"ALFA","beta":"BETA","delta":"DELTA","gama":"GAMA","omega":"OMEGA","zeta":"ZETA"}`

	for i := 0; i < 20; i++ {
		got, err := exercise.Marshal(m)
		if err != nil {
			t.Fatalf("Marshal vrátil chybu %v", err)
		}
		if got != want {
			t.Fatalf("Marshal = %s, chci %s (klíče musí být seřazené)", got, want)
		}
	}
}
