// Package exercise obsahuje cvičení lekce 41.
package exercise

// Generate pošle všechna čísla do vráceného kanálu a kanál sám zavře.
func Generate(nums ...int) <-chan int {
	// TODO: úkol A
	return nil
}

// Collect přečte kanál až do zavření a vrátí hodnoty v pořadí, v jakém dorazily.
func Collect(ch <-chan int) []int {
	// TODO: úkol A
	return nil
}

// Merge sloučí několik vstupních kanálů do jednoho (fan-in). Výstup se zavře
// až po zavření všech vstupů.
func Merge(chs ...<-chan int) <-chan int {
	// TODO: úkol B
	return nil
}

// Split rozdělí hodnoty ze vstupního kanálu mezi n výstupních kanálů. Každá
// hodnota skončí právě v jednom z nich. Po zavření vstupu se zavřou i výstupy.
func Split(ch <-chan int, n int) []<-chan int {
	// TODO: úkol B
	return nil
}

// Broker je jednoduchý publish/subscribe nad kanály.
type Broker struct {
	// TODO: doplň pole. Budeš potřebovat mutex, seznam odběratelů,
	// velikost bufferu, počítadlo zahozených zpráv a příznak zavření.
}

// NewBroker vytvoří brokera, jehož odběratelé mají buffer dané velikosti.
// Záporný buffer se chová jako nula.
func NewBroker(buffer int) *Broker {
	// TODO: úkol C
	return nil
}

// Subscribe zaregistruje nového odběratele a vrátí jeho kanál. Po zavolání
// Close vrací už zavřený kanál.
func (b *Broker) Subscribe() <-chan string {
	// TODO: úkol C
	return nil
}

// Publish rozešle zprávu všem odběratelům. Odběratele, který nestíhá, nesmí
// zablokovat — zprávu pro něj zahodí a započítá do Dropped.
func (b *Broker) Publish(msg string) {
	// TODO: úkol C
}

// Dropped vrací počet zpráv zahozených kvůli pomalým odběratelům.
func (b *Broker) Dropped() int {
	// TODO: úkol C
	return 0
}

// Close ukončí brokera a zavře kanály všech odběratelů. Opakované volání
// je bezpečné, Publish po zavření nic nedělá.
func (b *Broker) Close() {
	// TODO: úkol C
}
