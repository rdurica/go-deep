// Package solutions obsahuje referenční řešení checkpointu fáze 1 — balíček ledger.
package solutions

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
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

// --- Stupeň: obtížný ---
// String formátuje částku jako desetinné číslo se dvěma místy.
func (m Money) String() string {
	v := int64(m)
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	return fmt.Sprintf("%s%d.%02d", sign, v/100, v%100)
}

// --- Stupeň: jednoduchý ---
// Error implementuje rozhraní error.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("transakce %d: pole %q: %s", e.Index, e.Field, e.Reason)
}

// --- Stupeň: střední ---
// ParseTransactions načte pole transakcí z JSON a ověří je.
func ParseTransactions(r io.Reader) ([]Transaction, error) {
	var txs []Transaction
	if err := json.NewDecoder(r).Decode(&txs); err != nil {
		return nil, fmt.Errorf("ledger: dekódování: %w", err)
	}

	for i, tx := range txs {
		if err := validate(i, tx); err != nil {
			return nil, fmt.Errorf("ledger: %w", err)
		}
	}
	return txs, nil
}

func validate(i int, tx Transaction) error {
	switch {
	case strings.TrimSpace(tx.ID) == "":
		return &ValidationError{Index: i, Field: "id", Reason: "nesmí být prázdné"}
	case strings.TrimSpace(tx.Payee) == "":
		return &ValidationError{Index: i, Field: "payee", Reason: "nesmí být prázdné"}
	case strings.TrimSpace(tx.Category) == "":
		return &ValidationError{Index: i, Field: "category", Reason: "nesmí být prázdné"}
	case tx.Amount == 0:
		return &ValidationError{Index: i, Field: "amount", Reason: "nesmí být nula"}
	default:
		return nil
	}
}

// TotalsByCategory sečte částky podle kategorie.
func TotalsByCategory(txs []Transaction) map[string]Money {
	totals := make(map[string]Money, len(txs))
	for _, tx := range txs {
		totals[tx.Category] += tx.Amount
	}
	return totals
}

// GroupBy seskupí prvky podle klíče spočítaného funkcí key.
func GroupBy[T any, K comparable](items []T, key func(T) K) map[K][]T {
	groups := make(map[K][]T)
	for _, item := range items {
		k := key(item)
		groups[k] = append(groups[k], item)
	}
	return groups
}

// String implementuje fmt.Stringer.
func (r Report) String() string {
	top := r.Top
	if top == "" {
		top = "-"
	}
	return fmt.Sprintf("transakcí: %d, celkem: %s, top kategorie: %s", r.Count, r.Total, top)
}

// BuildReport načte transakce ze vstupu a sestaví z nich souhrn.
func BuildReport(r io.Reader) (Report, error) {
	txs, err := ParseTransactions(r)
	if err != nil {
		return Report{}, err
	}
	if len(txs) == 0 {
		return Report{}, ErrEmptyLedger
	}

	rep := Report{Count: len(txs)}
	for _, tx := range txs {
		rep.Total += tx.Amount
	}

	var bestTotal Money
	for category, total := range TotalsByCategory(txs) {
		if rep.Top == "" || total > bestTotal || (total == bestTotal && category < rep.Top) {
			rep.Top, bestTotal = category, total
		}
	}
	return rep, nil
}
