// Package solutions obsahuje referenční řešení lekce 54.
package solutions

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// Doménové chyby cvičení.
var (
	ErrNotStruct       = errors.New("hodnota není struct")
	ErrNotPointer      = errors.New("cíl není ukazatel na struct")
	ErrBadDefault      = errors.New("neplatná výchozí hodnota")
	ErrUnsupportedKind = errors.New("nepodporovaný typ pole")
)

// Result nese buď hodnotu, nebo chybu. Zero value je Ok se zero value T.
type Result[T any] struct {
	value T
	err   error
}

// User je ukázkový struct s tagy pro StructToMap a UserToMap.
type User struct {
	ID       int    `map:"id"`
	Name     string `map:"name"`
	Active   bool
	Email    string `map:"-"`
	password string
}

// Config je struktura s default tagy pro SetDefaults.
type Config struct {
	Host    string `default:"localhost"`
	Port    int    `default:"8080"`
	Debug   bool   `default:"true"`
	Retries int    `default:"3"`
	Timeout int
	secret  string
}

// BadConfig má pole s default tagem na nepodporovaném typu.
type BadConfig struct {
	Ratio float64 `default:"0.5"`
}

// Ok vytvoří úspěšný výsledek.
func Ok[T any](v T) Result[T] {
	return Result[T]{value: v}
}

// Err vytvoří chybový výsledek. Err(nil) je platný Ok.
func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
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

// Map převede hodnotu uvnitř výsledku. Nad chybou se f nevolá.
func Map[T, U any](r Result[T], f func(T) U) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return Ok(f(r.value))
}

// Must vrátí v, nebo panikuje hodnotou err.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// --- Stupeň: střední ---
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

// Secret vrací neexportované pole Config.
func (c Config) Secret() string {
	return c.secret
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

// UserToMap dělá totéž co StructToMap pro User, ale ručně bez reflexe.
func UserToMap(u User) map[string]any {
	return map[string]any{
		"id":     u.ID,
		"name":   u.Name,
		"Active": u.Active,
	}
}

// SetDefaults nastaví nulová exportovaná pole podle tagu default.
func SetDefaults(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("%T: %w", v, ErrNotPointer)
	}
	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("ukazatel na %s: %w", elem.Kind(), ErrNotPointer)
	}

	rt := elem.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}
		def, ok := f.Tag.Lookup("default")
		if !ok {
			continue
		}
		fv := elem.Field(i)
		if !fv.CanSet() || !fv.IsZero() {
			continue
		}
		if err := setDefault(fv, def); err != nil {
			return fmt.Errorf("pole %q: %w", f.Name, err)
		}
	}
	return nil
}

func setDefault(fv reflect.Value, def string) error {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(def)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(def)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrBadDefault, err)
		}
		fv.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(def, 10, fv.Type().Bits())
		if err != nil {
			return fmt.Errorf("%w: %v", ErrBadDefault, err)
		}
		fv.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(def, 10, fv.Type().Bits())
		if err != nil {
			return fmt.Errorf("%w: %v", ErrBadDefault, err)
		}
		fv.SetUint(n)
		return nil
	default:
		return ErrUnsupportedKind
	}
}
