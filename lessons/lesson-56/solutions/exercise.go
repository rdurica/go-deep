// Package solutions obsahuje referenční řešení lekce 56.
package solutions

import (
	"strings"
)

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
	switch s {
	case SeverityWarn:
		return "WARN"
	case SeverityError:
		return "ERROR"
	default:
		return "INFO"
	}
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

// foldMap mapuje písmena s diakritikou na ASCII protějšek.
var foldMap = map[rune]rune{
	'á': 'a', 'ä': 'a', 'à': 'a', 'â': 'a', 'ą': 'a',
	'č': 'c', 'ć': 'c', 'ç': 'c',
	'ď': 'd',
	'é': 'e', 'ě': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e', 'ę': 'e',
	'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
	'ľ': 'l', 'ĺ': 'l', 'ł': 'l',
	'ň': 'n', 'ń': 'n',
	'ó': 'o', 'ö': 'o', 'ô': 'o', 'ò': 'o',
	'ř': 'r', 'ŕ': 'r',
	'š': 's', 'ś': 's',
	'ť': 't',
	'ú': 'u', 'ů': 'u', 'ü': 'u', 'ù': 'u', 'û': 'u',
	'ý': 'y', 'ÿ': 'y',
	'ž': 'z', 'ź': 'z', 'ż': 'z',
}

// --- Stupeň: jednoduchý ---
// Check projde specifikaci a vrátí nálezy v pořadí pravidel.
func (c SpecCheck) Check(spec string) []Finding {
	folded := Fold(spec)
	var findings []Finding
	for _, r := range c.Rules {
		if ruleMatches(r, folded) {
			continue
		}
		findings = append(findings, Finding{RuleID: r.ID, Severity: r.Severity, Message: r.Message})
	}
	return findings
}

func ruleMatches(r Rule, foldedSpec string) bool {
	for _, kw := range r.Keywords {
		kw = strings.TrimSpace(Fold(kw))
		if kw == "" {
			continue
		}
		if strings.Contains(foldedSpec, kw) {
			return true
		}
	}
	return false
}

// CheckSpec zkontroluje specifikaci výchozí sadou pravidel.
func CheckSpec(spec string) []Finding {
	return DefaultSpecCheck().Check(spec)
}

// --- Stupeň: střední ---
// Fold převede text na malá písmena bez diakritiky.
func Fold(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range strings.ToLower(s) {
		if repl, ok := foldMap[r]; ok {
			r = repl
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Slug převede titulek na URL-bezpečný tvar bez diakritiky.
func Slug(title string) string {
	var b strings.Builder
	pendingDash := false
	for _, r := range Fold(title) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			if pendingDash && b.Len() > 0 {
				b.WriteByte('-')
			}
			pendingDash = false
			b.WriteRune(r)
		default:
			pendingDash = true
		}
	}
	return b.String()
}

// --- Stupeň: obtížný ---
// DefaultSpecCheck vrací výchozí sadu pravidel pro spec-first zadání.
func DefaultSpecCheck() SpecCheck {
	return SpecCheck{Rules: []Rule{
		{
			ID:       "acceptance",
			Keywords: []string{"akceptační kritéria", "acceptance criteria", "kritéria přijetí"},
			Severity: SeverityError,
			Message:  "chybí akceptační kritéria",
		},
		{
			ID:       "edge-cases",
			Keywords: []string{"hraniční případ", "edge case"},
			Severity: SeverityError,
			Message:  "chybí hraniční případy",
		},
		{
			ID:       "errors",
			Keywords: []string{"chybový stav", "chybové stavy", "error handling"},
			Severity: SeverityError,
			Message:  "chybí popis chybových stavů",
		},
		{
			ID:       "go-version",
			Keywords: []string{"go 1."},
			Severity: SeverityWarn,
			Message:  "chybí cílová verze Go",
		},
		{
			ID:       "deps",
			Keywords: []string{"bez závislostí", "žádné závislosti", "pouze stdlib", "stdlib only"},
			Severity: SeverityWarn,
			Message:  "chybí pravidlo o závislostech",
		},
	}}
}
