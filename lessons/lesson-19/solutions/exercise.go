// Package solutions obsahuje referenční řešení lekce 19.
package solutions

import (
	"errors"
	"fmt"
	"sort"
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
func ParseUserID(raw string) (int, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, ErrEmptyID
	}

	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, ErrInvalidID
	}
	if id <= 0 {
		return 0, ErrNonPositiveID
	}
	return id, nil
}

// --- Stupeň: střední ---

// ParseUserIDs převede čárkami oddělený seznam ID na slice čísel.
func ParseUserIDs(raw string) ([]int, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	ids := make([]int, 0, len(parts))
	for i, part := range parts {
		id, err := ParseUserID(part)
		if err != nil {
			return nil, fmt.Errorf("id at index %d: %w", i, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// --- Stupeň: obtížný ---

// ProcessOrders agreguje objednávky do souhrnu.
func ProcessOrders(orders []Order) (Summary, error) {
	var sum Summary
	seen := make(map[string]bool, len(orders))

	for i, o := range orders {
		if o.ID == "" {
			return Summary{}, fmt.Errorf("order at index %d: %w", i, ErrMissingOrderID)
		}
		switch o.Status {
		case "cancelled":
			continue
		case "paid", "pending":
		default:
			return Summary{}, fmt.Errorf("order %q: %w: %q", o.ID, ErrUnknownStatus, o.Status)
		}

		items, total, err := sumItems(o)
		if err != nil {
			return Summary{}, err
		}

		sum.OrderCount++
		sum.ItemCount += items
		sum.TotalCents += total
		if o.Customer != "" && !seen[o.Customer] {
			seen[o.Customer] = true
			sum.Customers = append(sum.Customers, o.Customer)
		}
	}

	sort.Strings(sum.Customers)
	return sum, nil
}

// sumItems spočítá počet kusů a cenu jedné objednávky.
func sumItems(o Order) (count, totalCents int, err error) {
	for _, it := range o.Items {
		if it.Quantity <= 0 {
			return 0, 0, fmt.Errorf("order %q, item %q: %w", o.ID, it.SKU, ErrInvalidQuantity)
		}
		if it.UnitCents < 0 {
			return 0, 0, fmt.Errorf("order %q, item %q: %w", o.ID, it.SKU, ErrInvalidPrice)
		}
		count += it.Quantity
		totalCents += it.Quantity * it.UnitCents
	}
	return count, totalCents, nil
}
