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
	// TODO: úkol A
	return ""
}

// Error implementuje rozhraní error.
func (e *ValidationError) Error() string {
	// TODO: úkol A
	return ""
}

// ParseTransactions načte pole transakcí z JSON a ověří je.
func ParseTransactions(r io.Reader) ([]Transaction, error) {
	// TODO: úkol B
	return nil, nil
}

// TotalsByCategory sečte částky podle kategorie.
func TotalsByCategory(txs []Transaction) map[string]Money {
	// TODO: úkol B
	return nil
}

// GroupBy seskupí prvky podle klíče spočítaného funkcí key.
func GroupBy[T any, K comparable](items []T, key func(T) K) map[K][]T {
	// TODO: úkol C
	return nil
}

// String implementuje fmt.Stringer.
func (r Report) String() string {
	// TODO: úkol C
	return ""
}

// BuildReport načte transakce ze vstupu a sestaví z nich souhrn.
func BuildReport(r io.Reader) (Report, error) {
	// TODO: úkol C
	return *new(Report), nil
}
