// Package exercise obsahuje kumulativní cvičení checkpointu fáze 1 — balíček ledger.
package exercise

import (
	"errors"
	"io"
)

// Money je částka v celých centech.
type Money int64

// ErrEmptyLedger signalizuje knihu bez jediné transakce.
var ErrEmptyLedger = errors.New("ledger: prázdná kniha")

// ValidationError popisuje neplatnou hodnotu konkrétního pole transakce.
type ValidationError struct {
	Index  int
	Field  string
	Reason string
}

// Transaction je jedna transakce načtená z JSON.
type Transaction struct {
	ID       string `json:"id"`
	Payee    string `json:"payee"`
	Amount   Money  `json:"amount"`
	Category string `json:"category"`
}

// Report je souhrn nad knihou transakcí.
type Report struct {
	Count int
	Total Money
	Top   string
}

// String formátuje částku jako desetinné číslo se dvěma místy.
func (m Money) String() string {
	panic("TODO: úkol A")
}

// Error implementuje rozhraní error.
func (e *ValidationError) Error() string {
	panic("TODO: úkol A")
}

// ParseTransactions načte pole transakcí z JSON a ověří je.
func ParseTransactions(r io.Reader) ([]Transaction, error) {
	panic("TODO: úkol B")
}

// TotalsByCategory sečte částky podle kategorie.
func TotalsByCategory(txs []Transaction) map[string]Money {
	panic("TODO: úkol B")
}

// GroupBy seskupí prvky podle klíče spočítaného funkcí key.
func GroupBy[T any, K comparable](items []T, key func(T) K) map[K][]T {
	panic("TODO: úkol C")
}

// String implementuje fmt.Stringer.
func (r Report) String() string {
	panic("TODO: úkol C")
}

// BuildReport načte transakce ze vstupu a sestaví z nich souhrn.
func BuildReport(r io.Reader) (Report, error) {
	panic("TODO: úkol C")
}
