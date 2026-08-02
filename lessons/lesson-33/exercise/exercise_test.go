package exercise_test

import (
	"errors"
	"fmt"
	"regexp"
	"sync"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-33/exercise"
)

// fixedClock je testovací adaptér portu Clock. Žádný mockovací framework
// není potřeba — port má jednu metodu.
type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

// seqIDs je testovací adaptér portu IDGen s předvídatelnou řadou.
type seqIDs struct{ n int }

func (g *seqIDs) NewID() string {
	g.n++
	return fmt.Sprintf("ord-%d", g.n)
}

var refTime = time.Date(2024, time.March, 7, 10, 30, 0, 0, time.UTC)

// Kontrola, že adaptéry skutečně splňují porty. Tohle je celý „implements".
var (
	_ exercise.Clock      = fixedClock{}
	_ exercise.Clock      = exercise.SystemClock{}
	_ exercise.IDGen      = (*seqIDs)(nil)
	_ exercise.IDGen      = exercise.RandomIDGen{}
	_ exercise.OrderStore = (*exercise.MemoryStore)(nil)
	_ exercise.OrderStore = exercise.FailingStore{}
)

func TestNewServiceRequiresDeps(t *testing.T) {
	tests := []struct {
		name  string
		clock exercise.Clock
		ids   exercise.IDGen
	}{
		{"missing clock", nil, &seqIDs{}},
		{"missing idgen", fixedClock{refTime}, nil},
		{"both missing", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := exercise.NewService(tt.clock, tt.ids)
			if !errors.Is(err, exercise.ErrMissingDependency) {
				t.Errorf("NewService = %v, chci ErrMissingDependency", err)
			}
		})
	}
}

func TestNewOrderIsDeterministic(t *testing.T) {
	svc, err := exercise.NewService(fixedClock{refTime}, &seqIDs{})
	if err != nil {
		t.Fatalf("NewService = %v", err)
	}

	first, err := svc.NewOrder("  Alice  ", 1999)
	if err != nil {
		t.Fatalf("NewOrder = %v", err)
	}
	want := exercise.Order{
		ID:         "ord-1",
		Customer:   "Alice",
		TotalCents: 1999,
		PlacedAt:   refTime,
	}
	if first != want {
		t.Errorf("NewOrder = %+v, chci %+v", first, want)
	}

	second, err := svc.NewOrder("Bob", 500)
	if err != nil {
		t.Fatalf("NewOrder = %v", err)
	}
	if second.ID != "ord-2" {
		t.Errorf("druhé ID = %q, chci %q", second.ID, "ord-2")
	}
	if !second.PlacedAt.Equal(refTime) {
		t.Errorf("PlacedAt = %v, chci %v — služba nesmí volat time.Now", second.PlacedAt, refTime)
	}
}

func TestNewOrderValidation(t *testing.T) {
	svc, err := exercise.NewService(fixedClock{refTime}, &seqIDs{})
	if err != nil {
		t.Fatalf("NewService = %v", err)
	}

	tests := []struct {
		name     string
		customer string
		total    int64
		want     error
	}{
		{"empty customer", "", 100, exercise.ErrEmptyCustomer},
		{"customer only spaces", "   ", 100, exercise.ErrEmptyCustomer},
		{"zero amount", "Alice", 0, exercise.ErrInvalidTotal},
		{"negative amount", "Alice", -1, exercise.ErrInvalidTotal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := svc.NewOrder(tt.customer, tt.total); !errors.Is(err, tt.want) {
				t.Errorf("NewOrder = %v, chci %v", err, tt.want)
			}
		})
	}
}

func TestMemoryStore(t *testing.T) {
	store := exercise.NewMemoryStore()

	if _, ok := store.Get("nic"); ok {
		t.Error("Get na prázdném úložišti vrátil ok=true")
	}

	o := exercise.Order{ID: "a", Customer: "Alice", TotalCents: 100, PlacedAt: refTime}
	if err := store.Save(o); err != nil {
		t.Fatalf("Save = %v, chci nil", err)
	}
	got, ok := store.Get("a")
	if !ok {
		t.Fatal("Get po Save vrátil ok=false")
	}
	if got != o {
		t.Errorf("Get = %+v, chci %+v", got, o)
	}

	updated := o
	updated.TotalCents = 250
	if err := store.Save(updated); err != nil {
		t.Fatalf("Save = %v", err)
	}
	if got, _ := store.Get("a"); got.TotalCents != 250 {
		t.Errorf("po přepsání TotalCents = %d, chci 250", got.TotalCents)
	}
}

func TestMemoryStoreConcurrent(t *testing.T) {
	store := exercise.NewMemoryStore()

	const n = 100
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("ord-%d", i)
			if err := store.Save(exercise.Order{ID: id, Customer: "X", TotalCents: 1}); err != nil {
				t.Errorf("Save = %v", err)
			}
			store.Get(id)
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		if _, ok := store.Get(fmt.Sprintf("ord-%d", i)); !ok {
			t.Fatalf("objednávka ord-%d chybí", i)
		}
	}
}

func newTestService(t *testing.T, store exercise.OrderStore) *exercise.OrderService {
	t.Helper()
	svc, err := exercise.NewOrderService(store, fixedClock{refTime}, &seqIDs{})
	if err != nil {
		t.Fatalf("NewOrderService = %v", err)
	}
	return svc
}

func TestNewOrderServiceRequiresStore(t *testing.T) {
	_, err := exercise.NewOrderService(nil, fixedClock{refTime}, &seqIDs{})
	if !errors.Is(err, exercise.ErrMissingDependency) {
		t.Errorf("NewOrderService(nil store) = %v, chci ErrMissingDependency", err)
	}
	_, err = exercise.NewOrderService(exercise.NewMemoryStore(), nil, &seqIDs{})
	if !errors.Is(err, exercise.ErrMissingDependency) {
		t.Errorf("NewOrderService(nil clock) = %v, chci ErrMissingDependency", err)
	}
}

func TestPlaceFindCancel(t *testing.T) {
	svc := newTestService(t, exercise.NewMemoryStore())

	placed, err := svc.Place("Alice", 1999)
	if err != nil {
		t.Fatalf("Place = %v", err)
	}
	if placed.ID != "ord-1" || placed.Canceled {
		t.Errorf("Place = %+v, chci ord-1 nestornovanou", placed)
	}

	found, err := svc.Find("ord-1")
	if err != nil {
		t.Fatalf("Find = %v", err)
	}
	if found != placed {
		t.Errorf("Find = %+v, chci %+v", found, placed)
	}

	canceled, err := svc.Cancel("ord-1")
	if err != nil {
		t.Fatalf("Cancel = %v", err)
	}
	if !canceled.Canceled {
		t.Error("Cancel nevrátil stornovanou objednávku")
	}

	again, err := svc.Find("ord-1")
	if err != nil {
		t.Fatalf("Find po Cancel = %v", err)
	}
	if !again.Canceled {
		t.Error("storno se neuložilo do úložiště")
	}
}

func TestFindAndCancelUnknownID(t *testing.T) {
	svc := newTestService(t, exercise.NewMemoryStore())

	if _, err := svc.Find("nic"); !errors.Is(err, exercise.ErrNotFound) {
		t.Errorf("Find(nic) = %v, chci ErrNotFound", err)
	}
	if _, err := svc.Cancel("nic"); !errors.Is(err, exercise.ErrNotFound) {
		t.Errorf("Cancel(nic) = %v, chci ErrNotFound", err)
	}
}

func TestCancelTwice(t *testing.T) {
	svc := newTestService(t, exercise.NewMemoryStore())

	if _, err := svc.Place("Alice", 100); err != nil {
		t.Fatalf("Place = %v", err)
	}
	if _, err := svc.Cancel("ord-1"); err != nil {
		t.Fatalf("první Cancel = %v", err)
	}
	if _, err := svc.Cancel("ord-1"); !errors.Is(err, exercise.ErrAlreadyCanceled) {
		t.Errorf("druhý Cancel = %v, chci ErrAlreadyCanceled", err)
	}
}

func TestSwapAdapterForFailing(t *testing.T) {
	boom := errors.New("disk je plný")
	svc := newTestService(t, exercise.FailingStore{Err: boom})

	_, err := svc.Place("Alice", 1999)
	if err == nil {
		t.Fatal("Place s padajícím adaptérem musí vrátit chybu")
	}
	if !errors.Is(err, exercise.ErrStore) {
		t.Errorf("Place = %v, chci obal s ErrStore", err)
	}
	if !errors.Is(err, boom) {
		t.Errorf("Place = %v, chci zachovanou původní chybu adaptéru", err)
	}
}

func TestCancelWrapsAdapterError(t *testing.T) {
	boom := errors.New("spojení spadlo")
	store := exercise.FailingStore{
		Err: boom,
		Orders: map[string]exercise.Order{
			"ord-1": {ID: "ord-1", Customer: "Alice", TotalCents: 100, PlacedAt: refTime},
		},
	}
	svc := newTestService(t, store)

	if _, err := svc.Find("ord-1"); err != nil {
		t.Fatalf("Find přes FailingStore = %v, chci nil", err)
	}

	_, err := svc.Cancel("ord-1")
	if !errors.Is(err, exercise.ErrStore) || !errors.Is(err, boom) {
		t.Errorf("Cancel = %v, chci obal ErrStore i původní chybu", err)
	}
}

func TestWire(t *testing.T) {
	svc, err := exercise.Wire()
	if err != nil {
		t.Fatalf("Wire = %v", err)
	}

	before := time.Now()
	o, err := svc.Place("Alice", 4200)
	if err != nil {
		t.Fatalf("Place = %v", err)
	}
	after := time.Now()

	if o.PlacedAt.Before(before) || o.PlacedAt.After(after) {
		t.Errorf("PlacedAt = %v, chci mezi %v a %v — Wire má dát systémový čas", o.PlacedAt, before, after)
	}

	re := regexp.MustCompile(`^[0-9a-f]{32}$`)
	if !re.MatchString(o.ID) {
		t.Errorf("ID = %q, chci 32 hex znaků z RandomIDGen", o.ID)
	}

	found, err := svc.Find(o.ID)
	if err != nil {
		t.Fatalf("Find = %v", err)
	}
	if found.Customer != "Alice" {
		t.Errorf("Find = %+v, chci Alice", found)
	}
}

func TestRandomIDGenIsUnique(t *testing.T) {
	var g exercise.RandomIDGen
	seen := make(map[string]bool, 500)
	for i := 0; i < 500; i++ {
		id := g.NewID()
		if seen[id] {
			t.Fatalf("duplicitní ID %q", id)
		}
		seen[id] = true
	}
}
