// Package solutions obsahuje referenční řešení lekce 27.
package solutions

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
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
	return context.WithValue(ctx, userKey{}, u)
}

// UserFrom vytáhne uživatele z kontextu.
func UserFrom(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey{}).(User)
	return u, ok
}

// unauthorized odešle 401 s výzvou k autentizaci.
func unauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	_ = WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: msg})
}

// Authenticate vrací middleware, který ověří Bearer token a vloží uživatele do kontextu.
func Authenticate(users map[string]User) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scheme, token, found := strings.Cut(strings.TrimSpace(r.Header.Get("Authorization")), " ")
			token = strings.TrimSpace(token)
			if !found || !strings.EqualFold(scheme, "Bearer") || token == "" {
				unauthorized(w, "missing bearer token")
				return
			}

			user, ok := users[token]
			if !ok {
				unauthorized(w, "unknown token")
				return
			}

			next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
		})
	}
}

// WhoAmI vrací handler, který odpoví uživatelem z kontextu.
func WhoAmI() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFrom(r.Context())
		if !ok {
			_ = WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "user missing in context"})
			return
		}
		_ = WriteJSON(w, http.StatusOK, user)
	})
}

// FetchWithTimeout zavolá fn s kontextem omezeným na d a respektuje deadline.
func FetchWithTimeout(ctx context.Context, fn func(context.Context) (string, error), d time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()

	type result struct {
		value string
		err   error
	}

	// buffer 1: goroutina musí mít kam odložit výsledek i po našem odchodu,
	// jinak by po timeoutu navždy visela na zápisu do kanálu
	done := make(chan result, 1)
	go func() {
		value, err := fn(ctx)
		done <- result{value: value, err: err}
	}()

	select {
	case res := <-done:
		return res.value, res.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// SlowHandler vrací handler, který pracuje work dlouho a reaguje na zrušení kontextu.
func SlowHandler(work time.Duration) http.Handler {
	return SlowHandlerWithHook(work, nil)
}

// SlowHandlerWithHook je jako SlowHandler, ale při odchodu zavolá onExit
// s důvodem ukončení (nil při úspěchu). Slouží k otestování zrušení.
func SlowHandlerWithHook(work time.Duration, onExit func(error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		deadline := time.Now().Add(work)

		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// klient se odpojil nebo vypršel deadline — dál pracovat nemá smysl
				if onExit != nil {
					onExit(ctx.Err())
				}
				return
			case <-ticker.C:
				if time.Now().Before(deadline) {
					continue
				}
				_ = WriteJSON(w, http.StatusOK, StatusResponse{Status: "done"})
				if onExit != nil {
					onExit(nil)
				}
				return
			}
		}
	})
}
