package solutions_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-10/solutions"
)

func TestDeferOrderIsLIFO(t *testing.T) {
	got := exercise.DeferOrder()
	want := []string{"third", "second", "first"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DeferOrder() = %v, chci %v (defery běží v opačném pořadí registrace)", got, want)
	}
}

func TestDeferOrderIsStable(t *testing.T) {
	first := exercise.DeferOrder()
	for i := 0; i < 10; i++ {
		if got := exercise.DeferOrder(); !reflect.DeepEqual(got, first) {
			t.Fatalf("DeferOrder() vrátil %v, předtím %v", got, first)
		}
	}
}

func TestSumWithLog(t *testing.T) {
	tests := []struct {
		name      string
		nums      []int
		wantTotal int
		wantSteps []string
	}{
		{"empty", nil, 0, []string{"total=0"}},
		{"one number", []int{5}, 5, []string{"+5=5", "total=5"}},
		{"three numbers", []int{1, 2, 3}, 6, []string{"+1=1", "+2=3", "+3=6", "total=6"}},
		{"with negatives", []int{10, -4}, 6, []string{"+10=10", "+-4=6", "total=6"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, steps := exercise.SumWithLog(tt.nums)
			if total != tt.wantTotal {
				t.Errorf("SumWithLog(%v) total = %d, chci %d", tt.nums, total, tt.wantTotal)
			}
			if !reflect.DeepEqual(steps, tt.wantSteps) {
				t.Errorf("SumWithLog(%v) steps = %v, chci %v", tt.nums, steps, tt.wantSteps)
			}
		})
	}
}

func TestSumWithLogLastStepAddsDefer(t *testing.T) {
	// Kdyby defer neupravoval pojmenovanou návratovou hodnotu, poslední
	// krok by se do výsledku vůbec nedostal.
	_, steps := exercise.SumWithLog([]int{4, 4})
	if len(steps) == 0 {
		t.Fatal("SumWithLog nevrátil žádné kroky")
	}
	if last := steps[len(steps)-1]; last != "total=8" {
		t.Errorf("poslední krok = %q, chci %q", last, "total=8")
	}
}

func TestSafeDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    int
		want    int
		wantErr bool
	}{
		{"normal division", 10, 2, 5, false},
		{"integer truncates down", 7, 2, 3, false},
		{"negative divisor", -9, 3, -3, false},
		{"zero numerator", 0, 5, 0, false},
		{"divide by zero", 10, 0, 0, true},
		{"zero divided by zero", 0, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exercise.SafeDivide(tt.a, tt.b)
			if tt.wantErr && err == nil {
				t.Fatalf("SafeDivide(%d, %d) vrátil err = nil, chci chybu", tt.a, tt.b)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("SafeDivide(%d, %d) vrátil err = %v, chci nil", tt.a, tt.b, err)
			}
			if got != tt.want {
				t.Errorf("SafeDivide(%d, %d) = %d, chci %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSafeDivideDoesNotLeakPanic(t *testing.T) {
	// Kdyby recover chyběl, tenhle test shodí celý běh testů.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SafeDivide propustila paniku: %v", r)
		}
	}()
	if _, err := exercise.SafeDivide(1, 0); err == nil {
		t.Error("SafeDivide(1, 0) nevrátila chybu")
	}
}

// fakeWriteCloser simuluje io.WriteCloser pro TestWriteAndClose.
type fakeWriteCloser struct {
	buf      bytes.Buffer
	writeErr error
	closeErr error
	closed   int
}

func (f *fakeWriteCloser) Write(p []byte) (int, error) {
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	return f.buf.Write(p)
}

func (f *fakeWriteCloser) Close() error {
	f.closed++
	return f.closeErr
}

func TestWriteAndClose(t *testing.T) {
	errWrite := errors.New("write failed")
	errClose := errors.New("close failed")

	t.Run("success", func(t *testing.T) {
		f := &fakeWriteCloser{}
		data := []byte("ahoj")
		if err := exercise.WriteAndClose(f, data); err != nil {
			t.Fatalf("WriteAndClose = %v, chci nil", err)
		}
		if f.buf.String() != "ahoj" {
			t.Errorf("zapsáno %q, chci %q", f.buf.String(), "ahoj")
		}
		if f.closed != 1 {
			t.Errorf("Close volán %dx, chci 1x", f.closed)
		}
	})

	t.Run("write fails still closes", func(t *testing.T) {
		f := &fakeWriteCloser{writeErr: errWrite}
		err := exercise.WriteAndClose(f, []byte("x"))
		if !errors.Is(err, errWrite) {
			t.Errorf("WriteAndClose = %v, chci %v", err, errWrite)
		}
		if f.closed != 1 {
			t.Errorf("Close volán %dx, chci 1x — Close se musí zavolat i po chybě Write", f.closed)
		}
	})

	t.Run("close fails after successful write", func(t *testing.T) {
		f := &fakeWriteCloser{closeErr: errClose}
		err := exercise.WriteAndClose(f, []byte("ok"))
		if !errors.Is(err, errClose) {
			t.Errorf("WriteAndClose = %v, chci %v — chyba Close se nesmí ztratit", err, errClose)
		}
		if f.buf.String() != "ok" {
			t.Errorf("zapsáno %q, chci %q", f.buf.String(), "ok")
		}
		if f.closed != 1 {
			t.Errorf("Close volán %dx, chci 1x", f.closed)
		}
	})

	t.Run("write fails close fails keeps write error", func(t *testing.T) {
		f := &fakeWriteCloser{writeErr: errWrite, closeErr: errClose}
		err := exercise.WriteAndClose(f, []byte("x"))
		if !errors.Is(err, errWrite) {
			t.Errorf("WriteAndClose = %v, chci %v — při selhání Write má zůstat chyba Write", err, errWrite)
		}
		if errors.Is(err, errClose) {
			t.Error("WriteAndClose nesmí přepsat chybu Write chybou Close")
		}
		if f.closed != 1 {
			t.Errorf("Close volán %dx, chci 1x", f.closed)
		}
	})

	t.Run("nil writer", func(t *testing.T) {
		err := exercise.WriteAndClose(nil, []byte("x"))
		if err == nil {
			t.Fatal("WriteAndClose(nil) = nil, chci chybu")
		}
	})
}

func TestStackPushPop(t *testing.T) {
	var s exercise.Stack

	s.Push(1)
	s.Push(2)
	s.Push(3)

	for _, want := range []int{3, 2, 1} {
		if got := s.Pop(); got != want {
			t.Errorf("Pop() = %d, chci %d", got, want)
		}
	}

	// Další Pop na prázdném zásobníku paniká — ověříme, že Push/Pop
	// opravdu vyprázdnily zásobník, bez volání Len z pozdějšího stupně.
	defer func() {
		if recover() == nil {
			t.Error("Pop() po vyprázdnění nepanikoval")
		}
	}()
	s.Pop()
}

func TestStackPopOnEmptyPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Pop() na prázdném zásobníku nepanikoval")
		}
		if got, ok := r.(string); !ok || got != "pop from empty stack" {
			t.Errorf("panika s hodnotou %v, chci %q", r, "pop from empty stack")
		}
	}()

	var s exercise.Stack
	s.Pop()
}

func TestStackLenOnNilPointer(t *testing.T) {
	var s *exercise.Stack
	if got := s.Len(); got != 0 {
		t.Errorf("Len() na nil pointeru = %d, chci 0", got)
	}
}

func TestTryPop(t *testing.T) {
	var s exercise.Stack
	s.Push(42)

	v, ok := exercise.TryPop(&s)
	if !ok || v != 42 {
		t.Errorf("TryPop = (%d, %v), chci (42, true)", v, ok)
	}
}

func TestTryPopOnEmpty(t *testing.T) {
	var s exercise.Stack
	v, ok := exercise.TryPop(&s)
	if ok {
		t.Error("TryPop na prázdném zásobníku vrátil ok = true, chci false")
	}
	if v != 0 {
		t.Errorf("TryPop na prázdném zásobníku vrátil v = %d, chci 0", v)
	}
}

func TestTryPopOnNilPointer(t *testing.T) {
	// Nil dereference je taky panika a recover ji musí pobrat.
	v, ok := exercise.TryPop(nil)
	if ok || v != 0 {
		t.Errorf("TryPop(nil) = (%d, %v), chci (0, false)", v, ok)
	}
}

func TestStackUsableAfterRecover(t *testing.T) {
	var s exercise.Stack

	if _, ok := exercise.TryPop(&s); ok {
		t.Fatal("TryPop na prázdném zásobníku vrátil ok = true")
	}

	s.Push(7)
	s.Push(8)
	if got := s.Len(); got != 2 {
		t.Fatalf("Len() po zotavení = %d, chci 2", got)
	}

	if v, ok := exercise.TryPop(&s); !ok || v != 8 {
		t.Errorf("TryPop po zotavení = (%d, %v), chci (8, true)", v, ok)
	}
	if got := s.Pop(); got != 7 {
		t.Errorf("Pop() po zotavení = %d, chci 7", got)
	}
	if got := s.Len(); got != 0 {
		t.Errorf("Len() = %d, chci 0", got)
	}
}

func TestStackAlternatingOps(t *testing.T) {
	var s exercise.Stack
	for i := 0; i < 100; i++ {
		s.Push(i)
	}
	for i := 0; i < 50; i++ {
		if _, ok := exercise.TryPop(&s); !ok {
			t.Fatalf("TryPop selhal na %d. iteraci, ačkoli zásobník není prázdný", i)
		}
	}
	if got := s.Len(); got != 50 {
		t.Fatalf("Len() = %d, chci 50", got)
	}
	if got := s.Pop(); got != 49 {
		t.Errorf("Pop() = %d, chci 49", got)
	}
}
