// Package solutions obsahuje referenční řešení lekce 14.
package solutions

import (
	"errors"
	"fmt"
	"strings"
)

// ErrDivideByZero je sentinel chyba pro dělení nulou.
var ErrDivideByZero = errors.New("division by zero")

// ValidationError popisuje jedno neplatné pole vstupu.
type ValidationError struct {
	Field  string
	Reason string
}

// NotFoundError říká, že záznam s daným ID neexistuje.
type NotFoundError struct {
	ID string
}

var knownUsers = map[string]struct{}{
	"u1": {},
	"u2": {},
}

// --- Stupeň: jednoduchý ---

// Divide dělí a a b. Při b == 0 vrací chybu obalující ErrDivideByZero.
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("divide %d: %w", a, ErrDivideByZero)
	}
	return a / b, nil
}

// --- Stupeň: střední ---

// Error implementuje interface error.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Reason)
}

// ValidateUser ověří jméno a e-mail a vrátí spojení všech nalezených chyb.
func ValidateUser(name, email string) error {
	var errs []error

	if name == "" {
		errs = append(errs, &ValidationError{Field: "name", Reason: "must not be empty"})
	}
	switch {
	case email == "":
		errs = append(errs, &ValidationError{Field: "email", Reason: "must not be empty"})
	case !strings.Contains(email, "@"):
		errs = append(errs, &ValidationError{Field: "email", Reason: "must contain @"})
	}

	return errors.Join(errs...)
}

// --- Stupeň: obtížný ---

// Error implementuje interface error.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("id %s not found", e.ID)
}

func fetchFromStore(id string) error {
	if _, ok := knownUsers[id]; !ok {
		return &NotFoundError{ID: id}
	}
	return nil
}

// LoadUser načte uživatele a případnou chybu obalí kontextem.
func LoadUser(id string) error {
	if err := fetchFromStore(id); err != nil {
		return fmt.Errorf("load user %s: %w", id, err)
	}
	return nil
}

// IsNotFound vrací true, pokud kdekoli v řetězu chyb leží *NotFoundError.
func IsNotFound(err error) bool {
	var nf *NotFoundError
	return errors.As(err, &nf)
}
