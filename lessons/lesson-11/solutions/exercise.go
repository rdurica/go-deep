// Package solutions obsahuje referenční řešení lekce 11.
package solutions

import (
	"github.com/rdurica/go-deep/lessons/lesson-11/solutions/money"
)

// Amount je alias na money.Amount.
type Amount = money.Amount

// NewAmount je fasáda nad money.New.
func NewAmount(cents int64) Amount { return money.New(cents) }

// SumCents je fasáda nad money.SumCents.
func SumCents(amounts []Amount) int64 { return money.SumCents(amounts) }

// Split je fasáda nad money.Split.
func Split(a Amount, n int) ([]Amount, bool) { return money.Split(a, n) }
