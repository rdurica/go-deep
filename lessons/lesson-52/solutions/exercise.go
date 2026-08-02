// Package solutions obsahuje referenční řešení lekce 52.
package solutions

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ErrFormat označuje vstup, který neodpovídá formátu záznamů.
var ErrFormat = errors.New("neplatný formát záznamu")

// Record je jeden záznam textového formátu.
type Record struct {
	ID    string
	Name  string
	Score int
}

// --- Stupeň: jednoduchý ---
// Normalize ořízne okrajové bílé znaky, sjednotí vnitřní bílé znaky
// na jednu mezeru a převede text na malá písmena.
func Normalize(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	pendingSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if b.Len() > 0 {
				pendingSpace = true
			}
			continue
		}
		if pendingSpace {
			b.WriteByte(' ')
			pendingSpace = false
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// --- Stupeň: střední ---
// Encode zapíše záznamy do textového formátu "id|name|score" po řádcích.
// Znaky '\\', '|', '\n' a '\r' se escapují zpětným lomítkem.
func Encode(recs []Record) string {
	var b strings.Builder
	for i, r := range recs {
		if i > 0 {
			b.WriteByte('\n')
		}
		writeEscaped(&b, r.ID)
		b.WriteByte('|')
		writeEscaped(&b, r.Name)
		b.WriteByte('|')
		b.WriteString(strconv.Itoa(r.Score))
	}
	return b.String()
}

func writeEscaped(b *strings.Builder, s string) {
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\\':
			b.WriteString(`\\`)
		case '|':
			b.WriteString(`\|`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteByte(c)
		}
	}
}

// Decode přečte formát vyrobený funkcí Encode. Přijímá jen kanonický tvar,
// takže Encode(Decode(s)) == s pro každý vstup, který projde.
func Decode(s string) ([]Record, error) {
	if s == "" {
		return nil, nil
	}
	lines := strings.Split(s, "\n")
	out := make([]Record, 0, len(lines))
	for i, line := range lines {
		rec, err := decodeLine(line)
		if err != nil {
			return nil, fmt.Errorf("řádek %d: %w", i+1, err)
		}
		out = append(out, rec)
	}
	return out, nil
}

func decodeLine(line string) (Record, error) {
	var fields [3]string
	var b strings.Builder
	field := 0

	for i := 0; i < len(line); i++ {
		switch c := line[i]; c {
		case '\\':
			i++
			if i >= len(line) {
				return Record{}, fmt.Errorf("useknutá escape sekvence: %w", ErrFormat)
			}
			switch line[i] {
			case '\\':
				b.WriteByte('\\')
			case '|':
				b.WriteByte('|')
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			default:
				return Record{}, fmt.Errorf("neznámá escape sekvence \\%s: %w", string(line[i]), ErrFormat)
			}
		case '|':
			if field == 2 {
				return Record{}, fmt.Errorf("víc než tři pole: %w", ErrFormat)
			}
			fields[field] = b.String()
			b.Reset()
			field++
		case '\r':
			return Record{}, fmt.Errorf("neescapovaný znak CR: %w", ErrFormat)
		default:
			b.WriteByte(c)
		}
	}
	if field != 2 {
		return Record{}, fmt.Errorf("%d pole místo tří: %w", field+1, ErrFormat)
	}
	fields[2] = b.String()

	score, err := strconv.Atoi(fields[2])
	if err != nil || strconv.Itoa(score) != fields[2] {
		return Record{}, fmt.Errorf("skóre %q není kanonické celé číslo: %w", fields[2], ErrFormat)
	}
	return Record{ID: fields[0], Name: fields[1], Score: score}, nil
}

// --- Stupeň: obtížný ---
// RenderTable vykreslí záznamy jako zarovnanou textovou tabulku.
// Prázdný vstup dá prázdný řetězec.
func RenderTable(recs []Record) string {
	if len(recs) == 0 {
		return ""
	}
	w := columnWidths(recs)

	out := renderRow(w, "ID", "NAME", "SCORE")
	out += renderRule(w)
	for _, r := range recs {
		out += renderRow(w, r.ID, r.Name, strconv.Itoa(r.Score))
	}
	return out
}

// RenderTableFast vrací totéž co RenderTable, ale staví výstup
// v jediném bufferu s předalokací.
func RenderTableFast(recs []Record) string {
	if len(recs) == 0 {
		return ""
	}
	w := columnWidths(recs)
	lineLen := w[0] + w[1] + w[2] + 5 // dva oddělovače po dvou mezerách + '\n'

	var b strings.Builder
	b.Grow(lineLen * (len(recs) + 2))

	writeCell(&b, "ID", w[0], false)
	b.WriteString("  ")
	writeCell(&b, "NAME", w[1], false)
	b.WriteString("  ")
	writeCell(&b, "SCORE", w[2], true)
	b.WriteByte('\n')

	writeRepeat(&b, '-', w[0])
	b.WriteString("  ")
	writeRepeat(&b, '-', w[1])
	b.WriteString("  ")
	writeRepeat(&b, '-', w[2])
	b.WriteByte('\n')

	for _, r := range recs {
		writeCell(&b, r.ID, w[0], false)
		b.WriteString("  ")
		writeCell(&b, r.Name, w[1], false)
		b.WriteString("  ")
		writeCell(&b, strconv.Itoa(r.Score), w[2], true)
		b.WriteByte('\n')
	}
	return b.String()
}

func columnWidths(recs []Record) [3]int {
	w := [3]int{2, 4, 5} // délky hlaviček ID, NAME, SCORE
	for _, r := range recs {
		w[0] = max(w[0], utf8.RuneCountInString(r.ID))
		w[1] = max(w[1], utf8.RuneCountInString(r.Name))
		w[2] = max(w[2], len(strconv.Itoa(r.Score)))
	}
	return w
}

func renderRow(w [3]int, a, b, c string) string {
	return padRight(a, w[0]) + "  " + padRight(b, w[1]) + "  " + padLeft(c, w[2]) + "\n"
}

func renderRule(w [3]int) string {
	return strings.Repeat("-", w[0]) + "  " + strings.Repeat("-", w[1]) + "  " + strings.Repeat("-", w[2]) + "\n"
}

func padRight(s string, w int) string {
	if n := w - utf8.RuneCountInString(s); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

func padLeft(s string, w int) string {
	if n := w - utf8.RuneCountInString(s); n > 0 {
		return strings.Repeat(" ", n) + s
	}
	return s
}

func writeCell(b *strings.Builder, s string, w int, alignRight bool) {
	pad := w - utf8.RuneCountInString(s)
	if pad < 0 {
		pad = 0
	}
	if alignRight {
		writeRepeat(b, ' ', pad)
		b.WriteString(s)
		return
	}
	b.WriteString(s)
	writeRepeat(b, ' ', pad)
}

func writeRepeat(b *strings.Builder, c byte, n int) {
	for i := 0; i < n; i++ {
		b.WriteByte(c)
	}
}
