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

// --- Stupeň: obtížný ---
// String formátuje částku jako desetinné číslo se dvěma místy.
// 1999 → "19.99", 5 → "0.05", 0 → "0.00", -250 → "-2.50". Hodnotový receiver pro fmt.Stringer.
func (m Money) String() string {
	// TODO
	return ""
}

// --- Stupeň: jednoduchý ---
// Error implementuje rozhraní error.
// Text musí obsahovat index transakce, jméno pole a důvod, např.
// transakce 2: pole "amount": nesmí být nula.
func (e *ValidationError) Error() string {
	// TODO
	return ""
}

// --- Stupeň: střední ---
// ParseTransactions načte pole transakcí z JSON přes json.Decoder (ne Unmarshal nad řetězcem).
// Ověř id, payee, category (neprázdné po oříznutí) a amount (nesmí být nula; záporné OK).
// První neplatná transakce skončí chybou obalující *ValidationError (errors.As).
// Prázdné [] je platný vstup; rozbitý JSON je chyba.
func ParseTransactions(r io.Reader) ([]Transaction, error) {
	// TODO
	return nil, nil
}

// TotalsByCategory sečte částky podle kategorie.
// Prázdný vstup vrací prázdnou, ale nenilovou mapu.
func TotalsByCategory(txs []Transaction) map[string]Money {
	// TODO
	return nil
}

// GroupBy seskupí prvky podle klíče spočítaného funkcí key.
// Pořadí prvků uvnitř skupiny odpovídá vstupu; prázdný vstup → prázdná mapa.
func GroupBy[T any, K comparable](items []T, key func(T) K) map[K][]T {
	// TODO
	return nil
}

// String implementuje fmt.Stringer.
// Výstup: transakcí: N, celkem: X.XX, top kategorie: cat. Prázdná Top → "-".
func (r Report) String() string {
	// TODO
	return ""
}

// BuildReport načte transakce ze vstupu a sestaví z nich souhrn.
// Kniha bez transakcí → ErrEmptyLedger (errors.Is); chyba validace se propíše beze změny.
// Top kategorie má nejvyšší součet; při shodě rozhoduje abeceda.
func BuildReport(r io.Reader) (Report, error) {
	// TODO
	return *new(Report), nil
}
