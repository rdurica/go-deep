package pool_test

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p04-worker-pool/pool"
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

// barrier pustí dál vždy až n goroutin najednou. Díky tomu je naměřený souběh
// deterministický: když pool opravdu pouští n úloh současně, bariéra se uvolní
// okamžitě. Když jich pouští méně, spadne na timeout a test to pozná na metrice.
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

func upper(_ context.Context, s string) (string, error) {
	return strings.ToUpper(s), nil
}

func TestNewValidatesConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  pool.Config[string, string]
		want error
	}{
		{"bez workerů", pool.Config[string, string]{Workers: 0, Handler: upper}, pool.ErrInvalidWorkers},
		{"záporná fronta", pool.Config[string, string]{Workers: 2, QueueSize: -1, Handler: upper}, pool.ErrInvalidQueueSize},
		{"bez handleru", pool.Config[string, string]{Workers: 2}, pool.ErrNilHandler},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := pool.New(context.Background(), tt.cfg); !errors.Is(err, tt.want) {
				t.Errorf("New = %v, chci %v", err, tt.want)
			}
		})
	}
}

func TestCollectKeepsOrder(t *testing.T) {
	before := runtime.NumGoroutine()

	inputs := make([]string, 50)
	for i := range inputs {
		inputs[i] = fmt.Sprintf("job-%02d", i)
	}

	got, stats, err := pool.Collect(context.Background(), pool.Config[string, string]{
		Workers:   4,
		QueueSize: 8,
		Handler:   upper,
	}, inputs)
	if err != nil {
		t.Fatalf("Collect = %v, chci nil", err)
	}

	if len(got) != len(inputs) {
		t.Fatalf("Collect vrátil %d výsledků, chci %d", len(got), len(inputs))
	}
	for i, in := range inputs {
		if got[i].Index != i {
			t.Errorf("výsledek na pozici %d má Index %d", i, got[i].Index)
		}
		if got[i].Input != in {
			t.Errorf("výsledek[%d].Input = %q, chci %q", i, got[i].Input, in)
		}
		if want := strings.ToUpper(in); got[i].Value != want {
			t.Errorf("výsledek[%d].Value = %q, chci %q", i, got[i].Value, want)
		}
		if got[i].Err != nil {
			t.Errorf("výsledek[%d].Err = %v, chci nil", i, got[i].Err)
		}
	}

	if stats.Submitted != len(inputs) || stats.Processed != len(inputs) || stats.Failed != 0 {
		t.Errorf("Stats = %+v, chci submitted=processed=%d a failed=0", stats, len(inputs))
	}
	if stats.Total() != len(inputs) {
		t.Errorf("Stats.Total() = %d, chci %d", stats.Total(), len(inputs))
	}
	if stats.Elapsed <= 0 {
		t.Errorf("Stats.Elapsed = %v, chci měřitelnou dobu", stats.Elapsed)
	}
	waitNoLeak(t, before)
}

func TestCollectEmptyInput(t *testing.T) {
	before := runtime.NumGoroutine()

	got, stats, err := pool.Collect(context.Background(), pool.Config[string, string]{
		Workers: 3,
		Handler: upper,
	}, nil)
	if err != nil {
		t.Fatalf("Collect(nil) = %v, chci nil", err)
	}
	if len(got) != 0 {
		t.Errorf("Collect(nil) = %v, chci prázdný výsledek", got)
	}
	if stats.Submitted != 0 || stats.Total() != 0 {
		t.Errorf("Stats = %+v, chci nuly", stats)
	}
	waitNoLeak(t, before)
}

func TestPoolRespectsConcurrencyLimit(t *testing.T) {
	before := runtime.NumGoroutine()
	const (
		workers = 4
		jobs    = 20 // násobek workers, ať bariéra vždy dorazí do plného počtu
	)

	b := newBarrier(workers)
	inputs := make([]int, jobs)
	for i := range inputs {
		inputs[i] = i
	}

	got, stats, err := pool.Collect(context.Background(), pool.Config[int, int]{
		Workers:   workers,
		QueueSize: 2,
		Handler: func(_ context.Context, v int) (int, error) {
			b.wait()
			return v * 3, nil
		},
	}, inputs)
	if err != nil {
		t.Fatalf("Collect = %v, chci nil", err)
	}

	for i := range inputs {
		if got[i].Value != i*3 {
			t.Errorf("výsledek[%d] = %d, chci %d", i, got[i].Value, i*3)
		}
	}
	if stats.MaxInFlight > workers {
		t.Errorf("MaxInFlight = %d, pool nesmí spustit víc než %d úloh naráz", stats.MaxInFlight, workers)
	}
	if stats.MaxInFlight < workers {
		t.Errorf("MaxInFlight = %d, chci %d — pool nevyužívá všechny workery", stats.MaxInFlight, workers)
	}
	waitNoLeak(t, before)
}

func TestPoolCollectsErrors(t *testing.T) {
	before := runtime.NumGoroutine()
	errOdd := errors.New("liché číslo")

	inputs := []int{0, 1, 2, 3, 4, 5}
	got, stats, err := pool.Collect(context.Background(), pool.Config[int, int]{
		Workers: 3,
		Handler: func(_ context.Context, v int) (int, error) {
			if v%2 == 1 {
				return 0, fmt.Errorf("úloha %d: %w", v, errOdd)
			}
			return v, nil
		},
	}, inputs)

	// Chyby jednotlivých úloh nejsou chybou dávky.
	if err != nil {
		t.Fatalf("Collect = %v, chci nil", err)
	}
	for i, res := range got {
		wantErr := i%2 == 1
		if wantErr && !errors.Is(res.Err, errOdd) {
			t.Errorf("výsledek[%d].Err = %v, chci %v", i, res.Err, errOdd)
		}
		if !wantErr && res.Err != nil {
			t.Errorf("výsledek[%d].Err = %v, chci nil", i, res.Err)
		}
	}
	if stats.Failed != 3 || stats.Processed != 3 {
		t.Errorf("Stats = %+v, chci processed=3 a failed=3", stats)
	}
	waitNoLeak(t, before)
}

func TestPoolBackpressure(t *testing.T) {
	before := runtime.NumGoroutine()

	// Jeden worker, fronta na jednu úlohu. Handler drží workera, dokud
	// ho nepustíme — třetí Submit tedy nemá kam odložit úlohu.
	release := make(chan struct{})
	busy := make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p, err := pool.New(ctx, pool.Config[int, int]{
		Workers:   1,
		QueueSize: 1,
		Handler: func(_ context.Context, v int) (int, error) {
			busy <- struct{}{}
			<-release
			return v, nil
		},
	})
	if err != nil {
		t.Fatalf("New = %v, chci nil", err)
	}

	if err := p.Submit(ctx, 1); err != nil { // vezme si ho worker
		t.Fatalf("první Submit = %v, chci nil", err)
	}
	<-busy                                   // worker opravdu pracuje
	if err := p.Submit(ctx, 2); err != nil { // sedne do fronty
		t.Fatalf("druhý Submit = %v, chci nil", err)
	}

	// Třetí Submit musí blokovat. S krátkým deadlinem se z něj vrátí chyba.
	tight, cancelTight := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancelTight()
	start := time.Now()
	if err := p.Submit(tight, 3); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Submit do plné fronty = %v, chci DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed < 50*time.Millisecond {
		t.Errorf("Submit se vrátil za %v — fronta netlačila zpět", elapsed)
	}

	close(release)
	p.Close()

	var n int
	for range p.Results() {
		n++
	}
	if n != 2 {
		t.Errorf("dorazilo %d výsledků, chci 2", n)
	}
	if stats := p.Stats(); stats.Submitted != 2 {
		t.Errorf("Stats.Submitted = %d, chci 2", stats.Submitted)
	}
	waitNoLeak(t, before)
}

func TestSubmitAfterCloseFails(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx := context.Background()
	p, err := pool.New(ctx, pool.Config[string, string]{Workers: 2, QueueSize: 4, Handler: upper})
	if err != nil {
		t.Fatalf("New = %v, chci nil", err)
	}

	if err := p.Submit(ctx, "a"); err != nil {
		t.Fatalf("Submit = %v, chci nil", err)
	}
	p.Close()
	p.Close() // opakované volání musí být bezpečné

	if err := p.Submit(ctx, "b"); !errors.Is(err, pool.ErrClosed) {
		t.Errorf("Submit po Close = %v, chci ErrClosed", err)
	}

	var n int
	for range p.Results() {
		n++
	}
	if n != 1 {
		t.Errorf("dorazil %d výsledek, chci 1", n)
	}
	waitNoLeak(t, before)
}

func TestPoolStopsOnContextCancel(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	var started atomic.Int64

	p, err := pool.New(ctx, pool.Config[int, int]{
		Workers:   2,
		QueueSize: 64,
		Handler: func(ctx context.Context, v int) (int, error) {
			started.Add(1)
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-time.After(3 * time.Second):
				return v, nil
			}
		},
	})
	if err != nil {
		t.Fatalf("New = %v, chci nil", err)
	}

	for i := 0; i < 20; i++ {
		if err := p.Submit(ctx, i); err != nil {
			t.Fatalf("Submit(%d) = %v, chci nil", i, err)
		}
	}

	time.AfterFunc(30*time.Millisecond, cancel)
	start := time.Now()

	// Kanál výsledků se musí zavřít, i když je fronta ještě plná.
	for range p.Results() { //nolint:revive // jen dočerpáme
	}
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Errorf("pool se zastavoval %v, chci rychlé ukončení po zrušení", elapsed)
	}
	if got := started.Load(); got > 4 {
		t.Errorf("po zrušení se rozběhlo %d úloh, chci jen ty rozdělané", got)
	}
	p.Close()
	waitNoLeak(t, before)
}

func TestCollectReturnsContextError(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	inputs := make([]int, 200)

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	_, _, err := pool.Collect(ctx, pool.Config[int, int]{
		Workers:   2,
		QueueSize: 1,
		Handler: func(ctx context.Context, v int) (int, error) {
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-time.After(20 * time.Millisecond):
				return v, nil
			}
		},
	}, inputs)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Collect po zrušení = %v, chci Canceled", err)
	}
	cancel()
	waitNoLeak(t, before)
}

func TestCollectWithAlreadyCanceledContext(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := pool.Collect(ctx, pool.Config[string, string]{
		Workers: 2,
		Handler: upper,
	}, []string{"a", "b", "c"})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Collect se zrušeným kontextem = %v, chci Canceled", err)
	}
	waitNoLeak(t, before)
}

func TestStatsDuringRun(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx := context.Background()
	p, err := pool.New(ctx, pool.Config[string, string]{Workers: 1, Handler: upper})
	if err != nil {
		t.Fatalf("New = %v, chci nil", err)
	}
	time.Sleep(2 * time.Millisecond)
	if stats := p.Stats(); stats.Elapsed <= 0 {
		t.Errorf("Stats.Elapsed během běhu = %v, chci kladnou dobu", stats.Elapsed)
	}
	p.Close()
	for range p.Results() { //nolint:revive
	}
	waitNoLeak(t, before)
}
