// Package exercise obsahuje cvičení lekce 48 — paměťový model a happens-before.
package exercise

import (
	"sync"
	"sync/atomic"
)

// Flag je sdílený booleovský příznak, ke kterému sahá víc goroutin najednou.
type Flag interface {
	Set(v bool)
	Get() bool
}

// AtomicFlag je Flag postavený na atomické operaci.
// Atomika neřeší jen nedělitelnost zápisu, ale hlavně viditelnost:
// hodnota zapsaná přes Store je zaručeně vidět z Load v jiné goroutině.
type AtomicFlag struct {
	v atomic.Bool
}

// --- Stupeň: jednoduchý ---
// Set nastaví hodnotu příznaku.
// Zápis je viditelný z Load v jiné goroutině bez další synchronizace.
func (f *AtomicFlag) Set(v bool) {
	// TODO
}

// Get vrátí aktuální hodnotu příznaku.
// Čtení je atomické a vidí poslední Set z jiné goroutiny.
func (f *AtomicFlag) Get() bool {
	// TODO
	return false
}

// MutexFlag je Flag chráněný zámkem; zamykej i čtení, ne jen zápis.
type MutexFlag struct {
	mu sync.Mutex
	v  bool
}

// Set nastaví hodnotu příznaku.
// Zamykej i čtení — Get musí držet stejný mutex.
func (f *MutexFlag) Set(v bool) {
	// TODO
}

// Get vrátí aktuální hodnotu příznaku.
// Čtení pod zámkem; nesmí číst v bez mutexu.
func (f *MutexFlag) Get() bool {
	// TODO
	return false
}

// --- Stupeň: střední ---
// StressFlag prožene příznak zátěží: writers goroutin ho nastavuje na true,
// readers goroutin ho čte, každá z nich iterations krát. Funkce počká na
// všechny a vrátí, kolikrát čtenáři viděli true.
//
// Pro nil příznak nebo nekladné počty vrací 0 a nic nespouští.
func StressFlag(f Flag, writers, readers, iterations int) int {
	// TODO
	return 0
}

// LazyInit je líná inicializace hodnoty. Inicializační funkce se musí zavolat
// právě jednou, i když Value volá sto goroutin současně.
// Nulová hodnota není použitelná, vytvoř ho přes NewLazyInit.
type LazyInit struct {
	once  sync.Once
	init  func() int
	value int
}

// NewLazyInit vrátí líný inicializátor s danou funkci.
// Pro nil funkci panikuje. Nulová hodnota LazyInit není použitelná.
func NewLazyInit(init func() int) *LazyInit {
	// TODO
	return nil
}

// Value vrátí hodnotu a při prvním volání ji nechá spočítat.
// Všechna volání musí vidět stejnou hodnotu.
func (l *LazyInit) Value() int {
	// TODO
	return 0
}

// ConcurrentValues zavolá Value() z n goroutin najednou; pořadí odpovídá pořadí goroutin.
// Startovní čáru srovnej zavřením sdíleného kanálu. Pro n <= 0 prázdný výsledek.
func ConcurrentValues(l *LazyInit, n int) []int {
	// TODO
	return nil
}

// Box publikuje data jedním zápisem a zpřístupní je libovolnému počtu čtenářů.
// Synchronizačním bodem je zavření kanálu: co bylo zapsáno před close,
// je vidět po každém úspěšném příjmu z něj.
// Nulová hodnota není použitelná, vytvoř ho přes NewBox.
type Box struct {
	data  []int
	ready chan struct{}
}

// --- Stupeň: obtížný ---
// NewBox vrátí prázdný box pro jednorázovou publikaci dat.
// Consume blokuje do Publish; nulová hodnota Box není použitelná.
func NewBox() *Box {
	// TODO
	return nil
}

// Publish uloží data a oznámí to čtenářům zavřením synchronizačního kanálu.
// Volá se právě jednou; opakované volání panikuje.
func (b *Box) Publish(data []int) {
	// TODO
}

// Consume blokuje do Publish a pak vrací publikovaná data okamžitě.
// Libovolný počet čtenářů po Publish dostane stejná kompletní data.
func (b *Box) Consume() []int {
	// TODO
	return nil
}

// PublishAndConsume publikuje data z jedné goroutiny a přečte je z readers
// goroutin. Vrátí, co který čtenář viděl — všichni musí vidět kompletní data.
// Pro readers <= 0 vrací prázdný výsledek.
func PublishAndConsume(data []int, readers int) [][]int {
	// TODO
	return nil
}

// WaitGroupVisibility spustí n goroutin, každá zapíše i*i na svůj index
// sdíleného slice. Po Wait jsou všechny zápisy viditelné, takže se dají
// bez dalšího zamykání sečíst. Vrací ten součet, pro n <= 0 nulu.
func WaitGroupVisibility(n int) int {
	// TODO
	return 0
}
