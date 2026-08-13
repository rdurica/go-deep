// Package exercise obsahuje cvičení lekce 43.
package exercise

import (
	"errors"
	"sort"
	"sync"
)

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
	n int64
}

// --- Stupeň: jednoduchý ---

// Inc zvýší čítač o jedna.
// Zero value Counter je použitelná bez konstruktoru; test běží s -race.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Inc, Add a Value nejsou chráněné proti datovému závodu.
// Najdi chybu a oprav — testy s -race před opravou padají.
func (c *Counter) Inc() {
	c.n++
}

// Add přičte n k čítači (n může být záporné).
func (c *Counter) Add(n int64) {
	c.n += n
}

// Value vrací aktuální hodnotu čítače.
func (c *Counter) Value() int64 {
	return c.n
}

// Cache je mapa chráněná zámkem, bezpečná pro souběžné použití.
type Cache struct {
	mu    sync.RWMutex
	items map[string]string
}

// NewCache vytvoří prázdnou cache s mapou chráněnou zámkem.
func NewCache() *Cache {
	// TODO
	return nil
}

// --- Stupeň: střední ---

// Get vrací hodnotu a true, pokud klíč existuje.
func (c *Cache) Get(key string) (string, bool) {
	// TODO
	return "", false
}

// Set uloží hodnotu pod klíč. Přepis existujícího klíče je povolený.
func (c *Cache) Set(key, value string) {
	// TODO
}

// Delete smaže klíč. Neexistující klíč je no-op.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Len vrací počet uložených klíčů v cache pod zámkem.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// --- Stupeň: obtížný ---

// account je jeden účet s vlastním zámkem.
type account struct {
	mu      sync.Mutex
	balance int64
}

// Bank drží zůstatky a umí bezpečné převody bez deadlocku.
type Bank struct {
	accounts map[string]*account
}

// NewBank vytvoří banku s počátečními zůstatky.
// Vstupní mapa se okopíruje; volající ji po vytvoření nesmí měnit.
func NewBank(balances map[string]int64) *Bank {
	accounts := make(map[string]*account, len(balances))
	for name, amount := range balances {
		accounts[name] = &account{balance: amount}
	}
	return &Bank{accounts: accounts}
}

// Balance vrací zůstatek účtu a true, pokud účet existuje.
func (b *Bank) Balance(name string) (int64, bool) {
	acc, ok := b.accounts[name]
	if !ok {
		return 0, false
	}
	acc.mu.Lock()
	defer acc.mu.Unlock()
	return acc.balance, true
}

func (b *Bank) sortedNames() []string {
	names := make([]string, 0, len(b.accounts))
	for name := range b.accounts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Total vrací součet všech zůstatků v konzistentním okamžiku (pomocný test).
func (b *Bank) Total() int64 {
	names := b.sortedNames()
	for _, name := range names {
		b.accounts[name].mu.Lock()
	}
	total := int64(0)
	for _, name := range names {
		total += b.accounts[name].balance
	}
	for i := len(names) - 1; i >= 0; i-- {
		b.accounts[names[i]].mu.Unlock()
	}
	return total
}

// Transfer převede amount z from do to. ErrUnknownAccount, ErrInvalidAmount
// (amount <= 0), ErrInsufficientFunds. Neúspěch nemění zůstatky.
// from == to je no-op. Zamykání v pevném pořadí účtů (bez deadlocku).
func (b *Bank) Transfer(from, to string, amount int64) error {
	// TODO
	return nil
}
