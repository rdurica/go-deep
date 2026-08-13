// Package exercise obsahuje cvičení lekce 48 — paměťový model a happens-before.
package exercise

import (
	"sync"
)

// NaiveFlag je sdílený booleovský příznak bez synchronizace.
type NaiveFlag struct {
	v bool
}

// --- Stupeň: jednoduchý ---

// Set nastaví hodnotu příznaku.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Set a Get nejsou chráněné proti datovému závodu.
// Najdi chybu a oprav — testy s -race před opravou padají.
func (f *NaiveFlag) Set(v bool) {
	f.v = v
}

// Get vrátí aktuální hodnotu příznaku.
func (f *NaiveFlag) Get() bool {
	return f.v
}

// LazyInit je líná inicializace hodnoty. Inicializační funkce se musí zavolat
// právě jednou. Nulová hodnota není použitelná — vytvoř přes NewLazyInit.
type LazyInit struct {
	once  sync.Once
	init  func() int
	value int
}

// NewLazyInit vrátí líný inicializátor s danou funkci.
// Pro nil funkci panikuje.
func NewLazyInit(init func() int) *LazyInit {
	if init == nil {
		panic("NewLazyInit: init nesmí být nil")
	}
	return &LazyInit{init: init}
}

// Value vrátí hodnotu a při prvním volání ji nechá spočítat.
func (l *LazyInit) Value() int {
	l.once.Do(func() {
		l.value = l.init()
	})
	return l.value
}

// AtomicFlag je Flag postavený na atomické operaci.
type AtomicFlag struct {
	// TODO
}

// --- Stupeň: střední ---

// Set nastaví hodnotu atomického příznaku.
// Zápis je viditelný z Get v jiné goroutině bez další synchronizace.
func (f *AtomicFlag) Set(v bool) {
	// TODO
}

// Get vrátí aktuální hodnotu atomického příznaku.
func (f *AtomicFlag) Get() bool {
	// TODO
	return false
}

// Box publikuje data jedním zápisem a zpřístupní je čtenářům.
// Synchronizačním bodem je zavření kanálu ready.
type Box struct {
	data  []int
	ready chan struct{}
}

// NewBox vrátí prázdný box pro jednorázovou publikaci dat.
func NewBox() *Box {
	return &Box{ready: make(chan struct{})}
}

// --- Stupeň: obtížný ---

// Publish uloží data a oznámí to čtenářům zavřením synchronizačního kanálu.
// Volá se právě jednou; opakované volání panikuje.
func (b *Box) Publish(data []int) {
	// TODO
}

// Consume blokuje do Publish a pak vrací publikovaná data.
// Libovolný počet čtenářů po Publish dostane stejná kompletní data.
func (b *Box) Consume() []int {
	// TODO
	return nil
}
