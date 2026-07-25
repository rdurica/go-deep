package solutions_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-54/solutions"
)

// Sinky pro benchmarky.
var sinkMap map[string]any

var errTest = errors.New("testovací chyba")

func TestResultOk(t *testing.T) {
	r := exercise.Ok(42)
	if !r.IsOk() {
		t.Error("Ok(42).IsOk() = false, chci true")
	}
	v, err := r.Unwrap()
	if err != nil {
		t.Errorf("Ok(42).Unwrap() vrátil chybu %v", err)
	}
	if v != 42 {
		t.Errorf("Ok(42).Unwrap() = %d, chci 42", v)
	}
}

func TestResultErr(t *testing.T) {
	r := exercise.Err[string](errTest)
	if r.IsOk() {
		t.Error("Err(...).IsOk() = true, chci false")
	}
	v, err := r.Unwrap()
	if !errors.Is(err, errTest) {
		t.Errorf("Err(...).Unwrap() vrátil chybu %v, chci errTest", err)
	}
	if v != "" {
		t.Errorf("Err(...).Unwrap() = %q, chci nulovou hodnotu", v)
	}
}

func TestResultNuloveHodnoty(t *testing.T) {
	var zero exercise.Result[int]
	if !zero.IsOk() {
		t.Error("nulový Result má být Ok")
	}
	if v, err := zero.Unwrap(); v != 0 || err != nil {
		t.Errorf("nulový Result.Unwrap() = (%d, %v), chci (0, nil)", v, err)
	}
	if !exercise.Err[int](nil).IsOk() {
		t.Error("Err(nil) má být Ok")
	}
}

func TestMap(t *testing.T) {
	r := exercise.Map(exercise.Ok(21), func(n int) string { return strconv.Itoa(n * 2) })
	v, err := r.Unwrap()
	if err != nil {
		t.Fatalf("Map nad Ok vrátil chybu %v", err)
	}
	if v != "42" {
		t.Errorf("Map(Ok(21), ×2 na string) = %q, chci %q", v, "42")
	}

	failed := exercise.Map(exercise.Err[int](errTest), func(n int) string {
		t.Error("f se nad chybou nesmí volat")
		return ""
	})
	if _, err := failed.Unwrap(); !errors.Is(err, errTest) {
		t.Errorf("Map nad Err vrátil chybu %v, chci errTest", err)
	}
}

func TestMapZmenaTypu(t *testing.T) {
	// Řetěz Map přes tři různé typy ověří, že typový parametr U opravdu funguje.
	step1 := exercise.Map(exercise.Ok("7"), func(s string) int {
		n, _ := strconv.Atoi(s)
		return n
	})
	step2 := exercise.Map(step1, func(n int) []int { return []int{n, n} })
	v, err := step2.Unwrap()
	if err != nil {
		t.Fatalf("řetěz Map vrátil chybu %v", err)
	}
	if len(v) != 2 || v[0] != 7 || v[1] != 7 {
		t.Errorf("řetěz Map = %v, chci [7 7]", v)
	}
}

func TestMust(t *testing.T) {
	if got := exercise.Must(strconv.Atoi("123")); got != 123 {
		t.Errorf("Must(Atoi(\"123\")) = %d, chci 123", got)
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Must nad chybou nepanikoval")
		}
		if err, ok := r.(error); !ok || !errors.Is(err, errTest) {
			t.Errorf("Must panikoval s %v, chci errTest", r)
		}
	}()
	exercise.Must("x", errTest)
}

func TestStructToMap(t *testing.T) {
	u := exercise.NewUser(7, "Alice", "alice@example.com", true, "tajne")
	got, err := exercise.StructToMap(u)
	if err != nil {
		t.Fatalf("StructToMap = chyba %v", err)
	}
	want := map[string]any{"id": 7, "name": "Alice", "Active": true}
	overMapy(t, "StructToMap", got, want)

	if _, ok := got["Email"]; ok {
		t.Error("pole s tagem map:\"-\" se nemá do mapy dostat")
	}
	if _, ok := got["password"]; ok {
		t.Error("neexportované pole se nemá do mapy dostat")
	}
	if u.Password() != "tajne" {
		t.Error("StructToMap nemá měnit vstup")
	}
}

func TestStructToMapPointer(t *testing.T) {
	u := exercise.NewUser(1, "Bob", "b@example.com", false, "x")
	got, err := exercise.StructToMap(&u)
	if err != nil {
		t.Fatalf("StructToMap(&u) = chyba %v", err)
	}
	overMapy(t, "StructToMap(&u)", got, map[string]any{"id": 1, "name": "Bob", "Active": false})
}

func TestStructToMapChyby(t *testing.T) {
	var nilUser *exercise.User
	bad := []any{nil, 42, "text", []int{1}, nilUser}
	for _, in := range bad {
		if got, err := exercise.StructToMap(in); err == nil {
			t.Errorf("StructToMap(%#v) = %v, chci chybu", in, got)
		} else if !errors.Is(err, exercise.ErrNotStruct) {
			t.Errorf("StructToMap(%#v) vrátil %v, chci obalený ErrNotStruct", in, err)
		}
	}
}

func TestUserToMapSeShodujeSReflexi(t *testing.T) {
	for i := 0; i < 5; i++ {
		u := exercise.NewUser(i, fmt.Sprintf("user-%d", i), "x@example.com", i%2 == 0, "p")
		want, err := exercise.StructToMap(u)
		if err != nil {
			t.Fatalf("StructToMap = chyba %v", err)
		}
		overMapy(t, "UserToMap", exercise.UserToMap(u), want)
	}
}

func TestSetDefaults(t *testing.T) {
	var c exercise.Config
	if err := exercise.SetDefaults(&c); err != nil {
		t.Fatalf("SetDefaults = chyba %v", err)
	}
	if c.Host != "localhost" {
		t.Errorf("Host = %q, chci %q", c.Host, "localhost")
	}
	if c.Port != 8080 {
		t.Errorf("Port = %d, chci 8080", c.Port)
	}
	if !c.Debug {
		t.Error("Debug = false, chci true")
	}
	if c.Retries != 3 {
		t.Errorf("Retries = %d, chci 3", c.Retries)
	}
	if c.Timeout != 0 {
		t.Errorf("Timeout = %d, chci 0 (pole bez tagu)", c.Timeout)
	}
	if c.Secret() != "" {
		t.Errorf("secret = %q, neexportované pole se nemá nastavovat", c.Secret())
	}
}

func TestSetDefaultsNeprepisuje(t *testing.T) {
	c := exercise.Config{Host: "db.internal", Port: 5432}
	if err := exercise.SetDefaults(&c); err != nil {
		t.Fatalf("SetDefaults = chyba %v", err)
	}
	if c.Host != "db.internal" {
		t.Errorf("Host = %q, vyplněné pole se nemá přepsat", c.Host)
	}
	if c.Port != 5432 {
		t.Errorf("Port = %d, vyplněné pole se nemá přepsat", c.Port)
	}
	if !c.Debug {
		t.Error("Debug = false, nulové pole se doplnit má")
	}
}

func TestSetDefaultsChyby(t *testing.T) {
	t.Run("není pointer", func(t *testing.T) {
		if err := exercise.SetDefaults(exercise.Config{}); !errors.Is(err, exercise.ErrNotPointer) {
			t.Errorf("chci ErrNotPointer, dostal jsem %v", err)
		}
	})
	t.Run("nil pointer", func(t *testing.T) {
		var c *exercise.Config
		if err := exercise.SetDefaults(c); !errors.Is(err, exercise.ErrNotPointer) {
			t.Errorf("chci ErrNotPointer, dostal jsem %v", err)
		}
	})
	t.Run("pointer na jiný typ", func(t *testing.T) {
		n := 0
		if err := exercise.SetDefaults(&n); !errors.Is(err, exercise.ErrNotPointer) {
			t.Errorf("chci ErrNotPointer, dostal jsem %v", err)
		}
	})
	t.Run("nepodporovaný typ", func(t *testing.T) {
		var bc exercise.BadConfig
		if err := exercise.SetDefaults(&bc); !errors.Is(err, exercise.ErrUnsupportedKind) {
			t.Errorf("chci ErrUnsupportedKind, dostal jsem %v", err)
		}
	})
}

func TestDiscount(t *testing.T) {
	if got := exercise.Discount(0); got != 0 {
		t.Errorf("Discount(0) = %d, chci 0", got)
	}
	if got := exercise.Discount(-100); got != 0 {
		t.Errorf("Discount(-100) = %d, chci 0", got)
	}
}

// overMapy porovná dvě mapy klíč po klíči.
func overMapy(t *testing.T, jmeno string, got, want map[string]any) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %v (%d klíčů), chci %v (%d klíčů)", jmeno, got, len(got), want, len(want))
	}
	for k, w := range want {
		g, ok := got[k]
		if !ok {
			t.Errorf("%s: chybí klíč %q", jmeno, k)
			continue
		}
		if g != w {
			t.Errorf("%s[%q] = %v (%T), chci %v (%T)", jmeno, k, g, g, w, w)
		}
	}
}

func BenchmarkStructToMapReflexe(b *testing.B) {
	u := exercise.NewUser(7, "Alice", "alice@example.com", true, "tajne")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkMap, _ = exercise.StructToMap(u)
	}
}

func BenchmarkUserToMapRucne(b *testing.B) {
	u := exercise.NewUser(7, "Alice", "alice@example.com", true, "tajne")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkMap = exercise.UserToMap(u)
	}
}
