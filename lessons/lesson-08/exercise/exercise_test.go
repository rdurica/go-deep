package exercise_test

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-08/exercise"
)

func TestWordCount(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want map[string]int
	}{
		{"prázdný", []string{}, map[string]int{}},
		{"nil", nil, map[string]int{}},
		{"bez duplicit", []string{"a", "b"}, map[string]int{"a": 1, "b": 1}},
		{
			"s duplicitami",
			[]string{"go", "php", "go", "go", "php"},
			map[string]int{"go": 3, "php": 2},
		},
		{"prázdné slovo je klíč", []string{"", ""}, map[string]int{"": 2}},
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

func TestWordCountVracaZapisovatelnouMapu(t *testing.T) {
	// Ne-nil je požadavek proto, aby šlo do výsledku dál zapisovat.
	got := exercise.WordCount(nil)
	got["novy"] = 1
	if got["novy"] != 1 {
		t.Error("do výsledku WordCount nejde zapisovat")
	}
}

func TestNewSet(t *testing.T) {
	s := exercise.NewSet("a", "b", "a", "c", "b")
	if got := s.Len(); got != 3 {
		t.Errorf("NewSet se třemi unikátními prvky má Len %d, chci 3", got)
	}
	for _, item := range []string{"a", "b", "c"} {
		if !s.Has(item) {
			t.Errorf("Has(%q) = false, chci true", item)
		}
	}
	if s.Has("d") {
		t.Error(`Has("d") = true, chci false`)
	}
}

func TestNewSetBezArgumentu(t *testing.T) {
	s := exercise.NewSet()
	if s == nil {
		t.Fatal("NewSet() vrátil nil, chci prázdnou ne-nil množinu")
	}
	if s.Len() != 0 {
		t.Errorf("NewSet().Len() = %d, chci 0", s.Len())
	}
	s.Add("x") // na nil množině by tohle panikovalo
	if !s.Has("x") {
		t.Error("do množiny z NewSet() nejde přidávat")
	}
}

func TestSetAddJeIdempotentni(t *testing.T) {
	s := exercise.NewSet()
	s.Add("go")
	s.Add("go")
	s.Add("go")
	if got := s.Len(); got != 1 {
		t.Errorf("po třech Add stejného prvku je Len %d, chci 1", got)
	}
}

func TestSetRemove(t *testing.T) {
	s := exercise.NewSet("a", "b")
	s.Remove("a")
	if s.Has("a") {
		t.Error(`po Remove("a") je prvek pořád v množině`)
	}
	if got := s.Len(); got != 1 {
		t.Errorf("Len = %d, chci 1", got)
	}
	s.Remove("neexistuje") // nesmí panikovat
	if got := s.Len(); got != 1 {
		t.Errorf("po odebrání neexistujícího prvku je Len %d, chci 1", got)
	}
}

func TestSetNaNilMnozine(t *testing.T) {
	// Čtení z nil mapy i delete jsou legální operace.
	var s exercise.Set
	if s.Has("cokoliv") {
		t.Error("Has na nil množině má vrátit false")
	}
	if s.Len() != 0 {
		t.Errorf("Len na nil množině = %d, chci 0", s.Len())
	}
	s.Remove("cokoliv") // nesmí panikovat
	if got := s.Sorted(); len(got) != 0 {
		t.Errorf("Sorted na nil množině = %v, chci prázdný výsledek", got)
	}
}

func TestSetSorted(t *testing.T) {
	tests := []struct {
		name  string
		items []string
		want  []string
	}{
		{"prázdná", nil, []string{}},
		{"jeden prvek", []string{"b"}, []string{"b"}},
		{"seřadí", []string{"cesta", "a", "b"}, []string{"a", "b", "cesta"}},
		{"velká písmena první", []string{"b", "A", "a"}, []string{"A", "a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exercise.NewSet(tt.items...).Sorted()
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

func TestSetSortedJeDeterministicky(t *testing.T) {
	// Iterace mapy je náhodná, Sorted musí přesto dávat pořád stejný výsledek.
	s := exercise.NewSet("delta", "alfa", "charlie", "bravo", "echo")
	first := s.Sorted()
	for i := 0; i < 50; i++ {
		if got := s.Sorted(); !reflect.DeepEqual(got, first) {
			t.Fatalf("Sorted() vrátil %v, předtím %v — výstup není deterministický", got, first)
		}
	}
}

func TestSetUnion(t *testing.T) {
	a := exercise.NewSet("a", "b")
	b := exercise.NewSet("b", "c")

	got := a.Union(b).Sorted()
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Union = %v, chci %v", got, want)
	}
	if a.Len() != 2 || b.Len() != 2 {
		t.Errorf("Union změnil vstupní množiny: a.Len = %d, b.Len = %d, chci 2 a 2", a.Len(), b.Len())
	}
}

func TestSetUnionSPrazdnou(t *testing.T) {
	a := exercise.NewSet("x")
	got := a.Union(exercise.NewSet()).Sorted()
	if !reflect.DeepEqual(got, []string{"x"}) {
		t.Errorf("Union s prázdnou množinou = %v, chci [x]", got)
	}
}

func TestSetIntersect(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
		want []string
	}{
		{"překryv", []string{"a", "b", "c"}, []string{"b", "c", "d"}, []string{"b", "c"}},
		{"bez průniku", []string{"a"}, []string{"b"}, []string{}},
		{"s prázdnou", []string{"a", "b"}, nil, []string{}},
		{"stejné", []string{"a", "b"}, []string{"b", "a"}, []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := exercise.NewSet(tt.a...)
			b := exercise.NewSet(tt.b...)

			got := a.Intersect(b).Sorted()
			if len(got) != len(tt.want) {
				t.Fatalf("Intersect = %v, chci %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("Intersect = %v, chci %v", got, tt.want)
				}
			}
			if a.Len() != len(tt.a) || b.Len() != len(tt.b) {
				t.Error("Intersect změnil vstupní množiny")
			}
		})
	}
}

func TestSetUnionVraciNezavislouMnozinu(t *testing.T) {
	a := exercise.NewSet("a")
	b := exercise.NewSet("b")

	u := a.Union(b)
	u.Add("navic")

	if a.Has("navic") || b.Has("navic") {
		t.Error("výsledek Union sdílí mapu se vstupem")
	}
}

func TestAddStockZakladaChybejiciPolozku(t *testing.T) {
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

func TestAddStockPricitaPresPointer(t *testing.T) {
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

func TestAddStockHraniceniPripady(t *testing.T) {
	t.Run("nil inventář nepanikuje", func(t *testing.T) {
		var inv exercise.Inventory
		exercise.AddStock(inv, "GO-1", 5)
		if len(inv) != 0 {
			t.Errorf("nil inventář má %d položek, chci 0", len(inv))
		}
	})

	t.Run("nekladné n nezakládá položku", func(t *testing.T) {
		inv := exercise.Inventory{}
		exercise.AddStock(inv, "GO-1", 0)
		exercise.AddStock(inv, "GO-2", -3)
		if len(inv) != 0 {
			t.Errorf("inventář má %d položek, chci 0", len(inv))
		}
	})

	t.Run("nekladné n nemění existující položku", func(t *testing.T) {
		inv := exercise.Inventory{"GO-1": {SKU: "GO-1", Qty: 7}}
		exercise.AddStock(inv, "GO-1", 0)
		exercise.AddStock(inv, "GO-1", -2)
		if inv["GO-1"].Qty != 7 {
			t.Errorf("Qty = %d, chci 7", inv["GO-1"].Qty)
		}
	})
}

func TestAddStockMutujeMapuVolajiciho(t *testing.T) {
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

func TestGroupByZachovavaPoradiVeSkupine(t *testing.T) {
	words := []string{"c", "b", "a", "cc", "bb", "aa"}
	got := exercise.GroupBy(words, func(s string) string {
		return strings.Repeat("x", len(s))
	})
	want := []string{"c", "b", "a"}
	if !reflect.DeepEqual(got["x"], want) {
		t.Errorf("skupina \"x\" = %v, chci %v (pořadí vstupu)", got["x"], want)
	}
}

func TestGroupByPrazdnyVstup(t *testing.T) {
	got := exercise.GroupBy(nil, strings.ToUpper)
	if got == nil {
		t.Fatal("GroupBy(nil) vrátil nil mapu, chci ne-nil")
	}
	if len(got) != 0 {
		t.Errorf("GroupBy(nil) = %v, chci prázdnou mapu", got)
	}
}
