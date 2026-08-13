//go:build fancy

package solutions

// --- Stupeň: obtížný ---
// FeatureName vrací jméno feature buildu s tagem fancy.
func FeatureName() string {
	return "fancy"
}

// Discount vrací slevu z částky (20 % dolů). Nekladný vstup dává 0.
func Discount(amount int) int {
	if amount <= 0 {
		return 0
	}
	return amount * 20 / 100
}
