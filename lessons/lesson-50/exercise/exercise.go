// Package exercise obsahuje kumulativní cvičení checkpointu fáze 5.
package exercise

import (
	"context"
	"errors"
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

// --- Stupeň: jednoduchý ---
// Total vrací počet dokončených úloh, tedy Processed + Failed.
// Nezahrnuje úlohy, které se ještě zpracovávají.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Počítá jen Processed a zapomíná Failed.
func (s Stats) Total() int {
	return s.Processed
}

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

// --- Stupeň: střední ---
// Enter oznámí začátek úlohy a aktualizuje maximum souběhu přes CompareAndSwap ve smyčce.
// Nulová Metrics je použitelná.
func (m *Metrics) Enter() {
	// TODO
}

// Leave oznámí konec jedné úlohy. Nenulové err se počítá jako chyba.
// Volá se vždy po Enter, typicky přes defer.
func (m *Metrics) Leave(err error) {
	// TODO
}

// Snapshot vrátí Stats s Processed, Failed a MaxInFlight z atomik a Elapsed z parametru.
// Nulová Metrics je použitelná.
func (m *Metrics) Snapshot(elapsed time.Duration) Stats {
	return Stats{
		Processed:   int(m.processed.Load()),
		Failed:      int(m.failed.Load()),
		MaxInFlight: int(m.maxInFlight.Load()),
		Elapsed:     elapsed,
	}
}

// --- Stupeň: obtížný ---

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

// Process zpracuje dávku s limitem nejvýš cfg.Workers souběžných handlerů.
// Fronta má kapacitu cfg.QueueSize (0 = předání z ruky do ruky). Výsledky v pořadí vstupu.
// Chyby úloh patří do Outcome.Err (dávka kvůli nim neselže); FailFast zruší odvozený
// kontext, počká na rozběhnuté goroutiny a vrátí (nil, stats, první chyba).
// Zrušený ctx → ctx.Err(); neplatná konfigurace → ErrInvalidWorkers, ErrInvalidQueueSize,
// ErrNilHandler; prázdná dávka → prázdný výsledek a nil.
// Po návratu nesmí zůstat živá goroutina.
func Process(ctx context.Context, cfg Config, items []Item) ([]Outcome, Stats, error) {
	// TODO
	return nil, *new(Stats), nil
}
