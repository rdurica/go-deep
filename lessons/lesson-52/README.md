# Lekce 52 — Testování do hloubky: benchmarky, fuzz, golden files

> **Čas:** ~75 min · **Fáze:** 6 — Production Go · **AI režim:** `JUNIOR POD REVIEW`

## Co budeš umět

- Napsat benchmark, který měří to, co si myslíš, a přečíst z jeho výstupu ns/op, B/op a allocs/op.
- Vysvětlit, proč jedno měření nic neznamená, a zabránit tomu, aby kompilátor benchmark vyoptimalizoval.
- Navrhnout fuzz invariant, který má šanci najít chybu, a založit seed korpus v `testdata/fuzz/`.
- Rozhodnout, kdy je golden file lepší než ruční assert, a napsat k němu `-update` flag.
- Snížit počet alokací v horké cestě a doložit to testem, ne dojmem.

## Teorie

### `testing.B` a jak číst výstup

```
BenchmarkRenderReport-16          2574 ns/op    1837 B/op    23 allocs/op
BenchmarkRenderReportFast-16       716 ns/op     864 B/op     2 allocs/op
```

- `-16` je `GOMAXPROCS`, ne počet opakování.
- `ns/op` je průměrná doba jedné iterace.
- `B/op` je počet **heap** bajtů na iteraci, `allocs/op` počet alokací.

Z těch tří je `allocs/op` nejstabilnější a nejinformativnější: je deterministické, kdežto
`ns/op` závisí na tom, co zrovna dělá tvůj notebook. Proto se regresní testy píšou na
alokace, ne na čas.

Doporučený zápis od Go 1.24 je `for b.Loop()` — runtime sám řídí počet iterací
(výchozí `-benchtime=1s`) a kompilátor výsledek volání uvnitř smyčky nevyhodí jako
mrtvý kód. Když si potřebuješ počet iterací určit sám, použij `-benchtime=100x`.

```go
func BenchmarkParse(b *testing.B) {
	data := loadFixture()   // příprava se měřit nemá
	b.ReportAllocs()
	for b.Loop() {
		_, _ = Parse(data)
	}
}
```

Starší tvar `for i := 0; i < b.N; i++` pořád funguje. U něj musíš často volat
`b.ResetTimer()` po přípravě a výsledek ukládat do package-level `sink`, jinak
kompilátor smyčku optimalizuje pryč. U `b.Loop()` tohle obvykle řešit nemusíš.
`b.StopTimer`/`b.StartTimer` uvnitř cyklu zůstanou drahé — přípravu dělej předem.

### Past mrtvého kódu

```go
// ŠPATNĚ se starým b.N — výsledek se nikam nepoužije, kompilátor smí volání zahodit
for i := 0; i < b.N; i++ {
	Normalize(in)
}
```

Kompilátor vidí, že funkce nemá vedlejší efekty a její výsledek nikdo nečte. Pak měříš
prázdný cyklus a vyjde ti podezřele hezkých 0,3 ns/op. U `b.N` ulož výsledek do
**package-level** proměnné, na kterou kompilátor nevidí — nebo rovnou použij `b.Loop()`:

```go
var sink string

func BenchmarkNormalize(b *testing.B) {
	for b.Loop() {
		sink = Normalize(in) // u b.Loop() sink často není nutný; u b.N ano
	}
}
```

Lokální proměnná u `b.N` nestačí — tu by escape analýza vyhodila taky.

### `testing.AllocsPerRun` jako regresní test

Benchmark ti řekne, jak na tom jsi. Test ti pohlídá, že to tak zůstane:

```go
func TestNormalizeDoesNotAllocate(t *testing.T) {
	in := "už čistý text"
	if n := testing.AllocsPerRun(200, func() { sink = Normalize(in) }); n != 0 {
		t.Errorf("Normalize alokuje %.1f×, chci 0", n)
	}
}
```

Funkce vrátí průměrný počet alokací na jedno zavolání. Je deterministická, takže test
není flaky — ale porovnávej ji vždy s **rezervou**, ne na přesnou hodnotu, pokud netrváš
na nule.

### Jedno měření neznamená nic

Rozdíl 3 % mezi dvěma běhy je šum. Než prohlásíš optimalizaci za úspěšnou:

```bash
go test -run xxx -bench BenchmarkRender -benchmem -count=10 . > new.txt
benchstat old.txt new.txt
```

`benchstat` (z `golang.org/x/perf`) spočítá medián a rozptyl a řekne ti, jestli je rozdíl
statisticky významný. Bez něj hlásíš zlepšení, které je náhoda.

### Fuzzing

Fuzz test je table-driven test, do kterého vstupy dodává engine. Ty popíšeš **invariant**,
který má platit vždycky, a Go generuje mutace vstupu, dokud ho neporuší.

```go
func FuzzRoundTrip(f *testing.F) {
	f.Add("u-1", "Alice", 42)          // seed korpus
	f.Fuzz(func(t *testing.T, id, name string, score int) {
		in := []Record{{ID: id, Name: name, Score: score}}
		out, err := Decode(Encode(in))
		if err != nil {
			t.Fatalf("Decode(%q) = %v", Encode(in), err)
		}
		if out[0] != in[0] {
			t.Errorf("round-trip rozbil %+v", in[0])
		}
	})
}
```

Dobré invarianty jsou tři:

| Invariant | Formulace | Typicky odhalí |
|-----------|-----------|----------------|
| round-trip | `Decode(Encode(x)) == x` | chybějící escapování, ztrátu prázdných polí |
| nepanikuje | libovolný vstup dá výsledek nebo chybu | index out of range, useknuté sekvence |
| idempotence | `f(f(x)) == f(x)` | normalizace, která se sama sebou rozbije |

Špatný invariant je „výsledek se rovná tomu, co spočítá druhá implementace téhož", pokud
jsou obě tvoje a obě špatně.

```bash
go test -run xxx -fuzz=FuzzRoundTrip -fuzztime=10s .
```

Bez `-fuzz` se fuzz funkce chová jako obyčejný test a projede jen seed korpus — to je
důvod, proč můžou být v repozitáři a nezdrží CI. Seed korpus je v
`testdata/fuzz/FuzzRoundTrip/` a je verzovaný:

```
go test fuzz v1
string("id|s|rourou")
string("jmeno\\se\\zpetnym")
int(-42)
```

Když fuzzing najde protipříklad, uloží ho do stejného adresáře. **Ten soubor patří do
commitu** — je z něj regresní test zdarma.

### Golden files

Golden file je očekávaný výstup uložený v `testdata/`. Vyplatí se, když je výstup velký a
strukturovaný (report, vyrenderovaná šablona, JSON odpověď) a když je čitelný, takže diff
v code review dává smysl. Nevyplatí se na tři pole — tam napiš normální assert.

```go
var update = flag.Bool("update", false, "přepiš golden soubory")

func TestReport(t *testing.T) {
	got := RenderReport(records)
	if *update {
		os.WriteFile(goldenPath, []byte(got), 0o644)
		return
	}
	want, err := os.ReadFile(goldenPath)
	// …porovnej
}
```

`-update` je pohodlí i riziko: po jeho spuštění **musíš** diff přečíst, jinak si zafixuješ
rozbitý výstup jako správný. Proto se golden soubory nikdy negenerují v CI.

## Rozdíly proti PHP

V PHP se výkon měří skoro vždy až v produkci — Blackfire, XHProf, Tideways. Do PHPUnit se
mikrobenchmark nedává, protože běh interpretu je tak proměnlivý, že by to nic neřeklo.

```php
$t = microtime(true);
for ($i = 0; $i < 100000; $i++) { normalize($s); }
echo microtime(true) - $t;   // číslo, které příště vyjde jinak
```

V Go je měření součástí `go test`. Benchmark je funkce jako každá jiná, běží ve stejném
balíčku a runtime si sám určí, kolikrát ji spustit, aby měření dávalo smysl:

```go
func BenchmarkNormalize(b *testing.B) {
	in := "  Ahoj  Světe  "
	b.ReportAllocs()
	for b.Loop() {
		_ = Normalize(in)
	}
}
```

Co se mění v uvažování: **výkon je testovatelná vlastnost, ne dojem.** Můžeš na něj napsat
regresní test (`testing.AllocsPerRun`) úplně stejně jako na chování. A protože benchmark
běží vedle unit testů, optimalizace nikdy neproběhne bez důkazu, že se něco zlepšilo.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Benchmark bez sinku | funkce „se přece zavolá" | ulož výsledek do package-level proměnné |
| Příprava dat uvnitř měřeného cyklu | copy-paste z testu | připrav před cyklem, pak `b.ResetTimer()` |
| Optimalizace podle jednoho běhu | zvyk na `microtime()` v PHP | `-count=10` a `benchstat` |
| Test na přesný počet alokací | snaha o „přísnost" | limit s rezervou, nula jen když je nula cílem |
| Fuzz invariant „výsledek == očekávaná konstanta" | myšlení table-driven testem | invariant musí platit pro **libovolný** vstup |
| `-update` spuštěný bez přečtení diffu | golden test „opraven" | diff je součást review, ne vedlejší efekt |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 52`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od jednodušších
ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `Normalize` (kód je záměrně vadný — chybí převod na malá písmena)

```bash
make lesson L=52 PART=1
```

Pak **`/go-deep-review 52 easy`**.

### Střední

Implementuj: `Encode`, `Decode` (fuzz invariant round-trip)

```bash
make lesson L=52 PART=2
```

Pak **`/go-deep-review 52 medium`**.

### Obtížný

Doplň: `RenderTable`, `RenderTableFast` (golden file v `testdata/`)

```bash
make lesson L=52 PART=3
```

Pak **`/go-deep-review 52 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 52 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=52` (+ `make race L=52`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, kdy stačí `b.Loop()` a kdy u starého `b.N` potřebuješ package-level sink
- [ ] Umíš vysvětlit rozdíl mezi `b.Loop()`, `b.N`, `-benchtime=1s` a `-benchtime=100x`
- [ ] Umíš uvést tři invarianty, na kterých se dá postavit fuzz test
- [ ] Umíš vysvětlit, proč se soubor nalezený fuzzingem commituje
- [ ] Umíš říct, kdy je golden file lepší než assert a kdy horší

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Agenti píšou benchmarky bez sinku a fuzz testy, které jen kontrolují, že funkce vrátila
nil chybu. Nejdřív si sám napiš invariant jednou větou; teprve pak nech agenta doplnit kód.

## Další čtení

1. [go.dev — Go Fuzzing](https://go.dev/doc/security/fuzz/)
2. [Go blog — Fuzzing is Beta Ready](https://go.dev/blog/fuzz-beta)
3. [pkg.go.dev — testing.B](https://pkg.go.dev/testing#B)
4. [pkg.go.dev — golang.org/x/perf/cmd/benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
