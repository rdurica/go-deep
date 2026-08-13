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

// New vytvoří částku z celých centů. Jediná cesta, jak Amount vyrobit zvenku balíčku.
func New(cents int64) Amount {
	return Amount{cents: cents}
}

// Cents vrací částku v celých centech. Zero value Amount{} vrací 0.
func (a Amount) Cents() int64 {
	return a.cents
}

// --- Stupeň: jednoduchý ---

// String implementuje fmt.Stringer. Formát s dvěma desetinnými místy a znaménkem:
// 0 → "0.00", 5 → "0.05", 1999 → "19.99", -1 → "-0.01", -250 → "-2.50".
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ — u záporných částek ztrácí znaménko a
// desetinná místa nejsou vždy dvě. Najdi chybu a oprav.
func (a Amount) String() string {
	c := a.cents
	if c < 0 {
		c = -c
	}
	return strconv.FormatInt(c/100, 10) + "." + strconv.FormatInt(c%100, 10)
}

// --- Stupeň: střední ---

// SumCents sečte částky a vrátí výsledek v centech.
// Sáhne přímo na pole cents (ne přes Cents()). Nil vstup dá 0.
func SumCents(amounts []Amount) int64 {
	// TODO
	return 0
}

// --- Stupeň: obtížný ---

// Split rozdělí částku na n dílů bez ztráty centů.
// Zbytek po dělení rozdá po jednom centu od prvního dílu: 1000/3 → 334, 333, 333.
// U záporných částek zbytek „ven od nuly": -250/3 → -84, -83, -83. n <= 0 → (nil, false).
func Split(a Amount, n int) ([]Amount, bool) {
	// TODO
	return nil, false
}
