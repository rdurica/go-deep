package exercise_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-56/exercise"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		in   exercise.Status
		want string
	}{
		{exercise.StatusUnknown, "Unknown"},
		{exercise.StatusProposed, "Proposed"},
		{exercise.StatusAccepted, "Accepted"},
		{exercise.StatusRejected, "Rejected"},
		{exercise.StatusSuperseded, "Superseded"},
		{exercise.Status(42), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.in.String(); got != tt.want {
			t.Errorf("Status(%d).String() = %q, chci %q", int(tt.in), got, tt.want)
		}
	}
}

func TestParseStatus(t *testing.T) {
	ok := map[string]exercise.Status{
		"proposed":   exercise.StatusProposed,
		"ACCEPTED":   exercise.StatusAccepted,
		" Rejected ": exercise.StatusRejected,
		"Superseded": exercise.StatusSuperseded,
	}
	for in, want := range ok {
		got, err := exercise.ParseStatus(in)
		if err != nil || got != want {
			t.Errorf("ParseStatus(%q) = (%v, %v), chci (%v, nil)", in, got, err, want)
		}
	}
	for _, in := range []string{"", "draft", "accpeted"} {
		got, err := exercise.ParseStatus(in)
		if !errors.Is(err, exercise.ErrInvalidStatus) {
			t.Errorf("ParseStatus(%q) = (%v, %v), chci ErrInvalidStatus", in, got, err)
		}
	}
}

func TestSlug(t *testing.T) {
	tests := map[string]string{
		"Použij stdlib router":   "pouzij-stdlib-router",
		"Use stdlib router":      "use-stdlib-router",
		"  Ěščřžýáíé  ":          "escrzyaie",
		"HTTP/2 & gRPC":          "http-2-grpc",
		"---":                    "",
		"":                       "",
		"Ukládej ADR do repa!":   "ukladej-adr-do-repa",
		"Verze Go 1.26 je nutná": "verze-go-1-26-je-nutna",
	}
	for in, want := range tests {
		if got := exercise.Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, chci %q", in, got, want)
		}
	}
}

func TestFold(t *testing.T) {
	tests := map[string]string{
		"Akceptační Kritéria": "akceptacni kriteria",
		"HRANIČNÍ PŘÍPAD":     "hranicni pripad",
		"already ascii":       "already ascii",
	}
	for in, want := range tests {
		if got := exercise.Fold(in); got != want {
			t.Errorf("Fold(%q) = %q, chci %q", in, got, want)
		}
	}
}

func TestFilename(t *testing.T) {
	tests := []struct {
		adr  exercise.ADR
		want string
	}{
		{exercise.ADR{Number: 7, Title: "Use stdlib router"}, "0007-use-stdlib-router.md"},
		{exercise.ADR{Number: 7, Title: "Použij stdlib router"}, "0007-pouzij-stdlib-router.md"},
		{exercise.ADR{Number: 123, Title: "Errors wrap with %w"}, "0123-errors-wrap-with-w.md"},
		{exercise.ADR{Number: 12345, Title: "Big"}, "12345-big.md"},
	}
	for _, tt := range tests {
		if got := tt.adr.Filename(); got != tt.want {
			t.Errorf("ADR{%d, %q}.Filename() = %q, chci %q", tt.adr.Number, tt.adr.Title, got, tt.want)
		}
	}
}

func sampleADR() exercise.ADR {
	return exercise.ADR{
		Number:       7,
		Title:        "Použij stdlib router",
		Status:       exercise.StatusAccepted,
		Date:         date(2024, time.May, 1),
		Context:      "Potřebujeme routing s metodami a wildcardy.",
		Decision:     "Použijeme net/http ServeMux se vzory (metody, wildcardy).",
		Consequences: "Méně závislostí.\nMusíme si napsat vlastní middleware chain.",
	}
}

func TestRender(t *testing.T) {
	want := strings.Join([]string{
		"# 7. Použij stdlib router",
		"",
		"- Status: Accepted",
		"- Date: 2024-05-01",
		"",
		"## Context",
		"",
		"Potřebujeme routing s metodami a wildcardy.",
		"",
		"## Decision",
		"",
		"Použijeme net/http ServeMux se vzory (metody, wildcardy).",
		"",
		"## Consequences",
		"",
		"Méně závislostí.",
		"Musíme si napsat vlastní middleware chain.",
		"",
	}, "\n")

	if got := sampleADR().Render(); got != want {
		t.Errorf("Render() =\n%q\nchci\n%q", got, want)
	}
}

func TestParseADRRoundTrip(t *testing.T) {
	inputs := []exercise.ADR{
		sampleADR(),
		{
			Number:       1,
			Title:        "Errors wrap with %w",
			Status:       exercise.StatusProposed,
			Date:         date(2023, time.January, 31),
			Context:      "Chyby se ztrácejí v logu.",
			Decision:     "Každá vrstva obaluje chybu přes %w.",
			Consequences: "errors.Is a errors.As začnou fungovat napříč vrstvami.",
		},
		{
			Number:       42,
			Title:        "Nahraď in-memory store PostgreSQL",
			Status:       exercise.StatusSuperseded,
			Date:         date(2025, time.December, 24),
			Context:      "Data se ztrácejí při restartu.",
			Decision:     "Přidáme adaptér nad database/sql.",
			Consequences: "Přibude migrace a integrační testy.",
		},
	}

	for _, want := range inputs {
		got, err := exercise.ParseADR(want.Render())
		if err != nil {
			t.Fatalf("ParseADR(Render(%d)) = chyba %v", want.Number, err)
		}
		if got.Number != want.Number || got.Title != want.Title || got.Status != want.Status {
			t.Errorf("round-trip hlavičky = %+v, chci %+v", got, want)
		}
		if !got.Date.Equal(want.Date) {
			t.Errorf("round-trip Date = %v, chci %v", got.Date, want.Date)
		}
		if got.Context != want.Context || got.Decision != want.Decision || got.Consequences != want.Consequences {
			t.Errorf("round-trip sekcí:\n got %+v\nchci %+v", got, want)
		}
	}
}

func TestParseADRErrors(t *testing.T) {
	full := sampleADR().Render()

	tests := []struct {
		name string
		in   string
		want error
	}{
		{"prázdný vstup", "   \n\n", exercise.ErrInvalidHeader},
		{"chybí mřížka", strings.TrimPrefix(full, "# "), exercise.ErrInvalidHeader},
		{"chybí číslo", "# Použij stdlib router\n\n- Status: Accepted\n", exercise.ErrInvalidHeader},
		{"nečíselné číslo", "# X. Titulek\n\n- Status: Accepted\n", exercise.ErrInvalidHeader},
		{"neznámý status", strings.Replace(full, "Accepted", "Done", 1), exercise.ErrInvalidStatus},
		{"neplatné datum", strings.Replace(full, "2024-05-01", "1. 5. 2024", 1), exercise.ErrInvalidDate},
		{"chybí status", strings.Replace(full, "- Status: Accepted\n", "", 1), exercise.ErrMissingSection},
		{"chybí datum", strings.Replace(full, "- Date: 2024-05-01\n", "", 1), exercise.ErrMissingSection},
		{"chybí Decision", strings.Replace(full, "Použijeme net/http ServeMux se vzory (metody, wildcardy).", "", 1), exercise.ErrMissingSection},
		{"chybí sekce Consequences", strings.Split(full, "## Consequences")[0], exercise.ErrMissingSection},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := exercise.ParseADR(tt.in)
			if !errors.Is(err, tt.want) {
				t.Errorf("ParseADR(...) = %v, chci %v", err, tt.want)
			}
		})
	}
}

func TestIndexSorted(t *testing.T) {
	adrs := []exercise.ADR{
		{Number: 12, Title: "Zavedeme slog", Status: exercise.StatusAccepted, Date: date(2024, time.June, 2)},
		{Number: 3, Title: "Bez ORM", Status: exercise.StatusRejected, Date: date(2024, time.February, 9)},
		{Number: 7, Title: "Použij stdlib router", Status: exercise.StatusAccepted, Date: date(2024, time.May, 1)},
	}
	got := exercise.Index(adrs)

	wantLines := []string{
		"| Číslo | Titulek | Status | Datum |",
		"|-------|---------|--------|-------|",
		"| 0003 | Bez ORM | Rejected | 2024-02-09 |",
		"| 0007 | Použij stdlib router | Accepted | 2024-05-01 |",
		"| 0012 | Zavedeme slog | Accepted | 2024-06-02 |",
		"",
	}
	if want := strings.Join(wantLines, "\n"); got != want {
		t.Errorf("Index() =\n%s\nchci\n%s", got, want)
	}
	if strings.Contains(got, "duplicitní") {
		t.Error("Index() hlásí duplicitu tam, kde žádná není")
	}
}

func TestIndexDuplicates(t *testing.T) {
	adrs := []exercise.ADR{
		{Number: 7, Title: "B", Status: exercise.StatusAccepted, Date: date(2024, time.May, 1)},
		{Number: 7, Title: "A", Status: exercise.StatusProposed, Date: date(2024, time.May, 2)},
		{Number: 9, Title: "C", Status: exercise.StatusAccepted, Date: date(2024, time.May, 3)},
		{Number: 9, Title: "D", Status: exercise.StatusAccepted, Date: date(2024, time.May, 4)},
	}
	got := exercise.Index(adrs)

	if !strings.Contains(got, "> pozor: duplicitní číslo 7 (2×)") {
		t.Errorf("Index() nehlásí duplicitu čísla 7:\n%s", got)
	}
	if !strings.Contains(got, "> pozor: duplicitní číslo 9 (2×)") {
		t.Errorf("Index() nehlásí duplicitu čísla 9:\n%s", got)
	}
	i7 := strings.Index(got, "duplicitní číslo 7")
	i9 := strings.Index(got, "duplicitní číslo 9")
	if i7 > i9 {
		t.Error("upozornění na duplicity nejsou seřazená vzestupně")
	}
	// stabilní řazení uvnitř stejného čísla: podle titulku
	iA := strings.Index(got, "| 0007 | A |")
	iB := strings.Index(got, "| 0007 | B |")
	if iA < 0 || iB < 0 || iA > iB {
		t.Errorf("řádky se stejným číslem nejsou seřazené podle titulku:\n%s", got)
	}
}

func TestIndexEmpty(t *testing.T) {
	if got, want := exercise.Index(nil), "_Žádné ADR._\n"; got != want {
		t.Errorf("Index(nil) = %q, chci %q", got, want)
	}
}

const completeSpec = `# Spec: služba záložek

Cílová verze: Go 1.26, pouze stdlib (bez závislostí).

## Akceptační kritéria
- POST /bookmarks vytvoří záložku.

## Hraniční případy
- prázdný titulek, duplicitní URL

## Chybové stavy
- 400 při neplatné URL
`

func TestCheckSpecComplete(t *testing.T) {
	if got := exercise.CheckSpec(completeSpec); len(got) != 0 {
		t.Errorf("CheckSpec(kompletní spec) = %+v, chci žádné nálezy", got)
	}
}

func TestCheckSpecHoley(t *testing.T) {
	spec := "Napiš mi službu na záložky. Ať to funguje."
	got := exercise.CheckSpec(spec)
	if len(got) != 5 {
		t.Fatalf("CheckSpec(děravý spec) vrátil %d nálezů, chci 5: %+v", len(got), got)
	}

	wantIDs := []string{"acceptance", "edge-cases", "errors", "go-version", "deps"}
	for i, want := range wantIDs {
		if got[i].RuleID != want {
			t.Errorf("nález %d má RuleID %q, chci %q (pořadí musí odpovídat pravidlům)", i, got[i].RuleID, want)
		}
		if got[i].Message == "" {
			t.Errorf("nález %d nemá zprávu", i)
		}
	}
	for _, f := range got[:3] {
		if f.Severity != exercise.SeverityError {
			t.Errorf("nález %q má závažnost %v, chci ERROR", f.RuleID, f.Severity)
		}
	}
	for _, f := range got[3:] {
		if f.Severity != exercise.SeverityWarn {
			t.Errorf("nález %q má závažnost %v, chci WARN", f.RuleID, f.Severity)
		}
	}
}

func TestCheckSpecPartial(t *testing.T) {
	spec := "Akceptační kritéria: služba vrací 201.\nEdge case: prázdný vstup.\nGo 1.26, pouze stdlib."
	got := exercise.CheckSpec(spec)
	if len(got) != 1 || got[0].RuleID != "errors" {
		t.Fatalf("CheckSpec(částečný spec) = %+v, chci jediný nález 'errors'", got)
	}
}

func TestCheckSpecCustomRules(t *testing.T) {
	c := exercise.SpecCheck{Rules: []exercise.Rule{
		{ID: "sla", Keywords: []string{"p99"}, Severity: exercise.SeverityWarn, Message: "chybí SLA"},
		{ID: "owner", Keywords: []string{"vlastník", "owner"}, Severity: exercise.SeverityInfo, Message: "chybí vlastník"},
	}}

	if got := c.Check("Vlastník: platební tým. p99 pod 200 ms."); len(got) != 0 {
		t.Errorf("Check() = %+v, chci žádné nálezy", got)
	}
	got := c.Check("Nic užitečného.")
	if len(got) != 2 {
		t.Fatalf("Check() = %+v, chci 2 nálezy", got)
	}
	if got[0].RuleID != "sla" || got[0].Severity != exercise.SeverityWarn {
		t.Errorf("první nález = %+v, chci sla/WARN", got[0])
	}
	if got[1].RuleID != "owner" || got[1].Severity != exercise.SeverityInfo {
		t.Errorf("druhý nález = %+v, chci owner/INFO", got[1])
	}

	empty := exercise.SpecCheck{}
	if got := empty.Check("cokoli"); len(got) != 0 {
		t.Errorf("Check() bez pravidel = %+v, chci žádné nálezy", got)
	}
}

func TestSeverityString(t *testing.T) {
	tests := map[exercise.Severity]string{
		exercise.SeverityInfo:  "INFO",
		exercise.SeverityWarn:  "WARN",
		exercise.SeverityError: "ERROR",
	}
	for in, want := range tests {
		if got := fmt.Sprintf("%v", in); got != want {
			t.Errorf("Severity(%d) = %q, chci %q", int(in), got, want)
		}
	}
}
