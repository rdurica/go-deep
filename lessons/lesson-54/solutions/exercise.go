// Package solutions obsahuje referenční řešení lekce 54.
package solutions

import (
	"errors"
	"fmt"
	"reflect"
)

// Doménové chyby cvičení.
var (
	ErrNotStruct = errors.New("hodnota není struct")
)

// Result nese buď hodnotu, nebo chybu. Zero value je Ok se zero value T.
type Result[T any] struct {
	value T
	err   error
}

// User je ukázkový struct s tagy pro StructToMap.
type User struct {
	ID       int    `map:"id"`
	Name     string `map:"name"`
	Active   bool
	Email    string `map:"-"`
	password string
}

// Ok vytvoří úspěšný výsledek.
func Ok[T any](v T) Result[T] {
	return Result[T]{value: v}
}

// Err vytvoří chybový výsledek. Err(nil) je platný Ok.
func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
}

// NewUser sestaví User včetně neexportovaného hesla.
func NewUser(id int, name, email string, active bool, password string) User {
	return User{
		ID:       id,
		Name:     name,
		Email:    email,
		Active:   active,
		password: password,
	}
}

// Password vrací neexportované heslo.
func (u User) Password() string {
	return u.password
}

// --- Stupeň: jednoduchý ---
// IsOk vrací true, pokud výsledek nese hodnotu (chyba je nil).
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// Unwrap vrací hodnotu a chybu. Při chybě vrací nulovou hodnotu T.
func (r Result[T]) Unwrap() (T, error) {
	if r.err != nil {
		var zero T
		return zero, r.err
	}
	return r.value, nil
}

// --- Stupeň: střední ---
// Map převede hodnotu uvnitř výsledku. Nad chybou se f nevolá.
func Map[T, U any](r Result[T], f func(T) U) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return Ok(f(r.value))
}

// --- Stupeň: obtížný ---
// StructToMap převede struct (nebo ukazatel na struct) na mapu podle tagů map.
func StructToMap(v any) (map[string]any, error) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil, fmt.Errorf("%v: %w", v, ErrNotStruct)
	}
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, fmt.Errorf("%v: %w", v, ErrNotStruct)
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%T: %w", v, ErrNotStruct)
	}

	rt := rv.Type()
	out := make(map[string]any)
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}
		tag, ok := f.Tag.Lookup("map")
		if ok {
			if tag == "-" {
				continue
			}
		}
		name := f.Name
		if ok && tag != "" {
			name = tag
		}
		out[name] = rv.Field(i).Interface()
	}
	return out, nil
}
