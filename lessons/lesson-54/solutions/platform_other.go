//go:build !linux

package solutions

// PlatformHint vrací jméno platformy.
func PlatformHint() string {
	return "other"
}
