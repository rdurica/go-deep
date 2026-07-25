// Package exercise obsahuje cvičení lekce 51.
package exercise

import (
	"errors"
	"fmt"
	"time"
)

// ErrSyntax označuje syntakticky chybnou verzi nebo pseudo-verzi.
var ErrSyntax = errors.New("neplatný formát verze")

// ErrMajorSuffix označuje cestu modulu s chybným major sufixem (/v0, /v1, /v01).
var ErrMajorSuffix = errors.New("neplatný major sufix v cestě modulu")

// ErrNoVersions označuje modul bez jediného požadavku na verzi.
var ErrNoVersions = errors.New("modul bez požadované verze")

// ErrIncompatible označuje rozpor mezi major verzí v cestě a ve verzi modulu.
var ErrIncompatible = errors.New("major verze v cestě neodpovídá verzi modulu")

// Version je sémantická verze modulu ve tvaru vMAJOR.MINOR.PATCH[-PRERELEASE].
type Version struct {
	Major int
	Minor int
	Patch int
	Pre   string // bez vedoucí pomlčky; prázdné u finálního vydání
}

// String vrací verzi v kanonickém tvaru s prefixem "v".
func (v Version) String() string {
	s := fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		s += "-" + v.Pre
	}
	return s
}

// ParseSemver rozparsuje sémantickou verzi. Prefix "v" je volitelný.
func ParseSemver(s string) (Version, error) {
	// TODO: úkol A
	return *new(Version), nil
}

// Compare porovná dvě verze podle semver pravidel a vrací -1, 0 nebo 1.
func Compare(a, b Version) int {
	// TODO: úkol A
	return 0
}

// ParsePseudoVersion rozloží pseudo-verzi na základ, čas commitu a revizi.
func ParsePseudoVersion(s string) (base string, ts time.Time, rev string, err error) {
	// TODO: úkol B
	return
}

// IsPseudo vrací true, pokud je s platná pseudo-verze.
func IsPseudo(s string) bool {
	// TODO: úkol B
	return false
}

// MajorSuffix vrátí major verzi zakódovanou v cestě modulu.
func MajorSuffix(modulePath string) (int, error) {
	// TODO: úkol B
	return 0, nil
}

// SelectVersions implementuje minimal version selection nad požadavky modulů.
func SelectVersions(reqs map[string][]string) (map[string]string, error) {
	// TODO: úkol C
	return nil, nil
}

// CheckCompat ověří, že major verze v import path odpovídá verzi modulu.
func CheckCompat(importPath, moduleVersion string) error {
	// TODO: úkol C
	return nil
}
