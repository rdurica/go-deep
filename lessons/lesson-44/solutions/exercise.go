// Package solutions obsahuje referenční řešení lekce 44.
package solutions

import (
	"sync"
	"sync/atomic"
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
func SafeIncrement(n int) int {
	// counter++ je čtení, přičtení a zápis. atomic.Int64 to udělá jako
	// jednu nedělitelnou operaci.
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
	once  sync.Once // líná inicializace právě jednou
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

// --- Stupeň: střední ---
// Len vrací počet uložených klíčů.
func (r *Registry) Len() int {
	r.init()
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items)
}

// ParallelAppend vrátí druhé mocniny všech čísel (v libovolném pořadí),
// spočítané souběžně.
func ParallelAppend(nums []int) []int {
	out := make([]int, 0, len(nums))
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(nums))
	for _, n := range nums {
		go func() {
			defer wg.Done()
			v := n * n // výpočet mimo zámek, pod zámkem jen append
			mu.Lock()
			out = append(out, v)
			mu.Unlock()
		}()
	}
	wg.Wait()
	return out
}

// StressTest zavolá f ve StressGoroutines goroutinách, každá StressIterations
// krát, a počká na jejich dokončení.
func StressTest(t *testing.T, f func()) {
	t.Helper()

	// Startovní výstřel: goroutiny se rozeběhnou naráz, takže mají největší
	// šanci se do sebe pustit.
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

// --- Stupeň: obtížný ---
// Consistent hlásí, jestli snapshot dává smysl jako celek.
func (s Snapshot) Consistent() bool {
	return s.Checksum == len(s.Endpoint)+s.Timeout
}

// Config drží aktuální konfiguraci, kterou na pozadí přepisuje reloader,
// zatímco ji ostatní goroutiny čtou.
type Config struct {
	// Celá konfigurace se vyměňuje jako jedna hodnota, takže čtenář nikdy
	// neuvidí půlku staré a půlku nové. Alternativa je RWMutex.
	v atomic.Value
}

// Store uloží novou konfiguraci.
func (c *Config) Store(s Snapshot) {
	c.v.Store(s)
}

// Load vrací aktuální konfiguraci.
func (c *Config) Load() Snapshot {
	s, _ := c.v.Load().(Snapshot) // nenaplněná Value vrací nil
	return s
}

// StartReloader spustí goroutinu, která dokola ukládá snapshoty ze seznamu,
// dokud volající nezavře stop.
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
