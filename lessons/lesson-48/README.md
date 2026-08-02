# Lekce 48 — Paměťový model a happens-before

> **Čas:** ~90 min · **Fáze:** 5 — Concurrency do hloubky · **AI režim:** `JUNIOR POD REVIEW`

## Co budeš umět

- Vysvětlit, co Go memory model garantuje a co naopak výslovně nezaručuje.
- Vyjmenovat synchronizační body a použít je jako důkaz, že je kód správně.
- Poznat klasický rozbitý vzor s `bool done` a vědět, proč není řešením `time.Sleep`.
- Vysvětlit, proč atomika není jen o nedělitelnosti zápisu, ale hlavně o viditelnosti.
- Napsat správný double-checked locking přes `sync.Once` a říct, proč je naivní verze rozbitá.

## Teorie

### Co model garantuje

[Go memory model](https://go.dev/ref/mem) říká, za jakých podmínek je zaručeno, že čtení
proměnné uvidí konkrétní zápis. Základní věta zní: pokud dvě goroutiny sahají na stejnou
proměnnou a aspoň jedna z nich zapisuje, **musí být uspořádané synchronizací**. Jinak to
není „skoro vždycky správně" — je to program s nedefinovaným chováním.

Klíčové je slovo *uspořádané*. Happens-before není o čase. Není to „stalo se dřív podle
hodinek". Je to relace, kterou vytváří jazyk a runtime, a existuje jen tam, kde ji některý
z níže uvedených mechanismů skutečně vyrobí. Dvě události mohou být „časově po sobě" a
přesto mezi nimi žádné happens-before není — a pak je viditelnost nedefinovaná.

### Synchronizační body

Tohle je celý seznam, který v praxi potřebuješ:

| Konstrukce | Co se řadí před co |
|-----------|--------------------|
| `go f()` | všechno před `go` happens-before první instrukce `f` |
| `close(ch)` | vše před `close` happens-before příjem, který zjistí zavření |
| `ch <- v` | odeslání happens-before dokončení odpovídajícího příjmu |
| `<-ch` (nebufferovaný) | příjem happens-before dokončení odeslání |
| `mu.Unlock()` | happens-before následující `mu.Lock()` |
| `once.Do(f)` | návrat `f` happens-before návrat všech ostatních `Do` |
| `wg.Done()` | happens-before návrat `wg.Wait()`, který kvůli němu skončil |
| `atomic.Store` | happens-before `atomic.Load`, který tu hodnotu přečte |

Praktický důsledek pro čtení kódu: když se ptáš „je tenhle zápis vidět?", hledáš cestu
z jednoho řádku do druhého složenou z těchhle hran. Když ji nenajdeš, je to závod — bez
ohledu na to, jak nepravděpodobně to vypadá.

### Proč to bez synchronizace nefunguje

Nejde jen o to, že dvě goroutiny „píšou přes sebe". Jsou tam dvě vrstvy přeuspořádání:

1. **Kompilátor** smí kód přeskládat, pokud se tím nemění chování *jedné* goroutiny.
   Může proměnnou držet v registru a do paměti ji zapsat až o kus dál. Nebo taky nikdy.
2. **CPU** má vlastní cache a store buffery. Zápis udělaný na jádru A se na jádře B
   objeví se zpožděním a ne nutně ve stejném pořadí jako ostatní zápisy.

Odtud klasika, na kterou naletí každý:

```go
var done bool
var result string

func worker() {
    result = "hotovo"
    done = true // ŠPATNĚ
}

func main() {
    go worker()
    for !done { // ŠPATNĚ: kompilátor smí načíst done jednou a točit navždy
    }
    fmt.Println(result) // a i kdyby se ven dostal, result může být ""
}
```

Dvě nezávislé chyby v pěti řádcích. Za prvé smyčka se nemusí nikdy ukončit, protože
kompilátor vidí, že `done` v té goroutině nikdo nemění, a načte ji do registru. Za druhé
i kdyby se ukončila, není zaručeno, že zápis do `result` bude vidět — nic tyhle dva
zápisy nesvazuje.

Nejhorší na tom je, že to na tvém stroji nejspíš „funguje". Až v produkci s jiným
zatížením, jinou verzí kompilátoru a jinou architekturou přestane. Proto je jediná
rozumná strategie pouštět souběžný kód pod `-race` a ne se spoléhat na pozorování.

### Atomika řeší viditelnost, ne jen atomicitu

Běžné vysvětlení „atomic zajistí, že se zápis nerozpůlí" je zavádějící. Zápis do `bool`
nebo `int64` je na běžných architekturách nedělitelný sám o sobě. To, co ti `sync/atomic`
přidává, je **hrana v happens-before grafu**:

```go
var done atomic.Bool
var result string

func worker() {
    result = "hotovo"
    done.Store(true) // Store se řadí před Load, který tuhle hodnotu uvidí
}

func main() {
    go worker()
    for !done.Load() {
        runtime.Gosched()
    }
    fmt.Println(result) // "hotovo" — zaručeně, nejen "nejspíš"
}
```

Ta záruka je přesně to, co dělá kód z předchozí ukázky správným. `result` není atomický a
nemusí být: je publikovaný *skrz* atomický zápis. Tenhle vzor se jmenuje **publish/subscribe
přes synchronizační bod** a používá se pořád — jen častěji přes kanál než přes atomiku.

Kanál to umí taky, a čitelněji:

```go
ready := make(chan struct{})
var result string

go func() {
    result = "hotovo"
    close(ready) // vše před close je vidět po příjmu
}()

<-ready
fmt.Println(result) // "hotovo"
```

Zavřený kanál je broadcast: takhle můžeš publikovat libovolnému počtu čtenářů najednou.

### Double-checked locking

Klasický pokus o zrychlení líné inicializace, který je v Go rozbitý:

```go
// ŠPATNĚ
var mu sync.Mutex
var conn *Conn

func Get() *Conn {
    if conn == nil {         // rychlá cesta bez zámku — a rovnou datový závod
        mu.Lock()
        defer mu.Unlock()
        if conn == nil {
            conn = dial()
        }
    }
    return conn
}
```

Ta první kontrola čte `conn` bez jakékoli synchronizace, zatímco jiná goroutina do něj
pod zámkem zapisuje. Race detektor to hlásí okamžitě. A i kdyby ne: čtenář může vidět
nenulový ukazatel dřív, než uvidí obsah struktury, na kterou ukazuje — dostal by
polovičatě inicializované spojení. To je ten druh chyby, který se ladí týden.

`sync.Once` dělá přesně tohle, jen správně — má rychlou cestu přes atomický load a
garantuje, že návrat z `f` se řadí před návrat *všech* volání `Do`:

```go
var once sync.Once
var conn *Conn

func Get() *Conn {
    once.Do(func() { conn = dial() })
    return conn // zaručeně kompletně inicializovaný
}
```

Pozor na dvě věci. `Once` neumí opakovat pokus, když inicializace selže — když
potřebuješ retry, `Once` není správný nástroj (od Go 1.21 je k dispozici
`sync.OnceValue` a `sync.OnceValues`, ale ty se chovají stejně). A `Once` se nesmí
kopírovat, stejně jako `Mutex` a `WaitGroup`.

## Rozdíly proti PHP

V PHP tenhle problém neexistuje a je dobré vědět proč. Každý request má vlastní proces
a vlastní paměť; nic se nesdílí:

```php
final class Config
{
    private static ?array $cache = null;

    public static function get(): array
    {
        // žádný zámek, žádná atomika — v tomhle procesu jsem sám
        return self::$cache ??= self::loadFromDisk();
    }
}
```

Ten samý kód v Go je datový závod, protože `Get()` může běžet z tisíce goroutin v jednom
procesu naráz:

```go
var cache map[string]string

func Get() map[string]string {
    if cache == nil {          // ZÁVOD: čtení bez synchronizace
        cache = loadFromDisk() // ZÁVOD: zápis bez synchronizace
    }
    return cache
}
```

Co se mění v uvažování: v PHP je „sdílený stav" cizí slovo — sdílí se přes Redis nebo DB,
což jsou systémy s vlastní synchronizací. V Go je sdílený stav běžná věc a **ty** jsi ten,
kdo musí dokázat, že je k němu přístup uspořádaný. Ne odhadnout. Dokázat.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `for !done {}` s obyčejným `bool` | v PHP sdílený stav mezi „vlákny" neexistuje | `atomic.Bool`, kanál nebo `WaitGroup` |
| „Zápis do bool je stejně atomický" | zaměňuje se atomicita a viditelnost | atomika není o velikosti zápisu, ale o uspořádání |
| Naivní double-checked locking | přenesený vzor z Javy/C++ | `sync.Once` |
| `time.Sleep` jako synchronizace | „dám tomu chvilku a bude to" | synchronizační bod, ne odhad času |
| `??=` reflex na lazy cache | v PHP je proces sám | `sync.Once` nebo inicializace v konstruktoru |
| Mutex jen kolem zápisu, ne kolem čtení | „čtení přece nic nerozbije" | závod potřebuje jen jeden zápis; zamykej obojí |
| Kopírování `sync.Once` / `Mutex` ve struct | struct se předává hodnotou | předávej ukazatel, `go vet` to hlásí |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 48`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `Set`, `Get`, `Set`, `Get`

```bash
make lesson L=48 PART=1
```

Pak **`/go-deep-review 48 easy`**.

### Střední

Funkce: `StressFlag`, `NewLazyInit`, `Value`, `ConcurrentValues`

```bash
make lesson L=48 PART=2
```

Pak **`/go-deep-review 48 medium`**.

### Obtížný

Funkce: `NewBox`, `Publish`, `Consume`, `PublishAndConsume`, `WaitGroupVisibility`

```bash
make lesson L=48 PART=3
```

Pak **`/go-deep-review 48 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 48 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=48` (+ `make race L=48`, pokud to lekce vyžaduje).

- [ ] Umíš vyjmenovat aspoň šest synchronizačních bodů
- [ ] Umíš vysvětlit, proč happens-before není o čase
- [ ] Umíš vysvětlit, proč `for !done {}` s obyčejným `bool` nemusí nikdy skončit
- [ ] Umíš vysvětlit, co přesně přidává `atomic.Store` nad obyčejný zápis
- [ ] Umíš vysvětlit, proč je naivní double-checked locking rozbitý

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Tohle je téma, kde jsou vygenerované odpovědi nejčastěji *skoro* správné. Když ti agent
napíše cache s dvojitou kontrolou nebo příznak přes obyčejný `bool`, nespokoj se s
vysvětlením — pusť na to `-race`. Užitečný prompt: „ukaž mi happens-before hranu, která
zaručuje, že tenhle zápis uvidí druhá goroutina." Pokud ji nedokáže pojmenovat, hrana
tam není.

## Další čtení

1. [The Go Memory Model](https://go.dev/ref/mem) — krátký a čitelný, přečti celý
2. [pkg.go.dev — sync/atomic](https://pkg.go.dev/sync/atomic)
3. [pkg.go.dev — sync.Once](https://pkg.go.dev/sync#Once)
4. [Go blog — Introducing the Go Race Detector](https://go.dev/blog/race-detector)
