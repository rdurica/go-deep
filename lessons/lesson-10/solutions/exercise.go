// Package solutions obsahuje referenční řešení lekce 10.
package solutions

import (
	"fmt"
	"io"
)

// --- Stupeň: jednoduchý ---

// SumWithLog sečte čísla a vede protokol kroků ve tvaru "+N=součet".
func SumWithLog(nums []int) (total int, steps []string) {
	defer func() {
		steps = append(steps, fmt.Sprintf("total=%d", total))
	}()

	for _, n := range nums {
		total += n
		steps = append(steps, fmt.Sprintf("+%d=%d", n, total))
	}
	return total, steps
}

// --- Stupeň: střední ---

// DeferOrder zaregistruje tři defery v LIFO pořadí.
func DeferOrder() (order []string) {
	defer func() { order = append(order, "first") }()
	defer func() { order = append(order, "second") }()
	defer func() { order = append(order, "third") }()
	return order
}

// SafeDivide vydělí a/b a paniku z dělení nulou převede na error.
func SafeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = 0
			err = fmt.Errorf("dělení %d/%d selhalo: %v", a, b, r)
		}
	}()
	return a / b, nil
}

// --- Stupeň: obtížný ---

// WriteAndClose zapíše data a vždy zavolá Close s korektní propagací chyb.
func WriteAndClose(w io.WriteCloser, data []byte) (err error) {
	if w == nil {
		return fmt.Errorf("WriteAndClose: nil writer")
	}
	defer func() {
		if cerr := w.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	_, err = w.Write(data)
	return err
}
