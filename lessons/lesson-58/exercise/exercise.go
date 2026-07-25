// Package exercise obsahuje cvičení lekce 58.
package exercise

import (
	"errors"
	"time"
)

// Severity je závažnost položky checklistu.
type Severity int

// Závažnosti položek checklistu.
const (
	SeverityInfo Severity = iota
	SeverityWarn
	SeverityError
)

// String implementuje fmt.Stringer.
func (s Severity) String() string {
	panic("TODO: úkol A")
}

// CheckItem je jedna položka review checklistu.
type CheckItem struct {
	ID       string
	Text     string
	Severity Severity
}

// MergeChecklists spojí základní a osobní checklist. Osobní položka se stejným ID
// přepíše základní, pořadí základních položek zůstává zachované.
func MergeChecklists(base, personal []CheckItem) []CheckItem {
	panic("TODO: úkol A")
}

// Role je role v pairing session s agentem.
type Role int

// Role v pairing protokolu. RoleNone je zero value, tedy "session nezačala".
const (
	RoleNone Role = iota
	RoleSpec
	RoleTests
	RoleImpl
	RoleReview
	RoleDone
)

// String implementuje fmt.Stringer.
func (r Role) String() string {
	panic("TODO: úkol B")
}

// Chyby pairing session.
var (
	ErrNotStarted        = errors.New("session: ještě nezačala")
	ErrAlreadyStarted    = errors.New("session: už začala")
	ErrFinished          = errors.New("session: už skončila")
	ErrInvalidRole       = errors.New("session: neplatná role")
	ErrInvalidTransition = errors.New("session: nepovolený přechod")
	ErrMissingReason     = errors.New("session: chybí důvod předání")
)

// Event je jeden přechod v pairing session.
type Event struct {
	From   Role
	To     Role
	Reason string
	At     time.Time
}

// Session modeluje pairing session: kdo drží spec, testy, implementaci a review.
type Session struct {
	now      func() time.Time
	current  Role
	started  bool
	timeline []Event
}

// NewSession vytvoří session. Hodiny se předávají jako závislost kvůli testovatelnosti;
// nil znamená time.Now.
func NewSession(now func() time.Time) *Session {
	panic("TODO: úkol B")
}

// Current vrací aktuální roli.
func (s *Session) Current() Role {
	panic("TODO: úkol B")
}

// Start zahájí session v dané roli.
func (s *Session) Start(r Role) error {
	panic("TODO: úkol B")
}

// Handoff předá session další roli. Důvod je povinný.
func (s *Session) Handoff(to Role, reason string) error {
	panic("TODO: úkol B")
}

// Finish uzavře session. Jde to jen z role review.
func (s *Session) Finish() error {
	panic("TODO: úkol B")
}

// Timeline vrací kopii historie přechodů.
func (s *Session) Timeline() []Event {
	panic("TODO: úkol B")
}

// Finding je jeden nález review drilu.
type Finding struct {
	ID       string
	Category string
}

// Score je vyhodnocení review drilu.
type Score struct {
	TruePositives  int
	FalsePositives int
	Missed         int
	Precision      float64
	Recall         float64
	Review         []string
}

// RecommendLesson vrací doporučení k zopakování podle kategorie zmeškaného nálezu.
func RecommendLesson(category string) string {
	panic("TODO: úkol C")
}

// ScoreReview porovná nalezené nálezy s nastraženými a spočítá precision a recall.
func ScoreReview(found, planted []Finding) Score {
	panic("TODO: úkol C")
}
