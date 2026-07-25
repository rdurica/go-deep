// Package order je doména objednávek.
//
// Neimportuje nic z okolního světa — žádné encoding/json, net/http ani
// database/sql. Kdyby se služba zítra přepsala na gRPC a Postgres,
// tenhle balíček se nezmění.
package order

import (
	"fmt"
	"strings"
	"time"
)

// Status je stav objednávky.
type Status int

// Stavy objednávky. StatusUnknown je nulová hodnota, tedy "nenastaveno".
const (
	StatusUnknown Status = iota
	StatusNew
	StatusPaid
	StatusShipped
	StatusCancelled
)

// String implementuje fmt.Stringer.
func (s Status) String() string {
	switch s {
	case StatusNew:
		return "new"
	case StatusPaid:
		return "paid"
	case StatusShipped:
		return "shipped"
	case StatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// MaxQuantity je horní mez množství v jedné položce.
const MaxQuantity = 1000

// Line je jedna položka objednávky.
type Line struct {
	SKU       string
	Quantity  int
	UnitPrice Money
}

// NewLine vytvoří položku a ověří její invarianty.
func NewLine(sku string, quantity int, unitPrice Money) (Line, error) {
	line := Line{SKU: strings.TrimSpace(sku), Quantity: quantity, UnitPrice: unitPrice}
	if err := line.validate(); err != nil {
		return Line{}, err
	}
	return line, nil
}

// validate ověří invarianty položky. Volá se i z Order.New, protože Line
// má exportovaná pole a nikdo nezaručí, že prošla konstruktorem.
func (l Line) validate() error {
	switch {
	case strings.TrimSpace(l.SKU) == "":
		return fmt.Errorf("%w: prázdné SKU", ErrInvalidLine)
	case l.Quantity <= 0:
		return fmt.Errorf("%w: množství %d není kladné", ErrInvalidLine, l.Quantity)
	case l.Quantity > MaxQuantity:
		return fmt.Errorf("%w: množství %d nad limitem %d", ErrInvalidLine, l.Quantity, MaxQuantity)
	case l.UnitPrice.IsZero() || l.UnitPrice.Cents() <= 0:
		return fmt.Errorf("%w: jednotková cena musí být kladná", ErrInvalidLine)
	}
	return nil
}

// Total vrací cenu položky.
func (l Line) Total() (Money, error) {
	return l.UnitPrice.Mul(l.Quantity)
}

// Order je objednávka. Přechody stavu vracejí novou hodnotu, takže
// nemůže vzniknout objednávka "napůl změněná" po selhání ukládání.
type Order struct {
	ID       string
	Customer string
	Lines    []Line
	Status   Status
	PlacedAt time.Time
}

// New vytvoří objednávku ve stavu StatusNew a ověří všechny invarianty.
func New(id, customer string, lines []Line, placedAt time.Time) (Order, error) {
	id = strings.TrimSpace(id)
	customer = strings.TrimSpace(customer)
	switch {
	case id == "":
		return Order{}, ErrMissingID
	case customer == "":
		return Order{}, ErrMissingCustomer
	case placedAt.IsZero():
		return Order{}, ErrMissingTimestamp
	case len(lines) == 0:
		return Order{}, ErrEmptyOrder
	}
	for i, line := range lines {
		if err := line.validate(); err != nil {
			return Order{}, fmt.Errorf("položka %d: %w", i, err)
		}
	}

	// Kopie: kdyby si volající slice ponechal a přepsal ho, obešel by
	// veškerou validaci provedenou o tři řádky výš.
	owned := make([]Line, len(lines))
	copy(owned, lines)

	o := Order{
		ID:       id,
		Customer: customer,
		Lines:    owned,
		Status:   StatusNew,
		PlacedAt: placedAt.UTC(),
	}
	// Součet ověří i to, že všechny položky mají stejnou měnu.
	if _, err := o.Total(); err != nil {
		return Order{}, err
	}
	return o, nil
}

// Total vrací celkovou cenu objednávky.
func (o Order) Total() (Money, error) {
	var total Money
	for i, line := range o.Lines {
		lineTotal, err := line.Total()
		if err != nil {
			return Money{}, fmt.Errorf("položka %d: %w", i, err)
		}
		total, err = total.Add(lineTotal)
		if err != nil {
			return Money{}, fmt.Errorf("položka %d: %w", i, err)
		}
	}
	return total, nil
}

// Pay převede novou objednávku do stavu StatusPaid.
func (o Order) Pay() (Order, error) { return o.transition(StatusPaid, StatusNew) }

// Ship převede zaplacenou objednávku do stavu StatusShipped.
func (o Order) Ship() (Order, error) { return o.transition(StatusShipped, StatusPaid) }

// Cancel zruší objednávku, která ještě není odeslaná ani zrušená.
func (o Order) Cancel() (Order, error) {
	return o.transition(StatusCancelled, StatusNew, StatusPaid)
}

// transition je jediné místo, kde se mění stav. Celý povolený graf
// přechodů je tak vidět na třech řádcích výš.
func (o Order) transition(to Status, from ...Status) (Order, error) {
	for _, allowed := range from {
		if o.Status == allowed {
			o.Status = to // o je kopie, volajícího to nezmění
			return o, nil
		}
	}
	return o, fmt.Errorf("%w: ze stavu %s nelze do %s", ErrInvalidTransition, o.Status, to)
}
