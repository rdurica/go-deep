// Package exercise obsahuje cvičení lekce 36.
package exercise

import (
	"errors"
	"net/http"
	"strings"
)

var (
	ErrNotFound        = errors.New("zdroj nenalezen")
	ErrConflict        = errors.New("konflikt se stavem zdroje")
	ErrMalformedJSON   = errors.New("neplatné JSON tělo")
	ErrEmptyEmail      = errors.New("e-mail je prázdný")
	ErrInvalidEmail    = errors.New("e-mail má neplatný tvar")
	ErrEmptyUsername   = errors.New("uživatelské jméno je prázdné")
	ErrInvalidUsername = errors.New("uživatelské jméno má neplatný tvar")
)

const maxEmailLength = 254
const maxRequestBody = 1 << 20

// Email je normalizovaná e-mailová adresa.
type Email struct{ value string }

// String vrací normalizovanou adresu.
func (e Email) String() string { return e.value }

// IsZero hlásí, zda adresa není nastavená.
func (e Email) IsZero() bool { return e.value == "" }

// ValidationErrors mapuje jméno pole na důvod neplatnosti.
type ValidationErrors map[string]string

// Username je ověřené uživatelské jméno.
type Username struct{ value string }

// String vrací uživatelské jméno.
func (u Username) String() string { return u.value }

// Validator je typ, který umí ověřit vlastní invarianty.
type Validator interface {
	Validate() error
}

// CreateUserRequest je vstup z HTTP hranice.
type CreateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Age      int    `json:"age"`
}

const ProblemContentType = "application/problem+json"

// ProblemDetails je tělo chybové odpovědi podle RFC 7807.
type ProblemDetails struct {
	Type   string            `json:"type"`
	Title  string            `json:"title"`
	Status int               `json:"status"`
	Detail string            `json:"detail,omitempty"`
	Errors map[string]string `json:"errors,omitempty"`
}

// --- Stupeň: jednoduchý ---

// Error vrátí deterministický popis: pole abecedně, tvar "pole: důvod" spojený "; ".
// Prázdná mapa vrátí "neplatná data".
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Iteruje mapu bez řazení klíčů, takže
// pořadí v řetězci je náhodné. Najdi chybu a oprav — testy před opravou padají.
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "neplatná data"
	}
	parts := make([]string, 0, len(v))
	for k, msg := range v {
		parts = append(parts, k+": "+msg)
	}
	return strings.Join(parts, "; ")
}

// --- Stupeň: střední ---

// ParseEmail ořeže, převede na malá písmena a ověří tvar adresy.
func ParseEmail(s string) (Email, error) {
	// TODO
	return Email{}, nil
}

// ParseUsername ořeže a ověří uživatelské jméno.
func ParseUsername(s string) (Username, error) {
	// TODO
	return Username{}, nil
}

// Validate posbírá všechny chyby polí najednou; při úspěchu vrátí nil.
func (r CreateUserRequest) Validate() error {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---

// WriteProblem zapíše problem+json odpověď se zadaným statusem.
func WriteProblem(w http.ResponseWriter, status int, title, detail string, fields map[string]string) {
	// TODO
}

// ErrorHandler namapuje chybu na HTTP status a ProblemDetails.
func ErrorHandler(err error) (int, ProblemDetails) {
	// TODO
	return 0, ProblemDetails{}
}

// WriteError zapíše chybovou odpověď přes ErrorHandler a WriteProblem.
func WriteError(w http.ResponseWriter, err error) {
	// TODO
}

// DecodeAndValidate dekóduje JSON tělo a zavolá Validate.
func DecodeAndValidate[T Validator](r *http.Request) (T, error) {
	// TODO
	var zero T
	return zero, nil
}
