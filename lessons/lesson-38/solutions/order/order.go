// Package order je doména objednávek.
//
// Nezná HTTP, databázi ani JSON. Kdyby se zítra služba přepsala na gRPC,
// tenhle balíček se nezmění ani o řádek.
package order

import (
	"errors"
	"fmt"
	"strings"
)

// Doménové chyby objednávky.
var (
	// ErrMissingID znamená, že objednávka nemá identifikátor.
	ErrMissingID = errors.New("objednávka nemá ID")
	// ErrEmptyOrder znamená, že objednávka nemá žádnou položku.
	ErrEmptyOrder = errors.New("objednávka nemá žádnou položku")
	// ErrInvalidLine znamená, že položka objednávky porušuje invariant.
	ErrInvalidLine = errors.New("neplatná položka objednávky")
	// ErrInvalidTransition znamená, že přechod mezi stavy není povolený.
	ErrInvalidTransition = errors.New("nepovolený přechod stavu")
	// ErrNotFound znamená, že objednávka neexistuje.
	ErrNotFound = errors.New("objednávka nenalezena")
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

// Line je jedna položka objednávky.
type Line struct {
	SKU            string
	Quantity       int
	UnitPriceCents int64
}

// TotalCents vrací cenu položky v celých centech.
func (l Line) TotalCents() int64 {
	return int64(l.Quantity) * l.UnitPriceCents
}

// validate ověří invarianty jedné položky.
func (l Line) validate() error {
	switch {
	case strings.TrimSpace(l.SKU) == "":
		return fmt.Errorf("%w: prázdné SKU", ErrInvalidLine)
	case l.Quantity <= 0:
		return fmt.Errorf("%w: množství %d není kladné", ErrInvalidLine, l.Quantity)
	case l.UnitPriceCents <= 0:
		return fmt.Errorf("%w: cena %d není kladná", ErrInvalidLine, l.UnitPriceCents)
	}
	return nil
}

// Order je objednávka. Metody přechodů vracejí novou hodnotu,
// původní objednávka zůstává nedotčená.
type Order struct {
	ID     string
	Lines  []Line
	Status Status
}

// New vytvoří novou objednávku ve stavu StatusNew a ověří invarianty.
func New(id string, lines []Line) (Order, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Order{}, ErrMissingID
	}
	if len(lines) == 0 {
		return Order{}, ErrEmptyOrder
	}
	for i, line := range lines {
		if err := line.validate(); err != nil {
			return Order{}, fmt.Errorf("položka %d: %w", i, err)
		}
	}
	// Kopie: kdyby si volající slice ponechal a změnil ho, prolezla by
	// změna dovnitř objednávky a obešla by validaci.
	owned := make([]Line, len(lines))
	copy(owned, lines)

	return Order{ID: id, Lines: owned, Status: StatusNew}, nil
}

// TotalCents vrací celkovou cenu objednávky v celých centech.
func (o Order) TotalCents() int64 {
	var total int64
	for _, line := range o.Lines {
		total += line.TotalCents()
	}
	return total
}

// Pay převede novou objednávku do stavu StatusPaid.
func (o Order) Pay() (Order, error) {
	return o.transition(StatusPaid, StatusNew)
}

// Ship převede zaplacenou objednávku do stavu StatusShipped.
func (o Order) Ship() (Order, error) {
	return o.transition(StatusShipped, StatusPaid)
}

// Cancel zruší objednávku, která ještě není odeslaná ani zrušená.
func (o Order) Cancel() (Order, error) {
	return o.transition(StatusCancelled, StatusNew, StatusPaid)
}

// transition je jediné místo, kde se stav objednávky mění. Celý stavový
// automat je tak vidět na pěti řádcích výš, ne rozeseto po aplikaci.
func (o Order) transition(to Status, from ...Status) (Order, error) {
	for _, allowed := range from {
		if o.Status == allowed {
			o.Status = to // o je kopie, originál voláním netrpí
			return o, nil
		}
	}
	return o, fmt.Errorf("%w: ze stavu %s nelze do %s", ErrInvalidTransition, o.Status, to)
}
