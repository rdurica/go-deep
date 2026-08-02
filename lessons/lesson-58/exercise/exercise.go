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

// --- Stupeň: jednoduchý ---
// String vrací INFO, WARN nebo ERROR.
// Hodnota mimo rozsah konstant vrací prázdný řetězec.
func (s Severity) String() string {
	// TODO
	return ""
}

// CheckItem je jedna položka review checklistu.
type CheckItem struct {
	ID       string
	Text     string
	Severity Severity
}

// MergeChecklists spojí základní a osobní checklist; zachová pořadí base.
// Stejné ID v personal přepíše položku na původním místě v base; nové ID z personal na konec.
// Duplicita uvnitř base: platí první výskyt; prázdné ID zahodí; prázdný vstup prázdný výsledek.
func MergeChecklists(base, personal []CheckItem) []CheckItem {
	// TODO
	return nil
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

// String vrací none, spec, tests, impl, review nebo done; mimo rozsah "none".
// RoleNone je zero value, tedy session ještě nezačala.
func (r Role) String() string {
	// TODO
	return ""
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

// --- Stupeň: střední ---
// NewSession vytvoří session; nil now znamená time.Now.
// Časy událostí v timeline ber vždy z těchto hodin.
func NewSession(now func() time.Time) *Session {
	// TODO
	return nil
}

// Current vrací aktuální roli pairing session.
// Po Finish vrací RoleDone.
func (s *Session) Current() Role {
	// TODO
	return RoleNone
}

// Start zahájí session v roli spec, tests, impl nebo review a zapíše událost RoleNone→r s důvodem "start".
// RoleNone, RoleDone a hodnoty mimo rozsah → ErrInvalidRole. Druhé volání ErrAlreadyStarted.
func (s *Session) Start(r Role) error {
	// TODO
	return nil
}

// Handoff předá session o jeden krok vpřed nebo zpět v řadě spec→tests→impl→review.
// Prázdný důvod → ErrMissingReason; před Start → ErrNotStarted; po Finish → ErrFinished.
// Jiný přechod (včetně RoleDone/RoleNone/sebe sama) → ErrInvalidTransition bez změny stavu.
func (s *Session) Handoff(to Role, reason string) error {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---
// Finish uzavře session jen z role review; jinak ErrInvalidTransition; podruhé ErrFinished.
// Zapíše událost RoleReview → RoleDone s důvodem "hotovo".
func (s *Session) Finish() error {
	// TODO
	return nil
}

// Timeline vrací kopii historie přechodů.
// Volající nesmí přepsat interní slice.
func (s *Session) Timeline() []Event {
	// TODO
	return nil
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

// RecommendLesson vrací doporučení ke zopakování podle kategorie (case-insensitive).
// Pokryté kategorie: errors, context, concurrency, http, design, testing; neznámá výchozí text.
func RecommendLesson(category string) string {
	// TODO
	return ""
}

// ScoreReview spočítá precision, recall a doporučení pro review dril; párování podle ID.
// Duplicitní ID počítá jednou; prázdné ID ignoruj. Precision = TP/(TP+FP), při 0 FP je 0.
// Recall = TP/počet nastražených, při 0 nastražených je 1. Review = doporučení pro zmeškané kategorie, abecedně.
func ScoreReview(found, planted []Finding) Score {
	// TODO
	return Score{}
}
