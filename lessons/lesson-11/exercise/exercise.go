// Package exercise obsahuje cvičení lekce 11.
//
// Tenhle balíček je konzument balíčku money — vidí z něj jen to, co má velké
// počáteční písmeno. Pole Amount.cents pro něj neexistuje.
package exercise

import (
	"github.com/rdurica/go-deep/lessons/lesson-11/exercise/money"
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
	// TODO: úkol B
	return *new(Amount)
}

// MustParse převede zápis částky ("19.99", "-2.5", "7") na Amount.
// Při neplatném vstupu panikuje.
func MustParse(s string) Amount {
	// TODO: úkol B
	return *new(Amount)
}
