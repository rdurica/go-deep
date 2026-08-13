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

type Email struct{ value string }

type ValidationErrors map[string]string

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "neplatná data"
	}
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+": "+v[k])
	}
	return strings.Join(parts, "; ")
}

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
	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") || strings.Contains(domain, "..") {
		return Email{}, fmt.Errorf("%w: neplatná doména %q", ErrInvalidEmail, domain)
	}
	return Email{value: s}, nil
}

func (e Email) String() string { return e.value }
func (e Email) IsZero() bool   { return e.value == "" }

type Username struct{ value string }

func ParseUsername(s string) (Username, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Username{}, ErrEmptyUsername
	}
	if len([]rune(s)) < 3 || len([]rune(s)) > 20 {
		return Username{}, fmt.Errorf("%w: %q", ErrInvalidUsername, s)
	}
	first := s[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
		return Username{}, fmt.Errorf("%w: %q", ErrInvalidUsername, s)
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return Username{}, fmt.Errorf("%w: %q", ErrInvalidUsername, s)
	}
	return Username{value: s}, nil
}

func (u Username) String() string { return u.value }

type Validator interface {
	Validate() error
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Age      int    `json:"age"`
}

func (r CreateUserRequest) Validate() error {
	problems := ValidationErrors{}
	if _, err := ParseEmail(r.Email); err != nil {
		problems["email"] = "e-mail nemá platný tvar"
	}
	if _, err := ParseUsername(r.Username); err != nil {
		problems["username"] = "uživatelské jméno nemá platný tvar"
	}
	if r.Age < 13 || r.Age > 150 {
		problems["age"] = "věk musí být mezi 13 a 150"
	}
	if len(problems) == 0 {
		return nil
	}
	return problems
}

const ProblemContentType = "application/problem+json"

type ProblemDetails struct {
	Type   string            `json:"type"`
	Title  string            `json:"title"`
	Status int               `json:"status"`
	Detail string            `json:"detail,omitempty"`
	Errors map[string]string `json:"errors,omitempty"`
}

func WriteProblem(w http.ResponseWriter, status int, title, detail string, fields map[string]string) {
	w.Header().Set("Content-Type", ProblemContentType)
	w.WriteHeader(status)
	body := ProblemDetails{Type: "about:blank", Title: title, Status: status}
	if detail != "" {
		body.Detail = detail
	}
	if len(fields) > 0 {
		body.Errors = fields
	}
	_ = json.NewEncoder(w).Encode(body)
}

func ErrorHandler(err error) (int, ProblemDetails) {
	var problems ValidationErrors
	switch {
	case errors.As(err, &problems):
		return 422, ProblemDetails{Type: "about:blank", Title: "Neplatná data", Status: 422, Errors: problems}
	case errors.Is(err, ErrMalformedJSON):
		return 400, ProblemDetails{Type: "about:blank", Title: "Neplatný požadavek", Status: 400, Detail: "tělo není platný JSON"}
	case errors.Is(err, ErrNotFound):
		return 404, ProblemDetails{Type: "about:blank", Title: "Nenalezeno", Status: 404}
	case errors.Is(err, ErrConflict):
		return 409, ProblemDetails{Type: "about:blank", Title: "Konflikt", Status: 409}
	default:
		return 500, ProblemDetails{Type: "about:blank", Title: "Vnitřní chyba serveru", Status: 500}
	}
}

func WriteError(w http.ResponseWriter, err error) {
	status, body := ErrorHandler(err)
	WriteProblem(w, status, body.Title, body.Detail, body.Errors)
}

func DecodeAndValidate[T Validator](r *http.Request) (T, error) {
	var zero, v T
	dec := json.NewDecoder(io.LimitReader(r.Body, maxRequestBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return zero, fmt.Errorf("%w: %v", ErrMalformedJSON, err)
	}
	if dec.More() {
		return zero, fmt.Errorf("%w: extra data", ErrMalformedJSON)
	}
	if err := v.Validate(); err != nil {
		return zero, err
	}
	return v, nil
}
