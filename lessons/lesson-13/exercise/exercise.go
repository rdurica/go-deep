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
	// TODO: úkol A
	return nil
}

// CountLines spočítá řádky přečtené z r.
func CountLines(r io.Reader) (int, error) {
	// TODO: úkol B
	return 0, nil
}

// NewUpperReader vrací Reader, který za běhu převádí ASCII písmena na velká.
func NewUpperReader(r io.Reader) io.Reader {
	// TODO: úkol B
	return *new(io.Reader)
}

// Tail vrací posledních n řádků ze vstupu.
func Tail(r io.Reader, n int) ([]string, error) {
	// TODO: úkol B
	return nil, nil
}

// Write implementuje io.Writer.
func (cw *CountingWriter) Write(p []byte) (int, error) {
	// TODO: úkol C
	return 0, nil
}

// Bytes vrací počet úspěšně zapsaných bajtů.
func (cw *CountingWriter) Bytes() int64 {
	// TODO: úkol C
	return 0
}

// Lines vrací počet zapsaných znaků nového řádku.
func (cw *CountingWriter) Lines() int {
	// TODO: úkol C
	return 0
}

// Pipeline přečte src po řádcích, na každý řádek zavolá transform
// a výsledek zapíše do dst. Vrací počet zpracovaných řádků.
func Pipeline(src io.Reader, dst io.Writer, transform func(string) string) (int, error) {
	// TODO: úkol C
	return 0, nil
}
