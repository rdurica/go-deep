// Package money je doménový balíček pro peněžní částky v celých centech.
package money

import "strconv"

// Amount je peněžní částka v celých centech.
type Amount struct {
	cents int64
}

// New vytvoří částku z celých centů.
func New(cents int64) Amount {
	return Amount{cents: cents}
}

// Cents vrací částku v celých centech.
func (a Amount) Cents() int64 {
	return a.cents
}

// --- Stupeň: jednoduchý ---

// String implementuje fmt.Stringer.
func (a Amount) String() string {
	c := a.cents
	sign := ""
	if c < 0 {
		sign = "-"
		c = -c
	}
	whole := strconv.FormatInt(c/100, 10)
	frac := strconv.FormatInt(c%100, 10)
	if len(frac) == 1 {
		frac = "0" + frac
	}
	return sign + whole + "." + frac
}

// --- Stupeň: střední ---

// SumCents sečte částky přes neexportované pole cents.
func SumCents(amounts []Amount) int64 {
	var total int64
	for _, a := range amounts {
		total += a.cents
	}
	return total
}

// --- Stupeň: obtížný ---

// Split rozdělí částku na n dílů bez ztráty centů.
func Split(a Amount, n int) ([]Amount, bool) {
	if n <= 0 {
		return nil, false
	}

	base := a.cents / int64(n)
	rem := a.cents % int64(n)
	step := int64(1)
	if rem < 0 {
		step = -1
		rem = -rem
	}

	parts := make([]Amount, n)
	for i := range parts {
		c := base
		if int64(i) < rem {
			c += step
		}
		parts[i] = Amount{cents: c}
	}
	return parts, true
}
