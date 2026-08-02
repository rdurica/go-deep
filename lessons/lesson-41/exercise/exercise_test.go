package exercise_test

import (
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-41/exercise"
)

// waitNoLeak počká, až počet goroutin klesne zpět na výchozí úroveň.
// Opakované krátké čekání je spolehlivější než jeden pevný sleep.
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

// intsSource pošle nums do kanálu v goroutině a zavře ho — lokální fixture bez Generate.
func intsSource(nums ...int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, n := range nums {
			ch <- n
		}
	}()
	return ch
}

// collectInt přečte kanál do slice — lokální helper bez Collect.
func collectInt(ch <-chan int) []int {
	var out []int
	for v := range ch {
		out = append(out, v)
	}
	return out
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name string
		in   []int
	}{
		{"empty", nil},
		{"one element", []int{7}},
		{"several items", []int{1, 2, 3, 4, 5}},
		{"duplicity a nuly", []int{0, 0, -1, -1, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectInt(exercise.Generate(tt.in...))
			if len(got) != len(tt.in) {
				t.Fatalf("Generate(%v) = %v, chci %v", tt.in, got, tt.in)
			}
			for i := range tt.in {
				if got[i] != tt.in[i] {
					t.Errorf("výsledek[%d] = %d, chci %d", i, got[i], tt.in[i])
				}
			}
		})
	}
}

func TestCollect(t *testing.T) {
	tests := []struct {
		name string
		in   []int
	}{
		{"empty", nil},
		{"one element", []int{7}},
		{"several items", []int{1, 2, 3, 4, 5}},
		{"duplicity a nuly", []int{0, 0, -1, -1, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exercise.Collect(intsSource(tt.in...))
			if len(got) != len(tt.in) {
				t.Fatalf("Collect(...) = %v, chci %v", got, tt.in)
			}
			for i := range tt.in {
				if got[i] != tt.in[i] {
					t.Errorf("výsledek[%d] = %d, chci %d", i, got[i], tt.in[i])
				}
			}
		})
	}
}

func TestGenerateClosesChannel(t *testing.T) {
	before := runtime.NumGoroutine()

	ch := exercise.Generate(1, 2)
	if v := <-ch; v != 1 {
		t.Fatalf("první hodnota = %d, chci 1", v)
	}
	if v := <-ch; v != 2 {
		t.Fatalf("druhá hodnota = %d, chci 2", v)
	}
	select {
	case v, ok := <-ch:
		if ok {
			t.Fatalf("kanál měl být zavřený, dostal jsem %d", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Generate kanál nezavřel — Collect by na něm visel navždy")
	}

	waitNoLeak(t, before)
}

func TestCollectOnAlreadyClosedChannel(t *testing.T) {
	ch := make(chan int)
	close(ch)
	if got := exercise.Collect(ch); len(got) != 0 {
		t.Errorf("Collect(zavřený kanál) = %v, chci prázdný výsledek", got)
	}
}

func TestMerge(t *testing.T) {
	before := runtime.NumGoroutine()

	a := intsSource(1, 2, 3)
	b := intsSource(10, 20)
	c := intsSource(100)

	got := collectInt(exercise.Merge(a, b, c))
	sort.Ints(got)

	want := []int{1, 2, 3, 10, 20, 100}
	if len(got) != len(want) {
		t.Fatalf("Merge vrátil %v, chci (v libovolném pořadí) %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Merge vrátil %v, chci (v libovolném pořadí) %v", got, want)
		}
	}

	waitNoLeak(t, before)
}

func TestMergeWithoutInputsClosesImmediately(t *testing.T) {
	before := runtime.NumGoroutine()

	out := exercise.Merge()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("Merge() bez vstupů poslal hodnotu, chci rovnou zavřený kanál")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Merge() bez vstupů nezavřel výstup")
	}

	waitNoLeak(t, before)
}

func TestMergeLargeNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()

	const chans, perChan = 8, 500
	ins := make([]<-chan int, chans)
	for i := range ins {
		nums := make([]int, perChan)
		for j := range nums {
			nums[j] = 1
		}
		ins[i] = intsSource(nums...)
	}

	sum := 0
	for v := range exercise.Merge(ins...) {
		sum += v
	}
	if want := chans * perChan; sum != want {
		t.Errorf("Merge propustil součet %d, chci %d", sum, want)
	}

	waitNoLeak(t, before)
}

func TestSplit(t *testing.T) {
	before := runtime.NumGoroutine()

	const n = 500
	nums := make([]int, n)
	for i := range nums {
		nums[i] = i
	}

	outs := exercise.Split(intsSource(nums...), 4)
	if len(outs) != 4 {
		t.Fatalf("Split vrátil %d kanálů, chci 4", len(outs))
	}

	var mu sync.Mutex
	got := make([]int, 0, n)
	var wg sync.WaitGroup
	wg.Add(len(outs))
	for _, out := range outs {
		go func() {
			defer wg.Done()
			for v := range out { // musí skončit, jinak test spadne na timeoutu
				mu.Lock()
				got = append(got, v)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	sort.Ints(got)
	if len(got) != n {
		t.Fatalf("Split propustil %d hodnot, chci %d", len(got), n)
	}
	for i := 0; i < n; i++ {
		if got[i] != i {
			t.Fatalf("Split ztratil nebo zduplikoval hodnoty: na pozici %d je %d", i, got[i])
		}
	}

	waitNoLeak(t, before)
}

func TestSplitNonPositiveN(t *testing.T) {
	before := runtime.NumGoroutine()

	for _, n := range []int{0, -1} {
		outs := exercise.Split(intsSource(1, 2, 3), n)
		if len(outs) != 1 {
			t.Fatalf("Split(_, %d) vrátil %d kanálů, chci 1", n, len(outs))
		}
		if got := collectInt(outs[0]); len(got) != 3 {
			t.Errorf("Split(_, %d) propustil %v, chci tři hodnoty", n, got)
		}
	}

	waitNoLeak(t, before)
}

func TestBrokerDeliversToAllSubscribers(t *testing.T) {
	before := runtime.NumGoroutine()

	b := exercise.NewBroker(8)
	s1 := b.Subscribe()
	s2 := b.Subscribe()

	b.Publish("a")
	b.Publish("b")
	b.Close()

	for name, sub := range map[string]<-chan string{"s1": s1, "s2": s2} {
		var got []string
		for msg := range sub {
			got = append(got, msg)
		}
		if len(got) != 2 || got[0] != "a" || got[1] != "b" {
			t.Errorf("odběratel %s dostal %v, chci [a b]", name, got)
		}
	}

	if b.Dropped() != 0 {
		t.Errorf("Dropped() = %d, chci 0 — buffer stačil", b.Dropped())
	}

	waitNoLeak(t, before)
}

func TestBrokerDropsForSlowSubscriber(t *testing.T) {
	b := exercise.NewBroker(1)
	sub := b.Subscribe()
	defer b.Close()

	// Odběratel nečte vůbec. První zpráva se vejde do bufferu, další čtyři
	// se musí zahodit — a Publish nesmí zablokovat.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 5; i++ {
			b.Publish("zpráva")
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Publish zablokoval na pomalém odběrateli")
	}

	if got := b.Dropped(); got != 4 {
		t.Errorf("Dropped() = %d, chci 4", got)
	}
	if got := <-sub; got != "zpráva" {
		t.Errorf("odběratel dostal %q, chci %q", got, "zpráva")
	}
}

func TestBrokerCloseEndsSubscriberGoroutines(t *testing.T) {
	before := runtime.NumGoroutine()

	b := exercise.NewBroker(4)
	const subs = 20
	var wg sync.WaitGroup
	wg.Add(subs)
	for i := 0; i < subs; i++ {
		sub := b.Subscribe()
		go func() {
			defer wg.Done()
			for range sub { // skončí až zavřením kanálu
			}
		}()
	}

	b.Publish("ahoj")
	b.Close()

	finished := make(chan struct{})
	go func() {
		wg.Wait()
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(3 * time.Second):
		t.Fatal("Close neukončil odběratele — jejich goroutiny visí na range")
	}

	waitNoLeak(t, before)
}

func TestBrokerSubscribeAfterCloseReturnsClosedChannel(t *testing.T) {
	b := exercise.NewBroker(1)
	b.Close()

	sub := b.Subscribe()
	select {
	case _, ok := <-sub:
		if ok {
			t.Fatal("odběratel zaregistrovaný po Close dostal zprávu")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Subscribe po Close nevrátil zavřený kanál")
	}
}

func TestBrokerCloseIsIdempotentAndPublishAfterCloseIsNoop(t *testing.T) {
	b := exercise.NewBroker(1)
	sub := b.Subscribe()
	b.Close()
	b.Close() // druhé zavření nesmí panikovat
	b.Publish("pozdě")

	if _, ok := <-sub; ok {
		t.Error("Publish po Close poslal zprávu")
	}
}

func TestBrokerConcurrentPublishAndSubscribe(t *testing.T) {
	before := runtime.NumGoroutine()

	b := exercise.NewBroker(16)
	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				b.Publish("x")
			}
		}()
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sub := b.Subscribe()
			for range sub {
			}
		}()
	}

	time.Sleep(50 * time.Millisecond)
	b.Close()
	wg.Wait()

	_ = b.Dropped() // jen ověřujeme, že se dá volat souběžně s ostatními

	waitNoLeak(t, before)
}
