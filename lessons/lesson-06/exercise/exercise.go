// Package exercise obsahuje cvičení lekce 06.
package exercise

// Config je konfigurace serveru. Debug je pointer, protože potřebujeme
// rozlišit "nenastaveno" od "nastaveno na false".
type Config struct {
	Host  string
	Port  int
	Debug *bool
}

// Node je prvek jednosměrně zřetězeného seznamu.
type Node struct {
	Val  int
	Next *Node
}

// Swap prohodí hodnoty, na které ukazují a a b.
// Pokud je kterýkoli pointer nil, neudělá nic.
func Swap(a, b *int) {
	// TODO: úkol A
}

// ApplyDefaults doplní výchozí hodnoty tam, kde je zero value nebo nil.
// Nil pointer na Config je platný vstup a znamená "nedělej nic".
func ApplyDefaults(c *Config) {
	// TODO: úkol B
}

// IncrementAll zvýší každý prvek slice o jedna přímo na místě.
func IncrementAll(nums []int) {
	// TODO: úkol B
}

// AppendSafe vrátí nový slice s přidaným prvkem, aniž by sáhl
// na podkladové pole volajícího.
func AppendSafe(nums []int, v int) []int {
	// TODO: úkol B
	return nil
}

// Push vloží hodnotu na začátek seznamu a vrátí novou hlavu.
func Push(head *Node, v int) *Node {
	// TODO: úkol C
	return nil
}

// Len vrací počet prvků seznamu. Nil hlava znamená 0.
func Len(head *Node) int {
	// TODO: úkol C
	return 0
}

// Reverse otočí pořadí prvků seznamu a vrátí novou hlavu.
func Reverse(head *Node) *Node {
	// TODO: úkol C
	return nil
}
