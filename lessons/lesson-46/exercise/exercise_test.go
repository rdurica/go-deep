package exercise_test

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-46/exercise"
)

// waitNoLeak počká, až se počet goroutin vrátí zpátky na výchozí hodnotu.
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

// barrier pustí dál vždy až n goroutin najednou. Používáme ho k tomu, aby byl
// souběh deterministický: když implementace opravdu pouští n úloh současně,
// bariéra se okamžitě uvolní. Když jich pouští míň, spadneme na timeout a
// test to pozná na naměřeném maximu, ne na zaseknutí.
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

// tracker měří skutečně dosažený souběh.
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

	t.Run("už zrušený kontext", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := s.Acquire(ctx); !errors.Is(err, context.Canceled) {
			t.Errorf("Acquire se zrušeným kontextem = %v, chci Canceled", err)
		}
	})

	s.Release()

	t.Run("místo se po neúspěchu nezabralo", func(t *testing.T) {
		if !s.TryAcquire() {
			t.Error("TryAcquire() = false, neúspěšný Acquire nesmí zabrat místo")
		}
		s.Release()
	})
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

func TestLimitedMapKeepsOrder(t *testing.T) {
	before := runtime.NumGoroutine()
	inputs := []string{"a", "bb", "ccc", "dddd", "e", "ff", "g", "hh"}

	got := exercise.LimitedMap(context.Background(), inputs, 3, strings.ToUpper)

	if len(got) != len(inputs) {
		t.Fatalf("LimitedMap vrátil %d prvků, chci %d", len(got), len(inputs))
	}
	for i, in := range inputs {
		if want := strings.ToUpper(in); got[i] != want {
			t.Errorf("výsledek[%d] = %q, chci %q", i, got[i], want)
		}
	}
	waitNoLeak(t, before)
}

func TestLimitedMapRespectsLimit(t *testing.T) {
	before := runtime.NumGoroutine()
	const limit = 4
	inputs := make([]string, 12) // násobek limitu, aby bariéra vždy dorazila do plného počtu
	for i := range inputs {
		inputs[i] = "x"
	}

	var tr tracker
	b := newBarrier(limit)
	f := func(s string) string {
		tr.enter()
		b.wait()
		tr.leave()
		return s + "!"
	}

	got := exercise.LimitedMap(context.Background(), inputs, limit, f)

	if len(got) != len(inputs) {
		t.Fatalf("LimitedMap vrátil %d prvků, chci %d", len(got), len(inputs))
	}
	for i := range got {
		if got[i] != "x!" {
			t.Errorf("výsledek[%d] = %q, chci %q", i, got[i], "x!")
		}
	}
	if peak := tr.peak(); peak != limit {
		t.Errorf("maximální souběh = %d, chci přesně %d", peak, limit)
	}
	waitNoLeak(t, before)
}

func TestLimitedMapEdgeCases(t *testing.T) {
	before := runtime.NumGoroutine()

	if got := exercise.LimitedMap(context.Background(), []string{"a"}, 3, nil); got != nil {
		t.Errorf("LimitedMap s nil funkcí = %v, chci nil", got)
	}
	if got := exercise.LimitedMap(context.Background(), nil, 3, strings.ToUpper); len(got) != 0 {
		t.Errorf("LimitedMap(nil) = %v, chci prázdný výsledek", got)
	}
	if got := exercise.LimitedMap(context.Background(), []string{"a", "b"}, 0, strings.ToUpper); len(got) != 2 || got[0] != "A" || got[1] != "B" {
		t.Errorf("LimitedMap s limitem 0 = %v, chci [A B]", got)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got := exercise.LimitedMap(ctx, []string{"a", "b", "c"}, 2, strings.ToUpper)
	if len(got) != 3 {
		t.Fatalf("LimitedMap se zrušeným kontextem vrátil %d prvků, chci 3", len(got))
	}
	for i, v := range got {
		if v != "" {
			t.Errorf("výsledek[%d] = %q, po zrušení kontextu chci prázdný řetězec", i, v)
		}
	}
	waitNoLeak(t, before)
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
	p.Close() // opakované volání musí být bezpečné

	err := p.Submit(exercise.Job{ID: 2, Run: func() (int, error) { return 0, nil }})
	if !errors.Is(err, exercise.ErrPoolClosed) {
		t.Errorf("Submit po Close = %v, chci ErrPoolClosed", err)
	}

	for range p.Results() { //nolint:revive // dočerpáme, kanál se musí zavřít
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

	for range p.Results() { //nolint:revive
	}
	waitNoLeak(t, before)
}

func TestPoolPanicsOnNonPositiveWorkers(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("New(0) měl panikovat")
		}
	}()
	_ = exercise.New(0)
}

func TestMapErrKeepsOrder(t *testing.T) {
	before := runtime.NumGoroutine()
	in := []int{1, 2, 3, 4, 5, 6, 7}

	got, err := exercise.MapErr(context.Background(), in, 3, func(_ context.Context, v int) (string, error) {
		return strings.Repeat("*", v), nil
	})
	if err != nil {
		t.Fatalf("MapErr = %v, chci nil", err)
	}
	if len(got) != len(in) {
		t.Fatalf("MapErr vrátil %d prvků, chci %d", len(got), len(in))
	}
	for i, v := range in {
		if want := strings.Repeat("*", v); got[i] != want {
			t.Errorf("výsledek[%d] = %q, chci %q", i, got[i], want)
		}
	}
	waitNoLeak(t, before)
}

func TestMapErrRespectsLimit(t *testing.T) {
	before := runtime.NumGoroutine()
	const limit = 3
	in := make([]int, 12)
	for i := range in {
		in[i] = i
	}

	var tr tracker
	b := newBarrier(limit)
	got, err := exercise.MapErr(context.Background(), in, limit, func(_ context.Context, v int) (int, error) {
		tr.enter()
		b.wait()
		tr.leave()
		return v + 100, nil
	})
	if err != nil {
		t.Fatalf("MapErr = %v, chci nil", err)
	}
	for i := range in {
		if got[i] != i+100 {
			t.Errorf("výsledek[%d] = %d, chci %d", i, got[i], i+100)
		}
	}
	if peak := tr.peak(); peak != limit {
		t.Errorf("maximální souběh = %d, chci přesně %d", peak, limit)
	}
	waitNoLeak(t, before)
}

func TestMapErrFirstErrorCancelsRest(t *testing.T) {
	before := runtime.NumGoroutine()
	errBoom := errors.New("boom")
	const limit = 4
	in := []int{0, 1, 2, 3, 4, 5, 6, 7}

	var canceled atomic.Int64
	// Bariéra zaručí, že se všechny první čtyři úlohy opravdu rozběhnou dřív,
	// než ta první ohlásí chybu. Bez toho by test byl nedeterministický.
	b := newBarrier(limit)

	start := time.Now()
	got, err := exercise.MapErr(context.Background(), in, limit, func(ctx context.Context, v int) (int, error) {
		b.wait()
		if v == 0 {
			return 0, errBoom
		}
		// Ostatní úlohy čekají na zrušení. Kdyby ho MapErr neudělal,
		// spadneme až na velkorysém timeoutu a test to pozná.
		select {
		case <-ctx.Done():
			canceled.Add(1)
			return 0, ctx.Err()
		case <-time.After(3 * time.Second):
			return v, nil
		}
	})
	elapsed := time.Since(start)

	if !errors.Is(err, errBoom) {
		t.Errorf("MapErr = %v, chci první chybu %v", err, errBoom)
	}
	if got != nil {
		t.Errorf("MapErr vrátil %v, při chybě chci nil výsledek", got)
	}
	if got := canceled.Load(); got < limit-1 {
		t.Errorf("zrušení dostalo %d rozběhnutých úloh, chci aspoň %d", got, limit-1)
	}
	if elapsed > 2*time.Second {
		t.Errorf("MapErr se vrátil až za %v, chci rychle po první chybě", elapsed)
	}
	waitNoLeak(t, before)
}

func TestMapErrValidation(t *testing.T) {
	f := func(_ context.Context, v int) (int, error) { return v, nil }

	if _, err := exercise.MapErr(context.Background(), []int{1}, 0, f); !errors.Is(err, exercise.ErrInvalidWorkers) {
		t.Errorf("MapErr s limitem 0 = %v, chci ErrInvalidWorkers", err)
	}
	if _, err := exercise.MapErr[int, int](context.Background(), []int{1}, 2, nil); !errors.Is(err, exercise.ErrNilJob) {
		t.Errorf("MapErr s nil funkcí = %v, chci ErrNilJob", err)
	}

	got, err := exercise.MapErr(context.Background(), nil, 2, f)
	if err != nil || len(got) != 0 {
		t.Errorf("MapErr(nil) = (%v, %v), chci (prázdné, nil)", got, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := exercise.MapErr(ctx, []int{1, 2}, 2, f); !errors.Is(err, context.Canceled) {
		t.Errorf("MapErr se zrušeným kontextem = %v, chci Canceled", err)
	}
}
