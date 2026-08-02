// Package exercise obsahuje cvičení lekce 49 — scheduler a mentální model G-M-P.
package exercise

import (
	"time"
)

// BlockingDuration je délka jednoho simulovaného blokujícího volání v Compare.
const BlockingDuration = 50 * time.Millisecond

// --- Stupeň: jednoduchý ---
// RunWithMaxProcs dočasně nastaví GOMAXPROCS na n, spustí f a původní hodnotu
// zase vrátí — i když f panikuje. Pro n <= 0 GOMAXPROCS nemění, pro nil f nedělá nic.
func RunWithMaxProcs(n int, f func()) {
	// TODO
}

// ObserveParallelism změří maximum souběžně běžících goroutin.
// Nejdřív je nech „dorazit" do bufferovaného kanálu, pak je pusť zavřením druhého.
// Pro workers <= 0 vrací 0; po návratu žádná živá goroutina.
func ObserveParallelism(workers int) int {
	// TODO
	return 0
}

// --- Stupeň: střední ---
// CPUBound udělá work jednotek čistě výpočetní práce (žádné čekání)
// a vrátí kontrolní součet. Pro stejný vstup vrací vždy stejnou hodnotu.
func CPUBound(work int) uint64 {
	// TODO
	return 0
}

// Blocking simuluje blokující syscall — nedělá nic, jen zabere čas.
// Pro d <= 0 se vrací hned.
func Blocking(d time.Duration) {
	// TODO
}

// Compare změří, jak dlouho trvá workers souběžných CPU-bound úloh a jak dlouho
// workers souběžných blokujících volání o délce BlockingDuration.
//
// Blokující varianta škáluje i s jedním P, protože čekající goroutina P uvolní.
// CPU-bound varianta škáluje jen do počtu P. Pro workers <= 0 se použije 1.
func Compare(workers int) (cpu, blocking time.Duration) {
	// TODO
	return
}

// --- Stupeň: obtížný ---
// StackGrowth rekurzivně zabere [1024]byte v každém rámci (pole musíš použít).
// Vrací dosaženou hloubku; pro depth <= 0 nulu.
func StackGrowth(depth int) int {
	// TODO
	return 0
}

// GoroutineCost spustí n goroutin, počká, až všechny opravdu běží, a vrátí
// počet goroutin před jejich spuštěním a v okamžiku, kdy všechny běží.
// Před návratem je všechny uklidí. Pro n <= 0 vrací dvakrát aktuální počet.
func GoroutineCost(n int) (before, after int) {
	// TODO
	return
}

// BytesPerGoroutine odhadne, kolik bajtů zásobníku připadá na jednu goroutinu.
// Měří přes runtime.ReadMemStats a StackInuse, takže jde o hrubý odhad —
// když měření nic nezachytí, vrací 0. Pro n <= 0 vrací 0.
func BytesPerGoroutine(n int) uint64 {
	// TODO
	return 0
}
