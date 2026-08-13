package solutions_test

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-48/solutions"
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

func TestNaiveFlagConcurrentSet(t *testing.T) {
	var f exercise.NaiveFlag
	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				f.Set(true)
			}
		}()
	}
	wg.Wait()
	if !f.Get() {
		t.Error("Get() = false po souběžném Set(true) — příznak ztrácí zápis")
	}
}

func TestAtomicFlagBasics(t *testing.T) {
	f := &exercise.AtomicFlag{}
	if f.Get() {
		t.Error("nulová hodnota příznaku má být false")
	}
	f.Set(true)
	if !f.Get() {
		t.Error("po Set(true) chci Get() == true")
	}
	f.Set(false)
	if f.Get() {
		t.Error("po Set(false) chci Get() == false")
	}
}

func TestAtomicFlagConcurrentStress(t *testing.T) {
	before := runtime.NumGoroutine()
	f := &exercise.AtomicFlag{}

	const (
		writers    = 8
		readers    = 8
		iterations = 500
	)
	var trueReads atomic.Int64
	var wg sync.WaitGroup
	wg.Add(writers + readers)
	for w := 0; w < writers; w++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				f.Set(true)
			}
		}()
	}
	for r := 0; r < readers; r++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				if f.Get() {
					trueReads.Add(1)
				}
			}
		}()
	}
	wg.Wait()

	if !f.Get() {
		t.Error("po zátěži chci Get() == true")
	}
	if got := int(trueReads.Load()); got < 0 || got > readers*iterations {
		t.Errorf("počet true čtení = %d, chci hodnotu mezi 0 a %d", got, readers*iterations)
	}
	waitNoLeak(t, before)
}

func TestLazyInitRunsExactlyOnce(t *testing.T) {
	before := runtime.NumGoroutine()

	var calls atomic.Int64
	l := exercise.NewLazyInit(func() int {
		calls.Add(1)
		time.Sleep(20 * time.Millisecond)
		return 42
	})

	const n = 100
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			<-start
			if v := l.Value(); v != 42 {
				t.Errorf("Value() = %d, chci 42", v)
			}
		}()
	}
	close(start)
	wg.Wait()

	if c := calls.Load(); c != 1 {
		t.Errorf("inicializace proběhla %dkrát, chci právě jednou", c)
	}
	if v := l.Value(); v != 42 {
		t.Errorf("Value() po zátěži = %d, chci 42", v)
	}
	waitNoLeak(t, before)
}

func TestLazyInitEdgeCases(t *testing.T) {
	t.Run("nil init panics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("NewLazyInit(nil) měl panikovat")
			}
		}()
		_ = exercise.NewLazyInit(nil)
	})
}

func TestBoxConsumeBlocksUntilPublish(t *testing.T) {
	before := runtime.NumGoroutine()

	b := exercise.NewBox()
	done := make(chan []int, 1)
	go func() { done <- b.Consume() }()

	select {
	case v := <-done:
		t.Fatalf("Consume vrátil %v ještě před Publish, měl blokovat", v)
	case <-time.After(50 * time.Millisecond):
	}

	b.Publish([]int{1, 2, 3})
	select {
	case v := <-done:
		if len(v) != 3 || v[0] != 1 || v[2] != 3 {
			t.Errorf("Consume vrátil %v, chci [1 2 3]", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Consume se po Publish neodblokoval")
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if v := b.Consume(); len(v) != 3 {
				t.Errorf("opakovaný Consume vrátil %v, chci [1 2 3]", v)
			}
		}()
	}
	wg.Wait()
	waitNoLeak(t, before)
}

func TestBoxManyReadersSeeFullData(t *testing.T) {
	before := runtime.NumGoroutine()

	data := make([]int, 1000)
	for i := range data {
		data[i] = i * 3
	}

	b := exercise.NewBox()
	const readers = 16
	out := make([][]int, readers)
	var wg sync.WaitGroup
	wg.Add(readers)
	for i := 0; i < readers; i++ {
		go func(idx int) {
			defer wg.Done()
			out[idx] = b.Consume()
		}(i)
	}
	b.Publish(data)
	wg.Wait()

	for r, seen := range out {
		if len(seen) != len(data) {
			t.Fatalf("čtenář %d viděl %d prvků, chci %d", r, len(seen), len(data))
		}
		for i := range data {
			if seen[i] != data[i] {
				t.Fatalf("čtenář %d na indexu %d viděl %d, chci %d", r, i, seen[i], data[i])
			}
		}
	}
	waitNoLeak(t, before)
}
