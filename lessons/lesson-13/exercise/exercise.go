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

// WriteReport zapíše číslovaný seznam řádků a souhrn na konci.
func WriteReport(w io.Writer, lines []string) error {
	panic("TODO: úkol A")
}

// CountLines spočítá řádky přečtené z r.
func CountLines(r io.Reader) (int, error) {
	panic("TODO: úkol B")
}

// NewUpperReader vrací Reader, který za běhu převádí ASCII písmena na velká.
func NewUpperReader(r io.Reader) io.Reader {
	panic("TODO: úkol B")
}

// Tail vrací posledních n řádků ze vstupu.
func Tail(r io.Reader, n int) ([]string, error) {
	panic("TODO: úkol B")
}

// Write implementuje io.Writer.
func (cw *CountingWriter) Write(p []byte) (int, error) {
	panic("TODO: úkol C")
}

// Bytes vrací počet úspěšně zapsaných bajtů.
func (cw *CountingWriter) Bytes() int64 {
	panic("TODO: úkol C")
}

// Lines vrací počet zapsaných znaků nového řádku.
func (cw *CountingWriter) Lines() int {
	panic("TODO: úkol C")
}

// Pipeline přečte src po řádcích, na každý řádek zavolá transform
// a výsledek zapíše do dst. Vrací počet zpracovaných řádků.
func Pipeline(src io.Reader, dst io.Writer, transform func(string) string) (int, error) {
	panic("TODO: úkol C")
}
