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

// WithContext vrátí skupinu a odvozený kontext, který se zruší, jakmile
// první úloha vrátí chybu — nebo až Wait skončí, podle toho, co nastane dřív.
func WithContext(ctx context.Context) (*Group, context.Context) {
	// TODO: úkol B
	return nil, *new(context.Context)
}

// SetLimit omezí počet souběžně běžících úloh. Pro n <= 0 se limit zruší.
// Volání po prvním Go panikuje — limit se za běhu měnit nedá.
func (g *Group) SetLimit(n int) {
	// TODO: úkol B
}

// Go spustí f v nové goroutině. Při nastaveném limitu blokuje, dokud se
// neuvolní místo. Chybu si skupina zapamatuje jen z prvního neúspěchu.
// Nil funkce se tiše přeskočí.
func (g *Group) Go(f func() error) {
	// TODO: úkol A
}

// Wait počká na všechny úlohy a vrátí chybu té, která selhala jako první.
// Pokud skupina vznikla přes WithContext, Wait odvozený kontext zruší.
func (g *Group) Wait() error {
	// TODO: úkol A
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

// RunAll spustí všechny úlohy souběžně a počká na jejich doběhnutí.
//
// Běžné úlohy dostanou kontext odvozený od ctx, který se zruší při první chybě
// nebo při zrušení rodiče. Úlohy s Detached == true dostanou kontext bez
// zrušení (context.WithoutCancel), takže doběhnou i po zrušení rodiče.
// Chyba úlohy se obalí jménem úlohy; vrací se první z nich.
func RunAll(ctx context.Context, tasks []Task) error {
	// TODO: úkol C
	return nil
}

// Cause rozbalí řetězec chyb až na tu nejhlubší a vrátí ji.
// Pro nil vrací nil, pro chybu bez Unwrap vrací ji samotnou.
func Cause(err error) error {
	// TODO: úkol C
	return nil
}
