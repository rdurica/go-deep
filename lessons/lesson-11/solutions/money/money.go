// Package money je doménový balíček pro peněžní částky v celých centech.
//
// Typ Amount schválně nemá exportované pole. Kdo je mimo tenhle balíček, se
// k centům dostane jen přes New a metodu Cents — nemůže tedy vyrobit částku
// jinak než konstruktorem. Uvnitř balíčku je naopak pole cents běžně dostupné,
// a to i na cizí instanci (viz SumCents).
package money

import "strconv"

// Amount je peněžní částka v celých centech. Zero value je nula měny.
type Amount struct {
	cents int64
}

// --- Stupeň: jednoduchý ---
// New vytvoří částku z celých centů.
func New(cents int64) Amount {
	return Amount{cents: cents}
}

// --- Stupeň: střední ---
// Cents vrací částku v celých centech.
func (a Amount) Cents() int64 {
	return a.cents
}

// String vrací částku ve tvaru s dvěma desetinnými místy, například "-2.50".
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

// --- Stupeň: obtížný ---
// SumCents sečte částky a vrátí výsledek v centech.
//
// Funkce je uvnitř balíčku money, takže smí sáhnout přímo na pole cents
// cizích instancí bez volání metody Cents.
func SumCents(amounts []Amount) int64 {
	var total int64
	for _, a := range amounts {
		total += a.cents // uvnitř balíčku je neexportované pole běžně dostupné
	}
	return total
}

// Split rozdělí částku na n dílů tak, aby se žádný cent neztratil.
// Zbytek po dělení se rozdá po jednom centu od prvního dílu.
// Pro n <= 0 vrací (nil, false).
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
