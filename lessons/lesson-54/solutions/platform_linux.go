//go:build linux

package solutions

// PlatformHint vrací jméno platformy. Sufix _linux.go by stačil sám,
// značka //go:build linux je tu navíc jako ukázka.
func PlatformHint() string {
	return "linux"
}
