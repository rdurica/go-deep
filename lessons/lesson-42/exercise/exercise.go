// Package exercise obsahuje cvičení lekce 42.
package exercise

import (
	"context"
	"time"
)

// TrySend se pokusí neblokujícím způsobem poslat v do ch. Vrací true, když
// se to povedlo, jinak false.
func TrySend(ch chan<- int, v int) bool {
	panic("TODO: úkol A")
}

// RecvWithTimeout přečte hodnotu z kanálu, nejdéle však po dobu d. Druhá
// návratová hodnota je false při timeoutu i při zavřeném kanálu.
func RecvWithTimeout(ch <-chan int, d time.Duration) (int, bool) {
	panic("TODO: úkol A")
}

// First spustí všechny funkce souběžně, vrátí první úspěšný výsledek a
// ostatní zruší. Když selžou všechny, vrátí jejich spojenou chybu.
func First(ctx context.Context, fns ...func(context.Context) (string, error)) (string, error) {
	panic("TODO: úkol B")
}

// Debounce propustí hodnotu, až když po dobu d nepřijde nic nového.
// Po zavření vstupu pošle poslední čekající hodnotu a výstup zavře.
func Debounce(in <-chan string, d time.Duration) <-chan string {
	panic("TODO: úkol B")
}

// Heartbeat každých interval zavolá work a pošle tep do vráceného kanálu.
// Po zrušení kontextu skončí a kanál zavře.
func Heartbeat(ctx context.Context, interval time.Duration, work func()) <-chan time.Time {
	panic("TODO: úkol C")
}
