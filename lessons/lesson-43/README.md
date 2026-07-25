# Lekce 43 — Mutex vs kanál

> **Čas:** ~90 min · **Fáze:** 5 — Concurrency do hloubky · **AI režim:** `JUNIOR POD REVIEW`

## Co budeš umět

- Rozhodnout, jestli daný problém patří zámku, kanálu, nebo atomice — a obhájit to.
- Navrhnout typ, jehož zero value je bezpečná pro souběžné použití a nepotřebuje konstruktor.
- Napsat single-flight cache, kde se drahý výpočet pro klíč provede právě jednou.
- Vyhnout se deadlocku z pořadí zámků a vysvětlit, proč pevné pořadí problém řeší.
- Zdůvodnit, proč `RWMutex` a `sync.Map` skoro nikdy nechceš jako první volbu.

## PHP → Go most

V PHP sdílený stav v paměti prostě neexistuje. Request má vlastní proces, po odpovědi je
paměť pryč. Když potřebuješ, aby si dva requesty něco předaly, musí to jít přes něco
mimo proces — a s tím přijde i zamykání, na které nikdo nemyslí, dokud to nerozbije data:

```php
// dva souběžné requesty, klasický lost update
$balance = $cache->get('balance');
$cache->set('balance', $balance + 100);
```

Symfony ti na to nabídne `LockFactory` nebo `Cache` s `CacheInterface::get()` a callbackem
(což je mimochodem přesně single-flight). V Go je sdílený stav **normální stav věcí**:
mapa v paměti serveru žije mezi requesty a sahá na ni každá goroutina.

```go
type Cache struct {
    mu    sync.RWMutex      // zámek stojí hned nad daty, která chrání
    items map[string]string
}
```

Změna v uvažování: v PHP je sdílený stav něco výjimečného, kolem čeho se staví
infrastruktura. V Go je to obyčejná struktura a **jediná ochrana, kterou má, je ta,
kterou napíšeš ty**. Kompilátor ti nic nepřipomene, chybu najde až race detektor.

## Teorie

### Kdy zámek a kdy kanál

Obojí umí totéž a přesto se nezaměňují:

| Použij zámek, když… | Použij kanál, když… |
|---|---|
| stav někde musí **zůstat** a jen se mění (cache, čítač, registr) | data **procházejí** z jedné goroutiny do druhé |
| operace je krátká a čistě lokální | jde o orchestraci, pořadí kroků nebo ukončení |
| jde ti o výkon jednoduché operace | chceš předat vlastnictví a přestat na data sahat |

Praktické vodítko: zeptej se, jestli řešíš **ochranu stavu**, nebo **předání vlastnictví**.
Cache je stav — kanál by z ní udělal servírovací goroutinu, která je pomalejší a hůř se
čte. Pipeline je předání — mutex by ji změnil na sdílený buffer, který nikdo nechápe.

Neplatí, že kanály jsou „idiomatičtější". Standardní knihovna používá `sync.Mutex` na
stovkách míst, protože je to na ochranu stavu prostě nejlepší nástroj.

### `sync.Mutex` v praxi

```go
type Registry struct {
    mu    sync.Mutex
    items map[string]Handler
}

func (r *Registry) Add(name string, h Handler) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.items[name] = h
}
```

Čtyři věci, které z toho udělají správný kód:

1. **Zámek je neexportované pole hned nad daty, která chrání.** Ne globální `var mu`, ne
   zámek předávaný zvenčí. Kdo vidí data, musí vidět i zámek.
2. **Zero value mutexu je odemčený mutex**, takže `var r Registry` je použitelné a
   `sync.Mutex` nikdy nevytváříš přes `new`.
3. **`defer mu.Unlock()` hned za `Lock()`.** Odemknout dřív má smysl jen tehdy, když
   následuje pomalá operace, která zámek nepotřebuje (typicky volání cizí funkce nebo
   I/O) — pak to ale piš explicitně a s komentářem.
4. **Strukturu se zámkem nikdy nekopíruj.** Metody musí mít pointer receiver a hodnotu
   nesmíš předávat kopií — kopie má vlastní zámek a chrání nic. `go vet` tuhle chybu
   hlásí (`passes lock by value`) a stojí za to jí věřit.

```go
func (r Registry) Broken() { … } // ŠPATNĚ: hodnotový receiver zkopíruje mutex
```

Granularita: jeden zámek na celou strukturu je výchozí volba. Rozdělovat ho má smysl, až
když měříš skutečnou kontenci — a pak spíš než víc zámků chceš víc nezávislých instancí
(sharding).

### `RWMutex` a kdy uškodí

`sync.RWMutex` dovolí libovolně mnoho souběžných čtenářů, ale jen jednoho zapisovatele.
Zní to jako lepší mutex zdarma. Není:

- `RLock`/`RUnlock` je dražší než `Lock`/`Unlock` — udržuje čítač čtenářů.
- Když je kritická sekce krátká (přečtení hodnoty z mapy), režie navíc převáží zisk.
- Zapisovatel musí počkat na všechny čtenáře, takže při náporu čtení může hladovět.

Vyplatí se tam, kde je čtení výrazně častější **a** kritická sekce netriviální. Pro
běžnou cache jde spíš o zvyk než o měření. Když si nejsi jistý, začni s `Mutex` a
`RWMutex` nasaď, až to ukáže benchmark.

### `sync.Once` a single-flight

`sync.Once` zaručí, že se funkce provede právě jednou, ať ji volá kolik goroutin chce.
Ostatní počkají, než ta první doběhne.

```go
var (
    once   sync.Once
    client *http.Client
)

func Client() *http.Client {
    once.Do(func() { client = newClient() })
    return client
}
```

Pozor na dvě věci: `Once` se nedá „resetovat" a `Do` počítá i volání, které panikovalo
(podruhé už se funkce nespustí). Když potřebuješ jednorázovost **na klíč**, uděláš mapu
`map[string]*sync.Once` (nebo strukturu s `Once` uvnitř) chráněnou mutexem — přesně to
je úkol B v téhle lekci a přesně tak funguje single-flight cache.

Kritická věc v návrhu: zámek cache **nesmíš držet během volání `f`**. Kdybys ho držel,
jeden pomalý výpočet zablokuje celou cache pro všechny klíče.

### Atomika

Pro čítač je mutex zbytečně těžký. Balíček `sync/atomic` má od Go 1.19 typové obálky:

```go
type Counter struct {
    n atomic.Int64 // zero value je připravená, konstruktor netřeba
}

func (c *Counter) Inc()         { c.n.Add(1) }
func (c *Counter) Value() int64 { return c.n.Load() }
```

`atomic.Int64` je lepší než `atomic.AddInt64(&c.n, 1)` nad holým `int64`: typ nejde
omylem přečíst neatomicky a je správně zarovnaný i na 32bitových platformách.

Hranice použitelnosti je ostrá. Atomika chrání **jednu hodnotu**, ne invariant mezi
dvěma. Jakmile potřebuješ „odečti tady a přičti tam a mezitím to nesmí nikdo vidět",
atomika nestačí a jsi zpátky u zámku. Pro celou strukturu, která se vyměňuje najednou,
existuje `atomic.Value` / `atomic.Pointer[T]` — hodí se na konfiguraci s hot reloadem
(lekce 44).

### Deadlock z pořadí zámků

Nejtypičtější deadlock v aplikačním kódu:

```go
// goroutina 1: Transfer("a", "b")   →   zamkne a, čeká na b
// goroutina 2: Transfer("b", "a")   →   zamkne b, čeká na a
```

Obě drží, co ta druhá potřebuje. Řešení není chytřejší zámek, ale **pevné globální
pořadí**: zámky se vždy zamykají ve stejném pořadí, například abecedně podle klíče.

```go
first, second := src, dst
if from > to {
    first, second = dst, src
}
first.mu.Lock()
defer first.mu.Unlock()
second.mu.Lock()
defer second.mu.Unlock()
```

Když se stejné pravidlo dodrží ve všech operacích včetně těch, které zamykají všechno
(sumarizace, výpis), zaklesnutí je vyloučené. A nezapomeň na hraniční případ `from == to`
— dvojí `Lock` téhož mutexu je okamžitý deadlock, protože `sync.Mutex` není rekurzivní.

Go umí deadlock detekovat jen v triviálním případě, kdy uvíznou úplně všechny goroutiny
(`fatal error: all goroutines are asleep - deadlock!`). Jakmile běží třeba jen HTTP
server, runtime mlčí a ty vidíš jen requesty, které nikdy neodpoví. Proto se deadlocky
testují časovým limitem, jak to dělá test v části C.

### `sync.Map` a proč ji skoro nikdy nechceš

`sync.Map` není „mapa se zámkem zadarmo". Je to specializovaná struktura pro dva úzké
případy popsané v její dokumentaci: klíč se zapíše jednou a pak se jen čte, nebo různé
goroutiny pracují nad disjunktními množinami klíčů. Za to platíš ztrátou typové kontroly
(`any` klíč i hodnota), horším výkonem u zápisů a horší čitelností. Výchozí volba je
`map[K]V` plus `sync.Mutex`.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Hodnotový receiver u typu se zámkem | metody se píšou bez rozmyslu | pointer receiver; `go vet` hlásí `passes lock by value` |
| Držení zámku během cizího volání | `defer mu.Unlock()` na začátku a hotovo | pusť zámek před pomalou operací, hodnotu ulož až pak |
| `RWMutex` „pro jistotu" | zní výkonněji | začni `Mutex`, `RWMutex` až podle benchmarku |
| Zamykání účtů v pořadí argumentů | v PHP se zamyká přes databázi, ne v paměti | pevné pořadí zámků podle klíče |
| `sync.Map` jako výchozí mapa | jméno slibuje víc, než dělá | `map[K]V` + `Mutex` |
| Mutex kolem čítače | „stav se musí zamknout" | `atomic.Int64` |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

### A — rozcvička (~10 min)

`Counter` — čítač bezpečný pro souběžné použití s metodami `Inc()`, `Add(n int64)` a
`Value() int64`. Podmínka: **zero value musí být rovnou použitelná**, tedy `var c Counter`
funguje bez konstruktoru. Test spustí 100 goroutin po 1000 zvýšeních a běží s `-race`.

### B — jádro (~35 min)

`Cache` — mapa chráněná zámkem:

- `NewCache() *Cache`, `Get(key) (string, bool)`, `Set(key, value)`, `Delete(key)`, `Len() int`.
- `GetOrCompute(key string, f func() string) string` — vrátí uloženou hodnotu, nebo ji
  spočítá pomocí `f`, uloží a vrátí. Klíčová podmínka: pro daný klíč se `f` zavolá
  **právě jednou**, i když `GetOrCompute` volá sto goroutin naráz. Ostatní počkají na
  výsledek toho prvního. Test to hlídá atomickým čítačem volání.
- Během volání `f` nesmíš držet zámek cache — pomalý výpočet jednoho klíče nesmí
  zablokovat práci s ostatními klíči.

### C — rozšíření (~25 min)

`Bank` — účty s převody:

- `NewBank(balances map[string]int64) *Bank`, `Balance(name) (int64, bool)`, `Total() int64`.
- `Transfer(from, to string, amount int64) error` — vrací `ErrUnknownAccount`,
  `ErrInvalidAmount` (pro `amount <= 0`) a `ErrInsufficientFunds`. Neúspěšný převod
  nesmí změnit žádný zůstatek. Převod na stejný účet je no-op s `nil` chybou.
- `Transfer` musí být odolný vůči souběžným převodům v **opačném směru** — test pouští
  všechny dvojice účtů proti sobě ve třiceti goroutinách a hlídá deadlock časovým
  limitem.
- `Total()` musí vracet konzistentní součet i uprostřed převodů. Test ho volá souběžně
  a jakákoli jiná hodnota než počáteční suma je chyba.

```bash
make lesson L=43
make race L=43
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `43`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=43` prochází
- [ ] `make race L=43` prochází
- [ ] Umíš na příkladu vysvětlit, kdy volíš zámek a kdy kanál
- [ ] Umíš říct, proč `go vet` hlásí kopírování struktury se zámkem
- [ ] Umíš vysvětlit, proč se `f` v `GetOrCompute` nevolá pod zámkem
- [ ] Umíš popsat deadlock z pořadí zámků a jak ho pevné pořadí odstraní
- [ ] Umíš jmenovat dva případy, kdy má `sync.Map` smysl, a proč to nejsou ty tvoje

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Nech si vygenerovat `Bank` a čti diff jako reviewer: v jakém pořadí se zamyká? Co dělá
`Transfer("a", "a")`? Drží se zámek během něčeho pomalého? Je někde `sync.Map`, protože
„je určená pro souběžný přístup"? Agenti tady spolehlivě generují kód, který projde
testy na šťastné cestě a zaklesne se pod zátěží.

## Další čtení

1. [pkg.go.dev — sync](https://pkg.go.dev/sync) — hlavně poznámky u `Mutex`, `Once` a `Map`
2. [pkg.go.dev — sync/atomic](https://pkg.go.dev/sync/atomic)
3. [Go blog — Share Memory By Communicating](https://go.dev/blog/codelab-share)
4. [Go Code Review Comments — Synchronous functions](https://go.dev/wiki/CodeReviewComments#synchronous-functions)
