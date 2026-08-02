//go:build fancy

package solutions

// --- Stupeň: jednoduchý ---
// FeatureName vrací jméno feature buildu.
func FeatureName() string {
	return "fancy"
}

// --- Stupeň: střední ---
// Discount vrací slevu z částky (20 % dolů). Nekladný vstup dává 0.
func Discount(amount int) int {
	if amount <= 0 {
		return 0
	}
	return amount * 20 / 100
}
