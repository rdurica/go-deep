// Package exercise obsahuje cvičení lekce 19.
package exercise

import "errors"

// Chyby vracené při parsování uživatelských ID.
var (
	ErrEmptyID       = errors.New("empty user id")
	ErrInvalidID     = errors.New("invalid user id")
	ErrNonPositiveID = errors.New("user id must be positive")
)

// Chyby vracené funkcí ProcessOrders.
var (
	ErrMissingOrderID  = errors.New("missing order id")
	ErrUnknownStatus   = errors.New("unknown order status")
	ErrInvalidQuantity = errors.New("invalid item quantity")
	ErrInvalidPrice    = errors.New("invalid item price")
)

// Item je jedna položka objednávky.
type Item struct {
	SKU       string
	Quantity  int
	UnitCents int
}

// Order je objednávka ve stavu "paid", "pending" nebo "cancelled".
type Order struct {
	ID       string
	Customer string
	Status   string
	Items    []Item
}

// Summary je agregovaný přehled zpracovaných objednávek.
type Summary struct {
	OrderCount int
	ItemCount  int
	TotalCents int
	Customers  []string
}

// InvoiceLine je jeden řádek faktury.
type InvoiceLine struct {
	Description string
	Quantity    int
	UnitCents   int
}

// Invoice je faktura připravená k vykreslení.
type Invoice struct {
	Number   string
	Customer string
	Lines    []InvoiceLine
}

// --- Stupeň: jednoduchý ---
// ParseUserID převede textové ID na kladné celé číslo.
// Ořízne bílé znaky; prázdný vstup → ErrEmptyID, nečíslo → ErrInvalidID, ≤0 → ErrNonPositiveID.
// Maximální hloubka zanoření 1, žádné else.
func ParseUserID(raw string) (int, error) {
	// TODO
	return 0, nil
}

// --- Stupeň: střední ---
// ParseUserIDs převede čárkami oddělený seznam ID na slice čísel.
// Prázdný nebo jen bílý vstup → prázdný výsledek a nil chybu.
// Chybu obal indexem (0-based): fmt.Errorf("id at index %d: %w", i, err).
func ParseUserIDs(raw string) ([]int, error) {
	// TODO
	return nil, nil
}

// --- Stupeň: obtížný ---
// ProcessOrders agreguje objednávky do souhrnu.
// Prázdné ID → chyba s indexem; cancelled se přeskočí včetně validace položek.
// paid/pending: Quantity <= 0 → ErrInvalidQuantity (SKU v textu), UnitCents < 0 → ErrInvalidPrice (ID v textu).
// Jiný status → ErrUnknownStatus (ID v textu). Při chybě nulový Summary; Customers vzestupně.
// nil i prázdný vstup → nulový Summary a nil chybu.
func ProcessOrders(orders []Order) (Summary, error) {
	// TODO
	return *new(Summary), nil
}

// RenderInvoice vykreslí fakturu do textové podoby.
// Složí renderHeader, renderLines a renderTotal; strings.Builder, ne += v cyklu.
// Oddělovač 32 pomlček; částky z centů na dvě desetinná místa; každý řádek končí \n.
func RenderInvoice(inv Invoice) string {
	// TODO
	return ""
}

// renderHeader vykreslí hlavičku faktury včetně oddělovače.
// Formát: "INVOICE <Number>\nCUSTOMER: <Customer>\n" + 32 pomlček + "\n".
func renderHeader(inv Invoice) string {
	// TODO
	return ""
}

// renderLines vykreslí položky faktury.
// Řádek: "<Description> | <Quantity> x <jednotková cena> = <cena za řádek>\n".
// Prázdný seznam → prázdný řetězec; součet počítej v centech, ne ve float64.
func renderLines(lines []InvoiceLine) string {
	// TODO
	return ""
}

// renderTotal vykreslí oddělovač a celkovou částku.
// Formát: 32 pomlček + "\nTOTAL: <částka z centů>\n"; bez položek TOTAL: 0.00.
func renderTotal(totalCents int) string {
	// TODO
	return ""
}
