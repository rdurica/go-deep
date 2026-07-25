package memstore_test

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/app"
	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/memstore"
	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/order"
)

var placedAt = time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

// Kontrola, že adaptér skutečně splňuje port. Kdyby se port změnil,
// nezkompiluje se test, ne až main.
var _ app.Repository = (*memstore.Repository)(nil)

func testOrder(t *testing.T, id string) order.Order {
	t.Helper()
	price, err := order.NewMoney(49900, "CZK")
	if err != nil {
		t.Fatalf("NewMoney = chyba %v", err)
	}
	line, err := order.NewLine("kniha-go", 2, price)
	if err != nil {
		t.Fatalf("NewLine = chyba %v", err)
	}
	o, err := order.New(id, "radek@example.com", []order.Line{line}, placedAt)
	if err != nil {
		t.Fatalf("order.New = chyba %v", err)
	}
	return o
}

func TestSaveFind(t *testing.T) {
	repo := memstore.New()
	ctx := context.Background()

	if _, err := repo.Find(ctx, "ord-1"); !errors.Is(err, order.ErrNotFound) {
		t.Errorf("Find(prázdné) = %v, chci ErrNotFound", err)
	}

	o := testOrder(t, "ord-1")
	if err := repo.Save(ctx, o); err != nil {
		t.Fatalf("Save = chyba %v", err)
	}
	got, err := repo.Find(ctx, "ord-1")
	if err != nil {
		t.Fatalf("Find = chyba %v", err)
	}
	if got.ID != o.ID || got.Status != o.Status || len(got.Lines) != len(o.Lines) {
		t.Errorf("Find = %+v, chci %+v", got, o)
	}
	if !got.PlacedAt.Equal(o.PlacedAt) {
		t.Errorf("PlacedAt = %v, chci %v", got.PlacedAt, o.PlacedAt)
	}
}

func TestSaveBezID(t *testing.T) {
	if err := memstore.New().Save(context.Background(), order.Order{}); !errors.Is(err, order.ErrMissingID) {
		t.Errorf("Save(bez ID) = %v, chci ErrMissingID", err)
	}
}

func TestSavePrepise(t *testing.T) {
	repo := memstore.New()
	ctx := context.Background()

	o := testOrder(t, "ord-1")
	if err := repo.Save(ctx, o); err != nil {
		t.Fatalf("Save = chyba %v", err)
	}
	paid, err := o.Pay()
	if err != nil {
		t.Fatalf("Pay = chyba %v", err)
	}
	if err := repo.Save(ctx, paid); err != nil {
		t.Fatalf("Save = chyba %v", err)
	}

	got, err := repo.Find(ctx, "ord-1")
	if err != nil {
		t.Fatalf("Find = chyba %v", err)
	}
	if got.Status != order.StatusPaid {
		t.Errorf("stav = %v, chci paid", got.Status)
	}
}

func TestIzolaceDat(t *testing.T) {
	repo := memstore.New()
	ctx := context.Background()
	o := testOrder(t, "ord-1")

	if err := repo.Save(ctx, o); err != nil {
		t.Fatalf("Save = chyba %v", err)
	}

	o.Lines[0].Quantity = 999 // volající si slice ponechal
	got, err := repo.Find(ctx, "ord-1")
	if err != nil {
		t.Fatalf("Find = chyba %v", err)
	}
	if got.Lines[0].Quantity == 999 {
		t.Error("Save musí položky zkopírovat, jinak fake lže proti databázi")
	}

	got.Lines[0].Quantity = 111
	again, err := repo.Find(ctx, "ord-1")
	if err != nil {
		t.Fatalf("Find = chyba %v", err)
	}
	if again.Lines[0].Quantity == 111 {
		t.Error("Find musí položky zkopírovat")
	}
}

func TestList(t *testing.T) {
	repo := memstore.New()
	ctx := context.Background()

	orders, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List = chyba %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("List(prázdné) = %d objednávek, chci 0", len(orders))
	}

	for _, id := range []string{"ord-3", "ord-1", "ord-2"} {
		if err := repo.Save(ctx, testOrder(t, id)); err != nil {
			t.Fatalf("Save = chyba %v", err)
		}
	}

	// Pořadí je součást kontraktu, jinak by test byl flaky.
	for i := 0; i < 10; i++ {
		orders, err = repo.List(ctx)
		if err != nil {
			t.Fatalf("List = chyba %v", err)
		}
		want := []string{"ord-1", "ord-2", "ord-3"}
		if len(orders) != len(want) {
			t.Fatalf("List = %d objednávek, chci %d", len(orders), len(want))
		}
		for j, id := range want {
			if orders[j].ID != id {
				t.Fatalf("List[%d].ID = %q, chci %q (seřazeno podle ID)", j, orders[j].ID, id)
			}
		}
	}
}

func TestZrusenyKontext(t *testing.T) {
	repo := memstore.New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := repo.Save(ctx, testOrder(t, "ord-1")); !errors.Is(err, context.Canceled) {
		t.Errorf("Save(zrušený ctx) = %v, chci context.Canceled", err)
	}
	if _, err := repo.Find(ctx, "ord-1"); !errors.Is(err, context.Canceled) {
		t.Errorf("Find(zrušený ctx) = %v, chci context.Canceled", err)
	}
	if _, err := repo.List(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("List(zrušený ctx) = %v, chci context.Canceled", err)
	}
}

func TestSoubezneUziti(t *testing.T) {
	repo := memstore.New()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := "ord-" + strconv.Itoa(i%4)
			price, err := order.NewMoney(100, "CZK")
			if err != nil {
				t.Errorf("NewMoney = chyba %v", err)
				return
			}
			line, err := order.NewLine("x", 1, price)
			if err != nil {
				t.Errorf("NewLine = chyba %v", err)
				return
			}
			o, err := order.New(id, "radek", []order.Line{line}, placedAt)
			if err != nil {
				t.Errorf("order.New = chyba %v", err)
				return
			}

			for j := 0; j < 50; j++ {
				if err := repo.Save(ctx, o); err != nil {
					t.Errorf("Save = chyba %v", err)
					return
				}
				if _, err := repo.Find(ctx, id); err != nil {
					t.Errorf("Find = chyba %v", err)
					return
				}
				if _, err := repo.List(ctx); err != nil {
					t.Errorf("List = chyba %v", err)
					return
				}
			}
		}(i)
	}
	wg.Wait()
}
