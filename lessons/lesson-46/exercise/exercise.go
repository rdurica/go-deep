// Package exercise obsahuje cvičení lekce 46 — worker pool a omezování souběžnosti.
package exercise

import (
	"context"
	"errors"
)

// ErrInvalidWorkers signalizuje nekladný počet workerů nebo nekladný limit.
var ErrInvalidWorkers = errors.New("pool: workers must be positive")

// ErrNilJob signalizuje úlohu bez funkce Run.
var ErrNilJob = errors.New("pool: job has no Run function")

// ErrPoolClosed vrací Submit poté, co byl pool zavřený.
var ErrPoolClosed = errors.New("pool: closed")

// Semaphore omezuje počet souběžně běžících úseků kódu.
// Nulová hodnota není použitelná, vytvoř ho přes NewSemaphore.
type Semaphore struct {
	slots chan struct{}
}

// NewSemaphore vrátí semafor s kapacitou n. Pro n <= 0 panikuje.
func NewSemaphore(n int) *Semaphore {
	// TODO: úkol A
	return nil
}

// Acquire zabere jedno místo. Blokuje, dokud se místo neuvolní nebo se nezruší ctx.
// Při zrušení kontextu vrací ctx.Err() a místo NEzabere.
func (s *Semaphore) Acquire(ctx context.Context) error {
	// TODO: úkol A
	return nil
}

// TryAcquire zabere místo, pokud je zrovna volné, a vrátí true. Nikdy neblokuje.
func (s *Semaphore) TryAcquire() bool {
	// TODO: úkol A
	return false
}

// Release uvolní jedno místo. Uvolnění bez předchozího zabrání panikuje.
func (s *Semaphore) Release() {
	// TODO: úkol A
}

// LimitedMap aplikuje f na všechny vstupy, nejvýš limit z nich souběžně,
// a vrátí výsledky ve stejném pořadí jako vstup.
// Pro limit <= 0 se použije limit 1, pro f == nil vrací nil.
// Když se ctx zruší, nespuštěné položky zůstanou na prázdném řetězci.
func LimitedMap(ctx context.Context, inputs []string, limit int, f func(string) string) []string {
	// TODO: úkol A
	return nil
}

// Job je jedna úloha pro Pool.
type Job struct {
	// ID slouží ke spárování úlohy s výsledkem.
	ID int
	// Run je vlastní práce. Když je nil, výsledek nese ErrNilJob.
	Run func() (int, error)
}

// Result je výsledek jedné úlohy.
type Result struct {
	ID    int
	Value int
	Err   error
}

// Pool je worker pool s pevným počtem goroutin nad sdíleným kanálem úloh.
//
// Vstupní kanál je nebufferovaný, takže Submit tlačí zpět (backpressure),
// dokud se neuvolní worker. Z Results musíš číst souběžně se Submit,
// jinak se pool zablokuje.
type Pool struct {
	jobs    chan Job
	results chan Result
}

// New spustí pool s daným počtem workerů. Pro workers <= 0 panikuje.
func New(workers int) *Pool {
	// TODO: úkol B
	return nil
}

// Submit předá úlohu poolu. Blokuje, dokud ji nepřevezme worker.
// Po Close vrací ErrPoolClosed a úlohu zahodí.
func (p *Pool) Submit(job Job) error {
	// TODO: úkol B
	return nil
}

// Results vrací kanál výsledků. Zavře se, až doběhnou všechny úlohy přijaté před Close.
func (p *Pool) Results() <-chan Result {
	// TODO: úkol B
	ch := make(chan Result)
	close(ch)
	return ch
}

// Close oznámí poolu, že další úlohy nepřijdou. Je bezpečné ho volat opakovaně.
func (p *Pool) Close() {
	// TODO: úkol B
}

// MapErr aplikuje f na všechny prvky in, nejvýš limit z nich souběžně,
// a vrátí výsledky ve stejném pořadí jako vstup.
//
// První chyba zruší odvozený kontext (takže se běžící volání f dozvědí, že
// nemá cenu pokračovat), MapErr počká na doběhnutí rozběhnutých goroutin
// a vrátí (nil, tu chybu).
// Pro limit <= 0 vrací ErrInvalidWorkers, pro f == nil ErrNilJob.
// Zrušený vstupní kontext vrací ctx.Err().
func MapErr[T, U any](ctx context.Context, in []T, limit int, f func(context.Context, T) (U, error)) ([]U, error) {
	// TODO: úkol C
	return nil, nil
}
