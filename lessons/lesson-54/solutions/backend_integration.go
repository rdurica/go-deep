//go:build integration

package solutions

// --- Stupeň: jednoduchý ---
// Backend vrací jméno úložiště zapečeného do buildu.
// Tahle varianta se přeloží jen s "go build -tags integration".
func Backend() string {
	return "postgres"
}
