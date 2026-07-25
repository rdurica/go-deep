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

// Total vrací počet dokončených úloh, tedy Processed + Failed.
func (s Stats) Total() int {
	panic("TODO: úkol A")
}

// Throughput vrací průchodnost v úlohách za sekundu.
// Pro nekladnou dobu běhu vrací 0.
func (s Stats) Throughput() float64 {
	panic("TODO: úkol A")
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
	panic("TODO: úkol A")
}

// Leave oznámí konec jedné úlohy. Nenulové err se počítá jako chyba.
func (m *Metrics) Leave(err error) {
	panic("TODO: úkol A")
}

// Snapshot vrátí aktuální metriky doplněné o dobu běhu.
func (m *Metrics) Snapshot(elapsed time.Duration) Stats {
	panic("TODO: úkol A")
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

// Process zpracuje celou dávku a vrátí výsledky ve stejném pořadí jako vstup.
//
// Chyby jednotlivých úloh zůstávají v Outcome. Návratová chyba znamená selhání
// celé dávky: špatná konfigurace, zrušený kontext, nebo první chyba v režimu
// FailFast (v tom případě je výsledek nil).
func Process(ctx context.Context, cfg Config, items []Item) ([]Outcome, Stats, error) {
	panic("TODO: úkol B")
}

// ProcessStream zpracuje neomezený proud úloh z kanálu in a výsledky posílá
// do out, který na konci zavře.
//
// Mezi in a workery je vlastní fronta o kapacitě cfg.QueueSize, takže odesílatel
// do in dostane backpressure, jakmile se zaplní. Funkce skončí, až se in zavře
// a fronta se vyprázdní, nebo při zrušení kontextu, nebo při první chybě
// v režimu FailFast.
func ProcessStream(ctx context.Context, cfg Config, in <-chan Item, out chan<- Outcome) (Stats, error) {
	panic("TODO: úkol C")
}
