// Package exercise obsahuje cvičení lekce 43.
package exercise

import "errors"

// Chyby, které vrací Bank.Transfer.
var (
	// ErrUnknownAccount znamená, že zdrojový nebo cílový účet neexistuje.
	ErrUnknownAccount = errors.New("unknown account")
	// ErrInvalidAmount znamená nekladnou částku.
	ErrInvalidAmount = errors.New("invalid amount")
	// ErrInsufficientFunds znamená, že na zdrojovém účtu není dost peněz.
	ErrInsufficientFunds = errors.New("insufficient funds")
)

// Counter je čítač bezpečný pro souběžné použití. Zero value je použitelná.
type Counter struct {
	// TODO
}

// --- Stupeň: jednoduchý ---
// Inc zvýší čítač o jedna atomicky.
// Zero value Counter je použitelná bez konstruktoru; test běží s -race.
func (c *Counter) Inc() {
	// TODO
}

// Add přičte n k čítači atomicky (n může být záporné).
// Souběžné volání z více goroutin musí být bezpečné pod -race.
func (c *Counter) Add(n int64) {
	// TODO
}

// Value vrací aktuální hodnotu čítače atomicky.
// Čtení je bezpečné i během souběžných zápisů z jiných goroutin.
func (c *Counter) Value() int64 {
	// TODO
	return 0
}

// Cache je mapa chráněná zámkem, bezpečná pro souběžné použití.
type Cache struct {
	// TODO
}

// NewCache vytvoří prázdnou cache s mapou chráněnou zámkem.
// Bezpečná pro souběžný přístup z více goroutin.
func NewCache() *Cache {
	// TODO
	return nil
}

// --- Stupeň: střední ---
// Get vrací hodnotu a true, pokud klíč existuje.
// Neexistující klíč vrací prázdný řetězec a false.
func (c *Cache) Get(key string) (string, bool) {
	// TODO
	return "", false
}

// Set uloží hodnotu pod klíč.
// Přepis existujícího klíče je povolený.
// Pod mutexem cache (jako Len); bezpečné pod -race.
func (c *Cache) Set(key, value string) {
	// TODO
}

// Delete smaže klíč. Neexistující klíč je no-op.
// Pod mutexem cache (jako Len); bezpečné pod -race.
func (c *Cache) Delete(key string) {
	// TODO
}

// Len vrací počet uložených klíčů v cache pod zámkem.
func (c *Cache) Len() int {
	// TODO
	return 0
}

// --- Stupeň: obtížný ---
// GetOrCompute vrátí uloženou hodnotu nebo spočítá přes f a uloží.
// Pro daný klíč se f zavolá právě jednou (i při sto goroutinách).
// Během f nesmíš držet zámek cache.
func (c *Cache) GetOrCompute(key string, f func() string) string {
	// TODO
	return ""
}

// Bank drží zůstatky a umí bezpečné převody bez deadlocku.
type Bank struct {
	// TODO
}

// NewBank vytvoří banku s počátečními zůstatky.
// Vstupní mapa se okopíruje; volající ji po vytvoření nesmí měnit.
func NewBank(balances map[string]int64) *Bank {
	// TODO
	return nil
}

// Balance vrací zůstatek účtu a true, pokud účet existuje.
// Neznámý účet vrací 0 a false.
func (b *Bank) Balance(account string) (int64, bool) {
	// TODO
	return 0, false
}

// Total vrací součet všech zůstatků v konzistentním okamžiku.
// Test ho volá souběžně s převody; jiná hodnota než počáteční suma je chyba.
func (b *Bank) Total() int64 {
	// TODO
	return 0
}

// Transfer převede amount z from do to. ErrUnknownAccount, ErrInvalidAmount
// (amount <= 0), ErrInsufficientFunds. Neúspěch nemění zůstatky.
// from == to je no-op. Zamykání v pevném pořadí účtů (bez deadlocku).
func (b *Bank) Transfer(from, to string, amount int64) error {
	// TODO
	return nil
}
