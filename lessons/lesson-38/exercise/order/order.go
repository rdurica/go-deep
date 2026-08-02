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

// String implementuje fmt.Stringer pro všech pět stavů objednávky.
// new, paid, shipped, cancelled mají odpovídající řetězce; neznámá hodnota → "unknown".
func (s Status) String() string {
	// TODO
	return ""
}

// Line je jedna položka objednávky.
type Line struct {
	SKU            string
	Quantity       int
	UnitPriceCents int64
}

// TotalCents vrací cenu položky: UnitPriceCents * Quantity.
// Např. Line{Quantity:3, UnitPriceCents:1999} → 5997.
func (l Line) TotalCents() int64 {
	// TODO
	return 0
}

// Order je objednávka. Metody přechodů vracejí novou hodnotu,
// původní objednávka zůstává nedotčená.
type Order struct {
	ID     string
	Lines  []Line
	Status Status
}

// New vytvoří objednávku ve stavu new. Ořízne ID; prázdné ID → ErrMissingID;
// prázdné lines → ErrEmptyOrder; vadná položka → ErrInvalidLine.
// Vrátí kopii slice položek. Při chybě nulová Order.
func New(id string, lines []Line) (Order, error) {
	// TODO
	return *new(Order), nil
}

// TotalCents vrací součet cen všech položek objednávky v centech.
// Sčítá TotalCents() každé položky.
func (o Order) TotalCents() int64 {
	// TODO
	return 0
}

// Pay převede stav new → paid. Hodnotový receiver — vrací novou Order, původní zůstane.
// Jiný stav → chyba obalující ErrInvalidTransition.
func (o Order) Pay() (Order, error) {
	// TODO
	return *new(Order), nil
}

// Ship převede stav paid → shipped. Hodnotový receiver vrací novou hodnotu.
// Jiný stav (včetně new, shipped, cancelled) → ErrInvalidTransition.
func (o Order) Ship() (Order, error) {
	// TODO
	return *new(Order), nil
}

// Cancel zruší objednávku ve stavu new nebo paid.
// Odeslaná nebo už zrušená → ErrInvalidTransition. Vrací novou hodnotu.
func (o Order) Cancel() (Order, error) {
	// TODO
	return *new(Order), nil
}
