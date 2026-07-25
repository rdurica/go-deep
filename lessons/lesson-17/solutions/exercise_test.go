package solutions_test

import (
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-17/solutions"
)

// equalFloat je pomocník pro porovnání desetinných čísel s tolerancí.
// t.Helper() zajistí, že se chyba nahlásí na řádku volajícího.
func equalFloat(t *testing.T, got, want float64, format string, args ...any) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Errorf(format, args...)
	}
}

func sampleRecords() []exercise.Record {
	return []exercise.Record{
		{Name: "Ada", Amount: 120.50, Category: "food"},
		{Name: "Bob", Amount: 80, Category: "transport"},
		{Name: "Grace", Amount: 200.25, Category: "food"},
		{Name: "Linus", Amount: 15.75, Category: "transport"},
		{Name: "Ken", Amount: 60, Category: "fun"},
	}
}

func TestMedian(t *testing.T) {
	tests := []struct {
		name   string
		in     []float64
		want   float64
		wantOK bool
	}{
		{"prázdný vstup", nil, 0, false},
		{"prázdný slice", []float64{}, 0, false},
		{"jeden prvek", []float64{4.2}, 4.2, true},
		{"lichý počet", []float64{3, 1, 2}, 2, true},
		{"sudý počet", []float64{4, 1, 3, 2}, 2.5, true},
		{"záporná čísla", []float64{-5, -1, -3}, -3, true},
		{"duplicity", []float64{2, 2, 2, 2}, 2, true},
		{"už seřazeno", []float64{1, 2, 3, 4, 5}, 3, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := exercise.Median(tt.in)
			if ok != tt.wantOK {
				t.Fatalf("Median(%v) = _, %v, chci %v", tt.in, ok, tt.wantOK)
			}
			equalFloat(t, got, tt.want, "Median(%v) = %v, chci %v", tt.in, got, tt.want)
		})
	}
}

func TestMedianNemeniVstup(t *testing.T) {
	in := []float64{9, 1, 5, 3}
	before := append([]float64(nil), in...)

	exercise.Median(in)

	for i := range before {
		if in[i] != before[i] {
			t.Fatalf("Median vstup přeuspořádal: %v, chci %v", in, before)
		}
	}
}

func TestMedianNahodnaData(t *testing.T) {
	// Medián permutace 1..n je znám dopředu, takže test nejde splnit
	// zadrátovanou konstantou.
	rnd := rand.New(rand.NewSource(42))
	for _, n := range []int{1, 2, 3, 10, 101} {
		nums := make([]float64, n)
		for i := range nums {
			nums[i] = float64(i + 1)
		}
		rnd.Shuffle(n, func(i, j int) { nums[i], nums[j] = nums[j], nums[i] })

		want := float64(n+1) / 2
		got, ok := exercise.Median(nums)
		if !ok {
			t.Fatalf("Median(permutace 1..%d) = _, false, chci true", n)
		}
		equalFloat(t, got, want, "Median(permutace 1..%d) = %v, chci %v", n, got, want)
	}
}

func TestParseRecords(t *testing.T) {
	in := "name,amount,category\nAda,120.50,food\nBob,80,transport\n"

	got, err := exercise.ParseRecords(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ParseRecords(...) = _, %v, chci nil", err)
	}
	want := []exercise.Record{
		{Name: "Ada", Amount: 120.50, Category: "food"},
		{Name: "Bob", Amount: 80, Category: "transport"},
	}
	if len(got) != len(want) {
		t.Fatalf("ParseRecords(...) vrátilo %d záznamů, chci %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("záznam %d = %+v, chci %+v", i, got[i], want[i])
		}
	}
}

func TestParseRecordsJenHlavicka(t *testing.T) {
	got, err := exercise.ParseRecords(strings.NewReader("name,amount,category\n"))
	if err != nil {
		t.Fatalf("ParseRecords(jen hlavička) = _, %v, chci nil", err)
	}
	if len(got) != 0 {
		t.Errorf("ParseRecords(jen hlavička) = %+v, chci prázdný výsledek", got)
	}
}

func TestParseRecordsChyby(t *testing.T) {
	tests := map[string]string{
		"prázdný vstup":     "",
		"špatná hlavička":   "jmeno,castka,kategorie\nAda,1,food\n",
		"chybí sloupec":     "name,amount,category\nAda,120.50\n",
		"sloupec navíc":     "name,amount,category\nAda,120.50,food,navic\n",
		"nečíselná částka":  "name,amount,category\nAda,hodne,food\n",
		"prázdné jméno":     "name,amount,category\n ,120.50,food\n",
		"prázdná kategorie": "name,amount,category\nAda,120.50, \n",
	}
	for name, in := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := exercise.ParseRecords(strings.NewReader(in)); err == nil {
				t.Errorf("ParseRecords(%q) = _, nil, chci chybu", in)
			}
		})
	}
}

func TestSumByCategory(t *testing.T) {
	got := exercise.SumByCategory(sampleRecords())
	want := map[string]float64{"food": 320.75, "transport": 95.75, "fun": 60}

	if len(got) != len(want) {
		t.Fatalf("SumByCategory(...) = %v, chci %d kategorií", got, len(want))
	}
	for cat, w := range want {
		equalFloat(t, got[cat], w, "SumByCategory(...)[%q] = %v, chci %v", cat, got[cat], w)
	}
}

func TestSumByCategoryPrazdny(t *testing.T) {
	got := exercise.SumByCategory(nil)
	if got == nil {
		t.Fatal("SumByCategory(nil) = nil, chci prázdnou mapu")
	}
	if len(got) != 0 {
		t.Errorf("SumByCategory(nil) = %v, chci prázdnou mapu", got)
	}
}

func TestTopN(t *testing.T) {
	recs := sampleRecords()

	tests := []struct {
		n     int
		want  []string
		popis string
	}{
		{0, nil, "n=0 nevrací nic"},
		{-1, nil, "záporné n nevrací nic"},
		{1, []string{"Grace"}, "největší útrata"},
		{3, []string{"Grace", "Ada", "Bob"}, "tři největší"},
		{99, []string{"Grace", "Ada", "Bob", "Ken", "Linus"}, "n větší než délka"},
	}
	for _, tt := range tests {
		t.Run(tt.popis, func(t *testing.T) {
			got := exercise.TopN(recs, tt.n)
			if len(got) != len(tt.want) {
				t.Fatalf("TopN(recs, %d) vrátilo %d záznamů (%+v), chci %d", tt.n, len(got), got, len(tt.want))
			}
			for i := range tt.want {
				if got[i].Name != tt.want[i] {
					t.Errorf("TopN(recs, %d)[%d] = %q, chci %q", tt.n, i, got[i].Name, tt.want[i])
				}
			}
		})
	}
}

func TestTopNStabilita(t *testing.T) {
	recs := []exercise.Record{
		{Name: "prvni", Amount: 10, Category: "a"},
		{Name: "druhy", Amount: 10, Category: "b"},
		{Name: "treti", Amount: 10, Category: "c"},
	}
	got := exercise.TopN(recs, 3)
	want := []string{"prvni", "druhy", "treti"}
	for i := range want {
		if got[i].Name != want[i] {
			t.Fatalf("TopN při shodných částkách = %+v, chci pořadí %v", got, want)
		}
	}
}

func TestTopNNemeniVstup(t *testing.T) {
	recs := sampleRecords()
	before := append([]exercise.Record(nil), recs...)

	exercise.TopN(recs, 2)

	for i := range before {
		if recs[i] != before[i] {
			t.Fatalf("TopN přeuspořádal vstup: %+v, chci %+v", recs, before)
		}
	}
}

func TestLoadFileTestdata(t *testing.T) {
	got, err := exercise.LoadFile(filepath.Join("testdata", "sample.csv"))
	if err != nil {
		t.Fatalf("LoadFile(testdata/sample.csv) = _, %v, chci nil", err)
	}
	if len(got) != 5 {
		t.Fatalf("LoadFile(...) vrátilo %d záznamů, chci 5: %+v", len(got), got)
	}
	sums := exercise.SumByCategory(got)
	equalFloat(t, sums["food"], 320.75, "součet food = %v, chci %v", sums["food"], 320.75)
}

func TestLoadFileTempDir(t *testing.T) {
	dir := t.TempDir() // uklidí se automaticky po testu
	path := filepath.Join(dir, "spend.csv")
	content := "name,amount,category\nZoe,42.5,fun\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("příprava souboru selhala: %v", err)
	}
	t.Cleanup(func() {
		// Jen ukázka: t.TempDir smaže adresář sám, tohle je hlídač navíc.
		if _, err := os.Stat(dir); err != nil {
			t.Logf("dočasný adresář už zmizel: %v", err)
		}
	})

	got, err := exercise.LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile(%s) = _, %v, chci nil", path, err)
	}
	if len(got) != 1 {
		t.Fatalf("LoadFile(...) vrátilo %d záznamů, chci 1: %+v", len(got), got)
	}
	want := exercise.Record{Name: "Zoe", Amount: 42.5, Category: "fun"}
	if got[0] != want {
		t.Errorf("LoadFile(...)[0] = %+v, chci %+v", got[0], want)
	}
}

func TestLoadFileChyby(t *testing.T) {
	t.Run("neexistující soubor", func(t *testing.T) {
		if _, err := exercise.LoadFile(filepath.Join(t.TempDir(), "neni.csv")); err == nil {
			t.Error("LoadFile(neexistující) = _, nil, chci chybu")
		}
	})

	t.Run("rozbitý obsah", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "bad.csv")
		if err := os.WriteFile(path, []byte("name,amount,category\nAda,neni-cislo,food\n"), 0o600); err != nil {
			t.Fatalf("příprava souboru selhala: %v", err)
		}
		if _, err := exercise.LoadFile(path); err == nil {
			t.Error("LoadFile(rozbitý CSV) = _, nil, chci chybu")
		}
	})
}
