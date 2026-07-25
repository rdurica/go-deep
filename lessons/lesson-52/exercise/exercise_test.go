package exercise_test

import (
	"errors"
	"flag"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-52/exercise"
)

var update = flag.Bool("update", false, "přepiš golden soubory v testdata/")

// Cíle pro benchmarky, aby kompilátor nevyhodil mrtvý kód.
var (
	sinkString  string
	sinkRecords []exercise.Record
)

var goldenRecords = []exercise.Record{
	{ID: "a1", Name: "Alice", Score: 100},
	{ID: "b2", Name: "Žofie Nováková", Score: 7},
	{ID: "c3", Name: "Bob", Score: -42},
	{ID: "dlouhe-id-42", Name: "Ch", Score: 12345},
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"   ", ""},
		{"\n\t ", ""},
		{"ahoj", "ahoj"},
		{"AHOJ", "ahoj"},
		{"  Ahoj   Svete \n", "ahoj svete"},
		{"a\t\tb\nc", "a b c"},
		{"ŽLUŤOUČKÝ  KŮŇ", "žluťoučký kůň"},
		{"already normalized", "already normalized"},
		{"  jedno  ", "jedno"},
	}
	for _, tt := range tests {
		if got := exercise.Normalize(tt.in); got != tt.want {
			t.Errorf("Normalize(%q) = %q, chci %q", tt.in, got, tt.want)
		}
	}
}

func TestNormalizeIdempotence(t *testing.T) {
	inputs := []string{"", "  A  b\tC  ", "žluťoučký   kůň\n", "x"}
	for _, in := range inputs {
		once := exercise.Normalize(in)
		twice := exercise.Normalize(once)
		if once != twice {
			t.Errorf("Normalize není idempotentní: %q -> %q -> %q", in, once, twice)
		}
	}
}

func TestNormalizeAllocs(t *testing.T) {
	in := "  " + strings.Repeat("Slovo   Dalsi\t", 40) + "  "
	const limit = 3

	got := testing.AllocsPerRun(100, func() {
		sinkString = exercise.Normalize(in)
	})
	if got > limit {
		t.Errorf("Normalize alokuje %.0f× na volání, chci nejvýš %d — stav výsledek v jednom bufferu", got, limit)
	}
}

func BenchmarkNormalize(b *testing.B) {
	in := "  " + strings.Repeat("Slovo   Dalsi\t", 40) + "  "
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkString = exercise.Normalize(in)
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		name string
		in   []exercise.Record
		want string
	}{
		{"nil", nil, ""},
		{"prázdný slice", []exercise.Record{}, ""},
		{"jeden", []exercise.Record{{ID: "a", Name: "b", Score: 1}}, "a|b|1"},
		{"zero value", []exercise.Record{{}}, "||0"},
		{
			"escapování",
			[]exercise.Record{{ID: "x|y", Name: `a\b`, Score: 0}, {ID: "n\nl", Name: "r\rl", Score: -1}},
			`x\|y|a\\b|0` + "\n" + `n\nl|r\rl|-1`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.Encode(tt.in); got != tt.want {
				t.Errorf("Encode(%+v) = %q, chci %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		in   string
		want []exercise.Record
	}{
		{"", nil},
		{"a|b|1", []exercise.Record{{ID: "a", Name: "b", Score: 1}}},
		{"||0", []exercise.Record{{}}},
		{`x\|y|a\\b|0`, []exercise.Record{{ID: "x|y", Name: `a\b`, Score: 0}}},
		{"a|b|1\nc|d|-2", []exercise.Record{{ID: "a", Name: "b", Score: 1}, {ID: "c", Name: "d", Score: -2}}},
	}
	for _, tt := range tests {
		got, err := exercise.Decode(tt.in)
		if err != nil {
			t.Errorf("Decode(%q) vrátila chybu %v", tt.in, err)
			continue
		}
		if len(got) != len(tt.want) {
			t.Errorf("Decode(%q) = %+v, chci %+v", tt.in, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("Decode(%q)[%d] = %+v, chci %+v", tt.in, i, got[i], tt.want[i])
			}
		}
	}
}

func TestDecodeChyby(t *testing.T) {
	bad := []string{
		"abc",       // jedno pole
		"a|b",       // dvě pole
		"a|b|c|d",   // čtyři pole
		"a|b|x",     // skóre není číslo
		"a|b|007",   // nekanonické číslo
		"a|b|+7",    // nekanonické číslo
		"a|b|1\n",   // prázdný poslední řádek
		"a|b|1\n\n", // prázdný řádek uprostřed
		`a\`,        // useknutá escape sekvence
		`a\q|b|1`,   // neznámá escape sekvence
		"a|b|1\rx",  // neescapovaný CR
	}
	for _, in := range bad {
		if _, err := exercise.Decode(in); err == nil {
			t.Errorf("Decode(%q) = nil chyba, chci chybu", in)
		} else if !errors.Is(err, exercise.ErrFormat) {
			t.Errorf("Decode(%q) vrátila %v, chci errors.Is(err, ErrFormat)", in, err)
		}
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	recs := []exercise.Record{
		{ID: "a1", Name: "Alice", Score: 100},
		{ID: "b|2", Name: "Bob\nnovy radek", Score: 0},
		{ID: `c\3`, Name: "Žofie", Score: -9},
		{ID: "", Name: "", Score: 1 << 40},
	}
	got, err := exercise.Decode(exercise.Encode(recs))
	if err != nil {
		t.Fatalf("Decode(Encode(recs)) vrátila chybu %v", err)
	}
	if len(got) != len(recs) {
		t.Fatalf("round-trip dal %d záznamů, chci %d", len(got), len(recs))
	}
	for i := range recs {
		if got[i] != recs[i] {
			t.Errorf("round-trip[%d] = %+v, chci %+v", i, got[i], recs[i])
		}
	}
}

func FuzzRoundTrip(f *testing.F) {
	f.Add("a1", "Alice", 100)
	f.Add("", "", 0)
	f.Add("x|y", `a\b`, -5)
	f.Add("radek\n", "navrat\r", 1<<40)

	f.Fuzz(func(t *testing.T, id, name string, score int) {
		recs := []exercise.Record{
			{ID: id, Name: name, Score: score},
			{ID: name, Name: id, Score: -score},
		}
		encoded := exercise.Encode(recs)
		got, err := exercise.Decode(encoded)
		if err != nil {
			t.Fatalf("Decode(Encode(%+v)) = chyba %v (kódováno jako %q)", recs, err, encoded)
		}
		if len(got) != len(recs) {
			t.Fatalf("round-trip dal %d záznamů, chci %d (kódováno jako %q)", len(got), len(recs), encoded)
		}
		for i := range recs {
			if got[i] != recs[i] {
				t.Fatalf("round-trip[%d] = %+v, chci %+v (kódováno jako %q)", i, got[i], recs[i], encoded)
			}
		}
	})
}

func FuzzDecodeCanonical(f *testing.F) {
	f.Add("")
	f.Add("a|b|1")
	f.Add("a|b|1\nc|d|2")
	f.Add(`a\|b|c\\d|3`)
	f.Add("a|b|007")
	f.Add("nesmysl")

	f.Fuzz(func(t *testing.T, s string) {
		recs, err := exercise.Decode(s)
		if err != nil {
			return // nevalidní vstup je v pořádku, hlídáme jen paniku a kanonicitu
		}
		if got := exercise.Encode(recs); got != s {
			t.Fatalf("Encode(Decode(%q)) = %q, chci %q — Decode přijal nekanonický tvar", s, got, s)
		}
	})
}

func TestRenderTableGolden(t *testing.T) {
	golden := filepath.Join("testdata", "table.golden")
	got := exercise.RenderTable(goldenRecords)

	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("vytvoření testdata: %v", err)
		}
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatalf("zápis %s: %v", golden, err)
		}
		t.Logf("golden soubor %s přepsán", golden)
		return
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("čtení %s: %v (vyrob ho přes go test -run TestRenderTableGolden -update)", golden, err)
	}
	if got != string(want) {
		t.Errorf("RenderTable se liší od %s\n--- chci ---\n%s\n--- mám ---\n%s", golden, want, got)
	}
}

func TestRenderTablePrazdny(t *testing.T) {
	if got := exercise.RenderTable(nil); got != "" {
		t.Errorf("RenderTable(nil) = %q, chci prázdný řetězec", got)
	}
	if got := exercise.RenderTableFast(nil); got != "" {
		t.Errorf("RenderTableFast(nil) = %q, chci prázdný řetězec", got)
	}
}

func TestRenderTableFastShoda(t *testing.T) {
	// Náhodná data, aby výstup nešlo zadrátovat.
	rnd := rand.New(rand.NewSource(20240115))
	for run := 0; run < 20; run++ {
		recs := make([]exercise.Record, rnd.Intn(6)+1)
		for i := range recs {
			recs[i] = exercise.Record{
				ID:    strings.Repeat("id", rnd.Intn(4)+1),
				Name:  strings.Repeat("Žo", rnd.Intn(5)+1),
				Score: rnd.Intn(200000) - 100000,
			}
		}
		slow := exercise.RenderTable(recs)
		fast := exercise.RenderTableFast(recs)
		if slow != fast {
			t.Fatalf("RenderTableFast se liší od RenderTable\n--- slow ---\n%s\n--- fast ---\n%s", slow, fast)
		}
	}
}

func TestRenderTableZarovnani(t *testing.T) {
	out := exercise.RenderTable(goldenRecords)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != len(goldenRecords)+2 {
		t.Fatalf("tabulka má %d řádků, chci %d (hlavička + oddělovač + data)", len(lines), len(goldenRecords)+2)
	}
	width := len([]rune(lines[0]))
	for i, l := range lines {
		if got := len([]rune(l)); got != width {
			t.Errorf("řádek %d má šířku %d run, chci %d:\n%s", i, got, width, out)
		}
	}
	if !strings.HasSuffix(out, "\n") {
		t.Error("tabulka má končit koncem řádku")
	}
}

func BenchmarkEncode(b *testing.B) {
	recs := make([]exercise.Record, 200)
	for i := range recs {
		recs[i] = exercise.Record{ID: "id", Name: "Jméno|s escapem", Score: i}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkString = exercise.Encode(recs)
	}
}

func BenchmarkDecode(b *testing.B) {
	recs := make([]exercise.Record, 200)
	for i := range recs {
		recs[i] = exercise.Record{ID: "id", Name: "Jméno|s escapem", Score: i}
	}
	in := exercise.Encode(recs)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := exercise.Decode(in)
		if err != nil {
			b.Fatal(err)
		}
		sinkRecords = out
	}
}

func benchRecords() []exercise.Record {
	recs := make([]exercise.Record, 100)
	for i := range recs {
		recs[i] = exercise.Record{ID: "id-x", Name: "Žofie Nováková", Score: i * 37}
	}
	return recs
}

func BenchmarkRenderTable(b *testing.B) {
	recs := benchRecords()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkString = exercise.RenderTable(recs)
	}
}

func BenchmarkRenderTableFast(b *testing.B) {
	recs := benchRecords()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkString = exercise.RenderTableFast(recs)
	}
}
