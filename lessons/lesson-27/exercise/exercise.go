// Package exercise obsahuje cvičení lekce 27.
package exercise

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// User je přihlášený uživatel.
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ErrorResponse je jednotný tvar chybové odpovědi.
type ErrorResponse struct {
	Error string `json:"error"`
}

// StatusResponse je tělo odpovědi SlowHandleru.
type StatusResponse struct {
	Status string `json:"status"`
}

// Middleware je funkce, která obalí handler jiným handlerem.
type Middleware func(http.Handler) http.Handler

// userKey je neexportovaný typ klíče do kontextu.
type userKey struct{}

// --- Stupeň: jednoduchý ---

// WriteJSON zapíše v jako JSON odpověď se status kódem status.
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// WithUser vrátí kopii kontextu s uloženým uživatelem.
// Ulož pod klíč userKey{}; rodičovský kontext zůstane beze změny.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Klíč typu string se srazí s cizími balíčky.
// Najdi chybu a oprav — testy před opravou padají.
func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, "user", u)
}

// UserFrom vytáhne uživatele z kontextu.
// Bez uloženého uživatele zero value a false.
func UserFrom(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey{}).(User)
	return u, ok
}

// --- Stupeň: střední ---

// Authenticate vrací middleware, který ověří Bearer token a vloží uživatele do kontextu.
// Chybějící/špatné schéma/prázdný token/neznámý token → 401 + WWW-Authenticate: Bearer.
func Authenticate(users map[string]User) Middleware {
	// TODO
	return *new(Middleware)
}

// WhoAmI vrací handler, který odpoví uživatelem z kontextu.
// Bez uživatele v kontextu → 500 (chyba wiringu, ne klienta).
func WhoAmI() http.Handler {
	// TODO
	return *new(http.Handler)
}

// --- Stupeň: obtížný ---

// SlowHandler vrací handler, který pracuje work dlouho a reaguje na zrušení kontextu.
// Kroky po ~5 ms; při dokončení 200 + StatusResponse{Status:"done"}; při zrušení bez zápisu do w.
func SlowHandler(work time.Duration) http.Handler {
	// TODO
	return *new(http.Handler)
}
