// Package exercise obsahuje cvičení lekce 45.
package exercise

import (
	"context"
	"errors"
)

// Chyby, které vznikají ve stupních Pipeline.
var (
	// ErrEmpty znamená prázdný vstup po ořezání bílých znaků.
	ErrEmpty = errors.New("empty input")
	// ErrDigits znamená, že vstup obsahuje číslici.
	ErrDigits = errors.New("input contains digits")
)

// Result je výsledek jednoho prvku pipeline. Chyba se nese s ním, ne mimo něj.
type Result struct {
	Input string
	Value string
	Err   error
}

// Gen pošle čísla do kanálu a zavře ho. Při zrušení kontextu skončí dřív.
func Gen(ctx context.Context, nums ...int) <-chan int {
	// TODO: úkol A
	return closed[int]()
}

// Square umocní každou hodnotu ze vstupu. Výstup zavírá, vstup nikdy.
func Square(ctx context.Context, in <-chan int) <-chan int {
	// TODO: úkol A
	return closed[int]()
}

// Stage je obecný stupeň pipeline: na každý prvek použije f.
func Stage[T, U any](ctx context.Context, in <-chan T, f func(T) U) <-chan U {
	// TODO: úkol B
	return closed[U]()
}

// FanIn sloučí několik kanálů do jednoho. Výstup se zavře po vyčerpání všech
// vstupů nebo po zrušení kontextu.
func FanIn[T any](ctx context.Context, chs ...<-chan T) <-chan T {
	// TODO: úkol B
	return closed[T]()
}

// Take propustí nejvýše n prvků a pak výstup zavře.
func Take[T any](ctx context.Context, in <-chan T, n int) <-chan T {
	// TODO: úkol B
	return closed[T]()
}

// Pipeline složí tři stupně (normalizace, obohacení s fan-outem, formátování)
// a vrátí kanál výsledků. Chyby se nesou uvnitř Result.
func Pipeline(ctx context.Context, in <-chan string) <-chan Result {
	// TODO: úkol C
	return closed[Result]()
}

// closed je fail-fast stub: nil kanál by v testech visel navždy.
func closed[T any]() <-chan T {
	ch := make(chan T)
	close(ch)
	return ch
}
