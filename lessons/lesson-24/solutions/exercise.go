// Package solutions obsahuje referenční řešení lekce 24.
package solutions

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"
)

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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// writeError je zkratka pro chybovou odpověď v jednotném tvaru.
func writeError(w http.ResponseWriter, status int, msg string) {
	_ = WriteJSON(w, status, ErrorResponse{Error: msg})
}

// HealthHandler vrací handler, který na jakoukoli metodu odpoví 200 a JSON {"status":"ok"}.
func HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = WriteJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
	})
}

// EchoHandler vrací handler, který přijme POST s JSON tělem EchoRequest
// a odpoví EchoResponse, nebo chybou v podobě ErrorResponse.
func EchoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			writeError(w, http.StatusUnsupportedMediaType, "expected application/json")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		var req EchoRequest
		if err := dec.Decode(&req); err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid json body")
			return
		}

		if strings.TrimSpace(req.Message) == "" {
			writeError(w, http.StatusBadRequest, "message is required")
			return
		}
		repeat := req.Repeat
		if repeat == 0 {
			repeat = 1
		}
		if repeat < 1 || repeat > 10 {
			writeError(w, http.StatusBadRequest, "repeat must be between 1 and 10")
			return
		}

		_ = WriteJSON(w, http.StatusOK, EchoResponse{
			Echo:  strings.Repeat(req.Message, repeat),
			Count: repeat,
		})
	})
}

// NotFoundHandler vrací handler, který vždy odpoví 404 a JSON ErrorResponse.
func NotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not found")
	})
}

// NewRouter složí health, echo a 404 fallback do jednoho http.ServeMux.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/healthz", HealthHandler())
	mux.Handle("/echo", EchoHandler())
	mux.Handle("/", NotFoundHandler())
	return mux
}

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
