//go:build integration

package solutions

// Backend vrací jméno úložiště zapečeného do buildu.
// Tahle varianta se přeloží jen s "go build -tags integration".
func Backend() string {
	return "postgres"
}
