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
	// TODO: úkol A
	return *new(Result[T])
}

// Err vytvoří chybový výsledek. Err(nil) je platný Ok.
func Err[T any](err error) Result[T] {
	// TODO: úkol A
	return *new(Result[T])
}

// IsOk vrací true, pokud výsledek nese hodnotu (chyba je nil).
func (r Result[T]) IsOk() bool {
	// TODO: úkol A
	return false
}

// Unwrap vrací hodnotu a chybu. Při chybě vrací nulovou hodnotu T.
func (r Result[T]) Unwrap() (T, error) {
	// TODO: úkol A
	return *new(T), nil
}

// Map převede hodnotu uvnitř výsledku. Nad chybou se f nevolá.
func Map[T, U any](r Result[T], f func(T) U) Result[U] {
	// TODO: úkol A
	return *new(Result[U])
}

// Must vrátí v, nebo panikuje hodnotou err.
func Must[T any](v T, err error) T {
	// TODO: úkol A
	return *new(T)
}

// NewUser sestaví User včetně neexportovaného hesla.
func NewUser(id int, name, email string, active bool, password string) User {
	// TODO: úkol B
	return *new(User)
}

// Password vrací neexportované heslo.
func (u User) Password() string {
	// TODO: úkol B
	return ""
}

// Secret vrací neexportované pole Config.
func (c Config) Secret() string {
	// TODO: úkol B
	return ""
}

// StructToMap převede struct (nebo ukazatel na struct) na mapu podle tagů map.
func StructToMap(v any) (map[string]any, error) {
	// TODO: úkol B
	return nil, nil
}

// UserToMap dělá totéž co StructToMap pro User, ale ručně bez reflexe.
func UserToMap(u User) map[string]any {
	// TODO: úkol B
	return nil
}

// SetDefaults nastaví nulová exportovaná pole podle tagu default.
func SetDefaults(v any) error {
	// TODO: úkol B
	return nil
}
