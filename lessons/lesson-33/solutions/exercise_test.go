package solutions_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-33/solutions"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type seqIDs struct{ n int }

func (g *seqIDs) NewID() string {
	g.n++
	return fmt.Sprintf("ord-%d", g.n)
}

var refTime = time.Date(2024, time.March, 7, 10, 30, 0, 0, time.UTC)

var (
	_ exercise.Clock      = fixedClock{}
	_ exercise.OrderStore = (*exercise.MemoryStore)(nil)
)

func mustService(t *testing.T) *exercise.Service {
	t.Helper()
	svc, err := exercise.NewService(fixedClock{refTime}, &seqIDs{})
	if err != nil {
		t.Fatalf("NewService = %v", err)
	}
	return svc
}

func TestNewOrderUsesClockPort(t *testing.T) {
	svc := mustService(t)
	o, err := svc.NewOrder("Alice", 1999)
	if err != nil {
		t.Fatalf("NewOrder = %v", err)
	}
	if o.PlacedAt != refTime {
		t.Errorf("PlacedAt = %v, chci %v — nesmí volat time.Now()", o.PlacedAt, refTime)
	}
	if o.ID != "ord-1" || o.Customer != "Alice" {
		t.Errorf("NewOrder = %+v, chci ord-1/Alice", o)
	}
}

func TestNewOrderValidation(t *testing.T) {
	svc := mustService(t)
	if _, err := svc.NewOrder("", 100); !errors.Is(err, exercise.ErrEmptyCustomer) {
		t.Errorf("NewOrder = %v, chci ErrEmptyCustomer", err)
	}
	if _, err := svc.NewOrder("Alice", 0); !errors.Is(err, exercise.ErrInvalidTotal) {
		t.Errorf("NewOrder = %v, chci ErrInvalidTotal", err)
	}
}

func TestMemoryStore(t *testing.T) {
	store := exercise.NewMemoryStore()
	o := exercise.Order{ID: "a", Customer: "Alice", TotalCents: 100, PlacedAt: refTime}
	if err := store.Save(o); err != nil {
		t.Fatalf("Save = %v", err)
	}
	got, ok := store.Get("a")
	if !ok || got != o {
		t.Errorf("Get = %+v, ok=%v, chci %+v", got, ok, o)
	}
	if _, ok := store.Get("nic"); ok {
		t.Error("Get na prázdném úložišti má vrátit false")
	}
}

func TestMemoryStoreConcurrent(t *testing.T) {
	store := exercise.NewMemoryStore()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("ord-%d", i)
			_ = store.Save(exercise.Order{ID: id, Customer: "X", TotalCents: 1})
			store.Get(id)
		}(i)
	}
	wg.Wait()
}

func TestPlace(t *testing.T) {
	store := exercise.NewMemoryStore()
	svc, err := exercise.NewOrderService(store, mustService(t))
	if err != nil {
		t.Fatalf("NewOrderService = %v", err)
	}
	o, err := svc.Place("Alice", 1999)
	if err != nil {
		t.Fatalf("Place = %v", err)
	}
	got, ok := store.Get(o.ID)
	if !ok || got.Customer != "Alice" {
		t.Errorf("uložená objednávka = %+v, ok=%v", got, ok)
	}
}

func TestPlaceDoesNotSaveInvalidOrder(t *testing.T) {
	store := exercise.NewMemoryStore()
	svc, _ := exercise.NewOrderService(store, mustService(t))
	if _, err := svc.Place("", 100); !errors.Is(err, exercise.ErrEmptyCustomer) {
		t.Fatalf("Place = %v, chci ErrEmptyCustomer", err)
	}
	if _, ok := store.Get("ord-1"); ok {
		t.Error("neplatná objednávka se nesmí uložit")
	}
}

func TestPlaceWrapsStoreError(t *testing.T) {
	boom := errors.New("disk plný")
	svc, _ := exercise.NewOrderService(exercise.FailingStore{Err: boom}, mustService(t))
	_, err := svc.Place("Alice", 100)
	if !errors.Is(err, exercise.ErrStore) || !errors.Is(err, boom) {
		t.Errorf("Place = %v, chci obal ErrStore i původní chybu", err)
	}
}

func TestWire(t *testing.T) {
	svc, err := exercise.Wire()
	if err != nil {
		t.Fatalf("Wire = %v", err)
	}
	before := time.Now()
	o, err := svc.Place("Bob", 500)
	after := time.Now()
	if err != nil {
		t.Fatalf("Place = %v", err)
	}
	if o.PlacedAt.Before(before) || o.PlacedAt.After(after) {
		t.Error("Wire má použít SystemClock")
	}
	if len(o.ID) != 32 {
		t.Errorf("ID = %q, chci 32 hex znaků", o.ID)
	}
}
