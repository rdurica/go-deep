// Package solutions obsahuje referenční řešení lekce 57.
package solutions

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"
)

// Severity je závažnost nálezu.
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

// Finding je jeden nález review nad zdrojovým kódem.
type Finding struct {
	Rule     string
	Severity Severity
	Line     int
	Message  string
}

// FuncInfo popisuje jednu funkci nebo metodu ve zdrojáku.
type FuncInfo struct {
	Name     string
	Params   int
	Results  int
	Lines    int
	Exported bool
}

// parseSrc rozebere zdroják jednoho souboru.
func parseSrc(src string) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "src.go", src, parser.SkipObjectResolution)
	if err != nil {
		return nil, nil, fmt.Errorf("parse: %w", err)
	}
	return fset, file, nil
}

// fieldCount spočítá parametry nebo návratové hodnoty včetně sdílených typů (a, b int = 2).
func fieldCount(list *ast.FieldList) int {
	if list == nil {
		return 0
	}
	n := 0
	for _, f := range list.List {
		if len(f.Names) == 0 {
			n++
			continue
		}
		n += len(f.Names)
	}
	return n
}

// ParseFuncs vytáhne ze zdrojáku přehled všech funkcí a metod v pořadí výskytu.
func ParseFuncs(src string) ([]FuncInfo, error) {
	fset, file, err := parseSrc(src)
	if err != nil {
		return nil, err
	}

	var out []FuncInfo
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		start := fset.Position(fn.Pos()).Line
		end := fset.Position(fn.End()).Line
		out = append(out, FuncInfo{
			Name:     fn.Name.Name,
			Params:   fieldCount(fn.Type.Params),
			Results:  fieldCount(fn.Type.Results),
			Lines:    end - start + 1,
			Exported: fn.Name.IsExported(),
		})
	}
	return out, nil
}

// parseFailure je nález vracený místo paniky, když zdroják nejde rozebrat.
func parseFailure() []Finding {
	return []Finding{{
		Rule:     "parse-error",
		Severity: SeverityError,
		Line:     0,
		Message:  "zdroják se nepodařilo rozebrat",
	}}
}

// isBlank vrací true pro identifikátor _.
func isBlank(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == "_"
}

// unwrapStar odstraní případnou hvězdičku z typu.
func unwrapStar(e ast.Expr) ast.Expr {
	if star, ok := e.(*ast.StarExpr); ok {
		return star.X
	}
	return e
}

// isContextType vrací true pro context.Context (i přes pointer).
func isContextType(e ast.Expr) bool {
	sel, ok := unwrapStar(e).(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "context" && sel.Sel.Name == "Context"
}

// CheckIgnoredErrors najde přiřazení návratové hodnoty volání do _.
func CheckIgnoredErrors(src string) []Finding {
	fset, file, err := parseSrc(src)
	if err != nil {
		return parseFailure()
	}

	var findings []Finding
	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || len(assign.Rhs) != 1 {
			return true
		}
		if _, ok := assign.Rhs[0].(*ast.CallExpr); !ok {
			return true
		}
		for _, lhs := range assign.Lhs {
			if !isBlank(lhs) {
				continue
			}
			findings = append(findings, Finding{
				Rule:     "ignored-error",
				Severity: SeverityError,
				Line:     fset.Position(lhs.Pos()).Line,
				Message:  "návratová hodnota volání je zahozená do _",
			})
		}
		return true
	})
	return findings
}

// CheckContextInStruct najde pole typu context.Context ve struct typech.
func CheckContextInStruct(src string) []Finding {
	fset, file, err := parseSrc(src)
	if err != nil {
		return parseFailure()
	}

	var findings []Finding
	ast.Inspect(file, func(n ast.Node) bool {
		st, ok := n.(*ast.StructType)
		if !ok || st.Fields == nil {
			return true
		}
		for _, f := range st.Fields.List {
			if !isContextType(f.Type) {
				continue
			}
			name := "vložené pole"
			if len(f.Names) > 0 {
				name = f.Names[0].Name
			}
			findings = append(findings, Finding{
				Rule:     "context-in-struct",
				Severity: SeverityError,
				Line:     fset.Position(f.Pos()).Line,
				Message:  fmt.Sprintf("pole %s typu context.Context patří do parametru, ne do structu", name),
			})
		}
		return true
	})
	return findings
}

// CheckContextNotFirst najde funkce, které berou context.Context, ale ne jako první parametr.
func CheckContextNotFirst(src string) []Finding {
	fset, file, err := parseSrc(src)
	if err != nil {
		return parseFailure()
	}

	var findings []Finding
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Type.Params == nil {
			continue
		}

		pos := 0
		for _, f := range fn.Type.Params.List {
			count := len(f.Names)
			if count == 0 {
				count = 1
			}
			if isContextType(f.Type) && pos > 0 {
				findings = append(findings, Finding{
					Rule:     "context-not-first",
					Severity: SeverityWarn,
					Line:     fset.Position(f.Pos()).Line,
					Message: fmt.Sprintf("%s: context.Context je %d. parametr, má být první",
						fn.Name.Name, pos+1),
				})
				break
			}
			pos += count
		}
	}
	return findings
}

// Hunk je jeden blok změn z unified diffu.
type Hunk struct {
	File     string
	OldStart int
	NewStart int
	Header   string
	Lines    []string
}

// parseHunkHeader rozebere "@@ -12,7 +12,9 @@ func Foo()".
func parseHunkHeader(line string) (oldStart, newStart int, header string, ok bool) {
	rest, ok := strings.CutPrefix(line, "@@ ")
	if !ok {
		return 0, 0, "", false
	}
	ranges, tail, ok := strings.Cut(rest, " @@")
	if !ok {
		return 0, 0, "", false
	}
	parts := strings.Fields(ranges)
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "-") || !strings.HasPrefix(parts[1], "+") {
		return 0, 0, "", false
	}
	oldStart, err := parseStart(parts[0][1:])
	if err != nil {
		return 0, 0, "", false
	}
	newStart, err = parseStart(parts[1][1:])
	if err != nil {
		return 0, 0, "", false
	}
	return oldStart, newStart, strings.TrimSpace(tail), true
}

// parseStart vytáhne počáteční řádek z rozsahu "12,7".
func parseStart(s string) (int, error) {
	if i := strings.IndexByte(s, ','); i >= 0 {
		s = s[:i]
	}
	return strconv.Atoi(s)
}

// hunkIsCritical vrací true, pokud hunk mění řádky a týká se funkce.
func hunkIsCritical(h Hunk) bool {
	changed := false
	mentionsFunc := strings.Contains(h.Header, "func ")
	for _, line := range h.Lines {
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			changed = true
		}
		if strings.Contains(line, "func ") {
			mentionsFunc = true
		}
	}
	return changed && mentionsFunc
}

// CriticalPath vrátí z unified diffu hunky, které mění kód uvnitř funkcí.
func CriticalPath(diff string) []Hunk {
	lines := strings.Split(strings.ReplaceAll(diff, "\r\n", "\n"), "\n")

	var (
		out     []Hunk
		current *Hunk
		file    string
	)
	flush := func() {
		if current == nil {
			return
		}
		if hunkIsCritical(*current) {
			out = append(out, *current)
		}
		current = nil
	}

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			flush()
			file = ""
		case strings.HasPrefix(line, "+++ "):
			flush()
			if path := strings.TrimSpace(strings.TrimPrefix(line, "+++ ")); path != "/dev/null" {
				file = strings.TrimPrefix(path, "b/")
			}
		case strings.HasPrefix(line, "--- "):
			flush()
			if file == "" {
				if path := strings.TrimSpace(strings.TrimPrefix(line, "--- ")); path != "/dev/null" {
					file = strings.TrimPrefix(path, "a/")
				}
			}
		case strings.HasPrefix(line, "@@ "):
			flush()
			oldStart, newStart, header, ok := parseHunkHeader(line)
			if !ok {
				continue
			}
			current = &Hunk{File: file, OldStart: oldStart, NewStart: newStart, Header: header}
		default:
			if current == nil {
				continue
			}
			if line == "" || strings.HasPrefix(line, " ") ||
				strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") ||
				strings.HasPrefix(line, "\\") {
				current.Lines = append(current.Lines, line)
				continue
			}
			flush()
		}
	}
	flush()
	return out
}

// ReviewReport sestaví textový přehled nálezů seskupený podle závažnosti.
func ReviewReport(findings []Finding) string {
	if len(findings) == 0 {
		return "Žádné nálezy.\n"
	}

	groups := map[Severity][]Finding{}
	for _, f := range findings {
		groups[f.Severity] = append(groups[f.Severity], f)
	}

	var blocks []string
	for _, sev := range []Severity{SeverityError, SeverityWarn, SeverityInfo} {
		group := groups[sev]
		if len(group) == 0 {
			continue
		}
		sort.SliceStable(group, func(i, j int) bool {
			if group[i].Line != group[j].Line {
				return group[i].Line < group[j].Line
			}
			if group[i].Rule != group[j].Rule {
				return group[i].Rule < group[j].Rule
			}
			return group[i].Message < group[j].Message
		})

		var b strings.Builder
		fmt.Fprintf(&b, "## %s (%d)\n", sev, len(group))
		for _, f := range group {
			fmt.Fprintf(&b, "- ř. %d [%s] %s\n", f.Line, f.Rule, f.Message)
		}
		blocks = append(blocks, b.String())
	}
	return strings.Join(blocks, "\n")
}
