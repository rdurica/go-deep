// Package solutions obsahuje referenční řešení lekce 16.
package solutions

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
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
	data, err := json.Marshal(u)
	if err != nil {
		return nil, fmt.Errorf("marshal user: %w", err)
	}
	return data, nil
}

// FromJSON deserializuje uživatele z JSON a ověří povinná pole.
func FromJSON(data []byte) (User, error) {
	var u User
	if err := json.Unmarshal(data, &u); err != nil {
		return User{}, fmt.Errorf("unmarshal user: %w", err)
	}
	if u.ID <= 0 {
		return User{}, fmt.Errorf("user: id musí být kladné, je %d", u.ID)
	}
	if strings.TrimSpace(u.Name) == "" {
		return User{}, errors.New("user: name nesmí být prázdné")
	}
	return u, nil
}

// DecodeEvent podle pole Kind dekóduje payload do odpovídajícího typu.
func DecodeEvent(data []byte) (any, error) {
	var ev Event
	if err := json.Unmarshal(data, &ev); err != nil {
		return nil, fmt.Errorf("unmarshal event: %w", err)
	}
	if len(ev.Payload) == 0 {
		return nil, fmt.Errorf("event %q: chybí payload", ev.Kind)
	}
	switch ev.Kind {
	case "user.created":
		var p UserCreated
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return nil, fmt.Errorf("event %q: %w", ev.Kind, err)
		}
		return p, nil
	case "user.deleted":
		var p UserDeleted
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return nil, fmt.Errorf("event %q: %w", ev.Kind, err)
		}
		return p, nil
	default:
		return nil, fmt.Errorf("event: neznámý kind %q", ev.Kind)
	}
}

// MarshalJSON zapíše částku jako desetinné číslo.
func (c Cents) MarshalJSON() ([]byte, error) {
	v := int64(c)
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	return []byte(fmt.Sprintf("%s%d.%02d", sign, v/100, v%100)), nil
}

// UnmarshalJSON načte částku z desetinného čísla nebo z řetězce.
func (c *Cents) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "null" {
		return nil
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		unquoted, err := strconv.Unquote(s)
		if err != nil {
			return fmt.Errorf("cents: %w", err)
		}
		s = strings.TrimSpace(unquoted)
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("cents: %q není částka: %w", s, err)
	}
	*c = Cents(math.Round(f * 100))
	return nil
}

// StrictDecode dekóduje data do v a odmítne neznámá pole i data navíc.
func StrictDecode(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("strict decode: %w", err)
	}
	if dec.More() {
		return errors.New("strict decode: data navíc za JSON hodnotou")
	}
	return nil
}
