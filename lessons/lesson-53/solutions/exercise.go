// Package solutions obsahuje referenční řešení lekce 53.
package solutions

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	httppprof "net/http/pprof"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"unicode"
)

// wordRE hledá slova jako souvislé úseky písmen a číslic.
var wordRE = regexp.MustCompile(`[\p{L}\p{N}]+`)

// SumDigitsSlow sečte desítkové číslice v řetězci. Referenční, ale pomalá varianta:
// každou runu převede na string a nechá ji rozparsovat strconv.
func SumDigitsSlow(s string) int {
	sum := 0
	for _, r := range s {
		n, err := strconv.Atoi(string(r))
		if err != nil {
			continue
		}
		sum += n
	}
	return sum
}

// CountWordsSlow spočítá výskyty slov. Referenční, ale pomalá varianta:
// lowercase celého textu a regulární výraz v horké cestě.
func CountWordsSlow(text string) map[string]int {
	counts := map[string]int{}
	for _, w := range wordRE.FindAllString(strings.ToLower(text), -1) {
		counts[w]++
	}
	return counts
}

// IsWordRune vrací true pro znaky, které patří do slova.
func IsWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

// SumDigits sečte desítkové číslice v řetězci bez jediné alokace.
func SumDigits(s string) int {
	sum := 0
	for i := 0; i < len(s); i++ {
		if c := s[i]; c >= '0' && c <= '9' {
			sum += int(c - '0')
		}
	}
	return sum
}

// CountWords spočítá výskyty slov (case-insensitive) s minimem alokací.
func CountWords(text string) map[string]int {
	// Odhad počtu slov: průměrné české slovo má kolem osmi bajtů.
	counts := make(map[string]int, len(text)/8+1)
	start := -1
	for i, r := range text {
		if IsWordRune(r) {
			if start < 0 {
				start = i
			}
			continue
		}
		if start >= 0 {
			countWord(counts, text[start:i])
			start = -1
		}
	}
	if start >= 0 {
		countWord(counts, text[start:])
	}
	return counts
}

// JoinIDs spojí čísla čárkou do jednoho řetězce.
func JoinIDs(ids []int) string {
	if len(ids) == 0 {
		return ""
	}
	var b strings.Builder
	// Nejdelší int64 má 20 znaků, plus oddělovač.
	b.Grow(len(ids) * 8)
	var scratch [24]byte
	for i, id := range ids {
		if i > 0 {
			b.WriteByte(',')
		}
		b.Write(strconv.AppendInt(scratch[:0], int64(id), 10))
	}
	return b.String()
}

// CaptureCPUProfile spustí f a zapíše CPU profil jejího běhu do w.
func CaptureCPUProfile(w io.Writer, f func()) error {
	if w == nil {
		return errors.New("cpu profil: chybí writer")
	}
	if f == nil {
		return errors.New("cpu profil: chybí funkce k profilování")
	}
	if err := pprof.StartCPUProfile(w); err != nil {
		return fmt.Errorf("start cpu profilu: %w", err)
	}
	defer pprof.StopCPUProfile()
	f()
	return nil
}

// CaptureHeapProfile zapíše aktuální heap profil do w.
func CaptureHeapProfile(w io.Writer) error {
	if w == nil {
		return errors.New("heap profil: chybí writer")
	}
	// Bez GC by profil obsahoval i objekty, které už nikdo nedrží.
	runtime.GC()
	p := pprof.Lookup("heap")
	if p == nil {
		return errors.New("heap profil: profil není registrovaný")
	}
	if err := p.WriteTo(w, 0); err != nil {
		return fmt.Errorf("zápis heap profilu: %w", err)
	}
	return nil
}

// PprofHandler vrátí handler s pprof endpointy na vlastním muxu.
func PprofHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", httppprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", httppprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", httppprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", httppprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", httppprof.Trace)
	return mux
}

// countWord započítá jedno slovo; malá písmena řeší až tehdy, když je potřeba.
func countWord(counts map[string]int, w string) {
	if hasUpper(w) {
		w = strings.ToLower(w)
	}
	counts[w]++
}

// hasUpper vrací true, pokud slovo obsahuje aspoň jedno velké písmeno.
func hasUpper(w string) bool {
	for i := 0; i < len(w); i++ {
		c := w[i]
		if c >= 'A' && c <= 'Z' {
			return true
		}
		if c >= utf8Self {
			// Vícebajtová runa: rozhodne až plný unicode test.
			for _, r := range w[i:] {
				if unicode.IsUpper(r) {
					return true
				}
			}
			return false
		}
	}
	return false
}

// utf8Self je hranice, nad kterou už bajt není samostatná ASCII runa.
const utf8Self = 0x80
