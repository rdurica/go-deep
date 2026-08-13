// Package exercise obsahuje cvičení lekce 44.
package exercise

import (
	"sync"
	"testing"
)

// Zátěž, kterou pouští StressTest.
const (
	// StressGoroutines je počet goroutin, ve kterých StressTest volá f.
	StressGoroutines = 50
	// StressIterations je počet volání f v každé goroutině.
	StressIterations = 100
)

// --- Stupeň: jednoduchý ---

// SafeIncrement spustí n goroutin, z nichž každá zvýší společný čítač o jedna,
// a vrátí jeho výslednou hodnotu.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Čtení, přičtení a zápis nejsou jedna operace.
// Najdi datový závod a oprav — testy s -race před opravou padají.
func SafeIncrement(n int) int {
	counter := 0
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			counter++
		}()
	}
	wg.Wait()
	return counter
}

// Registry je registr hodnot bezpečný pro souběžné použití.
// Mapa se vytváří líně při prvním zápisu; zero value musí být použitelná.
type Registry struct {
	items map[string]int
}

// Set uloží hodnotu pod klíč.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Sdílená mapa není chráněná proti souběžnému zápisu.
// Najdi datový závod a oprav — testy s -race před opravou padají.
func (r *Registry) Set(key string, value int) {
	if r.items == nil {
		r.items = make(map[string]int)
	}
	r.items[key] = value
}

// Get vrací hodnotu a true, pokud klíč existuje.
func (r *Registry) Get(key string) (int, bool) {
	v, ok := r.items[key]
	return v, ok
}

// Len vrací počet uložených klíčů.
func (r *Registry) Len() int {
	return len(r.items)
}

// --- Stupeň: střední ---

// StressTest zavolá f ve StressGoroutines goroutinách, každá StressIterations
// krát, a počká na dokončení. Goroutiny vyrazí naráz (společný startovní kanál).
func StressTest(t *testing.T, f func()) {
	// TODO
	t.Fatal("TODO")
}

// Snapshot je konfigurace načtená v jednom okamžiku.
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

// --- Stupeň: obtížný ---

// Config drží aktuální konfiguraci, kterou na pozadí přepisuje reloader.
type Config struct {
	snap Snapshot
}

// Store uloží novou konfiguraci.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Konfigurace se mění po polích, ne atomicky.
// Najdi datový závod a oprav — testy s -race před opravou padají.
func (c *Config) Store(s Snapshot) {
	c.snap.Endpoint = s.Endpoint
	c.snap.Timeout = s.Timeout
	c.snap.Checksum = s.Checksum
}

// Load vrací aktuální konfiguraci.
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
