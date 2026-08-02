// Package exercise obsahuje cvičení lekce 42.
package exercise

import (
	"context"
	"time"
)

// --- Stupeň: jednoduchý ---
// TrySend neblokujícím select+default pošle v do ch. Plný, nil nebo bez čtenáře → false.
func TrySend(ch chan<- int, v int) bool {
	// TODO
	return false
}

// --- Stupeň: střední ---
// RecvWithTimeout přečte z ch nejdéle po d. Timeout i zavřený kanál → (0, false).
// Použij NewTimer a defer Stop(), ne time.After.
func RecvWithTimeout(ch <-chan int, d time.Duration) (int, bool) {
	// TODO
	return 0, false
}

// First spustí všechny fns souběžně, vrátí první úspěch, poražené zruší.
// Počkej na doběhnutí všech goroutin. Všechny selžou → errors.Join.
// Bez funkcí vrať chybu.
func First(ctx context.Context, fns ...func(context.Context) (string, error)) (string, error) {
	// TODO
	return "", nil
}

// --- Stupeň: obtížný ---
// Debounce propustí hodnotu až po d ticha; z dávky projde poslední.
// Po zavření vstupu doruč čekající hodnotu a zavři výstup. Timer resetuj správně.
func Debounce(in <-chan string, d time.Duration) <-chan string {
	// TODO
	ch := make(chan string)
	close(ch)
	return ch
}

// Heartbeat každých interval zavolá work (může být nil) a pošle tep.
// Po zrušení ctx skončí a kanál zavře; musí skončit i bez čtenáře (select+ctx).
// interval <= 0 jako 1 ms; už zrušený ctx → okamžitý konec bez tepu.
func Heartbeat(ctx context.Context, interval time.Duration, work func()) <-chan time.Time {
	// TODO
	ch := make(chan time.Time)
	close(ch)
	return ch
}
