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

func TestParseTransactionsEmptyArray(t *testing.T) {
	txs, err := exercise.ParseTransactions(strings.NewReader(`[]`))
	if err != nil {
		t.Fatalf("ParseTransactions([]) = _, %v, chci nil", err)
	}
	if len(txs) != 0 {
		t.Errorf("ParseTransactions([]) = %+v, chci prázdný výsledek", txs)
	}
}

func TestParseTransactionsValidation(t *testing.T) {
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
		"zero amount": {
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

func TestParseTransactionsBrokenJSON(t *testing.T) {
	for _, in := range []string{``, `[`, `{"id":"t1"}`, `[{"id":"t1","amount":"hodně"}]`} {
		if _, err := exercise.ParseTransactions(strings.NewReader(in)); err == nil {
			t.Errorf("ParseTransactions(%q) = _, nil, chci chybu", in)
		}
	}
}

func TestTotalsByCategory(t *testing.T) {
	txs := []exercise.Transaction{
		{ID: "t1", Payee: "Albert", Amount: 12050, Category: "food"},
		{ID: "t2", Payee: "DPP", Amount: 8000, Category: "transport"},
		{ID: "t3", Payee: "Rohlik", Amount: 20025, Category: "food"},
		{ID: "t4", Payee: "Kino", Amount: -3000, Category: "fun"},
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

func TestTotalsByCategoryEmpty(t *testing.T) {
	got := exercise.TotalsByCategory(nil)
	if got == nil {
		t.Fatal("TotalsByCategory(nil) = nil, chci prázdnou mapu")
	}
	if len(got) != 0 {
		t.Errorf("TotalsByCategory(nil) = %v, chci prázdnou mapu", got)
	}
}

func TestMoneyString(t *testing.T) {
	if got := exercise.Money(1999).String(); got != "19.99" {
		t.Errorf("Money(1999).String() = %q, chci %q", got, "19.99")
	}
}

func TestReportString(t *testing.T) {
	rep := exercise.Report{Count: 4, Total: 37075, Top: "food"}
	got := rep.String()
	want := "transakcí: 4, celkem: 370.75, top kategorie: food"
	if got != want {
		t.Errorf("Report.String() = %q, chci %q", got, want)
	}
}

func TestReportIsStringer(t *testing.T) {
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

func TestBuildReportCategoryMatch(t *testing.T) {
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

func TestBuildReportErrors(t *testing.T) {
	t.Run("empty book", func(t *testing.T) {
		_, err := exercise.BuildReport(strings.NewReader(`[]`))
		if !errors.Is(err, exercise.ErrEmptyLedger) {
			t.Errorf("BuildReport([]) = _, %v, chci ErrEmptyLedger", err)
		}
	})

	t.Run("invalid transaction", func(t *testing.T) {
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

	t.Run("broken input", func(t *testing.T) {
		if _, err := exercise.BuildReport(strings.NewReader(`{`)); err == nil {
			t.Error("BuildReport(rozbitý JSON) = _, nil, chci chybu")
		}
	})
}
