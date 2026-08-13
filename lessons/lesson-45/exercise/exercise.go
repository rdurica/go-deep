// Package exercise obsahuje cvičení lekce 45.
package exercise

import (
	"context"
	"errors"
	"sync"
)

// Chyby, které vznikají ve stupních Pipeline.
var (
	// ErrEmpty znamená prázdný vstup po ořezání bílých znaků.
	ErrEmpty = errors.New("empty input")
	// ErrDigits znamená, že vstup obsahuje číslici.
	ErrDigits = errors.New("input contains digits")
)

// Result je výsledek jednoho prvku pipeline. Chyba se nese s ním, ne mimo něj.
type Result struct {
	Input string
	Value string
	Err   error
}

// --- Stupeň: jednoduchý ---

// Gen pošle nums do kanálu a zavře ho. Při zrušení kontextu skonči dřív (bez leaku).
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Najdi chybu a oprav ji — testy před opravou padají.
func Gen(ctx context.Context, nums ...int) <-chan int {
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

// Square umocní každou hodnotu ze vstupu. Zavírej jen svůj výstup; vstup nech otevřený.
// Zápis i čtení musí mít cestu ven přes ctx.Done() — po zrušení uprostřed nesmí zůstat goroutina.
func Square(ctx context.Context, in <-chan int) <-chan int {
	// TODO
	return closed[int]()
}

// Stage je obecný stupeň pipeline: na každý prvek aplikuje f. Zavírá jen výstup.
// Vstup nikdy nezavírej. select s ctx.Done() u zápisu.
func Stage[T, U any](ctx context.Context, in <-chan T, f func(T) U) <-chan U {
	out := make(chan U)
	go func() {
		defer close(out)
		for v := range in {
			select {
			case <-ctx.Done():
				return
			default:
			}
			select {
			case out <- f(v):
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// FanIn sloučí kanály do jednoho. Výstup zavři právě jednou po vyčerpání všech vstupů
// (nebo zrušení ctx). Bez vstupů → rovnou zavřený kanál.
func FanIn[T any](ctx context.Context, chs ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	wg.Add(len(chs))
	for _, ch := range chs {
		go func() {
			defer wg.Done()
			for v := range ch {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Take propustí nejvýše n prvků a výstup zavře. Vstup vyčerpán dřív → méně prvků.
// n <= 0 → rovnou zavřený kanál. Po Take + cancel nesmí zbýt goroutina.
func Take[T any](ctx context.Context, in <-chan T, n int) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for i := 0; i < n; i++ {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

// --- Stupeň: obtížný ---

// Pipeline: TrimSpace (prázdný → ErrEmpty), fan-out 4× obohacení (číslice → ErrDigits,
// jinak upper), formát prefixem "ok:". Result.Input = původní vstup před ořezáním.
// Chybný Result další stupně jen propouští. Pořadí není zaručené. Po cancel žádná goroutina.
func Pipeline(ctx context.Context, in <-chan string) <-chan Result {
	// TODO
	return closed[Result]()
}

// closed je fail-fast stub: nil kanál by v testech visel navždy.
func closed[T any]() <-chan T {
	ch := make(chan T)
	close(ch)
	return ch
}
