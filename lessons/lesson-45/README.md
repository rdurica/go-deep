# Lekce 45 — Pipelines

> **Čas:** ~90 min · **Fáze:** 5 — Concurrency do hloubky · **AI režim:** `JUNIOR POD REVIEW`

## Co budeš umět

- Složit pipeline ze stupňů propojených kanály a u každého říct, kdo zavírá jeho výstup.
- Propagovat zrušení celou pipeline tak, aby po předčasném konci nezůstala goroutina.
- Přidat do prostředního stupně fan-out a zase to slít fan-inem.
- Napsat obecný stupeň přes typové parametry a rozhodnout, kdy se to vyplatí.
- Poznat případ, kdy pipeline **nepsat**, a podložit to benchmarkem.

## Teorie

### Anatomie stupně

Stupeň je funkce, která dostane vstupní kanál a vrátí výstupní. Uvnitř si vytvoří
výstup, spustí goroutinu a hned se vrátí:

```go
func Square(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)      // stupeň zavírá svůj výstup, nikdy svůj vstup
        for v := range in {   // range skončí, až předchozí stupeň zavře svůj výstup
            select {
            case out <- v * v:
            case <-ctx.Done(): // cesta ven, i když nikdo neodebírá
                return
            }
        }
    }()
    return out
}
```

Tenhle tvar má tři vlastnosti, které tvoří celou lekci:

1. **Vlastnictví.** Stupeň vlastní jen svůj výstup, ten vytváří i zavírá. Vstup patří
   předchozímu stupni a zavřít ho nesmí (ostatně `<-chan int` mu to ani nedovolí).
2. **Ukončení zdola.** Když předchozí stupeň zavře výstup, `range` skončí, `defer`
   zavře další kanál a konec se propaguje dopředu jako domino.
3. **Ukončení shora.** Když konzument odejde, žádné domino nespadne — pomáhá jedině
   `ctx.Done()` ve větvi `select`.

### Backpressure zadarmo

Nebufferované kanály dělají z pipeline samoregulační systém. Rychlý producent se
zablokuje na zápisu, dokud si pomalý konzument nepřevezme hodnotu, takže rozpracovaných
dat je vždy jen tolik, kolik je stupňů. To je zásadní rozdíl proti frontě s neomezeným
bufferem, kde se práce hromadí, dokud nedojde paměť.

Buffer ve stupni je proto ladicí knoflík, ne výchozí stav. Malý buffer (jednotky až
desítky) vyhladí nárazovost, když je práce nerovnoměrná. Velký buffer říká „nevím, kdo
je pomalý, tak to zatím někam schovám".

### Fan-out a fan-in

Když je jeden stupeň výrazně dražší než ostatní, pustíš ho ve víc kopiích. Všechny čtou
ze **stejného** vstupního kanálu, takže si práci rozeberou samy:

```go
workers := make([]<-chan Result, 4)
for i := range workers {
    workers[i] = enrich(ctx, normalized) // čtyři goroutiny nad jedním vstupem
}
merged := FanIn(ctx, workers...)
```

`FanIn` je opak: jedna goroutina na každý vstup, `WaitGroup` a **jedno** zavření výstupu:

```go
func FanIn[T any](ctx context.Context, chs ...<-chan T) <-chan T {
    out := make(chan T)
    var wg sync.WaitGroup
    wg.Add(len(chs))
    for _, ch := range chs {
        go func() {
            defer wg.Done()
            for v := range ch {
                select {
                case out <- v:
                case <-ctx.Done():
                    return
                }
            }
        }()
    }
    go func() {
        wg.Wait()
        close(out) // právě jednou, až když už nikdo neposílá
    }()
    return out
}
```

Cena za fan-out je ztráta pořadí. Když ho potřebuješ zachovat, musíš s hodnotou nést
index a na konci ji seřadit — nebo fan-out nedělat.

### Zrušení a nejčastější chyba

Chyba, kterou v pipeline dělá skoro každý, vypadá takhle:

```go
for v := range Square(ctx, Gen(ctx, nums...)) {
    if v > 100 {
        break // konzument odchází...
    }
}
// ...a obě goroutiny zůstávají viset na zápisu do kanálu, který nikdo nečte
```

`break` ukončí smyčku, ne pipeline. Stupně visí na `out <- v` navždy. V testu si toho
nevšimneš, v HTTP handleru přibude leak na každý požadavek, který skončil dřív.

Řešení má dvě části a obě jsou nutné. Stupně musí mít v `select` větev `ctx.Done()` (viz
`Square` výše) **a** konzument musí kontext zrušit, když končí:

```go
ctx, cancel := context.WithCancel(ctx)
defer cancel() // ať odejdeš jakkoli, pipeline dostane signál
```

`defer cancel()` je to, co z toho dělá spolehlivý vzor: funguje i při `return` z půlky
funkce, i při panice. Historicky se místo kontextu používal kanál `done <-chan struct{}`
a v Go blogu o pipelines ho ještě uvidíš; `context.Context` dělá totéž, jen se navíc
propaguje přes API hranice a nese deadline.

### Generické stupně

Od Go 1.18 se dá stupeň napsat jednou pro všechny typy:

```go
func Stage[T, U any](ctx context.Context, in <-chan T, f func(T) U) <-chan U {
    out := make(chan U)
    go func() {
        defer close(out)
        for v := range in {
            select {
            case out <- f(v):
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

names := Stage(ctx, ids, func(id int) string { return lookup(id) })
```

Typové parametry se odvodí z argumentů, takže volání zůstane čitelné. Kde generický
stupeň nepomůže: když stupeň potřebuje vlastní chybové chování, dávkování nebo vnitřní
stav. Pak je poctivější napsat ho ručně — obecnost, kterou nikdo nevyužije, je jen
další vrstva.

### Chyby v pipeline

Kanály nesou jeden typ, takže chybu nemáš kam vrátit. Zavedený způsob je nést ji
s hodnotou:

```go
type Result struct {
    Input string
    Value string
    Err   error
}
```

Následující stupně chybný `Result` jen propustí dál (`if r.Err != nil { return r }`).
Konzument na konci má kompletní obrázek a rozhodne se, jestli chybu zaloguje, spočítá,
nebo pipeline zruší. Alternativa — samostatný kanál chyb — vypadá čistěji, ale rozbíjí
párování chyby se vstupem a přidává další kanál, který musí někdo zavřít.

### Kdy pipeline nepsat

Pipeline není zadarmo. Každý prvek projde několika kanály, každý přenos je synchronizace
a přepnutí goroutiny. Pro data v paměti a levnou transformaci je obyčejná smyčka řádově
rychlejší a čitelnější:

```go
// pipeline: ~3 kanály na prvek
for v := range Square(ctx, Gen(ctx, nums...)) { sum += v }

// smyčka: nic
for _, n := range nums { sum += n * n }
```

V `exercise/` jsou k tomu dva benchmarky, pusť si je:

```bash
cd lessons/lesson-45/exercise && go test -bench=Square -benchmem .
```

Rozdíl bývá dva až tři řády. Pipeline se vyplatí, když je práce ve stupni **dostatečně
drahá** (I/O, volání API, dekomprese, výpočet) nebo když skutečně potřebuješ streamovat
data, která se nevejdou do paměti. Pro `[]int` v paměti je to jen složitější smyčka.

## Rozdíly proti PHP

Zpracování dat po krocích v PHP typicky vypadá jako řetěz volání nad polem nebo jako
generátor s `yield`:

```php
function normalize(iterable $rows): Generator {
    foreach ($rows as $row) {
        yield trim($row);
    }
}

foreach (format(enrich(normalize($rows))) as $out) {
    echo $out;
}
```

Vypadá to skoro stejně jako Go pipeline a jeden zásadní rozdíl to skrývá: PHP generátor
je **líný a jednovláknový**. Nic neběží, dokud si o hodnotu neřekneš, a všechny kroky
běží v jediném vlákně, jeden po druhém.

```go
out := format(ctx, enrich(ctx, normalize(ctx, in)))
for res := range out {
    fmt.Println(res)
}
```

V Go každý stupeň **už běží** — je to samostatná goroutina, která pracuje, zatímco ty
čteš z konce. Tři důsledky, na které v PHP nemusíš myslet: stupně běží skutečně současně
(a mají tedy smysl u I/O), musíš řešit, kdo kanály zavírá, a když přestaneš odebírat,
stupně **nezmizí** — zůstanou viset. PHP generátor po `break` prostě zapomeneš a GC ho
uklidí. Goroutinu zablokovanou na zápisu do kanálu neuklidí nikdo.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `break` z konzumenta bez `cancel()` | v PHP generátor po `break` zmizí | `ctx, cancel := context.WithCancel(ctx)` a `defer cancel()` |
| Stupeň zavírá svůj vstup | „už jsem ho dočetl" | zavírá se jen vlastní výstup; vstup ber jako `<-chan T` |
| `close(out)` z každé goroutiny fan-inu | každá si myslí, že je poslední | `wg.Wait()` a jedno `close` v samostatné goroutině |
| Holé `out <- v` ve stupni | šťastná cesta vypadá jednodušeji | `select` s `ctx.Done()` |
| Velký buffer „pro plynulost" | reflex z fronty | nebufferovaný kanál dává backpressure; buffer jen s důvodem |
| Pipeline nad malým slicem | vypadá to idiomaticky | změř to; pro levnou transformaci piš smyčku |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 45`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `Gen`

```bash
make lesson L=45 PART=1
```

Pak **`/go-deep-review 45 easy`**.

### Střední

Funkce: `Square`

```bash
make lesson L=45 PART=2
```

Pak **`/go-deep-review 45 medium`**.

### Obtížný

Funkce: `Pipeline`

```bash
make lesson L=45 PART=3
```

Pak **`/go-deep-review 45 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 45 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=45` (+ `make race L=45`, pokud to lekce vyžaduje).

- [ ] Umíš u každého kanálu v pipeline říct, kdo ho vytváří a kdo zavírá
- [ ] Umíš vysvětlit, proč `break` v konzumentovi bez `cancel()` vyrobí leak
- [ ] Umíš popsat, proč `FanIn` potřebuje `WaitGroup` a samostatnou goroutinu na `close`
- [ ] Umíš říct, co se fan-outem ztrácí a jak to případně vrátit
- [ ] Umíš z benchmarku obhájit, kdy pipeline **nepsat**

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Nech si vygenerovat `Pipeline` a projdi ji jako reviewer. Typické nálezy: chybějící
`ctx.Done()` ve stupni, `close` na vstupním kanálu, fan-in bez `WaitGroup`, chybový kanál
navíc, který nikdo nezavírá, a buffery zvolené od oka. A poslední otázka na sebe: nebyla
by tady obyčejná smyčka lepší?

## Další čtení

1. [Go blog — Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
2. [Go blog — Go Concurrency Patterns (Rob Pike)](https://go.dev/blog/io2013-talk-concurrency)
3. [pkg.go.dev — context](https://pkg.go.dev/context)
4. [Go blog — An Introduction To Generics](https://go.dev/blog/intro-generics)
