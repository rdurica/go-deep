// Package exercise obsahuje cvičení lekce 56.
package exercise

import (
	"errors"
	"time"
)

// Chyby vracené parserem ADR.
var (
	ErrInvalidHeader  = errors.New("adr: neplatná hlavička")
	ErrInvalidStatus  = errors.New("adr: neznámý status")
	ErrInvalidDate    = errors.New("adr: neplatné datum")
	ErrMissingSection = errors.New("adr: chybí sekce")
)

// Status je stav rozhodnutí v ADR.
type Status int

// Stavy ADR. StatusUnknown je zero value, tedy "nenastaveno".
const (
	StatusUnknown Status = iota
	StatusProposed
	StatusAccepted
	StatusRejected
	StatusSuperseded
)

// String implementuje fmt.Stringer.
func (s Status) String() string {
	panic("TODO: úkol A")
}

// ParseStatus převede jméno stavu (case-insensitive) na Status.
func ParseStatus(s string) (Status, error) {
	panic("TODO: úkol A")
}

// Fold převede text na malá písmena bez diakritiky.
func Fold(s string) string {
	panic("TODO: úkol A")
}

// Slug převede titulek na URL-bezpečný tvar bez diakritiky.
func Slug(title string) string {
	panic("TODO: úkol A")
}

// ADR je záznam architektonického rozhodnutí.
type ADR struct {
	Number       int
	Title        string
	Status       Status
	Date         time.Time
	Context      string
	Decision     string
	Consequences string
}

// Filename vrací jméno souboru ADR, například "0007-use-stdlib-router.md".
func (a ADR) Filename() string {
	panic("TODO: úkol A")
}

// Render vykreslí ADR jako markdown dokument.
func (a ADR) Render() string {
	panic("TODO: úkol B")
}

// ParseADR rozebere markdown vyrobený metodou Render zpět na ADR.
func ParseADR(s string) (ADR, error) {
	panic("TODO: úkol B")
}

// Index vygeneruje markdown tabulku všech ADR seřazenou podle čísla.
func Index(adrs []ADR) string {
	panic("TODO: úkol B")
}

// Severity je závažnost nálezu v kontrole specifikace.
type Severity int

// Závažnosti nálezů.
const (
	SeverityInfo Severity = iota
	SeverityWarn
	SeverityError
)

// String implementuje fmt.Stringer.
func (s Severity) String() string {
	panic("TODO: úkol C")
}

// Rule je jedno pravidlo kontroly specifikace.
type Rule struct {
	ID       string
	Keywords []string
	Severity Severity
	Message  string
}

// Finding je nález kontroly specifikace.
type Finding struct {
	RuleID   string
	Severity Severity
	Message  string
}

// SpecCheck je konfigurovatelný analyzátor specifikace.
type SpecCheck struct {
	Rules []Rule
}

// DefaultSpecCheck vrací výchozí sadu pravidel pro spec-first zadání.
func DefaultSpecCheck() SpecCheck {
	panic("TODO: úkol C")
}

// Check projde specifikaci a vrátí nálezy v pořadí pravidel.
func (c SpecCheck) Check(spec string) []Finding {
	panic("TODO: úkol C")
}

// CheckSpec zkontroluje specifikaci výchozí sadou pravidel.
func CheckSpec(spec string) []Finding {
	panic("TODO: úkol C")
}
