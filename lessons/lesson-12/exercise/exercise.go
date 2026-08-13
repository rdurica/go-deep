// Package exercise obsahuje cvičení lekce 12.
package exercise

import "math"

// Shape je cokoli, co umí spočítat svůj obsah.
type Shape interface {
	Area() float64
}

// Rect je obdélník o šířce W a výšce H.
type Rect struct {
	W, H float64
}

// Circle je kruh o poloměru R.
type Circle struct {
	R float64
}

// Notifier je cokoli, co umí doručit zprávu.
type Notifier interface {
	Notify(msg string) error
}

// Recorder je testovací implementace Notifieru, která si zprávy jen pamatuje.
// Když je Err nenulová, Notify ji vrátí a zprávu nezaznamená.
type Recorder struct {
	Err      error
	messages []string
}

// MyErr je typ, jehož metoda Area má pointer receiver.
// Slouží k ukázce pasti "nil pointer v non-nil interfacu".
type MyErr struct{}

// --- Stupeň: jednoduchý ---

// TotalArea sečte obsahy všech tvarů. Prvky rovné nil přeskočí (bez paniky).
// Nil i prázdný slice dají 0.
//
// POZOR: kód níže je ZÁMĚRNĚ VADNÝ — nezkontroluje nil prvky v slice.
// Volání Area na nil interface panikuje. Najdi chybu a oprav.
func TotalArea(shapes []Shape) float64 {
	var total float64
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

// --- Stupeň: střední ---

// Describe vrací popis dynamického typu přes type switch:
// nil → "nil", int → "int:42", string → "string:%q", bool → "bool:true/false",
// []int → "[]int:len=3", ostatní → "other:<typ>".
func Describe(v any) string {
	// TODO
	return ""
}

// Notify zaznamená zprávu, pokud r.Err je nil. Při nenulové r.Err vrátí ji a nezaznamená zprávu.
// Zero value Recorder musí fungovat bez konstruktoru.
func (r *Recorder) Notify(msg string) error {
	// TODO
	return nil
}

// Messages vrací kopii zaznamenaných zpráv v pořadí vložení.
// Volající nesmí mutací výsledku změnit vnitřní stav Recorderu.
func (r *Recorder) Messages() []string {
	// TODO
	return nil
}

// --- Stupeň: obtížný ---

// Area vrací 0. Pointer receiver nesahá na pole — metoda funguje i na nil *MyErr.
func (e *MyErr) Area() float64 {
	return 0
}

// ReturnsNilPointer vrací Shape s dynamickým typem *MyErr a dynamickou hodnotou nil.
// Výsledek NENÍ roven nil (nesmíš vracet nil literál).
func ReturnsNilPointer() Shape {
	// TODO
	return *new(Shape)
}

// IsNilInterface vrací true jen když je celá interface hodnota nil (s == nil).
// Typed-nil (non-nil interface s nil pointerem uvnitř) vrací false.
func IsNilInterface(s Shape) bool {
	// TODO
	return false
}

// Area vrací obsah obdélníku W * H. Hotová — ilustruje implicitní implementaci Shape.
func (r Rect) Area() float64 {
	return r.W * r.H
}

// Area vrací obsah kruhu π·R². Hotová — používá math.Pi.
func (c Circle) Area() float64 {
	return math.Pi * c.R * c.R
}
