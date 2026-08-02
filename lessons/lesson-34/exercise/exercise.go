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

// --- Stupeň: jednoduchý ---
// NewMoney vytvoří částku v minoritních jednotkách a zadané měně.
// Měna musí mít přesně tři velká písmena A–Z, jinak ErrInvalidCurrency a nulová Money.
// Záporná částka (dobropis) je legální.
func NewMoney(cents int64, c Currency) (Money, error) {
	// TODO
	return *new(Money), nil
}

// Cents vrací částku v minoritních jednotkách (centech, haléřích).
// Hodnotový receiver — kopie Money se nemění.
func (m Money) Cents() int64 {
	// TODO
	return 0
}

// Currency vrací kód měny částky jako pojmenovaný typ Currency.
// U nulové Money{} může být prázdný string.
func (m Money) Currency() Currency {
	// TODO
	return *new(Currency)
}

// String implementuje fmt.Stringer: "19.99 EUR", "-0.05 EUR", "1.00 EUR".
// Nulová Money{} nemá měnu a vypíše se jako "0.00" bez mezery na konci.
// Formát odpovídá ParseMoney — musí platit round-trip.
func (m Money) String() string {
	// TODO
	return ""
}

// --- Stupeň: střední ---
// Add vrací novou částku m+o. Různé měny → ErrCurrencyMismatch a nulová Money.
// Nulová hodnota bez měny se s EUR také neshoduje. Operand m zůstane nedotčen.
func (m Money) Add(o Money) (Money, error) {
	// TODO
	return *new(Money), nil
}

// Sub vrací novou částku m-o. Různé měny → ErrCurrencyMismatch a nulová Money.
// Nulová hodnota bez měny se s EUR také neshoduje. Operand m zůstane nedotčen.
func (m Money) Sub(o Money) (Money, error) {
	// TODO
	return *new(Money), nil
}

// Mul vrací novou částku vynásobenou celým číslem. Měna se nemění, bez chyby.
// Hodnotový receiver — původní Money se nemění.
func (m Money) Mul(n int64) Money {
	// TODO
	return *new(Money)
}

// IsZero hlásí, jestli je částka nulová. Měna se neposuzuje.
// Money{0, "EUR"} i Money{} jsou obě nulové.
func (m Money) IsZero() bool {
	// TODO
	return false
}

// Neg vrací částku s opačným znaménkem; m.Neg().Neg() == m.
// Měna zůstává stejná.
func (m Money) Neg() Money {
	// TODO
	return *new(Money)
}

// --- Stupeň: obtížný ---
// Compare vrací -1, 0 nebo 1 podle velikosti částek ve stejné měně.
// Různé měny → ErrCurrencyMismatch. Nulová Money bez měny se s EUR neshoduje.
func (m Money) Compare(o Money) (int, error) {
	// TODO
	return 0, nil
}

// Allocate rozdělí částku na n dílů; součet dílů je přesně originál.
// n <= 0 → ErrInvalidSplit. Zbytek rozdá od začátku: 100 na 3 → 34,33,33;
// -100 na 3 → -34,-33,-33. Žádné dva díly se nesmí lišit o víc než 1.
func (m Money) Allocate(n int) ([]Money, error) {
	// TODO
	return nil, nil
}

// AllocateRatio rozdělí částku v poměrech; součet dílů je přesně m.
// Prázdný slice, záporný poměr nebo nulový součet → ErrInvalidRatios.
// 5 v poměru 3:7 → 2, 3. Zbytek rozdá od začátku jako u Allocate.
func (m Money) AllocateRatio(ratios []int) ([]Money, error) {
	// TODO
	return nil, nil
}

// ParseMoney přečte "19.99 EUR": volitelné mínus, tečka, přesně dvě desetinná
// místa, mezera, měna. Okolní bílé znaky ignoruj. Jiný tvar → ErrInvalidFormat.
// ParseMoney(m.String()) == m pro každou platnou částku.
func ParseMoney(s string) (Money, error) {
	// TODO
	return *new(Money), nil
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
