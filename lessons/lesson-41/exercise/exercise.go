// Package exercise obsahuje cvičení lekce 41.
package exercise

// --- Stupeň: jednoduchý ---
// Generate pošle všechna čísla do kanálu v goroutině a kanál sám zavře.
// Bez Collect nesmí blokovat volajícího.
func Generate(nums ...int) <-chan int {
	// TODO
	return closedInt()
}

// Collect přečte kanál do zavření a vrátí hodnoty v pořadí doručení.
// Prázdný nebo nil kanál → prázdný slice, ne panika.
func Collect(ch <-chan int) []int {
	// TODO
	return nil
}

// Merge sloučí vstupní kanály (fan-in). Výstup zavře jednou po zavření všech vstupů.
// Bez argumentů → rovnou zavřený kanál.
func Merge(chs ...<-chan int) <-chan int {
	// TODO
	return closedInt()
}

// --- Stupeň: střední ---
// Split rozdělí vstup mezi n kanálů; každá hodnota jen v jednom. Po vstupu
// zavře všechny výstupy. n < 1 se chová jako 1.
func Split(ch <-chan int, n int) []<-chan int {
	// TODO
	return nil
}

// Broker je publish/subscribe nad kanály s drop policy pro pomalé odběratele.
type Broker struct {
	// TODO
}

// NewBroker vytvoří brokera; buffer je velikost kanálu každého odběratele.
// Záporný buffer se chová jako nula.
func NewBroker(buffer int) *Broker {
	// TODO
	return nil
}

// Subscribe zaregistruje odběratele. Po Close vrací rovnou zavřený kanál.
// Kanál odběratele má kapacitu buffer z NewBroker. Souběžně bezpečné s Publish/Close.
func (b *Broker) Subscribe() <-chan string {
	// TODO
	ch := make(chan string)
	close(ch)
	return ch
}

// --- Stupeň: obtížný ---
// Publish rozešle zprávu všem. Pomalý odběratel nesmí blokovat — zahoď a započítej Dropped.
// Po Close nic nedělá.
func (b *Broker) Publish(msg string) {
	// TODO
}

// Dropped vrací počet zahozených zpráv.
// Roste, když Publish najde plný buffer odběratele.
func (b *Broker) Dropped() int {
	// TODO
	return 0
}

// Close zavře kanály všech odběratelů; opakované volání nepanikuje.
// Další Subscribe → zavřený kanál; Publish po Close nic nedělá.
func (b *Broker) Close() {
	// TODO
}

// closedInt je fail-fast stub: nil kanál by v testech visel navždy.
func closedInt() <-chan int {
	ch := make(chan int)
	close(ch)
	return ch
}
