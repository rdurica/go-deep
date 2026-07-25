// Package exercise obsahuje cvičení lekce 34 — doménové typy a value objekty.
//
// Money je value objekt: neměnný, porovnatelný operátorem ==, použitelný jako
// klíč mapy a bez jediného float64 uvnitř.
package exercise

import (
	"errors"
	"fmt"
	"regexp"
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
	panic("TODO: úkol A")
}

// Cents vrací částku v minoritních jednotkách.
func (m Money) Cents() int64 {
	panic("TODO: úkol A")
}

// Currency vrací měnu částky.
func (m Money) Currency() Currency {
	panic("TODO: úkol A")
}

// String implementuje fmt.Stringer, například "19.99 EUR" nebo "-0.05 EUR".
// Nulová hodnota Money nemá měnu a vypíše se jako "0.00".
func (m Money) String() string {
	panic("TODO: úkol A")
}

// Add vrací novou částku m+o. Různé měny jsou ErrCurrencyMismatch.
func (m Money) Add(o Money) (Money, error) {
	panic("TODO: úkol B")
}

// Sub vrací novou částku m-o. Různé měny jsou ErrCurrencyMismatch.
func (m Money) Sub(o Money) (Money, error) {
	panic("TODO: úkol B")
}

// Mul vrací novou částku vynásobenou celým číslem. Měna se nemění,
// takže tahle operace nemůže selhat.
func (m Money) Mul(n int64) Money {
	panic("TODO: úkol B")
}

// IsZero hlásí, jestli je částka nulová. Měna se neposuzuje.
func (m Money) IsZero() bool {
	panic("TODO: úkol B")
}

// Neg vrací částku s opačným znaménkem.
func (m Money) Neg() Money {
	panic("TODO: úkol B")
}

// Compare vrací -1, 0 nebo 1 podle toho, jestli je m menší, stejná nebo větší
// než o. Různé měny jsou ErrCurrencyMismatch.
func (m Money) Compare(o Money) (int, error) {
	panic("TODO: úkol B")
}

// Allocate rozdělí částku na n dílů tak, že jejich součet je přesně m.
// Zbylé minoritní jednotky rozdá po jedné od začátku: 100 centů na tři díly
// dá 34, 33, 33.
func (m Money) Allocate(n int) ([]Money, error) {
	panic("TODO: úkol C")
}

// AllocateRatio rozdělí částku v zadaných poměrech. Součet dílů je přesně m.
// Poměry musí být nezáporné a jejich součet kladný.
func (m Money) AllocateRatio(ratios []int) ([]Money, error) {
	panic("TODO: úkol C")
}

// ParseMoney přečte částku ve tvaru "19.99 EUR" — přesně dvě desetinná místa
// a mezera před kódem měny.
func ParseMoney(s string) (Money, error) {
	panic("TODO: úkol C")
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
