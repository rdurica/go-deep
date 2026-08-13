package solutions_test

import (
	"math/rand"
	"reflect"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-07/solutions"
)

func TestGrowReturnsOriginalWhenCapacityEnough(t *testing.T) {
	s := make([]int, 2, 8)
	s[0], s[1] = 10, 20

	got := exercise.Grow(s, 5)

	if len(got) != 2 || cap(got) != 8 {
		t.Fatalf("Grow(len=2 cap=8, 5) = len %d cap %d, chci len 2 cap 8", len(got), cap(got))
	}
	if &got[0] != &s[0] {
		t.Error("Grow měl vrátit původní slice beze změny, ale alokoval nové pole")
	}
}

func TestGrowAllocatesWhenCapacityInsufficient(t *testing.T) {
	s := []int{1, 2, 3}
	got := exercise.Grow(s, 10)

	if len(got) != 3 {
		t.Errorf("Grow zachovává len: len(got) = %d, chci 3", len(got))
	}
	if cap(got) < 10 {
		t.Errorf("cap(got) = %d, chci alespoň 10", cap(got))
	}
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("Grow ztratil data: got = %v, chci [1 2 3]", got)
	}

	got[0] = 99
	if s[0] != 1 {
		t.Error("nově alokovaný slice sdílí backing pole s originálem")
	}
}

func TestGrowNil(t *testing.T) {
	got := exercise.Grow(nil, 4)
	if len(got) != 0 {
		t.Errorf("Grow(nil, 4) má len %d, chci 0", len(got))
	}
	if cap(got) < 4 {
		t.Errorf("Grow(nil, 4) má cap %d, chci alespoň 4", cap(got))
	}
}

func TestSum(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"nil", nil, 0},
		{"empty", []int{}, 0},
		{"one element", []int{7}, 7},
		{"positive", []int{1, 2, 3, 4}, 10},
		{"with negatives", []int{-5, 5, -2}, -2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.Sum(tt.in); got != tt.want {
				t.Errorf("Sum(%v) = %d, chci %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestSumRandomData(t *testing.T) {
	nums := make([]int, 100)
	want := 0
	for i := range nums {
		nums[i] = rand.Intn(2000) - 1000
		want += nums[i]
	}
	if got := exercise.Sum(nums); got != want {
		t.Errorf("Sum(náhodných 100 čísel) = %d, chci %d", got, want)
	}
}

func TestRemoveAt(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		i    int
		want []int
	}{
		{"first", []int{1, 2, 3, 4}, 0, []int{2, 3, 4}},
		{"middle", []int{1, 2, 3, 4}, 1, []int{1, 3, 4}},
		{"last", []int{1, 2, 3, 4}, 3, []int{1, 2, 3}},
		{"single element", []int{9}, 0, []int{}},
		{"index too large", []int{1, 2, 3}, 3, []int{1, 2, 3}},
		{"negative index", []int{1, 2, 3}, -1, []int{1, 2, 3}},
		{"empty slice", []int{}, 0, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exercise.RemoveAt(tt.in, tt.i)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveAt(%v, %d) = %v, chci %v", tt.in, tt.i, got, tt.want)
			}
		})
	}
}

func TestClone(t *testing.T) {
	orig := []int{1, 2, 3}
	got := exercise.Clone(orig)

	if !reflect.DeepEqual(got, orig) {
		t.Fatalf("Clone(%v) = %v, chci stejný obsah", orig, got)
	}

	got[0] = 99
	if orig[0] != 1 {
		t.Error("zápis do klonu je vidět v originálu — kopie není nezávislá")
	}
	orig[2] = 77
	if got[2] != 3 {
		t.Error("zápis do originálu je vidět v klonu — kopie není nezávislá")
	}
}

func TestCloneNil(t *testing.T) {
	if got := exercise.Clone(nil); got != nil {
		t.Errorf("Clone(nil) = %v, chci nil", got)
	}
}

func TestCloneEmpty(t *testing.T) {
	got := exercise.Clone([]int{})
	if got == nil {
		t.Error("Clone([]int{}) = nil, chci prázdný ne-nil slice")
	}
	if len(got) != 0 {
		t.Errorf("Clone([]int{}) má len %d, chci 0", len(got))
	}
}
