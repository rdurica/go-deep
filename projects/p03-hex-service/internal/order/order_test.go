package order_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/order"
)

var placedAt = time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

func testLines(t *testing.T) []order.Line {
	t.Helper()
	book, err := order.NewLine("kniha-go", 2, mustMoney(t, 49900, "CZK"))
	if err != nil {
		t.Fatalf("NewLine = chyba %v", err)
	}
	mug, err := order.NewLine("hrnek", 1, mustMoney(t, 19900, "CZK"))
	if err != nil {
		t.Fatalf("NewLine = chyba %v", err)
	}
	return []order.Line{book, mug}
}

func testOrder(t *testing.T) order.Order {
	t.Helper()
	o, err := order.New("ord-1", "radek@example.com", testLines(t), placedAt)
	if err != nil {
		t.Fatalf("order.New = chyba %v", err)
	}
	return o
}

func TestStatusString(t *testing.T) {
	tests := map[order.Status]string{
		order.StatusUnknown:   "unknown",
		order.StatusNew:       "new",
		order.StatusPaid:      "paid",
		order.StatusShipped:   "shipped",
		order.StatusCancelled: "cancelled",
		order.Status(99):      "unknown",
	}
	for status, want := range tests {
		if got := status.String(); got != want {
			t.Errorf("Status(%d).String() = %q, chci %q", int(status), got, want)
		}
	}
	if got := fmt.Sprintf("%v", order.StatusShipped); got != "shipped" {
		t.Errorf("fmt.Sprintf(%%v) = %q, chci shipped", got)
	}
}

func TestNewLine(t *testing.T) {
	line, err := order.NewLine("  kniha-go ", 3, mustMoney(t, 1000, "CZK"))
	if err != nil {
		t.Fatalf("NewLine = chyba %v", err)
	}
	if line.SKU != "kniha-go" {
		t.Errorf("SKU = %q, chci ořezané", line.SKU)
	}
	total, err := line.Total()
	if err != nil {
		t.Fatalf("Total = chyba %v", err)
	}
	if total.Cents() != 3000 {
		t.Errorf("Total() = %d, chci 3000", total.Cents())
	}
}

func TestNewLineInvarianty(t *testing.T) {
	price := mustMoney(t, 1000, "CZK")
	tests := []struct {
		name     string
		sku      string
		quantity int
		price    order.Money
	}{
		{"prázdné SKU", "  ", 1, price},
		{"nulové množství", "x", 0, price},
		{"záporné množství", "x", -2, price},
		{"množství nad limitem", "x", order.MaxQuantity + 1, price},
		{"nulová cena", "x", 1, order.Money{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := order.NewLine(tt.sku, tt.quantity, tt.price); !errors.Is(err, order.ErrInvalidLine) {
				t.Errorf("NewLine = %v, chci ErrInvalidLine", err)
			}
		})
	}
}

func TestNewOrder(t *testing.T) {
	o := testOrder(t)
	if o.Status != order.StatusNew {
		t.Errorf("Status = %v, chci new", o.Status)
	}
	if o.PlacedAt.Location() != time.UTC {
		t.Errorf("PlacedAt má být v UTC, mám %v", o.PlacedAt.Location())
	}
	total, err := o.Total()
	if err != nil {
		t.Fatalf("Total = chyba %v", err)
	}
	if total.Cents() != 2*49900+19900 {
		t.Errorf("Total() = %d, chci %d", total.Cents(), 2*49900+19900)
	}
	if total.String() != "1197.00 CZK" {
		t.Errorf("Total().String() = %q, chci %q", total.String(), "1197.00 CZK")
	}
}

func TestNewOrderInvarianty(t *testing.T) {
	lines := testLines(t)
	tests := []struct {
		name     string
		id       string
		customer string
		lines    []order.Line
		at       time.Time
		wantErr  error
	}{
		{"prázdné ID", " ", "radek", lines, placedAt, order.ErrMissingID},
		{"prázdný zákazník", "ord-1", "  ", lines, placedAt, order.ErrMissingCustomer},
		{"nulový čas", "ord-1", "radek", lines, time.Time{}, order.ErrMissingTimestamp},
		{"nil položky", "ord-1", "radek", nil, placedAt, order.ErrEmptyOrder},
		{"prázdné položky", "ord-1", "radek", []order.Line{}, placedAt, order.ErrEmptyOrder},
		{"obejitý konstruktor položky", "ord-1", "radek",
			[]order.Line{{SKU: "x", Quantity: 0, UnitPrice: mustMoney(t, 100, "CZK")}},
			placedAt, order.ErrInvalidLine},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := order.New(tt.id, tt.customer, tt.lines, tt.at); !errors.Is(err, tt.wantErr) {
				t.Errorf("order.New = %v, chci %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewOrderOdmitneMichaniMen(t *testing.T) {
	czk, err := order.NewLine("kniha", 1, mustMoney(t, 100, "CZK"))
	if err != nil {
		t.Fatalf("NewLine = chyba %v", err)
	}
	eur, err := order.NewLine("hrnek", 1, mustMoney(t, 100, "EUR"))
	if err != nil {
		t.Fatalf("NewLine = chyba %v", err)
	}
	if _, err := order.New("ord-1", "radek", []order.Line{czk, eur}, placedAt); !errors.Is(err, order.ErrCurrencyMismatch) {
		t.Errorf("order.New s dvěma měnami = %v, chci ErrCurrencyMismatch", err)
	}
}

func TestNewOrderKopirujePolozky(t *testing.T) {
	lines := testLines(t)
	o, err := order.New("ord-1", "radek", lines, placedAt)
	if err != nil {
		t.Fatalf("order.New = chyba %v", err)
	}
	before, err := o.Total()
	if err != nil {
		t.Fatalf("Total = chyba %v", err)
	}

	lines[0].Quantity = 999

	after, err := o.Total()
	if err != nil {
		t.Fatalf("Total = chyba %v", err)
	}
	if after != before {
		t.Errorf("změna vstupního slice prosákla dovnitř: %v → %v", before, after)
	}
}

func TestPrechodyStavu(t *testing.T) {
	steps := map[string]func(order.Order) (order.Order, error){
		"Pay":    order.Order.Pay,
		"Ship":   order.Order.Ship,
		"Cancel": order.Order.Cancel,
	}
	allowed := map[order.Status]map[string]order.Status{
		order.StatusNew:       {"Pay": order.StatusPaid, "Cancel": order.StatusCancelled},
		order.StatusPaid:      {"Ship": order.StatusShipped, "Cancel": order.StatusCancelled},
		order.StatusShipped:   {},
		order.StatusCancelled: {},
	}

	for status, ok := range allowed {
		for name, step := range steps {
			t.Run(status.String()+"/"+name, func(t *testing.T) {
				base := testOrder(t)
				base.Status = status

				next, err := step(base)
				want, isAllowed := ok[name]
				if !isAllowed {
					if !errors.Is(err, order.ErrInvalidTransition) {
						t.Fatalf("%s ze stavu %v = %v, chci ErrInvalidTransition", name, status, err)
					}
					return
				}
				if err != nil {
					t.Fatalf("%s ze stavu %v = chyba %v", name, status, err)
				}
				if next.Status != want {
					t.Errorf("%s ze stavu %v dal %v, chci %v", name, status, next.Status, want)
				}
			})
		}
	}
}

func TestPrechodNemeniPuvodni(t *testing.T) {
	o := testOrder(t)
	if _, err := o.Pay(); err != nil {
		t.Fatalf("Pay = chyba %v", err)
	}
	if o.Status != order.StatusNew {
		t.Errorf("původní objednávka má stav %v, chci new", o.Status)
	}
}
