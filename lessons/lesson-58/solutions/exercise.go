// Package solutions obsahuje referenční řešení lekce 58.
package solutions

import (
	"errors"
	"sort"
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

// --- Stupeň: jednoduchý ---
// String implementuje fmt.Stringer.
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

// MergeChecklists spojí základní a osobní checklist. Osobní položka se stejným ID
// přepíše základní, pořadí základních položek zůstává zachované.
func MergeChecklists(base, personal []CheckItem) []CheckItem {
	merged := make([]CheckItem, 0, len(base)+len(personal))
	index := make(map[string]int, len(base)+len(personal))

	add := func(item CheckItem) {
		if item.ID == "" {
			return
		}
		if i, ok := index[item.ID]; ok {
			merged[i] = item
			return
		}
		index[item.ID] = len(merged)
		merged = append(merged, item)
	}

	for _, item := range base {
		if _, ok := index[item.ID]; ok {
			continue // duplicita v základu, první výskyt vyhrává
		}
		add(item)
	}
	for _, item := range personal {
		add(item)
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

// String implementuje fmt.Stringer.
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

// --- Stupeň: střední ---
// NewSession vytvoří session. Hodiny se předávají jako závislost kvůli testovatelnosti;
// nil znamená time.Now.
func NewSession(now func() time.Time) *Session {
	if now == nil {
		now = time.Now
	}
	return &Session{now: now}
}

// Current vrací aktuální roli.
func (s *Session) Current() Role {
	return s.current
}

// Start zahájí session v dané roli.
func (s *Session) Start(r Role) error {
	if s.started {
		return ErrAlreadyStarted
	}
	if r < RoleSpec || r > RoleReview {
		return ErrInvalidRole
	}
	s.started = true
	s.current = r
	s.timeline = append(s.timeline, Event{From: RoleNone, To: r, Reason: "start", At: s.now()})
	return nil
}

// allowedHandoff vrací true pro povolený přechod: o krok vpřed nebo o krok zpět.
func allowedHandoff(from, to Role) bool {
	if to < RoleSpec || to > RoleReview {
		return false
	}
	diff := int(to) - int(from)
	return diff == 1 || diff == -1
}

// Handoff předá session další roli. Důvod je povinný.
func (s *Session) Handoff(to Role, reason string) error {
	if !s.started {
		return ErrNotStarted
	}
	if s.current == RoleDone {
		return ErrFinished
	}
	if strings.TrimSpace(reason) == "" {
		return ErrMissingReason
	}
	if !allowedHandoff(s.current, to) {
		return ErrInvalidTransition
	}
	s.timeline = append(s.timeline, Event{From: s.current, To: to, Reason: reason, At: s.now()})
	s.current = to
	return nil
}

// --- Stupeň: obtížný ---
// Finish uzavře session. Jde to jen z role review.
func (s *Session) Finish() error {
	if !s.started {
		return ErrNotStarted
	}
	if s.current == RoleDone {
		return ErrFinished
	}
	if s.current != RoleReview {
		return ErrInvalidTransition
	}
	s.timeline = append(s.timeline, Event{From: s.current, To: RoleDone, Reason: "hotovo", At: s.now()})
	s.current = RoleDone
	return nil
}

// Timeline vrací kopii historie přechodů.
func (s *Session) Timeline() []Event {
	out := make([]Event, len(s.timeline))
	copy(out, s.timeline)
	return out
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

// dedupeIDs vrátí množinu ID nálezů.
func dedupeIDs(findings []Finding) map[string]Finding {
	out := make(map[string]Finding, len(findings))
	for _, f := range findings {
		if f.ID == "" {
			continue
		}
		if _, ok := out[f.ID]; !ok {
			out[f.ID] = f
		}
	}
	return out
}

// ScoreReview porovná nalezené nálezy s nastraženými a spočítá precision a recall.
func ScoreReview(found, planted []Finding) Score {
	foundSet := dedupeIDs(found)
	plantedSet := dedupeIDs(planted)

	var score Score
	for id := range foundSet {
		if _, ok := plantedSet[id]; ok {
			score.TruePositives++
		} else {
			score.FalsePositives++
		}
	}

	missedCategories := make(map[string]struct{})
	for id, f := range plantedSet {
		if _, ok := foundSet[id]; ok {
			continue
		}
		score.Missed++
		missedCategories[f.Category] = struct{}{}
	}

	if n := score.TruePositives + score.FalsePositives; n > 0 {
		score.Precision = float64(score.TruePositives) / float64(n)
	}
	if len(plantedSet) == 0 {
		score.Recall = 1
	} else {
		score.Recall = float64(score.TruePositives) / float64(len(plantedSet))
	}

	seen := make(map[string]struct{}, len(missedCategories))
	for cat := range missedCategories {
		rec := RecommendLesson(cat)
		if _, ok := seen[rec]; ok {
			continue
		}
		seen[rec] = struct{}{}
		score.Review = append(score.Review, rec)
	}
	sort.Strings(score.Review)
	return score
}
