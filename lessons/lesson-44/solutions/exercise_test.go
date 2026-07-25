package solutions_test

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-44/solutions"
)

func TestSafeIncrement(t *testing.T) {
	for _, n := range []int{0, 1, 100, 5000} {
		if got := exercise.SafeIncrement(n); got != n {
			t.Errorf("SafeIncrement(%d) = %d, chci %d — čítač ztrácí zvýšení", n, got, n)
		}
	}
}

func TestSafeIncrementRepeated(t *testing.T) {
	// Závod se nemusí projevit napoprvé. Opakování je nejlevnější způsob,
	// jak mu dát šanci.
	for i := 0; i < 20; i++ {
		if got := exercise.SafeIncrement(1000); got != 1000 {
			t.Fatalf("SafeIncrement(1000) = %d v pokusu %d, chci 1000", got, i)
		}
	}
}

func TestRegistryZeroValueAndLazyInit(t *testing.T) {
	var r exercise.Registry
	if got := r.Len(); got != 0 {
		t.Errorf("Len() na zero value = %d, chci 0", got)
	}
	if _, ok := r.Get("chybí"); ok {
		t.Error("Get na prázdném registru vrátil ok = true")
	}
	r.Set("a", 1)
	if got, ok := r.Get("a"); !ok || got != 1 {
		t.Errorf("Get(\"a\") = (%d, %v), chci (1, true)", got, ok)
	}
	if got := r.Len(); got != 1 {
		t.Errorf("Len() = %d, chci 1", got)
	}
}

func TestRegistryConcurrentWrites(t *testing.T) {
	var r exercise.Registry
	const goroutines, perGoroutine = 40, 25

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				r.Set(fmt.Sprintf("klíč-%d-%d", i, j), i*j)
			}
		}()
	}
	wg.Wait()

	if want := goroutines * perGoroutine; r.Len() != want {
		t.Errorf("Len() = %d, chci %d — souběžné zápisy se ztrácejí", r.Len(), want)
	}
}

func TestRegistryConcurrentReadWrite(t *testing.T) {
	var r exercise.Registry
	r.Set("start", 0)

	var wg sync.WaitGroup
	wg.Add(20)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				r.Set(fmt.Sprintf("k-%d", j%20), j)
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				r.Get(fmt.Sprintf("k-%d", j%20))
				r.Len()
			}
		}()
	}
	wg.Wait()
}

func TestParallelAppend(t *testing.T) {
	const n = 1000
	nums := make([]int, n)
	want := make([]int, n)
	for i := range nums {
		nums[i] = i
		want[i] = i * i
	}

	got := exercise.ParallelAppend(nums)
	if len(got) != n {
		t.Fatalf("ParallelAppend vrátilo %d prvků, chci %d — souběžný append ztrácí zápisy", len(got), n)
	}
	sort.Ints(got)
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ParallelAppend()[%d] = %d, chci %d", i, got[i], want[i])
		}
	}
}

func TestParallelAppendEmpty(t *testing.T) {
	if got := exercise.ParallelAppend(nil); len(got) != 0 {
		t.Errorf("ParallelAppend(nil) = %v, chci prázdný výsledek", got)
	}
}

func TestStressTestRunsEverything(t *testing.T) {
	var calls atomic.Int64
	exercise.StressTest(t, func() { calls.Add(1) })

	want := int64(exercise.StressGoroutines * exercise.StressIterations)
	if got := calls.Load(); got != want {
		t.Errorf("StressTest zavolal f %dx, chci %d", got, want)
	}
}

func TestStressTestRunsConcurrently(t *testing.T) {
	base := int64(runtime.NumGoroutine())
	var peak atomic.Int64
	exercise.StressTest(t, func() {
		cur := int64(runtime.NumGoroutine())
		for {
			old := peak.Load()
			if cur <= old || peak.CompareAndSwap(old, cur) {
				break
			}
		}
	})

	want := int64(exercise.StressGoroutines / 2)
	if extra := peak.Load() - base; extra < want {
		t.Errorf("během StressTestu žilo navíc nejvýše %d goroutin, chci alespoň %d — f se volá sekvenčně", extra, want)
	}
}

func TestStressTestFindsRegistryRace(t *testing.T) {
	// Přesně tenhle tvar testu odhalí závod, který jednorázové volání
	// nikdy nevyprovokuje.
	var r exercise.Registry
	var i atomic.Int64
	exercise.StressTest(t, func() {
		n := i.Add(1)
		r.Set(fmt.Sprintf("k-%d", n%50), int(n))
		r.Get("k-0")
		r.Len()
	})

	if got := r.Len(); got != 50 {
		t.Errorf("Len() = %d, chci 50", got)
	}
}

func TestStressTestFindsCounterRace(t *testing.T) {
	var c atomic.Int64
	exercise.StressTest(t, func() { c.Add(1) })

	want := int64(exercise.StressGoroutines * exercise.StressIterations)
	if got := c.Load(); got != want {
		t.Errorf("čítač = %d, chci %d", got, want)
	}
}

func TestConfigZeroValueIsConsistent(t *testing.T) {
	var c exercise.Config
	if got := c.Load(); !got.Consistent() {
		t.Errorf("Load() na zero value = %+v, chci konzistentní snapshot", got)
	}
}

func TestConfigStoreAndLoad(t *testing.T) {
	var c exercise.Config
	want := exercise.NewSnapshot("api.example.com", 30)
	c.Store(want)

	got := c.Load()
	if got != want {
		t.Errorf("Load() = %+v, chci %+v", got, want)
	}
}

func TestConfigHotReloadStaysConsistent(t *testing.T) {
	var c exercise.Config
	snaps := []exercise.Snapshot{
		exercise.NewSnapshot("a", 1),
		exercise.NewSnapshot("dlouhý-endpoint.example.com", 250),
		exercise.NewSnapshot("bb", 42),
	}

	stop := make(chan struct{})
	done := c.StartReloader(stop, snaps)

	var bad atomic.Int64
	var wg sync.WaitGroup
	wg.Add(8)
	for i := 0; i < 8; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 20000; j++ {
				if s := c.Load(); !s.Consistent() {
					// Endpoint z jednoho snapshotu a checksum z jiného:
					// čtenář viděl rozepsaný zápis.
					bad.Add(1)
					return
				}
			}
		}()
	}
	wg.Wait()

	close(stop)
	<-done

	if got := bad.Load(); got != 0 {
		t.Errorf("%d čtení vrátilo nekonzistentní konfiguraci — konfigurace se nemění atomicky", got)
	}
	if s := c.Load(); !s.Consistent() {
		t.Errorf("Load() po reloadu = %+v, chci konzistentní snapshot", s)
	}
}
