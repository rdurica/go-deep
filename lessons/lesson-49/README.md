# Lekce 49 — Scheduler: mentální model G-M-P

> **Čas:** ~90 min · **Fáze:** 5 — Concurrency do hloubky · **AI režim:** `JUNIOR POD REVIEW`

## Co budeš umět

- Popsat, co je G, M a P a jak si mezi sebou předávají práci.
- Vysvětlit, co `GOMAXPROCS` skutečně nastavuje — a co ne.
- Předpovědět, co se stane s vláknem při blokujícím syscallu a co při čtení ze socketu.
- Rozhodnout, kdy má ladění `GOMAXPROCS` smysl (spoiler: hlavně v kontejnerech).
- Změřit dosažený souběh a cenu goroutiny místo hádání.

## PHP → Go most

V PHP-FPM je model plánování viditelný a triviální: jeden request = jeden proces, kolik
procesů, tolik současně obsloužených requestů.

```ini
; php-fpm pool
pm = static
pm.max_children = 32   ; přesně 32 requestů naráz, 33. čeká ve frontě
```

Tvůj kód o tom nic neví a nemůže to ovlivnit. Souběžnost je konfigurace, ne kód.

V Go je „pool workerů" runtime samotný a chová se na dvou úrovních:

```go
runtime.GOMAXPROCS(0) // kolik goroutin může BĚŽET současně (typicky = počet jader)
runtime.NumGoroutine() // kolik goroutin EXISTUJE (klidně 100 000)
```

Co se mění v uvažování: přestaň si představovat „počet workerů" jako jedno číslo. V Go
jsou to dvě různá čísla a mezi nimi vrstva, která goroutiny na vlákna namapuje. Když se
tvůj program chová divně pod zatížením, odpověď je skoro vždycky v tom, jak se ta dvě
čísla potkávají.

## Teorie

### G, M, P

Tři písmena, tři různé věci:

| | Co to je | Kolik jich je |
|---|----------|---------------|
| **G** | goroutina — zásobník, PC a stav | tisíce až miliony |
| **M** | machine, tedy OS vlákno | podle potřeby, běžně desítky |
| **P** | processor — kontext nutný k běhu Go kódu, drží lokální frontu | `GOMAXPROCS`, typicky počet jader |

Pravidlo, které z toho plyne a které stojí za zapamatování: **aby G běžela, musí být
připojená k M a to k P.** P je omezený zdroj, a proto je počet *současně běžících*
goroutin nejvýš `GOMAXPROCS`. Počet *existujících* goroutin je neomezený.

Každé P má svou lokální frontu (kapacita 256) a k tomu existuje jedna globální. Nová
goroutina jde do lokální fronty toho P, které ji spustilo — proto se dvě goroutiny
spuštěné vedle sebe často provádějí na stejném jádru a data zůstanou v jeho cache.

### Work stealing

Kdyby každé P obsluhovalo jen svou frontu, jedno by se přetížilo a ostatní by zahálela.
Proto P, kterému práce došla, hledá v tomhle pořadí:

1. vlastní lokální frontu,
2. globální frontu (občas se do ní kouká i mimo tenhle případ, aby nevyhladověla),
3. netpoller (dokončené síťové operace),
4. **ukradne polovinu** lokální fronty náhodně vybraného jiného P.

Praktický dopad na tvůj kód: nemusíš rozdělovat práci rovnoměrně. Když jedna úloha trvá
desetkrát dýl než ostatní, scheduler to vyrovná sám. Ruční „sharding podle indexu" na N
goroutin je v Go skoro vždycky zbytečný.

### Co GOMAXPROCS nastavuje

`GOMAXPROCS` nastavuje **počet P**. Nic jiného. Nenastavuje:

- počet goroutin — ten určuje tvůj kód,
- počet OS vláken — těch může být klidně stokrát víc (viz níž),
- CPU limit procesu — to je věc cgroups.

```go
runtime.GOMAXPROCS(1)         // jedno P
go a(); go b(); go c()        // tři G, běží po jedné, ale všechny doběhnou
```

Při `GOMAXPROCS(1)` neztratíš souběžnost, jen paralelismus. Program se pořád prokládá
(a od Go 1.14 i preemptivně), jen v jednu chvíli běží nejvýš jedna goroutina Go kódu.

### Blokující syscall a netpoller

Tady je nejzajímavější část celého modelu. Co se stane, když goroutina zavolá něco, co
blokuje?

**Blokující syscall** (čtení z souboru, `time.Sleep` na úrovni OS, volání do C):

```go
data, err := os.ReadFile("/velky/soubor") // M zůstane zablokované v jádře
```

Runtime po chvíli zjistí, že M v syscallu trčí, **odpojí od něj P** a přidělí ho jinému M
(vezme si volné, nebo vyrobí nové). Díky tomu jedna blokující operace nezastaví ostatní
goroutiny. Cena je jedno vlákno navíc — a proto může mít Go program víc vláken než jader,
i když je `GOMAXPROCS` rovné počtu jader. Strop je `runtime/debug.SetMaxThreads`, defaultně
10 000.

**Síťové IO** se řeší jinak a lépe. Sockety runtime nastaví do neblokujícího režimu a
zaregistruje je do **netpolleru** (epoll na Linuxu, kqueue na BSD, IOCP na Windows):

```go
n, err := conn.Read(buf) // vypadá blokující, ale vlákno neblokuje
```

Když data nejsou, goroutina se odloží (`parked`) a P si okamžitě vezme jinou práci. Až
epoll ohlásí připravenost, goroutina se vrátí do fronty. To je důvod, proč Go zvládne
sto tisíc otevřených spojení na hrstce vláken a proč se v Go nepíše nic jako ReactPHP
nebo Swoole — asynchronní IO je pod tou synchronně vypadající funkcí schované.

### Preempce

Do Go 1.13 byl scheduler kooperativní: goroutinu bylo možné odebrat P jen na
„bezpečném bodě", což bylo v praxi volání funkce nebo alokace. Těsná smyčka bez volání
proto P okupovala navždy:

```go
// v Go 1.13 a starším: tohle při GOMAXPROCS=1 zastavilo celý program
for {
    x++
}
```

Od Go 1.14 je preempce **asynchronní** — runtime pošle vláknu signál (`SIGURG`) a
goroutinu odebere i v takové smyčce. Praktické důsledky dva. Za prvé: nemusíš do smyček
sypat `runtime.Gosched()`, což bylo dřív běžné. Za druhé: když ve starším kódu nebo
v článku na blogu narazíš na `Gosched()`, je to skoro vždycky mrtvá relikvie.

Co preempce **neřeší**: goroutina, která nikdy nekontroluje `ctx.Done()`, se pořád
nezastaví. Preempce ji jen přestane nechávat monopolizovat P.

### Goroutina není zdarma

Levná neznamená bezplatná:

- zásobník startuje na **2 KB** a roste zdvojnásobováním (runtime ho zkopíruje na nové,
  větší místo — proto se v Go nepředávají ukazatele na lokální proměnné mezi vlákny
  nijak zvlášť opatrně, kompilátor to řeší escape analýzou),
- k tomu struktura `g` v runtime a místo ve frontě,
- a hlavně: všechno, na co goroutina ukazuje, se nemůže uvolnit. Jedna zapomenutá
  goroutina držící 1 MB response body je milionkrát dražší než její vlastní zásobník.

Milion goroutin je reálný, ale je to už 2 GB jen na zásobníky. Sto tisíc je běžná
provozní hodnota.

### Kdy ladit GOMAXPROCS

V naprosté většině případů nikdy — default (počet jader) je správný. Jedna výjimka je ale
důležitá a týká se každého, kdo deployuje do Kubernetes.

Do Go 1.24 včetně runtime **nečte CPU limit z cgroups**. V kontejneru s `limits.cpu: 2`
běžícím na 64jádrovém nodu tedy `GOMAXPROCS` bude 64. Runtime bude plánovat na 64 P,
kernel ti dá dvě jádra, a rozdíl zaplatíš agresivním CPU throttlingem a latenčními
špičkami. Řešení do Go 1.24:

```yaml
env:
  - name: GOMAXPROCS
    value: "2"   # ručně podle limitu
```

Go 1.25 zavedlo automatické čtení cgroup limitu, takže na novějších verzích tenhle
problém mizí. Pořád ale platí, že limit se v produkci mění a runtime verze se liší podle
prostředí — takže tuhle proměnnou v manifestu chceš vidět explicitně.

A hlavně: **měř**. Naměřit se dá souběh (atomické maximum souběžně běžících goroutin),
počet goroutin (`runtime.NumGoroutine`), paměť zásobníků (`runtime.MemStats.StackInuse`)
a chování scheduleru (`GODEBUG=schedtrace=1000`). Změna `GOMAXPROCS` bez měření je hádání
s horším reportingem.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `GOMAXPROCS` = „počet goroutin" | název svádí | je to počet P, tedy strop paralelismu |
| Default `GOMAXPROCS` v kontejneru s CPU limitem | „runtime to přece ví" | do Go 1.24 nastav `GOMAXPROCS` podle limitu |
| `runtime.Gosched()` ve smyčkách | zvyk z Go před 1.14 | od 1.14 je preempce asynchronní, není potřeba |
| „Blokující `conn.Read` blokuje vlákno" | vypadá to synchronně | síťové IO jde přes netpoller, vlákno se uvolní |
| Ruční sharding práce mezi N goroutin | reflex z PHP: rozdělit dávku na procesy | work stealing to vyrovná sám |
| `GOMAXPROCS(1)` „pro bezpečnost" místo synchronizace | „když poběží po jedné, není závod" | preempce může přerušit kdekoli, závod zůstává |
| Ladění `GOMAXPROCS` bez měření | vypadá to jako levná optimalizace | změř souběh, latenci a throttling |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test. Testy jsou úmyslně
tolerantní — měříme chování scheduleru, ne stopky.

### A — rozcvička (~10 min)

1. `RunWithMaxProcs(n int, f func())` — nastaví `GOMAXPROCS` na `n`, spustí `f` a
   **v `defer`** obnoví původní hodnotu, takže obnova proběhne i při panice (a panika se
   propaguje ven). Pro `n <= 0` `GOMAXPROCS` nemění, pro `f == nil` nedělá nic.
   Nápověda: `runtime.GOMAXPROCS(n)` vrací předchozí hodnotu.
2. `ObserveParallelism(workers int) int` — spustí `workers` goroutin, každá si připočte
   k atomickému čítači souběhu (a udrží si jeho maximum) a pak počká na společné uvolnění.
   Vrací naměřené maximum. Pro `workers <= 0` nulu, po návratu žádná živá goroutina.

   Nápověda k determinismu: nejdřív nech všechny goroutiny „dorazit" (poslat do
   bufferovaného kanálu), pak je pusť zavřením druhého kanálu. Tím je souběh měřitelný
   spolehlivě, a ne podle štěstí.

### B — jádro (~35 min)

1. `CPUBound(work int) uint64` — `work` iterací čistého počítání (klidně FNV hash),
   deterministický výsledek, žádné čekání.
2. `Blocking(d time.Duration)` — simulace blokujícího volání, `d <= 0` se vrací hned.
3. `Compare(workers int) (cpu, blocking time.Duration)` — změří dobu `workers` souběžných
   CPU-bound úloh a dobu `workers` souběžných volání `Blocking(BlockingDuration)`.
   `workers <= 0` se chová jako 1.

Test kontroluje to, o čem je celá lekce: blokující část **neškáluje s počtem workerů**,
protože čekající goroutiny P uvolní, a tak osm spánků po 50 ms zabere zhruba 50 ms, ne
400. CPU-bound část naopak omezuje `GOMAXPROCS`. Zkus si `Compare` pustit ručně pod
`RunWithMaxProcs(1, …)` a porovnat čísla — to je ta lekce, kterou test nezachytí.

### C — rozšíření (~20 min)

1. `StackGrowth(depth int) int` — rekurze do hloubky `depth`, každý rámec obsahuje
   `[1024]byte`, které **musíš skutečně použít** (jinak ho kompilátor vyhodí). Vrací
   dosaženou hloubku, pro `depth <= 0` nulu. Test jde do hloubky 1000, tedy zhruba
   megabajt zásobníku — projde jen proto, že runtime zásobník zvětšuje.
2. `GoroutineCost(n int) (before, after int)` — spustí `n` goroutin, počká, až všechny
   opravdu běží, změří `runtime.NumGoroutine()` před a v tom okamžiku, a před návratem
   je všechny uklidí.
3. `BytesPerGoroutine(n int) uint64` — hrubý odhad zásobníku na goroutinu přes
   `runtime.ReadMemStats` a `StackInuse`. Když měření nic nezachytí, vrať 0.

```bash
make lesson L=49
make race L=49
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

- [ ] `make lesson L=49` prochází
- [ ] `make race L=49` prochází (žádné hlášení race detektoru)
- [ ] Umíš vysvětlit rozdíl mezi G, M a P a co z nich je omezený zdroj
- [ ] Umíš vysvětlit, proč může mít Go program víc vláken než jader
- [ ] Umíš vysvětlit, proč `conn.Read` neblokuje vlákno, ale `os.ReadFile` ano
- [ ] Umíš vysvětlit, co změnila asynchronní preempce v Go 1.14
- [ ] Umíš říct, kdy má smysl nastavovat `GOMAXPROCS` ručně

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Tohle je téma, kde jazykové modely rády mluví jako učebnice a přitom si vymýšlejí detaily
(velikosti fronty, verze, ve kterých se co změnilo). Ptej se na konkrétní tvrzení a
požaduj odkaz do `go.dev` nebo do zdrojáků `runtime/proc.go`. A když ti někdo doporučí
`runtime.Gosched()` nebo `GOMAXPROCS(1)` jako řešení závodu, máš rovnou vzorový příklad
odpovědi, kterou nesmíš vzít.

## Další čtení

1. [go.dev — Diagnostics: execution tracer](https://go.dev/doc/diagnostics)
2. [pkg.go.dev — runtime.GOMAXPROCS](https://pkg.go.dev/runtime#GOMAXPROCS)
3. [Go blog — Go 1.14 release notes: asynchronous preemption](https://go.dev/doc/go1.14#runtime)
4. [Go blog — Container-aware GOMAXPROCS](https://go.dev/blog/container-aware-gomaxprocs)
