# Lekce 09 — Stringy, runy a byty

> **Čas:** ~85 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Vysvětlit, proč `len("žluť")` není počet znaků, a spočítat obojí správně.
- Rozhodnout mezi `[]byte(s)` a `[]rune(s)` a vědět, co každá konverze stojí.
- Napsat operaci nad textem tak, aby nerozbila diakritiku ani emoji.
- Vysvětlit, proč se v cyklu skládá text přes `strings.Builder` a ne přes `+=`.
- Změřit rozdíl mezi oběma přístupy benchmarkem a přečíst si jeho výstup.

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

## Rozdíly proti PHP

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

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 09`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od jednodušších
ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `RuneLen` (kód je záměrně vadný — počítá bajty místo run)

```bash
make lesson L=09 PART=1
```

Pak **`/go-deep-review 09 easy`**.

### Střední

Implementuj: `ByteLen`, `ReverseRunes`

```bash
make lesson L=09 PART=2
```

Pak **`/go-deep-review 09 medium`**.

### Obtížný

Doplň: `Truncate` (limit v runách, výpustka …)

```bash
make lesson L=09 PART=3
```

Pak **`/go-deep-review 09 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 09 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=09` (+ `make race L=09`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč `len("kůň")` je 5 a ne 3
- [ ] Umíš vysvětlit, co znamená první proměnná v `for i, r := range s`
- [ ] Umíš vysvětlit, proč otočení po bajtech rozbije `ř` a otočení po runách ne
- [ ] Umíš odhadnout, kolik alokací udělá `+=` v cyklu přes 1000 prvků
- [ ] Umíš říct, kdy použít `[]byte(s)` a kdy `[]rune(s)`

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Mentor, kvíz i review (dialog) jsou vždy OK; v tomto režimu AI nesmí psát kód cvičení.
## Další čtení

1. [Go blog — Strings, bytes, runes and characters in Go](https://go.dev/blog/strings)
2. [Go blog — Text normalization in Go](https://go.dev/blog/normalization)
3. [pkg.go.dev — strings.Builder](https://pkg.go.dev/strings#Builder)
4. [pkg.go.dev — unicode/utf8](https://pkg.go.dev/unicode/utf8)
