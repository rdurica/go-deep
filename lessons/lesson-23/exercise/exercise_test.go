package exercise_test

import (
	"errors"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-23/exercise"
)

// stubLoader je testovací dvojník portu Loader. Právě proto, že je port malý,
// se dá napsat na pět řádků a bez mockovací knihovny.
type stubLoader struct {
	record exercise.Record
	err    error
	calls  []string
}

func (s *stubLoader) Load(sku string) (exercise.Record, error) {
	s.calls = append(s.calls, sku)
	if s.err != nil {
		return exercise.Record{}, s.err
	}
	return s.record, nil
}

func TestDescribe(t *testing.T) {
	loader := &stubLoader{record: exercise.Record{SKU: "A1", Name: "Šroub", Qty: 12}}

	got, err := exercise.Describe(loader, "A1")
	if err != nil {
		t.Fatalf("Describe vrátil chybu %v, chci nil", err)
	}
	if want := "Šroub: 12 ks"; got != want {
		t.Errorf("Describe = %q, chci %q", got, want)
	}
	if len(loader.calls) != 1 || loader.calls[0] != "A1" {
		t.Errorf("Loader dostal %v, chci [A1]", loader.calls)
	}
}

func TestDescribeErrors(t *testing.T) {
	t.Run("nil loader", func(t *testing.T) {
		got, err := exercise.Describe(nil, "A1")
		if !errors.Is(err, exercise.ErrMissingLoader) {
			t.Fatalf("chyba = %v, chci ErrMissingLoader", err)
		}
		if got != "" {
			t.Errorf("při chybě chci prázdný výstup, mám %q", got)
		}
	})

	t.Run("empty SKU", func(t *testing.T) {
		loader := &stubLoader{}
		if _, err := exercise.Describe(loader, ""); !errors.Is(err, exercise.ErrEmptySKU) {
			t.Fatalf("chyba = %v, chci ErrEmptySKU", err)
		}
		if len(loader.calls) != 0 {
			t.Errorf("Loader byl volaný %v, chci žádné volání", loader.calls)
		}
	})

	t.Run("loader error wrapped with context", func(t *testing.T) {
		boom := errors.New("database down")
		_, err := exercise.Describe(&stubLoader{err: boom}, "A1")
		if !errors.Is(err, boom) {
			t.Fatalf("chyba = %v, chci obalenou %v", err, boom)
		}
		if want := `describe "A1": database down`; err.Error() != want {
			t.Errorf("err.Error() = %q, chci %q", err.Error(), want)
		}
	})
}

func TestLoadRecords(t *testing.T) {
	in := strings.Join([]string{
		"# sklad",
		"",
		"A1;Šroub;12",
		"  B2 ; Matice ; 0 ",
		"C3;Podložka;7",
	}, "\n")

	got, err := exercise.LoadRecords(strings.NewReader(in))
	if err != nil {
		t.Fatalf("LoadRecords vrátil chybu %v, chci nil", err)
	}

	want := []exercise.Record{
		{SKU: "A1", Name: "Šroub", Qty: 12},
		{SKU: "B2", Name: "Matice", Qty: 0},
		{SKU: "C3", Name: "Podložka", Qty: 7},
	}
	if len(got) != len(want) {
		t.Fatalf("LoadRecords = %+v, chci %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("záznam %d = %+v, chci %+v", i, got[i], want[i])
		}
	}
}

func TestLoadRecordsEmptyInput(t *testing.T) {
	got, err := exercise.LoadRecords(strings.NewReader("# jen komentář\n\n"))
	if err != nil {
		t.Fatalf("LoadRecords vrátil chybu %v, chci nil", err)
	}
	if len(got) != 0 {
		t.Errorf("LoadRecords = %+v, chci prázdný výsledek", got)
	}
}

func TestLoadRecordsErrors(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr error
		wantMsg string
	}{
		{
			name:    "too few columns",
			in:      "A1;Šroub;1\nB2;Matice\n",
			wantErr: exercise.ErrMalformedLine,
			wantMsg: `line 2: malformed line: "B2;Matice"`,
		},
		{
			name:    "too many columns",
			in:      "A1;Šroub;1;navíc\n",
			wantErr: exercise.ErrMalformedLine,
			wantMsg: `line 1: malformed line: "A1;Šroub;1;navíc"`,
		},
		{
			name:    "empty SKU",
			in:      "A1;Šroub;1\n ;Matice;2\n",
			wantErr: exercise.ErrEmptySKU,
			wantMsg: "line 2: empty sku",
		},
		{
			name:    "quantity is not a number",
			in:      "A1;Šroub;mnoho\n",
			wantErr: exercise.ErrInvalidQty,
			wantMsg: `line 1: invalid quantity: "mnoho"`,
		},
		{
			name:    "negative quantity",
			in:      "A1;Šroub;1\nB2;Matice;-3\n",
			wantErr: exercise.ErrInvalidQty,
			wantMsg: `line 2: invalid quantity: "-3"`,
		},
		{
			name:    "duplicate SKU",
			in:      "A1;Šroub;1\n# komentář\nA1;Šroub znovu;2\n",
			wantErr: exercise.ErrDuplicateSKU,
			wantMsg: `line 3: duplicate sku: "A1"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exercise.LoadRecords(strings.NewReader(tt.in))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("chyba = %v, chci obalenou %v", err, tt.wantErr)
			}
			if err.Error() != tt.wantMsg {
				t.Errorf("err.Error() = %q, chci %q", err.Error(), tt.wantMsg)
			}
			if got != nil {
				t.Errorf("při chybě chci nil výsledek, mám %+v", got)
			}
		})
	}
}

// failingReader vždy selže.
type failingReader struct {
	err error
}

func (f failingReader) Read([]byte) (int, error) {
	return 0, f.err
}

func TestLoadRecordsReadError(t *testing.T) {
	boom := errors.New("io timeout")

	_, err := exercise.LoadRecords(failingReader{err: boom})
	if !errors.Is(err, boom) {
		t.Fatalf("chyba = %v, chci obalenou %v", err, boom)
	}
	if want := "read records: io timeout"; err.Error() != want {
		t.Errorf("err.Error() = %q, chci %q", err.Error(), want)
	}
}

func TestStoreZeroValue(t *testing.T) {
	var s exercise.Store // žádný konstruktor

	if got := s.TotalQty(); got != 0 {
		t.Errorf("TotalQty() = %d, chci 0", got)
	}
	if got := s.List(); len(got) != 0 {
		t.Errorf("List() = %+v, chci prázdný slice", got)
	}
	if _, err := s.Load("A1"); !errors.Is(err, exercise.ErrNotFound) {
		t.Errorf("Load na prázdném skladu = %v, chci ErrNotFound", err)
	}
	if err := s.Remove("A1"); !errors.Is(err, exercise.ErrNotFound) {
		t.Errorf("Remove na prázdném skladu = %v, chci ErrNotFound", err)
	}

	if err := s.Put(exercise.Record{SKU: "A1", Name: "Šroub", Qty: 5}); err != nil {
		t.Fatalf("Put vrátil chybu %v, chci nil", err)
	}
	if got := s.TotalQty(); got != 5 {
		t.Errorf("TotalQty() = %d, chci 5", got)
	}
}

func TestStoreLifecycle(t *testing.T) {
	var s exercise.Store

	records := []exercise.Record{
		{SKU: "C3", Name: "Podložka", Qty: 7},
		{SKU: "A1", Name: "Šroub", Qty: 12},
		{SKU: "B2", Name: "Matice", Qty: 3},
	}
	if err := s.PutAll(records); err != nil {
		t.Fatalf("PutAll vrátil chybu %v, chci nil", err)
	}

	if got := s.TotalQty(); got != 22 {
		t.Errorf("TotalQty() = %d, chci 22", got)
	}

	list := s.List()
	wantOrder := []string{"A1", "B2", "C3"}
	if len(list) != len(wantOrder) {
		t.Fatalf("List() = %+v, chci 3 položky", list)
	}
	for i, sku := range wantOrder {
		if list[i].SKU != sku {
			t.Errorf("List()[%d].SKU = %q, chci %q (řadí se podle SKU)", i, list[i].SKU, sku)
		}
	}

	// Put přepisuje existující SKU.
	if err := s.Put(exercise.Record{SKU: "A1", Name: "Šroub M6", Qty: 1}); err != nil {
		t.Fatalf("Put vrátil chybu %v, chci nil", err)
	}
	if got := len(s.List()); got != 3 {
		t.Errorf("po přepisu má sklad %d položek, chci 3", got)
	}
	if got := s.TotalQty(); got != 11 {
		t.Errorf("TotalQty() = %d, chci 11", got)
	}

	rec, err := s.Load("A1")
	if err != nil {
		t.Fatalf("Load vrátil chybu %v, chci nil", err)
	}
	if rec.Name != "Šroub M6" {
		t.Errorf("Load(\"A1\").Name = %q, chci %q", rec.Name, "Šroub M6")
	}

	if err := s.Remove("B2"); err != nil {
		t.Fatalf("Remove vrátil chybu %v, chci nil", err)
	}
	if err := s.Remove("B2"); !errors.Is(err, exercise.ErrNotFound) {
		t.Errorf("druhý Remove = %v, chci ErrNotFound", err)
	}
	if got := len(s.List()); got != 2 {
		t.Errorf("po smazání má sklad %d položek, chci 2", got)
	}
}

func TestStoreValidation(t *testing.T) {
	var s exercise.Store

	if err := s.Put(exercise.Record{SKU: "", Name: "Nic", Qty: 1}); !errors.Is(err, exercise.ErrEmptySKU) {
		t.Errorf("Put bez SKU = %v, chci ErrEmptySKU", err)
	}
	if err := s.Put(exercise.Record{SKU: "A1", Name: "Šroub", Qty: -1}); !errors.Is(err, exercise.ErrInvalidQty) {
		t.Errorf("Put se záporným množstvím = %v, chci ErrInvalidQty", err)
	}
	if got := len(s.List()); got != 0 {
		t.Errorf("neplatné položky se nesmí uložit, sklad má %d položek", got)
	}
}

func TestStorePutAllJoinsErrors(t *testing.T) {
	var s exercise.Store

	err := s.PutAll([]exercise.Record{
		{SKU: "A1", Name: "Šroub", Qty: 1},
		{SKU: "", Name: "Bez SKU", Qty: 1},
		{SKU: "B2", Name: "Matice", Qty: 2},
		{SKU: "C3", Name: "Rozbitá", Qty: -5},
	})
	if err == nil {
		t.Fatal("PutAll vrátil nil, chci spojené chyby")
	}
	if !errors.Is(err, exercise.ErrEmptySKU) {
		t.Errorf("chyba %v neobsahuje ErrEmptySKU", err)
	}
	if !errors.Is(err, exercise.ErrInvalidQty) {
		t.Errorf("chyba %v neobsahuje ErrInvalidQty", err)
	}
	if !strings.Contains(err.Error(), "record 1") || !strings.Contains(err.Error(), "record 3") {
		t.Errorf("chyba %q neobsahuje indexy vadných záznamů", err.Error())
	}

	// Platné záznamy se uložily i přes chyby ostatních.
	if got := s.TotalQty(); got != 3 {
		t.Errorf("TotalQty() = %d, chci 3 — platné záznamy se ukládají", got)
	}
}

func TestStoreSatisfiesLoader(t *testing.T) {
	// Sklad je jen jedna z možných implementací portu Loader.
	var loader exercise.Loader = &exercise.Store{}

	store, ok := loader.(*exercise.Store)
	if !ok {
		t.Fatal("typová aserce na *Store selhala")
	}
	if err := store.Put(exercise.Record{SKU: "A1", Name: "Šroub", Qty: 4}); err != nil {
		t.Fatalf("Put vrátil chybu %v", err)
	}

	got, err := exercise.Describe(loader, "A1")
	if err != nil {
		t.Fatalf("Describe vrátil chybu %v, chci nil", err)
	}
	if want := "Šroub: 4 ks"; got != want {
		t.Errorf("Describe = %q, chci %q", got, want)
	}

	if _, err := exercise.Describe(loader, "NENÍ"); !errors.Is(err, exercise.ErrNotFound) {
		t.Errorf("Describe neexistujícího SKU = %v, chci obalenou ErrNotFound", err)
	}
}

func TestIntegrationLoadAndStore(t *testing.T) {
	in := "A1;Šroub;12\nB2;Matice;3\nC3;Podložka;7\n"

	records, err := exercise.LoadRecords(strings.NewReader(in))
	if err != nil {
		t.Fatalf("LoadRecords vrátil chybu %v", err)
	}

	var s exercise.Store
	if err := s.PutAll(records); err != nil {
		t.Fatalf("PutAll vrátil chybu %v", err)
	}

	desc, err := exercise.Describe(&s, "B2")
	if err != nil {
		t.Fatalf("Describe vrátil chybu %v", err)
	}
	if want := "Matice: 3 ks"; desc != want {
		t.Errorf("Describe = %q, chci %q", desc, want)
	}
	if got := s.TotalQty(); got != 22 {
		t.Errorf("TotalQty() = %d, chci 22", got)
	}
}
