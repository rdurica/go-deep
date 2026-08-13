// Package exercise obsahuje cvičení lekce 01.
package exercise

import "fmt"

// --- Stupeň: jednoduchý ---

// Greet vrací pozdrav pro dané jméno.
// Prázdné jméno (i po ořezu bílých znaků) dá "Hello, Go!".
// Jinak vrátí "Hello, <name>!" s ořezanými bílými znaky.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Kontroluje prázdnost před ořezem mezer —
// řetězec jen z mezer se nesprávně nepovažuje za prázdný.
func Greet(name string) string {
	if name == "" {
		return "Hello, Go!"
	}
	return fmt.Sprintf("Hello, %s!", name)
}

// --- Stupeň: střední ---

// SumAll sečte libovolný počet celých čísel předaných jako argumenty.
// Bez argumentů vrátí 0. Záporná čísla sečte normálně.
func SumAll(nums ...int) int {
	// TODO
	return 0
}

// --- Stupeň: obtížný ---

// Describe popíše slice čísel.
// Nil nebo prázdný slice vrací přesně "empty".
// Jinak vrátí "count=<počet> sum=<součet> max=<maximum>" (formát musí sedět přesně).
func Describe(vals []int) string {
	// TODO
	return ""
}
