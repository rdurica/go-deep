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

// Counter je čítač bezpečný pro souběžné použití. Zero value je připravená
// k práci — žádný konstruktor není potřeba.
type Counter struct {
	// TODO: doplň pole (úkol A)
}

// Inc zvýší čítač o jedna.
func (c *Counter) Inc() {
	panic("TODO: úkol A")
}

// Add přičte n (může být i záporné).
func (c *Counter) Add(n int64) {
	panic("TODO: úkol A")
}

// Value vrací aktuální hodnotu čítače.
func (c *Counter) Value() int64 {
	panic("TODO: úkol A")
}

// Cache je mapa chráněná zámkem, bezpečná pro souběžné použití.
type Cache struct {
	// TODO: doplň pole (úkol B)
}

// NewCache vytvoří prázdnou cache.
func NewCache() *Cache {
	panic("TODO: úkol B")
}

// Get vrací hodnotu a true, pokud klíč existuje.
func (c *Cache) Get(key string) (string, bool) {
	panic("TODO: úkol B")
}

// Set uloží hodnotu pod klíč.
func (c *Cache) Set(key, value string) {
	panic("TODO: úkol B")
}

// Delete smaže klíč. Mazání neexistujícího klíče nic nedělá.
func (c *Cache) Delete(key string) {
	panic("TODO: úkol B")
}

// Len vrací počet uložených klíčů.
func (c *Cache) Len() int {
	panic("TODO: úkol B")
}

// GetOrCompute vrátí uloženou hodnotu, nebo ji spočítá pomocí f a uloží.
// Pro daný klíč se f zavolá právě jednou, i když GetOrCompute běží souběžně.
func (c *Cache) GetOrCompute(key string, f func() string) string {
	panic("TODO: úkol B")
}

// Bank drží zůstatky účtů a umí mezi nimi bezpečně převádět.
type Bank struct {
	// TODO: doplň pole (úkol C)
}

// NewBank vytvoří banku s počátečními zůstatky.
func NewBank(balances map[string]int64) *Bank {
	panic("TODO: úkol C")
}

// Balance vrací zůstatek účtu a true, pokud účet existuje.
func (b *Bank) Balance(account string) (int64, bool) {
	panic("TODO: úkol C")
}

// Total vrací součet všech zůstatků v konzistentním okamžiku.
func (b *Bank) Total() int64 {
	panic("TODO: úkol C")
}

// Transfer převede částku mezi účty. Musí být odolný vůči souběžným
// převodům v opačném směru, tedy bez deadlocku.
func (b *Bank) Transfer(from, to string, amount int64) error {
	panic("TODO: úkol C")
}
