// Package solutions obsahuje referenční řešení lekce 43.
package solutions

import (
	"errors"
	"sort"
	"sync"
	"sync/atomic"
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

// Counter je čítač bezpečný pro souběžné použití.
type Counter struct {
	// atomic.Int64 má užitečnou zero value, takže Counter nepotřebuje
	// konstruktor. Pro prostý čítač je atomika levnější než mutex.
	n atomic.Int64
}

// Inc zvýší čítač o jedna.
func (c *Counter) Inc() { c.n.Add(1) }

// Add přičte n (může být i záporné).
func (c *Counter) Add(n int64) { c.n.Add(n) }

// Value vrací aktuální hodnotu čítače.
func (c *Counter) Value() int64 { return c.n.Load() }

// Cache je mapa chráněná zámkem, bezpečná pro souběžné použití.
type Cache struct {
	mu    sync.RWMutex // zámek stojí hned nad daty, která chrání
	items map[string]string
	// calls drží rozdělaný výpočet pro klíč, aby se f nevolalo dvakrát.
	calls map[string]*computation
}

type computation struct {
	once sync.Once
	val  string
}

// NewCache vytvoří prázdnou cache.
func NewCache() *Cache {
	return &Cache{
		items: make(map[string]string),
		calls: make(map[string]*computation),
	}
}

// Get vrací hodnotu a true, pokud klíč existuje.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.items[key]
	return v, ok
}

// Set uloží hodnotu pod klíč.
func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

// Delete smaže klíč.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Len vrací počet uložených klíčů.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// GetOrCompute vrátí uloženou hodnotu, nebo ji spočítá pomocí f a uloží.
func (c *Cache) GetOrCompute(key string, f func() string) string {
	c.mu.RLock()
	v, ok := c.items[key]
	c.mu.RUnlock()
	if ok {
		return v
	}

	c.mu.Lock()
	if v, ok := c.items[key]; ok { // mezitím to mohl spočítat někdo jiný
		c.mu.Unlock()
		return v
	}
	comp, ok := c.calls[key]
	if !ok {
		comp = &computation{}
		c.calls[key] = comp
	}
	// Zámek pouštíme PŘED voláním f. Kdybychom ho drželi, zablokovali
	// bychom celou cache na dobu cizího výpočtu.
	c.mu.Unlock()

	comp.once.Do(func() {
		comp.val = f()
		c.mu.Lock()
		c.items[key] = comp.val
		delete(c.calls, key)
		c.mu.Unlock()
	})
	return comp.val
}

// account je jeden účet s vlastním zámkem, aby se převody mezi různými
// dvojicemi účtů navzájem neblokovaly.
type account struct {
	mu      sync.Mutex
	balance int64
}

// Bank drží zůstatky účtů a umí mezi nimi bezpečně převádět.
type Bank struct {
	accounts map[string]*account // po vytvoření se mapa už nemění
}

// NewBank vytvoří banku s počátečními zůstatky.
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

// Total vrací součet všech zůstatků v konzistentním okamžiku.
func (b *Bank) Total() int64 {
	names := b.sortedNames()
	// Zamykáme v abecedním pořadí — stejně jako Transfer, takže se ty dvě
	// operace nemohou zaklesnout.
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

// Transfer převede částku mezi účty.
func (b *Bank) Transfer(from, to string, amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	src, ok := b.accounts[from]
	if !ok {
		return ErrUnknownAccount
	}
	dst, ok := b.accounts[to]
	if !ok {
		return ErrUnknownAccount
	}
	if from == to {
		return nil // převod na sebe je no-op; jinak bychom zamykali dvakrát
	}

	// Klíč k odolnosti proti deadlocku: pevné pořadí zámků. Bez něj by se
	// souběžné převody A→B a B→A zaklesly do sebe.
	first, second := src, dst
	if from > to {
		first, second = dst, src
	}
	first.mu.Lock()
	defer first.mu.Unlock()
	second.mu.Lock()
	defer second.mu.Unlock()

	if src.balance < amount {
		return ErrInsufficientFunds
	}
	src.balance -= amount
	dst.balance += amount
	return nil
}

func (b *Bank) sortedNames() []string {
	names := make([]string, 0, len(b.accounts))
	for name := range b.accounts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
