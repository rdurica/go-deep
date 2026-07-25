package solutions_test

import (
	"bytes"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-53/solutions"
)

// Sinky pro benchmarky a měření alokací.
var (
	sinkInt    int
	sinkString string
	sinkMap    map[string]int
)

// vzorekTextu je vstup pro benchmarky a měření alokací.
const vzorekTextu = `Go je jazyk, ve kterém se profiluje AŽ TEHDY, když máš benchmark.
Benchmark bez sinku měří prázdný cyklus. Alokace v cyklu jsou 90 % nálezů;
konverze string na []byte a zpět je druhá polovina. Regex v horké cestě je klasika.`

func TestSumDigits(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"abc", 0},
		{"123", 6},
		{"a1b2c3", 6},
		{"0000", 0},
		{"9876543210", 45},
		{"cena 1999 Kč", 28},
		{"žluťoučký 7", 7},
	}
	for _, tt := range tests {
		if got := exercise.SumDigits(tt.in); got != tt.want {
			t.Errorf("SumDigits(%q) = %d, chci %d", tt.in, got, tt.want)
		}
	}
}

// TestSumDigitsShodaSPomalou porovnává rychlou verzi s referenční na náhodných datech,
// aby test nešlo splnit zadrátovanou tabulkou.
func TestSumDigitsShodaSPomalou(t *testing.T) {
	rnd := rand.New(rand.NewSource(3))
	abeceda := []rune("0123456789abcxyzžčř .,-")
	for i := 0; i < 300; i++ {
		b := make([]rune, rnd.Intn(40))
		for j := range b {
			b[j] = abeceda[rnd.Intn(len(abeceda))]
		}
		in := string(b)
		want := exercise.SumDigitsSlow(in)
		if got := exercise.SumDigits(in); got != want {
			t.Fatalf("SumDigits(%q) = %d, SumDigitsSlow = %d", in, got, want)
		}
	}
}

func TestSumDigitsNealokuje(t *testing.T) {
	in := "objednávka 2024 položek 1999 kusů 42"
	if n := testing.AllocsPerRun(200, func() { sinkInt = exercise.SumDigits(in) }); n != 0 {
		t.Errorf("SumDigits alokuje %.1f×, chci 0 — pracuj s bajty, ne s konverzemi", n)
	}
}

func TestCountWords(t *testing.T) {
	got := exercise.CountWords("Pes a kočka. PES! pes? kocka2 kocka2")
	want := map[string]int{"pes": 3, "a": 1, "kočka": 1, "kocka2": 2}
	if len(got) != len(want) {
		t.Fatalf("CountWords vrátil %d klíčů (%v), chci %d", len(got), got, len(want))
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("CountWords[%q] = %d, chci %d", k, got[k], v)
		}
	}
}

func TestCountWordsPrazdny(t *testing.T) {
	got := exercise.CountWords("")
	if got == nil {
		t.Fatal("CountWords(\"\") = nil mapa, chci prázdnou nenilovou")
	}
	if len(got) != 0 {
		t.Errorf("CountWords(\"\") = %v, chci prázdnou mapu", got)
	}
}

func TestCountWordsShodaSPomalou(t *testing.T) {
	rnd := rand.New(rand.NewSource(11))
	slova := []string{"Pes", "pes", "KOČKA", "kočka2", "Žába", "x", "42", "ř"}
	oddelovace := []string{" ", ", ", ".\n", "!  ", " -- ", "\t"}
	for i := 0; i < 100; i++ {
		var sb strings.Builder
		for j := 0; j < rnd.Intn(30); j++ {
			sb.WriteString(slova[rnd.Intn(len(slova))])
			sb.WriteString(oddelovace[rnd.Intn(len(oddelovace))])
		}
		in := sb.String()
		want := exercise.CountWordsSlow(in)
		got := exercise.CountWords(in)
		if len(got) != len(want) {
			t.Fatalf("CountWords(%q) má %d klíčů, CountWordsSlow %d", in, len(got), len(want))
		}
		for k, v := range want {
			if got[k] != v {
				t.Fatalf("CountWords(%q)[%q] = %d, chci %d", in, k, got[k], v)
			}
		}
	}
}

func TestCountWordsAlokace(t *testing.T) {
	slow := testing.AllocsPerRun(50, func() { sinkMap = exercise.CountWordsSlow(vzorekTextu) })
	fast := testing.AllocsPerRun(50, func() { sinkMap = exercise.CountWords(vzorekTextu) })
	if fast >= slow {
		t.Errorf("CountWords alokuje %.1f×, CountWordsSlow %.1f× — rychlá verze má alokovat míň", fast, slow)
	}
	if fast > 12 {
		t.Errorf("CountWords alokuje %.1f×, chci nejvýš 12 (předalokovaná mapa, žádný regex)", fast)
	}
}

func TestJoinIDs(t *testing.T) {
	tests := []struct {
		in   []int
		want string
	}{
		{nil, ""},
		{[]int{}, ""},
		{[]int{1}, "1"},
		{[]int{1, 2, 3}, "1,2,3"},
		{[]int{-1, 0, 1}, "-1,0,1"},
		{[]int{1000000, 999999999}, "1000000,999999999"},
	}
	for _, tt := range tests {
		if got := exercise.JoinIDs(tt.in); got != tt.want {
			t.Errorf("JoinIDs(%v) = %q, chci %q", tt.in, got, tt.want)
		}
	}
}

func TestJoinIDsAlokace(t *testing.T) {
	ids := make([]int, 64)
	for i := range ids {
		ids[i] = i * 137
	}
	if n := testing.AllocsPerRun(200, func() { sinkString = exercise.JoinIDs(ids) }); n > 2 {
		t.Errorf("JoinIDs alokuje %.1f×, chci nejvýš 2 (Builder s Grow a AppendInt)", n)
	}
}

func TestCaptureCPUProfile(t *testing.T) {
	var buf bytes.Buffer
	spusteno := false
	err := exercise.CaptureCPUProfile(&buf, func() {
		spusteno = true
		sinkInt = 0
		for i := 0; i < 3_000_000; i++ {
			sinkInt += i % 7
		}
	})
	if err != nil {
		t.Fatalf("CaptureCPUProfile = chyba %v", err)
	}
	if !spusteno {
		t.Error("CaptureCPUProfile nespustil předanou funkci")
	}
	overGzip(t, "CPU", buf.Bytes())
}

func TestCaptureCPUProfileChybnyVstup(t *testing.T) {
	var buf bytes.Buffer
	if err := exercise.CaptureCPUProfile(&buf, nil); err == nil {
		t.Error("CaptureCPUProfile(w, nil) = nil, chci chybu")
	}
	if err := exercise.CaptureCPUProfile(nil, func() {}); err == nil {
		t.Error("CaptureCPUProfile(nil, f) = nil, chci chybu")
	}
}

func TestCaptureHeapProfile(t *testing.T) {
	// Něco naalokujeme, ať profil není prázdný.
	drzeno := make([][]byte, 0, 64)
	for i := 0; i < 64; i++ {
		drzeno = append(drzeno, make([]byte, 4096))
	}
	var buf bytes.Buffer
	if err := exercise.CaptureHeapProfile(&buf); err != nil {
		t.Fatalf("CaptureHeapProfile = chyba %v", err)
	}
	overGzip(t, "heap", buf.Bytes())
	if len(drzeno) != 64 {
		t.Error("alokovaná data se nesmí ztratit před snímkem")
	}
}

func TestCaptureHeapProfileChybnyVstup(t *testing.T) {
	if err := exercise.CaptureHeapProfile(nil); err == nil {
		t.Error("CaptureHeapProfile(nil) = nil, chci chybu")
	}
}

func TestPprofHandler(t *testing.T) {
	h := exercise.PprofHandler()
	if h == nil {
		t.Fatal("PprofHandler = nil")
	}
	cesty := []string{
		"/debug/pprof/",
		"/debug/pprof/heap",
		"/debug/pprof/goroutine?debug=1",
		"/debug/pprof/cmdline",
	}
	for _, cesta := range cesty {
		req := httptest.NewRequest(http.MethodGet, cesta, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, chci 200", cesta, rec.Code)
		}
		if rec.Body.Len() == 0 {
			t.Errorf("GET %s vrátil prázdné tělo", cesta)
		}
	}
}

func TestPprofHandlerNeznamaCesta(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	exercise.PprofHandler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /metrics = %d, chci 404 (mux má nést jen pprof)", rec.Code)
	}
}

// overGzip ověří, že zapsaná data vypadají jako pprof profil (gzipovaný protobuf).
func overGzip(t *testing.T, jmeno string, data []byte) {
	t.Helper()
	if len(data) == 0 {
		t.Fatalf("%s profil je prázdný", jmeno)
	}
	if len(data) < 2 || data[0] != 0x1f || data[1] != 0x8b {
		t.Fatalf("%s profil nezačíná gzip hlavičkou, začátek = % x", jmeno, data[:min(8, len(data))])
	}
}

func BenchmarkSumDigitsSlow(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkInt = exercise.SumDigitsSlow(vzorekTextu)
	}
}

func BenchmarkSumDigits(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkInt = exercise.SumDigits(vzorekTextu)
	}
}

func BenchmarkCountWordsSlow(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkMap = exercise.CountWordsSlow(vzorekTextu)
	}
}

func BenchmarkCountWords(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkMap = exercise.CountWords(vzorekTextu)
	}
}

func BenchmarkJoinIDs(b *testing.B) {
	ids := make([]int, 64)
	for i := range ids {
		ids[i] = i * 137
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkString = exercise.JoinIDs(ids)
	}
}
