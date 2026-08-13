package exercise_test

import (
	"runtime"
	"sort"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-41/exercise"
)

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

func collectInt(ch <-chan int) []int {
	var out []int
	for v := range ch {
		out = append(out, v)
	}
	return out
}

func TestForgetCloseFinishes(t *testing.T) {
	before := runtime.NumGoroutine()

	ch := exercise.ForgetClose(1, 2, 3)
	done := make(chan struct{})
	go func() {
		for range ch {
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ForgetClose nezavřel kanál — příjemce visí navždy")
	}

	waitNoLeak(t, before)
}

func TestForgetCloseDeliversValues(t *testing.T) {
	got := collectInt(exercise.ForgetClose(1, 2, 3))
	want := []int{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("ForgetClose = %v, chci %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("výsledek[%d] = %d, chci %d", i, got[i], want[i])
		}
	}
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
