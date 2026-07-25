// Package solutions obsahuje referenční řešení lekce 06.
package solutions

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
	if a == nil || b == nil {
		return
	}
	*a, *b = *b, *a
}

// ApplyDefaults doplní výchozí hodnoty tam, kde je zero value nebo nil.
// Nil pointer na Config je platný vstup a znamená "nedělej nic".
func ApplyDefaults(c *Config) {
	if c == nil {
		return
	}
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 8080
	}
	if c.Debug == nil {
		debug := false
		c.Debug = &debug
	}
}

// IncrementAll zvýší každý prvek slice o jedna přímo na místě.
func IncrementAll(nums []int) {
	for i := range nums {
		nums[i]++
	}
}

// AppendSafe vrátí nový slice s přidaným prvkem, aniž by sáhl
// na podkladové pole volajícího.
func AppendSafe(nums []int, v int) []int {
	out := make([]int, len(nums), len(nums)+1)
	copy(out, nums)
	return append(out, v)
}

// Push vloží hodnotu na začátek seznamu a vrátí novou hlavu.
func Push(head *Node, v int) *Node {
	return &Node{Val: v, Next: head}
}

// Len vrací počet prvků seznamu. Nil hlava znamená 0.
func Len(head *Node) int {
	n := 0
	for node := head; node != nil; node = node.Next {
		n++
	}
	return n
}

// Reverse otočí pořadí prvků seznamu a vrátí novou hlavu.
func Reverse(head *Node) *Node {
	var prev *Node
	for node := head; node != nil; {
		next := node.Next
		node.Next = prev
		prev = node
		node = next
	}
	return prev
}
