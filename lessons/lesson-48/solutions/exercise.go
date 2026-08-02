// Package solutions obsahuje referenční řešení lekce 48.
package solutions

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
func (f *AtomicFlag) Set(v bool) { f.v.Store(v) }

// Get vrátí aktuální hodnotu příznaku.
func (f *AtomicFlag) Get() bool { return f.v.Load() }

// MutexFlag je Flag chráněný zámkem. Unlock jednoho zápisu se řadí před
// Lock následujícího čtení, takže i tady je viditelnost zaručená.
type MutexFlag struct {
	mu sync.Mutex
	v  bool
}

// Set nastaví hodnotu příznaku.
func (f *MutexFlag) Set(v bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.v = v
}

// Get vrátí aktuální hodnotu příznaku.
func (f *MutexFlag) Get() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.v
}

// --- Stupeň: střední ---
// StressFlag prožene příznak zátěží: writers goroutin ho nastavuje na true,
// readers goroutin ho čte, každá z nich iterations krát. Funkce počká na
// všechny a vrátí, kolikrát čtenáři viděli true.
//
// Pro nil příznak nebo nekladné počty vrací 0 a nic nespouští.
func StressFlag(f Flag, writers, readers, iterations int) int {
	if f == nil || iterations <= 0 || writers <= 0 || readers <= 0 {
		return 0
	}

	var trueReads atomic.Int64
	var wg sync.WaitGroup

	wg.Add(writers + readers)
	for w := 0; w < writers; w++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				f.Set(true)
			}
		}()
	}
	for r := 0; r < readers; r++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				if f.Get() {
					trueReads.Add(1)
				}
			}
		}()
	}
	wg.Wait()

	return int(trueReads.Load())
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
	if init == nil {
		panic("NewLazyInit: init nesmí být nil")
	}
	return &LazyInit{init: init}
}

// Value vrátí hodnotu a při prvním volání ji nechá spočítat.
// Všechna volání musí vidět stejnou hodnotu.
func (l *LazyInit) Value() int {
	// Once.Do dělá obě věci správného double-checked lockingu najednou:
	// rychlou cestu bez zámku a happens-before mezi zápisem uvnitř Do
	// a návratem ze všech ostatních volání.
	l.once.Do(func() {
		l.value = l.init()
	})
	return l.value
}

// ConcurrentValues zavolá l.Value() z n goroutin najednou a vrátí, co která
// viděla. Pořadí odpovídá pořadí goroutin. Pro n <= 0 vrací prázdný výsledek.
func ConcurrentValues(l *LazyInit, n int) []int {
	if l == nil || n <= 0 {
		return nil
	}

	out := make([]int, n)
	start := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			<-start // srovnáme startovní čáru, ať do Value spadnou opravdu naráz
			out[i] = l.Value()
		}()
	}
	close(start)
	wg.Wait()

	return out
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
// NewBox vrátí prázdný box, ze kterého se dá číst až po Publish.
func NewBox() *Box {
	return &Box{ready: make(chan struct{})}
}

// Publish uloží data a oznámí to čtenářům. Volá se právě jednou.
func (b *Box) Publish(data []int) {
	b.data = data
	close(b.ready) // close je broadcast a zároveň synchronizační bod
}

// Consume počká na Publish a vrátí publikovaná data.
func (b *Box) Consume() []int {
	<-b.ready
	return b.data
}

// PublishAndConsume publikuje data z jedné goroutiny a přečte je z readers
// goroutin. Vrátí, co který čtenář viděl — všichni musí vidět kompletní data.
// Pro readers <= 0 vrací prázdný výsledek.
func PublishAndConsume(data []int, readers int) [][]int {
	if readers <= 0 {
		return nil
	}

	b := NewBox()
	out := make([][]int, readers)
	var wg sync.WaitGroup

	wg.Add(readers)
	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()
			out[i] = b.Consume()
		}()
	}

	go b.Publish(data)

	wg.Wait()
	return out
}

// WaitGroupVisibility spustí n goroutin, každá zapíše i*i na svůj index
// sdíleného slice. Po Wait jsou všechny zápisy viditelné, takže se dají
// bez dalšího zamykání sečíst. Vrací ten součet, pro n <= 0 nulu.
func WaitGroupVisibility(n int) int {
	if n <= 0 {
		return 0
	}

	out := make([]int, n)
	var wg sync.WaitGroup

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			out[i] = i * i
		}()
	}
	wg.Wait() // Done -> Wait je synchronizační bod, zápisy jsou od téhle chvíle vidět

	sum := 0
	for _, v := range out {
		sum += v
	}
	return sum
}
