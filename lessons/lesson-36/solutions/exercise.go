// Package solutions obsahuje referenční řešení lekce 36.
package solutions

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
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

// maxEmailLength je horní mez délky adresy podle RFC 5321.
const maxEmailLength = 254

// maxRequestBody je strop pro tělo požadavku. Bez něj by útočník poslal
// nekonečný stream a server by ho poslušně dekódoval do paměti.
const maxRequestBody = 1 << 20

// Email je normalizovaná a ověřená e-mailová adresa.
//
// Pole je neexportované schválně: mimo tenhle balíček nelze vyrobit Email
// jinak než přes ParseEmail. Kdo drží Email, drží platnou adresu.
type Email struct {
	value string
}

// --- Stupeň: jednoduchý ---
// ParseEmail normalizuje a ověří e-mailovou adresu.
func ParseEmail(s string) (Email, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return Email{}, ErrEmptyEmail
	}
	if len(s) > maxEmailLength {
		return Email{}, fmt.Errorf("%w: délka %d nad limit %d", ErrInvalidEmail, len(s), maxEmailLength)
	}
	if strings.ContainsAny(s, " \t\r\n") {
		return Email{}, fmt.Errorf("%w: obsahuje bílý znak", ErrInvalidEmail)
	}
	if strings.Count(s, "@") != 1 {
		return Email{}, fmt.Errorf("%w: chybí právě jeden zavináč", ErrInvalidEmail)
	}
	local, domain, _ := strings.Cut(s, "@")
	if local == "" || domain == "" {
		return Email{}, fmt.Errorf("%w: prázdná část adresy", ErrInvalidEmail)
	}
	if !strings.Contains(domain, ".") ||
		strings.HasPrefix(domain, ".") ||
		strings.HasSuffix(domain, ".") ||
		strings.Contains(domain, "..") {
		return Email{}, fmt.Errorf("%w: neplatná doména %q", ErrInvalidEmail, domain)
	}
	return Email{value: s}, nil
}

// String vrací normalizovanou podobu adresy.
func (e Email) String() string { return e.value }

// IsZero vrací true pro nulovou hodnotu, tedy pro Email vyrobený bez ParseEmail.
func (e Email) IsZero() bool { return e.value == "" }

// --- Stupeň: střední ---
// Username je ověřené uživatelské jméno.
type Username struct {
	value string
}

// Meze délky uživatelského jména.
const (
	minUsernameLength = 3
	maxUsernameLength = 20
)

// ParseUsername ořízne bílé znaky a ověří tvar uživatelského jména.
func ParseUsername(s string) (Username, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Username{}, ErrEmptyUsername
	}
	runes := []rune(s)
	if len(runes) < minUsernameLength || len(runes) > maxUsernameLength {
		return Username{}, fmt.Errorf("%w: délka %d mimo rozsah %d–%d",
			ErrInvalidUsername, len(runes), minUsernameLength, maxUsernameLength)
	}
	if !isASCIILetter(runes[0]) {
		return Username{}, fmt.Errorf("%w: musí začínat písmenem", ErrInvalidUsername)
	}
	for _, r := range runes {
		if !isASCIILetter(r) && !(r >= '0' && r <= '9') && r != '_' {
			return Username{}, fmt.Errorf("%w: nepovolený znak %q", ErrInvalidUsername, r)
		}
	}
	return Username{value: s}, nil
}

func isASCIILetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// String vrací uživatelské jméno.
func (u Username) String() string { return u.value }

// ValidationErrors je mapa pole → důvod zamítnutí. Implementuje error.
type ValidationErrors map[string]string

// Error implementuje error. Výpis je deterministický, pole jsou seřazená.
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "validace selhala"
	}
	fields := make([]string, 0, len(v))
	for field := range v {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	var b strings.Builder
	b.WriteString("validace selhala: ")
	for i, field := range fields {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(field)
		b.WriteString(": ")
		b.WriteString(v[field])
	}
	return b.String()
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

// Meze věku, který systém přijme.
const (
	minAge = 13
	maxAge = 150
)

// --- Stupeň: obtížný ---
// Validate ověří všechna pole najednou a vrátí ValidationErrors,
// nebo nil, pokud je požadavek v pořádku.
func (r CreateUserRequest) Validate() error {
	problems := ValidationErrors{}
	if _, err := ParseEmail(r.Email); err != nil {
		problems["email"] = reasonOf(err)
	}
	if _, err := ParseUsername(r.Username); err != nil {
		problems["username"] = reasonOf(err)
	}
	if r.Age < minAge || r.Age > maxAge {
		problems["age"] = fmt.Sprintf("věk musí být mezi %d a %d", minAge, maxAge)
	}
	if len(problems) == 0 {
		// Pozor: `return problems` by vrátilo nenulový error interface
		// s prázdnou mapou uvnitř. Volající by pak marně testoval err != nil.
		return nil
	}
	return problems
}

// reasonOf vyrobí z chyby konstruktoru krátký důvod pro klienta.
func reasonOf(err error) string {
	switch {
	case errors.Is(err, ErrEmptyEmail):
		return "e-mail je povinný"
	case errors.Is(err, ErrInvalidEmail):
		return "e-mail nemá platný tvar"
	case errors.Is(err, ErrEmptyUsername):
		return "uživatelské jméno je povinné"
	case errors.Is(err, ErrInvalidUsername):
		return fmt.Sprintf("uživatelské jméno musí mít %d–%d znaků, začínat písmenem a obsahovat jen písmena, číslice a podtržítko",
			minUsernameLength, maxUsernameLength)
	default:
		return "hodnota není platná"
	}
}

// DecodeAndValidate dekóduje JSON tělo požadavku do T a zavolá jeho Validate.
func DecodeAndValidate[T Validator](r *http.Request) (T, error) {
	var zero, v T
	if r.Body == nil {
		return zero, fmt.Errorf("%w: chybí tělo požadavku", ErrMalformedJSON)
	}
	dec := json.NewDecoder(io.LimitReader(r.Body, maxRequestBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return zero, fmt.Errorf("%w: %v", ErrMalformedJSON, err)
	}
	if dec.More() {
		return zero, fmt.Errorf("%w: tělo obsahuje víc než jeden dokument", ErrMalformedJSON)
	}
	if err := v.Validate(); err != nil {
		return zero, err
	}
	return v, nil
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
	w.Header().Set("Content-Type", ProblemContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	body := ProblemDetails{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Detail: detail,
		Errors: fields,
	}
	// Hlavičky i status už jsou na drátě, takže s chybou zápisu se nedá
	// nic udělat. Jediná rozumná reakce je log, ne další WriteHeader.
	_ = json.NewEncoder(w).Encode(body)
}

// ErrorHandler přeloží chybu na HTTP status a tělo odpovědi.
//
// Vstup je vždy interní chyba se vším kontextem; výstup je to, co smí
// vidět klient. Neznámá chyba proto končí jako holá pětistovka.
func ErrorHandler(err error) (int, ProblemDetails) {
	problem := func(status int, title, detail string, fields map[string]string) (int, ProblemDetails) {
		return status, ProblemDetails{
			Type:   "about:blank",
			Title:  title,
			Status: status,
			Detail: detail,
			Errors: fields,
		}
	}

	var invalid ValidationErrors
	switch {
	case errors.As(err, &invalid):
		return problem(http.StatusUnprocessableEntity, "Neplatná data",
			"požadavek obsahuje neplatná pole", invalid)
	case errors.Is(err, ErrMalformedJSON):
		return problem(http.StatusBadRequest, "Neplatný požadavek",
			"tělo požadavku není platný JSON", nil)
	case errors.Is(err, ErrNotFound):
		return problem(http.StatusNotFound, "Nenalezeno",
			"požadovaný zdroj neexistuje", nil)
	case errors.Is(err, ErrConflict):
		return problem(http.StatusConflict, "Konflikt",
			"operace koliduje se současným stavem zdroje", nil)
	default:
		return problem(http.StatusInternalServerError, "Vnitřní chyba serveru",
			"požadavek se nepodařilo zpracovat", nil)
	}
}

// WriteError přeloží chybu přes ErrorHandler a zapíše ji přes WriteProblem.
func WriteError(w http.ResponseWriter, err error) {
	status, body := ErrorHandler(err)
	WriteProblem(w, status, body.Title, body.Detail, body.Errors)
}
