package solutions_test

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-14/solutions"
)

func TestDivide(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{10, 2, 5},
		{7, 2, 3},
		{-9, 3, -3},
		{0, 5, 0},
	}
	for _, tt := range tests {
		got, err := exercise.Divide(tt.a, tt.b)
		if err != nil {
			t.Errorf("Divide(%d, %d) = chyba %v, chci nil", tt.a, tt.b, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Divide(%d, %d) = %d, chci %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestDivideByZero(t *testing.T) {
	got, err := exercise.Divide(10, 0)
	if err == nil {
		t.Fatal("Divide(10, 0) vrátil nil chybu, chci ErrDivideByZero")
	}
	if got != 0 {
		t.Errorf("Divide(10, 0) = %d, chci 0", got)
	}
	if !errors.Is(err, exercise.ErrDivideByZero) {
		t.Errorf("errors.Is(err, ErrDivideByZero) = false pro %v", err)
	}
	if !strings.Contains(err.Error(), "division by zero") {
		t.Errorf("err.Error() = %q, chci text obsahující %q", err.Error(), "division by zero")
	}
	if err == exercise.ErrDivideByZero { // schválně porovnáváme identitu, ne errors.Is
		t.Error("chyba se má obalit kontextem přes %w, ne vracet holý sentinel")
	}
}

func TestValidationErrorText(t *testing.T) {
	ve := &exercise.ValidationError{Field: "email", Reason: "must contain @"}
	want := "invalid email: must contain @"
	if got := ve.Error(); got != want {
		t.Errorf("ValidationError.Error() = %q, chci %q", got, want)
	}

	// Musí splňovat error, a to s pointer receiverem.
	var err error = ve
	if err.Error() != want {
		t.Errorf("přes interface error = %q, chci %q", err.Error(), want)
	}
}

// validationFields vytáhne Field ze všech *ValidationError ve stromu přes errors.As —
// bez volání studentského FieldsWithErrors / Error().
func validationFields(err error) []string {
	var fields []string
	type multi interface{ Unwrap() []error }
	if m, ok := err.(multi); ok {
		for _, e := range m.Unwrap() {
			var ve *exercise.ValidationError
			if errors.As(e, &ve) {
				fields = append(fields, ve.Field)
			}
		}
		return fields
	}
	var ve *exercise.ValidationError
	if errors.As(err, &ve) {
		return []string{ve.Field}
	}
	return nil
}

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name       string
		userName   string
		email      string
		wantFields []string
	}{
		{"all ok", "radek", "radek@example.com", nil},
		{"missing name", "", "radek@example.com", []string{"name"}},
		{"missing email", "radek", "", []string{"email"}},
		{"email without at", "radek", "radek.example.com", []string{"email"}},
		{"both missing", "", "", []string{"name", "email"}},
		{"name missing, email bad", "", "nope", []string{"name", "email"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := exercise.ValidateUser(tt.userName, tt.email)

			if len(tt.wantFields) == 0 {
				if err != nil {
					t.Fatalf("ValidateUser(%q, %q) = %v, chci nil", tt.userName, tt.email, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ValidateUser(%q, %q) = nil, chci chyby pro %v",
					tt.userName, tt.email, tt.wantFields)
			}

			got := validationFields(err)
			if !slices.Equal(got, tt.wantFields) {
				t.Errorf("pole z ValidateUser = %v, chci %v", got, tt.wantFields)
			}
		})
	}
}

func TestValidateUserErrorsAs(t *testing.T) {
	err := exercise.ValidateUser("", "")

	var ve *exercise.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("errors.As(err, &ValidationError) = false pro %v", err)
	}
	if ve.Field != "name" {
		t.Errorf("první nalezená chyba je pro pole %q, chci %q", ve.Field, "name")
	}
}

func TestFieldsWithErrors(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		if got := exercise.FieldsWithErrors(nil); len(got) != 0 {
			t.Errorf("FieldsWithErrors(nil) = %v, chci prázdné", got)
		}
	})

	t.Run("foreign error", func(t *testing.T) {
		if got := exercise.FieldsWithErrors(errors.New("něco jiného")); len(got) != 0 {
			t.Errorf("FieldsWithErrors() = %v, chci prázdné", got)
		}
	})

	t.Run("wrapped ValidationError", func(t *testing.T) {
		inner := &exercise.ValidationError{Field: "age", Reason: "must be positive"}
		wrapped := fmt.Errorf("ověření vstupu: %w", inner)

		got := exercise.FieldsWithErrors(wrapped)
		if !slices.Equal(got, []string{"age"}) {
			t.Errorf("FieldsWithErrors() = %v, chci [age]", got)
		}
	})

	t.Run("mixed join", func(t *testing.T) {
		err := errors.Join(
			errors.New("nesouvisející"),
			&exercise.ValidationError{Field: "a", Reason: "r"},
			fmt.Errorf("kontext: %w", &exercise.ValidationError{Field: "b", Reason: "r"}),
		)

		got := exercise.FieldsWithErrors(err)
		if !slices.Equal(got, []string{"a", "b"}) {
			t.Errorf("FieldsWithErrors() = %v, chci [a b]", got)
		}
	})
}

func TestNotFoundErrorText(t *testing.T) {
	nf := &exercise.NotFoundError{ID: "u9"}
	want := "id u9 not found"
	if got := nf.Error(); got != want {
		t.Errorf("NotFoundError.Error() = %q, chci %q", got, want)
	}
}

func TestLoadUserExists(t *testing.T) {
	for _, id := range []string{"u1", "u2"} {
		if err := exercise.LoadUser(id); err != nil {
			t.Errorf("LoadUser(%q) = %v, chci nil", id, err)
		}
	}
}

// TestLoadUserWrapping je jádro lekce: obalená chyba si musí zachovat
// text obou vrstev A ZÁROVEŇ zůstat rozpoznatelná přes errors.As.
func TestLoadUserWrapping(t *testing.T) {
	err := exercise.LoadUser("u9")
	if err == nil {
		t.Fatal("LoadUser(\"u9\") = nil, chci NotFoundError")
	}

	msg := err.Error()
	if !strings.Contains(msg, "load user u9") {
		t.Errorf("err.Error() = %q, chybí vnější vrstva %q", msg, "load user u9")
	}
	if !strings.Contains(msg, "not found") {
		t.Errorf("err.Error() = %q, chybí vnitřní vrstva %q", msg, "not found")
	}

	var nf *exercise.NotFoundError
	if !errors.As(err, &nf) {
		t.Fatal("errors.As(err, &NotFoundError) = false — obaluješ chybu slovesem v místo w?")
	}
	if nf.ID != "u9" {
		t.Errorf("NotFoundError.ID = %q, chci %q", nf.ID, "u9")
	}

	if inner := errors.Unwrap(err); inner == nil {
		t.Error("errors.Unwrap(err) = nil, chyba není obalená")
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"foreign error", errors.New("boom"), false},
		{"direct NotFoundError", &exercise.NotFoundError{ID: "x"}, true},
		{"wrapped NotFoundError", fmt.Errorf("load user x: %w", &exercise.NotFoundError{ID: "x"}), true},
		{"wrapped twice", fmt.Errorf("handler: %w", fmt.Errorf("load: %w", &exercise.NotFoundError{ID: "nope"})), true},
		{"other error type", &exercise.ValidationError{Field: "name", Reason: "must not be empty"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound(%v) = %v, chci %v", tt.err, got, tt.want)
			}
		})
	}
}

// TestErrorTextConventions hlídá konvenci textů chyb: malé počáteční písmeno,
// bez tečky na konci a bez prefixu "error:".
func TestErrorTextConventions(t *testing.T) {
	_, divErr := exercise.Divide(1, 0)
	errs := []error{
		exercise.ErrDivideByZero,
		divErr,
		&exercise.ValidationError{Field: "f", Reason: "r"},
		&exercise.NotFoundError{ID: "x"},
		fmt.Errorf("load user nope: %w", &exercise.NotFoundError{ID: "nope"}),
	}
	for _, err := range errs {
		msg := err.Error()
		if msg == "" {
			t.Error("prázdný text chyby")
			continue
		}
		if strings.HasSuffix(msg, ".") {
			t.Errorf("text chyby %q končí tečkou", msg)
		}
		if r := []rune(msg)[0]; r >= 'A' && r <= 'Z' {
			t.Errorf("text chyby %q začíná velkým písmenem", msg)
		}
		if strings.HasPrefix(strings.ToLower(msg), "error:") {
			t.Errorf("text chyby %q začíná zbytečným %q", msg, "error:")
		}
	}
}
