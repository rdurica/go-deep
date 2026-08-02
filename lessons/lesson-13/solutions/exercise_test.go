package solutions_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-13/solutions"
)

// errWriter je Writer, který po okWrites zápisech začne vracet chybu.
type errWriter struct {
	okWrites int
	seen     int
}

var errDisk = errors.New("disk je plný")

func (w *errWriter) Write(p []byte) (int, error) {
	if w.seen >= w.okWrites {
		return 0, errDisk
	}
	w.seen++
	return len(p), nil
}

func TestWriteReport(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want string
	}{
		{"empty input", nil, "celkem: 0\n"},
		{"one line", []string{"první"}, "1. první\ncelkem: 1\n"},
		{"multiple lines", []string{"a", "b", "c"}, "1. a\n2. b\n3. c\ncelkem: 3\n"},
		{"empty string as line", []string{""}, "1. \ncelkem: 1\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := exercise.WriteReport(&buf, tt.in); err != nil {
				t.Fatalf("WriteReport() = %v, chci nil", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("WriteReport() zapsal %q, chci %q", got, tt.want)
			}
		})
	}
}

func TestWriteReportPropagatesError(t *testing.T) {
	w := &errWriter{okWrites: 1}
	err := exercise.WriteReport(w, []string{"a", "b", "c"})
	if !errors.Is(err, errDisk) {
		t.Errorf("WriteReport() = %v, chci %v", err, errDisk)
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"empty input", "", 0},
		{"one line without newline", "ahoj", 1},
		{"one line with newline", "ahoj\n", 1},
		{"two lines", "a\nb", 2},
		{"two lines with newline", "a\nb\n", 2},
		{"empty lines count", "a\n\nb\n", 3},
		{"CRLF", "a\r\nb\r\n", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exercise.CountLines(strings.NewReader(tt.in))
			if err != nil {
				t.Fatalf("CountLines(%q) = chyba %v, chci nil", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("CountLines(%q) = %d, chci %d", tt.in, got, tt.want)
			}
		})
	}
}

// TestCountLinesLongLine hlídá, že sis pohlídal výchozí 64 KiB limit
// bufio.Scanneru.
func TestCountLinesLongLine(t *testing.T) {
	long := strings.Repeat("x", 200*1024)
	in := "krátký\n" + long + "\nkrátký\n"

	got, err := exercise.CountLines(strings.NewReader(in))
	if err != nil {
		t.Fatalf("CountLines() = chyba %v, chci nil (zvětši buffer scanneru)", err)
	}
	if got != 3 {
		t.Errorf("CountLines() = %d, chci 3", got)
	}
}

func TestCountLinesReturnsReadError(t *testing.T) {
	want := errors.New("čtení selhalo")
	r := io.MultiReader(strings.NewReader("a\nb\n"), &errReader{err: want})

	if _, err := exercise.CountLines(r); !errors.Is(err, want) {
		t.Errorf("CountLines() = %v, chci %v", err, want)
	}
}

// errReader vždycky selže.
type errReader struct{ err error }

func (r *errReader) Read([]byte) (int, error) { return 0, r.err }

func TestUpperReader(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"ahoj", "AHOJ"},
		{"Ahoj Svete 123!", "AHOJ SVETE 123!"},
		{"UZ JE VELKE", "UZ JE VELKE"},
		{"a\nb\n", "A\nB\n"},
	}
	for _, tt := range tests {
		got, err := io.ReadAll(exercise.NewUpperReader(strings.NewReader(tt.in)))
		if err != nil {
			t.Fatalf("čtení z UpperReaderu selhalo: %v", err)
		}
		if string(got) != tt.want {
			t.Errorf("UpperReader(%q) = %q, chci %q", tt.in, string(got), tt.want)
		}
	}
}

// TestUpperReaderLeavesNonASCII ověřuje, že převod je bajtový a nerozbije UTF-8.
func TestUpperReaderLeavesNonASCII(t *testing.T) {
	got, err := io.ReadAll(exercise.NewUpperReader(strings.NewReader("Světe")))
	if err != nil {
		t.Fatalf("čtení selhalo: %v", err)
	}
	if string(got) != "SVěTE" {
		t.Errorf("UpperReader(%q) = %q, chci %q", "Světe", string(got), "SVěTE")
	}
}

// TestUpperReaderSmallChunks ověřuje, že převod probíhá průběžně a funguje
// i při čtení po třech bajtech.
func TestUpperReaderSmallChunks(t *testing.T) {
	r := exercise.NewUpperReader(strings.NewReader("abcdefghij"))

	var out []byte
	buf := make([]byte, 3)
	for {
		n, err := r.Read(buf)
		out = append(out, buf[:n]...)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Read() = %v", err)
		}
	}
	if string(out) != "ABCDEFGHIJ" {
		t.Errorf("po kusech přečteno %q, chci %q", string(out), "ABCDEFGHIJ")
	}
}

// TestUpperReaderWorksWithDecorators ověřuje, že UpperReader nečte dopředu:
// io.LimitReader nad ním musí vrátit přesně tolik bajtů, kolik povolí.
func TestUpperReaderWorksWithDecorators(t *testing.T) {
	r := io.LimitReader(exercise.NewUpperReader(strings.NewReader("abcdef")), 3)
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("čtení selhalo: %v", err)
	}
	if string(got) != "ABC" {
		t.Errorf("LimitReader(UpperReader(...), 3) = %q, chci %q", string(got), "ABC")
	}
}

func TestTail(t *testing.T) {
	tests := []struct {
		name string
		in   string
		n    int
		want []string
	}{
		{"last two", "a\nb\nc\nd\n", 2, []string{"c", "d"}},
		{"more than lines", "a\nb\n", 5, []string{"a", "b"}},
		{"exactly all", "a\nb\n", 2, []string{"a", "b"}},
		{"jeden", "a\nb\nc", 1, []string{"c"}},
		{"empty input", "", 3, nil},
		{"n je nula", "a\nb\n", 0, nil},
		{"n is negative", "a\nb\n", -2, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exercise.Tail(strings.NewReader(tt.in), tt.n)
			if err != nil {
				t.Fatalf("Tail() = chyba %v, chci nil", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("Tail(%q, %d) = %v, chci %v", tt.in, tt.n, got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("Tail(%q, %d)[%d] = %q, chci %q", tt.in, tt.n, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCountingWriterZeroValue(t *testing.T) {
	var cw exercise.CountingWriter

	n, err := cw.Write([]byte("ahoj\nsvete\n"))
	if err != nil {
		t.Fatalf("Write() = %v, chci nil", err)
	}
	if n != len("ahoj\nsvete\n") {
		t.Errorf("Write() vrátil n=%d, chci %d", n, len("ahoj\nsvete\n"))
	}
	if got := cw.Bytes(); got != 11 {
		t.Errorf("Bytes() = %d, chci 11", got)
	}
	if got := cw.Lines(); got != 2 {
		t.Errorf("Lines() = %d, chci 2", got)
	}
}

func TestCountingWriterAsDecorator(t *testing.T) {
	var sink bytes.Buffer
	cw := &exercise.CountingWriter{W: &sink}

	if _, err := io.Copy(cw, strings.NewReader("a\nb\nc")); err != nil {
		t.Fatalf("io.Copy() = %v, chci nil", err)
	}
	if _, err := fmt.Fprintf(cw, "\nd=%d\n", 4); err != nil {
		t.Fatalf("Fprintf() = %v, chci nil", err)
	}

	if got, want := sink.String(), "a\nb\nc\nd=4\n"; got != want {
		t.Errorf("podkladový writer obsahuje %q, chci %q", got, want)
	}
	if got, want := cw.Bytes(), int64(len("a\nb\nc\nd=4\n")); got != want {
		t.Errorf("Bytes() = %d, chci %d", got, want)
	}
	if got := cw.Lines(); got != 4 {
		t.Errorf("Lines() = %d, chci 4", got)
	}
}

func TestCountingWriterPropagatesError(t *testing.T) {
	cw := &exercise.CountingWriter{W: &errWriter{okWrites: 0}}

	if _, err := cw.Write([]byte("data\n")); !errors.Is(err, errDisk) {
		t.Errorf("Write() = %v, chci %v", err, errDisk)
	}
	if got := cw.Bytes(); got != 0 {
		t.Errorf("Bytes() = %d, chci 0 — nezapsané bajty se nepočítají", got)
	}
}

func TestPipeline(t *testing.T) {
	tests := []struct {
		name      string
		in        string
		transform func(string) string
		want      string
		wantLines int
	}{
		{"uppercase", "a\nb\n", strings.ToUpper, "A\nB\n", 2},
		{"unchanged", "x\ny\nz", func(s string) string { return s }, "x\ny\nz\n", 3},
		{"empty input", "", strings.ToUpper, "", 0},
		{
			"prefix",
			"jedna\ndva\n",
			func(s string) string { return "> " + s },
			"> jedna\n> dva\n",
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			n, err := exercise.Pipeline(strings.NewReader(tt.in), &out, tt.transform)
			if err != nil {
				t.Fatalf("Pipeline() = chyba %v, chci nil", err)
			}
			if n != tt.wantLines {
				t.Errorf("Pipeline() = %d řádků, chci %d", n, tt.wantLines)
			}
			if got := out.String(); got != tt.want {
				t.Errorf("Pipeline() zapsal %q, chci %q", got, tt.want)
			}
		})
	}
}

// TestPipelineWithCustomTypes propojí všechno dohromady: zdrojem je vlastní
// UpperReader, cílem vlastní CountingWriter.
func TestPipelineWithCustomTypes(t *testing.T) {
	src := exercise.NewUpperReader(strings.NewReader("ahoj\nsvete\n"))
	var sink bytes.Buffer
	cw := &exercise.CountingWriter{W: &sink}

	n, err := exercise.Pipeline(src, cw, func(s string) string { return s + "!" })
	if err != nil {
		t.Fatalf("Pipeline() = %v, chci nil", err)
	}
	if n != 2 {
		t.Errorf("Pipeline() = %d, chci 2", n)
	}
	if got, want := sink.String(), "AHOJ!\nSVETE!\n"; got != want {
		t.Errorf("výstup = %q, chci %q", got, want)
	}
	if got := cw.Lines(); got != 2 {
		t.Errorf("CountingWriter.Lines() = %d, chci 2", got)
	}
}

func TestPipelinePropagatesError(t *testing.T) {
	_, err := exercise.Pipeline(
		strings.NewReader("a\nb\nc\n"),
		&errWriter{okWrites: 1},
		strings.ToUpper,
	)
	if !errors.Is(err, errDisk) {
		t.Errorf("Pipeline() = %v, chci %v", err, errDisk)
	}
}
