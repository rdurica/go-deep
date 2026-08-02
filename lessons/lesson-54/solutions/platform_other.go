//go:build !linux

package solutions

// --- Stupeň: jednoduchý ---
// PlatformHint vrací jméno platformy.
func PlatformHint() string {
	return "other"
}
