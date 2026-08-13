// Package exercise obsahuje cvičení lekce 46 — worker pool a omezování souběžnosti.
package exercise

import (
	"context"
	"errors"
	"sync"
)

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

// NewSemaphore vytvoří semafor s kapacitou n nad bufferovaným kanálem.
// Pro n <= 0 panikuje.
func NewSemaphore(n int) *Semaphore {
	if n <= 0 {
		panic("NewSemaphore: kapacita musí být kladná")
	}
	return &Semaphore{slots: make(chan struct{}, n)}
}

// Acquire zabere jedno místo a blokuje, dokud není volné nebo se nezruší ctx.
// Při chybě místo nezabírá; u už zrušeného kontextu vrací chybu i když je volno.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Chybí kontrola ctx.Err() před selectem.
// Najdi chybu a oprav — testy před opravou padají.
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// --- Stupeň: střední ---

// TryAcquire zabere místo, pokud je zrovna volné, a vrátí true.
// Nikdy neblokuje; při plném semaforu vrací false.
func (s *Semaphore) TryAcquire() bool {
	// TODO
	return false
}

// Release uvolní jedno místo na semaforu.
// Uvolnění bez předchozího Acquire panikuje.
func (s *Semaphore) Release() {
	// TODO
}

// Job je jedna úloha pro Pool.
type Job struct {
	ID  int
	Run func() (int, error)
}

// Result je výsledek jedné úlohy.
type Result struct {
	ID    int
	Value int
	Err   error
}

// Pool je worker pool s pevným počtem goroutin nad sdíleným kanálem úloh.
type Pool struct {
	jobs    chan Job
	results chan Result
	mu      sync.Mutex
	closed  bool
}

// New spustí worker pool s daným počtem workerů nad nebufferovaným kanálem úloh.
// Pro workers <= 0 panikuje. Z Results musíš číst souběžně se Submit.
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
			for job := range p.jobs {
				p.results <- run(job)
			}
		}()
	}
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

// Submit předá úlohu poolu a blokuje, dokud ji nepřevezme worker.
// Po Close vrací ErrPoolClosed. Job s Run == nil dá Result{Err: ErrNilJob}.
func (p *Pool) Submit(job Job) error {
	// TODO
	return nil
}

// Results vrací kanál výsledků úloh přijatých před Close.
func (p *Pool) Results() <-chan Result {
	return p.results
}

// Close oznámí poolu, že další úlohy nepřijdou. Idempotentní.
func (p *Pool) Close() {
	// TODO
}
