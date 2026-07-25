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

// ParseUserID převede textové ID na kladné celé číslo.
func ParseUserID(raw string) (int, error) {
	// TODO: úkol A
	return 0, nil
}

// ParseUserIDs převede čárkami oddělený seznam ID na slice čísel.
func ParseUserIDs(raw string) ([]int, error) {
	// TODO: úkol A
	return nil, nil
}

// ProcessOrders agreguje objednávky do souhrnu.
func ProcessOrders(orders []Order) (Summary, error) {
	// TODO: úkol B
	return *new(Summary), nil
}

// RenderInvoice vykreslí fakturu do textové podoby.
func RenderInvoice(inv Invoice) string {
	// TODO: úkol C
	return ""
}

// renderHeader vykreslí hlavičku faktury včetně oddělovače.
func renderHeader(inv Invoice) string {
	// TODO: úkol C
	return ""
}

// renderLines vykreslí položky faktury.
func renderLines(lines []InvoiceLine) string {
	// TODO: úkol C
	return ""
}

// renderTotal vykreslí oddělovač a celkovou částku.
func renderTotal(totalCents int) string {
	// TODO: úkol C
	return ""
}
