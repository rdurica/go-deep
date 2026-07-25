package solutions_test

import (
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-02/solutions"
)

func TestApplyDiscount(t *testing.T) {
	tests := []struct {
		name    string
		price   int
		percent int
		want    int
	}{
		{"bez slevy", 1000, 0, 1000},
		{"plná sleva", 1000, 100, 0},
		{"desetina z 19.99", 1999, 10, 1799},
		{"zaokrouhlení půlky nahoru", 5, 50, 3},
		{"třetinová sleva", 1999, 33, 1339},
		{"záporné procento se ořízne", 1000, -20, 1000},
		{"procento nad sto se ořízne", 1000, 250, 0},
		{"nulová cena", 0, 50, 0},
		{"záporná cena", -100, 10, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.ApplyDiscount(tt.price, tt.percent); got != tt.want {
				t.Errorf("ApplyDiscount(%d, %d) = %d, chci %d", tt.price, tt.percent, got, tt.want)
			}
		})
	}
}

func TestApplyDiscountNulovaSlevaNemeniCenu(t *testing.T) {
	// Vlastnost, kterou nejde splnit zadrátovanou hodnotou.
	for price := 0; price < 5000; price += 37 {
		if got := exercise.ApplyDiscount(price, 0); got != price {
			t.Fatalf("ApplyDiscount(%d, 0) = %d, chci %d", price, got, price)
		}
	}
}

func TestTotalCents(t *testing.T) {
	tests := []struct {
		name  string
		items []exercise.Item
		want  int
	}{
		{"nil slice", nil, 0},
		{"prázdný slice", []exercise.Item{}, 0},
		{
			"jedna položka",
			[]exercise.Item{{Name: "kava", PriceCents: 4500, Qty: 1}},
			4500,
		},
		{
			"množství se počítá",
			[]exercise.Item{{Name: "kava", PriceCents: 4500, Qty: 3}},
			13500,
		},
		{
			"víc položek",
			[]exercise.Item{
				{Name: "kava", PriceCents: 4500, Qty: 2},
				{Name: "caj", PriceCents: 3200, Qty: 1},
			},
			12200,
		},
		{
			"nulové a záporné množství se ignoruje",
			[]exercise.Item{
				{Name: "kava", PriceCents: 4500, Qty: 0},
				{Name: "caj", PriceCents: 3200, Qty: -2},
				{Name: "voda", PriceCents: 1000, Qty: 1},
			},
			1000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.TotalCents(tt.items); got != tt.want {
				t.Errorf("TotalCents(%v) = %d, chci %d", tt.items, got, tt.want)
			}
		})
	}
}

func TestTotalCentsNemeniVstup(t *testing.T) {
	items := []exercise.Item{
		{Name: "kava", PriceCents: 4500, Qty: 2},
		{Name: "caj", PriceCents: 3200, Qty: 1},
	}
	before := append([]exercise.Item(nil), items...)
	exercise.TotalCents(items)
	for i := range items {
		if items[i] != before[i] {
			t.Fatalf("TotalCents změnila vstup: items[%d] = %v, chci %v", i, items[i], before[i])
		}
	}
}

func TestCheapest(t *testing.T) {
	items := []exercise.Item{
		{Name: "kava", PriceCents: 4500, Qty: 1},
		{Name: "voda", PriceCents: 1000, Qty: 5},
		{Name: "caj", PriceCents: 3200, Qty: 2},
	}
	got, ok := exercise.Cheapest(items)
	if !ok {
		t.Fatal("Cheapest vrátila ok=false, chci true")
	}
	if got.Name != "voda" {
		t.Errorf("Cheapest(...).Name = %q, chci %q", got.Name, "voda")
	}
	if got.PriceCents != 1000 || got.Qty != 5 {
		t.Errorf("Cheapest(...) = %v, chci celou položku voda", got)
	}
}

func TestCheapestPrazdnyVstup(t *testing.T) {
	for _, items := range [][]exercise.Item{nil, {}} {
		got, ok := exercise.Cheapest(items)
		if ok {
			t.Errorf("Cheapest(%v) vrátila ok=true, chci false", items)
		}
		if got != (exercise.Item{}) {
			t.Errorf("Cheapest(%v) = %v, chci zero value Item{}", items, got)
		}
	}
}

func TestCheapestPrvniPriShode(t *testing.T) {
	items := []exercise.Item{
		{Name: "prvni", PriceCents: 100, Qty: 1},
		{Name: "druhy", PriceCents: 100, Qty: 1},
	}
	got, _ := exercise.Cheapest(items)
	if got.Name != "prvni" {
		t.Errorf("Cheapest při shodě ceny vrátila %q, chci %q", got.Name, "prvni")
	}
}

func TestCheapestVraciKopii(t *testing.T) {
	items := []exercise.Item{{Name: "voda", PriceCents: 1000, Qty: 1}}
	got, _ := exercise.Cheapest(items)
	got.PriceCents = 1
	if items[0].PriceCents != 1000 {
		t.Errorf("změna vrácené položky ovlivnila slice: %d, chci %d", items[0].PriceCents, 1000)
	}
}

func newTestCatalog() *exercise.Catalog {
	return exercise.NewCatalog([]exercise.Item{
		{Name: "kava", PriceCents: 4500, Qty: 1},
		{Name: "caj", PriceCents: 3200, Qty: 1},
		{Name: "voda", PriceCents: 1000, Qty: 1},
	})
}

func TestCatalogPrice(t *testing.T) {
	c := newTestCatalog()

	price, ok := c.Price("caj")
	if !ok || price != 3200 {
		t.Errorf("Price(%q) = (%d, %v), chci (%d, %v)", "caj", price, ok, 3200, true)
	}

	price, ok = c.Price("pivo")
	if ok || price != 0 {
		t.Errorf("Price(%q) = (%d, %v), chci (%d, %v)", "pivo", price, ok, 0, false)
	}
}

func TestCatalogPrazdny(t *testing.T) {
	c := exercise.NewCatalog(nil)
	if _, ok := c.Price("kava"); ok {
		t.Error("prázdný ceník nemá nic znát")
	}
}

func TestCatalogKopirujeVstup(t *testing.T) {
	items := []exercise.Item{{Name: "kava", PriceCents: 4500, Qty: 1}}
	c := exercise.NewCatalog(items)
	items[0].PriceCents = 1

	if price, _ := c.Price("kava"); price != 4500 {
		t.Errorf("po změně vstupního slice Price(%q) = %d, chci %d", "kava", price, 4500)
	}
}

func TestCatalogCheckout(t *testing.T) {
	c := newTestCatalog()

	tests := []struct {
		name    string
		names   []string
		percent int
		want    int
		wantOK  bool
	}{
		{"jedna položka bez slevy", []string{"kava"}, 0, 4500, true},
		{"dvě položky bez slevy", []string{"kava", "voda"}, 0, 5500, true},
		{"stejná položka dvakrát", []string{"voda", "voda"}, 0, 2000, true},
		{"se slevou", []string{"kava", "caj"}, 10, 6930, true},
		{"prázdná objednávka", nil, 20, 0, true},
		{"neznámá položka", []string{"kava", "pivo"}, 0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := c.Checkout(tt.names, tt.percent)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("Checkout(%v, %d) = (%d, %v), chci (%d, %v)",
					tt.names, tt.percent, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
