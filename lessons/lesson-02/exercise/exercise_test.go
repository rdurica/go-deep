package exercise_test

import (
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-02/exercise"
)

func TestApplyDiscount(t *testing.T) {
	tests := []struct {
		name    string
		price   int
		percent int
		want    int
	}{
		{"bez slevy", 1000, 0, 1000},
		{"full discount", 1000, 100, 0},
		{"desetina z 19.99", 1999, 10, 1799},
		{"round half up", 5, 50, 3},
		{"one-third discount", 1999, 33, 1339},
		{"negative percent clamped", 1000, -20, 1000},
		{"percent over 100 clamped", 1000, 250, 0},
		{"zero price", 0, 50, 0},
		{"negative price", -100, 10, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exercise.ApplyDiscount(tt.price, tt.percent); got != tt.want {
				t.Errorf("ApplyDiscount(%d, %d) = %d, chci %d", tt.price, tt.percent, got, tt.want)
			}
		})
	}
}

func TestApplyDiscountZeroDoesNotChangePrice(t *testing.T) {
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
		{"empty slice", []exercise.Item{}, 0},
		{
			"one item",
			[]exercise.Item{{Name: "kava", PriceCents: 4500, Qty: 1}},
			4500,
		},
		{
			"quantity is counted",
			[]exercise.Item{{Name: "kava", PriceCents: 4500, Qty: 3}},
			13500,
		},
		{
			"multiple items",
			[]exercise.Item{
				{Name: "kava", PriceCents: 4500, Qty: 2},
				{Name: "caj", PriceCents: 3200, Qty: 1},
			},
			12200,
		},
		{
			"zero and negative qty ignored",
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

func TestTotalCentsDoesNotMutateInput(t *testing.T) {
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

func newTestCatalog() *exercise.Catalog {
	return exercise.NewCatalog([]exercise.Item{
		{Name: "kava", PriceCents: 4500, Qty: 1},
		{Name: "caj", PriceCents: 3200, Qty: 1},
		{Name: "voda", PriceCents: 1000, Qty: 1},
	})
}

func TestNewCatalog(t *testing.T) {
	c := exercise.NewCatalog(nil)
	if c == nil {
		t.Fatal("NewCatalog(nil) vrátil nil, chci prázdný ceník")
	}
	c2 := exercise.NewCatalog([]exercise.Item{{Name: "kava", PriceCents: 4500, Qty: 1}})
	if c2 == nil {
		t.Fatal("NewCatalog vrátil nil, chci ne-nil ceník")
	}
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

func TestCatalogPriceEmptyCatalog(t *testing.T) {
	c := exercise.NewCatalog(nil)
	if _, ok := c.Price("kava"); ok {
		t.Error("prázdný ceník nemá nic znát")
	}
}

func TestCatalogCopiesInput(t *testing.T) {
	items := []exercise.Item{{Name: "kava", PriceCents: 4500, Qty: 1}}
	c := exercise.NewCatalog(items)
	items[0].PriceCents = 1

	if price, ok := c.Price("kava"); ok && price != 4500 {
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
		{"one item no discount", []string{"kava"}, 0, 4500, true},
		{"two items no discount", []string{"kava", "voda"}, 0, 5500, true},
		{"same item twice", []string{"voda", "voda"}, 0, 2000, true},
		{"se slevou", []string{"kava", "caj"}, 10, 6930, true},
		{"empty order", nil, 20, 0, true},
		{"unknown item", []string{"kava", "pivo"}, 0, 0, false},
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
