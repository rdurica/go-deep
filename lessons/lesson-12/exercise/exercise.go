// Package exercise obsahuje cvičení lekce 12.
package exercise

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

// Area vrací obsah obdélníku.
func (r Rect) Area() float64 {
	// TODO: úkol A
	return 0
}

// Area vrací obsah kruhu.
func (c Circle) Area() float64 {
	// TODO: úkol A
	return 0
}

// TotalArea sečte obsahy všech tvarů. Prvky rovné nil přeskočí.
func TotalArea(shapes []Shape) float64 {
	// TODO: úkol A
	return 0
}

// Describe vrací popis dynamického typu hodnoty.
func Describe(v any) string {
	// TODO: úkol B
	return ""
}

// Notify zaznamená zprávu, nebo vrátí r.Err.
func (r *Recorder) Notify(msg string) error {
	// TODO: úkol B
	return nil
}

// Messages vrací kopii zaznamenaných zpráv.
func (r *Recorder) Messages() []string {
	// TODO: úkol B
	return nil
}

// NotifyAll pošle zprávu všem příjemcům a vrátí první chybu.
func NotifyAll(ns []Notifier, msg string) error {
	// TODO: úkol B
	return nil
}

// Area splňuje Shape s pointer receiverem, takže Shape implementuje *MyErr.
func (e *MyErr) Area() float64 {
	// TODO: úkol C
	return 0
}

// ReturnsNilPointer vrací Shape, uvnitř kterého je nil pointer typu *MyErr.
// Výsledek NENÍ roven nil, i když ukazatel uvnitř nil je.
func ReturnsNilPointer() Shape {
	// TODO: úkol C
	return *new(Shape)
}

// IsNilInterface vrací true, jen když je celá interface hodnota nil.
func IsNilInterface(s Shape) bool {
	// TODO: úkol C
	return false
}
