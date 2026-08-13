// Package exercise obsahuje cvičení lekce 49 — scheduler a mentální model G-M-P.
package exercise

import (
	"runtime"
	"time"
)

// BlockingDuration je délka jednoho simulovaného blokujícího volání v Compare.
const BlockingDuration = 50 * time.Millisecond

// --- Stupeň: jednoduchý ---

// RunWithMaxProcs dočasně nastaví GOMAXPROCS na n, spustí f a původní hodnotu
// zase vrátí — i když f panikuje. Pro n <= 0 GOMAXPROCS nemění, pro nil f nedělá nic.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Najdi chybu a oprav ji — testy před opravou padají.
func RunWithMaxProcs(n int, f func()) {
	if f == nil {
		return
	}
	if n > 0 {
		old := runtime.GOMAXPROCS(n)
		f()
		runtime.GOMAXPROCS(old)
		return
	}
	f()
}

// --- Stupeň: střední ---

// ObserveParallelism změří maximum souběžně běžících goroutin.
// Nejdřív je nech „dorazit" do bufferovaného kanálu, pak je pusť zavřením druhého.
// Pro workers <= 0 vrací 0; po návratu žádná živá goroutina.
func ObserveParallelism(workers int) int {
	// TODO
	return 0
}

// CPUBound udělá work jednotek čistě výpočetní práce (žádné čekání)
// a vrátí kontrolní součet. Pro stejný vstup vrací vždy stejnou hodnotu.
func CPUBound(work int) uint64 {
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
	// TODO
	return
}

// --- Stupeň: obtížný ---

// GoroutineCost spustí n goroutin, počká, až všechny opravdu běží, a vrátí
// počet goroutin před jejich spuštěním a v okamžiku, kdy všechny běží.
// Před návratem je všechny uklidí. Pro n <= 0 vrací dvakrát aktuální počet.
func GoroutineCost(n int) (before, after int) {
	// TODO
	return
}
