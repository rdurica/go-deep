// Package exercise obsahuje cvičení lekce 44.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Obsahuje datové závody, které máš najít
// a opravit. Testy s -race musí před opravou padat a po opravě procházet.
package exercise

import (
	"sync"
	"testing"
)

// Zátěž, kterou pouští StressTest. Testy s těmito čísly počítají.
const (
	// StressGoroutines je počet goroutin, ve kterých StressTest volá f.
	StressGoroutines = 50
	// StressIterations je počet volání f v každé goroutině.
	StressIterations = 100
)

// SafeIncrement spustí n goroutin, z nichž každá zvýší společný čítač o jedna,
// a vrátí jeho výslednou hodnotu.
//
// ZÁVODNÍ implementace — oprav ji.
func SafeIncrement(n int) int {
	counter := 0
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			counter++ // čtení, přičtení a zápis nejsou jedna operace
		}()
	}
	wg.Wait()
	return counter
}

// Registry je registr hodnot bezpečný pro souběžné použití. Mapa se vytváří
// líně, až při prvním zápisu, a zero value musí být použitelná.
//
// ZÁVODNÍ implementace — oprav ji.
type Registry struct {
	items map[string]int
}

// Set uloží hodnotu pod klíč.
//
// ZÁVODNÍ implementace — oprav ji.
func (r *Registry) Set(key string, value int) {
	if r.items == nil { // líná inicializace bez jakékoli synchronizace
		r.items = make(map[string]int)
	}
	r.items[key] = value
}

// Get vrací hodnotu a true, pokud klíč existuje.
//
// ZÁVODNÍ implementace — oprav ji.
func (r *Registry) Get(key string) (int, bool) {
	v, ok := r.items[key]
	return v, ok
}

// Len vrací počet uložených klíčů.
//
// ZÁVODNÍ implementace — oprav ji.
func (r *Registry) Len() int {
	return len(r.items)
}

// ParallelAppend vrátí druhé mocniny všech čísel (v libovolném pořadí),
// spočítané souběžně.
//
// ZÁVODNÍ implementace — oprav ji.
func ParallelAppend(nums []int) []int {
	out := make([]int, 0, len(nums))
	var wg sync.WaitGroup
	wg.Add(len(nums))
	for _, n := range nums {
		go func() {
			defer wg.Done()
			out = append(out, n*n) // append zapisuje sdílenou hlavičku slice
		}()
	}
	wg.Wait()
	return out
}

// StressTest zavolá f ve StressGoroutines goroutinách, každá StressIterations
// krát, a počká na jejich dokončení. Slouží k tomu, aby závod dostal šanci se
// projevit a aby ho race detektor uviděl.
func StressTest(t *testing.T, f func()) {
	// TODO: úkol C
	t.Fatal("TODO: úkol C")
}

// Snapshot je konfigurace načtená v jednom okamžiku. Checksum musí vždy
// odpovídat len(Endpoint) + Timeout — nekonzistentní dvojice znamená,
// že čtenář viděl rozepsaný zápis.
type Snapshot struct {
	Endpoint string
	Timeout  int
	Checksum int
}

// NewSnapshot vytvoří konzistentní snapshot.
func NewSnapshot(endpoint string, timeout int) Snapshot {
	return Snapshot{
		Endpoint: endpoint,
		Timeout:  timeout,
		Checksum: len(endpoint) + timeout,
	}
}

// Consistent hlásí, jestli snapshot dává smysl jako celek.
func (s Snapshot) Consistent() bool {
	return s.Checksum == len(s.Endpoint)+s.Timeout
}

// Config drží aktuální konfiguraci, kterou na pozadí přepisuje reloader,
// zatímco ji ostatní goroutiny čtou.
//
// ZÁVODNÍ implementace — oprav ji.
type Config struct {
	snap Snapshot
}

// Store uloží novou konfiguraci.
//
// ZÁVODNÍ implementace — oprav ji.
func (c *Config) Store(s Snapshot) {
	// Tři samostatné zápisy: čtenář může uvidět půlku staré a půlku nové
	// konfigurace, i kdyby byl každý zápis sám o sobě atomický.
	c.snap.Endpoint = s.Endpoint
	c.snap.Timeout = s.Timeout
	c.snap.Checksum = s.Checksum
}

// Load vrací aktuální konfiguraci.
//
// ZÁVODNÍ implementace — oprav ji.
func (c *Config) Load() Snapshot {
	return c.snap
}

// StartReloader spustí goroutinu, která dokola ukládá snapshoty ze seznamu,
// dokud volající nezavře stop. Vrácený kanál se zavře po jejím konci.
func (c *Config) StartReloader(stop <-chan struct{}, snaps []Snapshot) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			if len(snaps) > 0 {
				c.Store(snaps[i%len(snaps)])
			}
		}
	}()
	return done
}
