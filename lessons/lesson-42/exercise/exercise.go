// Package exercise obsahuje cvičení lekce 42.
package exercise

import (
	"context"
	"time"
)

// TrySend se pokusí neblokujícím způsobem poslat v do ch. Vrací true, když
// se to povedlo, jinak false.
func TrySend(ch chan<- int, v int) bool {
	// TODO: úkol A
	return false
}

// RecvWithTimeout přečte hodnotu z kanálu, nejdéle však po dobu d. Druhá
// návratová hodnota je false při timeoutu i při zavřeném kanálu.
func RecvWithTimeout(ch <-chan int, d time.Duration) (int, bool) {
	// TODO: úkol A
	return 0, false
}

// First spustí všechny funkce souběžně, vrátí první úspěšný výsledek a
// ostatní zruší. Když selžou všechny, vrátí jejich spojenou chybu.
func First(ctx context.Context, fns ...func(context.Context) (string, error)) (string, error) {
	// TODO: úkol B
	return "", nil
}

// Debounce propustí hodnotu, až když po dobu d nepřijde nic nového.
// Po zavření vstupu pošle poslední čekající hodnotu a výstup zavře.
func Debounce(in <-chan string, d time.Duration) <-chan string {
	// TODO: úkol B
	ch := make(chan string)
	close(ch)
	return ch
}

// Heartbeat každých interval zavolá work a pošle tep do vráceného kanálu.
// Po zrušení kontextu skončí a kanál zavře.
func Heartbeat(ctx context.Context, interval time.Duration, work func()) <-chan time.Time {
	// TODO: úkol C
	ch := make(chan time.Time)
	close(ch)
	return ch
}
