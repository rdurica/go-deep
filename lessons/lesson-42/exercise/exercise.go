// Package exercise obsahuje cvičení lekce 42.
package exercise

import (
	"context"
	"errors"
	"sync"
	"time"
)

// --- Stupeň: jednoduchý ---

// TrySend neblokujícím select+default pošle v do ch. Plný, nil nebo bez čtenáře → false.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Najdi chybu a oprav ji — testy před opravou padají.
func TrySend(ch chan<- int, v int) bool {
	ch <- v
	return true
}

// First spustí všechny fns souběžně, vrátí první úspěch, poražené zruší.
// Počkej na doběhnutí všech goroutin. Všechny selžou → errors.Join.
// Bez funkcí vrať chybu.
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

// RecvWithTimeout přečte z ch nejdéle po d. Timeout i zavřený kanál → (0, false).
// Použij NewTimer a defer Stop(), ne time.After.
func RecvWithTimeout(ch <-chan int, d time.Duration) (int, bool) {
	// TODO
	return 0, false
}

// --- Stupeň: obtížný ---

// Heartbeat každých interval zavolá work (může být nil) a pošle tep.
// Po zrušení ctx skončí a kanál zavře; musí skončit i bez čtenáře (select+ctx).
// interval <= 0 jako 1 ms; už zrušený ctx → okamžitý konec bez tepu.
func Heartbeat(ctx context.Context, interval time.Duration, work func()) <-chan time.Time {
	// TODO
	ch := make(chan time.Time)
	close(ch)
	return ch
}
