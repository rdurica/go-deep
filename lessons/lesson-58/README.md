# Lekce 58 — Osobní checklist, pairing protokol a manual rewrite

> **Čas:** ~90 min · **Fáze:** 7 — Inženýrství v době AI · **AI režim:** `TECH LEAD`

## Co budeš umět

- Sestavit si vlastní review checklist ze základu z playbooku a ze svých doložených slabin.
- Vést pairing session s agentem podle protokolu — vědět, kdo v každé fázi drží klávesnici a proč.
- Používat manual rewrite drill jako nástroj retence a poznat, kdy je čas AI úplně vypnout.
- Měřit vlastní review přes precision a recall a z čísel odvodit, které lekce zopakovat.

## Teorie

### Základ plus osobní vrstva

Checklist má dvě části. Základ je z `docs/ai-playbook.md` a je stejný pro každého
(interfacy u konzumenta, žádné `_ = err`, wrapping `%w`, context první a ne ve structu,
lifetime goroutin, žádný panický control-flow). Osobní vrstva vzniká z dat: z checkpointů,
z nálezů, které ti našel někdo jiný, a z produkčních incidentů.

Tři pravidla, aby checklist přežil měsíc:

1. **Položka má ID**, ne jen text. ID přežije přeformulování a jde podle něj měřit.
2. **Osobní verze přepisuje základní**, nepřidává se vedle. Dvě podobné položky znamenají,
   že jednu z nich přeskočíš.
3. **Maximálně 12 položek.** Delší seznam nikdo neprojde poctivě; místo toho se ho naučíš
   odklikávat.

Přesně tuhle logiku slučování implementuješ v úkolu A — a je to nástroj, který budeš
používat i po kurzu.

### Pairing protokol s agentem

Bez protokolu se pairing s agentem zvrtne do „on píše, já přijímám". Protokol říká, kdo
v které fázi vlastní výstup:

| Fáze | Vlastní | Agent smí |
|------|---------|-----------|
| `spec` | ty | ptát se na chybějící kritéria, navrhovat hraniční případy |
| `tests` | ty | doplnit tabulku dalšími případy, nikdy neměnit očekávání |
| `impl` | agent | psát kód, dokud testy neprojdou |
| `review` | ty | vysvětlit svoje rozhodnutí, navrhnout opravu konkrétního nálezu |
| `done` | ty | nic |

Přechody jsou jednosměrné, ale s návratem o krok zpět: `spec → tests → impl → review`,
a když review najde díru, `review → impl` nebo dál `impl → tests`. Skok
`spec → impl` je zakázaný záměrně — je to přesně ta zkratka, po které vzniká kód bez
akceptačního kritéria.

Každé předání má **důvod**. Věta „vracím do testů, protože chybí případ s `utm_`
parametry" je zároveň log rozhodnutí a materiál pro retrospektivu. Session s časovou osou
je model, který v úkolu B napíšeš — všimni si, že hodiny se předávají jako závislost:

```go
s := NewSession(func() time.Time { return fixed })  // v testu
s := NewSession(nil)                                // v produkci = time.Now
```

To je obecný Go idiom: **čas je závislost jako každá jiná.** Test, který volá `time.Now()`
uvnitř kódu, je test, který jednou v noci spadne.

### Manual rewrite drill

Retence znalosti nevzniká čtením diffů. Vzniká produkcí. Drill:

1. Vyber funkci z kódu, který ti agent napsal a který jsi schválil (60–100 řádků).
2. Zavři IDE, agenta i dokumentaci. Otevři prázdný soubor a napiš ji znovu z hlavy.
3. Pusť testy. Porovnej svoji verzi s původní.
4. Zapiš si, co ti nešlo: signatura? `errors.Is` vs `errors.As`? `defer` v cyklu?
   `sync.WaitGroup` a `Add` před `go`?

Co ti nešlo, je přesně obsah tvojí osobní vrstvy checklistu. Drill dělej jednou týdně na
kritické cestě — na doméně a na souběžnosti, ne na wiringu.

Doplněk je **spaced repetition na syntaxi**: krátká sada otázek („jaká je zero value
mapy?", „co dělá `append` u sdíleného backing array?", „kdy se vyhodnotí argumenty
`defer`?"), kterou projdeš za pět minut po 1, 3, 7 a 21 dnech. Syntaxe je jediná část Go,
kterou má smysl drilovat mechanicky; všechno ostatní je úsudek.

### Kniha vzorů a vypnutá AI

Veď si soubor s vlastními vzory — ne s teorií, ale s hotovými kusy kódu, které jsi napsal
a kterým rozumíš: worker pool s `errgroup`, middleware chain, `Option` funkce, table test
s `t.Run`, graceful shutdown. Když to potřebuješ, kopíruješ ze svého, ne z modelu. Vzor,
který jsi jednou napsal, umíš i obhájit.

Kdy AI vypnout úplně:

- učíš se nové API stdlib (jinak se naučíš, že „to za mě někdo napíše"),
- řešíš souběžnost s netriviálním lifetime,
- píšeš doménový model — hranice si musíš promyslet ty,
- honíš heisenbug: model ti s radostí vygeneruje pravděpodobné vysvětlení, které je špatně.

### Měření: precision a recall

Review drill se dá vyhodnotit číslem. Do kódu se nastraží známé chyby, ty najdeš, co
najdeš, a spočítá se:

```
precision = nalezené správně / všechno, co jsi označil
recall    = nalezené správně / všechno, co tam bylo
```

Nízký **recall** znamená, že čteš moc rychle nebo bez checklistu. Nízká **precision**
znamená, že označuješ dojmy místo nálezů — a tím ztrácíš důvěru toho, komu review píšeš.
Kategorie zmeškaných nálezů ti navíc přímo řeknou, co zopakovat: tři přehlédnuté chyby
v kategorii `concurrency` znamenají lekce 40–47, ne „snažit se víc".

## Rozdíly proti PHP

V PHP týmu se kvalita drží nástroji: PHPStan level, ECS, Rector, coverage gate. Jsou to
sdílená pravidla, která platí pro všechny stejně, a tvoje osobní slabiny v nich nejsou —
protože je nikdo nezná.

```yaml
# phpstan.neon — pravidla týmu, ne tvoje
parameters:
    level: 8
    paths: [src, tests]
```

V Go je sdíleného minima málo (`gofmt`, `go vet`) a zbytek stojí na úsudku. To vypadá jako
nevýhoda, ale je to příležitost: checklist si můžeš zúžit na to, co **ty** děláš špatně.

```go
// osobní položka checklistu vypadá takhle konkrétně:
// [ ] ctx-first: v minulých třech PR jsem dvakrát nechal context až za id
```

Změna v uvažování: kvalita v Go není nastavená konfigurací, je to **návyk s pamětí**.
Checklist, který nevychází z tvých vlastních chyb, je jen opsaný seznam z internetu.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Opsaný checklist z internetu | vypadá kompletně | základ + položky doložené vlastními nálezy |
| Checklist o 40 položkách | „ať tam nic nechybí" | maximálně 12, jinak ho začneš odklikávat |
| Skok ze spec rovnou do implementace | reflex „mám málo času" | protokol povoluje jen krok vpřed nebo zpět |
| Předání bez důvodu | důvod je „přece zřejmý" | důvod je log rozhodnutí, bez něj přechod neplatí |
| `time.Now()` uvnitř testované logiky | v PHP to řešil Clock z frameworku | hodiny jako `func() time.Time` v konstruktoru |
| Měření kvality pocitem | „myslím, že se lepším" | precision a recall z review drilu, kategorie → lekce |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 58`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `String`, `MergeChecklists`, `String`

```bash
make lesson L=58 PART=1
```

Pak **`/go-deep-review 58 easy`**.

### Střední

Funkce: `NewSession`, `Current`, `Start`, `Handoff`

```bash
make lesson L=58 PART=2
```

Pak **`/go-deep-review 58 medium`**.

### Obtížný

Funkce: `Finish`, `Timeline`, `RecommendLesson`, `ScoreReview`

```bash
make lesson L=58 PART=3
```

Pak **`/go-deep-review 58 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 58 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=58` (+ `make race L=58`, pokud to lekce vyžaduje).

- [ ] Máš vlastní checklist, kde aspoň tři položky vychází z tvých doložených chyb
- [ ] Umíš vysvětlit, proč je přechod `spec → impl` zakázaný
- [ ] Umíš popsat manual rewrite drill a řekneš, na jaké funkci ho uděláš tenhle týden
- [ ] Umíš vysvětlit rozdíl mezi nízkou precision a nízkým recall a co s každou z nich
- [ ] Umíš jmenovat tři situace, ve kterých AI vypneš úplně

## AI režim

`TECH LEAD` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Protokol z téhle lekce je závazný pro zbytek kurzu včetně capstone: spec a testy píšeš ty,
implementaci smí psát agent, review vlastníš ty a každé předání má důvod.

## Další čtení

1. [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
2. [Go blog — Testable Examples in Go](https://go.dev/blog/examples)
3. [Effective Go — Interfaces and methods](https://go.dev/doc/effective_go#interfaces_and_types)
