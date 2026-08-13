// Package exercise obsahuje cvičení lekce 26.
package exercise

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorResponse je jednotný tvar chybové odpovědi.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Middleware je funkce, která obalí handler jiným handlerem.
type Middleware func(http.Handler) http.Handler

// --- Stupeň: jednoduchý ---

// WriteJSON zapíše v jako JSON odpověď se status kódem status.
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

// NewStatusRecorder obalí w pro zachycení statusu a velikosti odpovědi.
func NewStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{ResponseWriter: w, status: http.StatusOK}
}

// RecordedStatus vrátí zapamatovaný HTTP status.
func (rec *statusRecorder) RecordedStatus() int {
	return rec.status
}

// RecordedBytes vrátí počet bajtů zapsaných přes Write.
func (rec *statusRecorder) RecordedBytes() int {
	return rec.bytes
}

// WriteHeader zapamatuje status a přepošle ho podkladovému writeru — jen jednou.
// Druhé a další WriteHeader status nepřepisují (první vyhrává).
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Chybí ochrana proti dvojímu volání
// a výchozí status 200 pro zápis bez WriteHeader.
func (rec *statusRecorder) WriteHeader(status int) {
	rec.wroteHeader = true
	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
}

// Write přepošle data a přičte jejich délku k počítadlu.
// Pokud status ještě nebyl zapsán, doplní 200.
func (rec *statusRecorder) Write(b []byte) (int, error) {
	n, err := rec.ResponseWriter.Write(b)
	rec.bytes += n
	return n, err
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
// Použij NewStatusRecorder; výchozí status 200; log Info se zprávou "request".
func Logging(logger *slog.Logger) Middleware {
	// TODO
	return *new(Middleware)
}

// --- Stupeň: obtížný ---

// Recovery vrací middleware, který z paniky v handleru udělá 500 s JSON tělem.
// Text paniky se do odpovědi nesmí; http.ErrAbortHandler propusť dál (panic).
func Recovery() Middleware {
	// TODO
	return *new(Middleware)
}
