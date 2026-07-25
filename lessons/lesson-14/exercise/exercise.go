// Package exercise obsahuje cvičení lekce 14.
package exercise

import "errors"

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

// Divide dělí a a b. Při b == 0 vrací chybu obalující ErrDivideByZero.
func Divide(a, b int) (int, error) {
	panic("TODO: úkol A")
}

// Error implementuje interface error.
func (e *ValidationError) Error() string {
	panic("TODO: úkol B")
}

// ValidateUser ověří jméno a e-mail a vrátí spojení všech nalezených chyb.
func ValidateUser(name, email string) error {
	panic("TODO: úkol B")
}

// FieldsWithErrors posbírá jména polí ze všech ValidationError v řetězu chyb.
func FieldsWithErrors(err error) []string {
	panic("TODO: úkol B")
}

// Error implementuje interface error.
func (e *NotFoundError) Error() string {
	panic("TODO: úkol C")
}

// fetchFromStore simuluje čtení z úložiště. Neexportovaná — testy ji nevidí.
func fetchFromStore(id string) error {
	panic("TODO: úkol C")
}

// LoadUser načte uživatele a případnou chybu obalí kontextem.
func LoadUser(id string) error {
	panic("TODO: úkol C")
}

// IsNotFound vrací true, pokud kdekoli v řetězu chyb leží *NotFoundError.
func IsNotFound(err error) bool {
	panic("TODO: úkol C")
}
