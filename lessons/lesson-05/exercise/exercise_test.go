package exercise_test

import (
	"fmt"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-05/exercise"
)

func TestPointAdd(t *testing.T) {
	tests := []struct {
		name string
		p, q exercise.Point
		want exercise.Point
	}{
		{"zero values", exercise.Point{}, exercise.Point{}, exercise.Point{}},
		{"kladné", exercise.Point{X: 1, Y: 2}, exercise.Point{X: 3, Y: 4}, exercise.Point{X: 4, Y: 6}},
		{"záporné", exercise.Point{X: 1, Y: 2}, exercise.Point{X: -1, Y: -2}, exercise.Point{}},
		{"poziční literál", exercise.Point{5, 5}, exercise.Point{-2, 7}, exercise.Point{3, 12}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// struct je porovnatelný, stačí ==
			if got := tt.p.Add(tt.q); got != tt.want {
				t.Errorf("%v.Add(%v) = %v, chci %v", tt.p, tt.q, got, tt.want)
			}
		})
	}
}

func TestPointAddNemeniPrijemce(t *testing.T) {
	p := exercise.Point{X: 1, Y: 2}
	p.Add(exercise.Point{X: 10, Y: 10})

	if p != (exercise.Point{X: 1, Y: 2}) {
		t.Errorf("po Add je p = %v, chci %v — value receiver nesmí měnit originál",
			p, exercise.Point{X: 1, Y: 2})
	}
}

func TestPointString(t *testing.T) {
	tests := []struct {
		p    exercise.Point
		want string
	}{
		{exercise.Point{}, "(0,0)"},
		{exercise.Point{X: 1, Y: 2}, "(1,2)"},
		{exercise.Point{X: -3, Y: 40}, "(-3,40)"},
	}
	for _, tt := range tests {
		if got := tt.p.String(); got != tt.want {
			t.Errorf("Point{%d, %d}.String() = %q, chci %q", tt.p.X, tt.p.Y, got, tt.want)
		}
	}
}

func TestPointJeStringer(t *testing.T) {
	var s fmt.Stringer = exercise.Point{X: 7, Y: 8}
	if got := fmt.Sprintf("%v", s); got != "(7,8)" {
		t.Errorf("fmt.Sprintf(%%v, Point{7, 8}) = %q, chci %q", got, "(7,8)")
	}
}

func TestPointPorovnatelnost(t *testing.T) {
	a := exercise.Point{X: 1, Y: 2}
	b := exercise.Point{X: 1, Y: 2}
	c := exercise.Point{X: 2, Y: 1}

	if a != b {
		t.Error("dva structy se stejnými poli si mají být rovny")
	}
	if a == c {
		t.Error("dva structy s jinými poli si nemají být rovny")
	}
}

func TestCounter(t *testing.T) {
	var c exercise.Counter // zero value je použitelná

	if got := c.Value(); got != 0 {
		t.Errorf("zero value Counter má Value() = %d, chci 0", got)
	}

	c.Inc()
	c.Inc()
	if got := c.Value(); got != 2 {
		t.Errorf("po dvou Inc() je Value() = %d, chci 2", got)
	}

	c.Add(10)
	if got := c.Value(); got != 12 {
		t.Errorf("po Add(10) je Value() = %d, chci 12", got)
	}

	c.Add(-20)
	if got := c.Value(); got != -8 {
		t.Errorf("po Add(-20) je Value() = %d, chci -8", got)
	}
}

func TestCounterPresPointer(t *testing.T) {
	c := &exercise.Counter{}
	c.Inc()
	c.Add(4)

	if got := c.Value(); got != 5 {
		t.Errorf("Value() = %d, chci 5", got)
	}
}

// incCopy dostane kopii, protože Counter se předává hodnotou.
func incCopy(c exercise.Counter) int {
	c.Inc()
	c.Inc()
	return c.Value()
}

func TestCounterKopieNemeniOriginal(t *testing.T) {
	var c exercise.Counter
	c.Inc()

	if got := incCopy(c); got != 3 {
		t.Errorf("kopie po dvou Inc() má Value() = %d, chci 3", got)
	}
	if got := c.Value(); got != 1 {
		t.Errorf("originál má Value() = %d, chci 1 — kopie ho nesmí ovlivnit", got)
	}

	// přiřazení structu je kopie, ne sdílení
	d := c
	d.Add(100)
	if got := c.Value(); got != 1 {
		t.Errorf("po d := c a d.Add(100) má originál Value() = %d, chci 1", got)
	}
	if got := d.Value(); got != 101 {
		t.Errorf("kopie má Value() = %d, chci 101", got)
	}
}

type incrementer interface {
	Inc()
}

func TestCounterMethodSet(t *testing.T) {
	// Inc má pointer receiver, takže je v method setu *Counter, ne Counter.
	var value any = exercise.Counter{}
	if _, ok := value.(incrementer); ok {
		t.Error("Counter (hodnota) nemá mít Inc v method setu — použij pointer receiver")
	}

	var pointer any = &exercise.Counter{}
	if _, ok := pointer.(incrementer); !ok {
		t.Error("*Counter má mít Inc v method setu")
	}
}

func TestBaseDescribe(t *testing.T) {
	b := exercise.Base{ID: "b1"}
	if got, want := b.Describe(), "base:b1"; got != want {
		t.Errorf("Base.Describe() = %q, chci %q", got, want)
	}
}

func TestUserPromotionPole(t *testing.T) {
	u := exercise.User{Base: exercise.Base{ID: "u1"}, Name: "Radek"}

	if got := u.ID; got != "u1" {
		t.Errorf("u.ID = %q, chci %q — pole vloženého structu se promotuje", got, "u1")
	}
	if got := u.Base.ID; got != "u1" {
		t.Errorf("u.Base.ID = %q, chci %q", got, "u1")
	}
}

func TestUserShadowing(t *testing.T) {
	u := exercise.User{Base: exercise.Base{ID: "u1"}, Name: "Radek"}

	if got, want := u.Describe(), "user:Radek (base:u1)"; got != want {
		t.Errorf("User.Describe() = %q, chci %q", got, want)
	}
	if got, want := u.Base.Describe(), "base:u1"; got != want {
		t.Errorf("u.Base.Describe() = %q, chci %q — původní metoda musí zůstat dostupná", got, want)
	}
}

func newTestAdmin() exercise.Admin {
	return exercise.Admin{
		User: exercise.User{
			Base: exercise.Base{ID: "a1"},
			Name: "Root",
		},
		Level: 9,
	}
}

func TestAdminPromotionPresDveUrovne(t *testing.T) {
	a := newTestAdmin()

	if got := a.ID; got != "a1" {
		t.Errorf("a.ID = %q, chci %q", got, "a1")
	}
	if got := a.Name; got != "Root" {
		t.Errorf("a.Name = %q, chci %q", got, "Root")
	}
	if got, want := a.Describe(), "user:Root (base:a1)"; got != want {
		t.Errorf("Admin.Describe() = %q, chci %q — metoda se promotuje z User", got, want)
	}
	if got, want := a.User.Base.Describe(), "base:a1"; got != want {
		t.Errorf("a.User.Base.Describe() = %q, chci %q", got, want)
	}
}

func TestAdminTag(t *testing.T) {
	tests := []struct {
		a    exercise.Admin
		want string
	}{
		{newTestAdmin(), "admin:a1/9"},
		{exercise.Admin{}, "admin:/0"},
		{
			exercise.Admin{User: exercise.User{Base: exercise.Base{ID: "x"}}, Level: -1},
			"admin:x/-1",
		},
	}
	for _, tt := range tests {
		if got := tt.a.Tag(); got != tt.want {
			t.Errorf("Admin.Tag() = %q, chci %q", got, tt.want)
		}
	}
}

type describer interface {
	Describe() string
}

func TestEmbeddingSplnujeInterface(t *testing.T) {
	// Promotovaná metoda se počítá do method setu, takže Admin splní
	// interface, aniž by Describe sám implementoval.
	var d describer = newTestAdmin()

	if got, want := d.Describe(), "user:Root (base:a1)"; got != want {
		t.Errorf("describer.Describe() = %q, chci %q", got, want)
	}
}
