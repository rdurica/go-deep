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

	exercise "github.com/rdurica/go-deep/lessons/lesson-47/exercise"
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

// barrier pustí dál vždy až n goroutin najednou. Když implementace pouští
// míň, spadneme na timeout a test to pozná na naměřeném maximu souběhu.
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

func TestWithContextReturnsNonNil(t *testing.T) {
	g, ctx := exercise.WithContext(context.Background())
	if g == nil {
		t.Fatal("WithContext vrátil nil skupinu")
	}
	if ctx == nil {
		t.Fatal("WithContext vrátil nil kontext")
	}
}

func TestWithContextIsDerivedContext(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, ctx := exercise.WithContext(parent)
	if parent == ctx {
		t.Error("WithContext musí vrátit odvozený kontext, ne stejný jako rodič")
	}
}

func TestWithContextPropagatesParentCancel(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()

	_, ctx := exercise.WithContext(parent)
	cancelParent()

	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("odvozený kontext nereaguje na zrušení rodiče")
	}
}

func TestGroupRunsEverythingAndReturnsNil(t *testing.T) {
	before := runtime.NumGoroutine()

	var g exercise.Group // nulová hodnota musí být použitelná
	var done atomic.Int64
	for i := 0; i < 20; i++ {
		g.Go(func() error {
			done.Add(1)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		t.Errorf("Wait() = %v, chci nil", err)
	}
	if got := done.Load(); got != 20 {
		t.Errorf("proběhlo %d úloh, chci 20", got)
	}
	waitNoLeak(t, before)
}

func TestGroupSkipsNilFunc(t *testing.T) {
	before := runtime.NumGoroutine()

	var g exercise.Group
	g.Go(nil)
	g.Go(func() error { return nil })
	if err := g.Wait(); err != nil {
		t.Errorf("Wait() = %v, chci nil", err)
	}
	waitNoLeak(t, before)
}

func TestGroupReturnsFirstError(t *testing.T) {
	before := runtime.NumGoroutine()
	errBoom := errors.New("boom")

	var g exercise.Group
	release := make(chan struct{})
	var finished atomic.Int64

	// Devět úloh čeká na uvolnění, desátá selže. Pořadí je tím dané:
	// jediná chyba, kterou skupina může vidět, je errBoom.
	for i := 0; i < 9; i++ {
		g.Go(func() error {
			<-release
			finished.Add(1)
			return nil
		})
	}
	g.Go(func() error {
		close(release)
		return errBoom
	})

	err := g.Wait()
	if !errors.Is(err, errBoom) {
		t.Errorf("Wait() = %v, chci %v", err, errBoom)
	}
	if got := finished.Load(); got != 9 {
		t.Errorf("Wait se vrátil, ale dokončeno je jen %d z 9 úloh", got)
	}
	waitNoLeak(t, before)
}

func TestGroupWithContextCancelsOnFirstError(t *testing.T) {
	before := runtime.NumGoroutine()
	errBoom := errors.New("boom")

	g, ctx := exercise.WithContext(context.Background())
	var canceled atomic.Int64

	for i := 0; i < 5; i++ {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				canceled.Add(1)
				return nil
			case <-time.After(3 * time.Second):
				return errors.New("úloha nedostala zrušení")
			}
		})
	}
	g.Go(func() error { return errBoom })

	start := time.Now()
	err := g.Wait()
	elapsed := time.Since(start)

	if !errors.Is(err, errBoom) {
		t.Errorf("Wait() = %v, chci %v", err, errBoom)
	}
	if got := canceled.Load(); got != 5 {
		t.Errorf("zrušení dostalo %d úloh z 5", got)
	}
	if elapsed > 2*time.Second {
		t.Errorf("Wait trval %v, chci rychlé zrušení", elapsed)
	}
	if ctx.Err() == nil {
		t.Error("odvozený kontext měl být po Wait zrušený")
	}
	waitNoLeak(t, before)
}

func TestGroupWithContextCancelsAfterSuccess(t *testing.T) {
	before := runtime.NumGoroutine()

	g, ctx := exercise.WithContext(context.Background())
	g.Go(func() error { return nil })
	if err := g.Wait(); err != nil {
		t.Fatalf("Wait() = %v, chci nil", err)
	}
	// Wait musí kontext zrušit i v úspěšné větvi, jinak unikne.
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("ctx.Err() = %v, chci Canceled", ctx.Err())
	}
	waitNoLeak(t, before)
}

func TestGroupSetLimit(t *testing.T) {
	before := runtime.NumGoroutine()
	const limit = 3

	var g exercise.Group
	g.SetLimit(limit)

	var tr tracker
	b := newBarrier(limit)
	for i := 0; i < 12; i++ { // násobek limitu, ať bariéra vždy dorazí do plného počtu
		g.Go(func() error {
			tr.enter()
			b.wait()
			tr.leave()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatalf("Wait() = %v, chci nil", err)
	}
	if peak := tr.peak(); peak != limit {
		t.Errorf("maximální souběh = %d, chci přesně %d", peak, limit)
	}
	waitNoLeak(t, before)
}

func TestGroupSetLimitZeroMeansUnlimited(t *testing.T) {
	before := runtime.NumGoroutine()

	var g exercise.Group
	g.SetLimit(0)
	var done atomic.Int64
	for i := 0; i < 5; i++ {
		g.Go(func() error {
			done.Add(1)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Errorf("Wait() = %v, chci nil", err)
	}
	if done.Load() != 5 {
		t.Errorf("proběhlo %d úloh, chci 5", done.Load())
	}
	waitNoLeak(t, before)
}

func TestGroupSetLimitAfterGoPanics(t *testing.T) {
	var g exercise.Group
	g.Go(func() error { return nil })
	defer func() {
		if recover() == nil {
			t.Error("SetLimit po prvním Go měl panikovat")
		}
		_ = g.Wait()
	}()
	g.SetLimit(2)
}

func TestGroupWithContextRespectsParentCancel(t *testing.T) {
	before := runtime.NumGoroutine()

	parent, cancelParent := context.WithCancel(context.Background())
	g, ctx := exercise.WithContext(parent)
	g.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
			return errors.New("zrušení rodiče se nepropsalo")
		}
	})
	time.AfterFunc(20*time.Millisecond, cancelParent)

	if err := g.Wait(); !errors.Is(err, context.Canceled) {
		t.Errorf("Wait() = %v, chci Canceled", err)
	}
	cancelParent()
	waitNoLeak(t, before)
}

func TestRunAllSuccess(t *testing.T) {
	before := runtime.NumGoroutine()

	var done atomic.Int64
	tasks := []exercise.Task{
		{Name: "a", Run: func(context.Context) error { done.Add(1); return nil }},
		{Name: "b", Run: func(context.Context) error { done.Add(1); return nil }},
	}
	if err := exercise.RunAll(context.Background(), tasks); err != nil {
		t.Errorf("RunAll = %v, chci nil", err)
	}
	if done.Load() != 2 {
		t.Errorf("proběhlo %d úloh, chci 2", done.Load())
	}
	if err := exercise.RunAll(context.Background(), nil); err != nil {
		t.Errorf("RunAll(nil) = %v, chci nil", err)
	}
	waitNoLeak(t, before)
}

func TestRunAllNilRun(t *testing.T) {
	err := exercise.RunAll(context.Background(), []exercise.Task{{Name: "prazdna"}})
	if !errors.Is(err, exercise.ErrNilTask) {
		t.Errorf("RunAll s nil Run = %v, chci ErrNilTask", err)
		return
	}
	if err != nil && !strings.Contains(err.Error(), "prazdna") {
		t.Errorf("chyba %q neobsahuje jméno úlohy", err.Error())
	}
}

func TestRunAllDetachedSurvivesParentCancel(t *testing.T) {
	before := runtime.NumGoroutine()

	var detachedDone, normalDone atomic.Int64
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()

	tasks := []exercise.Task{
		{
			Name: "normalni",
			Run: func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(3 * time.Second):
					normalDone.Add(1)
					return nil
				}
			},
		},
		{
			Name:     "odpojena",
			Detached: true,
			Run: func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(150 * time.Millisecond):
					detachedDone.Add(1)
					return nil
				}
			},
		},
	}

	time.AfterFunc(20*time.Millisecond, cancel)
	start := time.Now()
	err := exercise.RunAll(parent, tasks)
	elapsed := time.Since(start)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("RunAll = %v, chci chybu obalující context.Canceled", err)
	}
	if err != nil && !strings.Contains(err.Error(), "normalni") {
		t.Errorf("chyba %q neobsahuje jméno úlohy, která selhala", err.Error())
	}
	if detachedDone.Load() != 1 {
		t.Error("odpojená úloha nedoběhla, přestože měla přežít zrušení rodiče")
	}
	if normalDone.Load() != 0 {
		t.Error("běžná úloha doběhla, přestože měla dostat zrušení")
	}
	if elapsed < 100*time.Millisecond {
		t.Errorf("RunAll se vrátil za %v — nepočkal na odpojenou úlohu", elapsed)
	}
	if elapsed > 2*time.Second {
		t.Errorf("RunAll trval %v, chci rychlý návrat po dokončení odpojené úlohy", elapsed)
	}
	waitNoLeak(t, before)
}

func TestRunAllDetachedIgnoresParentValuesLoss(t *testing.T) {
	type key struct{}

	var got atomic.Value
	parent, cancel := context.WithCancel(context.WithValue(context.Background(), key{}, "trace-123"))
	defer cancel()

	tasks := []exercise.Task{{
		Name:     "odpojena",
		Detached: true,
		Run: func(ctx context.Context) error {
			if v, ok := ctx.Value(key{}).(string); ok {
				got.Store(v)
			}
			return nil
		},
	}}
	if err := exercise.RunAll(parent, tasks); err != nil {
		t.Fatalf("RunAll = %v, chci nil", err)
	}
	// WithoutCancel zahazuje zrušení, ale hodnoty musí zůstat.
	if v, _ := got.Load().(string); v != "trace-123" {
		t.Errorf("odpojená úloha viděla hodnotu %q, chci %q", v, "trace-123")
	}
}

func TestCause(t *testing.T) {
	base := errors.New("root")
	tests := []struct {
		name string
		in   error
		want error
	}{
		{"nil", nil, nil},
		{"bez obalu", base, base},
		{"jeden obal", wrap("a", base), base},
		{"three wrappers", wrap("c", wrap("b", wrap("a", base))), base},
		{"context", wrap("task", context.Canceled), context.Canceled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.Cause(tt.in); got != tt.want {
				t.Errorf("Cause(%v) = %v, chci %v", tt.in, got, tt.want)
			}
		})
	}
}

func wrap(msg string, err error) error {
	return &wrapped{msg: msg, err: err}
}

type wrapped struct {
	msg string
	err error
}

func (w *wrapped) Error() string { return w.msg + ": " + w.err.Error() }
func (w *wrapped) Unwrap() error { return w.err }
