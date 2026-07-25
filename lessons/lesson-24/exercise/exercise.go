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

// WriteJSON zapíše v jako JSON odpověď se status kódem status.
// Musí nastavit hlavičku Content-Type dřív, než zavolá WriteHeader.
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	panic("TODO: úkol A")
}

// HealthHandler vrací handler, který na jakoukoli metodu odpoví 200 a JSON {"status":"ok"}.
func HealthHandler() http.Handler {
	panic("TODO: úkol A")
}

// EchoHandler vrací handler, který přijme POST s JSON tělem EchoRequest
// a odpoví EchoResponse, nebo chybou v podobě ErrorResponse.
func EchoHandler() http.Handler {
	panic("TODO: úkol B")
}

// NotFoundHandler vrací handler, který vždy odpoví 404 a JSON ErrorResponse.
func NotFoundHandler() http.Handler {
	panic("TODO: úkol C")
}

// NewRouter složí health, echo a 404 fallback do jednoho http.ServeMux.
func NewRouter() http.Handler {
	panic("TODO: úkol C")
}

// NewServer vrátí *http.Server s vyplněnou adresou, handlerem a všemi timeouty.
func NewServer(addr string, h http.Handler) *http.Server {
	panic("TODO: úkol C")
}
