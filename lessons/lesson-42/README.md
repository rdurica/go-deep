# Lekce 42 — `select` a timeouty

> **Čas:** ~35 min · **Fáze:** 5 — Concurrency do hloubky · **AI režim:** `JUNIOR POD REVIEW`

## Co budeš umět

- Napsat `select`, který obslouží víc kanálů najednou, a vysvětlit, proč si Go mezi
  připravenými větvemi vybírá náhodně.
- Rozhodnout mezi `time.After`, `time.NewTimer` a `time.NewTicker` a vědět, kdy který leakuje.
- Přidat ke každé blokující operaci cestu ven — timeout, deadline nebo `ctx.Done()`.
- Implementovat vzory „první výsledek vyhrává", debounce a heartbeat bez zbylých goroutin.
- Poznat `select` v cyklu, ze kterého neexistuje východ.

## Teorie

### `select` a náhodný výběr

`select` čeká na několik kanálových operací a provede tu, která je připravená jako první:

```go
select {
case v := <-a:
    fmt.Println("z a:", v)
case b <- 1:
    fmt.Println("zapsáno do b")
case <-ctx.Done():
    return ctx.Err()
}
```

Když je připravených větví víc, Go si vybere **náhodně** — ne první shora. Je to
záměrné: brání to hladovění, kdy by rychlý kanál trvale přehlušil pomalý. Důsledek pro
tebe: nikdy nespoléhej na pořadí větví jako na prioritu. Když prioritu skutečně
potřebuješ, napiš vnořený `select` s `default`:

```go
select {
case <-urgent:
    handleUrgent()
default:
    select {
    case <-urgent:
        handleUrgent()
    case <-normal:
        handleNormal()
    }
}
```

Větev `default` mění chování zásadně: `select` s `default` **nikdy neblokuje**. Když
není nic připravené, provede se `default`. Odtud neblokující zápis i čtení:

```go
select {
case ch <- v:
    sent = true
default: // buffer plný nebo nikdo nečte — zprávu raději zahodíme
    dropped++
}
```

Prázdný `select{}` bez větví je opak: blokuje navždy. Občas se použije na konci `main`
v démonu, kde má práci obstarat jen běžící goroutina, ale skoro vždy je to spíš známka,
že chybí normální čekání.

### `time.After` a proč není zdarma

`time.After(d)` vrátí kanál, do kterého po uplynutí `d` přijde čas. Vypadá jako nejhezčí
řešení timeoutu:

```go
select {
case v := <-ch:
    return v, nil
case <-time.After(2 * time.Second):
    return zero, errTimeout
}
```

Jednorázově je to v pořádku. Problém nastane v cyklu. `time.After` pokaždé alokuje nový
timer a ten **žije až do svého vypršení**, i když už na něj nikdo nečeká. Ve smyčce,
která iteruje tisíckrát za vteřinu s pětiminutovým timeoutem, ti v paměti leží stovky
tisíc živých timerů:

```go
for {
    select {
    case v := <-ch:
        process(v)
    case <-time.After(5 * time.Minute): // ŠPATNĚ — nový timer v každé iteraci
        return errIdle
    }
}
```

Správně je vytvořit timer jednou, po použití ho zastavit a resetovat:

```go
timer := time.NewTimer(5 * time.Minute)
defer timer.Stop()
for {
    select {
    case v := <-ch:
        process(v)
        if !timer.Stop() {
            <-timer.C // timer už stihl vystřelit, vyprázdníme kanál
        }
        timer.Reset(5 * time.Minute)
    case <-timer.C:
        return errIdle
    }
}
```

Ten tanec kolem `Stop`/`Reset` je otravný, ale nutný: `Reset` na timeru, který mezitím
vystřelil a má hodnotu v kanálu, by ti příště vrátil starý tik. Pravidlo zní **resetuj
jen zastavený a vyprázdněný timer**.

### `time.Tick` vs `time.NewTicker`

Totéž o patro výš. `time.Tick(d)` vrátí kanál, který tepe donekonečna, a **nedá se
zastavit** — ticker za ním žije po celou dobu běhu programu. V dokumentaci má proto
varování, že je určený jen pro globální, po celou dobu běžící tikání.

```go
ticker := time.NewTicker(time.Second)
defer ticker.Stop() // bez tohohle je to leak
for {
    select {
    case <-ctx.Done():
        return
    case <-ticker.C:
        collect()
    }
}
```

Ticker navíc tiky **zahazuje**, když je nestíháš odebírat: jeho kanál má buffer 1. To je
většinou přesně to, co chceš (nechceš dohánět zmeškané metriky), ale znamená to, že počet
tiků za daný čas není zaručený. Testy nad tickerem proto vždy piš s tolerancí.

### Timeout, deadline a `ctx.Done()`

`context` je standardní způsob, jak se rušení propaguje napříč voláními:

```go
ctx, cancel := context.WithTimeout(ctx, 2*time.Second) // relativní
defer cancel()

ctx, cancel := context.WithDeadline(ctx, start.Add(5*time.Second)) // absolutní
defer cancel()
```

Timeout je „od teď za dvě vteřiny", deadline „nejpozději v tenhle okamžik". Pro rozpočet
sdílený mezi několika voláními chceš deadline — jinak si každý krok naúčtuje vlastní
dvě vteřiny a celek trvá minutu.

`defer cancel()` piš vždy, i když timeout vyprší sám. `cancel` uvolní timer a odpojí
kontext od rodiče; bez něj rodičovský kontext drží potomka až do svého konce a `go vet`
tenhle případ hlásí jako `lostcancel`.

V `select` pak `ctx.Done()` vystupuje jako obyčejný kanál — a je to ta větev, která
odlišuje kód, který se dá zastavit, od kódu, který se dá jen zabít:

```go
select {
case out <- v:
case <-ctx.Done():
    return ctx.Err()
}
```

### Vzor „první výsledek vyhrává"

Spustíš několik variant téhož a použiješ tu, která odpoví první. Klíčové jsou tři věci:
bufferovaný kanál výsledků (aby poražení neuvízli na zápisu), `cancel()` hned po
zjištění vítěze a počkání na doběhnutí ostatních, než se vrátíš.

```go
results := make(chan result, len(fns)) // buffer = počet odesílatelů
```

Kdyby byl kanál nebufferovaný, poražená goroutina by po návratu volajícího zůstala viset
na `results <- res` — klasický leak schovaný v kódu, který se tváří jako optimalizace.

### `select` v cyklu bez východu

Nejčastější leak z celé téhle lekce vypadá nevinně:

```go
for {
    select {
    case v := <-in:
        process(v)
    }
}
```

Když `in` nikdo nezavře a nikdo nic nepošle, goroutina tu visí navždy. Cyklus se `select`
musí mít **vždy** aspoň jednu z těchto větví: `case <-ctx.Done()`, `case <-done`, nebo
čtení s `v, ok := <-in` a `return` při `ok == false`.

## Rozdíly proti PHP

V PHP jsou timeouty konfigurací někoho jiného. Nastavíš `timeout` v HTTP klientovi,
`max_execution_time` v php.ini, `lock_timeout` v databázi — a pak už jen doufáš.

```php
$response = $this->client->request('GET', $url, ['timeout' => 2.0]);
// buď přijde odpověď, nebo výjimka. Nic mezi tím neřídíš.
```

Čekat na dvě věci najednou („první z těchto dvou API, co odpoví") v PHP prakticky nejde
bez `curl_multi` nebo async knihovny. Go má na to jazykovou konstrukci:

```go
select {
case resp := <-primary:
    use(resp)
case resp := <-fallback:
    use(resp)
case <-ctx.Done():
    return ctx.Err()
}
```

Změna v uvažování: timeout přestává být nastavení a stává se **větví ve tvém kódu**. Za
každou blokující operací si od téhle lekce představuj otázku „a co když to nikdy
nepřijde?". V PHP tu otázku za tebe zodpověděl runtime tím, že proces zabil. V Go ne.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `time.After` uvnitř smyčky | vypadá jako nejjednodušší timeout | `time.NewTimer` mimo smyčku + `Stop`/`Reset` |
| `time.Tick` v dlouho žijícím kódu | kratší zápis než `NewTicker` | `NewTicker` + `defer ticker.Stop()` |
| Chybějící `defer cancel()` | „timeout přece vyprší sám" | `cancel` uvolní zdroje; `go vet` hlásí `lostcancel` |
| Pořadí větví jako priorita | v PHP se podmínky vyhodnocují shora | výběr je náhodný; prioritu si vynuť vnořeným `select` s `default` |
| Cyklus se `select` bez ukončovací větve | soustředění na šťastnou cestu | vždy `ctx.Done()`, `done`, nebo `v, ok := <-in` |
| Nebufferovaný kanál výsledků u „první vyhrává" | buffer se zdá zbytečný, když čtu jen jednou | buffer = počet odesílatelů, jinak poražení leakují |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 42`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od jednodušších
ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `TrySend`

```bash
make lesson L=42 PART=1
```

Pak **`/go-deep-review 42 easy`**.

### Střední

Implementuj: `RecvWithTimeout`

```bash
make lesson L=42 PART=2
```

Pak **`/go-deep-review 42 medium`**.

### Obtížný

Doplň: `Heartbeat`

```bash
make lesson L=42 PART=3
```

Pak **`/go-deep-review 42 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 42 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=42` (+ `make race L=42`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč `select` vybírá mezi připravenými větvemi náhodně
- [ ] Umíš říct, kdy je `time.After` v pořádku a kdy je to leak
- [ ] Umíš správně resetovat `time.Timer` a víš, proč se předtím vyprazdňuje `timer.C`
- [ ] Umíš vysvětlit, proč má kanál výsledků u „první vyhrává" buffer
- [ ] Umíš v cizím kódu najít `select` v cyklu bez ukončovací větve

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Agenti tuhle oblast píšou nebezpečně dobře vypadajícím kódem. Nejčastěji uvidíš
`time.After` ve smyčce, chybějící `timer.Stop()`, `select` bez `ctx.Done()` a
nebufferovaný kanál výsledků. Než diff přijmeš, projdi každý `select` a zeptej se: jak
tahle větev skončí, když druhá strana zmizí?

## Další čtení

1. [The Go Programming Language Specification — Select statements](https://go.dev/ref/spec#Select_statements)
2. [pkg.go.dev — time.Timer](https://pkg.go.dev/time#Timer) — hlavně poznámky u `Stop` a `Reset`
3. [Go blog — Go Concurrency Patterns: Timing out, moving on](https://go.dev/blog/go-concurrency-patterns-timing-out-and)
4. [Go blog — Contexts and structs](https://go.dev/blog/context-and-structs)
