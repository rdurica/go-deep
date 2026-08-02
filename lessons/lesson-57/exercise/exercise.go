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

// --- Stupeň: jednoduchý ---
// String vrací "INFO", "WARN" nebo "ERROR".
func (s Severity) String() string {
	// TODO
	return ""
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

// ParseFuncs přes go/parser rozebere jeden soubor a vrátí funkce/metody v pořadí výskytu.
// Params/Results = počty jednotlivých parametrů a návratů (přijímač se nepočítá).
// Lines = řádky deklarace včetně hlavičky i uzavírací závorky. Exported = velké písmeno.
// Nevalidní zdroják vrací chybu.
func ParseFuncs(src string) ([]FuncInfo, error) {
	// TODO
	return nil, nil
}

// --- Stupeň: střední ---
// CheckIgnoredErrors najde přiřazení, kde pravá strana je jediné volání funkce
// a levá obsahuje _ (např. _ = f() i v, _ := f()). Ne _ = x ani v, ok := m[k].
// Rule: "ignored-error", SeverityError. Řádek z pozice blank identifikátoru.
// Nálezy v pořadí výskytu; každý má vyplněné Line, Rule, Severity a neprázdné Message.
// Při chybě parsování vrať jediný nález parse-error (SeverityError, Line 0).
func CheckIgnoredErrors(src string) []Finding {
	// TODO
	return nil
}

// CheckContextInStruct najde pole typu context.Context (i *context.Context)
// v jakémkoli structu včetně anonymních. Rule: "context-in-struct", SeverityError.
// Řádek z pozice pole. Nálezy v pořadí výskytu s vyplněnými poli Finding.
// Při chybě parsování vrať jediný nález parse-error (SeverityError, Line 0).
func CheckContextInStruct(src string) []Finding {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---
// CheckContextNotFirst najde funkce/metody s parametrem context.Context mimo první pozici.
// Na jednu funkci nanejvýš jeden nález. Rule: "context-not-first", SeverityWarn.
// Nálezy v pořadí výskytu s vyplněnými poli Finding.
// Při chybě parsování vrať jediný nález parse-error (SeverityError, Line 0).
func CheckContextNotFirst(src string) []Finding {
	// TODO
	return nil
}

// Hunk je jeden blok změn z unified diffu.
type Hunk struct {
	File     string
	OldStart int
	NewStart int
	Header   string
	Lines    []string
}

// CriticalPath vrátí z unified diffu jen hunky se změněným řádkem (+/-), které se týkají
// funkce (func v hlavičce za @@ nebo v těle). File z +++ b/... (bez b/); u /dev/null z --- a/...
// OldStart/NewStart z @@; Header = text za druhým @@; Lines včetně vedoucího znaku.
// Hunk bez změn nebo bez Go funkce vynech. Testovací diffy jsou v testdata/.
func CriticalPath(diff string) []Hunk {
	// TODO
	return nil
}

// ReviewReport sestaví přehled seskupený podle závažnosti ERROR, WARN, INFO.
// Prázdné skupiny vynech; prázdný vstup → "Žádné nálezy.\n".
// Uvnitř skupiny řaď podle Line, pak Rule, pak Message. Mezi skupinami prázdný řádek;
// výstup končí jediným \n. Formát: "## ERROR (N)" a "- ř. L [rule] zpráva".
func ReviewReport(findings []Finding) string {
	// TODO
	return ""
}
