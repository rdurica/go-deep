//go:build !integration

package exercise

// Backend vrací jméno úložiště zapečeného do buildu.
// Tahle varianta se přeloží ve výchozím buildu.
func Backend() string {
	panic("TODO: úkol C")
}
