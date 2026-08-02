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

// --- Stupeň: jednoduchý ---
// NewAmount je fasáda nad money.New. Hotová, neimplementuj ji.
func NewAmount(cents int64) Amount { return money.New(cents) }

// --- Stupeň: střední ---
// SumCents je fasáda nad money.SumCents. Hotová, neimplementuj ji.
func SumCents(amounts []Amount) int64 { return money.SumCents(amounts) }

// Split je fasáda nad money.Split. Hotová, neimplementuj ji.
func Split(a Amount, n int) ([]Amount, bool) { return money.Split(a, n) }

// --- Stupeň: obtížný ---
// TotalOf sečte částky a vrátí výsledek jako Amount.
// Prázdný nebo nil vstup dá nulovou částku. Smí použít jen veřejné API money.
func TotalOf(amounts []Amount) Amount {
	// TODO
	return *new(Amount)
}

// MustParse převede text na Amount: volitelné +/-, celá část, volitelná desetinná (. a 1–2 číslice).
// Platné: "0", "7", "19.99", "-2.5", "+3.05". Neplatné vstupy panikují.
// Bílé znaky se neořezávají (" 1" a "1 " jsou neplatné).
func MustParse(s string) Amount {
	// TODO
	return *new(Amount)
}
