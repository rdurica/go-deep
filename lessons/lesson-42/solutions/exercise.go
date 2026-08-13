// Package solutions obsahuje referenční řešení lekce 42.
package solutions

import (
	"context"
	"errors"
	"sync"
	"time"
)

// --- Stupeň: jednoduchý ---

// TrySend se pokusí neblokujícím způsobem poslat v do ch.
func TrySend(ch chan<- int, v int) bool {
	select {
	case ch <- v:
		return true
	default:
		return false
	}
}

// First spustí všechny funkce souběžně, vrátí první úspěšný výsledek a
// ostatní zruší.
func First(ctx context.Context, fns ...func(context.Context) (string, error)) (string, error) {
	if len(fns) == 0 {
		return "", errors.New("first: no functions")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		val string
		err error
	}
	results := make(chan result, len(fns))

	var wg sync.WaitGroup
	wg.Add(len(fns))
	for _, fn := range fns {
		go func() {
			defer wg.Done()
			v, err := fn(ctx)
			results <- result{val: v, err: err}
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	var errs []error
	for res := range results {
		if res.err == nil {
			cancel()
			for range results {
			}
			return res.val, nil
		}
		errs = append(errs, res.err)
	}
	return "", errors.Join(errs...)
}

// --- Stupeň: střední ---

// RecvWithTimeout přečte hodnotu z kanálu, nejdéle však po dobu d.
func RecvWithTimeout(ch <-chan int, d time.Duration) (int, bool) {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case v, ok := <-ch:
		if !ok {
			return 0, false
		}
		return v, true
	case <-timer.C:
		return 0, false
	}
}

// --- Stupeň: obtížný ---

// Heartbeat každých interval zavolá work a pošle tep do vráceného kanálu.
func Heartbeat(ctx context.Context, interval time.Duration, work func()) <-chan time.Time {
	if interval <= 0 {
		interval = time.Millisecond
	}
	out := make(chan time.Time)
	go func() {
		defer close(out)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				if work != nil {
					work()
				}
				select {
				case out <- now:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
