// Package solutions obsahuje řešení lekce 34.
package solutions

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrInvalidCurrency  = errors.New("money: invalid currency")
	ErrCurrencyMismatch = errors.New("money: currency mismatch")
	ErrInvalidSplit     = errors.New("money: split count must be positive")
)

type Currency string

type Money struct {
	cents    int64
	currency Currency
}

func NewMoney(cents int64, c Currency) (Money, error) {
	if !validCurrency(c) {
		return Money{}, fmt.Errorf("money: %q: %w", string(c), ErrInvalidCurrency)
	}
	return Money{cents: cents, currency: c}, nil
}

func (m Money) Cents() int64       { return m.cents }
func (m Money) Currency() Currency { return m.currency }
func (m Money) IsZero() bool       { return m.cents == 0 }

func (m Money) String() string {
	amount := formatAmount(m.cents)
	if m.currency == "" {
		return amount
	}
	return amount + " " + string(m.currency)
}

func (m Money) Add(o Money) (Money, error) {
	if err := m.sameCurrency(o); err != nil {
		return Money{}, err
	}
	return Money{cents: m.cents + o.cents, currency: m.currency}, nil
}

func (m Money) Allocate(n int) ([]Money, error) {
	if n <= 0 {
		return nil, fmt.Errorf("money: %d: %w", n, ErrInvalidSplit)
	}
	base := m.cents / int64(n)
	parts := make([]Money, n)
	for i := range parts {
		parts[i] = Money{cents: base, currency: m.currency}
	}
	distribute(parts, m.cents-base*int64(n))
	return parts, nil
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

func distribute(parts []Money, remainder int64) {
	step := int64(1)
	if remainder < 0 {
		step = -1
		remainder = -remainder
	}
	for i := int64(0); i < remainder && i < int64(len(parts)); i++ {
		parts[i].cents += step
	}
}
