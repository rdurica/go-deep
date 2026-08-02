// Package solutions obsahuje referenční řešení lekce 56.
package solutions

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
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

// --- Stupeň: obtížný ---
// String implementuje fmt.Stringer.
func (s Status) String() string {
	switch s {
	case StatusProposed:
		return "Proposed"
	case StatusAccepted:
		return "Accepted"
	case StatusRejected:
		return "Rejected"
	case StatusSuperseded:
		return "Superseded"
	default:
		return "Unknown"
	}
}

// --- Stupeň: jednoduchý ---
// ParseStatus převede jméno stavu (case-insensitive) na Status.
func ParseStatus(s string) (Status, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "proposed":
		return StatusProposed, nil
	case "accepted":
		return StatusAccepted, nil
	case "rejected":
		return StatusRejected, nil
	case "superseded":
		return StatusSuperseded, nil
	default:
		return StatusUnknown, fmt.Errorf("%w: %q", ErrInvalidStatus, strings.TrimSpace(s))
	}
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

// --- Stupeň: střední ---
// Filename vrací jméno souboru ADR, například "0007-use-stdlib-router.md".
func (a ADR) Filename() string {
	slug := Slug(a.Title)
	if slug == "" {
		slug = "adr"
	}
	return fmt.Sprintf("%04d-%s.md", a.Number, slug)
}

// Render vykreslí ADR jako markdown dokument.
func (a ADR) Render() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %d. %s\n\n", a.Number, strings.TrimSpace(a.Title))
	fmt.Fprintf(&b, "- Status: %s\n", a.Status)
	fmt.Fprintf(&b, "- Date: %s\n\n", a.Date.Format("2006-01-02"))
	fmt.Fprintf(&b, "## Context\n\n%s\n\n", strings.TrimSpace(a.Context))
	fmt.Fprintf(&b, "## Decision\n\n%s\n\n", strings.TrimSpace(a.Decision))
	fmt.Fprintf(&b, "## Consequences\n\n%s\n", strings.TrimSpace(a.Consequences))
	return b.String()
}

// parseHeader rozebere řádek "# 7. Titulek".
func parseHeader(line string) (int, string, error) {
	rest, ok := strings.CutPrefix(strings.TrimSpace(line), "# ")
	if !ok {
		return 0, "", fmt.Errorf("%w: %q", ErrInvalidHeader, line)
	}
	numText, title, ok := strings.Cut(rest, ". ")
	if !ok {
		return 0, "", fmt.Errorf("%w: chybí číslo v %q", ErrInvalidHeader, line)
	}
	num, err := strconv.Atoi(strings.TrimSpace(numText))
	if err != nil || num < 0 {
		return 0, "", fmt.Errorf("%w: %q není číslo", ErrInvalidHeader, numText)
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return 0, "", fmt.Errorf("%w: prázdný titulek", ErrInvalidHeader)
	}
	return num, title, nil
}

// ParseADR rozebere markdown vyrobený metodou Render zpět na ADR.
func ParseADR(s string) (ADR, error) {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i == len(lines) {
		return ADR{}, fmt.Errorf("%w: prázdný vstup", ErrInvalidHeader)
	}

	num, title, err := parseHeader(lines[i])
	if err != nil {
		return ADR{}, err
	}
	adr := ADR{Number: num, Title: title}

	sections := make(map[string][]string)
	order := make([]string, 0, 3)
	current := ""
	statusSet, dateSet := false, false

	for i++; i < len(lines); i++ {
		line := lines[i]
		switch {
		case strings.HasPrefix(line, "## "):
			current = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			if _, seen := sections[current]; !seen {
				sections[current] = nil
				order = append(order, current)
			}
		case current == "" && strings.HasPrefix(line, "- Status:"):
			st, err := ParseStatus(strings.TrimPrefix(line, "- Status:"))
			if err != nil {
				return ADR{}, err
			}
			adr.Status, statusSet = st, true
		case current == "" && strings.HasPrefix(line, "- Date:"):
			raw := strings.TrimSpace(strings.TrimPrefix(line, "- Date:"))
			d, err := time.Parse("2006-01-02", raw)
			if err != nil {
				return ADR{}, fmt.Errorf("%w: %q", ErrInvalidDate, raw)
			}
			adr.Date, dateSet = d, true
		default:
			if current != "" {
				sections[current] = append(sections[current], line)
			}
		}
	}

	if !statusSet {
		return ADR{}, fmt.Errorf("%w: Status", ErrMissingSection)
	}
	if !dateSet {
		return ADR{}, fmt.Errorf("%w: Date", ErrMissingSection)
	}

	targets := []struct {
		name string
		dst  *string
	}{
		{"Context", &adr.Context},
		{"Decision", &adr.Decision},
		{"Consequences", &adr.Consequences},
	}
	for _, t := range targets {
		body := strings.TrimSpace(strings.Join(sections[t.name], "\n"))
		if body == "" {
			return ADR{}, fmt.Errorf("%w: %s", ErrMissingSection, t.name)
		}
		*t.dst = body
	}
	return adr, nil
}

// Index vygeneruje markdown tabulku všech ADR seřazenou podle čísla.
func Index(adrs []ADR) string {
	if len(adrs) == 0 {
		return "_Žádné ADR._\n"
	}
	sorted := make([]ADR, len(adrs))
	copy(sorted, adrs)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Number != sorted[j].Number {
			return sorted[i].Number < sorted[j].Number
		}
		return sorted[i].Title < sorted[j].Title
	})

	var b strings.Builder
	b.WriteString("| Číslo | Titulek | Status | Datum |\n")
	b.WriteString("|-------|---------|--------|-------|\n")
	counts := make(map[int]int, len(sorted))
	for _, a := range sorted {
		counts[a.Number]++
		fmt.Fprintf(&b, "| %04d | %s | %s | %s |\n",
			a.Number, strings.TrimSpace(a.Title), a.Status, a.Date.Format("2006-01-02"))
	}

	dupes := make([]int, 0)
	for num, n := range counts {
		if n > 1 {
			dupes = append(dupes, num)
		}
	}
	if len(dupes) > 0 {
		sort.Ints(dupes)
		b.WriteString("\n")
		for _, num := range dupes {
			fmt.Fprintf(&b, "> pozor: duplicitní číslo %d (%d×)\n", num, counts[num])
		}
	}
	return b.String()
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

// ruleMatches vrací true, pokud specifikace obsahuje aspoň jedno klíčové slovo pravidla.
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
