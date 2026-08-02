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

// --- Stupeň: jednoduchý ---
// NewSemaphore vytvoří semafor s kapacitou n nad bufferovaným kanálem.
// Pro n <= 0 panikuje. Kapacita drží počet volných míst ve slots kanálu.
func NewSemaphore(n int) *Semaphore {
	// TODO
	return nil
}

// Acquire zabere jedno místo a blokuje, dokud není volné nebo se nezruší ctx.
// Při chybě místo nezabírá; u už zrušeného kontextu vrací chybu i když je volno.
func (s *Semaphore) Acquire(ctx context.Context) error {
	// TODO
	return nil
}

// TryAcquire zabere místo, pokud je zrovna volné, a vrátí true.
// Nikdy neblokuje; při plném semaforu vrací false.
func (s *Semaphore) TryAcquire() bool {
	// TODO
	return false
}

// Release uvolní jedno místo na semaforu.
// Uvolnění bez předchozího Acquire panikuje; NewSemaphore(0) taky panikuje.
func (s *Semaphore) Release() {
	// TODO
}

// --- Stupeň: střední ---
// LimitedMap aplikuje f na všechny vstupy, nejvýš limit z nich souběžně,
// a vrátí výsledky ve stejném pořadí jako vstup.
// Pro limit <= 0 se použije limit 1, pro f == nil vrací nil.
// Když se ctx zruší, nespuštěné položky zůstanou na prázdném řetězci.
// Po návratu nesmí zůstat běžící goroutina.
func LimitedMap(ctx context.Context, inputs []string, limit int, f func(string) string) []string {
	// TODO
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

// New spustí worker pool s daným počtem workerů nad nebufferovaným kanálem úloh.
// Pro workers <= 0 panikuje. Z Results musíš číst souběžně se Submit (backpressure).
func New(workers int) *Pool {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---
// Submit předá úlohu poolu a blokuje, dokud ji nepřevezme worker (backpressure).
// Po Close vrací ErrPoolClosed a úlohu zahodí; nesmí zapsat do zavřeného kanálu.
// Job s Run == nil dá Result{Err: ErrNilJob}, ne paniku.
func (p *Pool) Submit(job Job) error {
	// TODO
	return nil
}

// Results vrací kanál výsledků úloh přijatých před Close.
// Kanál se zavře, až doběhnou všechny tyto úlohy.
func (p *Pool) Results() <-chan Result {
	// TODO
	return nil
}

// Close oznámí poolu, že další úlohy nepřijdou.
// Idempotentní — opakované volání je bezpečné.
func (p *Pool) Close() {
	// TODO
}

// MapErr aplikuje f na všechny prvky in, nejvýš limit z nich souběžně,
// a vrátí výsledky ve stejném pořadí jako vstup.
//
// První chyba zruší odvozený kontext (takže se běžící volání f dozvědí, že
// nemá cenu pokračovat), MapErr počká na doběhnutí rozběhnutých goroutin
// a vrátí (nil, tu chybu).
// Pro limit <= 0 vrací ErrInvalidWorkers, pro f == nil ErrNilJob.
// Prázdný vstup vrací prázdný výsledek a nil; zrušený vstupní kontext vrací ctx.Err().
func MapErr[T, U any](ctx context.Context, in []T, limit int, f func(context.Context, T) (U, error)) ([]U, error) {
	// TODO
	return nil, nil
}
