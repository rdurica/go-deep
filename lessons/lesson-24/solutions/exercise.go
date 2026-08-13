// Package solutions obsahuje referenční řešení lekce 24.
package solutions

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse je tělo odpovědi health endpointu.
type HealthResponse struct {
	Status string `json:"status"`
}

// ErrorResponse je jednotný tvar chybové odpovědi celé služby.
type ErrorResponse struct {
	Error string `json:"error"`
}

// --- Stupeň: jednoduchý ---

// WriteJSON zapíše v jako JSON odpověď se status kódem status.
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// writeError je zkratka pro chybovou odpověď v jednotném tvaru.
func writeError(w http.ResponseWriter, status int, msg string) {
	_ = WriteJSON(w, status, ErrorResponse{Error: msg})
}

// --- Stupeň: střední ---

// HealthHandler vrací handler, který na jakoukoli metodu odpoví 200 a JSON {"status":"ok"}.
func HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = WriteJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
	})
}

// NotFoundHandler vrací handler, který vždy odpoví 404 a JSON ErrorResponse.
func NotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not found")
	})
}

// --- Stupeň: obtížný ---

// NewServer vrátí *http.Server s vyplněnou adresou, handlerem a všemi timeouty.
func NewServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
