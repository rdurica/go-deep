package solutions_test

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-46/solutions"
)

func waitNoLeak(t *testing.T, before int) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if runtime.NumGoroutine() <= before+1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("goroutine leak: před testem %d goroutin, po testu %d", before, runtime.NumGoroutine())
}

type barrier struct {
	mu    sync.Mutex
	n     int
	count int
	ch    chan struct{}
}

func newBarrier(n int) *barrier {
	return &barrier{n: n, ch: make(chan struct{})}
}

func (b *barrier) wait() {
	b.mu.Lock()
	b.count++
	ch := b.ch
	if b.count == b.n {
		b.count = 0
		b.ch = make(chan struct{})
		b.mu.Unlock()
		close(ch)
		return
	}
	b.mu.Unlock()

	select {
	case <-ch:
	case <-time.After(400 * time.Millisecond):
	}
}

type tracker struct {
	cur atomic.Int64
	max atomic.Int64
}

func (tr *tracker) enter() {
	cur := tr.cur.Add(1)
	for {
		max := tr.max.Load()
		if cur <= max || tr.max.CompareAndSwap(max, cur) {
			return
		}
	}
}

func (tr *tracker) leave() { tr.cur.Add(-1) }

func (tr *tracker) peak() int { return int(tr.max.Load()) }

func TestSemaphoreAcquireRespectsContext(t *testing.T) {
	s := exercise.NewSemaphore(1)
	if err := s.Acquire(context.Background()); err != nil {
		t.Fatalf("Acquire = %v, chci nil", err)
	}

	t.Run("deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		start := time.Now()
		err := s.Acquire(ctx)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Acquire na plném semaforu = %v, chci DeadlineExceeded", err)
		}
		if elapsed := time.Since(start); elapsed > 2*time.Second {
			t.Errorf("Acquire se vrátil až za %v, chci hned po deadline", elapsed)
		}
	})

	t.Run("already canceled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := s.Acquire(ctx); !errors.Is(err, context.Canceled) {
			t.Errorf("Acquire se zrušeným kontextem = %v, chci Canceled", err)
		}
	})

	s.Release()

	t.Run("slot not taken after failure", func(t *testing.T) {
		if !s.TryAcquire() {
			t.Error("TryAcquire() = false, neúspěšný Acquire nesmí zabrat místo")
		}
		s.Release()
	})
}

func TestSemaphoreCapacity(t *testing.T) {
	before := runtime.NumGoroutine()
	s := exercise.NewSemaphore(3)

	for i := 0; i < 3; i++ {
		if err := s.Acquire(context.Background()); err != nil {
			t.Fatalf("Acquire #%d = %v, chci nil", i, err)
		}
	}
	if s.TryAcquire() {
		t.Error("TryAcquire() = true na plném semaforu, chci false")
	}
	s.Release()
	if !s.TryAcquire() {
		t.Error("TryAcquire() = false po Release, chci true")
	}
	for i := 0; i < 3; i++ {
		s.Release()
	}
	waitNoLeak(t, before)
}

func TestSemaphoreReleaseWithoutAcquirePanics(t *testing.T) {
	s := exercise.NewSemaphore(2)
	defer func() {
		if recover() == nil {
			t.Error("Release bez Acquire měl panikovat")
		}
	}()
	s.Release()
}

func TestPoolPanicsOnNonPositiveWorkers(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("New(0) měl panikovat")
		}
	}()
	_ = exercise.New(0)
}

func TestPoolProcessesAllJobs(t *testing.T) {
	before := runtime.NumGoroutine()
	const (
		workers = 4
		jobs    = 20
	)

	var tr tracker
	b := newBarrier(workers)
	p := exercise.New(workers)

	go func() {
		for i := 0; i < jobs; i++ {
			id := i
			err := p.Submit(exercise.Job{ID: id, Run: func() (int, error) {
				tr.enter()
				b.wait()
				tr.leave()
				return id * 2, nil
			}})
			if err != nil {
				t.Errorf("Submit(%d) = %v, chci nil", id, err)
				return
			}
		}
		p.Close()
	}()

	seen := make(map[int]int)
	for res := range p.Results() {
		if res.Err != nil {
			t.Errorf("úloha %d skončila chybou %v", res.ID, res.Err)
		}
		seen[res.ID] = res.Value
	}

	if len(seen) != jobs {
		t.Fatalf("dorazilo %d výsledků, chci %d", len(seen), jobs)
	}
	for i := 0; i < jobs; i++ {
		if seen[i] != i*2 {
			t.Errorf("výsledek úlohy %d = %d, chci %d", i, seen[i], i*2)
		}
	}
	if peak := tr.peak(); peak > workers {
		t.Errorf("maximální souběh = %d, pool nesmí spustit víc než %d úloh najednou", peak, workers)
	}
	if peak := tr.peak(); peak < workers {
		t.Errorf("maximální souběh = %d, chci %d — pool nevyužívá všechny workery", peak, workers)
	}
	waitNoLeak(t, before)
}

func TestPoolSubmitAfterClose(t *testing.T) {
	before := runtime.NumGoroutine()
	p := exercise.New(2)

	if err := p.Submit(exercise.Job{ID: 1, Run: func() (int, error) { return 7, nil }}); err != nil {
		t.Fatalf("Submit = %v, chci nil", err)
	}
	if res := <-p.Results(); res.Value != 7 {
		t.Errorf("výsledek = %d, chci 7", res.Value)
	}

	p.Close()
	p.Close()

	err := p.Submit(exercise.Job{ID: 2, Run: func() (int, error) { return 0, nil }})
	if !errors.Is(err, exercise.ErrPoolClosed) {
		t.Errorf("Submit po Close = %v, chci ErrPoolClosed", err)
	}

	for range p.Results() {
	}
	waitNoLeak(t, before)
}

func TestPoolNilJobRun(t *testing.T) {
	before := runtime.NumGoroutine()
	p := exercise.New(1)

	if err := p.Submit(exercise.Job{ID: 9}); err != nil {
		t.Fatalf("Submit = %v, chci nil", err)
	}
	res := <-p.Results()
	if res.ID != 9 || !errors.Is(res.Err, exercise.ErrNilJob) {
		t.Errorf("výsledek = %+v, chci ID 9 a ErrNilJob", res)
	}
	p.Close()

	for range p.Results() {
	}
	waitNoLeak(t, before)
}
