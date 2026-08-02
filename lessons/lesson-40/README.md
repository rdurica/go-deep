# Lekce 40 — Goroutiny, WaitGroup a leaky

> **Čas:** ~90 min · **Fáze:** 5 — Concurrency do hloubky · **AI režim:** `JUNIOR POD REVIEW`

## Co budeš umět

- Vysvětlit, proč goroutina není vlákno a proč si můžeš dovolit spustit jich sto tisíc.
- Použít `sync.WaitGroup` správně — včetně toho, proč `Add` nikdy nepatří dovnitř goroutiny.
- Rozpoznat dva nejčastější způsoby, jak vzniká goroutine leak, a opravit je.
- Změřit počet živých goroutin a napsat test, který leak odhalí bez holého `time.Sleep`.
- Odpovědět u každého `go` na otázku „jak tahle goroutina skončí?".

## Teorie

### `go` a co se vlastně stane

Klíčové slovo `go` před voláním funkce znamená „spusť tohle jako samostatnou goroutinu
a hned pokračuj dál". Argumenty se **vyhodnotí okamžitě** v aktuální goroutině, ale tělo
běží jinde.

```go
go doWork(computeArg()) // computeArg() běží TEĎ, doWork jinde a později
```

Goroutina není OS vlákno. Startuje s ~2 KB zásobníku, který roste podle potřeby, a
runtime jich multiplexuje tisíce na hrstku vláken. Proto je běžné mít 100 000 goroutin
a nesmyslné mít 100 000 vláken. Přepnutí mezi goroutinami dělá scheduler v user space,
je řádově levnější než přepnutí kontextu v jádře.

Z toho plyne první past: **když skončí `main`, program skončí okamžitě** a nikdo nečeká
na rozběhnuté goroutiny.

```go
func main() {
    go fmt.Println("ahoj")
    // program končí, výpis se nejspíš nikdy neobjeví
}
```

Není to závod, který „většinou vyhraješ" — je to chybějící synchronizace.
`time.Sleep(100 * time.Millisecond)` na konci `main` to zdánlivě spraví a je to ten
nejhorší možný kód: funguje na notebooku a spadne v produkci.

### `sync.WaitGroup`

WaitGroup je čítač. `Add(n)` ho zvýší, `Done()` sníží o jedna, `Wait()` blokuje, dokud
není nula.

```go
var wg sync.WaitGroup
for _, job := range jobs {
    wg.Add(1)
    go func() {
        defer wg.Done()
        process(job)
    }()
}
wg.Wait()
```

Tři pravidla, která ber doslova:

1. **`Add` volej před `go`, nikdy uvnitř goroutiny.**

   ```go
   // ŠPATNĚ
   go func() {
       wg.Add(1) // Wait může proběhnout dřív, než se sem scheduler dostane
       defer wg.Done()
       process(job)
   }()
   ```

   `Wait()` uvidí nulu a vrátí se dřív, než goroutina vůbec začala. Test projde
   devětkrát z deseti a v produkci ti tiše zmizí část práce.

2. **`Done` dávej do `defer`.** Když funkce panikuje nebo se vrátí jinou větví, čítač
   zůstane nenulový a `Wait()` bude viset navždy.

3. **WaitGroup nikdy nekopíruj.** Předávej ji jako `*sync.WaitGroup`, nebo ji jen
   zavírej do closure. Kopie má vlastní čítač — `go vet` to naštěstí ohlásí.

WaitGroup neřeší výsledky ani chyby, jen „už jsou všichni hotoví". Na sběr výsledků
potřebuješ kanál (lekce 41) nebo předalokovaný slice.

### Proměnná cyklu: past, která od Go 1.22 neplatí

Do Go 1.21 měla smyčka `for i := range …` **jednu** proměnnou `i` sdílenou všemi
iteracemi. Goroutina spuštěná uvnitř těla viděla její aktuální hodnotu, takže tenhle kód
vypsal třikrát `3`:

```go
for i := 0; i < 3; i++ {
    go func() { fmt.Println(i) }() // do Go 1.21: 3 3 3
}
```

Proto se psalo `go func(i int) { … }(i)` nebo `i := i` na začátku těla. Od **Go 1.22**
má každá iterace vlastní kopii proměnné, takže původní kód vypíše `0 1 2` (v libovolném
pořadí) a obě obcházky jsou zbytečné.

Vědět to musíš z jednoho důvodu: **narazíš na starý kód a na starou radu od AI**. Když
uvidíš `i := i`, není to chyba, jen relikt. Když ale uvidíš goroutinu zavírající se nad
proměnnou deklarovanou **mimo** cyklus, past platí pořád — nová sémantika se týká jen
proměnných cyklu:

```go
buf := make([]byte, 0)
for _, chunk := range chunks {
    go func() { buf = append(buf, chunk...) }() // pořád závod, tohle 1.22 neřeší
}
```

### Sdílený slice bez mutexu

Tohle překvapí skoro každého, kdo přichází z jazyka bez sdílené paměti:

```go
results := make([]int, len(fns))
var wg sync.WaitGroup
for i, fn := range fns {
    wg.Add(1)
    go func() {
        defer wg.Done()
        results[i] = fn() // bez mutexu, a je to naprosto v pořádku
    }()
}
wg.Wait()
```

Datový závod vzniká, když dvě goroutiny sahají na **stejné paměťové místo**. Různé
indexy předalokovaného slice jsou různá místa a slice se během běhu nemění (žádný
`append`, žádná realokace), takže tu není co chránit. Race detektor tenhle kód označí za
čistý a je to zároveň nejlevnější způsob, jak zachovat pořadí výsledků.

Kdybys psal `results = append(results, fn())`, je to okamžitě závod: `append` čte a
zapisuje hlavičku slice, tedy sdílenou paměť.

### Goroutine leak

Leak je goroutina, která už nikdy neskončí a nikdo na ni nečeká. Skoro vždy vznikne
zablokováním na kanálu:

```go
// 1) zápis do kanálu, který nikdo nečte
func leak1() {
    ch := make(chan int) // nebufferovaný
    go func() { ch <- 42 }()
    // vrátíme se dřív, než někdo přečte — goroutina visí navždy
}

// 2) čtení z kanálu, do kterého už nikdo nezapíše
func leak2(ch chan int) {
    go func() {
        for v := range ch { fmt.Println(v) } // producent zapomněl close(ch)
    }()
}
```

V CLI nástroji, který za vteřinu skončí, si toho nevšimneš. V HTTP serveru, který
takovou funkci volá na každém requestu, přibývá goroutin lineárně s provozem. Každá drží
svůj zásobník a všechno, na co ukazuje — tedy i request, tělo odpovědi, connection. Za
pár hodin je z toho OOM kill, který se v žádném logu netváří jako tvoje chyba.

Pravidlo, které to celé řeší: **kdo goroutinu spustil, zodpovídá za její ukončení.** Než
napíšeš `go`, musíš umět odpovědět na otázku „jak tahle goroutina skončí?". Odpověď je
vždy jedna ze tří: doběhne sama do konce funkce, dostane signál přes `done` kanál, nebo
se zruší přes `context` (lekce 42).

Standardní tvar generátoru, který jde zastavit:

```go
func Generator(done <-chan struct{}) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out) // odesílatel zavírá výstup
        for i := 0; ; i++ {
            select {
            case <-done:
                return
            case out <- i: // nikdy holé `out <- i`
            }
        }
    }()
    return out
}
```

### Jak leak najít

V testu nejjednodušeji přes `runtime.NumGoroutine()`. Pevný `Sleep` je ale nespolehlivý —
doběhnutí goroutiny není okamžité a runtime má i vlastní pomocné goroutiny. Spolehlivé je
krátké opakované měření, dokud se číslo neustálí:

```go
func stableGoroutines() int {
    prev, stable := runtime.NumGoroutine(), 0
    for i := 0; i < 300; i++ {
        time.Sleep(5 * time.Millisecond)
        cur := runtime.NumGoroutine()
        if cur == prev {
            if stable++; stable >= 3 {
                return cur
            }
            continue
        }
        prev, stable = cur, 0
    }
    return runtime.NumGoroutine()
}
```

Ještě lepší je synchronizovat přes `done` kanál a `NumGoroutine` použít až jako pojistku.

V běžícím procesu se leak hledá přes pprof: naimportuješ `net/http/pprof` a otevřeš
`/debug/pprof/goroutine?debug=1`. Dostaneš seznam všech živých goroutin seskupený podle
stacku včetně počtu. Když tam vidíš 40 000 goroutin blokovaných na `chan send` ve tvém
`sendMetric`, máš hotovou diagnózu. Detailně se tomu věnuje lekce 53.

A vždycky pouštěj `go test -race`. Závod není totéž co leak, ale souběžný kód bez
`-race` je netestovaný kód.

## Rozdíly proti PHP

V PHP je souběžnost mimo proces. Když chceš zpracovat tisíc faktur „paralelně", pošleš
tisíc zpráv do fronty a necháš to na Messengeru a supervisoru:

```php
foreach ($invoices as $invoice) {
    $this->bus->dispatch(new GenerateInvoice($invoice->id));
}
// tady končí tvoje zodpovědnost — worker běží v jiném procesu,
// jeho životní cyklus řeší supervisor, restarty a memory limit
```

Request je izolovaný proces: má vlastní paměť, po odpovědi umře a všechno po sobě uklidí
operační systém. Sdílený stav neexistuje, protože není s kým sdílet.

V Go je souběžnost **uvnitř tvého procesu** a její životní cyklus je tvoje zodpovědnost:

```go
var wg sync.WaitGroup
for _, invoice := range invoices {
    wg.Add(1)
    go func() {
        defer wg.Done()
        generate(invoice)
    }()
}
wg.Wait() // tady teprve máš hotovo
```

Dva návyky je potřeba opustit. Za prvé: neexistuje supervisor, který by ti zapomenutou
goroutinu restartoval nebo zabil — když ji necháš viset, visí až do konce procesu.
Za druhé: Go server běží týdny, takže každá zapomenutá goroutina je trvalý únik paměti.
Reflex „to se nějak uklidí" je tady nejdražší věc, kterou si z PHP přineseš.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `time.Sleep` místo `wg.Wait()` | reflex „chvíli počkám, ono to doběhne" | `sync.WaitGroup` a `Wait()` |
| `wg.Add(1)` uvnitř goroutiny | zdá se logické mít to u té práce | `Add` před `go`, `Done` v `defer` |
| Goroutina bez plánu, jak skončí | v PHP proces po requestu umře sám | `done` kanál nebo `context`, vždy |
| Holé `out <- v` v dlouho žijící goroutině | v PHP zápis do fronty nikdy neblokuje | `select` s větví na `done` |
| `append` do sdíleného slice z goroutin | vypadá jako totéž co zápis na index | předalokuj a piš na index |
| `i := i` v novém kódu | rada z éry Go 1.21 (a od AI) | od 1.22 zbytečné; závod nad proměnnou **mimo** cyklus platí dál |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 40`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `ParallelSquares`

```bash
make lesson L=40 PART=1
```

Pak **`/go-deep-review 40 easy`**.

### Střední

Funkce: `FanOutSum`, `GoroutineDelta`

```bash
make lesson L=40 PART=2
```

Pak **`/go-deep-review 40 medium`**.

### Obtížný

Funkce: `LeakyGenerator`, `SafeGenerator`

```bash
make lesson L=40 PART=3
```

Pak **`/go-deep-review 40 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 40 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=40` (+ `make race L=40`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč je `wg.Add(1)` uvnitř goroutiny chyba
- [ ] Umíš popsat dva způsoby, jak vzniká goroutine leak, a jak se každý opravuje
- [ ] Umíš vysvětlit, proč je zápis na různé indexy slice z více goroutin bezpečný
- [ ] Umíš říct, co se změnilo v Go 1.22 u proměnné cyklu a co naopak zůstalo stejné
- [ ] Umíš u každého `go` ve svém kódu odpovědět, jak ta goroutina skončí

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Nech si od agenta navrhnout `FanOutSum`, ale nejdřív si napiš vlastní verzi a acceptance
test. V jeho kódu pak hledej přesně to, co je v téhle lekci: má každá goroutina cestu
ven? Je `Add` před `go`? Kdo zavírá kanál výsledků a čeká se předtím na odesílatele?
Neschovává se tam neomezený fan-out?

## Další čtení

1. [Effective Go — Goroutines](https://go.dev/doc/effective_go#goroutines)
2. [pkg.go.dev — sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup)
3. [Go Wiki — LoopvarExperiment](https://go.dev/wiki/LoopvarExperiment)
4. [pkg.go.dev — net/http/pprof](https://pkg.go.dev/net/http/pprof)
