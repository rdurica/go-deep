//go:build !fancy

package solutions

// --- Stupeň: jednoduchý ---
// FeatureName vrací jméno feature buildu.
func FeatureName() string {
	return "basic"
}

// --- Stupeň: střední ---
// Discount vrací slevu z částky. Ve výchozím buildu vždy 0.
func Discount(amount int) int {
	return 0
}
