// Package exercise obsahuje cvičení lekce 40.
package exercise

// --- Stupeň: jednoduchý ---
// ParallelSquares vrátí druhé mocniny všech čísel ve stejném pořadí jako vstup.
// Každý prvek ve vlastní goroutině, výsledek na index do předalokovaného slice,
// WaitGroup. nil/prázdný vstup → prázdný výsledek bez paniky.
func ParallelSquares(nums []int) []int {
	// TODO
	return nil
}

// --- Stupeň: střední ---
// FanOutSum sečte čísla nejvýše ve workers goroutinách přes kanál výsledků.
// workers < 1 jako 1; workers > len(nums) nespouští víc goroutin než prvků.
// Prázdný vstup → 0. Pozor na zavření kanálu výsledků.
func FanOutSum(nums []int, workers int) int {
	// TODO
	return 0
}

// GoroutineDelta vrátí přírůstek goroutin po f se stabilizací NumGoroutine
// (ne jeden Sleep). Čistá f → 0; leak tří goroutin → alespoň 3.
func GoroutineDelta(f func()) int {
	// TODO
	return 0
}

// --- Stupeň: obtížný ---
// LeakyGenerator záměrně leakuje: goroutina zablokovaná na zápis do
// nebufferovaného kanálu bez čtenáře. Reference pro GoroutineDelta test.
func LeakyGenerator() {
	// TODO
}

// SafeGenerator posílá 0,1,2,… dokud volající nezavře done.
// Po close(done) skončí i bez čtenáře výstupu; výstup zavírá generátor.
// GoroutineDelta po celém scénáři musí být 0.
func SafeGenerator(done <-chan struct{}) <-chan int {
	// TODO
	ch := make(chan int)
	close(ch)
	return ch
}
