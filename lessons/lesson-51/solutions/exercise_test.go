package solutions_test

import (
	"errors"
	"math/rand"
	"sort"
	"testing"
	"time"

	exercise "github.com/rdurica/go-deep/lessons/lesson-51/solutions"
)

func TestParseSemverValid(t *testing.T) {
	tests := []struct {
		in   string
		want exercise.Version
	}{
		{"v1.2.3", exercise.Version{Major: 1, Minor: 2, Patch: 3}},
		{"1.2.3", exercise.Version{Major: 1, Minor: 2, Patch: 3}},
		{"v0.0.0", exercise.Version{}},
		{"v10.20.30", exercise.Version{Major: 10, Minor: 20, Patch: 30}},
		{"v1.2.3-rc.1", exercise.Version{Major: 1, Minor: 2, Patch: 3, Pre: "rc.1"}},
		{"v2.0.0-alpha", exercise.Version{Major: 2, Pre: "alpha"}},
		{"v1.0.0-0.20240115120000-abcdef123456", exercise.Version{
			Major: 1, Pre: "0.20240115120000-abcdef123456",
		}},
	}
	for _, tt := range tests {
		got, err := exercise.ParseSemver(tt.in)
		if err != nil {
			t.Errorf("ParseSemver(%q) = chyba %v, chci %v", tt.in, err, tt.want)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseSemver(%q) = %+v, chci %+v", tt.in, got, tt.want)
		}
	}
}

func TestParseSemverInvalid(t *testing.T) {
	bad := []string{
		"", "v", "v1", "v1.2", "v1.2.3.4", "v1.2.x", "vx.2.3",
		"v01.2.3", "v1.02.3", "v1.2.3-", "v1.2.3-rc..1", "v1.2.3-rc.01",
		"v1.2.3-rc$1", "v-1.2.3", " v1.2.3",
	}
	for _, in := range bad {
		got, err := exercise.ParseSemver(in)
		if err == nil {
			t.Errorf("ParseSemver(%q) = %+v, chci chybu", in, got)
			continue
		}
		if !errors.Is(err, exercise.ErrSyntax) {
			t.Errorf("ParseSemver(%q) vrátil %v, chci obalený ErrSyntax", in, err)
		}
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		v    exercise.Version
		want string
	}{
		{exercise.Version{Major: 1, Minor: 2, Patch: 3}, "v1.2.3"},
		{exercise.Version{}, "v0.0.0"},
		{exercise.Version{Major: 2, Pre: "rc.1"}, "v2.0.0-rc.1"},
	}
	for _, tt := range tests {
		if got := tt.v.String(); got != tt.want {
			t.Errorf("Version%+v.String() = %q, chci %q", tt.v, got, tt.want)
		}
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b exercise.Version
		want int
	}{
		{exercise.Version{Major: 1}, exercise.Version{Major: 1}, 0},
		{exercise.Version{Major: 1}, exercise.Version{Major: 1, Patch: 1}, -1},
		{exercise.Version{Major: 1, Minor: 1}, exercise.Version{Major: 1, Patch: 9}, 1},
		{exercise.Version{Major: 2}, exercise.Version{Major: 10}, -1},
		{exercise.Version{Major: 1, Pre: "rc.1"}, exercise.Version{Major: 1}, -1},
		{exercise.Version{Major: 1}, exercise.Version{Major: 1, Pre: "rc.1"}, 1},
		{exercise.Version{Major: 1, Pre: "alpha"}, exercise.Version{Major: 1, Pre: "beta"}, -1},
		{exercise.Version{Major: 1, Pre: "rc.1"}, exercise.Version{Major: 1, Pre: "rc.2"}, -1},
		{exercise.Version{Major: 1, Pre: "rc.2"}, exercise.Version{Major: 1, Pre: "rc.10"}, -1},
		{exercise.Version{Major: 1, Pre: "rc"}, exercise.Version{Major: 1, Pre: "rc.1"}, -1},
		{exercise.Version{Major: 1, Pre: "1"}, exercise.Version{Major: 1, Pre: "alpha"}, -1},
	}
	for _, tt := range tests {
		if got := exercise.Compare(tt.a, tt.b); got != tt.want {
			t.Errorf("Compare(%s, %s) = %d, chci %d", tt.a, tt.b, got, tt.want)
		}
		if got := exercise.Compare(tt.b, tt.a); got != -tt.want {
			t.Errorf("Compare(%s, %s) = %d, chci %d (antisymetrie)", tt.b, tt.a, got, -tt.want)
		}
	}
}

// TestCompareOrders ověří, že Compare je použitelné jako řadicí funkce.
// Vstup je náhodně zamíchaný, takže test nejde splnit zadrátovanou hodnotou.
func TestCompareOrders(t *testing.T) {
	vs := []exercise.Version{
		{Minor: 9, Patch: 9},
		{Major: 1, Pre: "alpha"},
		{Major: 1, Pre: "alpha.1"},
		{Major: 1, Pre: "beta"},
		{Major: 1, Pre: "rc.1"},
		{Major: 1, Pre: "rc.2"},
		{Major: 1},
		{Major: 1, Patch: 1},
		{Major: 1, Minor: 2},
		{Major: 2},
	}
	rnd := rand.New(rand.NewSource(42))
	shuffled := make([]exercise.Version, len(vs))
	copy(shuffled, vs)
	rnd.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	sort.Slice(shuffled, func(i, j int) bool { return exercise.Compare(shuffled[i], shuffled[j]) < 0 })
	for i := range vs {
		if shuffled[i] != vs[i] {
			t.Fatalf("po seřazení je na indexu %d %s, chci %s", i, shuffled[i], vs[i])
		}
	}
}

func TestParsePseudoVersion(t *testing.T) {
	tests := []struct {
		in       string
		wantBase string
		wantTS   string
		wantRev  string
	}{
		{
			"v0.0.0-20230101120000-abcdef123456",
			"v0.0.0", "2023-01-01T12:00:00Z", "abcdef123456",
		},
		{
			"v1.2.4-0.20240115093015-0123456789ab",
			"v1.2.4", "2024-01-15T09:30:15Z", "0123456789ab",
		},
		{
			"v1.2.4-rc.1.0.20240115093015-0123456789ab",
			"v1.2.4-rc.1", "2024-01-15T09:30:15Z", "0123456789ab",
		},
	}
	for _, tt := range tests {
		base, ts, rev, err := exercise.ParsePseudoVersion(tt.in)
		if err != nil {
			t.Errorf("ParsePseudoVersion(%q) = chyba %v", tt.in, err)
			continue
		}
		if base != tt.wantBase {
			t.Errorf("ParsePseudoVersion(%q) base = %q, chci %q", tt.in, base, tt.wantBase)
		}
		if got := ts.UTC().Format(time.RFC3339); got != tt.wantTS {
			t.Errorf("ParsePseudoVersion(%q) čas = %q, chci %q", tt.in, got, tt.wantTS)
		}
		if rev != tt.wantRev {
			t.Errorf("ParsePseudoVersion(%q) revize = %q, chci %q", tt.in, rev, tt.wantRev)
		}
	}
}

func TestParsePseudoVersionInvalid(t *testing.T) {
	bad := []string{
		"",
		"v1.2.3",
		"v0.0.0-20230101120000",
		"v0.0.0-20230101120000-ABCDEF123456",
		"v0.0.0-20230101120000-abcdef12345",
		"v0.0.0-2023010112000-abcdef123456",
		"v0.0.0-20231301120000-abcdef123456",
		"v1.2.4-rc1-20240115093015-0123456789ab",
		"nesmysl-20240115093015-0123456789ab",
	}
	for _, in := range bad {
		if _, _, _, err := exercise.ParsePseudoVersion(in); err == nil {
			t.Errorf("ParsePseudoVersion(%q) = nil chyba, chci chybu", in)
		} else if !errors.Is(err, exercise.ErrSyntax) {
			t.Errorf("ParsePseudoVersion(%q) vrátil %v, chci obalený ErrSyntax", in, err)
		}
	}
}

func TestIsPseudo(t *testing.T) {
	tests := map[string]bool{
		"v0.0.0-20230101120000-abcdef123456":        true,
		"v1.2.4-0.20240115093015-0123456789ab":      true,
		"v1.2.4-rc.1.0.20240115093015-0123456789ab": true,
		"v1.2.3":         false,
		"v1.2.3-rc.1":    false,
		"":               false,
		"latest":         false,
		"v0.0.0-2023-ab": false,
	}
	for in, want := range tests {
		if got := exercise.IsPseudo(in); got != want {
			t.Errorf("IsPseudo(%q) = %v, chci %v", in, got, want)
		}
	}
}

func TestMajorSuffix(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"example.com/m", 1},
		{"example.com/m/v2", 2},
		{"example.com/m/v17", 17},
		{"github.com/rdurica/go-deep", 1},
		{"example.com/m/version", 1},
		{"example.com/m/v2x", 1},
		{"example.com/v2/sub", 1},
	}
	for _, tt := range tests {
		got, err := exercise.MajorSuffix(tt.in)
		if err != nil {
			t.Errorf("MajorSuffix(%q) = chyba %v, chci %d", tt.in, err, tt.want)
			continue
		}
		if got != tt.want {
			t.Errorf("MajorSuffix(%q) = %d, chci %d", tt.in, got, tt.want)
		}
	}
}

func TestMajorSuffixInvalid(t *testing.T) {
	bad := []string{"example.com/m/v0", "example.com/m/v1", "example.com/m/v02", "", "example.com/m/"}
	for _, in := range bad {
		if _, err := exercise.MajorSuffix(in); err == nil {
			t.Errorf("MajorSuffix(%q) = nil chyba, chci chybu", in)
		} else if !errors.Is(err, exercise.ErrMajorSuffix) {
			t.Errorf("MajorSuffix(%q) vrátil %v, chci obalený ErrMajorSuffix", in, err)
		}
	}
}

func TestSelectVersions(t *testing.T) {
	reqs := map[string][]string{
		"example.com/a":    {"v1.2.0", "v1.4.1", "v1.3.9"},
		"example.com/b":    {"v0.1.0"},
		"example.com/c/v2": {"v2.0.0-rc.1", "v2.0.0", "v2.0.0-rc.9"},
		"example.com/d":    {"v1.0.0", "v1.0.0"},
	}
	want := map[string]string{
		"example.com/a":    "v1.4.1",
		"example.com/b":    "v0.1.0",
		"example.com/c/v2": "v2.0.0",
		"example.com/d":    "v1.0.0",
	}
	got, err := exercise.SelectVersions(reqs)
	if err != nil {
		t.Fatalf("SelectVersions = chyba %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("SelectVersions vrátil %d modulů, chci %d", len(got), len(want))
	}
	for path, w := range want {
		if got[path] != w {
			t.Errorf("SelectVersions[%q] = %q, chci %q", path, got[path], w)
		}
	}
}

func TestSelectVersionsEmptyInput(t *testing.T) {
	got, err := exercise.SelectVersions(map[string][]string{})
	if err != nil {
		t.Fatalf("SelectVersions(prázdná mapa) = chyba %v", err)
	}
	if got == nil {
		t.Fatal("SelectVersions(prázdná mapa) = nil mapa, chci prázdnou nenilovou")
	}
	if len(got) != 0 {
		t.Errorf("SelectVersions(prázdná mapa) má %d položek, chci 0", len(got))
	}
}

func TestSelectVersionsErrors(t *testing.T) {
	t.Run("module without versions", func(t *testing.T) {
		_, err := exercise.SelectVersions(map[string][]string{"example.com/a": {}})
		if !errors.Is(err, exercise.ErrNoVersions) {
			t.Errorf("chci ErrNoVersions, dostal jsem %v", err)
		}
	})
	t.Run("broken version", func(t *testing.T) {
		_, err := exercise.SelectVersions(map[string][]string{"example.com/a": {"v1.0.0", "nesmysl"}})
		if !errors.Is(err, exercise.ErrSyntax) {
			t.Errorf("chci ErrSyntax, dostal jsem %v", err)
		}
	})
}

func TestCheckCompat(t *testing.T) {
	ok := []struct{ path, version string }{
		{"example.com/m", "v0.4.0"},
		{"example.com/m", "v1.9.3"},
		{"example.com/m", "v1.0.0-rc.1"},
		{"example.com/m/v2", "v2.0.0"},
		{"example.com/m/v3", "v3.1.4"},
	}
	for _, tt := range ok {
		if err := exercise.CheckCompat(tt.path, tt.version); err != nil {
			t.Errorf("CheckCompat(%q, %q) = %v, chci nil", tt.path, tt.version, err)
		}
	}

	bad := []struct{ path, version string }{
		{"example.com/m", "v2.0.0"},
		{"example.com/m/v2", "v1.0.0"},
		{"example.com/m/v2", "v3.0.0"},
		{"example.com/m/v2", "v0.1.0"},
	}
	for _, tt := range bad {
		err := exercise.CheckCompat(tt.path, tt.version)
		if err == nil {
			t.Errorf("CheckCompat(%q, %q) = nil, chci chybu", tt.path, tt.version)
			continue
		}
		if !errors.Is(err, exercise.ErrIncompatible) {
			t.Errorf("CheckCompat(%q, %q) vrátil %v, chci obalený ErrIncompatible", tt.path, tt.version, err)
		}
	}
}

func TestCheckCompatBrokenInput(t *testing.T) {
	if err := exercise.CheckCompat("example.com/m", "nesmysl"); !errors.Is(err, exercise.ErrSyntax) {
		t.Errorf("chci ErrSyntax, dostal jsem %v", err)
	}
	if err := exercise.CheckCompat("example.com/m/v1", "v1.0.0"); !errors.Is(err, exercise.ErrMajorSuffix) {
		t.Errorf("chci ErrMajorSuffix, dostal jsem %v", err)
	}
}
