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
type userKey struct{}

// --- Stupeň: jednoduchý ---

// WriteJSON zapíše v jako JSON odpověď se status kódem status.
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

// --- Stupeň: střední ---

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

// --- Stupeň: obtížný ---

// SlowHandler vrací handler, který pracuje work dlouho a reaguje na zrušení kontextu.
func SlowHandler(work time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		deadline := time.Now().Add(work)

		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if time.Now().Before(deadline) {
					continue
				}
				_ = WriteJSON(w, http.StatusOK, StatusResponse{Status: "done"})
				return
			}
		}
	})
}
