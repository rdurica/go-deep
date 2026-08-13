package exercise_test

import (
	"errors"
	"math"
	"regexp"
	"strings"
	"sync"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-32/exercise"
)

func p(sku, name string, cents int64) exercise.Product {
	return exercise.Product{SKU: sku, Name: name, Cents: cents}
}

func mustCatalog(t *testing.T, products ...exercise.Product) *exercise.Catalog {
	t.Helper()
	c, err := exercise.CatalogFixture(products...)
	if err != nil {
		t.Fatalf("CatalogFixture = %v", err)
	}
	return c
}

func TestSnapshotIndependent(t *testing.T) {
	orig := mustCatalog(t, p("A-1", "Kniha", 1999))
	got := exercise.Snapshot(orig)
	if got == orig {
		t.Fatal("Snapshot vrátil stejný pointer jako originál")
	}
	_, _ = got.BySKU("A-1")
	// Zápis přes BySKU nemění mapu — ověříme nezávislost přes druhý snapshot.
	got2 := exercise.Snapshot(orig)
	if got2 == orig || got2 == got {
		t.Error("Snapshot musí vracet novou instanci katalogu")
	}
}

func TestSnapshotNil(t *testing.T) {
	if got := exercise.Snapshot(nil); got != nil {
		t.Errorf("Snapshot(nil) = %v, chci nil", got)
	}
}

func TestBuildCatalog(t *testing.T) {
	t.Run("empty catalog is not error", func(t *testing.T) {
		c, err := exercise.BuildCatalog()
		if err != nil {
			t.Fatalf("BuildCatalog() = %v, chci nil", err)
		}
		if got := len(c.All()); got != 0 {
			t.Errorf("len(All()) = %d, chci 0", got)
		}
	})

	t.Run("invalid product fails validation", func(t *testing.T) {
		_, err := exercise.BuildCatalog(p("A-1", "Kniha", 100), p("B-2", "", 100))
		if !errors.Is(err, exercise.ErrEmptyName) {
			t.Fatalf("BuildCatalog = %v, chci ErrEmptyName", err)
		}
	})

	t.Run("duplicate SKU", func(t *testing.T) {
		_, err := exercise.BuildCatalog(p("A-1", "Kniha", 100), p("A-1", "Jiná kniha", 200))
		if !errors.Is(err, exercise.ErrDuplicateSKU) {
			t.Fatalf("BuildCatalog = %v, chci ErrDuplicateSKU", err)
		}
	})
}

func TestCatalogBySKU(t *testing.T) {
	c, err := exercise.BuildCatalog(p("A-1", "Kniha", 1999), p("B-2", "Tužka", 250))
	if err != nil {
		t.Fatalf("BuildCatalog = %v", err)
	}

	got, err := c.BySKU("B-2")
	if err != nil {
		t.Fatalf("BySKU(%q) = %v, chci nil", "B-2", err)
	}
	if got.Name != "Tužka" || got.Cents != 250 {
		t.Errorf("BySKU(%q) = %+v, chci Tužka/250", "B-2", got)
	}

	if _, err := c.BySKU("NEEXISTUJE"); !errors.Is(err, exercise.ErrNotFound) {
		t.Errorf("BySKU(neznámé) = %v, chci ErrNotFound", err)
	}
}

func TestCatalogAllIsSorted(t *testing.T) {
	c, err := exercise.BuildCatalog(
		p("C-3", "Guma", 100),
		p("A-1", "Kniha", 1999),
		p("B-2", "Tužka", 250),
	)
	if err != nil {
		t.Fatalf("BuildCatalog = %v", err)
	}

	want := []string{"A-1", "B-2", "C-3"}
	for i := 0; i < 20; i++ {
		var got []string
		for _, prod := range c.All() {
			got = append(got, prod.SKU)
		}
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("All() = %v, chci %v", got, want)
		}
	}
}

func TestPriceOf(t *testing.T) {
	c, err := exercise.BuildCatalog(p("A-1", "Kniha", 1999), p("B-2", "Tužka", 250))
	if err != nil {
		t.Fatalf("BuildCatalog = %v", err)
	}

	got, err := exercise.PriceOf(c, "A-1", 3)
	if err != nil {
		t.Fatalf("PriceOf = %v, chci nil", err)
	}
	if got != 5997 {
		t.Errorf("PriceOf(A-1, 3) = %d, chci 5997", got)
	}

	if _, err := exercise.PriceOf(c, "NIC", 1); !errors.Is(err, exercise.ErrNotFound) {
		t.Errorf("PriceOf(neznámé) = %v, chci ErrNotFound", err)
	}
	if _, err := exercise.PriceOf(c, "A-1", 0); !errors.Is(err, exercise.ErrInvalidQty) {
		t.Errorf("PriceOf(qty=0) = %v, chci ErrInvalidQty", err)
	}
}

func TestTotalOf(t *testing.T) {
	items := []exercise.Item{
		{Product: p("A-1", "Kniha", 1999), Qty: 2},
		{Product: p("B-2", "Tužka", 250), Qty: 4},
	}
	got, err := exercise.TotalOf(items)
	if err != nil {
		t.Fatalf("TotalOf = %v", err)
	}
	if got != 4998 {
		t.Errorf("TotalOf = %d, chci 4998", got)
	}
}

func TestTotalOfErrors(t *testing.T) {
	_, err := exercise.TotalOf([]exercise.Item{{Product: p("A-1", "Drahé", math.MaxInt64), Qty: 2}})
	if !errors.Is(err, exercise.ErrOverflow) {
		t.Errorf("TotalOf overflow = %v, chci ErrOverflow", err)
	}
}

func TestIDGen(t *testing.T) {
	g := exercise.NewIDGen("prod")
	if got := g.NewID(); got != "prod-000001" {
		t.Errorf("NewID() = %q, chci %q", got, "prod-000001")
	}
}

func TestIDGenConcurrent(t *testing.T) {
	g := exercise.NewIDGen("c")
	ids := make([]string, 0, 16)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := g.NewID()
			mu.Lock()
			ids = append(ids, id)
			mu.Unlock()
		}()
	}
	wg.Wait()
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("duplicitní ID %q", id)
		}
		seen[id] = true
		if !regexp.MustCompile(`^c-\d{6}$`).MatchString(id) {
			t.Errorf("NewID() = %q, nesedí formát", id)
		}
	}
}
