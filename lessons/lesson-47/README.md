# Lekce 47 — errgroup a rušení přes context

> **Čas:** ~35 min · **Fáze:** 5 — Concurrency do hloubky · **AI režim:** `JUNIOR POD REVIEW`

## Co budeš umět

- Napsat vlastní `errgroup.Group` — a tím pochopit, co ta knihovna vlastně dělá.
- Postavit vzor „první chyba zruší zbytek" tak, aby po něm nezůstala žádná goroutina.
- Rozlišit **zrušení** od **chyby** a vytáhnout skutečný důvod přes `context.Cause`.
- Poznat sedm anti-vzorů práce s kontextem a opravit je.
- Použít `context.WithoutCancel` pro práci na pozadí, která má přežít HTTP odpověď.

## Teorie

### Co errgroup řeší

`golang.org/x/sync/errgroup` je zhruba sto řádků kódu, které dělají tři věci, jež
`sync.WaitGroup` neumí:

1. sbírají **první** chybu,
2. při ní zruší kontext ostatním,
3. umí omezit souběžnost (`SetLimit`).

V tomhle kurzu si ji nemůžeme přidat jako závislost, což je vlastně dobře — napsat si ji
je nejrychlejší způsob, jak přestat být uživatelem a stát se čtenářem. Jádro vypadá takhle:

```go
type Group struct {
    wg      sync.WaitGroup
    errOnce sync.Once
    err     error
    cancel  context.CancelCauseFunc
}

func (g *Group) Go(f func() error) {
    g.wg.Add(1)
    go func() {
        defer g.wg.Done()
        if err := f(); err != nil {
            g.errOnce.Do(func() {
                g.err = err
                if g.cancel != nil {
                    g.cancel(err)
                }
            })
        }
    }()
}

func (g *Group) Wait() error {
    g.wg.Wait()
    if g.cancel != nil {
        g.cancel(g.err) // i v úspěšné větvi, jinak kontext unikne
    }
    return g.err
}
```

`sync.Once` tady dělá dvě věci najednou: zaručí, že se zapamatuje první chyba, a zároveň
poskytne synchronizaci pro zápis do `g.err`. Bez ní by to byl datový závod, protože do
`g.err` může sáhnout kdokoli z běžících goroutin.

Poslední `g.cancel(g.err)` ve `Wait` vypadá zbytečně, ale zbytečný není. Kontext vytvořený
přes `WithCancel` drží odkaz v rodiči, dokud ho nezrušíš. Bez toho řádku by ti při každém
úspěšném `Wait` zůstal viset jeden kus stromu — a `go vet` na to má vlastní kontrolu
(`lostcancel`).

### SetLimit a kde se bere vstupenka

Limit je ten samý semafor z lekce 46, jen schovaný uvnitř:

```go
func (g *Group) Go(f func() error) {
    if g.sem != nil {
        g.sem <- struct{}{} // ← v volajícím, ne v goroutině
    }
    g.wg.Add(1)
    go func() {
        defer func() { <-g.sem; g.wg.Done() }()
        // ...
    }()
}
```

Kritický detail je ten komentář. Kdybys vstupenku bral **uvnitř** spuštěné goroutiny,
`Go` by se okamžitě vrátilo, spustilo by se tisíc goroutin a limit by neomezoval nic —
jen by se tisíc goroutin řadilo do fronty na semafor. Právě proto `Go` s nastaveným
limitem blokuje, a je to vlastnost, ne chyba: je to backpressure na volajícího.

### Zrušení není chyba

Když se zruší kontext, dostaneš `context.Canceled`. To je ale jen *příznak*, ne *důvod*.
Důvod byl ta úplně první chyba, kvůli které se rušilo — a ta se ti v `ctx.Err()` ztratí:

```go
ctx, cancel := context.WithCancelCause(parent)
cancel(fmt.Errorf("platební brána: %w", errTimeout))

ctx.Err()            // context.Canceled — neužitečné
context.Cause(ctx)   // platební brána: timeout — tohle chceš v logu
```

Rozdíl je vidět i v tom, jak se chyba klasifikuje. `context.Canceled` obvykle znamená
„klient odešel, nic nelogovat jako incident". `context.DeadlineExceeded` znamená „byli
jsme pomalí, to už incident je". A `context.Cause` řekne, co bylo pomalé. Splácnout to
všechno do jednoho `if err != nil { log.Error(err) }` je promarněná informace.

### Sedm způsobů, jak zkazit kontext

Tohle je seznam, který se vyplatí umět nazpaměť, protože přesně tyhle věci generují AI
nástroje a přesně na tyhle věci se dívá reviewer:

```go
// 1) kontext ve struct fieldu
type Service struct {
    ctx context.Context // ŠPATNĚ — kontext patří k volání, ne k objektu
}
// správně: func (s *Service) Do(ctx context.Context) error

// 2) context.TODO natrvalo
rows, err := db.QueryContext(context.TODO(), q) // TODO je značka "sem doplnit",
// ne hodnota do produkce. Když ji necháš, dotaz nejde zrušit.

// 3) závislosti v kontextu
ctx = context.WithValue(ctx, "db", conn) // ŠPATNĚ — DI přes kontext,
// navíc s klíčem typu string (kolize napříč balíčky). Závislost patří do parametru
// nebo do struktury. V kontextu má co dělat jen request-scoped metadata:
// trace ID, uživatel, deadline.

// 4) nil kontext
DoSomething(nil) // panika při prvním ctx.Done(). Když nemáš co předat,
// je to context.Background().

// 5) ignorovaný ctx.Err()
for _, item := range items {
    process(item) // smyčka běží dál, i když je dávno po deadline
}
// správně: if err := ctx.Err(); err != nil { return err }

// 6) chybějící defer cancel()
ctx, cancel := context.WithTimeout(parent, time.Second)
_ = cancel // timer i uzel stromu žijí až do deadline; go vet: lostcancel

// 7) request kontext v práci na pozadí
go sendWelcomeEmail(r.Context(), user) // handler skončí, kontext se zruší,
// e-mail se neodešle — a stane se to jen občas, což je nejhorší druh chyby
```

### Práce, která má přežít odpověď

Sedmý bod má od Go 1.21 přesné řešení. `context.WithoutCancel` vrátí kontext, který
**zachová hodnoty** (trace ID, jazyk, uživatele) a **zahodí zrušení i deadline**:

```go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    // ...

    bg := context.WithoutCancel(r.Context())         // hodnoty ano, zrušení ne
    bg, cancel := context.WithTimeout(bg, time.Minute) // vlastní strop, ne nekonečno
    go func() {
        defer cancel()
        if err := h.mailer.SendWelcome(bg, user); err != nil {
            h.log.Error("welcome e-mail", "err", err)
        }
    }()

    w.WriteHeader(http.StatusCreated)
}
```

Dvě věci, které se u tohohle vzoru zapomínají. Za prvé: odpojený kontext potřebuje
**vlastní timeout**, jinak jsi si vyrobil práci bez horní meze. Za druhé: pořád platí
lekce 40 — musíš vědět, jak ta goroutina skončí. Při shutdownu serveru na ni nikdo
nečeká, takže pokud na výsledku záleží, nepatří do goroutiny, ale do fronty.

## Rozdíly proti PHP

V Symfony se „zruš to, klient je pryč" prakticky neřeší. Request běží, dokud neskončí,
a nejbližší obdoba je timeout někde v konfiguraci:

```php
// max_execution_time = 30 — a víc nástrojů na zrušení nemáš
$a = $this->api->fetchUser($id);        // 2 s
$b = $this->api->fetchOrders($id);      // 3 s
$c = $this->api->fetchInvoices($id);    // 4 s
// celkem 9 s sériově; když třetí volání selže, první dvě už proběhla zbytečně
```

V Go tyhle tři dotazy pustíš najednou a zrušení je součást API:

```go
g, ctx := WithContext(r.Context())
var a User
var b []Order
var c []Invoice
g.Go(func() error { var err error; a, err = fetchUser(ctx, id); return err })
g.Go(func() error { var err error; b, err = fetchOrders(ctx, id); return err })
g.Go(func() error { var err error; c, err = fetchInvoices(ctx, id); return err })
if err := g.Wait(); err != nil { // 4 s místo 9 — a při chybě rovnou míň
    return err
}
```

Co se mění v uvažování: `context.Context` není „další parametr, který se všude vláčí".
Je to **strom zrušení**. Když se zruší uzel, zruší se celý podstrom pod ním. Tvoje práce
je ten strom postavit tak, aby kopíroval strukturu volání — a nikdy ho neobejít.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Kontext jako pole struktury | reflex z DI: „závislost dám do konstruktoru" | `ctx` je vždy první parametr metody |
| `context.WithValue` na závislosti | v Symfony je kontejner všude | závislosti do parametrů, do kontextu jen request metadata |
| Chybějící `defer cancel()` | zdá se, že po návratu se to uklidí samo | vždy `defer cancel()` hned po `WithTimeout` |
| `r.Context()` pro úlohu na pozadí | „mám ho po ruce" | `context.WithoutCancel` + vlastní timeout |
| Smyčka ignorující `ctx.Err()` | v PHP smyčku nikdo nepřeruší | kontrola na začátku každé iterace |
| Logování `context.Canceled` jako chyby | vypadá to jako selhání | odlišit `Canceled` (klient odešel) od `DeadlineExceeded` |
| `Wait()` bez závěrečného `cancel()` | „vrátil jsem nil, není co rušit" | `cancel` i v úspěšné větvi, jinak uzel stromu unikne |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 47`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od jednodušších
ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `WithContext`

```bash
make lesson L=47 PART=1
```

Pak **`/go-deep-review 47 easy`**.

### Střední

Implementuj: `Go`, `Wait`

```bash
make lesson L=47 PART=2
```

Pak **`/go-deep-review 47 medium`**.

### Obtížný

Doplň: `RunAll`

```bash
make lesson L=47 PART=3
```

Pak **`/go-deep-review 47 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 47 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=47` (+ `make race L=47`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč `Go` s limitem blokuje volajícího
- [ ] Umíš vysvětlit rozdíl mezi `ctx.Err()` a `context.Cause(ctx)`
- [ ] Umíš vyjmenovat aspoň pět anti-vzorů práce s kontextem
- [ ] Umíš vysvětlit, co `context.WithoutCancel` zachová a co zahodí
- [ ] Umíš vysvětlit, proč `Wait` ruší kontext i po úspěchu

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Napiš si nejdřív acceptance test na `Group` a teprve pak nech agenta navrhnout
implementaci. V jeho kódu hledej přesně tři věci: bere vstupenku semaforu ve volajícím,
nebo v goroutině? Chrání zápis do `g.err`? Ruší kontext i v úspěšné větvi? Vygenerované
verze skoro vždy zvládnou první dvě a zapomenou na třetí.

## Další čtení

1. [pkg.go.dev — golang.org/x/sync/errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup) — zdrojový kód se vyplatí přečíst celý
2. [Go blog — Go Concurrency Patterns: Context](https://go.dev/blog/context)
3. [pkg.go.dev — context.WithoutCancel](https://pkg.go.dev/context#WithoutCancel)
4. [pkg.go.dev — context.Cause](https://pkg.go.dev/context#Cause)
