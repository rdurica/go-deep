package exercise_test

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-42/exercise"
)

// waitNoLeak počká, až počet goroutin klesne zpět na výchozí úroveň.
func waitNoLeak(t *testing.T, before int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if runtime.NumGoroutine() <= before {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("goroutine leak: před testem %d goroutin, po testu %d", before, runtime.NumGoroutine())
}

func TestTrySend(t *testing.T) {
	buf := make(chan int, 1)
	if !exercise.TrySend(buf, 1) {
		t.Error("TrySend do prázdného bufferu = false, chci true")
		return
	}
	if exercise.TrySend(buf, 2) {
		t.Error("TrySend do plného bufferu = true, chci false")
	}
	if got := <-buf; got != 1 {
		t.Errorf("v kanálu je %d, chci 1", got)
	}

	if exercise.TrySend(make(chan int), 1) {
		t.Error("TrySend do nebufferovaného kanálu bez čtenáře = true, chci false")
	}
	if exercise.TrySend(nil, 1) {
		t.Error("TrySend do nil kanálu = true, chci false (nil kanál blokuje navždy)")
	}
}

func TestRecvWithTimeoutValueReady(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 42
	got, ok := exercise.RecvWithTimeout(ch, time.Second)
	if !ok || got != 42 {
		t.Errorf("RecvWithTimeout = (%d, %v), chci (42, true)", got, ok)
	}
}

func TestRecvWithTimeoutValueArrivesLater(t *testing.T) {
	before := runtime.NumGoroutine()

	ch := make(chan int)
	go func() {
		time.Sleep(30 * time.Millisecond)
		ch <- 7
	}()

	got, ok := exercise.RecvWithTimeout(ch, 3*time.Second)
	if !ok || got != 7 {
		t.Errorf("RecvWithTimeout = (%d, %v), chci (7, true)", got, ok)
	}

	waitNoLeak(t, before)
}

func TestRecvWithTimeoutExpires(t *testing.T) {
	before := runtime.NumGoroutine()

	start := time.Now()
	got, ok := exercise.RecvWithTimeout(make(chan int), 50*time.Millisecond)
	elapsed := time.Since(start)

	if ok || got != 0 {
		t.Errorf("RecvWithTimeout = (%d, %v), chci (0, false)", got, ok)
	}
	if elapsed < 40*time.Millisecond {
		t.Errorf("RecvWithTimeout se vrátilo po %v, chci alespoň zhruba 50ms", elapsed)
	}
	if elapsed > 3*time.Second {
		t.Errorf("RecvWithTimeout se vrátilo až po %v — timeout nefunguje", elapsed)
	}

	waitNoLeak(t, before)
}

func TestRecvWithTimeoutClosedChannel(t *testing.T) {
	ch := make(chan int)
	close(ch)
	got, ok := exercise.RecvWithTimeout(ch, 3*time.Second)
	if ok || got != 0 {
		t.Errorf("RecvWithTimeout(zavřený kanál) = (%d, %v), chci (0, false)", got, ok)
	}
}

func TestRecvWithTimeoutDoesNotLeakTimers(t *testing.T) {
	before := runtime.NumGoroutine()

	ch := make(chan int, 1)
	for i := 0; i < 2000; i++ {
		ch <- i
		if got, ok := exercise.RecvWithTimeout(ch, time.Minute); !ok || got != i {
			t.Fatalf("RecvWithTimeout = (%d, %v), chci (%d, true)", got, ok, i)
		}
	}

	waitNoLeak(t, before)
}

var (
	errSlow = errors.New("pomalá funkce selhala")
	errFast = errors.New("rychlá funkce selhala")
)

func TestFirstReturnsFastestSuccess(t *testing.T) {
	before := runtime.NumGoroutine()

	var canceled atomic.Int64
	slow := func(ctx context.Context) (string, error) {
		select {
		case <-ctx.Done():
			canceled.Add(1)
			return "", ctx.Err()
		case <-time.After(5 * time.Second):
			return "pomalý", nil
		}
	}
	fast := func(context.Context) (string, error) {
		return "rychlý", nil
	}

	start := time.Now()
	got, err := exercise.First(context.Background(), slow, fast, slow)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("First vrátil chybu %v, chci úspěch", err)
	}
	if got != "rychlý" {
		t.Errorf("First = %q, chci %q", got, "rychlý")
	}
	if elapsed > 4*time.Second {
		t.Errorf("First čekal %v — nezrušil ostatní a čekal na nejpomalejšího", elapsed)
	}
	if n := canceled.Load(); n != 2 {
		t.Errorf("zrušené funkce: %d, chci 2 — First nezrušil poražené", n)
	}

	waitNoLeak(t, before)
}

func TestFirstJoinsAllErrors(t *testing.T) {
	before := runtime.NumGoroutine()

	a := func(context.Context) (string, error) { return "", errFast }
	b := func(context.Context) (string, error) {
		time.Sleep(20 * time.Millisecond)
		return "", errSlow
	}

	got, err := exercise.First(context.Background(), a, b)
	if err == nil {
		t.Fatalf("First = (%q, nil), chci chybu", got)
	}
	if !errors.Is(err, errFast) {
		t.Errorf("chyba %v neobsahuje %v", err, errFast)
	}
	if !errors.Is(err, errSlow) {
		t.Errorf("chyba %v neobsahuje %v", err, errSlow)
	}

	waitNoLeak(t, before)
}

func TestFirstWithoutFunctions(t *testing.T) {
	if _, err := exercise.First(context.Background()); err == nil {
		t.Error("First bez funkcí vrátil nil chybu, chci chybu")
	}
}

func TestFirstRespectsCanceledContext(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	waiter := func(ctx context.Context) (string, error) {
		<-ctx.Done()
		return "", ctx.Err()
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		if _, err := exercise.First(ctx, waiter, waiter); err == nil {
			t.Error("First se zrušeným kontextem vrátil nil chybu")
		}
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("First se zrušeným kontextem se nevrátil")
	}

	waitNoLeak(t, before)
}

func TestFirstManyCompetitors(t *testing.T) {
	before := runtime.NumGoroutine()

	fns := make([]func(context.Context) (string, error), 0, 50)
	for i := 0; i < 49; i++ {
		fns = append(fns, func(ctx context.Context) (string, error) {
			<-ctx.Done()
			return "", ctx.Err()
		})
	}
	fns = append(fns, func(context.Context) (string, error) { return "vítěz", nil })

	got, err := exercise.First(context.Background(), fns...)
	if err != nil || got != "vítěz" {
		t.Fatalf("First = (%q, %v), chci (\"vítěz\", nil)", got, err)
	}

	waitNoLeak(t, before)
}

func TestDebounceCollapsesBurst(t *testing.T) {
	before := runtime.NumGoroutine()

	const d = 200 * time.Millisecond
	in := make(chan string, 128)
	out := exercise.Debounce(in, d)

	for i := 0; i < 100; i++ {
		in <- fmt.Sprintf("v%d", i)
	}

	select {
	case got := <-out:
		if got != "v99" {
			t.Errorf("Debounce propustil %q, chci poslední hodnotu %q", got, "v99")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Debounce nepropustil ani jednu hodnotu")
	}

	// Z celé dávky smí projít jen jedna hodnota.
	time.Sleep(3 * d)
	select {
	case extra := <-out:
		t.Errorf("Debounce propustil navíc %q, chci z dávky jedinou hodnotu", extra)
	default:
	}

	close(in)
	select {
	case _, ok := <-out:
		if ok {
			t.Error("po zavření vstupu měl být výstup zavřený")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Debounce nezavřel výstup po zavření vstupu")
	}

	waitNoLeak(t, before)
}

func TestDebounceFlushesPendingOnClose(t *testing.T) {
	before := runtime.NumGoroutine()

	in := make(chan string, 4)
	out := exercise.Debounce(in, 10*time.Second) // timer nikdy nevyprší
	in <- "a"
	in <- "b"
	close(in)

	select {
	case got := <-out:
		if got != "b" {
			t.Errorf("Debounce po zavření vstupu poslal %q, chci %q", got, "b")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Debounce po zavření vstupu neposlal čekající hodnotu")
	}
	select {
	case _, ok := <-out:
		if ok {
			t.Error("Debounce poslal víc hodnot, než měl")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Debounce nezavřel výstup")
	}

	waitNoLeak(t, before)
}

func TestDebounceEmptyInput(t *testing.T) {
	before := runtime.NumGoroutine()

	in := make(chan string)
	out := exercise.Debounce(in, 20*time.Millisecond)
	close(in)

	select {
	case v, ok := <-out:
		if ok {
			t.Errorf("Debounce z prázdného vstupu poslal %q", v)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Debounce nezavřel výstup u prázdného vstupu")
	}

	waitNoLeak(t, before)
}

func TestHeartbeatBeatsAndCallsWork(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var work atomic.Int64
	beats := exercise.Heartbeat(ctx, 20*time.Millisecond, func() { work.Add(1) })

	deadline := time.After(400 * time.Millisecond)
	count := 0
loop:
	for {
		select {
		case _, ok := <-beats:
			if !ok {
				break loop
			}
			count++
		case <-deadline:
			break loop
		}
	}

	// Velkorysá tolerance: na vytíženém stroji tepy zpomalí, nikdy ale
	// nesmí být rychlejší, než dovolí interval.
	if count < 3 {
		t.Errorf("za 400ms dorazilo %d tepů při intervalu 20ms, chci alespoň 3", count)
	}
	if count > 60 {
		t.Errorf("za 400ms dorazilo %d tepů při intervalu 20ms — ticker tepe moc rychle", count)
	}
	if got := work.Load(); got < int64(count) {
		t.Errorf("work zavolán %dx, ale tepů dorazilo %d", got, count)
	}

	cancel()
	select {
	case _, ok := <-beats:
		if ok {
			// Jeden tep mohl být rozeslaný těsně před zrušením; přečteme
			// kanál až do zavření.
			for range beats {
			}
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Heartbeat po zrušení kontextu nezavřel kanál")
	}

	waitNoLeak(t, before)
}

func TestHeartbeatStopsWithoutReader(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	beats := exercise.Heartbeat(ctx, 10*time.Millisecond, nil)

	// Přečteme jeden tep a pak přestaneme číst úplně. Goroutina uvízlá na
	// `out <- now` by tady leakovala.
	select {
	case <-beats:
	case <-time.After(3 * time.Second):
		t.Fatal("Heartbeat neposlal ani jeden tep")
	}
	time.Sleep(50 * time.Millisecond)
	cancel()

	waitNoLeak(t, before)
}

func TestHeartbeatAlreadyCanceledContext(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	beats := exercise.Heartbeat(ctx, 10*time.Millisecond, nil)
	select {
	case _, ok := <-beats:
		if ok {
			t.Error("Heartbeat se zrušeným kontextem poslal tep")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Heartbeat se zrušeným kontextem nezavřel kanál")
	}

	waitNoLeak(t, before)
}
