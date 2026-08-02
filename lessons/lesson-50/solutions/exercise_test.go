package solutions_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-50/solutions"
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

// barrier pustí dál vždy až n goroutin najednou, takže je naměřený souběh
// deterministický. Když se jich sejde méně, spadneme na timeout a test to
// pozná na metrice MaxInFlight.
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

func items(n int) []exercise.Item {
	out := make([]exercise.Item, n)
	for i := range out {
		out[i] = exercise.Item{ID: 100 + i, Payload: fmt.Sprintf("payload-%02d", i)}
	}
	return out
}

func upperHandler(_ context.Context, it exercise.Item) (string, error) {
	return strings.ToUpper(it.Payload), nil
}

func TestStatsTotalAndThroughput(t *testing.T) {
	tests := []struct {
		name  string
		stats exercise.Stats
		total int
		tput  float64
	}{
		{"empty", exercise.Stats{}, 0, 0},
		{"no runtime", exercise.Stats{Processed: 10, Elapsed: 0}, 10, 0},
		{"negative duration", exercise.Stats{Processed: 10, Elapsed: -time.Second}, 10, 0},
		{"hundred jobs in two seconds", exercise.Stats{Processed: 50, Failed: 50, Elapsed: 2 * time.Second}, 100, 50},
		{"jen chyby", exercise.Stats{Failed: 8, Elapsed: 4 * time.Second}, 8, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.stats.Total(); got != tt.total {
				t.Errorf("Total() = %d, chci %d", got, tt.total)
			}
			if got := tt.stats.Throughput(); math.Abs(got-tt.tput) > 1e-9 {
				t.Errorf("Throughput() = %v, chci %v", got, tt.tput)
			}
		})
	}
}

func TestMetricsUnderConcurrency(t *testing.T) {
	before := runtime.NumGoroutine()
	const workers = 8

	var m exercise.Metrics
	b := newBarrier(workers)
	var wg sync.WaitGroup

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			m.Enter()
			b.wait() // všech osm se sejde naráz
			if i%2 == 0 {
				m.Leave(nil)
			} else {
				m.Leave(errors.New("bum"))
			}
		}()
	}
	wg.Wait()

	got := m.Snapshot(2 * time.Second)
	if got.Processed != workers/2 {
		t.Errorf("Processed = %d, chci %d", got.Processed, workers/2)
	}
	if got.Failed != workers/2 {
		t.Errorf("Failed = %d, chci %d", got.Failed, workers/2)
	}
	if got.MaxInFlight != workers {
		t.Errorf("MaxInFlight = %d, chci %d", got.MaxInFlight, workers)
	}
	if got.Elapsed != 2*time.Second {
		t.Errorf("Elapsed = %v, chci 2s", got.Elapsed)
	}
	wantTput := float64(got.Processed+got.Failed) / got.Elapsed.Seconds()
	if math.Abs(wantTput-float64(workers)/2) > 1e-9 {
		t.Errorf("průchodnost z polí = %v, chci %v", wantTput, float64(workers)/2)
	}
	waitNoLeak(t, before)
}

func TestProcessKeepsOrder(t *testing.T) {
	before := runtime.NumGoroutine()
	in := items(40)

	got, stats, err := exercise.Process(context.Background(), exercise.Config{
		Workers:   4,
		QueueSize: 8,
		Handler:   upperHandler,
	}, in)
	if err != nil {
		t.Fatalf("Process = %v, chci nil", err)
	}

	if len(got) != len(in) {
		t.Fatalf("Process vrátil %d výsledků, chci %d", len(got), len(in))
	}
	for i, item := range in {
		if got[i].ID != item.ID {
			t.Errorf("výsledek[%d].ID = %d, chci %d", i, got[i].ID, item.ID)
		}
		if want := strings.ToUpper(item.Payload); got[i].Value != want {
			t.Errorf("výsledek[%d].Value = %q, chci %q", i, got[i].Value, want)
		}
		if got[i].Err != nil {
			t.Errorf("výsledek[%d].Err = %v, chci nil", i, got[i].Err)
		}
	}
	if stats.Processed != len(in) || stats.Failed != 0 {
		t.Errorf("Stats = %+v, chci processed=%d a failed=0", stats, len(in))
	}
	if stats.Elapsed <= 0 {
		t.Errorf("Elapsed = %v, chci kladnou dobu běhu", stats.Elapsed)
	}
	if tput := float64(stats.Processed+stats.Failed) / stats.Elapsed.Seconds(); tput <= 0 {
		t.Errorf("průchodnost z polí = %v, chci kladnou", tput)
	}
	waitNoLeak(t, before)
}

func TestProcessRespectsWorkerLimit(t *testing.T) {
	before := runtime.NumGoroutine()
	const workers = 4

	b := newBarrier(workers)
	in := items(20) // násobek workers, ať bariéra vždy dorazí do plného počtu

	_, stats, err := exercise.Process(context.Background(), exercise.Config{
		Workers:   workers,
		QueueSize: 2,
		Handler: func(_ context.Context, it exercise.Item) (string, error) {
			b.wait()
			return it.Payload, nil
		},
	}, in)
	if err != nil {
		t.Fatalf("Process = %v, chci nil", err)
	}

	if stats.MaxInFlight > workers {
		t.Errorf("MaxInFlight = %d, nesmí být víc než %d", stats.MaxInFlight, workers)
	}
	if stats.MaxInFlight < workers {
		t.Errorf("MaxInFlight = %d, chci %d — nevyužíváš všechny workery", stats.MaxInFlight, workers)
	}
	waitNoLeak(t, before)
}

func TestProcessCollectsItemErrors(t *testing.T) {
	before := runtime.NumGoroutine()
	errOdd := errors.New("liché ID")
	in := items(10)

	got, stats, err := exercise.Process(context.Background(), exercise.Config{
		Workers: 3,
		Handler: func(_ context.Context, it exercise.Item) (string, error) {
			if it.ID%2 == 1 {
				return "", fmt.Errorf("úloha %d: %w", it.ID, errOdd)
			}
			return it.Payload, nil
		},
	}, in)

	// Bez FailFast není chyba úlohy chybou dávky.
	if err != nil {
		t.Fatalf("Process = %v, chci nil", err)
	}
	for i, res := range got {
		wantErr := in[i].ID%2 == 1
		if wantErr && !errors.Is(res.Err, errOdd) {
			t.Errorf("výsledek[%d].Err = %v, chci %v", i, res.Err, errOdd)
		}
		if !wantErr && res.Err != nil {
			t.Errorf("výsledek[%d].Err = %v, chci nil", i, res.Err)
		}
	}
	if stats.Failed != 5 || stats.Processed != 5 {
		t.Errorf("Stats = %+v, chci processed=5 a failed=5", stats)
	}
	waitNoLeak(t, before)
}

func TestProcessFailFast(t *testing.T) {
	before := runtime.NumGoroutine()
	errBoom := errors.New("bum")
	const workers = 4
	in := items(12)

	var canceled atomic.Int64
	// Bariéra zaručí, že se první čtyři úlohy opravdu rozběhnou dřív, než ta
	// první ohlásí chybu. Bez toho by test byl nedeterministický: workeři by
	// mohli chybu zaznamenat ještě předtím, než si vůbec vezmou úlohu.
	b := newBarrier(workers)

	start := time.Now()
	got, stats, err := exercise.Process(context.Background(), exercise.Config{
		Workers:   workers,
		QueueSize: 12,
		FailFast:  true,
		Handler: func(ctx context.Context, it exercise.Item) (string, error) {
			b.wait()
			if it.ID == in[0].ID {
				return "", errBoom
			}
			select {
			case <-ctx.Done():
				canceled.Add(1)
				return "", ctx.Err()
			case <-time.After(3 * time.Second):
				return it.Payload, nil
			}
		},
	}, in)
	elapsed := time.Since(start)

	if !errors.Is(err, errBoom) {
		t.Errorf("Process = %v, chci první chybu %v", err, errBoom)
	}
	if got != nil {
		t.Errorf("Process vrátil %v, v režimu FailFast chci nil", got)
	}
	if got := canceled.Load(); got < workers-1 {
		t.Errorf("zrušení dostalo %d rozběhnutých úloh, chci aspoň %d", got, workers-1)
	}
	if elapsed > 2*time.Second {
		t.Errorf("Process trval %v, chci rychlé ukončení po první chybě", elapsed)
	}
	if stats.Failed == 0 {
		t.Errorf("Stats = %+v, chci nenulové failed", stats)
	}
	waitNoLeak(t, before)
}

func TestProcessValidationAndEdgeCases(t *testing.T) {
	before := runtime.NumGoroutine()
	ctx := context.Background()

	tests := []struct {
		name string
		cfg  exercise.Config
		want error
	}{
		{"no workers", exercise.Config{Handler: upperHandler}, exercise.ErrInvalidWorkers},
		{"negative queue", exercise.Config{Workers: 2, QueueSize: -1, Handler: upperHandler}, exercise.ErrInvalidQueueSize},
		{"bez handleru", exercise.Config{Workers: 2}, exercise.ErrNilHandler},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := exercise.Process(ctx, tt.cfg, items(3)); !errors.Is(err, tt.want) {
				t.Errorf("Process = %v, chci %v", err, tt.want)
			}
		})
	}

	t.Run("empty batch", func(t *testing.T) {
		got, stats, err := exercise.Process(ctx, exercise.Config{Workers: 2, Handler: upperHandler}, nil)
		if err != nil {
			t.Fatalf("Process(nil) = %v, chci nil", err)
		}
		if len(got) != 0 {
			t.Errorf("Process(nil) = %v, chci prázdný výsledek", got)
		}
		if stats.Processed+stats.Failed != 0 {
			t.Errorf("Stats = %+v, chci nuly", stats)
		}
	})

	t.Run("already canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, err := exercise.Process(canceledCtx, exercise.Config{Workers: 2, Handler: upperHandler}, items(5)); !errors.Is(err, context.Canceled) {
			t.Errorf("Process se zrušeným kontextem = %v, chci Canceled", err)
		}
	})

	waitNoLeak(t, before)
}

func TestProcessCancelDuringRun(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	time.AfterFunc(30*time.Millisecond, cancel)

	start := time.Now()
	_, _, err := exercise.Process(ctx, exercise.Config{
		Workers:   2,
		QueueSize: 1,
		Handler: func(ctx context.Context, it exercise.Item) (string, error) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(3 * time.Second):
				return it.Payload, nil
			}
		},
	}, items(100))
	elapsed := time.Since(start)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Process po zrušení = %v, chci Canceled", err)
	}
	if elapsed > 2*time.Second {
		t.Errorf("Process se zastavoval %v, chci rychlé ukončení", elapsed)
	}
	waitNoLeak(t, before)
}

func TestProcessStreamProcessesEverything(t *testing.T) {
	before := runtime.NumGoroutine()
	in := items(40)

	src := make(chan exercise.Item)
	out := make(chan exercise.Outcome, len(in))

	go func() {
		defer close(src)
		for _, it := range in {
			src <- it
		}
	}()

	stats, err := exercise.ProcessStream(context.Background(), exercise.Config{
		Workers:   4,
		QueueSize: 4,
		Handler:   upperHandler,
	}, src, out)
	if err != nil {
		t.Fatalf("ProcessStream = %v, chci nil", err)
	}

	seen := make(map[int]string)
	for res := range out { // kanál musí být zavřený, jinak tady visíme
		if res.Err != nil {
			t.Errorf("úloha %d skončila chybou %v", res.ID, res.Err)
		}
		seen[res.ID] = res.Value
	}
	if len(seen) != len(in) {
		t.Fatalf("dorazilo %d výsledků, chci %d", len(seen), len(in))
	}
	for _, item := range in {
		if want := strings.ToUpper(item.Payload); seen[item.ID] != want {
			t.Errorf("výsledek úlohy %d = %q, chci %q", item.ID, seen[item.ID], want)
		}
	}
	if stats.Processed != len(in) || stats.MaxInFlight > 4 {
		t.Errorf("Stats = %+v, chci processed=%d a max_soubeznost <= 4", stats, len(in))
	}
	waitNoLeak(t, before)
}

// TestProcessStreamBackpressure ověřuje, že se tlak z plné fronty přenese až
// na odesílatele do vstupního kanálu.
func TestProcessStreamBackpressure(t *testing.T) {
	before := runtime.NumGoroutine()

	release := make(chan struct{})
	busy := make(chan struct{}, 8) // dost místa, ať hlášení "pracuji" nikdy neblokuje

	src := make(chan exercise.Item)
	out := make(chan exercise.Outcome, 8)
	done := make(chan struct{})

	go func() {
		defer close(done)
		// Jeden worker, fronta bez bufferu: v systému se udrží nejvýš
		// jedna rozdělaná úloha a jedna čekající u podavače.
		_, _ = exercise.ProcessStream(context.Background(), exercise.Config{
			Workers:   1,
			QueueSize: 0,
			Handler: func(_ context.Context, it exercise.Item) (string, error) {
				busy <- struct{}{}
				<-release
				return it.Payload, nil
			},
		}, src, out)
	}()

	src <- exercise.Item{ID: 1}
	select {
	case <-busy: // worker opravdu pracuje
	case <-time.After(2 * time.Second):
		t.Fatal("worker nezačal zpracovávat úlohu — ProcessStream musí spustit handler")
	}
	src <- exercise.Item{ID: 2}

	// Třetí odeslání nemá kam jít, musí blokovat.
	select {
	case src <- exercise.Item{ID: 3}:
		t.Error("třetí odeslání prošlo, backpressure nefunguje")
	case <-time.After(200 * time.Millisecond):
	}

	close(release)
	// Dopustíme zbytek: podavač i worker teď mají volno.
	go func() {
		src <- exercise.Item{ID: 3}
		close(src)
	}()

	var n int
	for range out {
		n++
	}
	<-done

	if n != 3 {
		t.Errorf("dorazilo %d výsledků, chci 3", n)
	}
	waitNoLeak(t, before)
}

func TestProcessStreamFailFastAndCancel(t *testing.T) {
	before := runtime.NumGoroutine()
	errBoom := errors.New("bum")

	t.Run("fail fast", func(t *testing.T) {
		src := make(chan exercise.Item, 20)
		for i := 0; i < 20; i++ {
			src <- exercise.Item{ID: i}
		}
		close(src)
		out := make(chan exercise.Outcome, 20)

		stats, err := exercise.ProcessStream(context.Background(), exercise.Config{
			Workers:   2,
			QueueSize: 2,
			FailFast:  true,
			Handler: func(_ context.Context, it exercise.Item) (string, error) {
				if it.ID == 0 {
					return "", errBoom
				}
				return "ok", nil
			},
		}, src, out)

		if !errors.Is(err, errBoom) {
			t.Errorf("ProcessStream = %v, chci %v", err, errBoom)
		}
		if stats.Processed+stats.Failed == 0 {
			t.Errorf("Stats = %+v, chci nenulový počet dokončených úloh", stats)
		}
		for range out { //nolint:revive // kanál musí být zavřený
		}
	})

	t.Run("context cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		time.AfterFunc(30*time.Millisecond, cancel)

		src := make(chan exercise.Item)
		go func() {
			defer close(src)
			for i := 0; ; i++ {
				select {
				case src <- exercise.Item{ID: i}:
				case <-ctx.Done():
					return
				}
			}
		}()
		out := make(chan exercise.Outcome, 64)
		drained := make(chan struct{})
		go func() {
			defer close(drained)
			for range out { //nolint:revive
			}
		}()

		start := time.Now()
		_, err := exercise.ProcessStream(ctx, exercise.Config{
			Workers:   2,
			QueueSize: 2,
			Handler: func(ctx context.Context, it exercise.Item) (string, error) {
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(20 * time.Millisecond):
					return "ok", nil
				}
			},
		}, src, out)
		elapsed := time.Since(start)
		<-drained

		if !errors.Is(err, context.Canceled) {
			t.Errorf("ProcessStream po zrušení = %v, chci Canceled", err)
		}
		if elapsed > 2*time.Second {
			t.Errorf("ProcessStream se zastavoval %v, chci rychlé ukončení", elapsed)
		}
	})

	t.Run("bad config closes output", func(t *testing.T) {
		out := make(chan exercise.Outcome)
		src := make(chan exercise.Item)
		close(src)

		_, err := exercise.ProcessStream(context.Background(), exercise.Config{Workers: 0}, src, out)
		if !errors.Is(err, exercise.ErrInvalidWorkers) {
			t.Errorf("ProcessStream = %v, chci ErrInvalidWorkers", err)
		}
		select {
		case _, ok := <-out:
			if ok {
				t.Error("čekal jsem zavřený výstupní kanál")
			}
		case <-time.After(time.Second):
			t.Error("výstupní kanál se nezavřel")
		}
	})

	waitNoLeak(t, before)
}
