// Package solutions obsahuje referenční řešení lekce 47.
package solutions

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// ErrNilTask signalizuje úlohu bez funkce Run.
var ErrNilTask = errors.New("task: no Run function")

// Group spouští skupinu goroutin a čeká, až všechny skončí.
// Nulová hodnota je použitelná: skupina bez limitu a bez kontextu.
type Group struct {
	wg      sync.WaitGroup
	sem     chan struct{}
	errOnce sync.Once
	err     error
	cancel  context.CancelCauseFunc
	started atomic.Bool
}

// WithContext vrátí skupinu a odvozený kontext, který se zruší, jakmile
// první úloha vrátí chybu — nebo až Wait skončí, podle toho, co nastane dřív.
func WithContext(ctx context.Context) (*Group, context.Context) {
	// WithCancelCause místo WithCancel: příjemce se pak přes context.Cause
	// dozví skutečný důvod, ne jen "context canceled".
	ctx, cancel := context.WithCancelCause(ctx)
	return &Group{cancel: cancel}, ctx
}

// SetLimit omezí počet souběžně běžících úloh. Pro n <= 0 se limit zruší.
// Volání po prvním Go panikuje — limit se za běhu měnit nedá.
func (g *Group) SetLimit(n int) {
	if g.started.Load() {
		panic("Group.SetLimit: limit nelze měnit po prvním Go")
	}
	if n <= 0 {
		g.sem = nil
		return
	}
	g.sem = make(chan struct{}, n)
}

// Go spustí f v nové goroutině. Při nastaveném limitu blokuje, dokud se
// neuvolní místo. Chybu si skupina zapamatuje jen z prvního neúspěchu.
// Nil funkce se tiše přeskočí.
func (g *Group) Go(f func() error) {
	g.started.Store(true)

	// Vstupenku bereme JEŠTĚ V VOLAJÍCÍM, ne uvnitř goroutiny. Jinak bychom
	// nejdřív spustili všechny goroutiny a limit by neomezoval nic.
	if g.sem != nil {
		g.sem <- struct{}{}
	}
	g.wg.Add(1)

	go func() {
		defer func() {
			if g.sem != nil {
				<-g.sem
			}
			g.wg.Done()
		}()
		if f == nil {
			return
		}
		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel(err) // první chyba ruší zbytek skupiny
				}
			})
		}
	}()
}

// Wait počká na všechny úlohy a vrátí chybu té, která selhala jako první.
// Pokud skupina vznikla přes WithContext, Wait odvozený kontext zruší.
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel(g.err)
	}
	return g.err
}

// Task je jedna pojmenovaná úloha pro RunAll.
type Task struct {
	// Name se objeví v textu chyby.
	Name string
	// Detached říká, že úloha má doběhnout i po zrušení rodičovského kontextu.
	Detached bool
	// Run je vlastní práce. Nil znamená chybu ErrNilTask.
	Run func(context.Context) error
}

// RunAll spustí všechny úlohy souběžně a počká na jejich doběhnutí.
//
// Běžné úlohy dostanou kontext odvozený od ctx, který se zruší při první chybě
// nebo při zrušení rodiče. Úlohy s Detached == true dostanou kontext bez
// zrušení (context.WithoutCancel), takže doběhnou i po zrušení rodiče.
// Chyba úlohy se obalí jménem úlohy; vrací se první z nich.
func RunAll(ctx context.Context, tasks []Task) error {
	if len(tasks) == 0 {
		return nil
	}

	g, groupCtx := WithContext(ctx)

	// WithoutCancel zachová hodnoty (trace ID, logger), ale zahodí zrušení
	// i deadline. Přesně to chceš u práce, která má přežít odpověď.
	detachedCtx := context.WithoutCancel(ctx)

	for _, t := range tasks {
		taskCtx := groupCtx
		if t.Detached {
			taskCtx = detachedCtx
		}
		g.Go(func() error {
			if t.Run == nil {
				return fmt.Errorf("task %q: %w", t.Name, ErrNilTask)
			}
			if err := t.Run(taskCtx); err != nil {
				return fmt.Errorf("task %q: %w", t.Name, err)
			}
			return nil
		})
	}

	return g.Wait()
}

// Cause rozbalí řetězec chyb až na tu nejhlubší a vrátí ji.
// Pro nil vrací nil, pro chybu bez Unwrap vrací ji samotnou.
func Cause(err error) error {
	for {
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return err
		}
		inner := u.Unwrap()
		if inner == nil {
			return err
		}
		err = inner
	}
}
