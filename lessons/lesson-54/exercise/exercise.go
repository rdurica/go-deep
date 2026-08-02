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

// Ok vytvoří úspěšný výsledek Result[T] s hodnotou v.
// Nulová hodnota Result[T] je použitelná Ok s nulovým T.
func Ok[T any](v T) Result[T] {
	// TODO
	return Result[T]{}
}

// Err vytvoří chybový výsledek Result[T] s danou chybou.
// Err(nil) je platný Ok s nulovou hodnotou T — IsOk závisí jen na nil chybě.
func Err[T any](err error) Result[T] {
	// TODO
	return Result[T]{}
}

// --- Stupeň: jednoduchý ---
// IsOk vrací true, pokud chyba uvnitř výsledku je nil.
// Err(nil) je platný Ok; nulová hodnota Result[T] je Ok s nulovým T.
func (r Result[T]) IsOk() bool {
	// TODO
	return false
}

// Unwrap vrací hodnotu a chybu z výsledku.
// Při chybě vrací nulovou hodnotu T, ne původní obsah uvnitř Result.
func (r Result[T]) Unwrap() (T, error) {
	// TODO
	return *new(T), nil
}

// Map převede hodnotu úspěšného výsledku přes f na Result[U].
// Nad chybou se f nesmí zavolat; chyba se propaguje beze změny.
func Map[T, U any](r Result[T], f func(T) U) Result[U] {
	// TODO
	return Result[U]{}
}

// Must vrátí v při úspěchu; při chybě panikuje hodnotou té chyby (panic(err)).
// Test chybu vytahuje přes recover() a errors.Is.
func Must[T any](v T, err error) T {
	// TODO
	return *new(T)
}

// --- Stupeň: střední ---
// NewUser sestaví User včetně neexportovaného hesla pro testy reflexe.
// Email má tag map:"-" — v mapě z reflexe se neobjeví.
func NewUser(id int, name, email string, active bool, password string) User {
	// TODO
	return User{}
}

// Password vrací neexportované heslo uživatele.
// Slouží testům reflexe; v mapě se heslo neobjeví.
func (u User) Password() string {
	// TODO
	return ""
}

// Secret vrací neexportované pole konfigurace.
// SetDefaults ho ignoruje; v mapě z reflexe chybí.
func (c Config) Secret() string {
	// TODO
	return ""
}

// --- Stupeň: obtížný ---
// StructToMap převede struct nebo pointer na struct podle tagů map (map:"-", prázdný tag).
// Neexportovaná pole přeskoč. Cokoli jiného než struct (včetně nil) → ErrNotStruct.
func StructToMap(v any) (map[string]any, error) {
	// TODO
	return nil, nil
}

// UserToMap dělá totéž co StructToMap pro User, ale ručně bez reflexe.
// Test ověřuje, že obě funkce vrátí identickou mapu; benchmark je porovnává.
func UserToMap(u User) map[string]any {
	// TODO
	return nil
}

// SetDefaults nastaví nulová exportovaná pole podle tagu default (string, int, bool).
// Vyžaduje nenilový pointer na struct. Vyplněné pole nepřepisuje; neexportované ignoruje.
// ErrNotPointer, ErrBadDefault nebo ErrUnsupportedKind podle typu chyby.
func SetDefaults(v any) error {
	// TODO
	return nil
}
