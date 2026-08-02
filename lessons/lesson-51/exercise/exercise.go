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

// --- Stupeň: jednoduchý ---
// String vrací verzi v kanonickém tvaru s prefixem "v".
// Prerelease se připojí za pomlčku; prázdné Pre u finálního vydání.
func (v Version) String() string {
	s := fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		s += "-" + v.Pre
	}
	return s
}

// ParseSemver rozparsuje semver: volitelný prefix v, tři číselné složky bez vedoucích nul,
// volitelná prerelease za pomlčkou. Každá chyba obaluje ErrSyntax (ověřitelné přes errors.Is).
func ParseSemver(s string) (Version, error) {
	// TODO
	return Version{}, nil
}

// --- Stupeň: střední ---
// Compare porovná dvě verze: major, minor, patch; prerelease bez hodnoty je vyšší než s ním.
// Prerelease se porovnává po identifikátorech (číselné číselně, kratší menší). Vrací -1, 0, 1.
func Compare(a, b Version) int {
	// TODO
	return 0
}

// ParsePseudoVersion rozloží pseudo-verzi na kanonický base, UTC čas a 12místnou revizi.
// Podporuje všechny tři tvary z dokumentace; jinak chyba obalující ErrSyntax.
func ParsePseudoVersion(s string) (base string, ts time.Time, rev string, err error) {
	// TODO
	return
}

// IsPseudo vrací true, pokud s je platná pseudo-verze.
// Postav na ParsePseudoVersion — neduplikuj parsovací logiku.
func IsPseudo(s string) bool {
	// TODO
	return false
}

// --- Stupeň: obtížný ---
// MajorSuffix vrátí major z cesty modulu (bez sufixu → 1).
// /v0, /v1, /v02, prázdná cesta nebo koncové lomítko jsou ErrMajorSuffix.
// Poslední segment, který nevypadá jako v<číslice>, znamená major 1.
func MajorSuffix(modulePath string) (int, error) {
	// TODO
	return 0, nil
}

// SelectVersions provede minimal version selection: pro každý modul nejvyšší z minim.
// Prázdná mapa dá prázdnou nenilovou mapu; prázdný seznam verzí je ErrNoVersions.
// Nerozparsovatelná verze propaguje ErrSyntax s cestou modulu v kontextu.
func SelectVersions(reqs map[string][]string) (map[string]string, error) {
	// TODO
	return nil, nil
}

// CheckCompat ověří import compatibility rule: major v cestě musí odpovídat verzi modulu.
// v0.x i v1.x patří k cestě bez sufixu. Nesoulad → ErrIncompatible; rozbitý vstup → ErrSyntax/ErrMajorSuffix.
func CheckCompat(importPath, moduleVersion string) error {
	// TODO
	return nil
}
