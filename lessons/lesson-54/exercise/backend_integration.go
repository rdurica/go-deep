//go:build integration

package exercise

// Backend vrací jméno úložiště při buildu s -tags integration.
// V tomto buildu vrací "postgres"; bez tagu se kompiluje backend_default.go.
func Backend() string {
	// TODO
	return ""
}
