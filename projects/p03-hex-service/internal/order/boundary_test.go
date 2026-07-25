package order_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"strconv"
	"strings"
	"testing"
)

// TestHraniceBalicku hlídá směr závislostí strojově. Slovní dohoda „doména
// nezná HTTP ani databázi" vydrží do prvního spěchu; tenhle test i po něm.
func TestHraniceBalicku(t *testing.T) {
	zakazane := map[string]string{
		"net/http":      "transport",
		"encoding/json": "serializace na hranici",
		"database/sql":  "persistence",
		"os":            "prostředí procesu",
	}
	// Testy balíčku klidně net/http importovat mohou; hranici porušuje až
	// produkční kód.
	bezTestu := func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}

	// Doména i aplikační vrstva jsou vnitřek hexagonu. Adaptéry (httpapi,
	// memstore) a wiring (cmd) se schválně nekontrolují — ty transport znát mají.
	for _, dir := range []string{".", "../app"} {
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, dir, bezTestu, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parsování %q selhalo: %v", dir, err)
		}

		// Bez počítadla by test prošel i tehdy, kdyby ParseDir nenašel
		// vůbec nic. Falešně zelený test je horší než žádný.
		souboru := 0
		for _, pkg := range pkgs {
			for name, file := range pkg.Files {
				souboru++
				for _, spec := range file.Imports {
					path, err := strconv.Unquote(spec.Path.Value)
					if err != nil {
						t.Fatalf("%s: nečitelný import %s", name, spec.Path.Value)
					}
					if duvod, zakazany := zakazane[path]; zakazany {
						t.Errorf("%s importuje %q (%s) — vnitřek hexagonu to znát nesmí",
							name, path, duvod)
					}
				}
			}
		}
		if souboru == 0 {
			t.Errorf("v %q se neprošel žádný soubor", dir)
		}
	}
}
