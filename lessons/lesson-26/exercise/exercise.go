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

// --- Stupeň: jednoduchý ---
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
// Druhé a další WriteHeader status nepřepisují (první vyhrává).
func (rec *statusRecorder) WriteHeader(status int) {
	// TODO
}

// Write přepošle data a přičte jejich délku k počítadlu.
// Pokud status ještě nebyl zapsán, doplní 200.
func (rec *statusRecorder) Write(b []byte) (int, error) {
	// TODO
	return 0, nil
}

// --- Stupeň: střední ---
// Chain obalí h middlewary tak, že první uvedený je nejvíc vnější.
// Bez middlewarů vrátí h beze změny.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	// TODO
	return *new(http.Handler)
}

// Logging vrací middleware, který po dokončení požadavku zaloguje
// metodu, cestu, status a velikost odpovědi.
// Výchozí status 200; log Info se zprávou "request" a poli method, path, status, bytes.
func Logging(logger *slog.Logger) Middleware {
	// TODO
	return *new(Middleware)
}

// Recovery vrací middleware, který z paniky v handleru udělá 500 s JSON tělem.
// Text paniky se do odpovědi nesmí; http.ErrAbortHandler propusť dál (panic).
func Recovery() Middleware {
	// TODO
	return *new(Middleware)
}

// --- Stupeň: obtížný ---
// RequestID vrací middleware, který doplní ID požadavku do kontextu i do odpovědi.
// Z hlavičky X-Request-ID nebo nové z crypto/rand; klíč neexportovaného typu.
func RequestID() Middleware {
	// TODO
	return *new(Middleware)
}

// RequestIDFrom vytáhne ID požadavku z kontextu.
// Bez middlewaru vrací "", false.
func RequestIDFrom(ctx context.Context) (string, bool) {
	// TODO
	return "", false
}

// Timeout vrací middleware, který požadavku nastaví deadline d.
// Při překročení 503 s tělem obsahujícím slovo timeout.
func Timeout(d time.Duration) Middleware {
	// TODO
	return *new(Middleware)
}
