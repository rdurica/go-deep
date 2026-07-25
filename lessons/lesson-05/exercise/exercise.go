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

// Add vrátí nový bod, který je součtem p a q.
func (p Point) Add(q Point) Point {
	// TODO: úkol A
	return *new(Point)
}

// String implementuje fmt.Stringer, formát je "(1,2)".
func (p Point) String() string {
	// TODO: úkol A
	return ""
}

// Inc zvýší počítadlo o jedna.
func (c *Counter) Inc() {
	// TODO: úkol B
}

// Add přičte n k počítadlu. Záporné n počítadlo snižuje.
func (c *Counter) Add(n int) {
	// TODO: úkol B
}

// Value vrací aktuální hodnotu počítadla.
func (c Counter) Value() int {
	// TODO: úkol B
	return 0
}

// Describe vrací popis základu ve tvaru "base:<ID>".
func (b Base) Describe() string {
	// TODO: úkol C
	return ""
}

// Describe překrývá promotovanou metodu Base.Describe.
// Formát je "user:<Name> (base:<ID>)".
func (u User) Describe() string {
	// TODO: úkol C
	return ""
}

// Tag vrací "admin:<ID>/<Level>" a používá pole promotovaná přes dvě úrovně.
func (a Admin) Tag() string {
	// TODO: úkol C
	return ""
}
