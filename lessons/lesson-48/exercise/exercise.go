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

// Set nastaví hodnotu příznaku.
func (f *AtomicFlag) Set(v bool) {
	// TODO: úkol A
}

// Get vrátí aktuální hodnotu příznaku.
func (f *AtomicFlag) Get() bool {
	// TODO: úkol A
	return false
}

// MutexFlag je Flag chráněný zámkem. Unlock jednoho zápisu se řadí před
// Lock následujícího čtení, takže i tady je viditelnost zaručená.
type MutexFlag struct {
	mu sync.Mutex
	v  bool
}

// Set nastaví hodnotu příznaku.
func (f *MutexFlag) Set(v bool) {
	// TODO: úkol A
}

// Get vrátí aktuální hodnotu příznaku.
func (f *MutexFlag) Get() bool {
	// TODO: úkol A
	return false
}

// StressFlag prožene příznak zátěží: writers goroutin ho nastavuje na true,
// readers goroutin ho čte, každá z nich iterations krát. Funkce počká na
// všechny a vrátí, kolikrát čtenáři viděli true.
//
// Pro nil příznak nebo nekladné počty vrací 0 a nic nespouští.
func StressFlag(f Flag, writers, readers, iterations int) int {
	// TODO: úkol A
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

// NewLazyInit vrátí líný inicializátor. Pro nil funkci panikuje.
func NewLazyInit(init func() int) *LazyInit {
	// TODO: úkol B
	return nil
}

// Value vrátí hodnotu a při prvním volání ji nechá spočítat.
// Všechna volání musí vidět stejnou hodnotu.
func (l *LazyInit) Value() int {
	// TODO: úkol B
	return 0
}

// ConcurrentValues zavolá l.Value() z n goroutin najednou a vrátí, co která
// viděla. Pořadí odpovídá pořadí goroutin. Pro n <= 0 vrací prázdný výsledek.
func ConcurrentValues(l *LazyInit, n int) []int {
	// TODO: úkol B
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

// NewBox vrátí prázdný box, ze kterého se dá číst až po Publish.
func NewBox() *Box {
	// TODO: úkol C
	return nil
}

// Publish uloží data a oznámí to čtenářům. Volá se právě jednou.
func (b *Box) Publish(data []int) {
	// TODO: úkol C
}

// Consume počká na Publish a vrátí publikovaná data.
func (b *Box) Consume() []int {
	// TODO: úkol C
	return nil
}

// PublishAndConsume publikuje data z jedné goroutiny a přečte je z readers
// goroutin. Vrátí, co který čtenář viděl — všichni musí vidět kompletní data.
// Pro readers <= 0 vrací prázdný výsledek.
func PublishAndConsume(data []int, readers int) [][]int {
	// TODO: úkol C
	return nil
}

// WaitGroupVisibility spustí n goroutin, každá zapíše i*i na svůj index
// sdíleného slice. Po Wait jsou všechny zápisy viditelné, takže se dají
// bez dalšího zamykání sečíst. Vrací ten součet, pro n <= 0 nulu.
func WaitGroupVisibility(n int) int {
	// TODO: úkol C
	return 0
}
