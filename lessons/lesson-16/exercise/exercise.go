// Package exercise obsahuje cvičení lekce 16.
package exercise

import (
	"encoding/json"
	"time"
)

// User je uživatel serializovaný do JSON.
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email,omitempty"`
	Active    bool      `json:"active"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Password  string    `json:"-"`
}

// Event je obálka události s dosud nedekódovaným payloadem.
type Event struct {
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

// UserCreated je payload události "user.created".
type UserCreated struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// UserDeleted je payload události "user.deleted".
type UserDeleted struct {
	ID     int    `json:"id"`
	Reason string `json:"reason,omitempty"`
}

// Cents je peněžní částka v celých centech.
// V JSON se objevuje jako desetinné číslo (1999 -> 19.99).
type Cents int64

// --- Stupeň: jednoduchý ---
// ToJSON serializuje uživatele do JSON.
// Password v JSON nikdy není; prázdný email a nil tags zmizí (omitempty).
// Active zůstane i při false; created_at je RFC 3339.
// Chybu z json.Marshal obal přes %w.
func ToJSON(u User) ([]byte, error) {
	// TODO
	return nil, nil
}

// FromJSON deserializuje uživatele z JSON a ověří povinná pole.
// ID musí být kladné, Name neprázdné i po oříznutí bílých znaků.
// Při chybě vrať nulový User a chybu; rozbitý JSON je taky chyba.
func FromJSON(data []byte) (User, error) {
	// TODO
	return *new(User), nil
}

// --- Stupeň: střední ---
// DecodeEvent podle pole Kind dekóduje payload do odpovídajícího typu.
// user.created → UserCreated, user.deleted → UserDeleted; vracej hodnotu, ne pointer.
// Neznámý kind, chybějící nebo nevalidní payload jsou chyby s názvem kindu v hlášce.
func DecodeEvent(data []byte) (any, error) {
	// TODO
	return nil, nil
}

// MarshalJSON zapíše částku jako desetinné číslo vždy se dvěma místy.
// 1999 → 19.99, 5 → 0.05, 0 → 0.00, -250 → -2.50.
func (c Cents) MarshalJSON() ([]byte, error) {
	// TODO
	return nil, nil
}

// --- Stupeň: obtížný ---
// UnmarshalJSON načte částku z desetinného čísla, řetězce nebo celého čísla (1 → 100 centů).
// null nechá hodnotu beze změny; round-trip musí být přesný (pozor na float).
func (c *Cents) UnmarshalJSON(data []byte) error {
	// TODO
	return nil
}

// StrictDecode dekóduje data do v přes json.Decoder s DisallowUnknownFields.
// Data za první JSON hodnotou jsou chyba.
func StrictDecode(data []byte, v any) error {
	// TODO
	return nil
}
