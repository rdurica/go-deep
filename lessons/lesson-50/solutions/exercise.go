// Package solutions obsahuje referenční řešení checkpointu fáze 5.
package solutions

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// Chyby konfigurace zpracovatele dávek.
var (
	// ErrInvalidWorkers znamená nekladný počet workerů.
	ErrInvalidWorkers = errors.New("batch: workers must be positive")
	// ErrInvalidQueueSize znamená zápornou kapacitu fronty.
	ErrInvalidQueueSize = errors.New("batch: queue size must not be negative")
	// ErrNilHandler znamená chybějící funkci pro zpracování úlohy.
	ErrNilHandler = errors.New("batch: handler must not be nil")
)

// Stats je snímek metrik jednoho běhu.
type Stats struct {
	// Processed je počet úloh dokončených bez chyby.
	Processed int
	// Failed je počet úloh, které skončily chybou.
	Failed int
	// MaxInFlight je nejvyšší pozorovaná souběžnost.
	MaxInFlight int
	// Elapsed je doba běhu.
	Elapsed time.Duration
}

// Total vrací počet dokončených úloh, tedy Processed + Failed.
func (s Stats) Total() int { return s.Processed + s.Failed }

// Throughput vrací průchodnost v úlohách za sekundu.
// Pro nekladnou dobu běhu vrací 0.
func (s Stats) Throughput() float64 {
	if s.Elapsed <= 0 {
		return 0
	}
	return float64(s.Total()) / s.Elapsed.Seconds()
}

// Metrics sbírá metriky souběžně z několika workerů.
// Nulová hodnota je použitelná.
type Metrics struct {
	processed   atomic.Int64
	failed      atomic.Int64
	inFlight    atomic.Int64
	maxInFlight atomic.Int64
}

// Enter oznámí začátek jedné úlohy a aktualizuje maximum souběhu.
func (m *Metrics) Enter() {
	cur := m.inFlight.Add(1)
	// Maximum se nedá zapsat jedním Store — mezi Load a Store by se vešla
	// jiná goroutina a maximum by se ztratilo. Proto CAS ve smyčce.
	for {
		old := m.maxInFlight.Load()
		if cur <= old || m.maxInFlight.CompareAndSwap(old, cur) {
			return
		}
	}
}

// Leave oznámí konec jedné úlohy. Nenulové err se počítá jako chyba.
func (m *Metrics) Leave(err error) {
	if err != nil {
		m.failed.Add(1)
	} else {
		m.processed.Add(1)
	}
	m.inFlight.Add(-1)
}

// Snapshot vrátí aktuální metriky doplněné o dobu běhu.
func (m *Metrics) Snapshot(elapsed time.Duration) Stats {
	return Stats{
		Processed:   int(m.processed.Load()),
		Failed:      int(m.failed.Load()),
		MaxInFlight: int(m.maxInFlight.Load()),
		Elapsed:     elapsed,
	}
}

// Item je jedna úloha v dávce.
type Item struct {
	ID      int
	Payload string
}

// Outcome je výsledek jedné úlohy.
type Outcome struct {
	ID    int
	Value string
	Err   error
}

// Config je konfigurace zpracovatele dávek.
type Config struct {
	// Workers je strop souběžnosti. Musí být kladný.
	Workers int
	// QueueSize je kapacita vstupní fronty. Nula znamená nejtvrdší
	// backpressure: producent čeká, dokud si úlohu nevezme volný worker.
	QueueSize int
	// FailFast zapne režim "první chyba ruší zbytek".
	FailFast bool
	// Handler je vlastní práce. Musí být zrušitelný přes kontext.
	Handler func(context.Context, Item) (string, error)
}

func (cfg Config) validate() error {
	if cfg.Workers <= 0 {
		return ErrInvalidWorkers
	}
	if cfg.QueueSize < 0 {
		return ErrInvalidQueueSize
	}
	if cfg.Handler == nil {
		return ErrNilHandler
	}
	return nil
}

// Process zpracuje celou dávku a vrátí výsledky ve stejném pořadí jako vstup.
//
// Chyby jednotlivých úloh zůstávají v Outcome. Návratová chyba znamená selhání
// celé dávky: špatná konfigurace, zrušený kontext, nebo první chyba v režimu
// FailFast (v tom případě je výsledek nil).
func Process(ctx context.Context, cfg Config, items []Item) ([]Outcome, Stats, error) {
	var m Metrics
	started := time.Now()

	if err := cfg.validate(); err != nil {
		return nil, m.Snapshot(0), err
	}
	if err := ctx.Err(); err != nil {
		return nil, m.Snapshot(time.Since(started)), err
	}
	if len(items) == 0 {
		return []Outcome{}, m.Snapshot(time.Since(started)), nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Pořadí drží index: každý worker zapisuje na vlastní políčko, což není
	// datový závod a je to levnější než řadit výsledky po dokončení.
	out := make([]Outcome, len(items))
	jobs := make(chan int, cfg.QueueSize)

	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
	)
	fail := func(err error) {
		once.Do(func() {
			firstErr = err
			cancel()
		})
	}

	wg.Add(cfg.Workers)
	for w := 0; w < cfg.Workers; w++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case i, ok := <-jobs:
					if !ok {
						return
					}
					m.Enter()
					v, err := cfg.Handler(ctx, items[i])
					m.Leave(err)
					out[i] = Outcome{ID: items[i].ID, Value: v, Err: err}
					if err != nil && cfg.FailFast {
						fail(err)
						return
					}
				}
			}
		}()
	}

	var feedErr error
	for i := range items {
		select {
		case jobs <- i:
		case <-ctx.Done():
			feedErr = ctx.Err()
		}
		if feedErr != nil {
			break
		}
	}
	close(jobs)
	wg.Wait()

	stats := m.Snapshot(time.Since(started))
	switch {
	case firstErr != nil:
		return nil, stats, firstErr
	case feedErr != nil:
		return nil, stats, feedErr
	default:
		return out, stats, nil
	}
}

// ProcessStream zpracuje neomezený proud úloh z kanálu in a výsledky posílá
// do out, který na konci zavře.
//
// Mezi in a workery je vlastní fronta o kapacitě cfg.QueueSize, takže odesílatel
// do in dostane backpressure, jakmile se zaplní. Funkce skončí, až se in zavře
// a fronta se vyprázdní, nebo při zrušení kontextu, nebo při první chybě
// v režimu FailFast. Kanál out nesmí být nil.
func ProcessStream(ctx context.Context, cfg Config, in <-chan Item, out chan<- Outcome) (Stats, error) {
	var m Metrics
	started := time.Now()

	// Kanál výsledků vlastníme, takže ho zavíráme za všech okolností —
	// i když skončíme na chybě konfigurace.
	defer close(out)

	if err := cfg.validate(); err != nil {
		return m.Snapshot(0), err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan Item, cfg.QueueSize)

	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
	)
	fail := func(err error) {
		once.Do(func() {
			firstErr = err
			cancel()
		})
	}

	wg.Add(cfg.Workers)
	for w := 0; w < cfg.Workers; w++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-jobs:
					if !ok {
						return
					}
					m.Enter()
					v, err := cfg.Handler(ctx, item)
					m.Leave(err)

					// Nikdy holé out <- …: kdyby konzument odešel,
					// worker by tady visel navždy.
					select {
					case out <- Outcome{ID: item.ID, Value: v, Err: err}:
					case <-ctx.Done():
						return
					}
					if err != nil && cfg.FailFast {
						fail(err)
						return
					}
				}
			}
		}()
	}

	var feedErr error
feed:
	for {
		select {
		case <-ctx.Done():
			feedErr = ctx.Err()
			break feed
		case item, ok := <-in:
			if !ok {
				break feed
			}
			select {
			case jobs <- item:
			case <-ctx.Done():
				feedErr = ctx.Err()
				break feed
			}
		}
	}
	close(jobs)
	wg.Wait()

	stats := m.Snapshot(time.Since(started))
	switch {
	case firstErr != nil:
		return stats, firstErr
	case feedErr != nil:
		return stats, feedErr
	default:
		return stats, nil
	}
}
