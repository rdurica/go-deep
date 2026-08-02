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

// --- Stupeň: jednoduchý ---
// Divide dělí a/b celočíselně. Při b == 0 vrací 0 a chybu obalující ErrDivideByZero
// s kontextem o dělenci. Vrácená chyba nesmí být holý sentinel.
func Divide(a, b int) (int, error) {
	// TODO
	return 0, nil
}

// --- Stupeň: obtížný ---
// Error implementuje error ve tvaru "invalid <Field>: <Reason>".
func (e *ValidationError) Error() string {
	// TODO
	return ""
}

// --- Stupeň: střední ---
// ValidateUser posbírá všechny problémy a vrátí je přes errors.Join.
// Prázdné name → ValidationError{name, must not be empty}; prázdný email → totéž;
// email bez @ → ValidationError{email, must contain @}. Pořadí: name, email.
// Bez problémů vrací nil.
func ValidateUser(name, email string) error {
	// TODO
	return nil
}

// FieldsWithErrors projde strom chyb (včetně errors.Join) a vrátí jména polí
// všech *ValidationError v pořadí. Použij errors.As / Unwrap — nesmí duplikovat
// stejnou chybu. Nil a chyba bez ValidationError → prázdný výsledek.
func FieldsWithErrors(err error) []string {
	// TODO
	return nil
}

// Error implementuje error ve tvaru "id <ID> not found".
func (e *NotFoundError) Error() string {
	// TODO
	return ""
}

// fetchFromStore simuluje úložiště. Zná "u1" a "u2" (nil), jinak &NotFoundError{ID: id}.
func fetchFromStore(id string) error {
	// TODO
	return nil
}

// LoadUser zavolá fetchFromStore a chybu obalí: fmt.Errorf("load user %s: %w", id, err),
// aby dál fungovalo errors.Is / errors.As na NotFoundError.
func LoadUser(id string) error {
	// TODO
	return nil
}

// IsNotFound vrací true, pokud kdekoli v řetězu leží *NotFoundError.
// Nil a cizí chyby → false.
func IsNotFound(err error) bool {
	// TODO
	return false
}
