// Package exercise obsahuje cvičení lekce 26.
package exercise

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// RequestIDHeader je jméno hlavičky, ve které se přenáší ID požadavku.
const RequestIDHeader = "X-Request-ID"

// ErrorResponse je jednotný tvar chybové odpovědi.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Middleware je funkce, která obalí handler jiným handlerem.
type Middleware func(http.Handler) http.Handler

// WriteJSON zapíše v jako JSON odpověď se status kódem status.
// Hotové z lekce 24.
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// statusRecorder obaluje http.ResponseWriter a pamatuje si status kód
// a počet zapsaných bajtů, aby je middleware mohl zalogovat.
type statusRecorder struct {
	http.ResponseWriter

	status      int
	bytes       int
	wroteHeader bool
}

// WriteHeader zapamatuje status a přepošle ho podkladovému writeru — jen jednou.
func (rec *statusRecorder) WriteHeader(status int) {
	// TODO: úkol B
}

// Write přepošle data a přičte jejich délku k počítadlu.
func (rec *statusRecorder) Write(b []byte) (int, error) {
	// TODO: úkol B
	return 0, nil
}

// Chain obalí h middlewary tak, že první uvedený je nejvíc vnější.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	// TODO: úkol A
	return *new(http.Handler)
}

// Logging vrací middleware, který po dokončení požadavku zaloguje
// metodu, cestu, status a velikost odpovědi.
func Logging(logger *slog.Logger) Middleware {
	// TODO: úkol B
	return *new(Middleware)
}

// Recovery vrací middleware, který z paniky v handleru udělá 500 s JSON tělem.
func Recovery() Middleware {
	// TODO: úkol B
	return *new(Middleware)
}

// RequestID vrací middleware, který doplní ID požadavku do kontextu i do odpovědi.
func RequestID() Middleware {
	// TODO: úkol C
	return *new(Middleware)
}

// RequestIDFrom vytáhne ID požadavku z kontextu.
func RequestIDFrom(ctx context.Context) (string, bool) {
	// TODO: úkol C
	return "", false
}

// Timeout vrací middleware, který požadavku nastaví deadline d.
func Timeout(d time.Duration) Middleware {
	// TODO: úkol C
	return *new(Middleware)
}
