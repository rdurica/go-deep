# Lekce 30 — HTTP klient, timeouty a graceful shutdown

> **Čas:** ~65 min · **Fáze:** 3 — net/http a tooling · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Vysvětlit, proč je `http.DefaultClient` v produkci chyba, a postavit klienta, který
  nikdy nevisí navěky.
- Správně uzavřít a dočíst tělo odpovědi, aby se spojení vracela do poolu.
- Napsat retry s exponenciálním backoffem a jitterem a rozhodnout, kdy retry nedělat.
- Otestovat HTTP klienta bez reálné sítě přes `httptest.NewServer`.
- Ukončit server tak, aby rozpracované požadavky doběhly, a vysvětlit, proč to
  v kontejneru potřebuješ.

## Teorie

### `http.DefaultClient` nemá timeout

```go
resp, err := http.Get(url)   // používá http.DefaultClient
```

`http.DefaultClient` má `Timeout: 0`, což znamená **žádný limit**. Když protistrana
otevře spojení a pak mlčí, tenhle `Get` visí, dokud se nezavře TCP spojení — což může
být minuty i hodiny. Goroutina drží paměť, request drží connection slot a fronta roste.
Klasický způsob, jak z pomalé závislosti udělat výpadek celé služby.

```go
client := &http.Client{Timeout: 3 * time.Second}
```

`Timeout` pokrývá celý požadavek: navázání spojení, odeslání, čekání na hlavičky
i **čtení těla**. Alternativa je per-request `context`:

```go
ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
```

Rozdíl je důležitý. `Client.Timeout` je pojistka pro celý klient a platí pro každý
požadavek stejně. Kontext se odvozuje od požadavku, který obsluhuješ, takže když klient
odejde, zruší se i odchozí volání. V praxi chceš obojí: `Timeout` jako strop a kontext
pro rušení a kratší deadline. Kratší z nich vyhraje.

### Transport a connection pooling

`http.Client` je tenká slupka, skutečnou práci dělá `http.Transport`, který drží pool
udržovaných spojení. Výchozí `MaxIdleConnsPerHost` je **2** — pro službu, která mluví
hlavně s jedním backendem, je to málo: každý požadavek nad limit navazuje nové TCP
(a TLS) spojení.

```go
transport := &http.Transport{
	Proxy:               http.ProxyFromEnvironment,
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 10,
	IdleConnTimeout:     90 * time.Second,
	TLSHandshakeTimeout: 5 * time.Second,
}
client := &http.Client{Timeout: 3 * time.Second, Transport: transport}
```

Klient s transportem vytvoř **jednou** a sdílej ho. Nový klient v každé funkci znamená
nový pool, tedy nulové znovupoužití spojení a v horším případě únik file descriptorů.

### Tělo odpovědi se musí zavřít *a dočíst*

`defer resp.Body.Close()` zná každý. Méně známé je, že spojení se vrátí do poolu jen
tehdy, když je tělo **přečtené až do konce**. Když skončíš po prvních bajtech, transport
spojení zahodí a příště naváže nové.

```go
defer func() {
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxDrain))
	_ = resp.Body.Close()
}()
```

`io.LimitReader` je tam schválně — bez něj by ses při nekonečné odpovědi zacyklil.
Ze stejného důvodu nikdy nečti neznámé tělo přes holé `io.ReadAll`:

```go
// ŠPATNĚ — protistrana rozhoduje o tom, kolik ti sežere paměti
data, err := io.ReadAll(resp.Body)

// SPRÁVNĚ — o jeden bajt víc, než chceš, ať poznáš přetečení
data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
if int64(len(data)) > maxBytes {
	return ErrBodyTooLarge
}
```

A status kód kontroluj vždycky: `err == nil` znamená jen „odpověď dorazila“, klidně
s kódem 500 a HTML chybovou stránkou v těle.

### Retry, backoff a jitter

Opakovat naslepo v těsné smyčce je nejrychlejší způsob, jak dorazit už tak přetíženou
službu. Standardní recept je exponenciální backoff — každá další prodleva je dvojnásobná
— plus **jitter**, tedy náhodné rozptýlení. Bez jitteru se všichni klienti po výpadku
sejdou v jednom okamžiku a znovu službu položí.

```go
d := base << attempt          // 100ms, 200ms, 400ms…
half := d / 2
delay := half + rand.N(half+1) // 50–100 %, každý klient jinak
```

Čekání musí respektovat kontext, jinak si retry prodlouží shutdown o desítky sekund:

```go
select {
case <-ctx.Done():
	return ctx.Err()
case <-timer.C:
}
```

Kdy retry **nedělat**:

- neidempotentní operace (`POST /payments`) — nevíš, jestli první pokus neprošel;
- chyby 4xx — `400`, `401`, `404` se opakováním nespraví;
- validační a doménové chyby.

Proto se do error modelu hodí typ, který říká „tohle je trvalé“:

```go
type PermanentError struct{ Err error }
func (e *PermanentError) Unwrap() error { return e.Err }
```

`Unwrap` je klíčový — volající pak pořád může použít `errors.Is(err, ErrNotFound)`.

### Testování klienta

`httptest.NewServer` spustí skutečný HTTP server na náhodném portu localhostu. Test tak
projde celou cestou (transport, hlavičky, tělo), ale je rychlý a nezávisí na síti:

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}))
defer srv.Close()

_, err := FetchJSON[User](ctx, srv.Client(), srv.URL)
```

`srv.Client()` je klient nakonfigurovaný pro daný server (u TLS varianty i s certifikátem).
Když testuješ timeout, nech handler čekat na `<-r.Context().Done()` — vrátí se ve chvíli,
kdy klient odejde, takže test nezdržuje pevný `time.Sleep`.

U konzumenta si definuj **malý interface**, ne `*http.Client`:

```go
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}
```

### Graceful shutdown

Když orchestrátor ruší pod, pošle procesu SIGTERM a po grace periodě SIGKILL. Bez obsluhy
signálu proces umře okamžitě uprostřed požadavků — klienti dostanou resetované spojení
a v logu máš chyby při každém deploy.

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
```

`signal.NotifyContext` vrátí kontext, který se zruší při prvním signálu. Zbytek je
`srv.Shutdown`:

```go
srvErr := make(chan error, 1)
go func() { srvErr <- srv.Serve(ln) }()

select {
case err := <-srvErr:
	return err
case <-ctx.Done():
}

shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
return srv.Shutdown(shutdownCtx)
```

Tři detaily, na kterých se to obvykle rozbije:

1. `Serve` po `Shutdown` vrací `http.ErrServerClosed`. To **není chyba** — musíš ji
   odfiltrovat přes `errors.Is`, jinak hlásíš selhání při každém korektním ukončení.
2. `Close()` utne spojení okamžitě, `Shutdown(ctx)` počká na dokončení požadavků.
   Kontext s timeoutem je strop — po jeho vypršení `Shutdown` vrátí chybu a spojení utne.
3. `Shutdown` **nesmí** dostat už zrušený kontext (typicky ten, kterým jsi shutdown
   spustil), jinak neposkytne žádnou grace periodu.

## Rozdíly proti PHP

V Symfony máš `HttpClientInterface`, timeouty v konfiguraci a retry přes dekorátor.
O ukončení procesu se stará PHP-FPM — request skončí, worker se vrátí do poolu, a když
přijde SIGTERM, FPM sám počká na dokončení běžících requestů.

```php
$response = $this->client->request('GET', $url, ['timeout' => 3.0]);
$data = $response->toArray();   // vyhodí výjimku na 4xx/5xx
```

V Go je klient obyčejný struct, který si nastavíš, a **nic ti nevyhodí výjimku**.
Status kód je jen číslo v odpovědi, o timeout se musíš postarat sám a graceful shutdown
si napíšeš:

```go
resp, err := client.Do(req)     // err jen při chybě sítě, ne při 500
if err != nil {
	return err
}
defer resp.Body.Close()
if resp.StatusCode != http.StatusOK {
	return fmt.Errorf("unexpected status %s", resp.Status)
}
```

Změna v uvažování: v PHP je životní cyklus procesu problém runtime, v Go je součástí
tvého kódu. Tři věci, které za tebe dělal FPM — timeout, uzavření zdrojů a čekání na
dokončení — jsou teď tvoje odpovědnost.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `http.Get` / `http.DefaultClient` v produkci | je to nejkratší zápis | vlastní `http.Client` s `Timeout` |
| Nový klient pro každý požadavek | zvyk vytvářet službu on-demand | jeden klient s transportem, sdílený |
| `err == nil` se bere jako úspěch | v PHP hodí klient výjimku na 5xx | zkontroluj `resp.StatusCode` |
| `io.ReadAll(resp.Body)` bez limitu | v PHP limit hlídal `memory_limit` | `io.LimitReader` a kontrola délky |
| Tělo se zavře, ale nedočte | `defer Close()` vypadá kompletně | dodrainuj přes `io.Copy(io.Discard, …)` |
| Retry i na `POST` a na 4xx | „zkusit to znovu neuškodí“ | trvalé chyby nevracet do retry smyčky |
| `os.Exit` po signálu | FPM to dřív řešil za tebe | `signal.NotifyContext` + `srv.Shutdown` |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 30`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od jednodušších
ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `Unwrap` na `PermanentError` (kód je záměrně vadný — vrací nil)

```bash
make lesson L=30 PART=1
```

Pak **`/go-deep-review 30 easy`**.

### Střední

Implementuj: `NewHTTPClient`

```bash
make lesson L=30 PART=2
```

Pak **`/go-deep-review 30 medium`**.

### Obtížný

Doplň: `RunServer` (graceful shutdown + `ErrServerClosed`)

```bash
make lesson L=30 PART=3
```

Pak **`/go-deep-review 30 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 30 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=30` (+ `make race L=30`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, co všechno pokrývá `http.Client.Timeout`
- [ ] Umíš vysvětlit, proč se tělo odpovědi musí dočíst, ne jen zavřít
- [ ] Umíš vysvětlit, k čemu je jitter v backoffu
- [ ] Umíš vyjmenovat tři situace, kdy se retry dělat nemá
- [ ] Umíš vysvětlit, proč `Serve` vrací `http.ErrServerClosed` a co s tím

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).
Nastavení transportu si nech vygenerovat, ale error model retry, hranice timeoutů
a chování při shutdownu navrhni a ověř sám.

## Další čtení

1. [pkg.go.dev — net/http.Transport](https://pkg.go.dev/net/http#Transport)
2. [pkg.go.dev — net/http.Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)
3. [pkg.go.dev — os/signal.NotifyContext](https://pkg.go.dev/os/signal#NotifyContext)
4. [Go blog — Contexts and structs](https://go.dev/blog/context-and-structs)
