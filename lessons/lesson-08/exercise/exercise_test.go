package exercise_test

import (
	"reflect"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-08/exercise"
)

func TestCloneMap(t *testing.T) {
	tests := []struct {
		name string
		in   map[string]int
		want map[string]int
	}{
		{"nil", nil, nil},
		{"empty", map[string]int{}, map[string]int{}},
		{"copy", map[string]int{"a": 1, "b": 2}, map[string]int{"a": 1, "b": 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exercise.CloneMap(tt.in)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("CloneMap(nil) = %v, chci nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("CloneMap vrátil nil, chci ne-nil mapu")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CloneMap(%v) = %v, chci %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestCloneMapIndependent(t *testing.T) {
	orig := map[string]int{"a": 1}
	got := exercise.CloneMap(orig)
	got["a"] = 99
	got["b"] = 2
	if orig["a"] != 1 {
		t.Errorf("zápis do klonu změnil originál: orig[a] = %d, chci 1", orig["a"])
	}
	if _, ok := orig["b"]; ok {
		t.Error("zápis nového klíče do klonu přidal klíč do originálu")
	}
}

func TestCloneMapEmptyNotNil(t *testing.T) {
	got := exercise.CloneMap(map[string]int{})
	if got == nil {
		t.Fatal("CloneMap(prázdná) vrátil nil, chci prázdnou ne-nil mapu")
	}
	got["x"] = 1
	if got["x"] != 1 {
		t.Error("do výsledku CloneMap nejde zapisovat")
	}
}

func TestWordCount(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want map[string]int
	}{
		{"empty", []string{}, map[string]int{}},
		{"nil", nil, map[string]int{}},
		{"no duplicates", []string{"a", "b"}, map[string]int{"a": 1, "b": 1}},
		{
			"with duplicates",
			[]string{"go", "php", "go", "go", "php"},
			map[string]int{"go": 3, "php": 2},
		},
		{"empty word is key", []string{"", ""}, map[string]int{"": 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exercise.WordCount(tt.in)
			if got == nil {
				t.Fatal("WordCount vrátil nil mapu, chci ne-nil")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WordCount(%v) = %v, chci %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestWordCountReturnsWritableMap(t *testing.T) {
	// Ne-nil je požadavek proto, aby šlo do výsledku dál zapisovat.
	got := exercise.WordCount(nil)
	got["novy"] = 1
	if got["novy"] != 1 {
		t.Error("do výsledku WordCount nejde zapisovat")
	}
}

func TestNewSet(t *testing.T) {
	s := exercise.NewSet("a", "b", "a", "c", "b")
	if len(s) != 3 {
		t.Errorf("NewSet se třemi unikátními prvky má len %d, chci 3", len(s))
	}
	for _, item := range []string{"a", "b", "c"} {
		if _, ok := s[item]; !ok {
			t.Errorf("prvek %q chybí v množině", item)
		}
	}
	if _, ok := s["d"]; ok {
		t.Error(`prvek "d" je v množině, nemá být`)
	}
}

func TestNewSetNoArgs(t *testing.T) {
	s := exercise.NewSet()
	if s == nil {
		t.Fatal("NewSet() vrátil nil, chci prázdnou ne-nil množinu")
	}
	if len(s) != 0 {
		t.Errorf("NewSet() má len %d, chci 0", len(s))
	}
	// Přímý zápis — na nil mapě by panikoval; Add nevoláme (jiný stub).
	s["x"] = struct{}{}
	if _, ok := s["x"]; !ok {
		t.Error("do množiny z NewSet() nejde přidávat")
	}
}

func TestSetAddIsIdempotent(t *testing.T) {
	s := exercise.Set{}
	s.Add("go")
	s.Add("go")
	s.Add("go")
	if len(s) != 1 {
		t.Errorf("po třech Add stejného prvku je len %d, chci 1", len(s))
	}
	if _, ok := s["go"]; !ok {
		t.Error(`po Add("go") prvek v množině chybí`)
	}
}

func TestSetHas(t *testing.T) {
	s := exercise.Set{"a": {}, "b": {}}
	if !s.Has("a") {
		t.Error(`Has("a") = false, chci true`)
	}
	if s.Has("d") {
		t.Error(`Has("d") = true, chci false`)
	}

	var nilSet exercise.Set
	if nilSet.Has("cokoliv") {
		t.Error("Has na nil množině má vrátit false")
	}
}

func TestAddStockCreatesMissingItem(t *testing.T) {
	inv := exercise.Inventory{}
	exercise.AddStock(inv, "GO-1", 5)

	item, ok := inv["GO-1"]
	if !ok {
		t.Fatal("AddStock nezaložil chybějící položku")
	}
	if item == nil {
		t.Fatal("AddStock uložil nil pointer")
	}
	if item.SKU != "GO-1" {
		t.Errorf("SKU = %q, chci %q", item.SKU, "GO-1")
	}
	if item.Qty != 5 {
		t.Errorf("Qty = %d, chci 5", item.Qty)
	}
}

func TestAddStockIncrementsViaPointer(t *testing.T) {
	inv := exercise.Inventory{"GO-1": {SKU: "GO-1", Qty: 2}}
	before := inv["GO-1"]

	exercise.AddStock(inv, "GO-1", 3)
	exercise.AddStock(inv, "GO-1", 4)

	if inv["GO-1"].Qty != 9 {
		t.Errorf("Qty po 2+3+4 = %d, chci 9", inv["GO-1"].Qty)
	}
	if inv["GO-1"] != before {
		t.Error("AddStock nahradil pointer v mapě, chci mutaci existující položky")
	}
	if before.Qty != 9 {
		t.Errorf("původní pointer ukazuje na Qty %d, chci 9", before.Qty)
	}
}

func TestAddStockEdgeCases(t *testing.T) {
	t.Run("nil inventory does not panic", func(t *testing.T) {
		var inv exercise.Inventory
		exercise.AddStock(inv, "GO-1", 5)
		if len(inv) != 0 {
			t.Errorf("nil inventář má %d položek, chci 0", len(inv))
		}
	})

	t.Run("non-positive n does not create item", func(t *testing.T) {
		inv := exercise.Inventory{}
		exercise.AddStock(inv, "GO-1", 0)
		exercise.AddStock(inv, "GO-2", -3)
		if len(inv) != 0 {
			t.Errorf("inventář má %d položek, chci 0", len(inv))
		}
	})

	t.Run("non-positive n does not change existing item", func(t *testing.T) {
		inv := exercise.Inventory{"GO-1": {SKU: "GO-1", Qty: 7}}
		exercise.AddStock(inv, "GO-1", 0)
		exercise.AddStock(inv, "GO-1", -2)
		if inv["GO-1"].Qty != 7 {
			t.Errorf("Qty = %d, chci 7", inv["GO-1"].Qty)
		}
	})
}

func TestAddStockMutatesCallerMap(t *testing.T) {
	// Mapa je reference type — funkce mění tabulku volajícího bez pointeru.
	inv := exercise.Inventory{}
	exercise.AddStock(inv, "A", 1)
	exercise.AddStock(inv, "B", 2)
	if len(inv) != 2 {
		t.Errorf("inventář volajícího má %d položek, chci 2", len(inv))
	}
}
