package order_test

import (
	"errors"
	"math"
	"testing"

	"github.com/rdurica/go-deep/projects/p03-hex-service/internal/order"
)

func mustMoney(t *testing.T, cents int64, currency string) order.Money {
	t.Helper()
	m, err := order.NewMoney(cents, currency)
	if err != nil {
		t.Fatalf("NewMoney(%d, %q) = chyba %v", cents, currency, err)
	}
	return m
}

func TestNewMoney(t *testing.T) {
	m := mustMoney(t, 19900, "czk")
	if m.Cents() != 19900 {
		t.Errorf("Cents() = %d, chci 19900", m.Cents())
	}
	if m.Currency() != "CZK" {
		t.Errorf("Currency() = %q, chci CZK (měna se normalizuje)", m.Currency())
	}
	if m.IsZero() {
		t.Error("nenulová částka nemá být IsZero()")
	}
}

func TestNewMoneyChyby(t *testing.T) {
	tests := []struct {
		name     string
		cents    int64
		currency string
		wantErr  error
	}{
		{"prázdná měna", 100, "", order.ErrInvalidCurrency},
		{"krátká měna", 100, "CZ", order.ErrInvalidCurrency},
		{"dlouhá měna", 100, "CZKK", order.ErrInvalidCurrency},
		{"číslice v měně", 100, "C1K", order.ErrInvalidCurrency},
		{"záporná částka", -1, "CZK", order.ErrNegativeAmount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := order.NewMoney(tt.cents, tt.currency); !errors.Is(err, tt.wantErr) {
				t.Errorf("NewMoney(%d, %q) = %v, chci %v", tt.cents, tt.currency, err, tt.wantErr)
			}
		})
	}
}

func TestMoneyNulovaHodnota(t *testing.T) {
	var zero order.Money
	if !zero.IsZero() {
		t.Error("nulová hodnota má být IsZero()")
	}
	if zero.String() != "0.00" {
		t.Errorf("String() = %q, chci %q", zero.String(), "0.00")
	}

	// Nulová hodnota je neutrální prvek, aby šlo sčítat od `var total Money`.
	sum, err := zero.Add(mustMoney(t, 500, "EUR"))
	if err != nil {
		t.Fatalf("Add = chyba %v", err)
	}
	if sum.Cents() != 500 || sum.Currency() != "EUR" {
		t.Errorf("Add = %v, chci 5.00 EUR", sum)
	}
}

func TestMoneyAdd(t *testing.T) {
	sum, err := mustMoney(t, 1999, "CZK").Add(mustMoney(t, 1, "CZK"))
	if err != nil {
		t.Fatalf("Add = chyba %v", err)
	}
	if sum.Cents() != 2000 {
		t.Errorf("Add = %d, chci 2000", sum.Cents())
	}

	if _, err := mustMoney(t, 100, "CZK").Add(mustMoney(t, 100, "EUR")); !errors.Is(err, order.ErrCurrencyMismatch) {
		t.Errorf("Add různých měn = %v, chci ErrCurrencyMismatch", err)
	}

	big := mustMoney(t, math.MaxInt64, "CZK")
	if _, err := big.Add(mustMoney(t, 1, "CZK")); !errors.Is(err, order.ErrAmountOverflow) {
		t.Errorf("Add přes limit = %v, chci ErrAmountOverflow", err)
	}
}

func TestMoneyMul(t *testing.T) {
	got, err := mustMoney(t, 19900, "CZK").Mul(3)
	if err != nil {
		t.Fatalf("Mul = chyba %v", err)
	}
	if got.Cents() != 59700 {
		t.Errorf("Mul(3) = %d, chci 59700", got.Cents())
	}

	zero, err := mustMoney(t, 19900, "CZK").Mul(0)
	if err != nil {
		t.Fatalf("Mul(0) = chyba %v", err)
	}
	if zero.Cents() != 0 || zero.Currency() != "CZK" {
		t.Errorf("Mul(0) = %v, chci 0 CZK", zero)
	}

	if _, err := mustMoney(t, 100, "CZK").Mul(-1); !errors.Is(err, order.ErrNegativeAmount) {
		t.Errorf("Mul(-1) = %v, chci ErrNegativeAmount", err)
	}
	if _, err := mustMoney(t, math.MaxInt64/2, "CZK").Mul(3); !errors.Is(err, order.ErrAmountOverflow) {
		t.Errorf("Mul přes limit = %v, chci ErrAmountOverflow", err)
	}
}

func TestMoneyString(t *testing.T) {
	tests := []struct {
		cents int64
		want  string
	}{
		{0, "0.00 CZK"},
		{5, "0.05 CZK"},
		{100, "1.00 CZK"},
		{19999, "199.99 CZK"},
	}
	for _, tt := range tests {
		m := mustMoney(t, tt.cents, "CZK")
		if got := m.String(); got != tt.want {
			t.Errorf("Money(%d).String() = %q, chci %q", tt.cents, got, tt.want)
		}
	}
}

func TestMoneyJePorovnatelna(t *testing.T) {
	if mustMoney(t, 100, "czk") != mustMoney(t, 100, "CZK") {
		t.Error("stejné částky se mají rovnat přes ==")
	}
	if mustMoney(t, 100, "CZK") == mustMoney(t, 100, "EUR") {
		t.Error("částky v různých měnách se rovnat nemají")
	}
}
