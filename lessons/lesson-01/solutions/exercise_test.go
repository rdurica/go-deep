package solutions_test

import (
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-01/solutions"
)

func TestGreet(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"name", "Radek", "Hello, Radek!"},
		{"empty", "", "Hello, Go!"},
		{"jen mezery", "   ", "Hello, Go!"},
		{"trim", "  Radek \t", "Hello, Radek!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.Greet(tt.input); got != tt.want {
				t.Errorf("Greet(%q) = %q, chci %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSumAll(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"no args", nil, 0},
		{"jeden", []int{5}, 5},
		{"more", []int{1, 2, 3}, 6},
		{"negative", []int{-4, 2}, -2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.SumAll(tt.in...); got != tt.want {
				t.Errorf("SumAll(%v) = %d, chci %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want string
	}{
		{"nil", nil, "empty"},
		{"empty", []int{}, "empty"},
		{"jeden", []int{7}, "count=1 sum=7 max=7"},
		{"more", []int{1, 2, 3}, "count=3 sum=6 max=3"},
		{"max is not last", []int{9, 2, 3}, "count=3 sum=14 max=9"},
		{"negative", []int{-5, -2}, "count=2 sum=-7 max=-2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.Describe(tt.in); got != tt.want {
				t.Errorf("Describe(%v) = %q, chci %q", tt.in, got, tt.want)
			}
		})
	}
}
