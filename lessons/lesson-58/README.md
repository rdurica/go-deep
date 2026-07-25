# Lekce 58 — Osobní checklist, pairing protokol a manual rewrite

> **Čas:** ~90 min · **Fáze:** 7 — Inženýrství v době AI · **AI režim:** `TECH LEAD`

## Co budeš umět

- Sestavit si vlastní review checklist ze základu z playbooku a ze svých doložených slabin.
- Vést pairing session s agentem podle protokolu — vědět, kdo v každé fázi drží klávesnici a proč.
- Používat manual rewrite drill jako nástroj retence a poznat, kdy je čas AI úplně vypnout.
- Měřit vlastní review přes precision a recall a z čísel odvodit, které lekce zopakovat.

## PHP → Go most

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

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Opsaný checklist z internetu | vypadá kompletně | základ + položky doložené vlastními nálezy |
| Checklist o 40 položkách | „ať tam nic nechybí" | maximálně 12, jinak ho začneš odklikávat |
| Skok ze spec rovnou do implementace | reflex „mám málo času" | protokol povoluje jen krok vpřed nebo zpět |
| Předání bez důvodu | důvod je „přece zřejmý" | důvod je log rozhodnutí, bez něj přechod neplatí |
| `time.Now()` uvnitř testované logiky | v PHP to řešil Clock z frameworku | hodiny jako `func() time.Time` v konstruktoru |
| Měření kvality pocitem | „myslím, že se lepším" | precision a recall z review drilu, kategorie → lekce |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

Stavíš tři nástroje své vlastní praxe: slučování checklistů, model pairing session
a vyhodnocení review drilu.

### A — rozcvička (~10 min)

1. `Severity.String()` → `"INFO"`, `"WARN"`, `"ERROR"`.
2. `MergeChecklists(base, personal []CheckItem) []CheckItem`:
   - výsledek zachovává pořadí `base`,
   - položka z `personal` se stejným `ID` **nahradí** základní na jejím původním místě
     (text i závažnost),
   - položky z `personal`, které v `base` nejsou, se přidají na konec v jejich pořadí,
   - duplicita uvnitř `base`: platí první výskyt,
   - položka s prázdným `ID` se zahodí,
   - prázdné vstupy dávají prázdný výsledek (ne paniku).

např. `MergeChecklists(base, personal)` → `ctx-first` přepsané, `rows-err` na konci

### B — jádro (~35 min)

Model pairing session. `NewSession(now func() time.Time) *Session`; `nil` znamená
`time.Now`. Časy událostí ber vždy z těchto hodin.

- `Start(r Role) error` — povolené role jsou `RoleSpec`, `RoleTests`, `RoleImpl`,
  `RoleReview` (`RoleNone`, `RoleDone` a hodnoty mimo rozsah dávají `ErrInvalidRole`).
  Druhé volání vrací `ErrAlreadyStarted`. Zapíše událost `RoleNone → r` s důvodem
  `"start"`.
- `Handoff(to Role, reason string) error` — povolený je jen posun o **jeden krok** vpřed
  nebo zpět v řadě `spec → tests → impl → review`. Cokoli jiného (včetně přechodu na
  `RoleDone`, na `RoleNone` a na sebe sama) je `ErrInvalidTransition`. Prázdný nebo jen
  bílými znaky tvořený důvod je `ErrMissingReason`. Před `Start` je to `ErrNotStarted`,
  po `Finish` `ErrFinished`. Neplatný pokus **nesmí** změnit stav ani zapsat událost.
- `Finish() error` — jen z `RoleReview`, jinak `ErrInvalidTransition`. Zapíše událost
  `RoleReview → RoleDone` s důvodem `"hotovo"`. Podruhé `ErrFinished`.
- `Current() Role` a `Timeline() []Event` — `Timeline` vrací **kopii**, aby volající
  nemohl přepsat historii.
- `Role.String()` → `"none"`, `"spec"`, `"tests"`, `"impl"`, `"review"`, `"done"`;
  hodnota mimo rozsah `"none"`.

např. `Start(spec)` → … → `Finish()` → `Current() = done`

### C — rozšíření (~25 min)

1. `RecommendLesson(category string) string` — case-insensitive mapa kategorie na
   doporučení. Pokrytá kategorie: `errors`, `context`, `concurrency`, `http`, `design`,
   `testing`; každá má jiný text. Neznámá kategorie dostane výchozí doporučení
   (neprázdné).
2. `ScoreReview(found, planted []Finding) Score` — vyhodnocení review drilu. Nálezy se
   párují podle `ID`, duplicitní `ID` se počítá jednou, prázdné `ID` se ignoruje.
   - `TruePositives` — nalezené, které byly nastražené,
   - `FalsePositives` — nalezené, které nastražené nebyly,
   - `Missed` — nastražené, které jsi nenašel,
   - `Precision` = TP / (TP + FP), při nule označených `0`,
   - `Recall` = TP / počet nastražených, při nule nastražených `1`,
   - `Review` — doporučení pro kategorie **zmeškaných** nálezů, bez duplicit, seřazená
     abecedně.

např. `ScoreReview` (TP=2, FP=1, Missed=2) → `Precision=2/3`, `Recall=0.5`

```bash
make lesson L=58
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `58`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=58` prochází
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
