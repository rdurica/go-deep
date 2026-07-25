package httpapi

import (
	"encoding/json"
	"net/http"
)

// Kódy chyb v odpovědi. Klient se rozhoduje podle nich, ne podle textu zprávy.
const (
	CodeBadRequest           = "bad_request"
	CodeValidationFailed     = "validation_failed"
	CodeNotFound             = "not_found"
	CodeMethodNotAllowed     = "method_not_allowed"
	CodeUnsupportedMediaType = "unsupported_media_type"
	CodeInternalError        = "internal_error"
)

// errorBody je vnitřní část chybové odpovědi.
type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// errorResponse je jednotný tvar každé chybové odpovědi API.
type errorResponse struct {
	Error errorBody `json:"error"`
}

// writeJSON pošle payload jako JSON s daným status kódem.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	// Hlavičky už jsou odeslané, takže na chybu zápisu se dá jen zalogovat —
	// a to dělá logovací middleware podle status kódu.
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError pošle chybu v konzistentním tvaru {"error":{"code","message"}}.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Error: errorBody{Code: code, Message: message}})
}
