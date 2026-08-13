// Package solutions obsahuje referenční řešení lekce 13.
package solutions

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// maxLine je horní mez délky jednoho řádku pro bufio.Scanner.
const maxLine = 1 << 20

// CountingWriter je io.Writer, který počítá zapsané bajty a řádky.
type CountingWriter struct {
	W io.Writer

	bytes int64
	lines int
}

// --- Stupeň: jednoduchý ---

// CountLines spočítá řádky přečtené z r.
func CountLines(r io.Reader) (int, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxLine)

	count := 0
	for sc.Scan() {
		count++
	}
	if err := sc.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

// upperReader je dekorátor nad jiným Readerem.
type upperReader struct {
	r io.Reader
}

func (u upperReader) Read(p []byte) (int, error) {
	n, err := u.r.Read(p)
	for i := 0; i < n; i++ {
		if c := p[i]; c >= 'a' && c <= 'z' {
			p[i] = c - ('a' - 'A')
		}
	}
	return n, err
}

// --- Stupeň: střední ---

// WriteReport zapíše číslovaný seznam řádků a souhrn na konci.
func WriteReport(w io.Writer, lines []string) error {
	for i, line := range lines {
		if _, err := fmt.Fprintf(w, "%d. %s\n", i+1, line); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "celkem: %d\n", len(lines)); err != nil {
		return err
	}
	return nil
}

// NewUpperReader vrací Reader, který za běhu převádí ASCII písmena na velká.
func NewUpperReader(r io.Reader) io.Reader {
	return upperReader{r: r}
}

// --- Stupeň: obtížný ---

// Write implementuje io.Writer.
func (cw *CountingWriter) Write(p []byte) (int, error) {
	n := len(p)
	var err error
	if cw.W != nil {
		n, err = cw.W.Write(p)
	}
	cw.bytes += int64(n)
	cw.lines += bytes.Count(p[:n], []byte{'\n'})
	return n, err
}

// Bytes vrací počet úspěšně zapsaných bajtů.
func (cw *CountingWriter) Bytes() int64 {
	return cw.bytes
}

// Lines vrací počet zapsaných znaků nového řádku.
func (cw *CountingWriter) Lines() int {
	return cw.lines
}
