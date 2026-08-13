// Package exercise obsahuje cvičení lekce 19.
package exercise

import (
	"errors"
	"strconv"
	"strings"
)

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

// --- Stupeň: jednoduchý ---

// ParseUserID převede textové ID na kladné celé číslo.
// Ořízne bílé znaky; prázdný vstup → ErrEmptyID, nečíslo → ErrInvalidID, ≤0 → ErrNonPositiveID.
// Maximální hloubka zanoření 1, žádné else.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Používá zbytečné else a hluboké zanoření.
// Přepiš na early return se stejným chováním — testy před opravou procházejí,
// ale kód nesplňuje kontrakt stylu.
func ParseUserID(raw string) (int, error) {
	s := strings.TrimSpace(raw)
	if s != "" {
		id, err := strconv.Atoi(s)
		if err == nil {
			if id > 0 {
				return id, nil
			}
			return 0, ErrNonPositiveID
		}
		return 0, ErrInvalidID
	}
	return 0, ErrEmptyID
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
