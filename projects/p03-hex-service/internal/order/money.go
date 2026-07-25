package order

import (
	"fmt"
	"math"
	"strings"
)

// Money je peněžní částka v celých centech (haléřích) jedné měny.
//
// Obě pole jsou neexportovaná, takže částku nejde vyrobit jinak než přes
// NewMoney. Zero value je "nulová částka bez měny" a chová se jako neutrální
// prvek sčítání — díky tomu jde suma začít od `var total Money`.
//
// Proč ne float64: 0.1 + 0.2 != 0.3. Účetní to pozná dřív než ty.
type Money struct {
	cents    int64
	currency string
}

// NewMoney vytvoří částku. Měna musí být trojpísmenný kód (ISO 4217),
// částka nesmí být záporná.
func NewMoney(cents int64, currency string) (Money, error) {
	code := strings.ToUpper(strings.TrimSpace(currency))
	if len(code) != 3 || !isAlpha(code) {
		return Money{}, fmt.Errorf("%w: %q", ErrInvalidCurrency, currency)
	}
	if cents < 0 {
		return Money{}, fmt.Errorf("%w: %d", ErrNegativeAmount, cents)
	}
	return Money{cents: cents, currency: code}, nil
}

func isAlpha(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

// Cents vrací částku v celých centech.
func (m Money) Cents() int64 { return m.cents }

// Currency vrací kód měny, nebo prázdný řetězec pro nulovou hodnotu.
func (m Money) Currency() string { return m.currency }

// IsZero vrací true pro nulovou hodnotu bez měny.
func (m Money) IsZero() bool { return m.cents == 0 && m.currency == "" }

// Add sečte dvě částky téže měny.
func (m Money) Add(other Money) (Money, error) {
	switch {
	case m.IsZero():
		return other, nil
	case other.IsZero():
		return m, nil
	case m.currency != other.currency:
		return Money{}, fmt.Errorf("%w: %s + %s", ErrCurrencyMismatch, m.currency, other.currency)
	case m.cents > math.MaxInt64-other.cents:
		return Money{}, fmt.Errorf("%w: %d + %d", ErrAmountOverflow, m.cents, other.cents)
	}
	return Money{cents: m.cents + other.cents, currency: m.currency}, nil
}

// Mul vynásobí částku nezáporným celým číslem.
func (m Money) Mul(n int) (Money, error) {
	if n < 0 {
		return Money{}, fmt.Errorf("%w: násobek %d", ErrNegativeAmount, n)
	}
	if n != 0 && m.cents > math.MaxInt64/int64(n) {
		return Money{}, fmt.Errorf("%w: %d × %d", ErrAmountOverflow, m.cents, n)
	}
	return Money{cents: m.cents * int64(n), currency: m.currency}, nil
}

// String vrací částku ve tvaru "199.00 CZK".
func (m Money) String() string {
	if m.IsZero() {
		return "0.00"
	}
	return fmt.Sprintf("%d.%02d %s", m.cents/100, m.cents%100, m.currency)
}
