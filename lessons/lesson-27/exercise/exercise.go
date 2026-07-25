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
// Prázdný struct nic nealokuje a nikdo mimo tenhle balíček ho nevyrobí.
type userKey struct{}

// WriteJSON zapíše v jako JSON odpověď se status kódem status.
// Hotové z lekce 24.
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// WithUser vrátí kopii kontextu s uloženým uživatelem.
func WithUser(ctx context.Context, u User) context.Context {
	panic("TODO: úkol A")
}

// UserFrom vytáhne uživatele z kontextu.
func UserFrom(ctx context.Context) (User, bool) {
	panic("TODO: úkol A")
}

// Authenticate vrací middleware, který ověří Bearer token a vloží uživatele do kontextu.
func Authenticate(users map[string]User) Middleware {
	panic("TODO: úkol B")
}

// WhoAmI vrací handler, který odpoví uživatelem z kontextu.
func WhoAmI() http.Handler {
	panic("TODO: úkol B")
}

// FetchWithTimeout zavolá fn s kontextem omezeným na d a respektuje deadline.
func FetchWithTimeout(ctx context.Context, fn func(context.Context) (string, error), d time.Duration) (string, error) {
	panic("TODO: úkol C")
}

// SlowHandler vrací handler, který pracuje work dlouho a reaguje na zrušení kontextu.
func SlowHandler(work time.Duration) http.Handler {
	panic("TODO: úkol C")
}

// SlowHandlerWithHook je jako SlowHandler, ale při odchodu zavolá onExit
// s důvodem ukončení (nil při úspěchu). Slouží k otestování zrušení.
func SlowHandlerWithHook(work time.Duration, onExit func(error)) http.Handler {
	panic("TODO: úkol C")
}
