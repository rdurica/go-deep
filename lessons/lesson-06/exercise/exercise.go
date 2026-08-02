// Package exercise obsahuje cvičení lekce 06.
package exercise

// Config je konfigurace serveru. Debug je pointer, protože potřebujeme
// rozlišit "nenastaveno" od "nastaveno na false".
type Config struct {
	Host  string
	Port  int
	Debug *bool
}

// Node je nod jednosměrně zřetězeného seznamu.
type Node struct {
	Val  int
	Next *Node
}

// --- Stupeň: jednoduchý ---
// Swap prohodí hodnoty, na které ukazují a a b.
// Nil pointer nepanikuje (no-op). Swap(&x, &x) nechá x beze změny.
func Swap(a, b *int) {
	// TODO
}

// ApplyDefaults doplní výchozí hodnoty: prázdný Host → "localhost", Port 0 → 8080,
// Debug nil → pointer na false. Už nastavené hodnoty (včetně &false) se nepřepisují.
// Nil Config je no-op bez paniky.
func ApplyDefaults(c *Config) {
	// TODO
}

// --- Stupeň: střední ---
// IncrementAll zvýší každý prvek o jedna na místě. Nil i prázdný slice projdou bez paniky.
func IncrementAll(nums []int) {
	// TODO
}

// AppendSafe vrátí nový slice s přidaným prvkem bez zásahu do podkladového pole volajícího.
// Musíš data zkopírovat (ne jen append do sdíleného backing pole). Nil vstup → slice s jedním prvkem.
func AppendSafe(nums []int, v int) []int {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---
// Push vloží v na začátek seznamu a vrátí nový první nod. Push(nil, 42) vytvoří jednoprvkový seznam.
func Push(head *Node, v int) *Node {
	// TODO
	return nil
}

// Len vrací počet prvků. nil (prázdný seznam) dá 0. Cyklem, ne rekurzí.
func Len(head *Node) int {
	// TODO
	return 0
}

// Reverse otočí pořadí prvků jedním průchodem (prev/next) a vrátí nový první nod.
// Reverse(nil) vrací nil.
func Reverse(head *Node) *Node {
	// TODO
	return nil
}
