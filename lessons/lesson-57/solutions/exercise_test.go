package solutions_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-57/solutions"
)

const srcFuncs = `package demo

import "context"

func Add(a, b int) int {
	return a + b
}

func noop() {}

func (s *Store) Get(ctx context.Context, id string) (Bookmark, error) {
	return Bookmark{}, nil
}

func Split(s string) (head string, tail string, err error) {
	return "", "", nil
}
`

func TestParseFuncs(t *testing.T) {
	got, err := exercise.ParseFuncs(srcFuncs)
	if err != nil {
		t.Fatalf("ParseFuncs() = chyba %v", err)
	}

	want := []exercise.FuncInfo{
		{Name: "Add", Params: 2, Results: 1, Lines: 3, Exported: true},
		{Name: "noop", Params: 0, Results: 0, Lines: 1, Exported: false},
		{Name: "Get", Params: 2, Results: 2, Lines: 3, Exported: true},
		{Name: "Split", Params: 1, Results: 3, Lines: 3, Exported: true},
	}
	if len(got) != len(want) {
		t.Fatalf("ParseFuncs() vrátil %d funkcí (%+v), chci %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("funkce %d = %+v, chci %+v", i, got[i], want[i])
		}
	}
}

func TestParseFuncsInvalidSource(t *testing.T) {
	if _, err := exercise.ParseFuncs("tohle není Go"); err == nil {
		t.Error("ParseFuncs(nevalidní zdroják) = nil chyba, chci chybu")
	}
}

func TestParseFuncsEmptyFile(t *testing.T) {
	got, err := exercise.ParseFuncs("package demo\n")
	if err != nil {
		t.Fatalf("ParseFuncs() = chyba %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ParseFuncs(soubor bez funkcí) = %+v, chci prázdný výsledek", got)
	}
}

const srcIgnored = `package demo

import "os"

func run() error {
	_ = os.Remove("x")
	f, _ := os.Open("y")
	defer f.Close()
	v, ok := lookup("z")
	if !ok {
		return nil
	}
	_ = v
	return nil
}
`

func TestCheckIgnoredErrors(t *testing.T) {
	got := exercise.CheckIgnoredErrors(srcIgnored)
	wantLines := []int{6, 7}

	if len(got) != len(wantLines) {
		t.Fatalf("CheckIgnoredErrors() = %+v, chci %d nálezů na řádcích %v", got, len(wantLines), wantLines)
	}
	for i, line := range wantLines {
		if got[i].Line != line {
			t.Errorf("nález %d je na řádku %d, chci %d", i, got[i].Line, line)
		}
		if got[i].Rule != "ignored-error" {
			t.Errorf("nález %d má Rule %q, chci %q", i, got[i].Rule, "ignored-error")
		}
		if got[i].Severity != exercise.SeverityError {
			t.Errorf("nález %d má závažnost %v, chci ERROR", i, got[i].Severity)
		}
		if got[i].Message == "" {
			t.Errorf("nález %d nemá zprávu", i)
		}
	}
}

func TestCheckIgnoredErrorsClean(t *testing.T) {
	src := `package demo

func run() error {
	v, ok := lookup("z")
	if !ok {
		return nil
	}
	_ = v
	return nil
}
`
	if got := exercise.CheckIgnoredErrors(src); len(got) != 0 {
		t.Errorf("CheckIgnoredErrors(čistý kód) = %+v, chci žádné nálezy", got)
	}
}

const srcContextStruct = `package demo

import "context"

type Worker struct {
	ctx  context.Context
	name string
}

type Job struct {
	Deadline int64
}

func run(ctx context.Context) {
	s := struct {
		ctx context.Context
	}{ctx: ctx}
	_ = s
}
`

func TestCheckContextInStruct(t *testing.T) {
	got := exercise.CheckContextInStruct(srcContextStruct)
	wantLines := []int{6, 16}

	if len(got) != len(wantLines) {
		t.Fatalf("CheckContextInStruct() = %+v, chci nálezy na řádcích %v", got, wantLines)
	}
	for i, line := range wantLines {
		if got[i].Line != line {
			t.Errorf("nález %d je na řádku %d, chci %d", i, got[i].Line, line)
		}
		if got[i].Rule != "context-in-struct" {
			t.Errorf("nález %d má Rule %q, chci %q", i, got[i].Rule, "context-in-struct")
		}
		if got[i].Severity != exercise.SeverityError {
			t.Errorf("nález %d má závažnost %v, chci ERROR", i, got[i].Severity)
		}
	}
}

const srcContextParam = `package demo

import "context"

func Good(ctx context.Context, id string) error { return nil }

func Bad(id string, ctx context.Context) error { return nil }

func (s *Store) AlsoBad(id string, limit int, ctx context.Context) error { return nil }

func None(id string) error { return nil }
`

func TestCheckContextNotFirst(t *testing.T) {
	got := exercise.CheckContextNotFirst(srcContextParam)
	wantLines := []int{7, 9}

	if len(got) != len(wantLines) {
		t.Fatalf("CheckContextNotFirst() = %+v, chci nálezy na řádcích %v", got, wantLines)
	}
	for i, line := range wantLines {
		if got[i].Line != line {
			t.Errorf("nález %d je na řádku %d, chci %d", i, got[i].Line, line)
		}
		if got[i].Rule != "context-not-first" {
			t.Errorf("nález %d má Rule %q, chci %q", i, got[i].Rule, "context-not-first")
		}
		if got[i].Severity != exercise.SeverityWarn {
			t.Errorf("nález %d má závažnost %v, chci WARN", i, got[i].Severity)
		}
	}
}

func TestChecksReportParseError(t *testing.T) {
	broken := "func ("
	checks := map[string]func(string) []exercise.Finding{
		"CheckIgnoredErrors":   exercise.CheckIgnoredErrors,
		"CheckContextInStruct": exercise.CheckContextInStruct,
		"CheckContextNotFirst": exercise.CheckContextNotFirst,
	}
	for name, check := range checks {
		got := check(broken)
		if len(got) != 1 || got[0].Rule != "parse-error" || got[0].Severity != exercise.SeverityError {
			t.Errorf("%s(rozbitý zdroják) = %+v, chci jediný nález parse-error/ERROR", name, got)
		}
	}
}

func readDiff(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("nelze načíst %s: %v", name, err)
	}
	return string(data)
}

func TestCriticalPathHandler(t *testing.T) {
	got := exercise.CriticalPath(readDiff(t, "handler.diff"))
	if len(got) != 1 {
		t.Fatalf("CriticalPath(handler.diff) vrátil %d hunků, chci 1: %+v", len(got), got)
	}

	h := got[0]
	if h.File != "internal/httpapi/server.go" {
		t.Errorf("File = %q, chci %q", h.File, "internal/httpapi/server.go")
	}
	if h.OldStart != 24 || h.NewStart != 25 {
		t.Errorf("(OldStart, NewStart) = (%d, %d), chci (24, 25)", h.OldStart, h.NewStart)
	}
	if !strings.Contains(h.Header, "func (s *Server) createBookmark") {
		t.Errorf("Header = %q, chci hlavičku s funkcí createBookmark", h.Header)
	}
	added := 0
	for _, line := range h.Lines {
		if strings.HasPrefix(line, "+") {
			added++
		}
	}
	if added != 4 {
		t.Errorf("hunk má %d přidaných řádků, chci 4: %q", added, h.Lines)
	}
}

func TestCriticalPathWorker(t *testing.T) {
	got := exercise.CriticalPath(readDiff(t, "worker.diff"))
	if len(got) != 2 {
		t.Fatalf("CriticalPath(worker.diff) vrátil %d hunků, chci 2: %+v", len(got), got)
	}
	wantStarts := [][2]int{{30, 30}, {60, 64}}
	for i, want := range wantStarts {
		if got[i].OldStart != want[0] || got[i].NewStart != want[1] {
			t.Errorf("hunk %d = (%d, %d), chci (%d, %d)", i, got[i].OldStart, got[i].NewStart, want[0], want[1])
		}
		if got[i].File != "internal/worker/pool.go" {
			t.Errorf("hunk %d má File %q, chci %q", i, got[i].File, "internal/worker/pool.go")
		}
	}
}

func TestCriticalPathSkipsNonCode(t *testing.T) {
	if got := exercise.CriticalPath(readDiff(t, "docs.diff")); len(got) != 0 {
		t.Errorf("CriticalPath(docs.diff) = %+v, chci žádné hunky", got)
	}
	if got := exercise.CriticalPath(""); len(got) != 0 {
		t.Errorf("CriticalPath(\"\") = %+v, chci žádné hunky", got)
	}
	noChange := "--- a/x.go\n+++ b/x.go\n@@ -1,3 +1,3 @@ func Foo() {\n \treturn\n"
	if got := exercise.CriticalPath(noChange); len(got) != 0 {
		t.Errorf("CriticalPath(hunk bez změn) = %+v, chci žádné hunky", got)
	}
}

func TestReviewReport(t *testing.T) {
	findings := []exercise.Finding{
		{Rule: "context-not-first", Severity: exercise.SeverityWarn, Line: 12, Message: "ctx není první"},
		{Rule: "ignored-error", Severity: exercise.SeverityError, Line: 30, Message: "chyba do _"},
		{Rule: "context-in-struct", Severity: exercise.SeverityError, Line: 7, Message: "ctx ve structu"},
		{Rule: "naming", Severity: exercise.SeverityInfo, Line: 3, Message: "zkratka v názvu"},
	}

	want := strings.Join([]string{
		"## ERROR (2)",
		"- ř. 7 [context-in-struct] ctx ve structu",
		"- ř. 30 [ignored-error] chyba do _",
		"",
		"## WARN (1)",
		"- ř. 12 [context-not-first] ctx není první",
		"",
		"## INFO (1)",
		"- ř. 3 [naming] zkratka v názvu",
		"",
	}, "\n")

	if got := exercise.ReviewReport(findings); got != want {
		t.Errorf("ReviewReport() =\n%q\nchci\n%q", got, want)
	}
}

func TestReviewReportEmpty(t *testing.T) {
	if got, want := exercise.ReviewReport(nil), "Žádné nálezy.\n"; got != want {
		t.Errorf("ReviewReport(nil) = %q, chci %q", got, want)
	}
}

func TestReviewReportSingleSeverity(t *testing.T) {
	got := exercise.ReviewReport([]exercise.Finding{
		{Rule: "a", Severity: exercise.SeverityWarn, Line: 2, Message: "x"},
	})
	want := "## WARN (1)\n- ř. 2 [a] x\n"
	if got != want {
		t.Errorf("ReviewReport() = %q, chci %q", got, want)
	}
}

func TestSeverityString(t *testing.T) {
	tests := map[exercise.Severity]string{
		exercise.SeverityInfo:  "INFO",
		exercise.SeverityWarn:  "WARN",
		exercise.SeverityError: "ERROR",
	}
	for in, want := range tests {
		if got := in.String(); got != want {
			t.Errorf("Severity(%d).String() = %q, chci %q", int(in), got, want)
		}
	}
}
