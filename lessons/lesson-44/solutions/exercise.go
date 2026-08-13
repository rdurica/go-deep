// Package solutions obsahuje referenční řešení lekce 44.
package solutions

import (
	"sync"
	"sync/atomic"
	"testing"
)

// Zátěž, kterou pouští StressTest.
const (
	StressGoroutines = 50
	StressIterations = 100
)

// --- Stupeň: jednoduchý ---

// SafeIncrement spustí n goroutin, z nichž každá zvýší společný čítač o jedna.
func SafeIncrement(n int) int {
	var counter atomic.Int64
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			counter.Add(1)
		}()
	}
	wg.Wait()
	return int(counter.Load())
}

// Registry je registr hodnot bezpečný pro souběžné použití.
type Registry struct {
	once  sync.Once
	mu    sync.RWMutex
	items map[string]int
}

func (r *Registry) init() {
	r.once.Do(func() {
		r.items = make(map[string]int)
	})
}

// Set uloží hodnotu pod klíč.
func (r *Registry) Set(key string, value int) {
	r.init()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[key] = value
}

// Get vrací hodnotu a true, pokud klíč existuje.
func (r *Registry) Get(key string) (int, bool) {
	r.init()
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.items[key]
	return v, ok
}

// Len vrací počet uložených klíčů.
func (r *Registry) Len() int {
	r.init()
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items)
}

// --- Stupeň: střední ---

// StressTest zavolá f ve StressGoroutines goroutinách, každá StressIterations krát.
func StressTest(t *testing.T, f func()) {
	t.Helper()
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(StressGoroutines)
	for i := 0; i < StressGoroutines; i++ {
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < StressIterations; j++ {
				f()
			}
		}()
	}
	close(start)
	wg.Wait()
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

// Config drží aktuální konfiguraci.
type Config struct {
	v atomic.Value
}

// Store uloží novou konfiguraci.
func (c *Config) Store(s Snapshot) {
	c.v.Store(s)
}

// Load vrací aktuální konfiguraci.
func (c *Config) Load() Snapshot {
	s, _ := c.v.Load().(Snapshot)
	return s
}

// StartReloader spustí goroutinu, která dokola ukládá snapshoty.
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
