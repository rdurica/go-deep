// Package exercise obsahuje cvičení lekce 40.
package exercise

// ParallelSquares vrátí druhé mocniny všech čísel ve stejném pořadí jako vstup.
// Každý prvek zpracuj ve vlastní goroutině a na dokončení počkej WaitGroupou.
func ParallelSquares(nums []int) []int {
	panic("TODO: úkol A")
}

// FanOutSum sečte všechna čísla pomocí nejvýše workers goroutin, které si
// výsledky předávají kanálem. workers < 1 se chová jako 1.
func FanOutSum(nums []int, workers int) int {
	panic("TODO: úkol B")
}

// GoroutineDelta vrátí, o kolik goroutin přibylo (nebo ubylo) po zavolání f.
// Měření musí být stabilizované, ne jeden holý time.Sleep.
func GoroutineDelta(f func()) int {
	panic("TODO: úkol B")
}

// LeakyGenerator je záměrně vadná funkce: spustí goroutinu, která uvízne
// navždy. Slouží jako referenční leak pro test GoroutineDelta.
func LeakyGenerator() {
	panic("TODO: úkol C")
}

// SafeGenerator posílá rostoucí čísla od nuly, dokud volající nezavře done.
// Po zavření done goroutina skončí a výstupní kanál zavře.
func SafeGenerator(done <-chan struct{}) <-chan int {
	panic("TODO: úkol C")
}
