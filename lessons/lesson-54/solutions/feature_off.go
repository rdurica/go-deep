//go:build !fancy

package solutions

// --- Stupeň: obtížný ---
// FeatureName vrací jméno feature buildu bez tagu fancy.
func FeatureName() string {
	return "basic"
}

// Discount vrací slevu z částky. Ve výchozím buildu vždy 0.
func Discount(amount int) int {
	return 0
}
