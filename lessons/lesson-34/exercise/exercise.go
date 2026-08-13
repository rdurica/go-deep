// Package exercise obsahuje cvičení lekce 34 — doménové typy a value objekty.
package exercise

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrInvalidCurrency  = errors.New("money: invalid currency")
	ErrCurrencyMismatch = errors.New("money: currency mismatch")
	ErrInvalidSplit     = errors.New("money: split count must be positive")
	ErrInvalidFormat    = errors.New("money: invalid format")
)

// Currency je kód měny podle ISO 4217.
type Currency string

// Money je peněžní částka v minoritních jednotkách.
type Money struct {
	cents    int64
	currency Currency
}

// --- Stupeň: jednoduchý ---

// NewMoney vytvoří částku. Měna musí mít tři velká písmena A–Z, jinak ErrInvalidCurrency.
func NewMoney(cents int64, c Currency) (Money, error) {
	// TODO
	return *new(Money), nil
}

// Cents vrací částku v minoritních jednotkách.
func (m Money) Cents() int64 { return m.cents }

// Currency vrací kód měny.
func (m Money) Currency() Currency { return m.currency }

// String implementuje fmt.Stringer: "19.99 EUR". Nulová Money{} → "0.00".
func (m Money) String() string {
	// TODO
	return ""
}

// --- Stupeň: střední ---

// Add vrací novou částku m+o. Různé měny → ErrCurrencyMismatch.
// Hodnotový receiver — operand m se nesmí změnit.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Mutuje příjemce místo vrácení nové hodnoty.
func (m Money) Add(o Money) (Money, error) {
	if m.currency != o.currency || (m.currency == "" && o.currency != "") || (m.currency != "" && o.currency == "") {
		return Money{}, ErrCurrencyMismatch
	}
	m.cents += o.cents
	return m, nil
}

// --- Stupeň: obtížný ---

// Allocate rozdělí částku na n dílů; součet dílů je přesně originál.
// n <= 0 → ErrInvalidSplit. Zbytek rozdá od začátku: 100/3 → 34,33,33.
func (m Money) Allocate(n int) ([]Money, error) {
	// TODO
	return nil, nil
}

var moneyRe = regexp.MustCompile(`^(-?)(\d+)\.(\d{2})$`)

func validCurrency(c Currency) bool {
	if len(c) != 3 {
		return false
	}
	for i := 0; i < 3; i++ {
		if c[i] < 'A' || c[i] > 'Z' {
			return false
		}
	}
	return true
}

func formatAmount(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}

func (m Money) sameCurrency(o Money) error {
	if m.currency != o.currency {
		return fmt.Errorf("money: %q vs %q: %w", m.currency, o.currency, ErrCurrencyMismatch)
	}
	if m.currency == "" && !o.IsZero() {
		return ErrCurrencyMismatch
	}
	if o.currency == "" && !m.IsZero() {
		return ErrCurrencyMismatch
	}
	return nil
}

// IsZero hlásí nulovou částku.
func (m Money) IsZero() bool { return m.cents == 0 }
