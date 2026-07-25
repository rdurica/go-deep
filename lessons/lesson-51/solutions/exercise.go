// Package solutions obsahuje referenční řešení lekce 51.
package solutions

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

// pseudoTimeLayout je formát časového razítka uvnitř pseudo-verze.
const pseudoTimeLayout = "20060102150405"

// ParseSemver rozparsuje sémantickou verzi. Prefix "v" je volitelný.
func ParseSemver(s string) (Version, error) {
	raw := strings.TrimPrefix(s, "v")
	if raw == "" {
		return Version{}, fmt.Errorf("%q: %w", s, ErrSyntax)
	}
	core, pre := raw, ""
	if i := strings.IndexByte(raw, '-'); i >= 0 {
		core, pre = raw[:i], raw[i+1:]
		if !validPreRelease(pre) {
			return Version{}, fmt.Errorf("%q: neplatné prerelease: %w", s, ErrSyntax)
		}
	}
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("%q: chci tři složky MAJOR.MINOR.PATCH: %w", s, ErrSyntax)
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, ok := parseNumericIdent(p)
		if !ok {
			return Version{}, fmt.Errorf("%q: složka %q není číslo bez vedoucí nuly: %w", s, p, ErrSyntax)
		}
		nums[i] = n
	}
	return Version{Major: nums[0], Minor: nums[1], Patch: nums[2], Pre: pre}, nil
}

// Compare porovná dvě verze podle semver pravidel a vrací -1, 0 nebo 1.
func Compare(a, b Version) int {
	if c := compareInt(a.Major, b.Major); c != 0 {
		return c
	}
	if c := compareInt(a.Minor, b.Minor); c != 0 {
		return c
	}
	if c := compareInt(a.Patch, b.Patch); c != 0 {
		return c
	}
	return comparePreRelease(a.Pre, b.Pre)
}

// ParsePseudoVersion rozloží pseudo-verzi na základ, čas commitu a revizi.
func ParsePseudoVersion(s string) (base string, ts time.Time, rev string, err error) {
	var zero time.Time

	dash := strings.LastIndexByte(s, '-')
	if dash < 0 {
		return "", zero, "", fmt.Errorf("%q: chybí revize: %w", s, ErrSyntax)
	}
	rev = s[dash+1:]
	if len(rev) != 12 || !isLowerHex(rev) {
		return "", zero, "", fmt.Errorf("%q: revize %q není 12 znaků hex: %w", s, rev, ErrSyntax)
	}

	// Razítko je posledních 14 číslic; odděluje ho '-' (bez tagu) nebo '.' (s tagem).
	rest := s[:dash]
	cut := len(rest) - len(pseudoTimeLayout)
	if cut < 2 {
		return "", zero, "", fmt.Errorf("%q: chybí časové razítko: %w", s, ErrSyntax)
	}
	stamp := rest[cut:]
	if !isDigits(stamp) {
		return "", zero, "", fmt.Errorf("%q: razítko %q není 14 číslic: %w", s, stamp, ErrSyntax)
	}
	sep := rest[cut-1]
	if sep != '-' && sep != '.' {
		return "", zero, "", fmt.Errorf("%q: razítko není odděleno '-' ani '.': %w", s, ErrSyntax)
	}
	ts, err = time.Parse(pseudoTimeLayout, stamp)
	if err != nil {
		return "", zero, "", fmt.Errorf("%q: razítko %q není platný čas: %w", s, stamp, ErrSyntax)
	}

	prefix := rest[:cut-1]
	v, err := ParseSemver(prefix)
	if err != nil {
		return "", zero, "", fmt.Errorf("%q: základ %q není semver: %w", s, prefix, ErrSyntax)
	}
	switch {
	case sep == '-' && v.Pre == "":
		// v0.0.0-20230101120000-abcdef123456 — commit bez jakéhokoli tagu.
		base = v.String()
	case sep == '.' && v.Pre == "0":
		// v1.2.4-0.20230101120000-abcdef123456 — commit za tagem v1.2.3.
		base = Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch}.String()
	case sep == '.' && strings.HasSuffix(v.Pre, ".0"):
		// v1.2.4-rc.1.0.20230101120000-abcdef123456 — commit za prerelease tagem.
		base = Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch, Pre: strings.TrimSuffix(v.Pre, ".0")}.String()
	default:
		return "", zero, "", fmt.Errorf("%q: základ %q neodpovídá tvaru pseudo-verze: %w", s, prefix, ErrSyntax)
	}
	return base, ts, rev, nil
}

// IsPseudo vrací true, pokud je s platná pseudo-verze.
func IsPseudo(s string) bool {
	_, _, _, err := ParsePseudoVersion(s)
	return err == nil
}

// MajorSuffix vrátí major verzi zakódovanou v cestě modulu.
func MajorSuffix(modulePath string) (int, error) {
	if modulePath == "" || strings.HasSuffix(modulePath, "/") {
		return 0, fmt.Errorf("%q: prázdná nebo neúplná cesta modulu: %w", modulePath, ErrMajorSuffix)
	}
	slash := strings.LastIndexByte(modulePath, '/')
	if slash < 0 {
		return 1, nil
	}
	last := modulePath[slash+1:]
	if len(last) < 2 || last[0] != 'v' || !isDigits(last[1:]) {
		return 1, nil
	}
	digits := last[1:]
	if len(digits) > 1 && digits[0] == '0' {
		return 0, fmt.Errorf("%q: sufix %q má vedoucí nulu: %w", modulePath, last, ErrMajorSuffix)
	}
	n, err := strconv.Atoi(digits)
	if err != nil {
		return 0, fmt.Errorf("%q: sufix %q není číslo: %w", modulePath, last, ErrMajorSuffix)
	}
	if n < 2 {
		return 0, fmt.Errorf("%q: /v0 a /v1 se nepíšou: %w", modulePath, ErrMajorSuffix)
	}
	return n, nil
}

// SelectVersions implementuje minimal version selection nad požadavky modulů.
func SelectVersions(reqs map[string][]string) (map[string]string, error) {
	out := make(map[string]string, len(reqs))
	for path, list := range reqs {
		if len(list) == 0 {
			return nil, fmt.Errorf("%s: %w", path, ErrNoVersions)
		}
		var best Version
		for i, raw := range list {
			v, err := ParseSemver(raw)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			if i == 0 || Compare(v, best) > 0 {
				best = v
			}
		}
		out[path] = best.String()
	}
	return out, nil
}

// CheckCompat ověří, že major verze v import path odpovídá verzi modulu.
func CheckCompat(importPath, moduleVersion string) error {
	pathMajor, err := MajorSuffix(importPath)
	if err != nil {
		return err
	}
	v, err := ParseSemver(moduleVersion)
	if err != nil {
		return err
	}
	want := v.Major
	if want < 2 {
		want = 1
	}
	if want != pathMajor {
		return fmt.Errorf("cesta %q nese major %d, ale verze %s vyžaduje %d: %w",
			importPath, pathMajor, moduleVersion, want, ErrIncompatible)
	}
	return nil
}

// compareInt vrací -1, 0 nebo 1 podle vzájemného pořadí a a b.
func compareInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// comparePreRelease řadí prerelease identifikátory; prázdný (finální) je největší.
func comparePreRelease(a, b string) int {
	if a == b {
		return 0
	}
	if a == "" {
		return 1
	}
	if b == "" {
		return -1
	}
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		if c := compareIdent(as[i], bs[i]); c != 0 {
			return c
		}
	}
	return compareInt(len(as), len(bs))
}

// compareIdent porovná jeden prerelease identifikátor; číselný je menší než textový.
func compareIdent(a, b string) int {
	an, aNum := parseNumericIdent(a)
	bn, bNum := parseNumericIdent(b)
	switch {
	case aNum && bNum:
		return compareInt(an, bn)
	case aNum:
		return -1
	case bNum:
		return 1
	default:
		return strings.Compare(a, b)
	}
}

// parseNumericIdent vrátí číselnou hodnotu identifikátoru bez vedoucích nul.
func parseNumericIdent(s string) (int, bool) {
	if s == "" || !isDigits(s) {
		return 0, false
	}
	if len(s) > 1 && s[0] == '0' {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}

// validPreRelease ověří tvar prerelease části podle semver.
func validPreRelease(pre string) bool {
	if pre == "" {
		return false
	}
	for _, ident := range strings.Split(pre, ".") {
		if ident == "" {
			return false
		}
		numeric := true
		for i := 0; i < len(ident); i++ {
			c := ident[i]
			switch {
			case c >= '0' && c <= '9':
			case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '-':
				numeric = false
			default:
				return false
			}
		}
		if numeric && len(ident) > 1 && ident[0] == '0' {
			return false
		}
	}
	return true
}

// isDigits vrací true, pokud je s neprázdný a tvořený jen číslicemi.
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// isLowerHex vrací true, pokud je s tvořený jen malými hexadecimálními znaky.
func isLowerHex(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
