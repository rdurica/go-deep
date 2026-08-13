// Package solutions obsahuje referenční řešení lekce 41.
package solutions

import "sync"

// --- Stupeň: jednoduchý ---

// ForgetClose pošle všechna čísla do kanálu v goroutině a kanál zavře.
func ForgetClose(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// --- Stupeň: střední ---

// Generate pošle všechna čísla do vráceného kanálu a kanál sám zavře.
func Generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// Collect přečte kanál až do zavření a vrátí hodnoty v pořadí, v jakém dorazily.
func Collect(ch <-chan int) []int {
	out := []int{}
	for v := range ch {
		out = append(out, v)
	}
	return out
}

// --- Stupeň: obtížný ---

// Merge sloučí několik vstupních kanálů do jednoho (fan-in).
func Merge(chs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(chs))
	for _, ch := range chs {
		go func() {
			defer wg.Done()
			for v := range ch {
				out <- v
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
