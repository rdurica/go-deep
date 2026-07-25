// Package pool obsahuje generický worker pool s omezenou souběžností,
// backpressure na vstupní frontě, gracefulním ukončením a metrikami.
//
// Životní cyklus je vždycky stejný:
//
//	p, err := pool.New(ctx, cfg)     // spustí workery
//	go func() {                      // producent
//	    for _, v := range inputs {
//	        _ = p.Submit(ctx, v)     // blokuje, když je fronta plná
//	    }
//	    p.Close()                    // "další úlohy nepřijdou"
//	}()
//	for res := range p.Results() { … } // konzument dočte kanál do konce
//	stats := p.Stats()
//
// Kdo pool vlastní, ten ho zavírá. Kanál výsledků se zavře sám, až doběhnou
// všechny úlohy přijaté před Close.
package pool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Chyby konfigurace a životního cyklu poolu.
var (
	// ErrInvalidWorkers znamená nekladný počet workerů.
	ErrInvalidWorkers = errors.New("pool: workers must be positive")
	// ErrInvalidQueueSize znamená zápornou velikost vstupní fronty.
	ErrInvalidQueueSize = errors.New("pool: queue size must not be negative")
	// ErrNilHandler znamená chybějící funkci pro zpracování úlohy.
	ErrNilHandler = errors.New("pool: handler must not be nil")
	// ErrClosed vrací Submit po zavření poolu.
	ErrClosed = errors.New("pool: closed")
)

// Config je konfigurace poolu.
type Config[T, U any] struct {
	// Workers je počet goroutin, tedy strop souběžnosti. Musí být kladný.
	Workers int
	// QueueSize je kapacita vstupní fronty. Nula znamená nejtvrdší
	// backpressure: Submit čeká, dokud si úlohu nevezme volný worker.
	QueueSize int
	// Handler je vlastní práce nad jednou úlohou. Musí být zrušitelný
	// přes předaný kontext.
	Handler func(context.Context, T) (U, error)
}

// Result je výsledek jedné úlohy. Index umožňuje obnovit původní pořadí,
// Input spárování bez ohledu na pořadí dokončení.
type Result[T, U any] struct {
	Index int
	Input T
	Value U
	Err   error
}

// Stats je souhrn běhu poolu.
type Stats struct {
	// Submitted je počet úloh, které pool přijal.
	Submitted int
	// Processed je počet úloh dokončených bez chyby.
	Processed int
	// Failed je počet úloh, které skončily chybou.
	Failed int
	// MaxInFlight je nejvyšší pozorovaná souběžnost.
	MaxInFlight int
	// Elapsed je doba od New do zavření kanálu výsledků. Během běhu
	// vrací dobu od spuštění.
	Elapsed time.Duration
}

// Total vrací počet dokončených úloh, tedy Processed + Failed.
func (s Stats) Total() int { return s.Processed + s.Failed }

type job[T any] struct {
	index int
	value T
}

// Pool zpracovává úlohy typu T na výsledky typu U pevným počtem workerů.
type Pool[T, U any] struct {
	handler   func(context.Context, T) (U, error)
	jobs      chan job[T]
	results   chan Result[T, U]
	startedAt time.Time

	// mu chrání closed a serializuje odeslání do jobs, aby Submit nikdy
	// nezapsal do kanálu zavřeného mezitím metodou Close.
	mu     sync.Mutex
	closed bool

	submitted   atomic.Int64
	processed   atomic.Int64
	failed      atomic.Int64
	inFlight    atomic.Int64
	maxInFlight atomic.Int64
	elapsedNs   atomic.Int64
}

// New zvaliduje konfiguraci a spustí workery. Workery dostávají ctx, takže
// jeho zrušení pool zastaví bez ohledu na stav fronty.
func New[T, U any](ctx context.Context, cfg Config[T, U]) (*Pool[T, U], error) {
	if cfg.Workers <= 0 {
		return nil, ErrInvalidWorkers
	}
	if cfg.QueueSize < 0 {
		return nil, ErrInvalidQueueSize
	}
	if cfg.Handler == nil {
		return nil, ErrNilHandler
	}

	p := &Pool[T, U]{
		handler:   cfg.Handler,
		jobs:      make(chan job[T], cfg.QueueSize),
		results:   make(chan Result[T, U], cfg.Workers),
		startedAt: time.Now(),
	}

	var wg sync.WaitGroup
	wg.Add(cfg.Workers)
	for i := 0; i < cfg.Workers; i++ {
		go func() {
			defer wg.Done()
			p.work(ctx)
		}()
	}

	// Kanál výsledků zavírá vlastník, a to až když do něj přestali psát
	// všichni workeři. Tady se zároveň zmrazí naměřená doba běhu.
	go func() {
		wg.Wait()
		p.elapsedNs.Store(int64(time.Since(p.startedAt)))
		close(p.results)
	}()

	return p, nil
}

// Submit zařadí úlohu do fronty. Když je fronta plná, blokuje — to je ten
// backpressure. Vrací ErrClosed po Close a ctx.Err() při zrušení kontextu.
//
// Předávej stejný kontext jako do New; jinak se může stát, že Submit bude
// čekat na frontu, kterou už zastavení workerů nikdy nevyprázdní.
func (p *Pool[T, U]) Submit(ctx context.Context, v T) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return ErrClosed
	}

	j := job[T]{index: int(p.submitted.Load()), value: v}
	select {
	case p.jobs <- j:
		p.submitted.Add(1)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Results vrací kanál výsledků. Volající ho musí dočíst do konce, jinak
// zůstanou workeři viset na zápisu. Zavře se, až doběhnou všechny úlohy
// přijaté před Close (nebo hned po zrušení kontextu).
func (p *Pool[T, U]) Results() <-chan Result[T, U] {
	return p.results
}

// Close oznámí poolu, že další úlohy nepřijdou. Rozpracované a zařazené úlohy
// se dokončí (graceful drain). Opakované volání je bezpečné.
func (p *Pool[T, U]) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return
	}
	p.closed = true
	close(p.jobs)
}

// Stats vrací aktuální metriky. Kompletní jsou až po dočtení Results().
func (p *Pool[T, U]) Stats() Stats {
	elapsed := time.Duration(p.elapsedNs.Load())
	if elapsed == 0 {
		elapsed = time.Since(p.startedAt)
	}
	return Stats{
		Submitted:   int(p.submitted.Load()),
		Processed:   int(p.processed.Load()),
		Failed:      int(p.failed.Load()),
		MaxInFlight: int(p.maxInFlight.Load()),
		Elapsed:     elapsed,
	}
}

func (p *Pool[T, U]) work(ctx context.Context) {
	for {
		// Zrušení kontextu má přednost před další úlohou z fronty.
		select {
		case <-ctx.Done():
			return
		default:
		}

		select {
		case <-ctx.Done():
			return
		case j, ok := <-p.jobs:
			if !ok {
				return // Close: fronta je vyčerpaná, končíme
			}
			res := p.handle(ctx, j)
			select {
			case p.results <- res:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (p *Pool[T, U]) handle(ctx context.Context, j job[T]) Result[T, U] {
	cur := p.inFlight.Add(1)
	for {
		old := p.maxInFlight.Load()
		if cur <= old || p.maxInFlight.CompareAndSwap(old, cur) {
			break
		}
	}
	defer p.inFlight.Add(-1)

	v, err := p.handler(ctx, j.value)
	if err != nil {
		p.failed.Add(1)
	} else {
		p.processed.Add(1)
	}
	return Result[T, U]{Index: j.index, Input: j.value, Value: v, Err: err}
}

// Collect zpracuje celou dávku a vrátí výsledky ve stejném pořadí jako vstup.
//
// Chyby jednotlivých úloh zůstávají ve výsledcích — návratová chyba znamená
// selhání celé dávky (špatná konfigurace nebo zrušený kontext).
func Collect[T, U any](ctx context.Context, cfg Config[T, U], inputs []T) ([]Result[T, U], Stats, error) {
	p, err := New(ctx, cfg)
	if err != nil {
		return nil, Stats{}, err
	}

	submitErr := make(chan error, 1)
	go func() {
		defer close(submitErr)
		defer p.Close()
		for _, v := range inputs {
			if err := p.Submit(ctx, v); err != nil {
				submitErr <- err
				return
			}
		}
	}()

	out := make([]Result[T, U], len(inputs))
	seen := make([]bool, len(inputs))
	for res := range p.Results() {
		if res.Index >= 0 && res.Index < len(out) {
			out[res.Index] = res
			seen[res.Index] = true
		}
	}

	if err := <-submitErr; err != nil {
		return nil, p.Stats(), err
	}
	if err := ctx.Err(); err != nil {
		return nil, p.Stats(), err
	}
	for i, ok := range seen {
		if !ok {
			return nil, p.Stats(), fmt.Errorf("pool: úloha %d nedoběhla", i)
		}
	}
	return out, p.Stats(), nil
}
