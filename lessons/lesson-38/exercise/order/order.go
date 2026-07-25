// Package order je doména objednávek.
//
// Nezná HTTP, databázi ani JSON. Kdyby se zítra služba přepsala na gRPC,
// tenhle balíček se nezmění ani o řádek.
package order

import "errors"

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
	panic("TODO: úkol A")
}

// Line je jedna položka objednávky.
type Line struct {
	SKU            string
	Quantity       int
	UnitPriceCents int64
}

// TotalCents vrací cenu položky v celých centech.
func (l Line) TotalCents() int64 {
	panic("TODO: úkol A")
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
	panic("TODO: úkol A")
}

// TotalCents vrací celkovou cenu objednávky v celých centech.
func (o Order) TotalCents() int64 {
	panic("TODO: úkol A")
}

// Pay převede novou objednávku do stavu StatusPaid.
func (o Order) Pay() (Order, error) {
	panic("TODO: úkol A")
}

// Ship převede zaplacenou objednávku do stavu StatusShipped.
func (o Order) Ship() (Order, error) {
	panic("TODO: úkol A")
}

// Cancel zruší objednávku, která ještě není odeslaná ani zrušená.
func (o Order) Cancel() (Order, error) {
	panic("TODO: úkol A")
}
