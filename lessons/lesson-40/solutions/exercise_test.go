package solutions_test

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-40/solutions"
)

// waitNoLeak počká, až počet goroutin klesne zpět na výchozí úroveň.
// Opakované krátké čekání je spolehlivější než jeden pevný sleep — doběhnutí
// goroutiny není okamžité.
func waitNoLeak(t *testing.T, before int) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if runtime.NumGoroutine() <= before {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("goroutine leak: před testem %d goroutin, po testu %d", before, runtime.NumGoroutine())
}

func TestParallelSquares(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"nil", nil, []int{}},
		{"empty", []int{}, []int{}},
		{"one element", []int{5}, []int{25}},
		{"negative numbers", []int{-3, 0, 4}, []int{9, 0, 16}},
		{"duplicity", []int{2, 2, 2}, []int{4, 4, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exercise.ParallelSquares(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("ParallelSquares(%v) vrátilo %d prvků, chci %d", tt.in, len(got), len(tt.want))
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("ParallelSquares(%v)[%d] = %d, chci %d", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParallelSquaresKeepsOrder(t *testing.T) {
	before := runtime.NumGoroutine()

	const n = 1000
	in := make([]int, n)
	for i := range in {
		in[i] = i
	}

	got := exercise.ParallelSquares(in)

	if len(got) != n {
		t.Fatalf("ParallelSquares vrátilo %d prvků, chci %d", len(got), n)
	}
	for i := 0; i < n; i++ {
		if got[i] != i*i {
			t.Fatalf("ParallelSquares()[%d] = %d, chci %d", i, got[i], i*i)
		}
	}

	waitNoLeak(t, before)
}

// TestParallelSquaresSpawnsGoroutines hlídá, že funkce práci opravdu rozdá
// goroutinám. Sekvenční smyčka by tímhle testem neprošla.
func TestParallelSquaresSpawnsGoroutines(t *testing.T) {
	const n = 2000
	in := make([]int, n)
	for i := range in {
		in[i] = i
	}

	base := runtime.NumGoroutine()
	var peak atomic.Int64
	var running atomic.Bool
	running.Store(true)
	watching := make(chan struct{})
	watcherDone := make(chan struct{})

	go func() {
		defer close(watcherDone)
		close(watching)
		for running.Load() {
			if cur := int64(runtime.NumGoroutine()); cur > peak.Load() {
				peak.Store(cur)
			}
			runtime.Gosched()
		}
	}()
	<-watching

	got := exercise.ParallelSquares(in)
	running.Store(false)
	<-watcherDone

	if len(got) != n {
		t.Fatalf("ParallelSquares vrátilo %d prvků, chci %d", len(got), n)
	}
	if extra := peak.Load() - int64(base); extra < 20 {
		t.Errorf("během ParallelSquares běželo navíc nejvýše %d goroutin — funkce nic nespouští", extra)
	}
}

func TestFanOutSum(t *testing.T) {
	tests := []struct {
		name    string
		nums    []int
		workers int
		want    int
	}{
		{"nil", nil, 4, 0},
		{"empty", []int{}, 4, 0},
		{"workers = 0", []int{1, 2, 3, 4}, 0, 10},
		{"workers = -5", []int{1, 2, 3, 4}, -5, 10},
		{"workers = 1", []int{1, 2, 3, 4}, 1, 10},
		{"workers = 3", []int{1, 2, 3, 4, 5, 6, 7}, 3, 28},
		{"more workers than items", []int{1, 2, 3}, 64, 6},
		{"negative numbers", []int{-5, 5, -10, 10}, 2, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.FanOutSum(tt.nums, tt.workers); got != tt.want {
				t.Errorf("FanOutSum(%v, %d) = %d, chci %d", tt.nums, tt.workers, got, tt.want)
			}
		})
	}
}

func TestFanOutSumLargeInput(t *testing.T) {
	before := runtime.NumGoroutine()

	const n = 100_000
	nums := make([]int, n)
	want := 0
	for i := range nums {
		nums[i] = i % 7
		want += nums[i]
	}

	for _, workers := range []int{1, 2, 8, 33} {
		if got := exercise.FanOutSum(nums, workers); got != want {
			t.Errorf("FanOutSum(100k prvků, %d) = %d, chci %d", workers, got, want)
		}
	}

	waitNoLeak(t, before)
}

// TestFanOutSumSpawnsGoroutines ověřuje, že se práce skutečně rozdělí mezi
// souběžně běžící goroutiny.
func TestFanOutSumSpawnsGoroutines(t *testing.T) {
	const n = 200_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = 1
	}

	base := runtime.NumGoroutine()
	var peak atomic.Int64
	var running atomic.Bool
	running.Store(true)
	watching := make(chan struct{})
	watcherDone := make(chan struct{})

	go func() {
		defer close(watcherDone)
		close(watching)
		for running.Load() {
			if cur := int64(runtime.NumGoroutine()); cur > peak.Load() {
				peak.Store(cur)
			}
			runtime.Gosched()
		}
	}()
	<-watching

	got := exercise.FanOutSum(nums, 500)
	running.Store(false)
	<-watcherDone

	if got != n {
		t.Fatalf("FanOutSum = %d, chci %d", got, n)
	}
	if extra := peak.Load() - int64(base); extra < 20 {
		t.Errorf("během FanOutSum běželo navíc nejvýše %d goroutin — práce se nerozdělila", extra)
	}
}

// stableGoroutines počká, až se počet goroutin ustálí — lokální helper místo
// studentova GoroutineDelta.
func stableGoroutines() int {
	const (
		needStable = 3
		maxRounds  = 300
		step       = 5 * time.Millisecond
	)
	runtime.Gosched()
	prev := runtime.NumGoroutine()
	stable := 0
	for i := 0; i < maxRounds; i++ {
		time.Sleep(step)
		cur := runtime.NumGoroutine()
		if cur == prev {
			stable++
			if stable >= needStable {
				return cur
			}
			continue
		}
		prev = cur
		stable = 0
	}
	return runtime.NumGoroutine()
}

func goroutineDelta(f func()) int {
	before := stableGoroutines()
	f()
	after := stableGoroutines()
	return after - before
}

func TestGoroutineDeltaCleanCode(t *testing.T) {
	if got := exercise.GoroutineDelta(func() {}); got != 0 {
		t.Errorf("GoroutineDelta(prázdná funkce) = %d, chci 0", got)
	}

	got := exercise.GoroutineDelta(func() {
		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
			}()
		}
		wg.Wait()
	})
	if got != 0 {
		t.Errorf("GoroutineDelta(funkce s WaitGroup) = %d, chci 0", got)
	}
}

func TestGoroutineDeltaCountsLeaks(t *testing.T) {
	release := make(chan struct{})
	defer close(release) // po testu goroutiny pustíme, ať neleakují doopravdy

	got := exercise.GoroutineDelta(func() {
		for i := 0; i < 3; i++ {
			go func() { <-release }()
		}
	})
	if got < 3 {
		t.Errorf("GoroutineDelta(3 zablokované goroutiny) = %d, chci alespoň 3", got)
	}
}

func TestLeakyGeneratorReallyLeaks(t *testing.T) {
	if got := goroutineDelta(exercise.LeakyGenerator); got < 1 {
		t.Errorf("LeakyGenerator: delta goroutin = %d, chci alespoň 1 — funkce má leakovat", got)
	}
}

func TestSafeGeneratorProducesSequence(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	ch := exercise.SafeGenerator(done)
	for want := 0; want < 10; want++ {
		select {
		case got := <-ch:
			if got != want {
				t.Fatalf("SafeGenerator poslal %d, chci %d", got, want)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("SafeGenerator neposlal hodnotu %d", want)
		}
	}
}

func TestSafeGeneratorClosesChannel(t *testing.T) {
	done := make(chan struct{})
	ch := exercise.SafeGenerator(done)
	<-ch
	close(done)

	deadline := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return // kanál je zavřený, přesně jak má být
			}
		case <-deadline:
			t.Fatal("SafeGenerator po zavření done nezavřel výstupní kanál")
		}
	}
}

func TestSafeGeneratorDoesNotLeak(t *testing.T) {
	got := goroutineDelta(func() {
		done := make(chan struct{})
		ch := exercise.SafeGenerator(done)
		for i := 0; i < 5; i++ {
			<-ch
		}
		close(done)
	})
	if got > 0 {
		t.Errorf("SafeGenerator: delta goroutin = %d, chci 0 — po zavření done nesmí nic zůstat", got)
	}
}

func TestSafeGeneratorSurvivesAbandonedReader(t *testing.T) {
	// Nejtvrdší případ: přečteme jednu hodnotu a pak z kanálu přestaneme
	// číst úplně. Goroutina uvízlá na `out <- i` by tady leakovala.
	got := goroutineDelta(func() {
		done := make(chan struct{})
		ch := exercise.SafeGenerator(done)
		<-ch
		close(done)
	})
	if got > 0 {
		t.Errorf("SafeGenerator: delta goroutin = %d, chci 0 — generátor uvízl na zápisu do opuštěného kanálu", got)
	}
}
