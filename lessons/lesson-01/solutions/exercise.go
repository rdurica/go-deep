// Package solutions obsahuje referenční řešení lekce 01.
package solutions

import (
	"fmt"
	"strings"
)

// --- Stupeň: jednoduchý ---

// Greet vrací pozdrav pro dané jméno. Prázdné jméno (i po ořezu bílých znaků)
// dá "Hello, Go!".
func Greet(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Go"
	}
	return fmt.Sprintf("Hello, %s!", name)
}

// --- Stupeň: střední ---
// SumAll sečte libovolný počet celých čísel.
func SumAll(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// --- Stupeň: obtížný ---
// Describe popíše slice ve tvaru "count=3 sum=6 max=3".
// Prázdný nebo nil slice vrací "empty".
func Describe(vals []int) string {
	if len(vals) == 0 {
		return "empty"
	}
	max := vals[0]
	for _, v := range vals[1:] {
		if v > max {
			max = v
		}
	}
	return fmt.Sprintf("count=%d sum=%d max=%d", len(vals), SumAll(vals...), max)
}
