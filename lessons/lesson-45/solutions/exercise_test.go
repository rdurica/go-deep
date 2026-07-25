package solutions_test

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-45/solutions"
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

// stringSource je nekonečný zdroj řetězců, který respektuje kontext.
func stringSource(ctx context.Context, values ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for i := 0; ; i++ {
			select {
			case out <- values[i%len(values)]:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func TestGenAndSquare(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var got []int
	for v := range exercise.Square(ctx, exercise.Gen(ctx, 1, 2, 3, 4)) {
		got = append(got, v)
	}

	want := []int{1, 4, 9, 16}
	if len(got) != len(want) {
		t.Fatalf("pipeline vrátila %v, chci %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("výsledek[%d] = %d, chci %d — pipeline musí zachovat pořadí", i, got[i], want[i])
		}
	}

	cancel()
	waitNoLeak(t, before)
}

func TestGenEmpty(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	select {
	case _, ok := <-exercise.Gen(ctx):
		if ok {
			t.Error("Gen bez čísel poslal hodnotu")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Gen bez čísel nezavřel kanál")
	}

	waitNoLeak(t, before)
}

func TestGenStopsOnCancel(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	nums := make([]int, 10_000)
	for i := range nums {
		nums[i] = i
	}

	ch := exercise.Gen(ctx, nums...)
	if got := <-ch; got != 0 {
		t.Fatalf("první hodnota = %d, chci 0", got)
	}
	cancel() // konzument končí uprostřed proudu

	waitNoLeak(t, before)
}

func TestSquareStopsOnCancel(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	nums := make([]int, 10_000)
	for i := range nums {
		nums[i] = i
	}

	out := exercise.Square(ctx, exercise.Gen(ctx, nums...))
	<-out
	cancel()

	waitNoLeak(t, before)
}

func TestStage(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	words := exercise.Stage(ctx, exercise.Gen(ctx, 1, 2, 3), func(n int) string {
		return fmt.Sprintf("#%d", n)
	})

	var got []string
	for w := range words {
		got = append(got, w)
	}
	if want := "#1,#2,#3"; strings.Join(got, ",") != want {
		t.Errorf("Stage vrátil %v, chci %s", got, want)
	}

	cancel()
	waitNoLeak(t, before)
}

func TestStageStopsOnCancel(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	nums := make([]int, 10_000)
	for i := range nums {
		nums[i] = i
	}

	out := exercise.Stage(ctx, exercise.Gen(ctx, nums...), func(n int) int { return n + 1 })
	<-out
	cancel()

	waitNoLeak(t, before)
}

func TestFanIn(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := exercise.Gen(ctx, 1, 2, 3)
	b := exercise.Gen(ctx, 4, 5)
	c := exercise.Gen(ctx, 6)

	var got []int
	for v := range exercise.FanIn(ctx, a, b, c) {
		got = append(got, v)
	}
	sort.Ints(got)

	want := []int{1, 2, 3, 4, 5, 6}
	if len(got) != len(want) {
		t.Fatalf("FanIn vrátil %v, chci (v libovolném pořadí) %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("FanIn vrátil %v, chci (v libovolném pořadí) %v", got, want)
		}
	}

	cancel()
	waitNoLeak(t, before)
}

func TestFanInWithoutInputs(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	select {
	case _, ok := <-exercise.FanIn[int](ctx):
		if ok {
			t.Error("FanIn bez vstupů poslal hodnotu")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("FanIn bez vstupů nezavřel výstup")
	}

	waitNoLeak(t, before)
}

func TestTake(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	got := []int{}
	for v := range exercise.Take(ctx, exercise.Gen(ctx, 1, 2, 3, 4, 5), 3) {
		got = append(got, v)
	}
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Errorf("Take(_, 3) = %v, chci [1 2 3]", got)
	}

	cancel()
	waitNoLeak(t, before)
}

func TestTakeMoreThanAvailable(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	got := []int{}
	for v := range exercise.Take(ctx, exercise.Gen(ctx, 1, 2), 10) {
		got = append(got, v)
	}
	if len(got) != 2 {
		t.Errorf("Take(_, 10) ze dvou hodnot = %v, chci dvě hodnoty", got)
	}

	waitNoLeak(t, before)
}

func TestTakeNonPositive(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, n := range []int{0, -1} {
		out := exercise.Take(ctx, exercise.Gen(ctx, 1, 2, 3), n)
		select {
		case v, ok := <-out:
			if ok {
				t.Errorf("Take(_, %d) poslal %d, chci rovnou zavřený kanál", n, v)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("Take(_, %d) nezavřel výstup", n)
		}
	}

	cancel()
	waitNoLeak(t, before)
}

// TestTakeStopsWholePipeline je jádro celé lekce: konzument odebere pár prvků
// a odejde. Všechny stupně nad ním musí skončit.
func TestTakeStopsWholePipeline(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())

	nums := make([]int, 100_000)
	for i := range nums {
		nums[i] = i
	}

	src := exercise.Gen(ctx, nums...)
	squared := exercise.Square(ctx, src)
	doubled := exercise.Stage(ctx, squared, func(n int) int { return 2 * n })
	out := exercise.Take(ctx, doubled, 5)

	count := 0
	for range out {
		count++
	}
	if count != 5 {
		t.Fatalf("Take propustil %d prvků, chci 5", count)
	}

	cancel() // až teď říkáme zbytku pipeline, že skončil
	waitNoLeak(t, before)
}

func TestPipeline(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inputs := []string{"alfa", "  beta  ", "", "   ", "gama7", "delta"}
	in := make(chan string, len(inputs))
	for _, s := range inputs {
		in <- s
	}
	close(in)

	type want struct {
		value string
		err   error
	}
	expected := map[string]want{
		"alfa":     {"ok:ALFA", nil},
		"  beta  ": {"ok:BETA", nil},
		"":         {"", exercise.ErrEmpty},
		"   ":      {"", exercise.ErrEmpty},
		"gama7":    {"", exercise.ErrDigits},
		"delta":    {"ok:DELTA", nil},
	}

	got := map[string]exercise.Result{}
	for res := range exercise.Pipeline(ctx, in) {
		if _, dup := got[res.Input]; dup {
			t.Errorf("výsledek pro %q dorazil dvakrát", res.Input)
		}
		got[res.Input] = res
	}

	if len(got) != len(expected) {
		t.Fatalf("Pipeline vrátila %d výsledků, chci %d", len(got), len(expected))
	}
	for input, w := range expected {
		res, ok := got[input]
		if !ok {
			t.Errorf("chybí výsledek pro vstup %q", input)
			continue
		}
		if !errors.Is(res.Err, w.err) {
			t.Errorf("Pipeline(%q).Err = %v, chci %v", input, res.Err, w.err)
		}
		if res.Value != w.value {
			t.Errorf("Pipeline(%q).Value = %q, chci %q", input, res.Value, w.value)
		}
	}

	cancel()
	waitNoLeak(t, before)
}

func TestPipelineEmptyInput(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan string)
	close(in)

	for res := range exercise.Pipeline(ctx, in) {
		t.Errorf("Pipeline z prázdného vstupu vrátila %+v", res)
	}

	waitNoLeak(t, before)
}

// TestPipelineConsumerLeavesEarly ověřuje nejčastější chybu v pipeline:
// konzument přestane odebírat a všechny stupně zůstanou viset.
func TestPipelineConsumerLeavesEarly(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())

	in := stringSource(ctx, "alfa", "beta", " gama ", "delta9", "")
	out := exercise.Pipeline(ctx, in)

	for i := 0; i < 3; i++ {
		select {
		case res, ok := <-out:
			if !ok {
				t.Fatal("Pipeline zavřela výstup dřív, než dodala tři výsledky")
			}
			if res.Err == nil && res.Value == "" {
				t.Errorf("výsledek %+v nemá ani hodnotu, ani chybu", res)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("Pipeline nedodala výsledky")
		}
	}

	cancel() // konzument odchází uprostřed proudu
	waitNoLeak(t, before)
}

func TestPipelineWithManyItems(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const n = 2000
	in := make(chan string, n)
	for i := 0; i < n; i++ {
		in <- "položka"
	}
	close(in)

	count := 0
	for res := range exercise.Pipeline(ctx, in) {
		if res.Err != nil {
			t.Fatalf("neočekávaná chyba %v", res.Err)
		}
		if res.Value != "ok:POLOŽKA" {
			t.Fatalf("Value = %q, chci %q", res.Value, "ok:POLOŽKA")
		}
		count++
	}
	if count != n {
		t.Errorf("Pipeline vrátila %d výsledků, chci %d", count, n)
	}

	cancel()
	waitNoLeak(t, before)
}

// BenchmarkSquarePipeline a BenchmarkSquareLoop ukazují, že pro malá data je
// obyčejná smyčka řádově rychlejší než pipeline. Spusť:
//
//	go test -bench=Square -benchmem .
func BenchmarkSquarePipeline(b *testing.B) {
	nums := make([]int, 100)
	for i := range nums {
		nums[i] = i
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for v := range exercise.Square(ctx, exercise.Gen(ctx, nums...)) {
			sum += v
		}
		_ = sum
	}
}

func BenchmarkSquareLoop(b *testing.B) {
	nums := make([]int, 100)
	for i := range nums {
		nums[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for _, v := range nums {
			sum += v * v
		}
		_ = sum
	}
}
