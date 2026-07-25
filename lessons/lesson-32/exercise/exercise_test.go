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

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		in   exercise.Product
		want error
	}{
		{"platný produkt", p("A-1", "Kniha", 1999), nil},
		{"nulová cena je v pořádku", p("A-1", "Kniha", 0), nil},
		{"prázdné SKU", p("", "Kniha", 1999), exercise.ErrEmptySKU},
		{"SKU jen z mezer", p("   ", "Kniha", 1999), exercise.ErrEmptySKU},
		{"prázdné jméno", p("A-1", "", 1999), exercise.ErrEmptyName},
		{"jméno jen z mezer", p("A-1", "\t ", 1999), exercise.ErrEmptyName},
		{"záporná cena", p("A-1", "Kniha", -1), exercise.ErrNegativeCents},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := exercise.Validate(tt.in)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Validate(%+v) = %v, chci %v", tt.in, err, tt.want)
			}
		})
	}
}

func TestValidateChybaNeseSKU(t *testing.T) {
	err := exercise.Validate(p("SKU-42", "", 100))
	if err == nil {
		t.Fatal("Validate u prázdného jména musí vrátit chybu")
	}
	if !strings.Contains(err.Error(), "SKU-42") {
		t.Errorf("chyba %q neobsahuje SKU produktu", err.Error())
	}
}

func TestBuildCatalog(t *testing.T) {
	t.Run("prázdný katalog není chyba", func(t *testing.T) {
		c, err := exercise.BuildCatalog()
		if err != nil {
			t.Fatalf("BuildCatalog() = %v, chci nil", err)
		}
		if got := len(c.All()); got != 0 {
			t.Errorf("len(All()) = %d, chci 0", got)
		}
	})

	t.Run("neplatný produkt propadne validací", func(t *testing.T) {
		_, err := exercise.BuildCatalog(p("A-1", "Kniha", 100), p("B-2", "", 100))
		if !errors.Is(err, exercise.ErrEmptyName) {
			t.Fatalf("BuildCatalog = %v, chci ErrEmptyName", err)
		}
	})

	t.Run("duplicitní SKU", func(t *testing.T) {
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

	var nilCatalog *exercise.Catalog
	if _, err := nilCatalog.BySKU("A-1"); !errors.Is(err, exercise.ErrNotFound) {
		t.Errorf("BySKU na nil katalogu = %v, chci ErrNotFound", err)
	}
}

func TestCatalogAllJeSerazene(t *testing.T) {
	c, err := exercise.BuildCatalog(
		p("C-3", "Guma", 100),
		p("A-1", "Kniha", 1999),
		p("B-2", "Tužka", 250),
	)
	if err != nil {
		t.Fatalf("BuildCatalog = %v", err)
	}

	// Mapa nemá pořadí, takže test běží víckrát — kdyby se pořadí bralo
	// z iterace mapy, dřív nebo později to praskne.
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

func TestTotalOf(t *testing.T) {
	tests := []struct {
		name  string
		items []exercise.Item
		want  int64
	}{
		{"nil košík", nil, 0},
		{"prázdný košík", []exercise.Item{}, 0},
		{
			"jedna položka",
			[]exercise.Item{{Product: p("A-1", "Kniha", 1999), Qty: 3}},
			5997,
		},
		{
			"víc položek",
			[]exercise.Item{
				{Product: p("A-1", "Kniha", 1999), Qty: 2},
				{Product: p("B-2", "Tužka", 250), Qty: 4},
				{Product: p("C-3", "Zdarma", 0), Qty: 10},
			},
			4998,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exercise.TotalOf(tt.items)
			if err != nil {
				t.Fatalf("TotalOf = %v, chci nil", err)
			}
			if got != tt.want {
				t.Errorf("TotalOf = %d, chci %d", got, tt.want)
			}
		})
	}
}

func TestTotalOfChyby(t *testing.T) {
	tests := []struct {
		name  string
		items []exercise.Item
		want  error
	}{
		{
			"nulové množství",
			[]exercise.Item{{Product: p("A-1", "Kniha", 1999), Qty: 0}},
			exercise.ErrInvalidQty,
		},
		{
			"záporné množství",
			[]exercise.Item{{Product: p("A-1", "Kniha", 1999), Qty: -2}},
			exercise.ErrInvalidQty,
		},
		{
			"neplatný produkt",
			[]exercise.Item{{Product: p("", "Kniha", 1999), Qty: 1}},
			exercise.ErrEmptySKU,
		},
		{
			"přetečení násobení",
			[]exercise.Item{{Product: p("A-1", "Drahé", math.MaxInt64), Qty: 2}},
			exercise.ErrOverflow,
		},
		{
			"přetečení součtu",
			[]exercise.Item{
				{Product: p("A-1", "Drahé", math.MaxInt64), Qty: 1},
				{Product: p("B-2", "Taky drahé", 1), Qty: 1},
			},
			exercise.ErrOverflow,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := exercise.TotalOf(tt.items)
			if !errors.Is(err, tt.want) {
				t.Errorf("TotalOf = %v, chci %v", err, tt.want)
			}
		})
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

func TestIDGen(t *testing.T) {
	g := exercise.NewIDGen("prod")

	want := []string{"prod-000001", "prod-000002", "prod-000003"}
	for _, w := range want {
		if got := g.NewID(); got != w {
			t.Fatalf("NewID() = %q, chci %q", got, w)
		}
	}

	def := exercise.NewIDGen("")
	if got := def.NewID(); got != "id-000001" {
		t.Errorf("NewID() s prázdným prefixem = %q, chci %q", got, "id-000001")
	}
}

func TestIDGenTvar(t *testing.T) {
	re := regexp.MustCompile(`^order-\d{6}$`)
	g := exercise.NewIDGen("order")
	for i := 0; i < 5; i++ {
		if id := g.NewID(); !re.MatchString(id) {
			t.Errorf("NewID() = %q, nesedí na %s", id, re)
		}
	}
}

func TestIDGenSoubezne(t *testing.T) {
	const workers, perWorker = 8, 200

	g := exercise.NewIDGen("c")
	ids := make([]string, 0, workers*perWorker)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := make([]string, 0, perWorker)
			for j := 0; j < perWorker; j++ {
				local = append(local, g.NewID())
			}
			mu.Lock()
			ids = append(ids, local...)
			mu.Unlock()
		}()
	}
	wg.Wait()

	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("duplicitní ID %q — generátor není bezpečný pro souběh", id)
		}
		seen[id] = true
	}
	if len(seen) != workers*perWorker {
		t.Errorf("unikátních ID = %d, chci %d", len(seen), workers*perWorker)
	}
}
