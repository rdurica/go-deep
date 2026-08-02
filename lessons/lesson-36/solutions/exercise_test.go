package solutions_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-36/solutions"
)

func TestParseEmailNormalizes(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"radek@example.com", "radek@example.com"},
		{"  radek@example.com  ", "radek@example.com"},
		{"Radek@Example.COM", "radek@example.com"},
		{"\tRADEK@EXAMPLE.CO.UK\n", "radek@example.co.uk"},
		{"a.b+tag@sub.example.org", "a.b+tag@sub.example.org"},
	}
	for _, tt := range tests {
		got, err := exercise.ParseEmail(tt.in)
		if err != nil {
			t.Errorf("ParseEmail(%q) = chyba %v, chci platný e-mail", tt.in, err)
			continue
		}
		if got.String() != tt.want {
			t.Errorf("ParseEmail(%q).String() = %q, chci %q", tt.in, got.String(), tt.want)
		}
	}
}

func TestParseEmailRandomData(t *testing.T) {
	// Zabraňuje řešení typu "vrať natvrdo radek@example.com".
	rnd := rand.New(rand.NewSource(1))
	const letters = "abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < 50; i++ {
		local := make([]byte, 1+rnd.Intn(8))
		for j := range local {
			local[j] = letters[rnd.Intn(len(letters))]
		}
		want := string(local) + "@example.com"
		got, err := exercise.ParseEmail(strings.ToUpper(want))
		if err != nil {
			t.Fatalf("ParseEmail(%q) = chyba %v, chci platný e-mail", want, err)
		}
		if got.String() != want {
			t.Fatalf("ParseEmail(%q).String() = %q, chci %q", want, got.String(), want)
		}
	}
}

func TestParseEmailRejects(t *testing.T) {
	tests := []struct {
		in      string
		wantErr error
	}{
		{"", exercise.ErrEmptyEmail},
		{"   ", exercise.ErrEmptyEmail},
		{"\t\n", exercise.ErrEmptyEmail},
		{"radek", exercise.ErrInvalidEmail},
		{"radek@", exercise.ErrInvalidEmail},
		{"@example.com", exercise.ErrInvalidEmail},
		{"radek@@example.com", exercise.ErrInvalidEmail},
		{"radek@a@b.com", exercise.ErrInvalidEmail},
		{"radek@example", exercise.ErrInvalidEmail},
		{"radek@.com", exercise.ErrInvalidEmail},
		{"radek@example.", exercise.ErrInvalidEmail},
		{"radek@exa..mple.com", exercise.ErrInvalidEmail},
		{"ra dek@example.com", exercise.ErrInvalidEmail},
		{strings.Repeat("a", 250) + "@example.com", exercise.ErrInvalidEmail},
	}
	for _, tt := range tests {
		got, err := exercise.ParseEmail(tt.in)
		if !errors.Is(err, tt.wantErr) {
			t.Errorf("ParseEmail(%q) = chyba %v, chci %v", tt.in, err, tt.wantErr)
		}
		if !got.IsZero() {
			t.Errorf("ParseEmail(%q) vrátil při chybě nenulový Email %q", tt.in, got.String())
		}
	}
}

func TestEmailZeroValue(t *testing.T) {
	var zero exercise.Email
	if !zero.IsZero() {
		t.Error("nulová hodnota Email má být IsZero() == true")
	}
	if zero.String() != "" {
		t.Errorf("nulová hodnota Email má String() == \"\", má %q", zero.String())
	}
	parsed, err := exercise.ParseEmail("radek@example.com")
	if err != nil {
		t.Fatalf("ParseEmail = chyba %v", err)
	}
	if parsed.IsZero() {
		t.Error("naparsovaný Email nemá být IsZero()")
	}
}

func TestEmailIsComparable(t *testing.T) {
	a, err := exercise.ParseEmail("Radek@Example.com")
	if err != nil {
		t.Fatalf("ParseEmail = chyba %v", err)
	}
	b, err := exercise.ParseEmail("  radek@example.com ")
	if err != nil {
		t.Fatalf("ParseEmail = chyba %v", err)
	}
	if a != b {
		t.Errorf("normalizované adresy se mají rovnat: %q != %q", a.String(), b.String())
	}
}

func TestParseUsername(t *testing.T) {
	ok := []struct {
		in   string
		want string
	}{
		{"radek", "radek"},
		{"  radek  ", "radek"},
		{"Radek_99", "Radek_99"},
		{"abc", "abc"},
		{strings.Repeat("a", 20), strings.Repeat("a", 20)},
	}
	for _, tt := range ok {
		got, err := exercise.ParseUsername(tt.in)
		if err != nil {
			t.Errorf("ParseUsername(%q) = chyba %v, chci platné jméno", tt.in, err)
			continue
		}
		if got.String() != tt.want {
			t.Errorf("ParseUsername(%q).String() = %q, chci %q", tt.in, got.String(), tt.want)
		}
	}

	bad := []struct {
		in      string
		wantErr error
	}{
		{"", exercise.ErrEmptyUsername},
		{"   ", exercise.ErrEmptyUsername},
		{"ab", exercise.ErrInvalidUsername},
		{strings.Repeat("a", 21), exercise.ErrInvalidUsername},
		{"9radek", exercise.ErrInvalidUsername},
		{"_radek", exercise.ErrInvalidUsername},
		{"ra dek", exercise.ErrInvalidUsername},
		{"radek-99", exercise.ErrInvalidUsername},
		{"radeček", exercise.ErrInvalidUsername},
	}
	for _, tt := range bad {
		if _, err := exercise.ParseUsername(tt.in); !errors.Is(err, tt.wantErr) {
			t.Errorf("ParseUsername(%q) = chyba %v, chci %v", tt.in, err, tt.wantErr)
		}
	}
}

func TestValidationErrorsError(t *testing.T) {
	ve := exercise.ValidationErrors{
		"username": "je povinné",
		"email":    "nemá platný tvar",
		"age":      "je mimo rozsah",
	}
	first := ve.Error()
	for i := 0; i < 20; i++ {
		if got := ve.Error(); got != first {
			t.Fatalf("Error() není deterministický: %q vs %q", got, first)
		}
	}
	for _, field := range []string{"age", "email", "username"} {
		if !strings.Contains(first, field) {
			t.Errorf("Error() = %q, chybí pole %q", first, field)
		}
	}
	iAge := strings.Index(first, "age")
	iEmail := strings.Index(first, "email")
	iUser := strings.Index(first, "username")
	if !(iAge < iEmail && iEmail < iUser) {
		t.Errorf("Error() = %q, chci pole seřazená abecedně", first)
	}

	var asErr error = exercise.ValidationErrors{"email": "chybí"}
	var target exercise.ValidationErrors
	if !errors.As(asErr, &target) {
		t.Fatal("ValidationErrors musí jít vytáhnout přes errors.As")
	}
	if target["email"] != "chybí" {
		t.Errorf("errors.As vrátil %v, chci mapu s klíčem email", target)
	}
}

func TestCreateUserRequestValidateOK(t *testing.T) {
	req := exercise.CreateUserRequest{Email: "Radek@Example.com", Username: "radek", Age: 40}
	if err := req.Validate(); err != nil {
		t.Errorf("Validate() = %v, chci nil", err)
	}
}

func TestCreateUserRequestCollectsAllErrors(t *testing.T) {
	req := exercise.CreateUserRequest{Email: "nope", Username: "x", Age: 3}
	err := req.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, chci ValidationErrors")
	}
	var ve exercise.ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("Validate() = %T, chci ValidationErrors", err)
	}
	want := []string{"age", "email", "username"}
	if len(ve) != len(want) {
		t.Errorf("Validate() vrátil %d chyb (%v), chci %d — validace nesmí končit na první chybě",
			len(ve), ve, len(want))
	}
	for _, field := range want {
		reason, ok := ve[field]
		if !ok {
			t.Errorf("chybí chyba pro pole %q, mám %v", field, ve)
			continue
		}
		if strings.TrimSpace(reason) == "" {
			t.Errorf("pole %q má prázdný důvod", field)
		}
	}
}

func TestCreateUserRequestIndividualErrors(t *testing.T) {
	tests := []struct {
		name  string
		req   exercise.CreateUserRequest
		field string
	}{
		{"empty email", exercise.CreateUserRequest{Username: "radek", Age: 30}, "email"},
		{"empty name", exercise.CreateUserRequest{Email: "a@b.cz", Age: 30}, "username"},
		{"low age", exercise.CreateUserRequest{Email: "a@b.cz", Username: "radek", Age: 12}, "age"},
		{"high age", exercise.CreateUserRequest{Email: "a@b.cz", Username: "radek", Age: 151}, "age"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ve exercise.ValidationErrors
			if !errors.As(tt.req.Validate(), &ve) {
				t.Fatalf("Validate() nevrátil ValidationErrors pro %+v", tt.req)
			}
			if len(ve) != 1 {
				t.Fatalf("Validate() = %v, chci právě jednu chybu u pole %q", ve, tt.field)
			}
			if _, ok := ve[tt.field]; !ok {
				t.Errorf("Validate() = %v, chci chybu u pole %q", ve, tt.field)
			}
		})
	}
}

func postJSON(body string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
}

func TestDecodeAndValidateOK(t *testing.T) {
	req := postJSON(`{"email":"Radek@Example.com","username":"radek","age":40}`)
	got, err := exercise.DecodeAndValidate[exercise.CreateUserRequest](req)
	if err != nil {
		t.Fatalf("DecodeAndValidate = chyba %v, chci úspěch", err)
	}
	want := exercise.CreateUserRequest{Email: "Radek@Example.com", Username: "radek", Age: 40}
	if got != want {
		t.Errorf("DecodeAndValidate = %+v, chci %+v", got, want)
	}
}

func TestDecodeAndValidateBadJSON(t *testing.T) {
	tests := map[string]string{
		"broken JSON":       `{"email":`,
		"špatný typ pole":    `{"email":"a@b.cz","username":"radek","age":"čtyřicet"}`,
		"unknown field":       `{"email":"a@b.cz","username":"radek","age":40,"role":"admin"}`,
		"dva dokumenty":      `{"email":"a@b.cz","username":"radek","age":40}{"email":"x@y.cz"}`,
		"empty body":       ``,
		"pole místo objektu": `[1,2,3]`,
	}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := exercise.DecodeAndValidate[exercise.CreateUserRequest](postJSON(body))
			if !errors.Is(err, exercise.ErrMalformedJSON) {
				t.Errorf("DecodeAndValidate(%q) = %v, chci ErrMalformedJSON", body, err)
			}
		})
	}
}

func TestDecodeAndValidateInvalidData(t *testing.T) {
	req := postJSON(`{"email":"nope","username":"x","age":3}`)
	_, err := exercise.DecodeAndValidate[exercise.CreateUserRequest](req)
	var ve exercise.ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("DecodeAndValidate = %v (%T), chci ValidationErrors", err, err)
	}
	if len(ve) != 3 {
		t.Errorf("DecodeAndValidate vrátil %v, chci chyby u tří polí", ve)
	}
}

func decodeProblem(t *testing.T, rec *httptest.ResponseRecorder) exercise.ProblemDetails {
	t.Helper()
	if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Errorf("Content-Type = %q, chci %q", got, "application/problem+json")
	}
	var p exercise.ProblemDetails
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("tělo není platný JSON (%v): %s", err, rec.Body.String())
	}
	return p
}

func TestWriteProblem(t *testing.T) {
	rec := httptest.NewRecorder()
	exercise.WriteProblem(rec, http.StatusUnprocessableEntity, "Neplatná data", "zkontroluj pole",
		map[string]string{"email": "nemá platný tvar"})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, chci %d", rec.Code, http.StatusUnprocessableEntity)
	}
	p := decodeProblem(t, rec)
	if p.Status != http.StatusUnprocessableEntity {
		t.Errorf("body.status = %d, chci %d", p.Status, http.StatusUnprocessableEntity)
	}
	if p.Title != "Neplatná data" {
		t.Errorf("body.title = %q, chci %q", p.Title, "Neplatná data")
	}
	if p.Detail != "zkontroluj pole" {
		t.Errorf("body.detail = %q, chci %q", p.Detail, "zkontroluj pole")
	}
	if p.Type == "" {
		t.Error("body.type musí být vyplněný (typicky about:blank)")
	}
	if p.Errors["email"] != "nemá platný tvar" {
		t.Errorf("body.errors = %v, chci klíč email", p.Errors)
	}

	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("tělo není JSON objekt: %v", err)
	}
	for _, key := range []string{"type", "title", "status", "detail", "errors"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("v těle chybí pole %q, mám %v", key, raw)
		}
	}
}

func TestWriteProblemWithoutFields(t *testing.T) {
	rec := httptest.NewRecorder()
	exercise.WriteProblem(rec, http.StatusNotFound, "Nenalezeno", "", nil)

	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("tělo není JSON objekt: %v", err)
	}
	if _, ok := raw["errors"]; ok {
		t.Errorf("prázdné errors se nemá serializovat (chybí omitempty): %v", raw)
	}
	if _, ok := raw["detail"]; ok {
		t.Errorf("prázdný detail se nemá serializovat (chybí omitempty): %v", raw)
	}
}

func TestErrorHandlerMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"not found", exercise.ErrNotFound, http.StatusNotFound},
		{"wrapped not found", fmt.Errorf("repozitář: %w", exercise.ErrNotFound), http.StatusNotFound},
		{"conflict", exercise.ErrConflict, http.StatusConflict},
		{"wrapped conflict", fmt.Errorf("uložení: %w", exercise.ErrConflict), http.StatusConflict},
		{"broken JSON", fmt.Errorf("%w: EOF", exercise.ErrMalformedJSON), http.StatusBadRequest},
		{"validation", exercise.ValidationErrors{"email": "chybí"}, http.StatusUnprocessableEntity},
		{"unknown error", errors.New("cokoli jiného"), http.StatusInternalServerError},
		{"nil", nil, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := exercise.ErrorHandler(tt.err)
			if status != tt.wantStatus {
				t.Errorf("ErrorHandler(%v) = status %d, chci %d", tt.err, status, tt.wantStatus)
			}
			if body.Status != status {
				t.Errorf("body.status = %d, chci %d (musí souhlasit s HTTP statusem)", body.Status, status)
			}
			if body.Title == "" {
				t.Error("body.title nesmí být prázdný")
			}
		})
	}
}

func TestErrorHandlerForwardsValidationFields(t *testing.T) {
	ve := exercise.ValidationErrors{"email": "nemá platný tvar", "age": "je mimo rozsah"}
	_, body := exercise.ErrorHandler(fmt.Errorf("hranice: %w", ve))
	if len(body.Errors) != 2 {
		t.Fatalf("body.errors = %v, chci obě pole", body.Errors)
	}
	if body.Errors["email"] != "nemá platný tvar" {
		t.Errorf("body.errors[email] = %q, chci %q", body.Errors["email"], "nemá platný tvar")
	}
}

func TestErrorHandlerDoesNotLeakInternalDetails(t *testing.T) {
	secrets := []string{
		"pq: password authentication failed for user \"orders\"",
		"dial tcp 10.0.0.17:5432: connect: connection refused",
		"/home/deploy/app/internal/repo/pg.go:118",
	}
	for _, secret := range secrets {
		internal := fmt.Errorf("uložení objednávky: %w", errors.New(secret))
		status, body := exercise.ErrorHandler(internal)
		if status != http.StatusInternalServerError {
			t.Errorf("ErrorHandler(%v) = %d, chci 500", internal, status)
		}
		blob := body.Title + " " + body.Detail + " " + fmt.Sprint(body.Errors)
		if strings.Contains(blob, secret) {
			t.Errorf("odpověď prozradila interní detail %q: %q", secret, blob)
		}
	}
}

func TestWriteErrorViaHTTP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := exercise.DecodeAndValidate[exercise.CreateUserRequest](r)
		if err != nil {
			exercise.WriteError(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"username": req.Username})
	})

	t.Run("valid request", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, postJSON(`{"email":"a@b.cz","username":"radek","age":40}`))
		if rec.Code != http.StatusCreated {
			t.Errorf("status = %d, chci 201 (tělo: %s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("invalid data", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, postJSON(`{"email":"nope","username":"x","age":3}`))
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, chci 422 (tělo: %s)", rec.Code, rec.Body.String())
		}
		p := decodeProblem(t, rec)
		if len(p.Errors) != 3 {
			t.Errorf("body.errors = %v, chci tři pole", p.Errors)
		}
	})

	t.Run("broken JSON", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, postJSON(`{"email":`))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, chci 400 (tělo: %s)", rec.Code, rec.Body.String())
		}
		p := decodeProblem(t, rec)
		if strings.Contains(strings.ToLower(p.Detail), "unexpected end of json") {
			t.Errorf("detail prozrazuje interní chybu dekodéru: %q", p.Detail)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		exercise.WriteError(rec, fmt.Errorf("repo: %w", errors.New("tajný stack trace")))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, chci 500", rec.Code)
		}
		decodeProblem(t, rec)
		if strings.Contains(rec.Body.String(), "tajný stack trace") {
			t.Errorf("tělo odpovědi prozradilo interní chybu: %s", rec.Body.String())
		}
	})
}
