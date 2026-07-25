// Package exercise obsahuje cvičení lekce 49 — scheduler a mentální model G-M-P.
package exercise

import (
	"time"
)

// BlockingDuration je délka jednoho simulovaného blokujícího volání v Compare.
const BlockingDuration = 50 * time.Millisecond

// RunWithMaxProcs dočasně nastaví GOMAXPROCS na n, spustí f a původní hodnotu
// zase vrátí — i když f panikuje. Pro n <= 0 GOMAXPROCS nemění, pro nil f nedělá nic.
func RunWithMaxProcs(n int, f func()) {
	panic("TODO: úkol A")
}

// ObserveParallelism změří, kolik goroutin se doopravdy sešlo naráz.
//
// Spustí workers goroutin, každá si připočte k čítači souběhu a počká na
// společné uvolnění. Vrací nejvyšší naměřenou hodnotu čítače.
// Po návratu nesmí zůstat běžící goroutina. Pro workers <= 0 vrací 0.
func ObserveParallelism(workers int) int {
	panic("TODO: úkol A")
}

// CPUBound udělá work jednotek čistě výpočetní práce (žádné čekání)
// a vrátí kontrolní součet. Pro stejný vstup vrací vždy stejnou hodnotu.
func CPUBound(work int) uint64 {
	panic("TODO: úkol B")
}

// Blocking simuluje blokující syscall — nedělá nic, jen zabere čas.
// Pro d <= 0 se vrací hned.
func Blocking(d time.Duration) {
	panic("TODO: úkol B")
}

// Compare změří, jak dlouho trvá workers souběžných CPU-bound úloh a jak dlouho
// workers souběžných blokujících volání o délce BlockingDuration.
//
// Blokující varianta škáluje i s jedním P, protože čekající goroutina P uvolní.
// CPU-bound varianta škáluje jen do počtu P. Pro workers <= 0 se použije 1.
func Compare(workers int) (cpu, blocking time.Duration) {
	panic("TODO: úkol B")
}

// StackGrowth zavolá sama sebe do hloubky depth, přičemž každý rámec zabere
// kilobajt lokálního pole. Vrací dosaženou hloubku, pro depth <= 0 nulu.
// Smysl: zásobník goroutiny začíná na dvou kilobajtech a runtime ho podle
// potřeby zvětšuje — což je vidět na tom, že tohle vůbec projde.
func StackGrowth(depth int) int {
	panic("TODO: úkol C")
}

// GoroutineCost spustí n goroutin, počká, až všechny opravdu běží, a vrátí
// počet goroutin před jejich spuštěním a v okamžiku, kdy všechny běží.
// Před návratem je všechny uklidí. Pro n <= 0 vrací dvakrát aktuální počet.
func GoroutineCost(n int) (before, after int) {
	panic("TODO: úkol C")
}

// BytesPerGoroutine odhadne, kolik bajtů zásobníku připadá na jednu goroutinu.
// Měří přes runtime.ReadMemStats a StackInuse, takže jde o hrubý odhad —
// když měření nic nezachytí, vrací 0. Pro n <= 0 vrací 0.
func BytesPerGoroutine(n int) uint64 {
	panic("TODO: úkol C")
}
