// Package exercise obsahuje cvičení lekce 05.
package exercise

// Point je bod v celočíselné mřížce. Struct je porovnatelný přes ==.
type Point struct {
	X, Y int
}

// Counter drží počítadlo. Pole n je neexportované, mění se jen metodami.
type Counter struct {
	n int
}

// Base je společný základ entit s identifikátorem.
type Base struct {
	ID string
}

// User vkládá Base a přidává jméno.
type User struct {
	Base
	Name string
}

// Admin vkládá User a přidává úroveň oprávnění.
type Admin struct {
	User
	Level int
}

// --- Stupeň: jednoduchý ---

// Inc zvýší počítadlo o jedna.
// Zero value musí být použitelná: var c Counter; c.Inc() dá Value() == 1.
// Metoda mění stav přes pointer receiver (*Counter).
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ. Receiver mění kopii místo originálu.
// Najdi chybu a oprav — testy před opravou padají.
func (c Counter) Inc() {
	c.n++
}

// --- Stupeň: střední ---

// Add vrátí nový bod se součtem složek X a Y.
// Přijímače p ani q se nesmí změnit — Point je hodnotový typ.
func (p Point) Add(q Point) Point {
	// TODO
	return *new(Point)
}

// String implementuje fmt.Stringer. Formát přesně "(x,y)" bez mezer.
// Test ověří přes fmt.Sprintf("%v", p).
func (p Point) String() string {
	// TODO
	return ""
}

// Add přičte n k počítadlu. Záporné n počítadlo snižuje.
// Pointer receiver jako u Inc; předání Counter hodnotou do funkce originál nezmění.
func (c *Counter) Add(n int) {
	c.n += n
}

// Value vrací aktuální hodnotu počítadla. Value receiver — čtení nepotřebuje pointer.
func (c Counter) Value() int {
	return c.n
}

// --- Stupeň: obtížný ---

// Describe vrací popis základu ve tvaru "base:<ID>".
func (b Base) Describe() string {
	return "base:" + b.ID
}

// Describe vrací "user:<Name> (<popis base>)".
// Popis base musíš získat voláním u.Base.Describe(), ne slepením řetězce ručně.
func (u User) Describe() string {
	// TODO
	return ""
}

// Tag vrací "admin:<ID>/<Level>" přes promotované pole (a.ID, ne a.User.Base.ID).
// Na převod Level použij strconv.Itoa. Admin nedefinuje Describe — promotuje se z User.
func (a Admin) Tag() string {
	// TODO
	return ""
}
