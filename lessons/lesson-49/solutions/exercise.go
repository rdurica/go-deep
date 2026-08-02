// Package solutions obsahuje referenční řešení lekce 49.
package solutions

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// BlockingDuration je délka jednoho simulovaného blokujícího volání v Compare.
const BlockingDuration = 50 * time.Millisecond

// cpuWork je počet iterací jedné CPU-bound úlohy v Compare.
const cpuWork = 2_000_000

// --- Stupeň: jednoduchý ---
// RunWithMaxProcs dočasně nastaví GOMAXPROCS na n, spustí f a původní hodnotu
// zase vrátí — i když f panikuje. Pro n <= 0 GOMAXPROCS nemění, pro nil f nedělá nic.
func RunWithMaxProcs(n int, f func()) {
	if f == nil {
		return
	}
	if n > 0 {
		// GOMAXPROCS(n) vrací PŘEDCHOZÍ hodnotu — to je celý trik obnovy.
		old := runtime.GOMAXPROCS(n)
		defer runtime.GOMAXPROCS(old)
	}
	f()
}

// ObserveParallelism změří, kolik goroutin se doopravdy sešlo naráz.
//
// Spustí workers goroutin, každá si připočte k čítači souběhu a počká na
// společné uvolnění. Vrací nejvyšší naměřenou hodnotu čítače.
// Po návratu nesmí zůstat běžící goroutina. Pro workers <= 0 vrací 0.
func ObserveParallelism(workers int) int {
	if workers <= 0 {
		return 0
	}

	var (
		inFlight atomic.Int64
		peak     atomic.Int64
		wg       sync.WaitGroup
	)
	arrived := make(chan struct{}, workers)
	release := make(chan struct{})

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			cur := inFlight.Add(1)
			for {
				max := peak.Load()
				if cur <= max || peak.CompareAndSwap(max, cur) {
					break
				}
			}
			arrived <- struct{}{}

			// Čekání na kanálu je "IO": goroutina uvolní P, takže se sem
			// dostanou všechny i při GOMAXPROCS=1.
			<-release
			inFlight.Add(-1)
		}()
	}

	for i := 0; i < workers; i++ {
		<-arrived
	}
	close(release)
	wg.Wait()

	return int(peak.Load())
}

// --- Stupeň: střední ---
// CPUBound udělá work jednotek čistě výpočetní práce (žádné čekání)
// a vrátí kontrolní součet. Pro stejný vstup vrací vždy stejnou hodnotu.
func CPUBound(work int) uint64 {
	// FNV-1a: levné, deterministické a kompilátor to nevyhodí.
	h := uint64(14695981039346656037)
	for i := 0; i < work; i++ {
		h ^= uint64(i)
		h *= 1099511628211
	}
	return h
}

// Blocking simuluje blokující syscall — nedělá nic, jen zabere čas.
// Pro d <= 0 se vrací hned.
func Blocking(d time.Duration) {
	if d <= 0 {
		return
	}
	time.Sleep(d)
}

// Compare změří, jak dlouho trvá workers souběžných CPU-bound úloh a jak dlouho
// workers souběžných blokujících volání o délce BlockingDuration.
//
// Blokující varianta škáluje i s jedním P, protože čekající goroutina P uvolní.
// CPU-bound varianta škáluje jen do počtu P. Pro workers <= 0 se použije 1.
func Compare(workers int) (cpu, blocking time.Duration) {
	if workers <= 0 {
		workers = 1
	}

	var sink atomic.Uint64

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			sink.Add(CPUBound(cpuWork))
		}()
	}
	wg.Wait()
	cpu = time.Since(start)

	start = time.Now()
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			Blocking(BlockingDuration)
		}()
	}
	wg.Wait()
	blocking = time.Since(start)

	return cpu, blocking
}

// --- Stupeň: obtížný ---
// StackGrowth zavolá sama sebe do hloubky depth, přičemž každý rámec zabere
// kilobajt lokálního pole. Vrací dosaženou hloubku, pro depth <= 0 nulu.
// Smysl: zásobník goroutiny začíná na dvou kilobajtech a runtime ho podle
// potřeby zvětšuje — což je vidět na tom, že tohle vůbec projde.
func StackGrowth(depth int) int {
	if depth <= 0 {
		return 0
	}
	// Pole musí být opravdu použité, jinak ho kompilátor vyhodí a rámec
	// zůstane malý.
	var buf [1024]byte
	buf[0] = 1
	buf[len(buf)-1] = buf[0]
	return int(buf[len(buf)-1]) + StackGrowth(depth-1)
}

// GoroutineCost spustí n goroutin, počká, až všechny opravdu běží, a vrátí
// počet goroutin před jejich spuštěním a v okamžiku, kdy všechny běží.
// Před návratem je všechny uklidí. Pro n <= 0 vrací dvakrát aktuální počet.
func GoroutineCost(n int) (before, after int) {
	if n <= 0 {
		now := runtime.NumGoroutine()
		return now, now
	}

	before = runtime.NumGoroutine()

	started := make(chan struct{}, n)
	release := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			started <- struct{}{}
			<-release
		}()
	}
	for i := 0; i < n; i++ {
		<-started
	}
	after = runtime.NumGoroutine()

	close(release)
	wg.Wait()

	return before, after
}

// BytesPerGoroutine odhadne, kolik bajtů zásobníku připadá na jednu goroutinu.
// Měří přes runtime.ReadMemStats a StackInuse, takže jde o hrubý odhad —
// když měření nic nezachytí, vrací 0. Pro n <= 0 vrací 0.
func BytesPerGoroutine(n int) uint64 {
	if n <= 0 {
		return 0
	}

	started := make(chan struct{}, n)
	release := make(chan struct{})
	var wg sync.WaitGroup

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			started <- struct{}{}
			<-release
		}()
	}
	for i := 0; i < n; i++ {
		<-started
	}
	runtime.ReadMemStats(&after)

	close(release)
	wg.Wait()

	if after.StackInuse <= before.StackInuse {
		return 0
	}
	return (after.StackInuse - before.StackInuse) / uint64(n)
}
