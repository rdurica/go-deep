// Package exercise obsahuje cvičení lekce 56.
package exercise

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
// Pravidlo splní, když spec obsahuje aspoň jedno klíčové slovo (po Fold).
// Prázdný keyword přeskoč. Bez pravidel vrať prázdný slice.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ — review-lab typického AI spec checkeru.
// Hledá klíčová slova case-sensitive a bez Fold, takže „Vlastník" mine „vlastník".
// Oprav podle kontraktu — testy před opravou padají.
func (c SpecCheck) Check(spec string) []Finding {
	var findings []Finding
	for _, r := range c.Rules {
		if ruleMatches(r, spec) {
			continue
		}
		findings = append(findings, Finding{RuleID: r.ID, Severity: r.Severity, Message: r.Message})
	}
	return findings
}

func ruleMatches(r Rule, spec string) bool {
	for _, kw := range r.Keywords {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		if strings.Contains(spec, kw) {
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
	// TODO
	return ""
}

// Slug převede titulek na URL-bezpečný tvar bez diakritiky.
func Slug(title string) string {
	// TODO
	return ""
}

// --- Stupeň: obtížný ---

// DefaultSpecCheck vrací výchozí sadu pravidel pro spec-first zadání.
func DefaultSpecCheck() SpecCheck {
	// TODO
	return SpecCheck{}
}
