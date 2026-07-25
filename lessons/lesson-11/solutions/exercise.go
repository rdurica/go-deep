// Package solutions obsahuje referenční řešení lekce 11.
//
// Tenhle balíček je konzument balíčku money — vidí z něj jen to, co má velké
// počáteční písmeno. Pole Amount.cents pro něj neexistuje.
package solutions

import (
	"strconv"
	"strings"

	"github.com/rdurica/go-deep/lessons/lesson-11/solutions/money"
)

// Amount je alias na money.Amount, aby testy nemusely importovat podbalíček.
// Alias (znaménko =) není nový typ, jen druhé jméno pro ten samý typ.
type Amount = money.Amount

// NewAmount je fasáda nad money.New. Hotová, neimplementuj ji.
func NewAmount(cents int64) Amount { return money.New(cents) }

// SumCents je fasáda nad money.SumCents. Hotová, neimplementuj ji.
func SumCents(amounts []Amount) int64 { return money.SumCents(amounts) }

// Split je fasáda nad money.Split. Hotová, neimplementuj ji.
func Split(a Amount, n int) ([]Amount, bool) { return money.Split(a, n) }

// TotalOf sečte částky a vrátí výsledek jako Amount.
// Smí použít jen veřejné API balíčku money.
func TotalOf(amounts []Amount) Amount {
	var total int64
	for _, a := range amounts {
		total += a.Cents() // zvenku jen přes metodu, pole cents není vidět
	}
	return money.New(total)
}

// MustParse převede zápis částky ("19.99", "-2.5", "7") na Amount.
// Při neplatném vstupu panikuje.
func MustParse(s string) Amount {
	cents, ok := parseCents(s)
	if !ok {
		panic("money: neplatná částka " + strconv.Quote(s))
	}
	return money.New(cents)
}

// parseCents je neexportovaný pomocník — mimo balíček ho nikdo nezavolá.
func parseCents(s string) (int64, bool) {
	negative := false
	switch {
	case strings.HasPrefix(s, "-"):
		negative = true
		s = s[1:]
	case strings.HasPrefix(s, "+"):
		s = s[1:]
	}

	whole, frac, hasFrac := strings.Cut(s, ".")
	if !isDigits(whole) {
		return 0, false
	}
	if hasFrac && (len(frac) > 2 || !isDigits(frac)) {
		return 0, false
	}

	units, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, false
	}

	var subunits int64
	if hasFrac {
		if len(frac) == 1 {
			frac += "0"
		}
		subunits, err = strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return 0, false
		}
	}

	cents := units*100 + subunits
	if negative {
		cents = -cents
	}
	return cents, true
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
