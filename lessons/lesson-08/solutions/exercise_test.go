package solutions_test

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-08/solutions"
)

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

func TestSetLen(t *testing.T) {
	s := exercise.Set{"a": {}, "b": {}, "c": {}}
	if got := s.Len(); got != 3 {
		t.Errorf("Len = %d, chci 3", got)
	}

	var nilSet exercise.Set
	if got := nilSet.Len(); got != 0 {
		t.Errorf("Len na nil množině = %d, chci 0", got)
	}
}

func TestSetRemove(t *testing.T) {
	s := exercise.Set{"a": {}, "b": {}}
	s.Remove("a")
	if _, ok := s["a"]; ok {
		t.Error(`po Remove("a") je prvek pořád v množině`)
	}
	if len(s) != 1 {
		t.Errorf("len = %d, chci 1", len(s))
	}
	s.Remove("neexistuje") // nesmí panikovat
	if len(s) != 1 {
		t.Errorf("po odebrání neexistujícího prvku je len %d, chci 1", len(s))
	}

	var nilSet exercise.Set
	nilSet.Remove("cokoliv") // nesmí panikovat
}

func TestSetSorted(t *testing.T) {
	tests := []struct {
		name string
		s    exercise.Set
		want []string
	}{
		{"empty", nil, []string{}},
		{"one element", exercise.Set{"b": {}}, []string{"b"}},
		{"sorts", exercise.Set{"cesta": {}, "a": {}, "b": {}}, []string{"a", "b", "cesta"}},
		{"uppercase first", exercise.Set{"b": {}, "A": {}, "a": {}}, []string{"A", "a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Sorted()
			if len(got) != len(tt.want) {
				t.Fatalf("Sorted() = %v, chci %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("Sorted() = %v, chci %v", got, tt.want)
				}
			}
		})
	}
}

func TestSetSortedIsDeterministic(t *testing.T) {
	// Iterace mapy je náhodná, Sorted musí přesto dávat pořád stejný výsledek.
	s := exercise.Set{
		"delta": {}, "alfa": {}, "charlie": {}, "bravo": {}, "echo": {},
	}
	first := s.Sorted()
	for i := 0; i < 50; i++ {
		if got := s.Sorted(); !reflect.DeepEqual(got, first) {
			t.Fatalf("Sorted() vrátil %v, předtím %v — výstup není deterministický", got, first)
		}
	}
}

func setKeys(s exercise.Set) []string {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func TestSetUnion(t *testing.T) {
	a := exercise.Set{"a": {}, "b": {}}
	b := exercise.Set{"b": {}, "c": {}}

	got := a.Union(b)
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(setKeys(got), want) {
		t.Errorf("Union = %v, chci %v", setKeys(got), want)
	}
	if len(a) != 2 || len(b) != 2 {
		t.Errorf("Union změnil vstupní množiny: len(a) = %d, len(b) = %d, chci 2 a 2", len(a), len(b))
	}
}

func TestSetUnionWithEmpty(t *testing.T) {
	a := exercise.Set{"x": {}}
	got := a.Union(exercise.Set{})
	if !reflect.DeepEqual(setKeys(got), []string{"x"}) {
		t.Errorf("Union s prázdnou množinou = %v, chci [x]", setKeys(got))
	}
}

func TestSetIntersect(t *testing.T) {
	tests := []struct {
		name string
		a, b exercise.Set
		want []string
	}{
		{"overlap", exercise.Set{"a": {}, "b": {}, "c": {}}, exercise.Set{"b": {}, "c": {}, "d": {}}, []string{"b", "c"}},
		{"no intersection", exercise.Set{"a": {}}, exercise.Set{"b": {}}, []string{}},
		{"with empty", exercise.Set{"a": {}, "b": {}}, nil, []string{}},
		{"same", exercise.Set{"a": {}, "b": {}}, exercise.Set{"b": {}, "a": {}}, []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aLen, bLen := len(tt.a), len(tt.b)
			got := tt.a.Intersect(tt.b)
			if !reflect.DeepEqual(setKeys(got), tt.want) {
				t.Fatalf("Intersect = %v, chci %v", setKeys(got), tt.want)
			}
			if len(tt.a) != aLen || len(tt.b) != bLen {
				t.Error("Intersect změnil vstupní množiny")
			}
		})
	}
}

func TestSetUnionReturnsIndependentSet(t *testing.T) {
	a := exercise.Set{"a": {}}
	b := exercise.Set{"b": {}}

	u := a.Union(b)
	u["navic"] = struct{}{}

	if _, ok := a["navic"]; ok {
		t.Error("výsledek Union sdílí mapu se vstupem a")
	}
	if _, ok := b["navic"]; ok {
		t.Error("výsledek Union sdílí mapu se vstupem b")
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

func TestGroupBy(t *testing.T) {
	firstLetter := func(s string) string {
		if s == "" {
			return ""
		}
		return s[:1]
	}

	words := []string{"gopher", "php", "go", "python", "perl", "gin"}
	got := exercise.GroupBy(words, firstLetter)

	want := map[string][]string{
		"g": {"gopher", "go", "gin"},
		"p": {"php", "python", "perl"},
	}

	// Klíče mapy porovnáváme seřazené, aby byl test deterministický.
	gotKeys := make([]string, 0, len(got))
	for k := range got {
		gotKeys = append(gotKeys, k)
	}
	sort.Strings(gotKeys)

	wantKeys := []string{"g", "p"}
	if !reflect.DeepEqual(gotKeys, wantKeys) {
		t.Fatalf("klíče = %v, chci %v", gotKeys, wantKeys)
	}
	for _, k := range wantKeys {
		if !reflect.DeepEqual(got[k], want[k]) {
			t.Errorf("skupina %q = %v, chci %v", k, got[k], want[k])
		}
	}
}

func TestGroupByPreservesOrderInGroup(t *testing.T) {
	words := []string{"c", "b", "a", "cc", "bb", "aa"}
	got := exercise.GroupBy(words, func(s string) string {
		return strings.Repeat("x", len(s))
	})
	want := []string{"c", "b", "a"}
	if !reflect.DeepEqual(got["x"], want) {
		t.Errorf("skupina \"x\" = %v, chci %v (pořadí vstupu)", got["x"], want)
	}
}

func TestGroupByEmptyInput(t *testing.T) {
	got := exercise.GroupBy(nil, strings.ToUpper)
	if got == nil {
		t.Fatal("GroupBy(nil) vrátil nil mapu, chci ne-nil")
	}
	if len(got) != 0 {
		t.Errorf("GroupBy(nil) = %v, chci prázdnou mapu", got)
	}
}
