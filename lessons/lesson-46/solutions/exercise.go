// Package solutions obsahuje referenční řešení lekce 46.
package solutions

import (
	"context"
	"errors"
	"sync"
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

// --- Stupeň: jednoduchý ---
// NewSemaphore vrátí semafor s kapacitou n. Pro n <= 0 panikuje.
func NewSemaphore(n int) *Semaphore {
	if n <= 0 {
		panic("NewSemaphore: kapacita musí být kladná")
	}
	return &Semaphore{slots: make(chan struct{}, n)}
}

// Acquire zabere jedno místo. Blokuje, dokud se místo neuvolní nebo se nezruší ctx.
// Při zrušení kontextu vrací ctx.Err() a místo NEzabere.
func (s *Semaphore) Acquire(ctx context.Context) error {
	// Zrušený kontext má přednost i tehdy, když je zrovna místo volné —
	// jinak by select vybral náhodně a chování by bylo nedeterministické.
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case s.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire zabere místo, pokud je zrovna volné, a vrátí true. Nikdy neblokuje.
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.slots <- struct{}{}:
		return true
	default:
		return false
	}
}

// Release uvolní jedno místo. Uvolnění bez předchozího zabrání panikuje.
func (s *Semaphore) Release() {
	select {
	case <-s.slots:
	default:
		panic("Semaphore.Release: uvolnění bez Acquire")
	}
}

// --- Stupeň: střední ---
// LimitedMap aplikuje f na všechny vstupy, nejvýš limit z nich souběžně,
// a vrátí výsledky ve stejném pořadí jako vstup.
// Pro limit <= 0 se použije limit 1, pro f == nil vrací nil.
// Když se ctx zruší, nespuštěné položky zůstanou na prázdném řetězci.
func LimitedMap(ctx context.Context, inputs []string, limit int, f func(string) string) []string {
	if f == nil {
		return nil
	}
	if limit <= 0 {
		limit = 1
	}

	// Pořadí drží index ve výsledku: každá goroutina zapisuje na vlastní
	// políčko předalokovaného slice, což není datový závod.
	out := make([]string, len(inputs))
	sem := NewSemaphore(limit)
	var wg sync.WaitGroup

	for i, in := range inputs {
		if err := sem.Acquire(ctx); err != nil {
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer sem.Release()
			out[i] = f(in)
		}()
	}
	wg.Wait()
	return out
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

	// mu chrání closed a zároveň serializuje odeslání do jobs, aby Submit
	// nikdy nezapsal do kanálu, který mezitím zavřel Close.
	mu     sync.Mutex
	closed bool
}

// New spustí pool s daným počtem workerů. Pro workers <= 0 panikuje.
func New(workers int) *Pool {
	if workers <= 0 {
		panic("pool.New: workers musí být kladné")
	}
	p := &Pool{
		jobs:    make(chan Job),
		results: make(chan Result, workers),
	}

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			// range nad kanálem skončí sám, jakmile ho Close zavře.
			for job := range p.jobs {
				p.results <- run(job)
			}
		}()
	}
	// Kanál výsledků zavírá ten, kdo do něj zapisuje — až všichni dopíšou.
	go func() {
		wg.Wait()
		close(p.results)
	}()

	return p
}

func run(job Job) Result {
	if job.Run == nil {
		return Result{ID: job.ID, Err: ErrNilJob}
	}
	v, err := job.Run()
	return Result{ID: job.ID, Value: v, Err: err}
}

// --- Stupeň: obtížný ---
// Submit předá úlohu poolu. Blokuje, dokud ji nepřevezme worker.
// Po Close vrací ErrPoolClosed a úlohu zahodí.
func (p *Pool) Submit(job Job) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return ErrPoolClosed
	}
	p.jobs <- job
	return nil
}

// Results vrací kanál výsledků. Zavře se, až doběhnou všechny úlohy přijaté před Close.
func (p *Pool) Results() <-chan Result {
	return p.results
}

// Close oznámí poolu, že další úlohy nepřijdou. Je bezpečné ho volat opakovaně.
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return
	}
	p.closed = true
	close(p.jobs)
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
	if limit <= 0 {
		return nil, ErrInvalidWorkers
	}
	if f == nil {
		return nil, ErrNilJob
	}

	out := make([]U, len(in))
	sem := NewSemaphore(limit)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
	)
	fail := func(err error) {
		once.Do(func() {
			firstErr = err
			cancel() // ostatní goroutiny se dozvědí, že se práce ruší
		})
	}

	for i, v := range in {
		if err := sem.Acquire(ctx); err != nil {
			fail(err)
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer sem.Release()
			u, err := f(ctx, v)
			if err != nil {
				fail(err)
				return
			}
			out[i] = u
		}()
	}
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	return out, nil
}
