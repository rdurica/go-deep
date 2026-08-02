# Lekce 13 — Interfaces II: io.Reader, io.Writer, kompozice

> **Čas:** ~85 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Napsat funkci, která bere `io.Writer` místo `*os.File`, a otestovat ji bez jediného souboru.
- Odrecitovat kontrakt `Read` včetně případu `n > 0` a `err == io.EOF` a nešlápnout do něj.
- Poskládat dekorátory (`io.LimitReader`, `io.TeeReader`, vlastní) do řetězu.
- Vybrat mezi `bufio.Scanner`, `io.ReadAll` a `io.Copy` a vědět, kde má Scanner limit.

## Teorie

### Dva interfacy, na kterých stojí celá stdlib

```go
type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}
```

To je vše. Jedna metoda, žádný stav, žádné `open`/`close`. A přesto je splňuje
`*os.File`, `net.Conn`, `*bytes.Buffer`, `*strings.Reader`, `http.ResponseWriter`,
`gzip.Writer`, `*sql.Rows` (skoro) a stovky typů mimo stdlib.

Právě proto je v Go tak levné psát univerzální kód. Funkce, která bere `io.Reader`,
umí zpracovat soubor, síťové spojení, HTTP tělo i řetězec v testu, aniž bys pro to
cokoli udělal.

Když navrhuješ vlastní API, ptej se: *nedá se to vyjádřit jako Reader nebo Writer?*
Když ano, dostaneš zdarma `io.Copy`, `io.ReadAll`, `bufio`, `gzip`, `httptest`
a všechno ostatní.

### Kontrakt `Read` — čti pozorně

`Read(p []byte) (n int, err error)` naplní `p` a vrátí, kolik bajtů dodal. Pravidla,
která překvapí každého nováčka:

1. `Read` smí vrátit **méně** bajtů, než je `len(p)`, i když nejde o konec dat.
   Nikdy nepředpokládej, že jedno volání přečte všechno.
2. Implementace smí vrátit `n > 0` **a zároveň** `err == io.EOF`. Zpracuj tedy nejdřív
   data a až pak koukej na chybu.
3. `n == 0, err == nil` je legální, ale znamená „nic se nestalo" — volající to nesmí
   brát jako konec.
4. Konec dat se hlásí `io.EOF`. To **není chyba** ve smyslu selhání, je to normální
   ukončení. Jakákoli jiná chyba je skutečný problém.

```go
// SPRÁVNĚ — data se zpracují i v posledním volání
for {
	n, err := r.Read(buf)
	process(buf[:n])   // nejdřív data
	if err == io.EOF { // pak konec
		break
	}
	if err != nil {
		return err
	}
}

// ŠPATNĚ — ztratí poslední kus dat u readerů, které vracejí (n>0, io.EOF)
for {
	n, err := r.Read(buf)
	if err != nil {
		break
	}
	process(buf[:n])
}
```

Dobrá zpráva: tenhle cyklus ručně skoro nikdy psát nebudeš. Napsat ho správně je práce
pro `io.Copy`, `io.ReadAll` nebo `bufio.Scanner`.

### Nástroje, které to za tebe udělají

```go
data, err := io.ReadAll(r)     // vše do paměti — jen když víš, že se to vejde
n, err := io.Copy(dst, src)    // stream bez mezipaměti, vrací počet bajtů
io.WriteString(w, "text")      // bez zbytečné konverze na []byte
```

`io.Copy` je chytrý: když `dst` umí `ReadFrom` nebo `src` umí `WriteTo`, použije je
a vyhne se kopírování přes buffer. Proto `io.Copy(os.Stdout, file)` u velkých souborů
nezabírá paměť.

Pro čtení po řádcích je `bufio.Scanner`:

```go
sc := bufio.NewScanner(r)
for sc.Scan() {
	line := sc.Text() // bez koncového \n a \r
}
if err := sc.Err(); err != nil { // io.EOF se sem NEDOSTANE, to není chyba
	return err
}
```

Pozor na dvě věci. Za prvé, `sc.Err()` musíš zkontrolovat — jinak nepoznáš rozdíl mezi
„konec souboru" a „síť spadla". Za druhé, **Scanner má výchozí limit 64 KiB na řádek**
(`bufio.MaxScanTokenSize`). Delší řádek ukončí smyčku a `sc.Err()` vrátí
`bufio.ErrTooLong`. Když je delší řádek možný, zvyš limit:

```go
sc.Buffer(make([]byte, 0, 64*1024), 1<<20) // až 1 MiB na řádek
```

Tohle je klasická produkční chyba: kód roky funguje, pak přijde jeden dlouhý JSON řádek
v logu a tiše se ztratí zbytek souboru.

### Zdroje a jímky v paměti

```go
var buf bytes.Buffer          // Writer i Reader, zero value použitelná
fmt.Fprintf(&buf, "%d", 42)
fmt.Println(buf.String())

r := strings.NewReader("obsah") // Reader nad řetězcem, umí i Seek
```

Tyhle dva typy jsou důvod, proč testy v téhle lekci nesahají na disk. Místo dočasného
souboru dáš do funkce `strings.NewReader(...)` a výsledek si přečteš z `bytes.Buffer`.
Rychlé, deterministické, bez úklidu.

### Dekorátory: Reader, který obaluje Reader

Protože je `io.Reader` jen jedna metoda, dá se triviálně obalit. Stdlib to dělá pořád:

```go
io.LimitReader(r, 1024)   // přečte nejvýš 1024 bajtů, pak hlásí EOF
io.TeeReader(r, log)      // co se přečte z r, zapíše se i do log
io.MultiReader(a, b, c)   // tři zdroje za sebou jako jeden
bufio.NewReader(r)        // bufferování
gzip.NewReader(r)         // dekomprese
```

Vlastní dekorátor je struct s vnořeným Readerem a jednou metodou:

```go
type upperReader struct{ r io.Reader }

func (u upperReader) Read(p []byte) (int, error) {
	n, err := u.r.Read(p)
	for i := 0; i < n; i++ {
		if c := p[i]; c >= 'a' && c <= 'z' {
			p[i] = c - ('a' - 'A')
		}
	}
	return n, err // n i err se propagují beze změny
}
```

Klíčové detaily: transformuj jen `p[:n]` (zbytek bufferu je cizí paměť) a chybu i počet
předej dál. Konstruktor vrací `io.Reader`, ne `upperReader` — typ může zůstat neexportovaný.
Tohle je ta jedna situace, kdy se „return structs" nedodržuje, protože typ nemá žádné
další užitečné metody.

`io.LimitReader(NewUpperReader(f), 100)` pak funguje samo od sebe. Řetězení dekorátorů
je Go ekvivalent Symfony middleware — bez konfigurace, jen skládáním funkcí.

### Kompozice interfaců

Interface se dá poskládat z jiných:

```go
type ReadWriter interface {
	Reader
	Writer
}
```

`io.ReadWriter`, `io.ReadCloser`, `io.ReadWriteCloser`, `io.WriteSeeker` — celá tabulka
kombinací je v balíčku `io` hotová. Ber jako parametr **nejmenší** kombinaci, kterou
skutečně potřebuješ: bereš-li `io.ReadCloser`, ale nikdy nezavíráš, ber `io.Reader`
a nech zavírání volajícímu.

Kompozice funguje i jako runtime dotaz. Type assertion na širší interface ti řekne,
jestli konkrétní hodnota umí něco navíc:

```go
if s, ok := r.(io.Seeker); ok {
	s.Seek(0, io.SeekStart) // umí přetočit
}
```

## Rozdíly proti PHP

V PHP je „něco, do čeho se dá psát" typicky `resource` z `fopen()` nebo objekt s vlastním
API — a testovat to znamená sáhnout po `vfsStream`, po `php://memory` nebo po dočasném souboru:

```php
function writeReport(string $path, array $lines): void
{
    $fh = fopen($path, 'w');          // funkce si sama otevře soubor
    foreach ($lines as $i => $line) {
        fwrite($fh, sprintf("%d. %s\n", $i + 1, $line));
    }
    fclose($fh);
}
```

Symfony to řeší přes `OutputInterface` v Console komponentě — správný instinkt, ale platí
jen uvnitř té komponenty. Go má jeden takový interface pro **celý ekosystém**:

```go
func WriteReport(w io.Writer, lines []string) error {
	for i, line := range lines {
		if _, err := fmt.Fprintf(w, "%d. %s\n", i+1, line); err != nil {
			return err
		}
	}
	return nil
}
```

Volající rozhodne, kam se píše: soubor, HTTP odpověď, gzip stream, `bytes.Buffer` v testu.
Změna v uvažování: **funkce nemá vlastnit zdroj**. Otevírání a zavírání patří o patro výš.
Je to dependency injection bez kontejneru, konfigurace a `services.yaml` — jen parametr.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Funkce bere cestu k souboru a otevírá si ho | reflex `fopen()` uvnitř funkce | ber `io.Reader`/`io.Writer`, otevírej o patro výš |
| `if err != nil { break }` před zpracováním dat | předpoklad, že EOF přijde s `n == 0` | zpracuj `buf[:n]`, pak řeš chybu |
| Chybějící `sc.Err()` po smyčce | Scanner vypadá, že chyby řeší sám | vždy zkontroluj `sc.Err()` |
| Řádek nad 64 KiB tiše useknutý | neznámý výchozí limit Scanneru | `sc.Buffer(...)` s vlastním maximem |
| `io.ReadAll` na HTTP tělo bez limitu | zvyk na `file_get_contents()` | `io.LimitReader` nebo `http.MaxBytesReader` |
| Test přes dočasný soubor | zvyk na `vfsStream` / tmp soubory | `strings.NewReader` a `bytes.Buffer` |
| Transformace celého `p` místo `p[:n]` | přehlédnuté `n` | pracuj jen s tím, co Read skutečně dodal |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 13`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `WriteReport`, `CountLines`

```bash
make lesson L=13 PART=1
```

Pak **`/go-deep-review 13 easy`**.

### Střední

Funkce: `NewUpperReader`, `Tail`, `Write`

```bash
make lesson L=13 PART=2
```

Pak **`/go-deep-review 13 medium`**.

### Obtížný

Funkce: `Bytes`, `Lines`, `Pipeline`

```bash
make lesson L=13 PART=3
```

Pak **`/go-deep-review 13 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 13 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=13` (+ `make race L=13`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč `Read` smí vrátit `n > 0` spolu s `io.EOF`
- [ ] Umíš vysvětlit, proč funkce bere `io.Writer` a ne `*os.File`
- [ ] Umíš říct, kdy použít `io.Copy`, kdy `io.ReadAll` a kdy `bufio.Scanner`
- [ ] Umíš popsat, co se stane s řádkem delším než 64 KiB
- [ ] Umíš napsat vlastní dekorátor nad `io.Reader` z hlavy

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Mentor, kvíz i review (dialog) jsou vždy OK; v tomto režimu AI nesmí psát kód cvičení.
## Další čtení

1. [pkg.go.dev — io](https://pkg.go.dev/io)
2. [pkg.go.dev — bufio.Scanner](https://pkg.go.dev/bufio#Scanner)
3. [Go blog — Errors are values](https://go.dev/blog/errors-are-values) — příklad `errWriter`
4. [Go Code Review Comments — Interfaces](https://go.dev/wiki/CodeReviewComments#interfaces)
