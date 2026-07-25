package solutions_test

import (
	"math/rand"
	"reflect"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-07/solutions"
)

func TestSum(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"nil", nil, 0},
		{"prázdný", []int{}, 0},
		{"jeden prvek", []int{7}, 7},
		{"kladné", []int{1, 2, 3, 4}, 10},
		{"se zápornými", []int{-5, 5, -2}, -2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.Sum(tt.in); got != tt.want {
				t.Errorf("Sum(%v) = %d, chci %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestSumNahodnaData(t *testing.T) {
	// Náhodná data brání tomu, aby test prošel se zadrátovanou hodnotou.
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

func TestReverse(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"prázdný", []int{}, []int{}},
		{"jeden prvek", []int{1}, []int{1}},
		{"sudá délka", []int{1, 2, 3, 4}, []int{4, 3, 2, 1}},
		{"lichá délka", []int{1, 2, 3}, []int{3, 2, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exercise.Reverse(tt.in)
			if !reflect.DeepEqual(tt.in, tt.want) {
				t.Errorf("po Reverse je slice %v, chci %v", tt.in, tt.want)
			}
		})
	}
}

func TestReverseNilNepanikuje(t *testing.T) {
	exercise.Reverse(nil)
}

func TestReverseMutujeVstup(t *testing.T) {
	// Reverse nic nevrací, takže musí měnit backing pole volajícího.
	nums := []int{1, 2, 3}
	view := nums[:2]
	exercise.Reverse(nums)
	if view[0] != 3 {
		t.Errorf("Reverse nezměnil sdílené backing pole: view = %v, chci [3 2]", view)
	}
}

func TestGrowVraciPuvodniSliceKdyzKapacitaStaci(t *testing.T) {
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

func TestGrowAlokujeKdyzKapacitaNestaci(t *testing.T) {
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

func TestRemoveAt(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		i    int
		want []int
	}{
		{"první", []int{1, 2, 3, 4}, 0, []int{2, 3, 4}},
		{"prostřední", []int{1, 2, 3, 4}, 1, []int{1, 3, 4}},
		{"poslední", []int{1, 2, 3, 4}, 3, []int{1, 2, 3}},
		{"jediný prvek", []int{9}, 0, []int{}},
		{"index moc velký", []int{1, 2, 3}, 3, []int{1, 2, 3}},
		{"záporný index", []int{1, 2, 3}, -1, []int{1, 2, 3}},
		{"prázdný slice", []int{}, 0, []int{}},
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

func TestRemoveFast(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		i    int
		want []int
	}{
		{"první", []int{1, 2, 3, 4}, 0, []int{4, 2, 3}},
		{"prostřední", []int{1, 2, 3, 4}, 1, []int{1, 4, 3}},
		{"poslední", []int{1, 2, 3, 4}, 3, []int{1, 2, 3}},
		{"jediný prvek", []int{9}, 0, []int{}},
		{"index moc velký", []int{1, 2, 3}, 5, []int{1, 2, 3}},
		{"záporný index", []int{1, 2, 3}, -2, []int{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exercise.RemoveFast(tt.in, tt.i)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveFast(%v, %d) = %v, chci %v", tt.in, tt.i, got, tt.want)
			}
		})
	}
}

func TestRemoveFastNealokuje(t *testing.T) {
	s := []int{1, 2, 3, 4}
	got := exercise.RemoveFast(s, 0)
	if len(got) > 0 && &got[0] != &s[0] {
		t.Error("RemoveFast má pracovat nad původním backing polem, ne alokovat nové")
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

func TestClonePrazdny(t *testing.T) {
	got := exercise.Clone([]int{})
	if got == nil {
		t.Error("Clone([]int{}) = nil, chci prázdný ne-nil slice")
	}
	if len(got) != 0 {
		t.Errorf("Clone([]int{}) má len %d, chci 0", len(got))
	}
}

func TestChunk(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		size int
		want [][]int
	}{
		{"beze zbytku", []int{1, 2, 3, 4}, 2, [][]int{{1, 2}, {3, 4}}},
		{"se zbytkem", []int{1, 2, 3, 4, 5}, 2, [][]int{{1, 2}, {3, 4}, {5}}},
		{"size větší než délka", []int{1, 2}, 10, [][]int{{1, 2}}},
		{"size 1", []int{1, 2, 3}, 1, [][]int{{1}, {2}, {3}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exercise.Chunk(tt.in, tt.size)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Chunk(%v, %d) = %v, chci %v", tt.in, tt.size, got, tt.want)
			}
		})
	}
}

func TestChunkHraniceniPripady(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		size int
	}{
		{"size 0", []int{1, 2, 3}, 0},
		{"záporný size", []int{1, 2, 3}, -1},
		{"prázdný vstup", []int{}, 2},
		{"nil vstup", nil, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.Chunk(tt.in, tt.size); len(got) != 0 {
				t.Errorf("Chunk(%v, %d) = %v, chci výsledek nulové délky", tt.in, tt.size, got)
			}
		})
	}
}

func TestChunkVraciNezavisleKopie(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6}
	got := exercise.Chunk(in, 2)
	if len(got) != 3 {
		t.Fatalf("Chunk vrátil %d chunků, chci 3", len(got))
	}

	got[0][0] = 99
	got[0][1] = 98

	if in[0] != 1 || in[1] != 2 {
		t.Errorf("zápis do chunku přepsal vstup: in = %v, chci [1 2 3 4 5 6]", in)
	}
	if !reflect.DeepEqual(got[1], []int{3, 4}) {
		t.Errorf("zápis do chunku 0 ovlivnil chunk 1: %v, chci [3 4]", got[1])
	}

	// Rozšíření jednoho chunku nesmí zasáhnout do sousedního.
	got[1] = append(got[1], 42)
	if !reflect.DeepEqual(got[2], []int{5, 6}) {
		t.Errorf("append do chunku 1 přepsal chunk 2: %v, chci [5 6]", got[2])
	}
}

func even(n int) bool { return n%2 == 0 }

func TestFilter(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		keep func(int) bool
		want []int
	}{
		{"sudá čísla", []int{1, 2, 3, 4, 5, 6}, even, []int{2, 4, 6}},
		{"nic neprojde", []int{1, 3, 5}, even, []int{}},
		{"všechno projde", []int{2, 4}, even, []int{2, 4}},
		{"prázdný vstup", []int{}, even, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exercise.Filter(tt.in, tt.keep)
			if len(got) != len(tt.want) {
				t.Fatalf("Filter vrátil %v, chci %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("Filter vrátil %v, chci %v", got, tt.want)
				}
			}
		})
	}
}

func TestFilterNil(t *testing.T) {
	if got := exercise.Filter(nil, even); len(got) != 0 {
		t.Errorf("Filter(nil, even) = %v, chci výsledek nulové délky", got)
	}
}

func TestFilterPrepisujeVstup(t *testing.T) {
	// Trik s[:0] skládá výsledek do backing pole vstupu. Test to schválně
	// odhaluje: implementace přes make() by tímhle testem neprošla.
	in := []int{1, 2, 3, 4, 5, 6}
	got := exercise.Filter(in, even)

	if len(got) != 3 {
		t.Fatalf("Filter vrátil %v, chci [2 4 6]", got)
	}
	if &got[0] != &in[0] {
		t.Fatal("Filter alokoval nový slice, chci trik s[:0] nad vstupním polem")
	}
	for i, want := range []int{2, 4, 6} {
		if in[i] != want {
			t.Errorf("vstup po Filter má in[%d] = %d, chci %d (výsledek se skládá do vstupu)", i, in[i], want)
		}
	}
}
