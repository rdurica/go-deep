// Package exercise obsahuje cvičení lekce 54.
package exercise

import "errors"

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
	panic("TODO: úkol A")
}

// Err vytvoří chybový výsledek. Err(nil) je platný Ok.
func Err[T any](err error) Result[T] {
	panic("TODO: úkol A")
}

// IsOk vrací true, pokud výsledek nese hodnotu (chyba je nil).
func (r Result[T]) IsOk() bool {
	panic("TODO: úkol A")
}

// Unwrap vrací hodnotu a chybu. Při chybě vrací nulovou hodnotu T.
func (r Result[T]) Unwrap() (T, error) {
	panic("TODO: úkol A")
}

// Map převede hodnotu uvnitř výsledku. Nad chybou se f nevolá.
func Map[T, U any](r Result[T], f func(T) U) Result[U] {
	panic("TODO: úkol A")
}

// Must vrátí v, nebo panikuje hodnotou err.
func Must[T any](v T, err error) T {
	panic("TODO: úkol A")
}

// NewUser sestaví User včetně neexportovaného hesla.
func NewUser(id int, name, email string, active bool, password string) User {
	panic("TODO: úkol B")
}

// Password vrací neexportované heslo.
func (u User) Password() string {
	panic("TODO: úkol B")
}

// Secret vrací neexportované pole Config.
func (c Config) Secret() string {
	panic("TODO: úkol B")
}

// StructToMap převede struct (nebo ukazatel na struct) na mapu podle tagů map.
func StructToMap(v any) (map[string]any, error) {
	panic("TODO: úkol B")
}

// UserToMap dělá totéž co StructToMap pro User, ale ručně bez reflexe.
func UserToMap(u User) map[string]any {
	panic("TODO: úkol B")
}

// SetDefaults nastaví nulová exportovaná pole podle tagu default.
func SetDefaults(v any) error {
	panic("TODO: úkol B")
}
