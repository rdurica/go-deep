package csvstats_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/rdurica/go-deep/projects/p01-csv-cli/csvstats"
)

// go test ./... -update přepíše golden soubory podle aktuálního výstupu.
var update = flag.Bool("update", false, "přepiš golden soubory")

func assertGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name)

	if *update {
		if err := os.WriteFile(path, got, 0o600); err != nil {
			t.Fatalf("zápis golden souboru %s selhal: %v", path, err)
		}
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("čtení golden souboru %s selhalo (spusť go test -update): %v", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("výstup se liší od %s\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}

func TestRenderSummaryGolden(t *testing.T) {
	recs, err := csvstats.LoadFile(filepath.Join("testdata", "sample.csv"))
	if err != nil {
		t.Fatalf("LoadFile(...) = _, %v, chci nil", err)
	}

	var buf bytes.Buffer
	if err := csvstats.RenderSummary(&buf, csvstats.Summarize(recs)); err != nil {
		t.Fatalf("RenderSummary(...) = %v, chci nil", err)
	}
	assertGolden(t, "summary.golden", buf.Bytes())
}

func TestRenderTopGolden(t *testing.T) {
	recs, err := csvstats.LoadFile(filepath.Join("testdata", "sample.csv"))
	if err != nil {
		t.Fatalf("LoadFile(...) = _, %v, chci nil", err)
	}

	var buf bytes.Buffer
	if err := csvstats.RenderTop(&buf, csvstats.TopN(recs, 3)); err != nil {
		t.Fatalf("RenderTop(...) = %v, chci nil", err)
	}
	assertGolden(t, "top3.golden", buf.Bytes())
}

func TestRenderSummaryPrazdny(t *testing.T) {
	var buf bytes.Buffer
	if err := csvstats.RenderSummary(&buf, csvstats.Summarize(nil)); err != nil {
		t.Fatalf("RenderSummary(prázdný) = %v, chci nil", err)
	}
	if buf.Len() == 0 {
		t.Error("RenderSummary(prázdný) nic nevypsal, chci aspoň hlavičku a řádek CELKEM")
	}
}
