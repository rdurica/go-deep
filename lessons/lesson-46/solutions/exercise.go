// Package solutions obsahuje referenční řešení lekce 46.
package solutions

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
type Semaphore struct {
	slots chan struct{}
}

// --- Stupeň: jednoduchý ---

// NewSemaphore vrátí semafor s kapacitou n.
func NewSemaphore(n int) *Semaphore {
	if n <= 0 {
		panic("NewSemaphore: kapacita musí být kladná")
	}
	return &Semaphore{slots: make(chan struct{}, n)}
}

// Acquire zabere jedno místo.
func (s *Semaphore) Acquire(ctx context.Context) error {
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

// --- Stupeň: střední ---

// TryAcquire zabere místo, pokud je volné.
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.slots <- struct{}{}:
		return true
	default:
		return false
	}
}

// Release uvolní jedno místo.
func (s *Semaphore) Release() {
	select {
	case <-s.slots:
	default:
		panic("Semaphore.Release: uvolnění bez Acquire")
	}
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

// Pool je worker pool s pevným počtem goroutin.
type Pool struct {
	jobs    chan Job
	results chan Result
	mu      sync.Mutex
	closed  bool
}

// New spustí pool s daným počtem workerů.
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

// Submit předá úlohu poolu.
func (p *Pool) Submit(job Job) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return ErrPoolClosed
	}
	p.jobs <- job
	return nil
}

// Results vrací kanál výsledků.
func (p *Pool) Results() <-chan Result {
	return p.results
}

// Close oznámí poolu, že další úlohy nepřijdou.
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return
	}
	p.closed = true
	close(p.jobs)
}
