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

// knownUsers je in-memory náhrada databáze.
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

// --- Stupeň: obtížný ---
// Error implementuje interface error.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Reason)
}

// --- Stupeň: střední ---
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

	// errors.Join nad prázdným seznamem vrací nil, takže žádné if navíc.
	return errors.Join(errs...)
}

// FieldsWithErrors posbírá jména polí ze všech ValidationError v řetězu chyb.
func FieldsWithErrors(err error) []string {
	var fields []string

	var walk func(error)
	walk = func(e error) {
		if e == nil {
			return
		}
		// Nejdřív se zanoř. errors.As se ptá na celý podstrom, takže by na
		// vnitřním uzlu našlo tu samou chybu podruhé.
		switch node := e.(type) {
		case interface{ Unwrap() []error }:
			for _, sub := range node.Unwrap() {
				walk(sub)
			}
			return
		case interface{ Unwrap() error }:
			walk(node.Unwrap())
			return
		}

		var ve *ValidationError
		if errors.As(e, &ve) {
			fields = append(fields, ve.Field)
		}
	}
	walk(err)

	return fields
}

// Error implementuje interface error.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("id %s not found", e.ID)
}

// fetchFromStore simuluje čtení z úložiště. Neexportovaná — testy ji nevidí.
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
