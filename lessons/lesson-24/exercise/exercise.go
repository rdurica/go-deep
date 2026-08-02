// Package exercise obsahuje cvičení lekce 24.
package exercise

import "net/http"

// MaxBodyBytes je maximální povolená velikost těla požadavku v bajtech.
const MaxBodyBytes = 1024

// HealthResponse je tělo odpovědi health endpointu.
type HealthResponse struct {
	Status string `json:"status"`
}

// ErrorResponse je jednotný tvar chybové odpovědi celé služby.
type ErrorResponse struct {
	Error string `json:"error"`
}

// EchoRequest je tělo požadavku pro EchoHandler.
type EchoRequest struct {
	Message string `json:"message"`
	Repeat  int    `json:"repeat"`
}

// EchoResponse je tělo úspěšné odpovědi EchoHandleru.
type EchoResponse struct {
	Echo  string `json:"echo"`
	Count int    `json:"count"`
}

// --- Stupeň: jednoduchý ---
// WriteJSON zapíše v jako JSON odpověď se status kódem status.
// Nastav Content-Type: application/json, pak WriteHeader, pak json.NewEncoder. Vrací chybu enkodéru.
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	// TODO
	return nil
}

// HealthHandler vrací handler, který na jakoukoli metodu odpoví 200 a JSON {"status":"ok"}.
func HealthHandler() http.Handler {
	// TODO
	return *new(http.Handler)
}

// --- Stupeň: střední ---
// EchoHandler vrací handler, který přijme POST s JSON tělem EchoRequest
// a odpoví EchoResponse, nebo chybou v podobě ErrorResponse.
// POST only (405 + Allow); Content-Type přes mime.ParseMediaType (415);
// tělo omez http.MaxBytesReader — překročení 413 (errors.As *http.MaxBytesError);
// DisallowUnknownFields; validace message a repeat 1–10 (repeat 0 → 1).
func EchoHandler() http.Handler {
	// TODO
	return *new(http.Handler)
}

// NotFoundHandler vrací handler, který vždy odpoví 404 a JSON ErrorResponse.
// Tělo ErrorResponse s neprázdným polem Error (např. "not found").
func NotFoundHandler() http.Handler {
	// TODO
	return *new(http.Handler)
}

// --- Stupeň: obtížný ---
// NewRouter složí health, echo a 404 fallback do jednoho http.ServeMux.
// Trasy /healthz, /echo; / jako fallback na NotFoundHandler.
func NewRouter() http.Handler {
	// TODO
	return *new(http.Handler)
}

// NewServer vrátí *http.Server s vyplněnou adresou, handlerem a všemi timeouty.
// ReadHeaderTimeout, ReadTimeout, WriteTimeout, IdleTimeout kladné; ReadHeaderTimeout ≤ ReadTimeout.
func NewServer(addr string, h http.Handler) *http.Server {
	// TODO
	return nil
}
