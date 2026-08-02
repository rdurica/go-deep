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

// --- Stupeň: jednoduchý ---
// ParseEmail ořízne bílé znaky, převede na malá písmena a ověří tvar.
// Prázdný → ErrEmptyEmail; délka >254, bílý znak uvnitř, chybějící/vícenásobný
// @, prázdná lokální část/doména, doména bez tečky nebo s tečkou na okraji
// → chyba obalující ErrInvalidEmail. Při chybě nulová hodnota.
func ParseEmail(s string) (Email, error) {
	// TODO
	return *new(Email), nil
}

// String vrací normalizovanou podobu adresy.
func (e Email) String() string {
	// TODO
	return ""
}

// IsZero vrací true pro nulovou hodnotu Email.
func (e Email) IsZero() bool {
	// TODO
	return false
}

// --- Stupeň: střední ---
// Username je ověřené uživatelské jméno.
type Username struct {
	value string
}

// ParseUsername ořízne bílé znaky. Prázdný → ErrEmptyUsername; 3–20 runů,
// jen ASCII písmena, číslice a _, první znak písmeno → jinak ErrInvalidUsername.
// Chyby obal %w pro errors.Is.
func ParseUsername(s string) (Username, error) {
	// TODO
	return *new(Username), nil
}

// String vrací uživatelské jméno.
func (u Username) String() string {
	// TODO
	return ""
}

// ValidationErrors je mapa pole → důvod zamítnutí. Implementuje error.
type ValidationErrors map[string]string

// Error implementuje error. Pole abecedně seřazená (deterministický výpis).
// Prázdná mapa dá rozumnou větu, ne prázdný string.
func (v ValidationErrors) Error() string {
	// TODO
	return ""
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

// --- Stupeň: obtížný ---
// Validate ověří všechna pole najednou. Klíče email, username, age.
// E-mail a jméno přes Parse z části A; věk 13–150. Vrať nil, ne typovaný nil mapy.
func (r CreateUserRequest) Validate() error {
	// TODO
	return nil
}

// DecodeAndValidate dekóduje JSON do T s LimitReader, DisallowUnknownFields,
// odmítne druhý dokument a chybějící tělo. Dekódování → ErrMalformedJSON;
// validace propusť pro errors.As. Při chybě nulová hodnota T.
func DecodeAndValidate[T Validator](r *http.Request) (T, error) {
	// TODO
	return *new(T), nil
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

// WriteProblem zapíše application/problem+json s Type about:blank.
// Prázdný detail ani prázdná fields se do těla nedostanou (omitempty).
func WriteProblem(w http.ResponseWriter, status int, title, detail string, fields map[string]string) {
	// TODO
}

// ErrorHandler mapuje chyby: ValidationErrors→422, ErrMalformedJSON→400,
// ErrNotFound→404, ErrConflict→409, nil i ostatní→500. body.Status = status.
// Funguje na obalené chyby (errors.Is/As).
func ErrorHandler(err error) (int, ProblemDetails) {
	// TODO
	return 0, *new(ProblemDetails)
}

// WriteError přeloží chybu přes ErrorHandler a zapíše přes WriteProblem.
// Text interní chyby se do těla nedostane.
func WriteError(w http.ResponseWriter, err error) {
	// TODO
}
