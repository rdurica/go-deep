package csvstats_test

import (
	"math"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rdurica/go-deep/projects/p01-csv-cli/csvstats"
)

func equalFloat(t *testing.T, got, want float64, format string, args ...any) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Errorf(format, args...)
	}
}

func loadSample(t *testing.T) []csvstats.Record {
	t.Helper()
	recs, err := csvstats.LoadFile(filepath.Join("testdata", "sample.csv"))
	if err != nil {
		t.Fatalf("LoadFile(testdata/sample.csv) = _, %v, chci nil", err)
	}
	return recs
}

func TestParseRecords(t *testing.T) {
	in := "name,amount,category\nAda,120.50,food\nBob,80,transport\n"

	got, err := csvstats.ParseRecords(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ParseRecords(...) = _, %v, chci nil", err)
	}
	want := []csvstats.Record{
		{Name: "Ada", Amount: 120.50, Category: "food"},
		{Name: "Bob", Amount: 80, Category: "transport"},
	}
	if len(got) != len(want) {
		t.Fatalf("ParseRecords(...) vrátilo %d záznamů, chci %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("záznam %d = %+v, chci %+v", i, got[i], want[i])
		}
	}
}

func TestParseRecordsChyby(t *testing.T) {
	tests := map[string]string{
		"prázdný vstup":     "",
		"špatná hlavička":   "jmeno,castka,kategorie\nAda,1,food\n",
		"chybí sloupec":     "name,amount,category\nAda,120.50\n",
		"nečíselná částka":  "name,amount,category\nAda,hodne,food\n",
		"prázdné jméno":     "name,amount,category\n ,120.50,food\n",
		"prázdná kategorie": "name,amount,category\nAda,120.50, \n",
	}
	for name, in := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := csvstats.ParseRecords(strings.NewReader(in)); err == nil {
				t.Errorf("ParseRecords(%q) = _, nil, chci chybu", in)
			}
		})
	}
}

func TestParseRecordsCisloRadku(t *testing.T) {
	in := "name,amount,category\nAda,1,food\nBob,rozbito,fun\n"
	_, err := csvstats.ParseRecords(strings.NewReader(in))
	if err == nil {
		t.Fatal("ParseRecords(rozbitý řádek) = _, nil, chci chybu")
	}
	if !strings.Contains(err.Error(), "řádek 3") {
		t.Errorf("chyba = %q, chci zmínku o řádku 3", err)
	}
}

func TestSummarize(t *testing.T) {
	s := csvstats.Summarize(loadSample(t))

	if s.Records != 7 {
		t.Errorf("Records = %d, chci 7", s.Records)
	}
	equalFloat(t, s.Total, 581.60, "Total = %v, chci %v", s.Total, 581.60)

	wantOrder := []string{"food", "fun", "transport"}
	if len(s.Categories) != len(wantOrder) {
		t.Fatalf("Categories = %+v, chci %d položek", s.Categories, len(wantOrder))
	}
	for i, want := range wantOrder {
		if s.Categories[i].Category != want {
			t.Errorf("Categories[%d] = %q, chci %q (řadí se sestupně podle součtu)", i, s.Categories[i].Category, want)
		}
	}

	food := s.Categories[0]
	if food.Count != 3 {
		t.Errorf("food.Count = %d, chci 3", food.Count)
	}
	equalFloat(t, food.Total, 365.85, "food.Total = %v, chci %v", food.Total, 365.85)
	equalFloat(t, food.Average, 365.85/3, "food.Average = %v, chci %v", food.Average, 365.85/3)
}

func TestSummarizePrazdny(t *testing.T) {
	s := csvstats.Summarize(nil)
	if s.Records != 0 || s.Total != 0 {
		t.Errorf("Summarize(nil) = %+v, chci nulový souhrn", s)
	}
	if s.Categories == nil {
		t.Error("Categories = nil, chci prázdný slice")
	}
}

func TestTopN(t *testing.T) {
	recs := loadSample(t)

	tests := []struct {
		n    int
		want []string
	}{
		{0, nil},
		{-3, nil},
		{2, []string{"Grace", "Ada"}},
		{99, []string{"Grace", "Ada", "Bob", "Ken", "Barbara", "Rob", "Linus"}},
	}
	for _, tt := range tests {
		got := csvstats.TopN(recs, tt.n)
		if len(got) != len(tt.want) {
			t.Fatalf("TopN(recs, %d) vrátilo %d záznamů (%+v), chci %d", tt.n, len(got), got, len(tt.want))
		}
		for i := range tt.want {
			if got[i].Name != tt.want[i] {
				t.Errorf("TopN(recs, %d)[%d] = %q, chci %q", tt.n, i, got[i].Name, tt.want[i])
			}
		}
	}
}

func TestTopNNemeniVstup(t *testing.T) {
	recs := loadSample(t)
	before := append([]csvstats.Record(nil), recs...)

	csvstats.TopN(recs, 3)

	for i := range before {
		if recs[i] != before[i] {
			t.Fatalf("TopN přeuspořádal vstup: %+v, chci %+v", recs, before)
		}
	}
}

func TestLoadFileNeexistuje(t *testing.T) {
	if _, err := csvstats.LoadFile(filepath.Join(t.TempDir(), "neni.csv")); err == nil {
		t.Error("LoadFile(neexistující) = _, nil, chci chybu")
	}
}
