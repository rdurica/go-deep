# Lekce 09 — Stringy, runy a byty

> **Čas:** ~85 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Vysvětlit, proč `len("žluť")` není počet znaků, a spočítat obojí správně.
- Rozhodnout mezi `[]byte(s)` a `[]rune(s)` a vědět, co každá konverze stojí.
- Napsat operaci nad textem tak, aby nerozbila diakritiku ani emoji.
- Vysvětlit, proč se v cyklu skládá text přes `strings.Builder` a ne přes `+=`.
- Změřit rozdíl mezi oběma přístupy benchmarkem a přečíst si jeho výstup.

## PHP → Go most

V PHP je `$s[0]` jeden bajt, `strlen()` počítá bajty a `mb_strlen()` znaky — a když
zapomeneš na `mb_`, rozbiješ češtinu. Tenhle dvojí svět znáš.

```php
$s = "žluť";
strlen($s);        // 6 bajtů
mb_strlen($s);     // 4 znaky
strrev($s);        // rozbité — otáčí bajty
mb_substr($s, 0, 2);
```

Go dvojí sadu funkcí nemá. Má jeden typ `string`, který je **vždycky posloupnost bajtů**,
a jazykové konstrukce, které ti dovolí dívat se na něj jako na znaky, když chceš.

```go
s := "žluť"
len(s)                       // 6 — bajty, vždycky bajty
utf8.RuneCountInString(s)    // 4 — znaky (runy)
s[0]                         // 197 (byte), ne "ž"
for i, r := range s { }      // i je bajtový offset, r je runa
[]rune(s)[0]                 // 'ž'
```

Přenos návyku: přestaň hledat `mb_` variantu. Místo toho se u každé operace zeptej,
jestli pracuješ s bajty (I/O, protokoly, hashe) nebo se znaky (uživatelský text, délka
pro zobrazení, otáčení). Go tě k té otázce donutí tím, že `len` má jen jeden význam.

## Teorie

### String je neměnná sekvence bajtů

`string` je v Go dvojice `{ptr, len}` — ukazatel na bajty a jejich počet. Bajty jsou
**neměnné**:

```go
s := "ahoj"
s[0] = 'A'      // chyba kompilace: cannot assign to s[0]
s = "Ahoj"      // OK — proměnná ukazuje na jiné bajty, staré se nemění
```

Neměnnost je důvod, proč je předání stringu do funkce levné (kopíruje se 16 bajtů
headeru, ne data) a proč může být string klíč mapy. Zároveň znamená, že **každá úprava
textu alokuje nový string**.

Go nikde negarantuje, že string obsahuje platné UTF-8 — může v něm být cokoliv, třeba
binární data z disku. Zdrojový kód ale UTF-8 je, takže literály platné jsou.

Indexace vrací `byte` (což je alias pro `uint8`), ne znak:

```go
s := "žluť"
fmt.Println(len(s))    // 6
fmt.Println(s[0])      // 197
fmt.Printf("%c\n", s[0]) // Å — půlka znaku ž, nesmysl
fmt.Println(s[0:2])    // ž — první dva bajty tvoří jeden znak
```

### UTF-8 v pěti řádcích

UTF-8 kóduje každý kódový bod (runu) do 1 až 4 bajtů:

| Rozsah | Bajtů | Příklad |
|--------|-------|---------|
| ASCII (`U+0000`–`U+007F`) | 1 | `a`, `9`, `,` |
| Latinka s diakritikou, řečtina, cyrilice | 2 | `ž`, `á`, `ř` |
| Většina ostatních písem, `…`, `€` | 3 | `…`, `€` |
| Emoji, historická písma | 4 | `🐹` |

Klíčová vlastnost: ASCII znaky mají v UTF-8 stejné bajty jako v ASCII, a žádný
následující bajt vícebajtové sekvence nevypadá jako ASCII. Proto `strings.Split(s, ",")`
funguje správně i pro český text — čárku nelze potkat uprostřed znaku.

Typ `rune` je alias pro `int32` a drží jeden kódový bod. Literál v jednoduchých uvozovkách
je runa, ne string:

```go
var r rune = 'ž'
fmt.Println(r)          // 382 — číselná hodnota kódového bodu
fmt.Printf("%c %U\n", r, r) // ž U+017E
fmt.Println(len(string(r))) // 2 — jako UTF-8 zabere 2 bajty
```

### `range` po stringu jde po runách

Tohle je jediné místo v jazyce, kde `range` dělá něco jiného, než by člověk čekal:

```go
for i, r := range "žluť" {
	fmt.Printf("i=%d r=%c\n", i, r)
}
// i=0 r=ž
// i=2 r=l
// i=3 r=u
// i=4 r=ť
```

`i` **není** pořadí runy, ale její **bajtový offset**. Skáče o 1 až 4. Proto se `range`
po stringu nikdy nepoužívá k počítání pozic ve smyslu „třetí znak".

Naproti tomu indexový cyklus jede po bajtech:

```go
for i := 0; i < len(s); i++ {
	// s[i] je byte, u vícebajtového znaku dostaneš kus
}
```

Obojí je legitimní — první pro text, druhý pro bajtové zpracování. Musíš jen vědět,
který právě píšeš.

Když narazí `range` na neplatný UTF-8 bajt, vrátí `utf8.RuneError` (`U+FFFD`) a posune se
o jeden bajt. Nespadne to, ale data jsou tichá ztráta.

### `[]byte(s)`, `[]rune(s)` a jejich cena

Obě konverze **alokují a kopírují** (kompilátor umí pár speciálních případů zoptimalizovat,
ale nespoléhej na to):

```go
b := []byte(s)   // kopie bajtů, len(b) == len(s)
r := []rune(s)   // dekóduje UTF-8, len(r) == počet run, alokuje 4 B na runu
back := string(r)
```

`[]byte` chceš, když text posíláš do `io.Writer`, hashuješ, nebo ho po bajtech měníš.
`[]rune` chceš, když potřebuješ **náhodný přístup ke znakům** — otočení, výběr n-tého
znaku, ořez na počet znaků. Za pohodlí platíš alokací velkou 4× počet run.

Když ti stačí projít text jednou zleva doprava, `range` je zdarma a `[]rune` je plýtvání.

Naopak **výřez stringu `s[a:b]` nekopíruje**. Vrátí nový string header ukazující do
stejných bajtů — je to operace za konstantní čas, přesně jako u slice. Protože jsou
stringy neměnné, není to problém pro korektnost, ale je to past na paměť: když si
z desetimegabajtového souboru uložíš `s[0:10]`, drží ti ten kousek celých deset megabajtů
naživu, dokud ho nezkopíruješ přes `strings.Clone` nebo `string([]byte(s[0:10]))`.

A pozor na dvojí význam konverze čísla na string:

```go
string(rune(65))       // "A" — kódový bod
strconv.Itoa(65)       // "65" — desítkový zápis
string(65)             // od Go 1.15 hlásí vet varování, dělá totéž co první řádek
```

Počet run bez alokace spočítá `utf8.RuneCountInString`:

```go
utf8.RuneCountInString("žluť")  // 4
len([]rune("žluť"))             // taky 4, ale s alokací
```

### Proč reverzování po bajtech rozbije češtinu

Vezmi větu, kterou zná každý český vývojář:

```go
s := "příliš žluťoučký kůň"

// ŠPATNĚ — otáčí bajty
b := []byte(s)
for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
	b[i], b[j] = b[j], b[i]
}
fmt.Println(string(b)) // rozsypaný čaj: ��k��uo��ul� ��il��rp
```

Dvoubajtová sekvence `ř` = `0xC5 0x99` se po otočení stane `0x99 0xC5`, což není platný
znak. Dekodér vrátí `U+FFFD` a text je nečitelný.

```go
// SPRÁVNĚ — otáčí runy
r := []rune(s)
for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
	r[i], r[j] = r[j], r[i]
}
fmt.Println(string(r)) // ňůk ýkčuoťulž šilířp
```

Pro úplnost: ani runová varianta není „správné otočení textu" v typografickém smyslu.
Emoji složená z několika kódových bodů (vlajky, rodiny, modifikátory pleti) se rozpadnou,
protože jedna vizuální *grapheme cluster* je víc run. Stdlib na to nemá nástroj —
řeší to balíček `golang.org/x/text`. Pro tenhle kurz stačí runy a vědomí, kde je hranice.

### `strings.Builder` vs `+=` v cyklu

String je neměnný, takže `a + b` musí alokovat nový string a zkopírovat do něj obě
strany. V cyklu je to kvadratická práce:

```go
// ŠPATNĚ — n alokací, celkem O(n²) zkopírovaných bajtů
out := ""
for _, p := range parts {
	out += p
}
```

`strings.Builder` drží rostoucí `[]byte` a `String()` z něj vyrobí string **bez kopie**
(využívá toho, že buffer už nikdo jiný nedrží):

```go
var sb strings.Builder
sb.Grow(totalLen)          // jedna alokace předem, když délku znáš
for _, p := range parts {
	sb.WriteString(p)
}
out := sb.String()
```

Zero value `strings.Builder` je rovnou použitelná, konstruktor není. `Grow(n)` je
volitelný, ale když umíš délku spočítat, ušetříš i to postupné zvětšování bufferu.
Metody `WriteString`, `WriteByte`, `WriteRune` a `Write` vracejí chybu jen kvůli
splnění `io.Writer` — u Builderu je vždycky `nil` a ignoruje se.

Rozdíl uvidíš v části C na vlastním benchmarku. Pro pár desítek krátkých kousků je to
jedno; pro tisíce je to rozdíl mezi milisekundou a sekundou.

### Co si najít v `strings` a `strconv`

Nemá smysl si to pamatovat, ale musíš vědět, že to existuje:

| Potřebuji | Funkce |
|-----------|--------|
| Obsahuje / začíná / končí | `strings.Contains`, `HasPrefix`, `HasSuffix` |
| Rozdělit a spojit | `strings.Split`, `strings.Fields`, `strings.Join` |
| Ořezat bílé znaky nebo znaky | `strings.TrimSpace`, `Trim`, `TrimPrefix` |
| Nahradit | `strings.ReplaceAll`, `strings.NewReplacer` |
| Velikost písmen | `strings.ToUpper`, `ToLower`, `unicode.ToUpper` |
| Porovnat bez ohledu na velikost | `strings.EqualFold` |
| Číslo ↔ text | `strconv.Itoa`, `Atoi`, `ParseFloat`, `FormatFloat`, `Quote` |

`strings.Fields` rozdělí podle libovolného počtu bílých znaků a prázdné kusy zahodí —
přesně to, co chceš na uživatelský vstup s dvojitými mezerami.

Za pozornost stojí `strings.EqualFold`:

```go
strings.ToLower(a) == strings.ToLower(b)  // dvě alokace
strings.EqualFold(a, b)                   // bez alokace, Unicode case-folding
```

`EqualFold` navíc řeší i případy, kde prosté `ToLower` selhává (například německé `ß`).
Pro porovnávání identifikátorů, hlaviček a příkazů je to správná volba. Pozor: **není
to** správný nástroj na porovnávání jmen nebo hesel v lidském smyslu — na to je potřeba
normalizace, kterou stdlib nemá.

A `strconv` je vždycky lepší než `fmt.Sprintf("%d", n)`: je rychlejší, nealokuje
interface a jasně říká, co dělá.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `len(s)` jako počet znaků | v PHP je `strlen` na ASCII totéž | `utf8.RuneCountInString(s)` |
| `s[i]` jako znak | zvyk na `$s[$i]` | `[]rune(s)[i]` nebo `range` |
| Otočení / ořez po bajtech | zapomenutá `mb_` varianta | pracuj s `[]rune` |
| `out += x` v cyklu | v PHP je konkatenace levná | `strings.Builder` + `Grow` |
| `strings.ToLower(a) == strings.ToLower(b)` | zvyk na `strtolower` | `strings.EqualFold(a, b)` |
| `i` z `range` jako index znaku | vypadá jako běžný `range` | `i` je bajtový offset |
| `fmt.Sprintf("%d", n)` pro převod | univerzální nástroj na všechno | `strconv.Itoa(n)` |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

### A — rozcvička (~10 min)

1. `ByteLen(s string) int` — počet bajtů.
2. `RuneLen(s string) int` — počet run. Použij `utf8.RuneCountInString`, ne `[]rune`.

Pro `"go"` obojí vrátí `2`, pro `"kůň"` `5` a `3` — `ů` i `ň` zaberou po dvou bajtech.

### B — jádro (~35 min)

1. `ReverseRunes(s string) string` — otočí pořadí run. Musí správně zvládnout češtinu
   (`"kůň"` → `"ňůk"`) i emoji (`"a🐹b"` → `"b🐹a"`). Prázdný vstup vrací prázdný string.
2. `Truncate(s string, maxRunes int) string` — zkrátí text tak, aby měl **nejvýš
   `maxRunes` run včetně** připojeného znaku `…` (U+2026, jedna runa):
   - `maxRunes <= 0` → prázdný string,
   - text kratší nebo přesně dlouhý `maxRunes` run → vrací se beze změny, **bez** `…`,
   - jinak prvních `maxRunes-1` run plus `…`.

   Takže `Truncate("příliš", 4)` je `"pří…"` a `Truncate("příliš", 6)` je `"příliš"`.
3. `Initials(fullName string) string` — iniciály z celého jména: první runa každého slova,
   převedená na velké písmeno. `"Radek Ďurica"` → `"RĎ"`. Ošetři vícenásobné mezery,
   tabulátory a okrajové mezery (`"  jan   novák "` → `"JN"`) i prázdný vstup (`""`).
   Hodí se `strings.Fields` a `unicode.ToUpper`.

### C — rozšíření (~25 min)

1. `Join(parts []string, sep string) string` — spojí kusy oddělovačem. **Nesmíš použít
   `strings.Join`.** Postav to na `strings.Builder` a **předalokuj** přes `Grow` na
   přesnou výslednou délku v bajtech, kterou si spočítáš dopředu. Prázdný vstup vrací
   `""`, jeden prvek vrací ten prvek bez oddělovače.

   Test kontroluje počet alokací přes `testing.AllocsPerRun` — bez `Grow` neprojde.
2. `CountRunes(s string) map[rune]int` — spočítá výskyty jednotlivých run. Vždy vrací
   ne-nil mapu. Nesmíš alokovat `[]rune` — projdi text přes `range`.

V testu jsou navíc připravené benchmarky `BenchmarkBuilder` a `BenchmarkConcat`.
Až budeš hotový, spusť je a podívej se na rozdíl:

```bash
cd exercise && go test -bench=. -benchmem -run=^$
```

Na 1000 kouscích to vypadá zhruba takhle:

```
BenchmarkBuilder-16      168592          7269 ns/op       12288 B/op          1 allocs/op
BenchmarkConcat-16          968       1288358 ns/op    12504270 B/op       1998 allocs/op
```

Sto sedmdesátkrát pomalejší a tisíckrát víc naalokované paměti — protože `+=` pokaždé
zkopíruje celý dosavadní výsledek. Benchmarky se při běžném `go test` nespouštějí, takže
ti nic nerozbijí.

```bash
make lesson L=09
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `09`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=09` prochází
- [ ] Umíš vysvětlit, proč `len("kůň")` je 5 a ne 3
- [ ] Umíš vysvětlit, co znamená první proměnná v `for i, r := range s`
- [ ] Umíš vysvětlit, proč otočení po bajtech rozbije `ř` a otočení po runách ne
- [ ] Umíš odhadnout, kolik alokací udělá `+=` v cyklu přes 1000 prvků
- [ ] Umíš říct, kdy použít `[]byte(s)` a kdy `[]rune(s)`

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

## Další čtení

1. [Go blog — Strings, bytes, runes and characters in Go](https://go.dev/blog/strings)
2. [Go blog — Text normalization in Go](https://go.dev/blog/normalization)
3. [pkg.go.dev — strings.Builder](https://pkg.go.dev/strings#Builder)
4. [pkg.go.dev — unicode/utf8](https://pkg.go.dev/unicode/utf8)
