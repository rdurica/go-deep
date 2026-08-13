// Package exercise obsahuje cvičení lekce 16.
package exercise

import (
	"encoding/json"
	"errors"
	"fmt"
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

// --- Stupeň: jednoduchý ---

// FromJSON deserializuje uživatele z JSON a ověří povinná pole.
// ID musí být kladné, Name neprázdné i po oříznutí bílých znaků.
// Při chybě vrať nulový User a chybu; rozbitý JSON je taky chyba.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Jméno nekontroluje po oříznutí bílých znaků.
// Najdi chybu a oprav — testy před opravou padají.
func FromJSON(data []byte) (User, error) {
	var u User
	if err := json.Unmarshal(data, &u); err != nil {
		return User{}, fmt.Errorf("unmarshal user: %w", err)
	}
	if u.ID <= 0 {
		return User{}, fmt.Errorf("user: id musí být kladné, je %d", u.ID)
	}
	if u.Name == "" {
		return User{}, errors.New("user: name nesmí být prázdné")
	}
	return u, nil
}

// --- Stupeň: střední ---

// DecodeEvent podle pole Kind dekóduje payload do odpovídajícího typu.
// user.created → UserCreated, user.deleted → UserDeleted; vracej hodnotu, ne pointer.
// Neznámý kind, chybějící nebo nevalidní payload jsou chyby s názvem kindu v hlášce.
func DecodeEvent(data []byte) (any, error) {
	// TODO
	return nil, nil
}

// --- Stupeň: obtížný ---

// StrictDecode dekóduje data do v přes json.Decoder s DisallowUnknownFields.
// Data za první JSON hodnotou jsou chyba.
func StrictDecode(data []byte, v any) error {
	// TODO
	return nil
}
