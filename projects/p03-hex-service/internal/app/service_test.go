package app_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/app"
	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/order"
)

var fixedTime = time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

// fakeRepo je ručně psaný fake portu Repository. Počítá volání, takže
// jde ověřit i to, co se NEstalo — třeba že se neplatná objednávka
// neuložila.
type fakeRepo struct {
	mu      sync.Mutex
	orders  map[string]order.Order
	saves   int
	finds   int
	saveErr error
	findErr error
	listErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{orders: make(map[string]order.Order)}
}

func (f *fakeRepo) Save(ctx context.Context, o order.Order) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.saves++
	if f.saveErr != nil {
		return f.saveErr
	}
	f.orders[o.ID] = o
	return nil
}

func (f *fakeRepo) Find(ctx context.Context, id string) (order.Order, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.finds++
	if f.findErr != nil {
		return order.Order{}, f.findErr
	}
	o, ok := f.orders[id]
	if !ok {
		return order.Order{}, fmt.Errorf("%w: %s", order.ErrNotFound, id)
	}
	return o, nil
}

func (f *fakeRepo) List(ctx context.Context) ([]order.Order, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := make([]order.Order, 0, len(f.orders))
	for _, o := range f.orders {
		out = append(out, o)
	}
	return out, nil
}

// stubIDs vydává předvídatelná ID, aby testy nezávisely na náhodě.
type stubIDs struct {
	mu sync.Mutex
	n  int
}

func (s *stubIDs) NewID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.n++
	return "ord-" + strconv.Itoa(s.n)
}

func newService(repo app.Repository) *app.Service {
	return app.NewService(repo, app.ClockFunc(func() time.Time { return fixedTime }), &stubIDs{})
}

func validCommand() app.PlaceCommand {
	return app.PlaceCommand{
		Customer: "radek@example.com",
		Currency: "CZK",
		Lines: []app.LineCommand{
			{SKU: "kniha-go", Quantity: 2, UnitPriceCents: 49900},
			{SKU: "hrnek", Quantity: 1, UnitPriceCents: 19900},
		},
	}
}

func TestPlace(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(repo)

	o, err := svc.Place(context.Background(), validCommand())
	if err != nil {
		t.Fatalf("Place = chyba %v", err)
	}
	if o.ID != "ord-1" {
		t.Errorf("ID = %q, chci ord-1 (z portu IDGen)", o.ID)
	}
	if !o.PlacedAt.Equal(fixedTime) {
		t.Errorf("PlacedAt = %v, chci %v (z portu Clock)", o.PlacedAt, fixedTime)
	}
	if o.Status != order.StatusNew {
		t.Errorf("Status = %v, chci new", o.Status)
	}
	total, err := o.Total()
	if err != nil {
		t.Fatalf("Total = chyba %v", err)
	}
	if total.Cents() != 2*49900+19900 {
		t.Errorf("Total = %d, chci %d", total.Cents(), 2*49900+19900)
	}
	if repo.saves != 1 {
		t.Errorf("Save volán %dkrát, chci 1", repo.saves)
	}
}

func TestPlaceNeplatneVstupy(t *testing.T) {
	tests := []struct {
		name    string
		cmd     app.PlaceCommand
		wantErr error
	}{
		{"bez zákazníka", app.PlaceCommand{Currency: "CZK", Lines: validCommand().Lines}, order.ErrMissingCustomer},
		{"bez položek", app.PlaceCommand{Customer: "radek", Currency: "CZK"}, order.ErrEmptyOrder},
		{"neplatná měna", app.PlaceCommand{Customer: "radek", Currency: "koruna", Lines: validCommand().Lines}, order.ErrInvalidCurrency},
		{"nulové množství", app.PlaceCommand{Customer: "radek", Currency: "CZK",
			Lines: []app.LineCommand{{SKU: "x", Quantity: 0, UnitPriceCents: 100}}}, order.ErrInvalidLine},
		{"nulová cena", app.PlaceCommand{Customer: "radek", Currency: "CZK",
			Lines: []app.LineCommand{{SKU: "x", Quantity: 1, UnitPriceCents: 0}}}, order.ErrInvalidLine},
		{"záporná cena", app.PlaceCommand{Customer: "radek", Currency: "CZK",
			Lines: []app.LineCommand{{SKU: "x", Quantity: 1, UnitPriceCents: -100}}}, order.ErrNegativeAmount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			svc := newService(repo)

			if _, err := svc.Place(context.Background(), tt.cmd); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Place = %v, chci %v", err, tt.wantErr)
			}
			if repo.saves != 0 {
				t.Errorf("Save volán %dkrát, chci 0 — neplatná objednávka se nesmí uložit", repo.saves)
			}
		})
	}
}

func TestPlaceObaliChybuPortu(t *testing.T) {
	sentinel := errors.New("disk plný")
	repo := newFakeRepo()
	repo.saveErr = sentinel

	if _, err := newService(repo).Place(context.Background(), validCommand()); !errors.Is(err, sentinel) {
		t.Errorf("Place = %v, chci chybu obalující chybu portu (%%w)", err)
	}
}

func TestZivotniCyklus(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(repo)
	ctx := context.Background()

	o, err := svc.Place(ctx, validCommand())
	if err != nil {
		t.Fatalf("Place = chyba %v", err)
	}

	paid, err := svc.Pay(ctx, o.ID)
	if err != nil {
		t.Fatalf("Pay = chyba %v", err)
	}
	if paid.Status != order.StatusPaid {
		t.Fatalf("po Pay stav %v, chci paid", paid.Status)
	}

	shipped, err := svc.Ship(ctx, o.ID)
	if err != nil {
		t.Fatalf("Ship = chyba %v", err)
	}
	if shipped.Status != order.StatusShipped {
		t.Fatalf("po Ship stav %v, chci shipped", shipped.Status)
	}

	reloaded, err := svc.Get(ctx, o.ID)
	if err != nil {
		t.Fatalf("Get = chyba %v", err)
	}
	if reloaded.Status != order.StatusShipped {
		t.Errorf("uložený stav %v, chci shipped — use-case musí ukládat", reloaded.Status)
	}
}

func TestZamitnutyPrechodNeuklada(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(repo)
	ctx := context.Background()

	o, err := svc.Place(ctx, validCommand())
	if err != nil {
		t.Fatalf("Place = chyba %v", err)
	}
	before := repo.saves

	if _, err := svc.Ship(ctx, o.ID); !errors.Is(err, order.ErrInvalidTransition) {
		t.Fatalf("Ship nezaplacené = %v, chci ErrInvalidTransition", err)
	}
	if repo.saves != before {
		t.Errorf("Save volán %dkrát navíc, chci 0", repo.saves-before)
	}
}

func TestNaNeexistujici(t *testing.T) {
	svc := newService(newFakeRepo())
	ctx := context.Background()

	for name, fn := range map[string]func(context.Context, string) (order.Order, error){
		"Get":    svc.Get,
		"Pay":    svc.Pay,
		"Ship":   svc.Ship,
		"Cancel": svc.Cancel,
	} {
		if _, err := fn(ctx, "neexistuje"); !errors.Is(err, order.ErrNotFound) {
			t.Errorf("%s = %v, chci ErrNotFound", name, err)
		}
	}
}

func TestListObaliChybuPortu(t *testing.T) {
	sentinel := errors.New("úložiště neodpovídá")
	repo := newFakeRepo()
	repo.listErr = sentinel

	if _, err := newService(repo).List(context.Background()); !errors.Is(err, sentinel) {
		t.Errorf("List = %v, chci chybu obalující chybu portu", err)
	}
}

func TestCancel(t *testing.T) {
	svc := newService(newFakeRepo())
	ctx := context.Background()

	o, err := svc.Place(ctx, validCommand())
	if err != nil {
		t.Fatalf("Place = chyba %v", err)
	}
	cancelled, err := svc.Cancel(ctx, o.ID)
	if err != nil {
		t.Fatalf("Cancel = chyba %v", err)
	}
	if cancelled.Status != order.StatusCancelled {
		t.Errorf("stav = %v, chci cancelled", cancelled.Status)
	}
	if _, err := svc.Pay(ctx, o.ID); !errors.Is(err, order.ErrInvalidTransition) {
		t.Errorf("Pay zrušené = %v, chci ErrInvalidTransition", err)
	}
}
