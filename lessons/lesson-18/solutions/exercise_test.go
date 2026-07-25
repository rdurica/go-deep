package solutions_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-18/solutions"
)

const sampleJSON = `[
	{"id":"t1","payee":"Albert","amount":12050,"category":"food"},
	{"id":"t2","payee":"DPP","amount":8000,"category":"transport"},
	{"id":"t3","payee":"Rohlik","amount":20025,"category":"food"},
	{"id":"t4","payee":"Kino","amount":-3000,"category":"fun"}
]`

func TestMoneyString(t *testing.T) {
	tests := []struct {
		in   exercise.Money
		want string
	}{
		{1999, "19.99"},
		{0, "0.00"},
		{5, "0.05"},
		{100, "1.00"},
		{-250, "-2.50"},
		{123456789, "1234567.89"},
	}
	for _, tt := range tests {
		if got := tt.in.String(); got != tt.want {
			t.Errorf("Money(%d).String() = %q, chci %q", int64(tt.in), got, tt.want)
		}
	}
}

func TestMoneyJeStringer(t *testing.T) {
	var s fmt.Stringer = exercise.Money(1999)
	if got := fmt.Sprintf("%v", s); got != "19.99" {
		t.Errorf("fmt.Sprintf(%%v, Money(1999)) = %q, chci %q", got, "19.99")
	}
}

func TestValidationErrorText(t *testing.T) {
	err := &exercise.ValidationError{Index: 2, Field: "amount", Reason: "nesmí být nula"}
	got := err.Error()
	for _, want := range []string{"2", "amount", "nesmí být nula"} {
		if !strings.Contains(got, want) {
			t.Errorf("Error() = %q, chci zmínku o %q", got, want)
		}
	}
}

func TestParseTransactions(t *testing.T) {
	txs, err := exercise.ParseTransactions(strings.NewReader(sampleJSON))
	if err != nil {
		t.Fatalf("ParseTransactions(...) = _, %v, chci nil", err)
	}
	if len(txs) != 4 {
		t.Fatalf("ParseTransactions(...) vrátilo %d transakcí, chci 4", len(txs))
	}
	want := exercise.Transaction{ID: "t1", Payee: "Albert", Amount: 12050, Category: "food"}
	if txs[0] != want {
		t.Errorf("txs[0] = %+v, chci %+v", txs[0], want)
	}
	if txs[3].Amount != exercise.Money(-3000) {
		t.Errorf("txs[3].Amount = %d, chci -3000 (záporné částky jsou platné)", int64(txs[3].Amount))
	}
}

func TestParseTransactionsPrazdnePole(t *testing.T) {
	txs, err := exercise.ParseTransactions(strings.NewReader(`[]`))
	if err != nil {
		t.Fatalf("ParseTransactions([]) = _, %v, chci nil", err)
	}
	if len(txs) != 0 {
		t.Errorf("ParseTransactions([]) = %+v, chci prázdný výsledek", txs)
	}
}

func TestParseTransactionsValidace(t *testing.T) {
	tests := map[string]struct {
		in    string
		field string
		index int
	}{
		"chybí id": {
			`[{"id":"","payee":"Albert","amount":100,"category":"food"}]`, "id", 0,
		},
		"chybí payee": {
			`[{"id":"t1","payee":"  ","amount":100,"category":"food"}]`, "payee", 0,
		},
		"chybí kategorie": {
			`[{"id":"t1","payee":"Albert","amount":100,"category":""}]`, "category", 0,
		},
		"nulová částka": {
			`[{"id":"t1","payee":"Albert","amount":100,"category":"food"},
			  {"id":"t2","payee":"DPP","amount":0,"category":"transport"}]`, "amount", 1,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := exercise.ParseTransactions(strings.NewReader(tt.in))
			if err == nil {
				t.Fatalf("ParseTransactions(%s) = _, nil, chci chybu", tt.in)
			}

			var verr *exercise.ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("chyba %v není *ValidationError (musí jít vytáhnout přes errors.As)", err)
			}
			if verr.Field != tt.field {
				t.Errorf("ValidationError.Field = %q, chci %q", verr.Field, tt.field)
			}
			if verr.Index != tt.index {
				t.Errorf("ValidationError.Index = %d, chci %d", verr.Index, tt.index)
			}
		})
	}
}

func TestParseTransactionsRozbityJSON(t *testing.T) {
	for _, in := range []string{``, `[`, `{"id":"t1"}`, `[{"id":"t1","amount":"hodně"}]`} {
		if _, err := exercise.ParseTransactions(strings.NewReader(in)); err == nil {
			t.Errorf("ParseTransactions(%q) = _, nil, chci chybu", in)
		}
	}
}

func TestTotalsByCategory(t *testing.T) {
	txs, err := exercise.ParseTransactions(strings.NewReader(sampleJSON))
	if err != nil {
		t.Fatalf("ParseTransactions(...) = _, %v, chci nil", err)
	}

	got := exercise.TotalsByCategory(txs)
	want := map[string]exercise.Money{"food": 32075, "transport": 8000, "fun": -3000}
	if len(got) != len(want) {
		t.Fatalf("TotalsByCategory(...) = %v, chci %d kategorií", got, len(want))
	}
	for cat, w := range want {
		if got[cat] != w {
			t.Errorf("TotalsByCategory(...)[%q] = %d, chci %d", cat, int64(got[cat]), int64(w))
		}
	}
}

func TestTotalsByCategoryPrazdny(t *testing.T) {
	got := exercise.TotalsByCategory(nil)
	if got == nil {
		t.Fatal("TotalsByCategory(nil) = nil, chci prázdnou mapu")
	}
	if len(got) != 0 {
		t.Errorf("TotalsByCategory(nil) = %v, chci prázdnou mapu", got)
	}
}

func TestGroupBy(t *testing.T) {
	t.Run("transakce podle kategorie", func(t *testing.T) {
		txs, err := exercise.ParseTransactions(strings.NewReader(sampleJSON))
		if err != nil {
			t.Fatalf("ParseTransactions(...) = _, %v, chci nil", err)
		}

		groups := exercise.GroupBy(txs, func(tx exercise.Transaction) string { return tx.Category })
		if len(groups) != 3 {
			t.Fatalf("GroupBy(...) = %d skupin, chci 3", len(groups))
		}
		if len(groups["food"]) != 2 {
			t.Errorf("skupina food má %d prvků, chci 2", len(groups["food"]))
		}
		if groups["food"][0].ID != "t1" || groups["food"][1].ID != "t3" {
			t.Errorf("skupina food = %+v, chci pořadí t1, t3", groups["food"])
		}
	})

	t.Run("jiný typ prvku i klíče", func(t *testing.T) {
		words := []string{"ada", "bob", "eva", "ken", "li"}
		groups := exercise.GroupBy(words, func(s string) int { return len(s) })
		if len(groups[3]) != 4 || len(groups[2]) != 1 {
			t.Errorf("GroupBy(words, len) = %v, chci 4 tříznakové a 1 dvouznakové", groups)
		}
	})

	t.Run("prázdný vstup", func(t *testing.T) {
		groups := exercise.GroupBy(nil, func(n int) int { return n })
		if groups == nil {
			t.Fatal("GroupBy(nil, ...) = nil, chci prázdnou mapu")
		}
		if len(groups) != 0 {
			t.Errorf("GroupBy(nil, ...) = %v, chci prázdnou mapu", groups)
		}
	})
}

func TestReportString(t *testing.T) {
	rep := exercise.Report{Count: 4, Total: 37075, Top: "food"}
	got := rep.String()
	want := "transakcí: 4, celkem: 370.75, top kategorie: food"
	if got != want {
		t.Errorf("Report.String() = %q, chci %q", got, want)
	}

	empty := exercise.Report{}
	wantEmpty := "transakcí: 0, celkem: 0.00, top kategorie: -"
	if got := empty.String(); got != wantEmpty {
		t.Errorf("Report{}.String() = %q, chci %q", got, wantEmpty)
	}
}

func TestReportJeStringer(t *testing.T) {
	var s fmt.Stringer = exercise.Report{Count: 1, Total: 100, Top: "food"}
	if got := fmt.Sprint(s); !strings.Contains(got, "1.00") {
		t.Errorf("fmt.Sprint(Report) = %q, chci částku 1.00", got)
	}
}

func TestBuildReport(t *testing.T) {
	got, err := exercise.BuildReport(strings.NewReader(sampleJSON))
	if err != nil {
		t.Fatalf("BuildReport(...) = _, %v, chci nil", err)
	}
	want := exercise.Report{Count: 4, Total: 37075, Top: "food"}
	if got != want {
		t.Errorf("BuildReport(...) = %+v, chci %+v", got, want)
	}
}

func TestBuildReportShodaKategorii(t *testing.T) {
	in := `[
		{"id":"t1","payee":"A","amount":5000,"category":"zzz"},
		{"id":"t2","payee":"B","amount":5000,"category":"aaa"}
	]`
	got, err := exercise.BuildReport(strings.NewReader(in))
	if err != nil {
		t.Fatalf("BuildReport(...) = _, %v, chci nil", err)
	}
	if got.Top != "aaa" {
		t.Errorf("Top = %q, chci %q (při shodě rozhoduje abeceda)", got.Top, "aaa")
	}
}

func TestBuildReportChyby(t *testing.T) {
	t.Run("prázdná kniha", func(t *testing.T) {
		_, err := exercise.BuildReport(strings.NewReader(`[]`))
		if !errors.Is(err, exercise.ErrEmptyLedger) {
			t.Errorf("BuildReport([]) = _, %v, chci ErrEmptyLedger", err)
		}
	})

	t.Run("neplatná transakce", func(t *testing.T) {
		in := `[{"id":"t1","payee":"A","amount":0,"category":"food"}]`
		_, err := exercise.BuildReport(strings.NewReader(in))
		var verr *exercise.ValidationError
		if !errors.As(err, &verr) {
			t.Fatalf("BuildReport(neplatná transakce) = _, %v, chci *ValidationError", err)
		}
		if errors.Is(err, exercise.ErrEmptyLedger) {
			t.Error("chyba validace se nesmí tvářit jako ErrEmptyLedger")
		}
	})

	t.Run("rozbitý vstup", func(t *testing.T) {
		if _, err := exercise.BuildReport(strings.NewReader(`{`)); err == nil {
			t.Error("BuildReport(rozbitý JSON) = _, nil, chci chybu")
		}
	})
}
