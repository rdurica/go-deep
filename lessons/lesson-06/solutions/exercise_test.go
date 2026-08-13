package solutions_test

import (
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-06/solutions"
)

func TestSwap(t *testing.T) {
	a, b := 1, 2
	exercise.Swap(&a, &b)

	if a != 2 || b != 1 {
		t.Errorf("po Swap je (a, b) = (%d, %d), chci (2, 1)", a, b)
	}
}

func TestSwapSamePointer(t *testing.T) {
	x := 7
	exercise.Swap(&x, &x)

	if x != 7 {
		t.Errorf("Swap(&x, &x) změnil x na %d, chci 7", x)
	}
}

func TestSwapNil(t *testing.T) {
	x := 5
	exercise.Swap(&x, nil)
	exercise.Swap(nil, &x)
	exercise.Swap(nil, nil)

	if x != 5 {
		t.Errorf("Swap s nil změnil x na %d, chci 5", x)
	}
}

func boolPtr(b bool) *bool { return &b }

func TestApplyDefaultsEmptyConfig(t *testing.T) {
	var c exercise.Config
	exercise.ApplyDefaults(&c)

	if c.Host != "localhost" {
		t.Errorf("Host = %q, chci %q", c.Host, "localhost")
	}
	if c.Port != 8080 {
		t.Errorf("Port = %d, chci 8080", c.Port)
	}
	if c.Debug == nil {
		t.Fatal("Debug je nil, chci pointer na false")
	}
	if *c.Debug {
		t.Errorf("*Debug = true, chci false")
	}
}

func TestApplyDefaultsKeepsSetValues(t *testing.T) {
	debug := true
	c := exercise.Config{Host: "example.com", Port: 9000, Debug: &debug}
	exercise.ApplyDefaults(&c)

	if c.Host != "example.com" {
		t.Errorf("Host = %q, chci %q", c.Host, "example.com")
	}
	if c.Port != 9000 {
		t.Errorf("Port = %d, chci 9000", c.Port)
	}
	if c.Debug != &debug {
		t.Error("Debug byl nahrazen jiným pointerem, chci zachovat původní")
	}
	if !*c.Debug {
		t.Error("*Debug = false, chci true")
	}
}

func TestApplyDefaultsKeepsExplicitFalse(t *testing.T) {
	// Tohle je důvod, proč je Debug pointer: false nastavené uživatelem
	// se musí odlišit od "nenastaveno".
	c := exercise.Config{Debug: boolPtr(false)}
	exercise.ApplyDefaults(&c)

	if c.Debug == nil || *c.Debug {
		t.Error("explicitní false se má zachovat")
	}
}

func TestApplyDefaultsPartial(t *testing.T) {
	c := exercise.Config{Port: 3000}
	exercise.ApplyDefaults(&c)

	if c.Host != "localhost" {
		t.Errorf("Host = %q, chci %q", c.Host, "localhost")
	}
	if c.Port != 3000 {
		t.Errorf("Port = %d, chci 3000", c.Port)
	}
}

func TestApplyDefaultsNilPointer(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ApplyDefaults(nil) panikovala: %v", r)
		}
	}()
	exercise.ApplyDefaults(nil)
}

// applyToCopy dostane kopii structu, takže venku není nic vidět.
func applyToCopy(c exercise.Config) exercise.Config {
	exercise.ApplyDefaults(&c)
	return c
}

func TestConfigPassedByValue(t *testing.T) {
	original := exercise.Config{}
	filled := applyToCopy(original)

	if original.Host != "" || original.Port != 0 || original.Debug != nil {
		t.Errorf("originál se změnil na %+v, chci zero value", original)
	}
	if filled.Host != "localhost" {
		t.Errorf("kopie má Host = %q, chci %q", filled.Host, "localhost")
	}
}

func TestIncrementAll(t *testing.T) {
	nums := []int{1, 2, 3}
	exercise.IncrementAll(nums)

	want := []int{2, 3, 4}
	for i := range want {
		if nums[i] != want[i] {
			t.Fatalf("IncrementAll dala %v, chci %v", nums, want)
		}
	}
}

func TestIncrementAllEmpty(t *testing.T) {
	exercise.IncrementAll(nil)
	exercise.IncrementAll([]int{})
}

func TestIncrementAllViaSubslice(t *testing.T) {
	// Podslice sdílí podkladové pole, takže změna je vidět i v originálu.
	nums := []int{1, 2, 3}
	exercise.IncrementAll(nums[1:])

	want := []int{1, 3, 4}
	for i := range want {
		if nums[i] != want[i] {
			t.Fatalf("po IncrementAll(nums[1:]) je nums = %v, chci %v", nums, want)
		}
	}
}

func TestAppendSafe(t *testing.T) {
	nums := []int{1, 2}
	got := exercise.AppendSafe(nums, 3)

	want := []int{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("AppendSafe vrátila %v, chci %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("AppendSafe vrátila %v, chci %v", got, want)
		}
	}
	if len(nums) != 2 {
		t.Errorf("vstupní slice má délku %d, chci 2", len(nums))
	}
}

func TestAppendSafeDoesNotTouchBacking(t *testing.T) {
	// Slice s rezervou v kapacitě: obyčejný append by zapsal do stejného pole.
	backing := make([]int, 1, 8)
	backing[0] = 1

	exercise.AppendSafe(backing, 99)

	extended := backing[:2]
	if extended[1] == 99 {
		t.Error("AppendSafe zapsala do podkladového pole volajícího, chci kopii")
	}
}

func TestAppendSafeNilInput(t *testing.T) {
	got := exercise.AppendSafe(nil, 5)
	if len(got) != 1 || got[0] != 5 {
		t.Errorf("AppendSafe(nil, 5) = %v, chci [5]", got)
	}
}

// toSlice převede seznam na slice, ať se dá pohodlně porovnávat.
func toSlice(head *exercise.Node) []int {
	var out []int
	for node := head; node != nil; node = node.Next {
		out = append(out, node.Val)
	}
	return out
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestPush(t *testing.T) {
	var head *exercise.Node
	head = exercise.Push(head, 1)
	head = exercise.Push(head, 2)
	head = exercise.Push(head, 3)

	if got, want := toSlice(head), []int{3, 2, 1}; !equal(got, want) {
		t.Errorf("seznam = %v, chci %v — Push vkládá na začátek", got, want)
	}
}

func TestPushNilHead(t *testing.T) {
	head := exercise.Push(nil, 42)
	if head == nil {
		t.Fatal("Push(nil, 42) vrátil nil, chci nový nod")
	}
	if head.Val != 42 || head.Next != nil {
		t.Errorf("nod = {%d, %v}, chci {42, nil}", head.Val, head.Next)
	}
}

// listFrom sestaví seznam z literálů Node — bez volání studentova Push.
func listFrom(vals ...int) *exercise.Node {
	var head *exercise.Node
	for i := len(vals) - 1; i >= 0; i-- {
		head = &exercise.Node{Val: vals[i], Next: head}
	}
	return head
}

func TestLen(t *testing.T) {
	if got := exercise.Len(nil); got != 0 {
		t.Errorf("Len(nil) = %d, chci 0", got)
	}

	tests := []struct {
		head *exercise.Node
		want int
	}{
		{&exercise.Node{Val: 0}, 1},
		{&exercise.Node{Val: 1, Next: &exercise.Node{Val: 0}}, 2},
		{listFrom(4, 3, 2, 1, 0), 5},
	}
	for _, tt := range tests {
		if got := exercise.Len(tt.head); got != tt.want {
			t.Fatalf("Len(%v) = %d, chci %d", toSlice(tt.head), got, tt.want)
		}
	}
}
