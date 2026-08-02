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
	default: // větev default dělá ze selectu neblokující operaci
		return false
	}
}

// --- Stupeň: střední ---
// RecvWithTimeout přečte hodnotu z kanálu, nejdéle však po dobu d.
func RecvWithTimeout(ch <-chan int, d time.Duration) (int, bool) {
	// time.After by tady taky fungoval, ale timer by žil až do vypršení.
	// NewTimer + Stop uvolní zdroje hned, jakmile hodnota dorazí dřív.
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
	// Buffer na všechny výsledky: ani opuštěná goroutina neuvízne na zápisu.
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
			cancel() // ostatním řekneme, že už je nepotřebujeme
			// Kdo goroutinu spustil, počká na její konec. Kanál se zavře
			// až po wg.Wait(), takže dočtení znamená "všichni skončili".
			for range results {
			}
			return res.val, nil
		}
		errs = append(errs, res.err)
	}
	return "", errors.Join(errs...)
}

// --- Stupeň: obtížný ---
// Debounce propustí hodnotu, až když po dobu d nepřijde nic nového.
func Debounce(in <-chan string, d time.Duration) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)

		timer := time.NewTimer(d)
		// Timer nechceme rozběhnutý, dokud nic nepřišlo.
		if !timer.Stop() {
			<-timer.C
		}
		defer timer.Stop()

		var pending string
		var have bool
		for {
			select {
			case v, ok := <-in:
				if !ok {
					if have {
						out <- pending // poslední slovo ještě doručíme
					}
					return
				}
				pending, have = v, true
				// Reset se smí volat až na zastaveném a vyprázdněném timeru.
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(d)
			case <-timer.C:
				if have {
					out <- pending
					have = false
				}
			}
		}
	}()
	return out
}

// Heartbeat každých interval zavolá work a pošle tep do vráceného kanálu.
func Heartbeat(ctx context.Context, interval time.Duration, work func()) <-chan time.Time {
	if interval <= 0 {
		interval = time.Millisecond
	}
	out := make(chan time.Time)
	go func() {
		defer close(out)

		// time.Tick by ticker nikdy neuvolnil — v dlouho žijícím procesu
		// je to leak. NewTicker se dá zastavit.
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
				// Ani tep neposíláme naslepo: kdyby konzument přestal
				// číst, uvízli bychom tu navždy.
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
