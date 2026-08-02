//go:build fancy

package exercise

// FeatureName vrací jméno feature buildu s tagem fancy.
// Při buildu s -tags fancy vrací "fancy"; bez tagu se kompiluje feature_off.go.
func FeatureName() string {
	// TODO
	return ""
}

// Discount vrací 20 % z částky celočíselně dolů; pro nekladný vstup 0.
func Discount(amount int) int {
	// TODO
	return 0
}
