// Package solutions obsahuje referenční řešení lekce 45.
package solutions

import (
	"context"
	"errors"
	"strings"
	"sync"
)

// Chyby, které vznikají ve stupních Pipeline.
var (
	// ErrEmpty znamená prázdný vstup po ořezání bílých znaků.
	ErrEmpty = errors.New("empty input")
	// ErrDigits znamená, že vstup obsahuje číslici.
	ErrDigits = errors.New("input contains digits")
)

// Result je výsledek jednoho prvku pipeline.
type Result struct {
	Input string
	Value string
	Err   error
}

// Gen pošle čísla do kanálu a zavře ho.
func Gen(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // stupeň zavírá svůj výstup, nikdy vstup
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Square umocní každou hodnotu ze vstupu.
func Square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			select {
			case out <- v * v:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Stage je obecný stupeň pipeline: na každý prvek použije f.
func Stage[T, U any](ctx context.Context, in <-chan T, f func(T) U) <-chan U {
	out := make(chan U)
	go func() {
		defer close(out)
		for v := range in {
			// Pozor na pořadí: nejdřív se ptáme, jestli ještě někdo
			// poslouchá, teprve pak se pokoušíme poslat.
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

// FanIn sloučí několik kanálů do jednoho.
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
	// Výstup zavírá jediná goroutina, až když už nikdo neposílá.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Take propustí nejvýše n prvků a pak výstup zavře.
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

// Pipeline složí tři stupně a vrátí kanál výsledků.
func Pipeline(ctx context.Context, in <-chan string) <-chan Result {
	const workers = 4

	normalized := normalize(ctx, in)

	// Fan-out: prostřední (nejdražší) stupeň běží ve více kopiích, které
	// čtou ze stejného kanálu.
	enriched := make([]<-chan Result, workers)
	for i := range enriched {
		enriched[i] = enrich(ctx, normalized)
	}

	return format(ctx, FanIn(ctx, enriched...))
}

func normalize(ctx context.Context, in <-chan string) <-chan Result {
	return Stage(ctx, in, func(s string) Result {
		trimmed := strings.TrimSpace(s)
		if trimmed == "" {
			return Result{Input: s, Err: ErrEmpty}
		}
		return Result{Input: s, Value: trimmed}
	})
}

func enrich(ctx context.Context, in <-chan Result) <-chan Result {
	return Stage(ctx, in, func(r Result) Result {
		if r.Err != nil {
			return r // chyba se dál jen veze, nepočítá se s ní znovu
		}
		if strings.ContainsAny(r.Value, "0123456789") {
			return Result{Input: r.Input, Err: ErrDigits}
		}
		r.Value = strings.ToUpper(r.Value)
		return r
	})
}

func format(ctx context.Context, in <-chan Result) <-chan Result {
	return Stage(ctx, in, func(r Result) Result {
		if r.Err != nil {
			return r
		}
		r.Value = "ok:" + r.Value
		return r
	})
}
