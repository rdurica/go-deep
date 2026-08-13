// Package exercise obsahuje cvičení lekce 11.
//
// Tenhle balíček je konzument balíčku money — vidí z něj jen to, co má velké
// počáteční písmeno. Pole Amount.cents pro něj neexistuje.
package exercise

import (
	"github.com/rdurica/go-deep/lessons/lesson-11/exercise/money"
)

// Amount je alias na money.Amount, aby testy nemusely importovat podbalíček.
type Amount = money.Amount

// NewAmount je fasáda nad money.New. Hotová, neimplementuj ji.
func NewAmount(cents int64) Amount { return money.New(cents) }

// SumCents je fasáda nad money.SumCents. Hotová, neimplementuj ji.
func SumCents(amounts []Amount) int64 { return money.SumCents(amounts) }

// Split je fasáda nad money.Split. Hotová, neimplementuj ji.
func Split(a Amount, n int) ([]Amount, bool) { return money.Split(a, n) }
