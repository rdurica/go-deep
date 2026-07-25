// Package exercise obsahuje cvičení lekce 36.
package exercise

import (
	"errors"
	"net/http"
)

// Doménové chyby, které umí ErrorHandler přeložit na HTTP status.
var (
	// ErrNotFound říká, že požadovaný zdroj neexistuje.
	ErrNotFound = errors.New("zdroj nenalezen")
	// ErrConflict říká, že operace koliduje se současným stavem.
	ErrConflict = errors.New("konflikt se stavem zdroje")
	// ErrMalformedJSON říká, že tělo požadavku nešlo dekódovat.
	ErrMalformedJSON = errors.New("neplatné JSON tělo")
)

// Chyby konstruktorů hodnotových typů.
var (
	// ErrEmptyEmail vrací ParseEmail pro prázdný vstup.
	ErrEmptyEmail = errors.New("e-mail je prázdný")
	// ErrInvalidEmail vrací ParseEmail pro vstup, který není e-mail.
	ErrInvalidEmail = errors.New("e-mail má neplatný tvar")
	// ErrEmptyUsername vrací ParseUsername pro prázdný vstup.
	ErrEmptyUsername = errors.New("uživatelské jméno je prázdné")
	// ErrInvalidUsername vrací ParseUsername pro vstup mimo povolený tvar.
	ErrInvalidUsername = errors.New("uživatelské jméno má neplatný tvar")
)

// Email je normalizovaná a ověřená e-mailová adresa.
//
// Pole je neexportované schválně: mimo tenhle balíček nelze vyrobit Email
// jinak než přes ParseEmail. Kdo drží Email, drží platnou adresu.
type Email struct {
	value string
}

// ParseEmail normalizuje a ověří e-mailovou adresu.
func ParseEmail(s string) (Email, error) {
	panic("TODO: úkol A")
}

// String vrací normalizovanou podobu adresy.
func (e Email) String() string {
	panic("TODO: úkol A")
}

// IsZero vrací true pro nulovou hodnotu, tedy pro Email vyrobený bez ParseEmail.
func (e Email) IsZero() bool {
	panic("TODO: úkol A")
}

// Username je ověřené uživatelské jméno.
type Username struct {
	value string
}

// ParseUsername ořízne bílé znaky a ověří tvar uživatelského jména.
func ParseUsername(s string) (Username, error) {
	panic("TODO: úkol A")
}

// String vrací uživatelské jméno.
func (u Username) String() string {
	panic("TODO: úkol A")
}

// ValidationErrors je mapa pole → důvod zamítnutí. Implementuje error.
type ValidationErrors map[string]string

// Error implementuje error. Výpis je deterministický, pole jsou seřazená.
func (v ValidationErrors) Error() string {
	panic("TODO: úkol B")
}

// Validator umí ověřit sám sebe. Splňují ho DTO na hranici systému.
type Validator interface {
	Validate() error
}

// CreateUserRequest je DTO požadavku na založení uživatele.
type CreateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Age      int    `json:"age"`
}

// Validate ověří všechna pole najednou a vrátí ValidationErrors,
// nebo nil, pokud je požadavek v pořádku.
func (r CreateUserRequest) Validate() error {
	panic("TODO: úkol B")
}

// DecodeAndValidate dekóduje JSON tělo požadavku do T a zavolá jeho Validate.
func DecodeAndValidate[T Validator](r *http.Request) (T, error) {
	panic("TODO: úkol B")
}

// ProblemContentType je hodnota hlavičky Content-Type podle RFC 7807.
const ProblemContentType = "application/problem+json"

// ProblemDetails je tělo chybové odpovědi podle RFC 7807.
type ProblemDetails struct {
	Type   string            `json:"type"`
	Title  string            `json:"title"`
	Status int               `json:"status"`
	Detail string            `json:"detail,omitempty"`
	Errors map[string]string `json:"errors,omitempty"`
}

// WriteProblem zapíše odpověď ve tvaru problem+json.
func WriteProblem(w http.ResponseWriter, status int, title, detail string, fields map[string]string) {
	panic("TODO: úkol C")
}

// ErrorHandler přeloží chybu na HTTP status a tělo odpovědi.
func ErrorHandler(err error) (int, ProblemDetails) {
	panic("TODO: úkol C")
}

// WriteError přeloží chybu přes ErrorHandler a zapíše ji přes WriteProblem.
func WriteError(w http.ResponseWriter, err error) {
	panic("TODO: úkol C")
}
