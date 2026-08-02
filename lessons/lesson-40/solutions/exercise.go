// Package solutions obsahuje referenční řešení lekce 40.
package solutions

import (
	"runtime"
	"sync"
	"time"
)

// --- Stupeň: jednoduchý ---
// ParallelSquares vrátí druhé mocniny všech čísel ve stejném pořadí jako vstup.
func ParallelSquares(nums []int) []int {
	out := make([]int, len(nums))
	var wg sync.WaitGroup
	for i, n := range nums {
		wg.Add(1) // Add vždy PŘED go, nikdy uvnitř goroutiny
		go func() {
			defer wg.Done()
			// Každá goroutina píše na vlastní index. Různé prvky slice jsou
			// různá paměťová místa a slice se nerealokuje, takže tu není
			// co zamykat.
			out[i] = n * n
		}()
	}
	wg.Wait()
	return out
}

// --- Stupeň: střední ---
// FanOutSum sečte všechna čísla pomocí nejvýše workers goroutin, které si
// dílčí součty předávají kanálem.
func FanOutSum(nums []int, workers int) int {
	if len(nums) == 0 {
		return 0
	}
	if workers < 1 {
		workers = 1
	}
	if workers > len(nums) {
		workers = len(nums)
	}

	partials := make(chan int, workers)
	var wg sync.WaitGroup
	chunk := (len(nums) + workers - 1) / workers
	for start := 0; start < len(nums); start += chunk {
		end := start + chunk
		if end > len(nums) {
			end = len(nums)
		}
		part := nums[start:end]
		wg.Add(1)
		go func() {
			defer wg.Done()
			sum := 0
			for _, n := range part {
				sum += n
			}
			partials <- sum
		}()
	}

	// Kdo kanál otevřel, ten ho zavírá — a to až po doběhnutí všech
	// odesílatelů. Zavření musí být v samostatné goroutině, jinak by
	// wg.Wait() blokoval dřív, než začneme číst.
	go func() {
		wg.Wait()
		close(partials)
	}()

	total := 0
	for sum := range partials {
		total += sum
	}
	return total
}

// GoroutineDelta vrátí, o kolik goroutin přibylo (nebo ubylo) po zavolání f.
func GoroutineDelta(f func()) int {
	before := stableGoroutines()
	f()
	after := stableGoroutines()
	return after - before
}

// stableGoroutines počká, až se počet goroutin ustálí, a vrátí ho. Opakované
// čtení je spolehlivější než jeden pevný sleep: doběhnutí goroutiny není
// okamžité a runtime má i vlastní pomocné goroutiny.
func stableGoroutines() int {
	const (
		needStable = 3
		maxRounds  = 300
		step       = 5 * time.Millisecond
	)
	runtime.Gosched()
	prev := runtime.NumGoroutine()
	stable := 0
	for i := 0; i < maxRounds; i++ {
		time.Sleep(step)
		cur := runtime.NumGoroutine()
		if cur == prev {
			stable++
			if stable >= needStable {
				return cur
			}
			continue
		}
		prev = cur
		stable = 0
	}
	return runtime.NumGoroutine()
}

// --- Stupeň: obtížný ---
// LeakyGenerator je záměrně vadná funkce: spustí goroutinu, která uvízne
// navždy na zápisu do kanálu, ze kterého nikdo nečte.
func LeakyGenerator() {
	ch := make(chan int) // nebufferovaný, bez čtenáře
	go func() {
		ch <- 42 // tenhle zápis se nikdy nedokončí
	}()
	// Vracíme se okamžitě. Goroutina výše zůstane viset do konce procesu.
}

// SafeGenerator posílá rostoucí čísla od nuly, dokud volající nezavře done.
func SafeGenerator(done <-chan struct{}) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // odesílatel zavírá výstup
		for i := 0; ; i++ {
			select {
			case <-done:
				return
			case out <- i:
			}
		}
	}()
	return out
}
