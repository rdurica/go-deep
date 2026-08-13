// Package solutions obsahuje referenční řešení lekce 48.
package solutions

import (
	"sync"
	"sync/atomic"
)

// NaiveFlag je sdílený booleovský příznak bez synchronizace (vadný vzor).
type NaiveFlag struct {
	v atomic.Bool
}

// --- Stupeň: jednoduchý ---

// Set nastaví hodnotu příznaku.
func (f *NaiveFlag) Set(v bool) { f.v.Store(v) }

// Get vrátí aktuální hodnotu příznaku.
func (f *NaiveFlag) Get() bool { return f.v.Load() }

// AtomicFlag je Flag postavený na atomické operaci.
type AtomicFlag struct {
	v atomic.Bool
}

// --- Stupeň: střední ---

// Set nastaví hodnotu atomického příznaku.
func (f *AtomicFlag) Set(v bool) { f.v.Store(v) }

// Get vrátí aktuální hodnotu atomického příznaku.
func (f *AtomicFlag) Get() bool { return f.v.Load() }

// LazyInit je líná inicializace hodnoty.
type LazyInit struct {
	once  sync.Once
	init  func() int
	value int
}

// NewLazyInit vrátí líný inicializátor.
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

// Box publikuje data jedním zápisem.
type Box struct {
	data  []int
	ready chan struct{}
}

// --- Stupeň: obtížný ---

// NewBox vrátí prázdný box.
func NewBox() *Box {
	return &Box{ready: make(chan struct{})}
}

// Publish uloží data a oznámí to čtenářům.
func (b *Box) Publish(data []int) {
	b.data = data
	close(b.ready)
}

// Consume počká na Publish a vrátí publikovaná data.
func (b *Box) Consume() []int {
	<-b.ready
	return b.data
}
