package ratelimit_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p04-worker-pool/ratelimit"
)

// fakeClock je posuvný čas — testy díky němu nemusí nic prospat.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func TestNewValidation(t *testing.T) {
	if _, err := ratelimit.New(0, 1, nil); !errors.Is(err, ratelimit.ErrInvalidRate) {
		t.Errorf("New(rate=0) = %v, chci ErrInvalidRate", err)
	}
	if _, err := ratelimit.New(-1, 1, nil); !errors.Is(err, ratelimit.ErrInvalidRate) {
		t.Errorf("New(rate=-1) = %v, chci ErrInvalidRate", err)
	}
	if _, err := ratelimit.New(1, 0, nil); !errors.Is(err, ratelimit.ErrInvalidBurst) {
		t.Errorf("New(burst=0) = %v, chci ErrInvalidBurst", err)
	}
	if _, err := ratelimit.New(1, 1, nil); err != nil {
		t.Errorf("New(platná konfigurace) = %v, chci nil", err)
	}
}

func TestBucketStartsFull(t *testing.T) {
	clock := newFakeClock()
	b, err := ratelimit.New(10, 3, clock.Now)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	for i := 0; i < 3; i++ {
		if !b.Allow() {
			t.Fatalf("Allow() č. %d = false, chci true (bucket začíná plný)", i+1)
		}
	}
	if b.Allow() {
		t.Error("Allow() po vyčerpání = true, chci false")
	}
}

func TestBucketRefillsOverTime(t *testing.T) {
	clock := newFakeClock()
	b, err := ratelimit.New(10, 2, clock.Now) // 10 tokenů/s = 1 token za 100 ms
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	if !b.Allow() || !b.Allow() {
		t.Fatal("první dvě Allow() mají projít")
	}
	if b.Allow() {
		t.Fatal("třetí Allow() má selhat")
	}

	clock.Advance(99 * time.Millisecond)
	if b.Allow() {
		t.Error("Allow() po 99 ms = true, chci false (token ještě není)")
	}

	clock.Advance(time.Millisecond)
	if !b.Allow() {
		t.Error("Allow() po 100 ms = false, chci true")
	}

	// Doplnění se nesmí přelít přes kapacitu.
	clock.Advance(time.Hour)
	if !b.Allow() || !b.Allow() {
		t.Error("po dlouhé pauze mají projít dvě Allow()")
	}
	if b.Allow() {
		t.Error("po dlouhé pauze prošlo víc než burst tokenů — bucket přetéká")
	}
}

func TestBucketAllowN(t *testing.T) {
	clock := newFakeClock()
	b, err := ratelimit.New(5, 5, clock.Now)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	if !b.AllowN(0) {
		t.Error("AllowN(0) = false, chci true")
	}
	if !b.AllowN(5) {
		t.Error("AllowN(5) = false, chci true (bucket je plný)")
	}
	if b.AllowN(1) {
		t.Error("AllowN(1) po vyčerpání = true, chci false")
	}

	clock.Advance(time.Second)
	if b.AllowN(6) {
		t.Error("AllowN(6) = true, chci false (víc než kapacita)")
	}
}

func TestBucketReserve(t *testing.T) {
	clock := newFakeClock()
	b, err := ratelimit.New(10, 1, clock.Now) // 1 token za 100 ms
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	if d := b.Reserve(); d != 0 {
		t.Errorf("první Reserve() = %v, chci 0", d)
	}

	// Rezervace se řadí za sebe, ne na stejný okamžik.
	first := b.Reserve()
	second := b.Reserve()
	if first <= 0 {
		t.Errorf("druhý Reserve() = %v, chci kladné čekání", first)
	}
	if second <= first {
		t.Errorf("třetí Reserve() = %v, chci víc než %v", second, first)
	}

	tolerance := 5 * time.Millisecond
	if diff := second - first - 100*time.Millisecond; diff > tolerance || diff < -tolerance {
		t.Errorf("rozestup rezervací = %v, chci ~100 ms", second-first)
	}
}

func TestBucketWaitDoesNotBlockWhenFull(t *testing.T) {
	clock := newFakeClock()
	b, err := ratelimit.New(1, 3, clock.Now)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	start := time.Now()
	for i := 0; i < 3; i++ {
		if err := b.Wait(context.Background()); err != nil {
			t.Fatalf("Wait() = %v, chci nil", err)
		}
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("tři Wait() na plném bucketu trvaly %v, chci téměř nic", elapsed)
	}
}

func TestBucketWaitRespectsContext(t *testing.T) {
	clock := newFakeClock()
	b, err := ratelimit.New(0.001, 1, clock.Now) // jeden token za 1000 s
	if err != nil {
		t.Fatalf("New() = %v", err)
	}
	if err := b.Wait(context.Background()); err != nil {
		t.Fatalf("první Wait() = %v, chci nil", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	start := time.Now()
	if err := b.Wait(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Wait() = %v, chci context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("Wait() trval %v — nečeká přes select s ctx.Done()", elapsed)
	}

	canceled, cancel2 := context.WithCancel(context.Background())
	cancel2()
	if err := b.Wait(canceled); !errors.Is(err, context.Canceled) {
		t.Errorf("Wait(zrušený ctx) = %v, chci context.Canceled", err)
	}
}

func TestBucketConcurrent(t *testing.T) {
	clock := newFakeClock()
	const burst = 100
	b, err := ratelimit.New(1, burst, clock.Now)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	var (
		allowed atomic.Int64
		wg      sync.WaitGroup
		start   = make(chan struct{})
	)
	const goroutines = 50
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 10; j++ {
				if b.Allow() {
					allowed.Add(1)
				}
			}
		}()
	}
	close(start)
	wg.Wait()

	// Čas stojí, takže se nedoplnil ani jeden token navíc:
	// projít smí přesně burst požadavků, ani o jeden víc.
	if got := allowed.Load(); got != burst {
		t.Errorf("prošlo %d požadavků, chci přesně %d", got, burst)
	}
}
