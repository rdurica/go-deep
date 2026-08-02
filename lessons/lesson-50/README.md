# Lekce 50 — Checkpoint fáze 5 + projekt P04

> **Čas:** ~90 min · **Fáze:** 5 — Concurrency do hloubky · **AI režim:** `JUNIOR POD REVIEW`

Checkpoint nemá novou látku. Projdeš si fázi 5 (lekce 40–49) formou otázek a tabulek,
naučíš se rozeznat tři nejčastější patologie souběžného kódu podle příznaků, postavíš
jeden kumulativní zpracovatel dávek a odevzdáš projekt **P04**.

## Co budeš umět

- Vysvětlit každou konstrukci z fáze 5 a říct, kdy je správnou volbou a kdy ne.
- Rozeznat deadlock, goroutine leak a datový závod podle příznaků, bez hádání.
- Přečíst goroutine dump a najít v něm zablokovanou goroutinu.
- Postavit zpracovatel dávek, který má limit, backpressure, rušení, sběr chyb a metriky.
- Sebehodnotit se podle rubriky a vědět, které lekce si zopakovat.

## Recap fáze 5

### Otázky, na které musíš umět odpovědět

| Otázka | Odpověď v jedné větě | Lekce |
|--------|----------------------|-------|
| Proč není `time.Sleep` synchronizace? | Nevytváří happens-before hranu, jen doufá. | 40, 48 |
| Kde patří `wg.Add`? | Před `go`, nikdy dovnitř goroutiny. | 40 |
| Kdo zavírá kanál? | Ten, kdo do něj zapisuje — a nikdo jiný. | 41 |
| Proč `close(done)` a ne `done <- struct{}{}`? | Zavření funguje bez čtenáře a pro všechny. | 40, 41 |
| Co udělá `select` s víc připravenými větvemi? | Vybere náhodně, na pořadí `case` nezáleží. | 42 |
| Kdy mutex a kdy kanál? | Mutex chrání stav, kanál předává vlastnictví. | 43 |
| Co `-race` najde a co ne? | Jen závody, které při běhu skutečně nastanou. | 44 |
| Jak ukončit pipeline? | Zavřením vstupu, nebo zrušením kontextu ve všech `select`ech. | 45 |
| Jak zachovat pořadí výsledků? | Index v úloze a zápis do předalokovaného slice. | 46 |
| Co dělá errgroup nad `WaitGroup`? | První chyba, zrušení ostatních, limit. | 47 |
| Rozdíl `ctx.Err()` a `context.Cause(ctx)`? | Příznak versus skutečný důvod. | 47 |
| Co garantuje memory model? | Viditelnost jen tam, kde je synchronizační bod. | 48 |
| Co nastavuje `GOMAXPROCS`? | Počet P, tedy strop paralelismu. Nic jiného. | 49 |

### Co si musíš pamatovat

| Konstrukce | Použij, když | Nepoužívej, když |
|-----------|--------------|------------------|
| `sync.WaitGroup` | chceš jen vědět „už jsou všichni hotoví" | potřebuješ výsledky nebo chyby |
| Kanál | předáváš vlastnictví dat mezi goroutinami | jen chráníš jedno pole ve struktuře |
| `sync.Mutex` | chráníš krátkou kritickou sekci nad stavem | držel by se přes IO nebo dlouhý výpočet |
| `sync.RWMutex` | čtení je řádově častější než zápis | zápisů je podobně jako čtení (režie se nevyplatí) |
| `atomic` | čítač, příznak, publikace jedné hodnoty | potřebuješ udržet invariant nad víc proměnnými |
| `sync.Once` | líná inicializace, která se nesmí opakovat | inicializace může selhat a chceš retry |
| `select` + `ctx.Done()` | kdekoli, kde goroutina může čekat | (vždycky to chceš) |
| Worker pool | dlouhá nebo neomezená fronta úloh | jednorázová dávka o desítkách prvků |
| Semafor (bufferovaný kanál) | omezit souběžnost jednorázové dávky | úloh je milion (režie goroutiny na položku) |
| Vlastní `Group` (errgroup) | pár souběžných volání, kde první chyba ruší zbytek | úlohy jsou nezávislé a chyby chceš všechny |

### Ukončovací vzory

```go
// 1) došla práce: zavři vstup
close(jobs)      // for range v workerech skončí sám
wg.Wait()        // rozdělané úlohy se dodělají — graceful drain

// 2) volající to vzdal: zruš kontext
cancel()         // každý select s ctx.Done() se probudí — okamžité zastavení

// 3) chyba: první chyba ruší zbytek
g, ctx := WithContext(parent)
// … g.Go(…) …
err := g.Wait()  // vrátí první chybu, ostatní dostaly zrušený ctx
```

## Rozdíly proti PHP

Na začátku fáze byla souběžnost něco, co v PHP dělá infrastruktura:

```php
// supervisor: numprocs=8; messenger; redis; retry v konfiguraci
$this->bus->dispatch(new ProcessBatch($ids));
```

Teď je to něco, co navrhuješ ty — včetně limitu, frontování, rušení a metrik:

```go
results, stats, err := Process(ctx, Config{
    Workers:   runtime.NumCPU(),
    QueueSize: 64,
    FailFast:  false,
    Handler:   handleItem,
}, items)
```

Ten druhý blok kódu má stejnou zodpovědnost jako celá `supervisor.conf` plus polovina
`messenger.yaml`. To je ta výměna: dostaneš plnou kontrolu a s ní plnou odpovědnost.

## Deadlock vs leak vs závod

Tři různé patologie, které se snadno pletou. Rozeznat je podle příznaků je nejužitečnější
diagnostická schopnost celé fáze.

| | Deadlock | Goroutine leak | Datový závod |
|---|----------|----------------|--------------|
| **Příznak** | program stojí a nedělá nic | paměť roste, `NumGoroutine` roste | občas špatný výsledek, občas panika |
| **Kdy** | hned a spolehlivě | pod provozem, po hodinách | nedeterministicky, hlavně pod zátěží |
| **Runtime to řekne?** | ano, když stojí **všechny** goroutiny | ne, nikdy | jen s `-race` |
| **Nástroj** | goroutine dump | `pprof/goroutine`, test s `NumGoroutine` | `go test -race` |
| **Typická příčina** | čekání na kanál, do kterého nikdo nepíše; `Lock` dvakrát | chybějící `case <-ctx.Done()`; nezavřený kanál | zápis do mapy nebo slice bez zámku |

### „all goroutines are asleep - deadlock!"

Tuhle hlášku vypíše runtime, když **žádná** goroutina nemůže pokračovat. Je to dobrá
zpráva: dostaneš kompletní dump a přesnou příčinu.

```
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan receive]:
main.main()
	/app/main.go:12 +0x38
```

Postup: podívej se na stav v hranatých závorkách (`chan receive`, `chan send`,
`sync.Mutex.Lock`, `select`), najdi ten řádek a zeptej se, **kdo měl na druhé straně
jednat**. Odpověď je skoro vždycky jedna ze čtyř:

- nikdo nezavřel kanál, nad kterým visí `for range`,
- nikdo nečte kanál, do kterého se zapisuje (nebufferovaný `out <- v`),
- `wg.Wait()` čeká na `Done`, které se kvůli chybějícímu `defer` nikdy nezavolalo,
- `Lock` na už zamčeném mutexu ve stejné goroutině (typicky metoda pod zámkem volá
  jinou metodu, která zamyká znovu).

Pozor na past: pokud v programu běží **jediná** živá goroutina, například HTTP server
nebo `time.Ticker`, runtime deadlock **neohlásí**. Program bude vypadat, že „jen nic
nedělá". Proto se v produkci hledá zablokovaná goroutina ručně.

### Jak číst goroutine dump

Dump si vyžádáš třemi způsoby:

```bash
kill -QUIT <pid>            # vypíše dump všech goroutin a program ukončí
GOTRACEBACK=all ./app       # při panice vypíše i goroutiny mimo tu padající
curl localhost:6060/debug/pprof/goroutine?debug=2   # s importem net/http/pprof
```

Čte se to takhle:

```
goroutine 4711 [chan send, 12 minutes]:
myapp/metrics.(*Sender).push(0xc0000b4000, ...)
	/app/metrics/sender.go:88 +0x94
```

Tři užitečné informace na jednom řádku. `chan send` říká, na čem to visí. `12 minutes`
říká, jak dlouho — a to je nejsilnější signál leaku, protože zdravá goroutina tak dlouho
nečeká. A když v dumpu vidíš tisíce goroutin se **stejným** stackem, máš hotovou diagnózu:
v `sender.go:88` chybí `select` s `ctx.Done()`.

V testech stačí postup z lekce 40: změř `runtime.NumGoroutine()` před a po, s krátkým
pollingem na stabilizaci.

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 50`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `Total`, `Throughput`

```bash
make lesson L=50 PART=1
```

Pak **`/go-deep-review 50 easy`**.

### Střední

Funkce: `Enter`, `Leave`

```bash
make lesson L=50 PART=2
```

Pak **`/go-deep-review 50 medium`**.

### Obtížný

Funkce: `Snapshot`, `Process`, `ProcessStream`

```bash
make lesson L=50 PART=3
```

Pak **`/go-deep-review 50 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Projekt P04

Zadání a akceptační kritéria: [projects/p04-worker-pool/ACCEPTANCE.md](../../projects/p04-worker-pool/ACCEPTANCE.md).

Postav znovupoužitelnou komponentu, ne skript: generický worker pool s limitem souběžnosti,
backpressure, gracefulním ukončením i okamžitým zastavením, sběrem výsledků a chyb a
metrikami. K tomu ukázkové CLI, které zpracuje seznam úloh ze stdin a vypíše report.

```bash
cd projects/p04-worker-pool
gofmt -l . && go vet ./... && go test -race -count=3 ./...
printf 'jedna\ndve tri\n!chybna\n' | go run ./cmd/crawl -workers=2 -queue=4
```

## Sebehodnocení

Za každou položku si dej body podle skutečnosti, ne podle dojmu. Maximum je 30.

| # | Kritérium | 0 bodů | 1 bod | 2 body |
|---|-----------|--------|-------|--------|
| 1 | Cvičení 46–49 | nedokončil jsem | dokončil s pomocí `solutions/` | dokončil sám |
| 2 | Testy s `-race` | neprošly | prošly po opravách | prošly hned |
| 3 | Leaky | test na leak jsem nepsal | píšu ho, když si vzpomenu | píšu ho automaticky |
| 4 | Ukončení goroutiny | nevím vždycky, jak skončí | většinou vím | u každého `go` to umím říct |
| 5 | Volba mutex vs kanál | hádám | rozhodnu s rozmyslem | rozhodnu a umím to obhájit |
| 6 | Worker pool | opsal jsem ho | napsal s nápovědou | napsal z hlavy včetně ukončení |
| 7 | Vlastní `Group` | nezvládl jsem | s nápovědou | z hlavy včetně `SetLimit` |
| 8 | Kontext | píšu ho, kam řeknou | znám hlavní anti-vzory | vysvětlím `WithoutCancel` i `Cause` |
| 9 | Happens-before | „atomic je bezpečnější" | vyjmenuji body | dokážu ukázat hranu v konkrétním kódu |
| 10 | Scheduler | pletu G, M, P | rozumím modelu | vysvětlím netpoller i preempci |
| 11 | Diagnostika | příznaky si pletu | rozliším deadlock a závod | rozliším všechny tři a vím čím |
| 12 | Goroutine dump | nečetl jsem ho | přečtu s pomocí | najdu v něm leak sám |
| 13 | Projekt P04 | neodevzdán | prochází `go test` | prochází `-race -count=3` a splňuje ACCEPTANCE |
| 14 | Metriky souběhu | nemám | čítače | čítače i maximum souběhu přes CAS |
| 15 | Backpressure | „přidám buffer" | vím, proč fronta tlačí zpět | umím ji navrhnout i otestovat |

**Vyhodnocení**

| Skóre | Co s tím |
|-------|----------|
| 27–30 | Fáze 5 je hotová, pokračuj na lekci 51. |
| 21–26 | Zopakuj si lekce, kde jsi dal 0–1 bodu, a přepiš odpovídající cvičení z hlavy. |
| 14–20 | Vrať se k lekcím 40–43 a projdi je znovu včetně cvičení. Bez jistoty v základech je zbytek fáze jen recitace. |
| 0–13 | Projdi celou fázi znovu. Nespěchej — concurrency je jediná část Go, kde nepochopení nezpůsobí chybu kompilace, ale incident v produkci. |

## Závěrečné otázky

Spusť **`/go-deep-review 50 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=50` (+ `make race L=50`, pokud to lekce vyžaduje).

- [ ] `make project P=04` prochází
- [ ] Umíš rozeznat deadlock, leak a závod podle příznaků a říct, čím se hledají
- [ ] Umíš přečíst řádek goroutine dumpu a určit, na čem goroutina visí
- [ ] Umíš vysvětlit rozdíl mezi graceful drainem a zrušením kontextem
- [ ] Sebehodnocení máš vyplněné a víš, co si zopakovat

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

U P04 postupuj přesně podle režimu: nejdřív `ACCEPTANCE.md` a testy, pak návrh od agenta,
pak review checklist. U worker poolu se dívej na pět věcí: kdo zavírá kanál výsledků, co
dělá `Submit` po `Close`, jestli je `select` na zrušení i v zápisu do výstupu, jestli
metriky jdou přes atomiku, a jestli někde nevznikl buffer „pro rychlost", který zabil
backpressure.

## Další čtení

1. [go.dev — Diagnostics](https://go.dev/doc/diagnostics) — kompletní přehled nástrojů
2. [The Go Memory Model](https://go.dev/ref/mem)
3. [Go blog — Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
4. [pkg.go.dev — runtime/debug SetTraceback](https://pkg.go.dev/runtime/debug#SetTraceback)
