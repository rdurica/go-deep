// Package exercise obsahuje cvičení lekce 57.
package exercise

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
	panic("TODO: úkol A")
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

// ParseFuncs vytáhne ze zdrojáku přehled všech funkcí a metod v pořadí výskytu.
func ParseFuncs(src string) ([]FuncInfo, error) {
	panic("TODO: úkol A")
}

// CheckIgnoredErrors najde přiřazení návratové hodnoty volání do _.
func CheckIgnoredErrors(src string) []Finding {
	panic("TODO: úkol B")
}

// CheckContextInStruct najde pole typu context.Context ve struct typech.
func CheckContextInStruct(src string) []Finding {
	panic("TODO: úkol B")
}

// CheckContextNotFirst najde funkce, které berou context.Context, ale ne jako první parametr.
func CheckContextNotFirst(src string) []Finding {
	panic("TODO: úkol B")
}

// Hunk je jeden blok změn z unified diffu.
type Hunk struct {
	File     string
	OldStart int
	NewStart int
	Header   string
	Lines    []string
}

// CriticalPath vrátí z unified diffu hunky, které mění kód uvnitř funkcí.
func CriticalPath(diff string) []Hunk {
	panic("TODO: úkol C")
}

// ReviewReport sestaví textový přehled nálezů seskupený podle závažnosti.
func ReviewReport(findings []Finding) string {
	panic("TODO: úkol C")
}
