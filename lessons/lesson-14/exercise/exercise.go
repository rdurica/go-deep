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

// knownUsers je in-memory náhrada databáze.
var knownUsers = map[string]struct{}{
	"u1": {},
	"u2": {},
}

// --- Stupeň: jednoduchý ---

// Divide dělí a/b celočíselně. Při b == 0 vrací 0 a chybu obalující ErrDivideByZero
// s kontextem o dělenci. Vrácená chyba nesmí být holý sentinel.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ — vrací sentinel bez obalení přes %w.
// Najdi chybu a oprav.
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return a / b, nil
}

// --- Stupeň: střední ---

// Error implementuje error ve tvaru "invalid <Field>: <Reason>".
func (e *ValidationError) Error() string {
	// TODO
	return ""
}

// ValidateUser posbírá všechny problémy a vrátí je přes errors.Join.
// Prázdné name → ValidationError{name, must not be empty}; prázdný email → totéž;
// email bez @ → ValidationError{email, must contain @}. Pořadí: name, email.
// Bez problémů vrací nil.
func ValidateUser(name, email string) error {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---

// Error implementuje error ve tvaru "id <ID> not found". Hotová — neimplementuj.
func (e *NotFoundError) Error() string {
	return "id " + e.ID + " not found"
}

// fetchFromStore simuluje úložiště. Hotová — neimplementuj.
func fetchFromStore(id string) error {
	if _, ok := knownUsers[id]; !ok {
		return &NotFoundError{ID: id}
	}
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
