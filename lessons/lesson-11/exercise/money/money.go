// Package money je doménový balíček pro peněžní částky v celých centech.
//
// Typ Amount schválně nemá exportované pole. Kdo je mimo tenhle balíček, se
// k centům dostane jen přes New a metodu Cents — nemůže tedy vyrobit částku
// jinak než konstruktorem. Uvnitř balíčku je naopak pole cents běžně dostupné,
// a to i na cizí instanci (viz SumCents).
package money

// Amount je peněžní částka v celých centech. Zero value je nula měny.
type Amount struct {
	cents int64
}

// New vytvoří částku z celých centů.
func New(cents int64) Amount {
	panic("TODO: úkol A")
}

// Cents vrací částku v celých centech.
func (a Amount) Cents() int64 {
	panic("TODO: úkol A")
}

// String vrací částku ve tvaru s dvěma desetinnými místy, například "-2.50".
func (a Amount) String() string {
	panic("TODO: úkol A")
}

// SumCents sečte částky a vrátí výsledek v centech.
//
// Funkce je uvnitř balíčku money, takže smí sáhnout přímo na pole cents
// cizích instancí bez volání metody Cents.
func SumCents(amounts []Amount) int64 {
	panic("TODO: úkol C")
}

// Split rozdělí částku na n dílů tak, aby se žádný cent neztratil.
// Zbytek po dělení se rozdá po jednom centu od prvního dílu.
// Pro n <= 0 vrací (nil, false).
func Split(a Amount, n int) ([]Amount, bool) {
	panic("TODO: úkol C")
}
