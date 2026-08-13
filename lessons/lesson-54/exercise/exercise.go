// Package exercise obsahuje cvičení lekce 54.
package exercise

import "errors"

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

// Ok vytvoří úspěšný výsledek Result[T] s hodnotou v.
func Ok[T any](v T) Result[T] {
	return Result[T]{value: v}
}

// Err vytvoří chybový výsledek Result[T] s danou chybou.
// Err(nil) je platný Ok s nulovou hodnotou T.
func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
}

// NewUser sestaví User včetně neexportovaného hesla pro testy reflexe.
func NewUser(id int, name, email string, active bool, password string) User {
	return User{
		ID:       id,
		Name:     name,
		Email:    email,
		Active:   active,
		password: password,
	}
}

// Password vrací neexportované heslo uživatele.
func (u User) Password() string {
	return u.password
}

// --- Stupeň: jednoduchý ---
// IsOk vrací true, pokud chyba uvnitř výsledku je nil.
// Err(nil) je platný Ok; nulová hodnota Result[T] je Ok s nulovým T.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Vrací opačnou logiku — true při chybě.
// Najdi chybu a oprav — testy před opravou padají.
func (r Result[T]) IsOk() bool {
	return r.err != nil
}

// Unwrap vrací hodnotu a chybu z výsledku.
// Při chybě vrací nulovou hodnotu T, ne původní obsah uvnitř Result.
func (r Result[T]) Unwrap() (T, error) {
	// TODO
	return *new(T), nil
}

// --- Stupeň: střední ---
// Map převede hodnotu úspěšného výsledku přes f na Result[U].
// Nad chybou se f nesmí zavolat; chyba se propaguje beze změny.
func Map[T, U any](r Result[T], f func(T) U) Result[U] {
	// TODO
	return Result[U]{}
}

// --- Stupeň: obtížný ---
// StructToMap převede struct nebo pointer na struct podle tagů map (map:"-", prázdný tag).
// Neexportovaná pole přeskoč. Cokoli jiného než struct (včetně nil) → ErrNotStruct.
func StructToMap(v any) (map[string]any, error) {
	// TODO
	return nil, nil
}
