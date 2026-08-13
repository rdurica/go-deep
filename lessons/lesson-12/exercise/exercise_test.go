package exercise_test

import (
	"errors"
	"fmt"
	"math"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-12/exercise"
)

const eps = 1e-9

// stubShape je lokální Shape, aby TotalArea nevolal studentské Area z pozdějšího PART.
type stubShape float64

func (s stubShape) Area() float64 { return float64(s) }

func TestTotalArea(t *testing.T) {
	tests := []struct {
		name string
		in   []exercise.Shape
		want float64
	}{
		{"nil slice", nil, 0},
		{"empty slice", []exercise.Shape{}, 0},
		{"jeden tvar", []exercise.Shape{stubShape(10)}, 10},
		{
			"mixed shapes",
			[]exercise.Shape{stubShape(10), stubShape(math.Pi)},
			10 + math.Pi,
		},
		{
			"nil element skipped",
			[]exercise.Shape{stubShape(10), nil, stubShape(1)},
			11,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.TotalArea(tt.in); math.Abs(got-tt.want) > eps {
				t.Errorf("TotalArea() = %v, chci %v", got, tt.want)
			}
		})
	}
}

func TestShapeIsImplicitlyImplemented(t *testing.T) {
	var _ exercise.Shape = exercise.Rect{}
	var _ exercise.Shape = exercise.Circle{}
}

type customStruct struct{ X int }

func TestDescribe(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, "nil"},
		{"int", 42, "int:42"},
		{"negative int", -7, "int:-7"},
		{"string", "ahoj", `string:"ahoj"`},
		{"empty string", "", `string:""`},
		{"bool true", true, "bool:true"},
		{"bool false", false, "bool:false"},
		{"int slice", []int{1, 2, 3}, "[]int:len=3"},
		{"empty int slice", []int{}, "[]int:len=0"},
		{"float", 1.5, "other:float64"},
		{"custom struct", customStruct{X: 1}, fmt.Sprintf("other:%T", customStruct{X: 1})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.Describe(tt.in); got != tt.want {
				t.Errorf("Describe(%#v) = %q, chci %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRecorder(t *testing.T) {
	var r exercise.Recorder

	if got := r.Messages(); len(got) != 0 {
		t.Errorf("Messages() na prázdném Recorderu = %v, chci prázdné", got)
	}

	if err := r.Notify("first"); err != nil {
		t.Fatalf("Notify() = %v, chci nil", err)
	}
	if err := r.Notify("druhá"); err != nil {
		t.Fatalf("Notify() = %v, chci nil", err)
	}

	got := r.Messages()
	want := []string{"first", "druhá"}
	if len(got) != len(want) {
		t.Fatalf("Messages() = %v, chci %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Messages()[%d] = %q, chci %q", i, got[i], want[i])
		}
	}

	got[0] = "podvrh"
	if again := r.Messages(); again[0] != "first" {
		t.Errorf("Messages() vrací vnitřní slice, změna zvenku se projevila: %v", again)
	}
}

func TestRecorderWithError(t *testing.T) {
	sentinel := errors.New("doručení selhalo")
	r := &exercise.Recorder{Err: sentinel}

	if err := r.Notify("zpráva"); !errors.Is(err, sentinel) {
		t.Errorf("Notify() = %v, chci %v", err, sentinel)
	}
	if got := r.Messages(); len(got) != 0 {
		t.Errorf("Messages() = %v, chci prázdné — chybující Notify nemá nic zaznamenat", got)
	}
}

func TestRecorderSatisfiesNotifier(t *testing.T) {
	var n exercise.Notifier = &exercise.Recorder{}
	if err := n.Notify("přes interface"); err != nil {
		t.Fatalf("Notify() = %v, chci nil", err)
	}
	rec, ok := n.(*exercise.Recorder)
	if !ok {
		t.Fatalf("type assertion na *Recorder selhala")
	}
	if got := rec.Messages(); len(got) != 1 || got[0] != "přes interface" {
		t.Errorf("Messages() = %v, chci [přes interface]", got)
	}
}

func TestNilPointerInInterface(t *testing.T) {
	s := exercise.ReturnsNilPointer()

	if s == nil {
		t.Fatal("ReturnsNilPointer() == nil, chci non-nil interface s nil pointerem uvnitř")
	}
	if exercise.IsNilInterface(s) {
		t.Error("IsNilInterface(ReturnsNilPointer()) = true, chci false")
	}

	p, ok := s.(*exercise.MyErr)
	if !ok {
		t.Fatalf("type assertion na *MyErr selhala, dynamický typ je %T", s)
	}
	if p != nil {
		t.Errorf("pointer uvnitř interfacu = %v, chci nil", p)
	}

	if got := s.Area(); got != 0 {
		t.Errorf("s.Area() = %v, chci 0", got)
	}
}

func TestIsNilInterface(t *testing.T) {
	if !exercise.IsNilInterface(nil) {
		t.Error("IsNilInterface(nil) = false, chci true")
	}
	if exercise.IsNilInterface(exercise.Rect{}) {
		t.Error("IsNilInterface(Rect{}) = true, chci false")
	}
	if exercise.IsNilInterface(&exercise.MyErr{}) {
		t.Error("IsNilInterface(&MyErr{}) = true, chci false")
	}
}

func TestRectAndCircleArea(t *testing.T) {
	if got := (exercise.Rect{W: 3, H: 4}).Area(); math.Abs(got-12) > eps {
		t.Errorf("Rect{3,4}.Area() = %v, chci 12", got)
	}
	if got := (exercise.Rect{}).Area(); math.Abs(got) > eps {
		t.Errorf("Rect{}.Area() = %v, chci 0", got)
	}
	want := math.Pi * 4
	if got := (exercise.Circle{R: 2}).Area(); math.Abs(got-want) > eps {
		t.Errorf("Circle{2}.Area() = %v, chci %v", got, want)
	}
}
