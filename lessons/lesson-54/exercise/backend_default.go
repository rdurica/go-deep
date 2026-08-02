//go:build !integration

package exercise

// Backend vrací jméno úložiště zapečeného do výchozího buildu.
// Bez build tagu integration (!integration) vrací "memory".
func Backend() string {
	// TODO
	return ""
}
