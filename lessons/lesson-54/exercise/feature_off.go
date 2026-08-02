//go:build !fancy

package exercise

// FeatureName vrací jméno feature buildu bez tagu fancy.
// Ve výchozím buildu (!fancy) vrací "basic"; s -tags fancy vrací "fancy".
func FeatureName() string {
	// TODO
	return ""
}

// Discount vrací slevu z částky; ve výchozím buildu (!fancy) vždy 0.
// S tagem fancy je to 20 % celočíselně dolů; pro nekladný vstup 0.
func Discount(amount int) int {
	// TODO
	return 0
}
