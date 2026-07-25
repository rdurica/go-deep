// Package solutions obsahuje referenční řešení lekce 05.
package solutions

import "strconv"

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
	return Point{X: p.X + q.X, Y: p.Y + q.Y}
}

// String implementuje fmt.Stringer, formát je "(1,2)".
func (p Point) String() string {
	return "(" + strconv.Itoa(p.X) + "," + strconv.Itoa(p.Y) + ")"
}

// Inc zvýší počítadlo o jedna.
func (c *Counter) Inc() {
	c.n++
}

// Add přičte n k počítadlu. Záporné n počítadlo snižuje.
func (c *Counter) Add(n int) {
	c.n += n
}

// Value vrací aktuální hodnotu počítadla.
func (c Counter) Value() int {
	return c.n
}

// Describe vrací popis základu ve tvaru "base:<ID>".
func (b Base) Describe() string {
	return "base:" + b.ID
}

// Describe překrývá promotovanou metodu Base.Describe.
// Formát je "user:<Name> (base:<ID>)".
func (u User) Describe() string {
	return "user:" + u.Name + " (" + u.Base.Describe() + ")"
}

// Tag vrací "admin:<ID>/<Level>" a používá pole promotovaná přes dvě úrovně.
func (a Admin) Tag() string {
	return "admin:" + a.ID + "/" + strconv.Itoa(a.Level)
}
