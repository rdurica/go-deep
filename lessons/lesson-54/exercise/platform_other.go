//go:build !linux

package exercise

// PlatformHint vrací jméno platformy mimo Linux.
// Soubor platform_other.go se kompiluje na ne-linuxových GOOS; vrací "other".
func PlatformHint() string {
	// TODO
	return ""
}
