//go:build !fancy

package solutions

// FeatureName vrací jméno feature buildu.
func FeatureName() string {
	return "basic"
}

// Discount vrací slevu z částky. Ve výchozím buildu vždy 0.
func Discount(amount int) int {
	return 0
}
