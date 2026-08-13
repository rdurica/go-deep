// Package exercise obsahuje kumulativní cvičení checkpointu fáze 1 — balíček ledger.
package exercise

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

// String formátuje částku jako desetinné číslo se dvěma místy.
// 1999 → "19.99", 5 → "0.05", 0 → "0.00", -250 → "-2.50".
func (m Money) String() string {
	v := int64(m)
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	return fmt.Sprintf("%s%d.%02d", sign, v/100, v%100)
}

// String implementuje fmt.Stringer.
// Výstup: transakcí: N, celkem: X.XX, top kategorie: cat. Prázdná Top → "-".
func (r Report) String() string {
	top := r.Top
	if top == "" {
		top = "-"
	}
	return fmt.Sprintf("transakcí: %d, celkem: %s, top kategorie: %s", r.Count, r.Total, top)
}

// --- Stupeň: jednoduchý ---

// Error implementuje rozhraní error.
// Text musí obsahovat index transakce, jméno pole a důvod, např.
// transakce 2: pole "amount": nesmí být nula.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Vrací jen důvod, bez indexu a pole.
// Najdi chybu a oprav — testy před opravou padají.
func (e *ValidationError) Error() string {
	return e.Reason
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

// --- Stupeň: obtížný ---

// BuildReport načte transakce ze vstupu a sestaví z nich souhrn.
// Kniha bez transakcí → ErrEmptyLedger (errors.Is); chyba validace se propíše beze změny.
// Top kategorie má nejvyšší součet; při shodě rozhoduje abeceda.
func BuildReport(r io.Reader) (Report, error) {
	// TODO
	return *new(Report), nil
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

// decodeTransactions je interní helper — student ho nemění.
func decodeTransactions(r io.Reader) ([]Transaction, error) {
	var txs []Transaction
	if err := json.NewDecoder(r).Decode(&txs); err != nil {
		return nil, fmt.Errorf("ledger: dekódování: %w", err)
	}
	return txs, nil
}
