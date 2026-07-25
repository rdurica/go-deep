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

func TestFlagImplementations(t *testing.T) {
	var _ exercise.Flag = (*exercise.AtomicFlag)(nil)
	var _ exercise.Flag = (*exercise.MutexFlag)(nil)
}

func TestFlagBasics(t *testing.T) {
	flags := map[string]exercise.Flag{
		"AtomicFlag": &exercise.AtomicFlag{},
		"MutexFlag":  &exercise.MutexFlag{},
	}
	for name, f := range flags {
		t.Run(name, func(t *testing.T) {
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
		})
	}
}

// TestStressFlagIsRaceFree je hlavní test úkolu A. Pod -race musí projít
// bez jediného hlášení závodu a příznak musí na konci držet zapsanou hodnotu.
func TestStressFlagIsRaceFree(t *testing.T) {
	flags := map[string]exercise.Flag{
		"AtomicFlag": &exercise.AtomicFlag{},
		"MutexFlag":  &exercise.MutexFlag{},
	}
	for name, f := range flags {
		t.Run(name, func(t *testing.T) {
			before := runtime.NumGoroutine()

			const (
				writers    = 8
				readers    = 8
				iterations = 500
			)
			trueReads := exercise.StressFlag(f, writers, readers, iterations)

			if !f.Get() {
				t.Error("po zátěži chci Get() == true, zapisovatelé nastavovali true")
			}
			if trueReads < 0 || trueReads > readers*iterations {
				t.Errorf("StressFlag vrátil %d, chci hodnotu mezi 0 a %d", trueReads, readers*iterations)
			}
			waitNoLeak(t, before)
		})
	}
}

func TestStressFlagEdgeCases(t *testing.T) {
	before := runtime.NumGoroutine()

	if got := exercise.StressFlag(nil, 2, 2, 10); got != 0 {
		t.Errorf("StressFlag(nil, …) = %d, chci 0", got)
	}
	f := &exercise.AtomicFlag{}
	if got := exercise.StressFlag(f, 0, 2, 10); got != 0 {
		t.Errorf("StressFlag bez zapisovatelů = %d, chci 0", got)
	}
	if got := exercise.StressFlag(f, 2, 2, 0); got != 0 {
		t.Errorf("StressFlag s nulou iterací = %d, chci 0", got)
	}
	if f.Get() {
		t.Error("žádná z hraničních variant neměla nic nastavit")
	}
	waitNoLeak(t, before)
}

func TestLazyInitRunsExactlyOnce(t *testing.T) {
	before := runtime.NumGoroutine()

	var calls atomic.Int64
	l := exercise.NewLazyInit(func() int {
		calls.Add(1)
		// Chvilka práce, ať mají ostatní goroutiny šanci dorazit na Value.
		time.Sleep(20 * time.Millisecond)
		return 42
	})

	const n = 100
	got := exercise.ConcurrentValues(l, n)

	if len(got) != n {
		t.Fatalf("ConcurrentValues vrátilo %d hodnot, chci %d", len(got), n)
	}
	for i, v := range got {
		if v != 42 {
			t.Fatalf("goroutina %d viděla %d, chci 42", i, v)
		}
	}
	if c := calls.Load(); c != 1 {
		t.Errorf("inicializace proběhla %dkrát, chci právě jednou", c)
	}
	if v := l.Value(); v != 42 {
		t.Errorf("Value() po zátěži = %d, chci 42", v)
	}
	if c := calls.Load(); c != 1 {
		t.Errorf("další Value() spustilo inicializaci znovu (%d volání)", c)
	}
	waitNoLeak(t, before)
}

func TestLazyInitEdgeCases(t *testing.T) {
	t.Run("nil init panikuje", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("NewLazyInit(nil) měl panikovat")
			}
		}()
		_ = exercise.NewLazyInit(nil)
	})

	t.Run("nekladné n", func(t *testing.T) {
		l := exercise.NewLazyInit(func() int { return 1 })
		if got := exercise.ConcurrentValues(l, 0); len(got) != 0 {
			t.Errorf("ConcurrentValues(l, 0) = %v, chci prázdný výsledek", got)
		}
	})
}

func TestBoxPublishConsume(t *testing.T) {
	before := runtime.NumGoroutine()

	data := make([]int, 1000)
	for i := range data {
		data[i] = i * 3
	}

	const readers = 16
	got := exercise.PublishAndConsume(data, readers)

	if len(got) != readers {
		t.Fatalf("PublishAndConsume vrátilo %d výsledků, chci %d", len(got), readers)
	}
	for r, seen := range got {
		if len(seen) != len(data) {
			t.Fatalf("čtenář %d viděl %d prvků, chci %d", r, len(seen), len(data))
		}
		for i := range data {
			if seen[i] != data[i] {
				t.Fatalf("čtenář %d viděl na indexu %d hodnotu %d, chci %d", r, i, seen[i], data[i])
			}
		}
	}
	if got := exercise.PublishAndConsume(data, 0); got != nil {
		t.Errorf("PublishAndConsume s nula čtenáři = %v, chci nil", got)
	}
	waitNoLeak(t, before)
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

	// Po Publish musí Consume vracet okamžitě libovolnému počtu čtenářů.
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

func TestWaitGroupVisibility(t *testing.T) {
	before := runtime.NumGoroutine()

	tests := []struct {
		n    int
		want int
	}{
		{0, 0},
		{-5, 0},
		{1, 0},
		{4, 0 + 1 + 4 + 9},
		{100, 328350},
	}
	for _, tt := range tests {
		if got := exercise.WaitGroupVisibility(tt.n); got != tt.want {
			t.Errorf("WaitGroupVisibility(%d) = %d, chci %d", tt.n, got, tt.want)
		}
	}
	waitNoLeak(t, before)
}
