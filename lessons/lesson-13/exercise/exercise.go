// Package exercise obsahuje cvičení lekce 13.
package exercise

import "io"

// CountingWriter je io.Writer, který počítá zapsané bajty a řádky.
// Je-li W nenulový, data se navíc přeposílají do něj.
// Zero value je použitelná a chová se jako io.Discard s počítadlem.
type CountingWriter struct {
	W io.Writer

	bytes int64
	lines int
}

// --- Stupeň: jednoduchý ---
// WriteReport zapíše číslovaný seznam řádků ve tvaru "%d. %s\n" (od 1)
// a na konci vždy "celkem: %d\n" (i pro prázdný vstup). Chyby zapisu propaguj.
func WriteReport(w io.Writer, lines []string) error {
	// TODO
	return nil
}

// CountLines spočítá řádky v r. Prázdný vstup → 0. Prázdné řádky se počítají.
// Použij bufio.Scanner, nezapomeň na sc.Err() a zvyš limit na délku řádku
// (test posílá ~200 KiB).
func CountLines(r io.Reader) (int, error) {
	// TODO
	return 0, nil
}

// --- Stupeň: střední ---
// NewUpperReader vrací Reader, který za běhu převádí ASCII a–z na velká písmena.
// Nesmí načíst celý vstup dopředu. Ne-ASCII bajty nech beze změny ("Světe" → "SVěTE").
func NewUpperReader(r io.Reader) io.Reader {
	// TODO
	return *new(io.Reader)
}

// Tail vrací posledních n řádků v původním pořadí. Méně řádků → všechny.
// n <= 0 → prázdný výsledek. Drž v paměti max n řádků.
func Tail(r io.Reader, n int) ([]string, error) {
	// TODO
	return nil, nil
}

// Write implementuje io.Writer. Počítá skutečně zapsané bajty (n z podkladového zápisu)
// a znaky '\n' v zapsaných datech. W nenulové → přeposílá a vrací jeho n, err.
// W nil → chová se jako io.Discard, jen počítá.
func (cw *CountingWriter) Write(p []byte) (int, error) {
	// TODO
	return 0, nil
}

// Bytes vrací počet úspěšně zapsaných bajtů.
func (cw *CountingWriter) Bytes() int64 {
	// TODO
	return 0
}

// Lines vrací počet zapsaných znaků nového řádku.
func (cw *CountingWriter) Lines() int {
	// TODO
	return 0
}

// --- Stupeň: obtížný ---
// Pipeline čte src po řádcích, na každý zavolá transform a výsledek zapíše do dst s '\n'.
// Vrací počet zpracovaných řádků. Při chybě zápisu vrátí chybu a dosavadní počet.
// transform == nil znamená identitu.
func Pipeline(src io.Reader, dst io.Writer, transform func(string) string) (int, error) {
	// TODO
	return 0, nil
}
