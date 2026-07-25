// Package task obsahuje doménový typ Task a pravidla, která pro něj platí.
// Balíček nezná HTTP ani JSON — to je práce balíčku httpapi.
package task

import (
	"errors"
	"strings"
	"time"
)

// MaxTitleLength je maximální délka názvu úkolu v bajtech.
const MaxTitleLength = 200

// Chyby domény. Volající je rozlišuje přes errors.Is a mapuje na HTTP kódy.
var (
	// ErrEmptyTitle znamená, že název úkolu je prázdný.
	ErrEmptyTitle = errors.New("title must not be empty")
	// ErrTitleTooLong znamená, že název úkolu překročil MaxTitleLength.
	ErrTitleTooLong = errors.New("title is too long")
	// ErrInvalidStatus znamená, že stav není jednou z povolených hodnot.
	ErrInvalidStatus = errors.New("invalid status")
	// ErrNotFound znamená, že úkol s daným ID neexistuje.
	ErrNotFound = errors.New("task not found")
)

// Status je stav úkolu.
type Status string

// Povolené stavy úkolu.
const (
	StatusTodo  Status = "todo"
	StatusDoing Status = "doing"
	StatusDone  Status = "done"
)

// Valid vrací true, pokud je stav jednou z povolených hodnot.
func (s Status) Valid() bool {
	switch s {
	case StatusTodo, StatusDoing, StatusDone:
		return true
	default:
		return false
	}
}

// Task je úkol v seznamu.
type Task struct {
	ID        string
	Title     string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Normalize ořeže bílé znaky v názvu a doplní výchozí stav.
func Normalize(title string, status Status) (string, Status) {
	if status == "" {
		status = StatusTodo
	}
	return strings.TrimSpace(title), status
}

// Validate ověří název a stav úkolu.
func Validate(title string, status Status) error {
	switch {
	case title == "":
		return ErrEmptyTitle
	case len(title) > MaxTitleLength:
		return ErrTitleTooLong
	case !status.Valid():
		return ErrInvalidStatus
	default:
		return nil
	}
}
