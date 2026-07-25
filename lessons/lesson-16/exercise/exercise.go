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

// ToJSON serializuje uživatele do JSON.
func ToJSON(u User) ([]byte, error) {
	// TODO: úkol A
	return nil, nil
}

// FromJSON deserializuje uživatele z JSON a ověří povinná pole.
func FromJSON(data []byte) (User, error) {
	// TODO: úkol B
	return *new(User), nil
}

// DecodeEvent podle pole Kind dekóduje payload do odpovídajícího typu.
func DecodeEvent(data []byte) (any, error) {
	// TODO: úkol B
	return nil, nil
}

// MarshalJSON zapíše částku jako desetinné číslo.
func (c Cents) MarshalJSON() ([]byte, error) {
	// TODO: úkol C
	return nil, nil
}

// UnmarshalJSON načte částku z desetinného čísla nebo z řetězce.
func (c *Cents) UnmarshalJSON(data []byte) error {
	// TODO: úkol C
	return nil
}

// StrictDecode dekóduje data do v a odmítne neznámá pole i data navíc.
func StrictDecode(data []byte, v any) error {
	// TODO: úkol C
	return nil
}
