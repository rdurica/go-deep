// Package exercise obsahuje cvičení lekce 56.
package exercise

import (
	"errors"
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
// String vrací "Unknown", "Proposed", "Accepted", "Rejected" nebo "Superseded".
// Hodnota mimo rozsah vrací "Unknown".
func (s Status) String() string {
	// TODO
	return ""
}

// --- Stupeň: jednoduchý ---
// ParseStatus převede jméno stavu (case-insensitive, s oříznutím mezer) na Status.
// Neznámý vstup vrací chybu obalující ErrInvalidStatus.
func ParseStatus(s string) (Status, error) {
	// TODO
	return *new(Status), nil
}

// Fold převede text na malá písmena bez diakritiky.
// Znaky bez mapování nechává beze změny ("Akceptační Kritéria" → "akceptacni kriteria").
func Fold(s string) string {
	// TODO
	return ""
}

// Slug převede titulek na URL-bezpečný tvar: Fold, pak jen [a-z0-9],
// ostatní skupiny znaků nahraď jedinou pomlčkou, bez pomlčky na začátku/konci.
// "---" → "".
func Slug(title string) string {
	// TODO
	return ""
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
// Filename vrací jméno souboru ADR: "<číslo4>-<slug>.md".
// Číslo je na čtyři místa doplněné nulami, delší se nezkracuje.
// Prázdný slug nahraď "adr".
func (a ADR) Filename() string {
	// TODO
	return ""
}

// Render vykreslí ADR jako markdown: hlavička "# N. Title", Status, Date (2006-01-02),
// sekce Context / Decision / Consequences. Texty sekcí ořízni o okolní bílé znaky.
// Výstup končí \n a bez odsazení.
func (a ADR) Render() string {
	// TODO
	return ""
}

// ParseADR rozebere markdown z Render zpět na ADR (včetně víceřádkových sekcí a \r\n).
// Chyby vždy obaluj kvůli errors.Is:
// neplatná hlavička → ErrInvalidHeader, neznámý status → ErrInvalidStatus,
// špatné datum → ErrInvalidDate, chybějící/prázdná sekce nebo Status/Date → ErrMissingSection.
func ParseADR(s string) (ADR, error) {
	// TODO
	return *new(ADR), nil
}

// Index vygeneruje markdown tabulku Číslo|Titulek|Status|Datum, seřazenou podle čísla
// (při shodě podle titulku). Prázdný vstup → "_Žádné ADR._\n".
// Při duplicitním čísle za tabulkou prázdný řádek a "> pozor: duplicitní číslo N (K×)" vzestupně.
func Index(adrs []ADR) string {
	// TODO
	return ""
}

// Severity je závažnost nálezu v kontrole specifikace.
type Severity int

// Závažnosti nálezů.
const (
	SeverityInfo Severity = iota
	SeverityWarn
	SeverityError
)

// String vrací "INFO", "WARN" nebo "ERROR".
func (s Severity) String() string {
	// TODO
	return ""
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

// DefaultSpecCheck vrací výchozí sadu pravidel v tomto pořadí:
// acceptance (ERROR: akceptační kritéria / acceptance criteria / kritéria přijetí),
// edge-cases (ERROR: hraniční případ / edge case),
// errors (ERROR: chybový stav(y) / error handling),
// go-version (WARN: "go 1."),
// deps (WARN: bez/žádné závislosti / pouze stdlib / stdlib only).
func DefaultSpecCheck() SpecCheck {
	// TODO
	return *new(SpecCheck)
}

// Check pro každé pravidlo, jehož žádné klíčové slovo se ve spec nevyskytuje, vrátí nález.
// Porovnávej přes Fold. Pořadí nálezů = pořadí pravidel. Bez pravidel → žádné nálezy.
func (c SpecCheck) Check(spec string) []Finding {
	// TODO
	return nil
}

// CheckSpec zkontroluje specifikaci výchozí sadou pravidel (DefaultSpecCheck).
func CheckSpec(spec string) []Finding {
	// TODO
	return nil
}
