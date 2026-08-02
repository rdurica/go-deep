// Package solutions obsahuje referenční řešení lekce 12.
package solutions

import (
	"fmt"
	"math"
	"slices"
)

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

// --- Stupeň: obtížný ---
// Area vrací obsah obdélníku.
func (r Rect) Area() float64 {
	return r.W * r.H
}

// Area vrací obsah kruhu π·R².
// Použij math.Pi (ne vlastní přibližnou konstantu).
func (c Circle) Area() float64 {
	return math.Pi * c.R * c.R
}

// --- Stupeň: jednoduchý ---
// TotalArea sečte obsahy všech tvarů. Prvky rovné nil přeskočí.
func TotalArea(shapes []Shape) float64 {
	var total float64
	for _, s := range shapes {
		if s == nil {
			continue
		}
		total += s.Area()
	}
	return total
}

// --- Stupeň: střední ---
// Describe vrací popis dynamického typu hodnoty.
func Describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("int:%d", x)
	case string:
		return fmt.Sprintf("string:%q", x)
	case bool:
		return fmt.Sprintf("bool:%t", x)
	case []int:
		return fmt.Sprintf("[]int:len=%d", len(x))
	default:
		return fmt.Sprintf("other:%T", x)
	}
}

// Notify zaznamená zprávu, nebo vrátí r.Err.
func (r *Recorder) Notify(msg string) error {
	if r.Err != nil {
		return r.Err
	}
	r.messages = append(r.messages, msg)
	return nil
}

// Messages vrací kopii zaznamenaných zpráv.
func (r *Recorder) Messages() []string {
	return slices.Clone(r.messages)
}

// NotifyAll pošle zprávu všem příjemcům a vrátí první chybu.
func NotifyAll(ns []Notifier, msg string) error {
	for _, n := range ns {
		if n == nil {
			continue
		}
		if err := n.Notify(msg); err != nil {
			return err
		}
	}
	return nil
}

// Area splňuje Shape s pointer receiverem, takže Shape implementuje *MyErr.
func (e *MyErr) Area() float64 {
	return 0
}

// ReturnsNilPointer vrací Shape, uvnitř kterého je nil pointer typu *MyErr.
// Výsledek NENÍ roven nil, i když ukazatel uvnitř nil je.
func ReturnsNilPointer() Shape {
	var p *MyErr // nil pointer
	return p     // zabalený do interfacu: (typ=*MyErr, hodnota=nil) != nil
}

// IsNilInterface vrací true jen když je celá interface hodnota nil (s == nil).
// Typed-nil (non-nil interface s nil pointerem uvnitř) vrací false.
func IsNilInterface(s Shape) bool {
	return s == nil
}
