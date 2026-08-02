// Package solutions obsahuje referenční řešení lekce 41.
package solutions

import "sync"

// --- Stupeň: jednoduchý ---
// Generate pošle všechna čísla do vráceného kanálu a kanál sám zavře.
func Generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		// Kanál zavírá ten, kdo do něj zapisuje. Bez toho by Collect
		// na rangi visel navždy.
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// Collect přečte kanál až do zavření a vrátí hodnoty v pořadí, v jakém dorazily.
func Collect(ch <-chan int) []int {
	out := []int{}
	for v := range ch { // range končí sám ve chvíli, kdy je kanál zavřený
		out = append(out, v)
	}
	return out
}

// Merge sloučí několik vstupních kanálů do jednoho (fan-in).
func Merge(chs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(chs))
	for _, ch := range chs {
		go func() {
			defer wg.Done()
			for v := range ch {
				out <- v
			}
		}()
	}
	// Výstup zavíráme právě jednou, až doběhnou všichni odesílatelé.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// --- Stupeň: střední ---
// Split rozdělí hodnoty ze vstupního kanálu mezi n výstupních kanálů.
func Split(ch <-chan int, n int) []<-chan int {
	if n < 1 {
		n = 1
	}
	outs := make([]<-chan int, n)
	for i := 0; i < n; i++ {
		out := make(chan int)
		outs[i] = out
		go func() {
			defer close(out)
			// Všechny goroutiny čtou ze stejného kanálu, takže hodnotu
			// dostane vždy právě jedna z nich. Rozdělení je na runtime.
			for v := range ch {
				out <- v
			}
		}()
	}
	return outs
}

// Broker je jednoduchý publish/subscribe nad kanály.
type Broker struct {
	mu      sync.Mutex
	subs    []chan string
	buffer  int
	dropped int
	closed  bool
}

// NewBroker vytvoří brokera, jehož odběratelé mají buffer dané velikosti.
func NewBroker(buffer int) *Broker {
	if buffer < 0 {
		buffer = 0
	}
	return &Broker{buffer: buffer}
}

// Subscribe zaregistruje odběratele. Po Close vrací rovnou zavřený kanál.
// Kanál odběratele má kapacitu buffer z NewBroker. Souběžně bezpečné s Publish/Close.
func (b *Broker) Subscribe() <-chan string {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan string, b.buffer)
	if b.closed {
		close(ch) // po Close už nikdo nic nepošle, ať odběratel nevisí
		return ch
	}
	b.subs = append(b.subs, ch)
	return ch
}

// --- Stupeň: obtížný ---
// Publish rozešle zprávu všem odběratelům.
func (b *Broker) Publish(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}
	for _, ch := range b.subs {
		// Neblokující zápis: pomalý odběratel nesmí zastavit celého
		// brokera ani ostatní odběratele.
		select {
		case ch <- msg:
		default:
			b.dropped++
		}
	}
}

// Dropped vrací počet zpráv zahozených kvůli pomalým odběratelům.
func (b *Broker) Dropped() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.dropped
}

// Close zavře kanály všech odběratelů; opakované volání nepanikuje.
// Další Subscribe → zavřený kanál; Publish po Close nic nedělá.
func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return // dvojí close by panikoval
	}
	b.closed = true
	for _, ch := range b.subs {
		close(ch) // zavírá odesílatel, tedy broker
	}
	b.subs = nil
}
