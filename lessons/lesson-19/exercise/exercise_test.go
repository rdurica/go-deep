package exercise_test

import (
	"errors"
	"strings"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-19/exercise"
)

func TestParseUserID(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr error
	}{
		{"42", 42, nil},
		{"  42  ", 42, nil},
		{"1", 1, nil},
		{"", 0, exercise.ErrEmptyID},
		{"   ", 0, exercise.ErrEmptyID},
		{"abc", 0, exercise.ErrInvalidID},
		{"4a2", 0, exercise.ErrInvalidID},
		{"4.2", 0, exercise.ErrInvalidID},
		{"99999999999999999999999", 0, exercise.ErrInvalidID},
		{"0", 0, exercise.ErrNonPositiveID},
		{"-5", 0, exercise.ErrNonPositiveID},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := exercise.ParseUserID(tt.in)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseUserID(%q) chyba = %v, chci %v", tt.in, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseUserID(%q) = %d, chci %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseUserIDs(t *testing.T) {
	t.Run("platný seznam", func(t *testing.T) {
		got, err := exercise.ParseUserIDs("1, 2 ,30")
		if err != nil {
			t.Fatalf("ParseUserIDs vrátil chybu %v, chci nil", err)
		}
		want := []int{1, 2, 30}
		if len(got) != len(want) {
			t.Fatalf("ParseUserIDs = %v, chci %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("ParseUserIDs = %v, chci %v", got, want)
			}
		}
	})

	t.Run("prázdný vstup", func(t *testing.T) {
		got, err := exercise.ParseUserIDs("   ")
		if err != nil {
			t.Fatalf("ParseUserIDs(\"   \") vrátil chybu %v, chci nil", err)
		}
		if len(got) != 0 {
			t.Errorf("ParseUserIDs(\"   \") = %v, chci prázdný slice", got)
		}
	})

	t.Run("chyba nese index", func(t *testing.T) {
		_, err := exercise.ParseUserIDs("1,2,")
		if !errors.Is(err, exercise.ErrEmptyID) {
			t.Fatalf("chyba = %v, chci obalenou ErrEmptyID", err)
		}
		if !strings.Contains(err.Error(), "index 2") {
			t.Errorf("chyba %q neobsahuje %q", err.Error(), "index 2")
		}
	})

	t.Run("chyba u záporného ID", func(t *testing.T) {
		_, err := exercise.ParseUserIDs("5,-1")
		if !errors.Is(err, exercise.ErrNonPositiveID) {
			t.Fatalf("chyba = %v, chci obalenou ErrNonPositiveID", err)
		}
	})
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestProcessOrders(t *testing.T) {
	orders := []exercise.Order{
		{
			ID:       "A-1",
			Customer: "zeta",
			Status:   "paid",
			Items: []exercise.Item{
				{SKU: "widget", Quantity: 2, UnitCents: 1999},
				{SKU: "gadget", Quantity: 1, UnitCents: 500},
			},
		},
		{
			ID:       "A-2",
			Customer: "acme",
			Status:   "pending",
			Items:    []exercise.Item{{SKU: "widget", Quantity: 3, UnitCents: 100}},
		},
		{
			ID:       "A-3",
			Customer: "zeta",
			Status:   "paid",
			Items:    []exercise.Item{{SKU: "bolt", Quantity: 10, UnitCents: 0}},
		},
		{
			ID:       "A-4",
			Customer: "ghost",
			Status:   "cancelled",
			Items:    []exercise.Item{{SKU: "nope", Quantity: -7, UnitCents: -1}},
		},
	}

	got, err := exercise.ProcessOrders(orders)
	if err != nil {
		t.Fatalf("ProcessOrders vrátil chybu %v, chci nil", err)
	}
	if got.OrderCount != 3 {
		t.Errorf("OrderCount = %d, chci 3", got.OrderCount)
	}
	if got.ItemCount != 16 {
		t.Errorf("ItemCount = %d, chci 16", got.ItemCount)
	}
	if got.TotalCents != 4798 {
		t.Errorf("TotalCents = %d, chci 4798", got.TotalCents)
	}
	if want := []string{"acme", "zeta"}; !sameStrings(got.Customers, want) {
		t.Errorf("Customers = %v, chci %v", got.Customers, want)
	}
}

func TestProcessOrdersEmpty(t *testing.T) {
	for name, in := range map[string][]exercise.Order{
		"nil":     nil,
		"prázdný": {},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := exercise.ProcessOrders(in)
			if err != nil {
				t.Fatalf("ProcessOrders(%v) vrátil chybu %v, chci nil", in, err)
			}
			if got.OrderCount != 0 || got.ItemCount != 0 || got.TotalCents != 0 || len(got.Customers) != 0 {
				t.Errorf("ProcessOrders(%v) = %+v, chci nulový Summary", in, got)
			}
		})
	}
}

func TestProcessOrdersErrors(t *testing.T) {
	tests := []struct {
		name    string
		orders  []exercise.Order
		wantErr error
		wantSub string
	}{
		{
			name:    "chybějící ID",
			orders:  []exercise.Order{{ID: "", Customer: "acme", Status: "paid"}},
			wantErr: exercise.ErrMissingOrderID,
			wantSub: "index 0",
		},
		{
			name: "neznámý status",
			orders: []exercise.Order{
				{ID: "A-1", Customer: "acme", Status: "paid"},
				{ID: "A-2", Customer: "acme", Status: "refunded"},
			},
			wantErr: exercise.ErrUnknownStatus,
			wantSub: "A-2",
		},
		{
			name: "nulové množství",
			orders: []exercise.Order{{
				ID: "A-9", Customer: "acme", Status: "paid",
				Items: []exercise.Item{{SKU: "widget", Quantity: 0, UnitCents: 100}},
			}},
			wantErr: exercise.ErrInvalidQuantity,
			wantSub: "widget",
		},
		{
			name: "záporná cena",
			orders: []exercise.Order{{
				ID: "A-9", Customer: "acme", Status: "pending",
				Items: []exercise.Item{{SKU: "widget", Quantity: 1, UnitCents: -1}},
			}},
			wantErr: exercise.ErrInvalidPrice,
			wantSub: "A-9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exercise.ProcessOrders(tt.orders)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("chyba = %v, chci obalenou %v", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("chyba %q neobsahuje kontext %q", err.Error(), tt.wantSub)
			}
			if got.OrderCount != 0 || got.TotalCents != 0 {
				t.Errorf("při chybě chci nulový Summary, mám %+v", got)
			}
		})
	}
}

func TestRenderInvoice(t *testing.T) {
	inv := exercise.Invoice{
		Number:   "2024-001",
		Customer: "Acme s.r.o.",
		Lines: []exercise.InvoiceLine{
			{Description: "Widget", Quantity: 2, UnitCents: 1999},
			{Description: "Gadget", Quantity: 1, UnitCents: 500},
		},
	}

	want := strings.Join([]string{
		"INVOICE 2024-001",
		"CUSTOMER: Acme s.r.o.",
		"--------------------------------",
		"Widget | 2 x 19.99 = 39.98",
		"Gadget | 1 x 5.00 = 5.00",
		"--------------------------------",
		"TOTAL: 44.98",
		"",
	}, "\n")

	if got := exercise.RenderInvoice(inv); got != want {
		t.Errorf("RenderInvoice() =\n%q\nchci\n%q", got, want)
	}
}

func TestRenderInvoiceEmpty(t *testing.T) {
	inv := exercise.Invoice{Number: "2024-002", Customer: "Nikdo"}

	want := strings.Join([]string{
		"INVOICE 2024-002",
		"CUSTOMER: Nikdo",
		"--------------------------------",
		"--------------------------------",
		"TOTAL: 0.00",
		"",
	}, "\n")

	if got := exercise.RenderInvoice(inv); got != want {
		t.Errorf("RenderInvoice() =\n%q\nchci\n%q", got, want)
	}
}

func TestRenderInvoiceRounding(t *testing.T) {
	inv := exercise.Invoice{
		Number:   "2024-003",
		Customer: "Bit s.r.o.",
		Lines: []exercise.InvoiceLine{
			{Description: "Bit", Quantity: 3, UnitCents: 1},
			{Description: "Byte", Quantity: 1, UnitCents: 100000},
		},
	}

	want := strings.Join([]string{
		"INVOICE 2024-003",
		"CUSTOMER: Bit s.r.o.",
		"--------------------------------",
		"Bit | 3 x 0.01 = 0.03",
		"Byte | 1 x 1000.00 = 1000.00",
		"--------------------------------",
		"TOTAL: 1000.03",
		"",
	}, "\n")

	if got := exercise.RenderInvoice(inv); got != want {
		t.Errorf("RenderInvoice() =\n%q\nchci\n%q", got, want)
	}
}
