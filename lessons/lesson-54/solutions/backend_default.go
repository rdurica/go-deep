//go:build !integration

package solutions

// --- Stupeň: jednoduchý ---
// Backend vrací jméno úložiště zapečeného do buildu.
// Tahle varianta se přeloží ve výchozím buildu.
func Backend() string {
	return "memory"
}
