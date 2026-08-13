// Package exercise obsahuje cvičení lekce 58.
package exercise

import (
	"errors"
	"strings"
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

// String vrací INFO, WARN nebo ERROR.
func (s Severity) String() string {
	switch s {
	case SeverityWarn:
		return "WARN"
	case SeverityError:
		return "ERROR"
	default:
		return "INFO"
	}
}

// CheckItem je jedna položka review checklistu.
type CheckItem struct {
	ID       string
	Text     string
	Severity Severity
}

// --- Stupeň: jednoduchý ---
// MergeChecklists spojí základní a osobní checklist; zachová pořadí base.
// Stejné ID v personal přepíše položku na původním místě v base; nové ID z personal na konec.
// Duplicita uvnitř base: platí první výskyt; prázdné ID zahodí; prázdný vstup prázdný výsledek.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ — osobní položka nepřepíše základní se stejným ID.
// Najdi chybu a oprav — testy před opravou padají.
func MergeChecklists(base, personal []CheckItem) []CheckItem {
	merged := append([]CheckItem{}, base...)
	for _, item := range personal {
		if item.ID != "" {
			merged = append(merged, item)
		}
	}
	return merged
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
func (r Role) String() string {
	switch r {
	case RoleSpec:
		return "spec"
	case RoleTests:
		return "tests"
	case RoleImpl:
		return "impl"
	case RoleReview:
		return "review"
	case RoleDone:
		return "done"
	default:
		return "none"
	}
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

// Current vrací aktuální roli pairing session. Po Finish vrací RoleDone.
func (s *Session) Current() Role {
	return s.current
}

// --- Stupeň: střední ---
// NewSession vytvoří session; nil now znamená time.Now.
// Časy událostí v timeline ber vždy z těchto hodin.
func NewSession(now func() time.Time) *Session {
	// TODO
	return &Session{now: now}
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
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "errors":
		return "lekce 14 a 21 — chyby jako hodnoty"
	case "context":
		return "lekce 27 — context v request scope"
	case "concurrency":
		return "lekce 40–47 — souběžnost"
	case "http":
		return "lekce 24–26 — net/http a middleware"
	case "design":
		return "lekce 32–35 — hranice balíčků"
	case "testing":
		return "lekce 17 a 52 — testování"
	default:
		return "lekce 57 — strukturované review"
	}
}

// --- Stupeň: obtížný ---
// Finish uzavře session jen z role review; jinak ErrInvalidTransition; podruhé ErrFinished.
// Zapíše událost RoleReview → RoleDone s důvodem "hotovo".
func (s *Session) Finish() error {
	// TODO
	return nil
}

// Timeline vrací kopii historie přechodů. Volající nesmí přepsat interní slice.
func (s *Session) Timeline() []Event {
	// TODO
	return []Event{}
}

// ScoreReview spočítá precision, recall a doporučení pro review dril; párování podle ID.
// Duplicitní ID počítá jednou; prázdné ID ignoruj. Precision = TP/(TP+FP), při 0 FP je 0.
// Recall = TP/počet nastražených, při 0 nastražených je 1. Review = doporučení pro zmeškané kategorie, abecedně.
func ScoreReview(found, planted []Finding) Score {
	// TODO
	return Score{}
}
