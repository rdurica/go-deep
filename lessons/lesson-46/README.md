# Lekce 46 — Worker pool

> **Čas:** ~90 min · **Fáze:** 5 — Concurrency do hloubky · **AI režim:** `JUNIOR POD REVIEW`

## Co budeš umět

- Vysvětlit, proč je neomezený fan-out chyba, i když goroutina stojí jen dva kilobajty.
- Postavit worker pool jako N goroutin nad jedním sdíleným kanálem úloh a korektně ho ukončit.
- Rozhodnout mezi plným poolem a semaforem přes bufferovaný kanál — a vědět, kdy stačí to druhé.
- Sbírat výsledky se zachovaným pořadím a s chybami jednotlivých úloh, bez mutexu nad slice.
- Odhadnout počet workerů podle toho, jestli je práce CPU-bound, nebo IO-bound.

## PHP → Go most

V Symfony se souběžnost obvykle neomezuje v kódu — omezuje ji infrastruktura. Napíšeš
handler, a kolik jich poběží najednou, rozhodne konfigurace supervisoru:

```php
// messenger: kolik zpráv se zpracuje současně, řeší počet procesů
// supervisor: numprocs=8  →  osm PHP procesů, osm spojení do DB
final class GenerateThumbnailHandler
{
    public function __invoke(GenerateThumbnail $msg): void { /* ... */ }
}
```

V Go je ten dial uvnitř tvého programu a nikdo ho za tebe nenastaví:

```go
// bez limitu — 50 000 obrázků = 50 000 souběžných konverzí a OOM
for _, img := range images {
    go generate(img)
}

// s limitem — přesná obdoba numprocs=8, jen o tři patra níž
sem := make(chan struct{}, 8)
for _, img := range images {
    sem <- struct{}{}
    go func() {
        defer func() { <-sem }()
        generate(img)
    }()
}
```

Návyk, který je potřeba opustit: „škálování je věc opsu". V Go je počet souběžných úloh
součást návrhu funkce, stejně jako její signatura. A přibývá k tomu druhá zodpovědnost,
kterou PHP nemá vůbec — musíš vědět, **jak pool skončí**, protože proces po requestu
nezemře.

## Teorie

### Proč vůbec omezovat

Goroutina je levná, ale to, co dělá, levné být nemusí. Neomezený fan-out narazí vždycky
na jeden ze čtyř stropů:

| Zdroj | Co se stane při 10 000 souběžných úlohách |
|-------|-------------------------------------------|
| Paměť | každá goroutina drží buffer, dekódovaný obrázek, tělo odpovědi — a všechny najednou |
| Spojení do DB | `database/sql` má `SetMaxOpenConns`; nad limit se čeká, ale ty už držíš 10 000 goroutin |
| Cizí API | rate limit, `429`, ban IP adresy |
| Souborové deskriptory | `too many open files` — a spadne i to, co s tvou dávkou nesouvisí |

Rozdíl proti PHP je v tom, že tady se dá strop překročit tiše. Program se nezastaví,
jen začne dělat všechno pomaleji a hůř, dokud ho nezabije OOM killer.

### Pool jako N goroutin nad jedním kanálem

Základní tvar je překvapivě malý. Kanál úloh je fronta, workeři jsou konzumenti:

```go
jobs := make(chan Job)
var wg sync.WaitGroup
wg.Add(workers)
for i := 0; i < workers; i++ {
    go func() {
        defer wg.Done()
        for job := range jobs { // range skončí, až kanál někdo zavře
            process(job)
        }
    }()
}
for _, j := range list {
    jobs <- j
}
close(jobs)
wg.Wait()
```

Všimni si, že workeři nemají žádný `stop` kanál ani `select`. Nepotřebují ho: ukončovací
signál je **zavření vstupního kanálu**. `for range` nad zavřeným a vyčerpaným kanálem
skončí sám. To je nejlevnější a nejčitelnější způsob ukončení poolu, jaký v Go existuje.

Nebufferovaný `jobs` navíc dává zadarmo backpressure. Zápis `jobs <- j` blokuje, dokud
si úlohu někdo nevezme, takže producent běží přesně tak rychle, jak stíhají workeři.
Kdybys dal kanálu buffer 100 000, producent by celou dávku nasypal do paměti a strop
si posunul jen o kus dál.

### Ukončení: zavřený kanál vs kontext

Existují dva různé důvody, proč pool končí, a pletou se:

```go
// 1) došla práce — graceful drain
close(jobs)
wg.Wait() // všechny rozdělané úlohy se dodělají

// 2) volající to vzdal — okamžité zastavení
for {
    select {
    case job, ok := <-jobs:
        if !ok {
            return
        }
        process(job)
    case <-ctx.Done():
        return // rozdělaná fronta se zahodí
    }
}
```

První varianta znamená „nic nového už nepřijde, ale co je ve frontě, dokonči". Druhá
znamená „nech všeho". Produkční pool umí obojí, protože to jsou různé situace: konec
dávky versus `SIGTERM`. Kdo je má slité do jednoho, buď při shutdownu ztrácí práci, nebo
při konci dávky zbytečně čeká.

### Sběr výsledků a pořadí

Nejčastější chyba v poolech od AI:

```go
// ŠPATNĚ — datový závod na hlavičce slice
var results []Result
for job := range jobs {
    results = append(results, process(job))
}
```

`append` čte a zapisuje délku, kapacitu i ukazatel na pole. Tři goroutiny, tři zápisy na
totéž místo, `-race` to hlásí okamžitě — a i kdyby ne, tiše přijdeš o výsledky.

Řešení jsou dvě a obě jsou lepší než mutex. Buď kanál výsledků:

```go
results := make(chan Result, workers)
```

…nebo, když chceš zachovat **pořadí**, předalokovaný slice a index nesený v úloze:

```go
out := make([]U, len(in))
for i, v := range in {
    i, v := i, v
    go func() { out[i] = f(v) }() // různé indexy = různá paměťová místa, žádný závod
}
```

Pořadí se nedá získat z kanálu — z něj výsledky lezou v pořadí dokončení, ne zadání.
Buď si tedy index neseš v úloze a řadíš až nakonec, nebo píšeš rovnou na index.

Kanál výsledků má jeden háček, na který se v poolech naráží pořád: **kdo ho zavírá?**
Zavřít ho nesmí jeden worker (ostatní ještě píšou) ani volající (workeři by panikovali).
Správně to dělá samostatná goroutina:

```go
go func() {
    wg.Wait()
    close(results)
}()
```

### Chyby jednotlivých úloh

Chyba jedné úlohy typicky neznamená konec celé dávky. Proto chyba patří **do výsledku**,
ne do návratové hodnoty poolu:

```go
type Result struct {
    ID    int
    Value int
    Err   error
}
```

Volající pak sám rozhodne, jestli se má při první chybě zastavit, nebo dojet dávku a
chyby posbírat. Kdyby pool na první chybě spadl, tuhle volbu bys mu vzal. Opačný vzor —
„první chyba ruší zbytek" — má taky své místo, a je to přesně úkol C.

### Semafor jako lehčí alternativa

Pool s kanálem a workery dává smysl, když máš dlouho žijící frontu. Když chceš jen
„zpracuj tenhle slice, nejvýš pět najednou", je pool zbytečně těžký. Stačí semafor:

```go
type Semaphore struct{ slots chan struct{} }

func (s *Semaphore) Acquire(ctx context.Context) error {
    select {
    case s.slots <- struct{}{}:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
func (s *Semaphore) Release() { <-s.slots }
```

Kapacita kanálu **je** ten limit. Rozdíl proti poolu je zásadní: pool spustí N goroutin
a ty si berou práci, semafor spustí goroutinu na každou položku a ty se řadí frontu na
vstupenku. U pár tisíc položek je to jedno, u milionu chceš pool.

Pozor na jeden detail, který má vlastní test: když `Acquire` skončí chybou, místo se
**nesmí** zabrat. Jinak by ho nikdo neuvolnil a semafor by se postupně ucpal.

### Kolik workerů

Nedá se to uhodnout, ale dá se to odhadnout podle typu práce:

- **CPU-bound** (komprese, hashování, parsování): `runtime.NumCPU()`. Víc goroutin než
  jader nepřidá výkon, jen režii přepínání. Menší hodnota naopak nechá jádra ležet ladem.
- **IO-bound** (HTTP, DB, disk): mnohem víc než jader, protože většinu času goroutina
  čeká a nic neblokuje. Strop tady nediktuje CPU, ale protistrana — velikost poolu spojení
  do DB nebo rate limit API. Nastav ho na to, ne na `NumCPU`.

```go
workers := runtime.NumCPU()  // CPU-bound
workers := 32                // IO-bound: tolik, kolik unese protistrana
```

A poslední pravidlo: v kontejneru s CPU limitem `NumCPU()` do Go 1.24 včetně vrací počet
jader **hostitele**, ne tvůj limit. Podrobně v lekci 49. Vždycky měř, nehádej.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `go f(x)` v cyklu bez limitu | v PHP by to udělal supervisor za tebe | semafor nebo pool s pevným N |
| `append` do sdíleného slice z workerů | vypadá jako zápis na index | kanál výsledků, nebo zápis na index |
| Kanál výsledků zavírá worker | „skončil jsem, zavřu" | `go func(){ wg.Wait(); close(results) }()` |
| Velký buffer na vstupním kanálu | „ať to nebrzdí" | nebufferovaný kanál = backpressure zadarmo |
| Pool vrací `error` místo chyby v `Result` | reflex z PHP, kde výjimka ukončí request | chyba patří k úloze, volající rozhodne |
| `workers = 1000` pro DB dotazy | „IO-bound, tak hodně" | limit diktuje protistrana, ne fantazie |
| Zabrané místo v semaforu i při chybě `Acquire` | zapomenutá větev v `select` | při chybě se místo nezabírá, `Release` jen po úspěchu |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

### A — rozcvička (~10 min)

1. `NewSemaphore(n int) *Semaphore` s metodami `Acquire(ctx) error`, `TryAcquire() bool`
   a `Release()`. Kapacitu drž v bufferovaném kanálu. `Acquire` musí respektovat zrušení
   kontextu a při chybě **nesmí** zabrat místo; u už zrušeného kontextu vrací chybu, i
   kdyby bylo volno. `Release` bez předchozího `Acquire` panikuje, `NewSemaphore(0)` taky.
2. `LimitedMap(ctx, inputs []string, limit int, f func(string) string) []string` — aplikuj
   `f` na všechny vstupy, nejvýš `limit` souběžně, výsledky ve **stejném pořadí** jako
   vstup. `limit <= 0` se chová jako 1, `f == nil` vrací `nil`, zrušený kontext nechá
   nespuštěné položky na prázdném řetězci. Po návratu nesmí zůstat běžící goroutina.

např. `LimitedMap(ctx, []string{"a", "b"}, 0, strings.ToUpper)` → `["A", "B"]`

### B — jádro (~35 min)

Postav `Pool` s pevným počtem workerů:

- `New(workers int) *Pool` — spustí workery, pro `workers <= 0` panikuje.
- `Submit(job Job) error` — blokuje, dokud úlohu nepřevezme worker (backpressure).
  Po `Close` vrací `ErrPoolClosed` a úlohu zahodí. **Nesmí** zapsat do zavřeného kanálu.
- `Results() <-chan Result` — kanál výsledků, který se zavře, až doběhnou všechny úlohy
  přijaté před `Close`.
- `Close()` — idempotentní, oznámí, že další úlohy nepřijdou.

`Job` s `Run == nil` musí dát `Result{Err: ErrNilJob}`, ne paniku. Test ověřuje tři věci:
že projde všech 20 úloh, že pool nikdy nespustí víc než `workers` úloh současně (měří to
atomickým maximem), a že po dočerpání `Results()` nezůstane žádná goroutina navíc.

Kanál úloh je nebufferovaný, takže z `Results()` se musí číst souběžně se `Submit` —
tohle si napiš do doc commentu, je to součást kontraktu.

např. `Submit(Job{ID: 1, Run: →7}); Close()` → `Result{Value: 7}`

### C — rozšíření (~20 min)

`MapErr[T, U any](ctx, in []T, limit int, f func(context.Context, T) (U, error)) ([]U, error)`:

- omezená souběžnost (`limit`), zachované pořadí výsledků,
- **první chyba** zruší odvozený kontext, takže se běžící volání `f` dozvědí, že nemá
  smysl pokračovat; funkce počká na jejich doběhnutí a vrátí `(nil, ta první chyba)`,
- `limit <= 0` → `ErrInvalidWorkers`, `f == nil` → `ErrNilJob`, prázdný vstup → prázdný
  výsledek a `nil`, už zrušený vstupní kontext → `ctx.Err()`.

Nápověda: `context.WithCancel` + `sync.Once` na zapamatování první chyby. A nezapomeň na
`defer cancel()`, jinak ti unikne kontext i v úspěšné větvi.

např. `MapErr(ctx, []int{1, 2, 3}, 2, f)` → `["*", "**", "***"]`

```bash
make lesson L=46
make race L=46
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `46`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=46` prochází
- [ ] `make race L=46` prochází (žádné hlášení race detektoru)
- [ ] Umíš vysvětlit, proč `for range jobs` nepotřebuje `stop` kanál
- [ ] Umíš vysvětlit, kdo smí zavřít kanál výsledků a proč
- [ ] Umíš vysvětlit rozdíl mezi graceful drain a zrušením kontextem
- [ ] Umíš říct, proč je zápis na různé indexy slice bez mutexu v pořádku, ale `append` ne
- [ ] Umíš zdůvodnit počet workerů zvlášť pro CPU-bound a IO-bound práci

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Worker pool je klasické zadání, na kterém se dá dobře trénovat review. Nech si ho
vygenerovat a projdi: zavírá kanál výsledků správná goroutina? Má `Submit` po `Close`
definované chování, nebo panikuje? Je vstupní kanál bufferovaný „pro rychlost"? Sbírá
výsledky přes `append` pod mutexem tam, kde by stačil index? Tyhle čtyři otázky odhalí
většinu vygenerovaných poolů.

## Další čtení

1. [Go blog — Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
2. [pkg.go.dev — golang.org/x/sync/semaphore](https://pkg.go.dev/golang.org/x/sync/semaphore) — vážený semafor ze stejné rodiny
3. [pkg.go.dev — database/sql SetMaxOpenConns](https://pkg.go.dev/database/sql#DB.SetMaxOpenConns)
4. [pkg.go.dev — runtime.NumCPU](https://pkg.go.dev/runtime#NumCPU)
