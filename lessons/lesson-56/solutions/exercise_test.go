package solutions_test

import (
	"fmt"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-56/solutions"
)

func TestCheckSpecComplete(t *testing.T) {
	const completeSpec = `# Spec: služba záložek

Cílová verze: Go 1.26, pouze stdlib (bez závislostí).

## Akceptační kritéria
- POST /bookmarks vytvoří záložku.

## Hraniční případy
- prázdný titulek, duplicitní URL

## Chybové stavy
- 400 při neplatné URL
`
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

func TestDefaultSpecCheck(t *testing.T) {
	c := exercise.DefaultSpecCheck()
	if len(c.Rules) != 5 {
		t.Fatalf("DefaultSpecCheck() má %d pravidel, chci 5", len(c.Rules))
	}
	wantIDs := []string{"acceptance", "edge-cases", "errors", "go-version", "deps"}
	for i, want := range wantIDs {
		if c.Rules[i].ID != want {
			t.Errorf("pravidlo %d má ID %q, chci %q", i, c.Rules[i].ID, want)
		}
		if c.Rules[i].Message == "" {
			t.Errorf("pravidlo %q nemá zprávu", want)
		}
	}
	if c.Rules[0].Severity != exercise.SeverityError {
		t.Errorf("acceptance má závažnost %v, chci ERROR", c.Rules[0].Severity)
	}
	if c.Rules[3].Severity != exercise.SeverityWarn {
		t.Errorf("go-version má závažnost %v, chci WARN", c.Rules[3].Severity)
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
