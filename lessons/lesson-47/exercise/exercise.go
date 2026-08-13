// Package exercise obsahuje cvičení lekce 47 — vlastní errgroup a rušení přes context.
package exercise

import (
	"context"
	"errors"
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

// --- Stupeň: jednoduchý ---

// WithContext vrátí skupinu a odvozený kontext zrušený při první chybě.
// Použij context.WithCancelCause; Wait musí kontext zrušit i v úspěšné větvi.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Najdi chybu a oprav ji — testy před opravou padají.
func WithContext(ctx context.Context) (*Group, context.Context) {
	_, cancel := context.WithCancelCause(ctx)
	_ = cancel
	return &Group{}, ctx
}

// SetLimit omezí počet souběžně běžících úloh; pro n <= 0 limit ruší.
// Volání po prvním Go panikuje. Vstupenku ber v Go, ne uvnitř goroutiny.
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

// --- Stupeň: střední ---

// Go spustí f v nové goroutině; nulová Group je použitelná.
// Go(nil) se tiše přeskočí. Při limitu blokuje volajícího, dokud se neuvolní místo.
// První chyba se zapamatuje přes sync.Once; Wait ji vrátí po doběhnutí všech úloh.
func (g *Group) Go(f func() error) {
	// TODO
}

// Wait počká na všechny úlohy a vrátí chybu té, která selhala jako první.
// Pokud skupina vznikla přes WithContext, Wait odvozený kontext zruší.
func (g *Group) Wait() error {
	// TODO
	return nil
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

// --- Stupeň: obtížný ---

// RunAll spustí všechny úlohy souběžně a počká i na odpojené (Detached).
// Běžné dostanou kontext skupiny, odpojené context.WithoutCancel(ctx).
// Chybu úlohy obal: fmt.Errorf("task %q: %w", …). Run == nil → ErrNilTask;
// prázdný seznam vrací nil.
func RunAll(ctx context.Context, tasks []Task) error {
	// TODO
	return nil
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
