package exercise_test

import (
	"runtime"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-49/exercise"
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

func TestRunWithMaxProcsRestoresValue(t *testing.T) {
	original := runtime.GOMAXPROCS(0)

	var inside int
	exercise.RunWithMaxProcs(1, func() {
		inside = runtime.GOMAXPROCS(0)
	})

	if inside != 1 {
		t.Errorf("uvnitř f je GOMAXPROCS = %d, chci 1", inside)
	}
	if now := runtime.GOMAXPROCS(0); now != original {
		runtime.GOMAXPROCS(original)
		t.Fatalf("po návratu je GOMAXPROCS = %d, chci obnovených %d", now, original)
	}
}

func TestRunWithMaxProcsRestoresAfterPanic(t *testing.T) {
	original := runtime.GOMAXPROCS(0)

	func() {
		defer func() {
			if recover() == nil {
				t.Error("panika z f se měla propagovat ven")
			}
		}()
		exercise.RunWithMaxProcs(1, func() { panic("bum") })
	}()

	if now := runtime.GOMAXPROCS(0); now != original {
		runtime.GOMAXPROCS(original)
		t.Fatalf("po panice je GOMAXPROCS = %d, chci obnovených %d", now, original)
	}
}

func TestRunWithMaxProcsEdgeCases(t *testing.T) {
	original := runtime.GOMAXPROCS(0)

	exercise.RunWithMaxProcs(1, nil) // nil f nesmí panikovat

	var seen int
	exercise.RunWithMaxProcs(0, func() { seen = runtime.GOMAXPROCS(0) })
	if seen != original {
		t.Errorf("pro n = 0 chci nezměněné GOMAXPROCS %d, dostal jsem %d", original, seen)
	}
	if now := runtime.GOMAXPROCS(0); now != original {
		runtime.GOMAXPROCS(original)
		t.Fatalf("GOMAXPROCS = %d, chci %d", now, original)
	}
}

func TestCPUBoundRunsEvenWithSingleProc(t *testing.T) {
	before := runtime.NumGoroutine()

	want := exercise.CPUBound(50_000)
	var got uint64
	exercise.RunWithMaxProcs(1, func() {
		got = exercise.CPUBound(50_000)
	})

	if got != want {
		t.Errorf("CPUBound při GOMAXPROCS=1 = %d, chci %d", got, want)
	}
	if exercise.CPUBound(0) == exercise.CPUBound(1) {
		t.Error("CPUBound(0) a CPUBound(1) mají dát různý výsledek — funkce nic nepočítá")
	}
	waitNoLeak(t, before)
}

func TestObserveParallelism(t *testing.T) {
	before := runtime.NumGoroutine()

	if got := exercise.ObserveParallelism(0); got != 0 {
		t.Errorf("ObserveParallelism(0) = %d, chci 0", got)
	}
	if got := exercise.ObserveParallelism(-3); got != 0 {
		t.Errorf("ObserveParallelism(-3) = %d, chci 0", got)
	}
	if got := exercise.ObserveParallelism(1); got != 1 {
		t.Errorf("ObserveParallelism(1) = %d, chci 1", got)
	}

	const workers = 8
	got := exercise.ObserveParallelism(workers)
	if got < 1 || got > workers {
		t.Errorf("ObserveParallelism(%d) = %d, chci hodnotu mezi 1 a %d", workers, got, workers)
	}
	if got <= 1 {
		t.Errorf("ObserveParallelism(%d) = %d — goroutiny čekající na kanálu se musí sejít naráz", workers, got)
	}
	waitNoLeak(t, before)
}

// TestObserveParallelismWithSingleProc ukazuje pointu netpolleru a blokování:
// i s jediným P se čekající goroutiny sejdou všechny najednou.
func TestObserveParallelismWithSingleProc(t *testing.T) {
	before := runtime.NumGoroutine()
	const workers = 8

	var got int
	exercise.RunWithMaxProcs(1, func() {
		got = exercise.ObserveParallelism(workers)
	})

	if got < 1 || got > workers {
		t.Errorf("při GOMAXPROCS=1 je souběh %d, chci hodnotu mezi 1 a %d", got, workers)
	}
	if got <= 1 {
		t.Errorf("při GOMAXPROCS=1 je souběh %d — blokující operace nedrží P, souběh má být vyšší", got)
	}
	waitNoLeak(t, before)
}

func TestBlocking(t *testing.T) {
	start := time.Now()
	exercise.Blocking(30 * time.Millisecond)
	if elapsed := time.Since(start); elapsed < 25*time.Millisecond {
		t.Errorf("Blocking(30ms) trval jen %v", elapsed)
	}

	start = time.Now()
	exercise.Blocking(0)
	exercise.Blocking(-time.Second)
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("Blocking s nekladnou dobou trval %v, chci návrat hned", elapsed)
	}
}

func TestCompareBlockingScales(t *testing.T) {
	before := runtime.NumGoroutine()
	const workers = 8

	cpu, blocking := exercise.Compare(workers)

	if cpu <= 0 {
		t.Errorf("CPU-bound část trvala %v, chci měřitelný čas", cpu)
	}
	if blocking < exercise.BlockingDuration/2 {
		t.Errorf("blokující část trvala %v, chci aspoň zhruba %v", blocking, exercise.BlockingDuration)
	}
	// Kdyby se blokující volání neprolínala, trvalo by to workers násobek.
	// Velkorysá mez: pětinásobek jednoho spánku místo osminásobku.
	if limit := 5 * exercise.BlockingDuration; blocking > limit {
		t.Errorf("blokující část trvala %v, chci pod %v — blokující goroutiny musí čekat souběžně", blocking, limit)
	}

	if _, b := exercise.Compare(0); b < exercise.BlockingDuration/2 {
		t.Errorf("Compare(0) se má chovat jako Compare(1), blokující část = %v", b)
	}
	waitNoLeak(t, before)
}

func TestGoroutineCost(t *testing.T) {
	before := runtime.NumGoroutine()

	if b, a := exercise.GoroutineCost(0); b != a {
		t.Errorf("GoroutineCost(0) = (%d, %d), chci dvě stejné hodnoty", b, a)
	}

	const n = 500
	b, a := exercise.GoroutineCost(n)
	if a-b < n {
		t.Errorf("GoroutineCost(%d) = (%d, %d), chci rozdíl aspoň %d", n, b, a, n)
	}
	waitNoLeak(t, before)
}
