// Package solutions obsahuje řešení lekce 34 — doménové typy a value objekty.
//
// Money je value objekt: neměnný, porovnatelný operátorem ==, použitelný jako
// klíč mapy a bez jediného float64 uvnitř.
package solutions

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Chyby práce s penězi.
var (
	// ErrInvalidCurrency hlásí kód měny, který nemá tvar tří velkých písmen.
	ErrInvalidCurrency = errors.New("money: invalid currency")
	// ErrCurrencyMismatch hlásí pokus o operaci nad dvěma různými měnami.
	ErrCurrencyMismatch = errors.New("money: currency mismatch")
	// ErrInvalidSplit hlásí nekladný počet dílů.
	ErrInvalidSplit = errors.New("money: split count must be positive")
	// ErrInvalidRatios hlásí prázdné, záporné nebo nulové poměry.
	ErrInvalidRatios = errors.New("money: invalid ratios")
	// ErrInvalidFormat hlásí řetězec, který nejde přečíst jako částka.
	ErrInvalidFormat = errors.New("money: invalid format")
)

// Currency je kód měny podle ISO 4217, například "EUR". Je to pojmenovaný typ
// nad stringem, takže ho nejde omylem zaměnit za jiný string.
type Currency string

// Money je peněžní částka v minoritních jednotkách (centech, haléřích).
//
// Obě pole jsou neexportovaná, takže hodnotu nelze zvenčí rozbít. Struct
// obsahuje jen porovnatelné typy, proto funguje == i použití jako klíč mapy.
type Money struct {
	cents    int64
	currency Currency
}

// NewMoney vytvoří částku. Měna musí mít tvar přesně tří velkých písmen A–Z.
func NewMoney(cents int64, c Currency) (Money, error) {
	if !validCurrency(c) {
		return Money{}, fmt.Errorf("money: %q: %w", string(c), ErrInvalidCurrency)
	}
	return Money{cents: cents, currency: c}, nil
}

// Cents vrací částku v minoritních jednotkách.
func (m Money) Cents() int64 { return m.cents }

// Currency vrací měnu částky.
func (m Money) Currency() Currency { return m.currency }

// String implementuje fmt.Stringer, například "19.99 EUR" nebo "-0.05 EUR".
// Nulová hodnota Money nemá měnu a vypíše se jako "0.00".
func (m Money) String() string {
	amount := formatAmount(m.cents)
	if m.currency == "" {
		return amount
	}
	return amount + " " + string(m.currency)
}

// Add vrací novou částku m+o. Různé měny jsou ErrCurrencyMismatch.
func (m Money) Add(o Money) (Money, error) {
	if err := m.sameCurrency(o); err != nil {
		return Money{}, err
	}
	return Money{cents: m.cents + o.cents, currency: m.currency}, nil
}

// Sub vrací novou částku m-o. Různé měny jsou ErrCurrencyMismatch.
func (m Money) Sub(o Money) (Money, error) {
	if err := m.sameCurrency(o); err != nil {
		return Money{}, err
	}
	return Money{cents: m.cents - o.cents, currency: m.currency}, nil
}

// Mul vrací novou částku vynásobenou celým číslem. Měna se nemění,
// takže tahle operace nemůže selhat.
func (m Money) Mul(n int64) Money {
	return Money{cents: m.cents * n, currency: m.currency}
}

// IsZero hlásí, jestli je částka nulová. Měna se neposuzuje.
func (m Money) IsZero() bool { return m.cents == 0 }

// Neg vrací částku s opačným znaménkem.
func (m Money) Neg() Money {
	return Money{cents: -m.cents, currency: m.currency}
}

// Compare vrací -1, 0 nebo 1 podle toho, jestli je m menší, stejná nebo větší
// než o. Různé měny jsou ErrCurrencyMismatch.
func (m Money) Compare(o Money) (int, error) {
	if err := m.sameCurrency(o); err != nil {
		return 0, err
	}
	switch {
	case m.cents < o.cents:
		return -1, nil
	case m.cents > o.cents:
		return 1, nil
	default:
		return 0, nil
	}
}

// Allocate rozdělí částku na n dílů tak, že jejich součet je přesně m.
// Zbylé minoritní jednotky rozdá po jedné od začátku: 100 centů na tři díly
// dá 34, 33, 33.
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

// AllocateRatio rozdělí částku v zadaných poměrech. Součet dílů je přesně m.
// Poměry musí být nezáporné a jejich součet kladný.
func (m Money) AllocateRatio(ratios []int) ([]Money, error) {
	if len(ratios) == 0 {
		return nil, fmt.Errorf("money: prázdné poměry: %w", ErrInvalidRatios)
	}
	var total int64
	for _, r := range ratios {
		if r < 0 {
			return nil, fmt.Errorf("money: záporný poměr %d: %w", r, ErrInvalidRatios)
		}
		total += int64(r)
	}
	if total == 0 {
		return nil, fmt.Errorf("money: nulový součet poměrů: %w", ErrInvalidRatios)
	}

	parts := make([]Money, len(ratios))
	var assigned int64
	for i, r := range ratios {
		share := m.cents * int64(r) / total
		parts[i] = Money{cents: share, currency: m.currency}
		assigned += share
	}
	distribute(parts, m.cents-assigned)
	return parts, nil
}

// ParseMoney přečte částku ve tvaru "19.99 EUR" — přesně dvě desetinná místa
// a mezera před kódem měny.
func ParseMoney(s string) (Money, error) {
	fields := strings.Fields(s)
	if len(fields) != 2 {
		return Money{}, fmt.Errorf("money: %q: %w", s, ErrInvalidFormat)
	}
	groups := moneyRe.FindStringSubmatch(fields[0])
	if groups == nil {
		return Money{}, fmt.Errorf("money: %q: %w", s, ErrInvalidFormat)
	}

	whole, err := strconv.ParseInt(groups[2], 10, 64)
	if err != nil {
		return Money{}, fmt.Errorf("money: %q: %w", s, ErrInvalidFormat)
	}
	frac, err := strconv.ParseInt(groups[3], 10, 64)
	if err != nil {
		return Money{}, fmt.Errorf("money: %q: %w", s, ErrInvalidFormat)
	}
	if whole > (math.MaxInt64-frac)/100 {
		return Money{}, fmt.Errorf("money: %q: %w", s, ErrInvalidFormat)
	}

	cents := whole*100 + frac
	if groups[1] == "-" {
		cents = -cents
	}
	return NewMoney(cents, Currency(fields[1]))
}

// moneyRe popisuje část s částkou: volitelné mínus, celá část, tečka
// a přesně dvě desetinná místa.
var moneyRe = regexp.MustCompile(`^(-?)(\d+)\.(\d{2})$`)

// validCurrency hlásí, jestli kód měny má tvar tří velkých písmen A–Z.
func validCurrency(c Currency) bool {
	if len(c) != 3 {
		return false
	}
	for i := 0; i < len(c); i++ {
		if c[i] < 'A' || c[i] > 'Z' {
			return false
		}
	}
	return true
}

// formatAmount naformátuje minoritní jednotky na dvě desetinná místa.
func formatAmount(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}

// sameCurrency ohlásí ErrCurrencyMismatch, pokud se měny liší.
func (m Money) sameCurrency(o Money) error {
	if m.currency != o.currency {
		return fmt.Errorf("money: %q vs %q: %w", string(m.currency), string(o.currency), ErrCurrencyMismatch)
	}
	return nil
}

// distribute rozdá zbytek po jedné minoritní jednotce od začátku slice.
// Zbytek může být záporný, pak se od začátku po jedné odebírá.
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
