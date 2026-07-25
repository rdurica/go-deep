# Lekce 60 — Checkpoint závěrečný: hardening, retrospektiva a export kurzu

> **Čas:** ~90 min · **Fáze:** 8 — Capstone · **AI režim:** `TECH LEAD`

## Co budeš umět

- Doložit u každé z osmi fází kurzu, co umíš, kde jsi to dělal a jak si to kdykoli ověříš.
- Zpevnit HTTP službu do produkčního stavu: limity, timeouty, hlavičky, logy bez PII, health, shutdown.
- Vyhodnotit připravenost projektu k nasazení podle checklistu místo podle pocitu.
- Udržet si znalost Go i po kurzu a vědět, co číst a kam přispívat dál.
- Udělat z tohohle repozitáře kurz pro ostatní.

## PHP → Go most

V Symfony přichází „produkční připravenost" z velké části s frameworkem a s tím, co má
tvoje firma nastavené: rate limiter je bundle, security hlavičky přidává nginx,
graceful reload dělá PHP-FPM, logy formátuje Monolog.

```yaml
framework:
    rate_limiter:
        anonymous_api: { policy: 'token_bucket', limit: 100, rate: { interval: '1 minute' } }
```

V Go je proces **tvůj**. Nikdo za tebe neukončí požadavky při deploy, nikdo neomezí
velikost těla a nikdo neschová heslo v logu. Zato všechno vidíš na jednom místě
a v jednom jazyce:

```go
srv := &http.Server{
	Handler:           Harden(mux, opts),
	ReadHeaderTimeout: 5 * time.Second,
	WriteTimeout:      10 * time.Second,
}
// a shutdown je tvoje funkce, ne konfigurace
```

Poslední přenos návyku celého kurzu: **produkční vlastnosti jsou kód, který napíšeš
a otestuješ**, ne konfigurace, kterou zdědíš. Proto je taky můžeš otestovat přes
`httptest` a mít je zelené v CI.

## Recap

Kurz měl osm fází. Tady je celá mapa: co umíš, kde to bylo a čím si to ověříš, když si
za rok nebudeš jistý.

| Fáze | Co umíš | Kde to bylo | Jak si to ověříš |
|------|---------|-------------|------------------|
| 0 — Setup a mentální reset | moduly, workspace, rozdíl mezi kompilovaným a interpretovaným během | 01–02 | `go mod init`, `go build ./...` na prázdném projektu |
| 1 — Jazyk a paměťový model | zero values, slices a aliasing, mapy, interfacy, chyby jako hodnoty, generics | 03–18 | přepiš `Slug` a `errors.Is/As` z hlavy, spusť testy lekcí 07 a 14 |
| 2 — Idiomatický Go | pojmenování, konstruktory, error texty, čtení stdlib | 19–23 | projdi cizí PR a najdi tři neidiomatická místa s odkazem na Code Review Comments |
| 3 — net/http a tooling | handler, ServeMux (vzory), middleware, context, konfigurace, slog, shutdown | 24–31 | napiš službu se dvěma endpointy a testy přes `httptest` za 30 minut |
| 4 — Architektura | `internal/`, porty a adaptéry, doménové typy, validace na hranici | 32–39 | ukaž v P03 a P05 test, který hlídá, že doména neimportuje `net/http` |
| 5 — Concurrency | goroutiny a leaky, kanály, `select`, mutex vs kanál, pipeline, worker pool, `errgroup` | 40–50 | spusť P04 s `-race` a vysvětli lifetime každé goroutiny |
| 6 — Production Go | verzování, benchmarky a fuzz, pprof, generics v API, kontejnery a health | 51–55 | najdi alokaci přes `testing.AllocsPerRun` a sniž ji |
| 7 — Inženýrství v době AI | spec-first, ADR, strukturované review, pairing protokol, měření review | 56–58 | pusť svůj linter z lekce 57 na cizí diff a spočítej precision |
| 8 — Capstone | kompletní služba: doména, store, HTTP, konfigurace, provoz | 59–60 | `go test -race ./projects/p05-capstone/...` |

Deset otázek, na které bys po kurzu měl umět odpovědět bez hledání:

1. Proč je zero value designový cíl a které tři typy ze stdlib to využívají?
2. Kdy `append` alokuje nové pole a co to udělá s druhým slicem nad stejným polem?
3. Čím se liší `errors.Is` od `errors.As` a kdy potřebuješ vlastní `Unwrap`?
4. Proč se interface definuje u konzumenta a jak to vypadá v testu?
5. Co přesně dělá `defer` s argumenty a v jakém pořadí se odloží?
6. Jak poznáš goroutine leak a jak ho ověříš v testu?
7. Proč `context.Context` nepatří do structu a co se stane, když tam je?
8. Co dělá `http.TimeoutHandler` a proč to nestačí místo `ReadHeaderTimeout`?
9. Proč je cursor lepší než offset a co dělat s neznámým cursorem?
10. Které tři informace musí být v promptu pro Go, aby výsledek nebyl Java?

### Hardening capstone

Produkční vrstva, kterou dnes v úkolu postavíš a v P05 zapojíš:

| Vlastnost | Proč | Jak v Go |
|-----------|------|----------|
| Rate limiting | ochrana před jedním klientem, který sežere kapacitu | token bucket na klíč, úklid starých kbelíků |
| Limit velikosti těla | `io.ReadAll` na neomezené tělo je OOM | kontrola `ContentLength` + `http.MaxBytesReader` |
| Timeouty | pomalý klient drží goroutinu i spojení | `ReadHeaderTimeout`, `WriteTimeout`, `http.TimeoutHandler` |
| Bezpečnostní hlavičky | levná obrana proti sniffingu a clickjackingu | `nosniff`, `DENY`, `no-referrer`, restriktivní CSP |
| Strukturované logy bez PII | log je datový zdroj, ne příběh; e-mail v logu je incident | `slog` s atributy, whitelist polí, nikdy celé tělo |
| Metriky a health | orchestrátor potřebuje odpověď, ne dojem | `/healthz` (žije) a `/readyz` (může přijímat provoz) |
| Recovery | jedna panika nesmí shodit celý proces | middleware s `recover()` a 500 |
| Graceful shutdown | deploy nesmí utnout běžící požadavky | signál → `srv.Shutdown(ctx)` s timeoutem |

Pořadí v chainu není libovolné. Recovery je **nejvíc venku** (musí zachytit i paniku
z timeout handleru), bezpečnostní hlavičky hned za ním (aby je měla i chybová odpověď),
pak rate limit (odmítni dřív, než uděláš práci), pak limit těla a nakonec timeout kolem
vlastního handleru.

### Retrospektiva: jak si znalost udržet

- **Piš Go aspoň hodinu týdně**, i kdyby to byl nástroj na vlastní logy. Jazyk, který
  čteš, ale nepíšeš, se rozpadne za tři měsíce.
- **Manual rewrite drill** z lekce 58 jednou týdně na kritické funkci.
- **Čti stdlib**, ne blogy: `net/http`, `io`, `errors`, `sync` a `sort` jsou učebnice
  psaná lidmi, kteří jazyk navrhli.
- **Sleduj release notes** každého Go vydání (vychází dvakrát ročně) a vyzkoušej jednu
  novinku na malém projektu.
- **Přispěj do OSS**: začni dokumentací a testy v knihovně, kterou používáš, pak issue
  označené `help wanted`. Pro Go samotné je vstupní bod
  [contribution guide](https://go.dev/doc/contribute); první příspěvek klidně může být
  oprava překlepu — proces se naučíš na něčem levném.

### Export kurzu

Repozitář je zároveň kurz. Materiál k tomu je v `course/` a postup v
[docs/course-export.md](../../docs/course-export.md):
`course/lesson-map.md` je osnova, `lessons/*` jsou lekce s testy, `projects/*` jsou
hodnocené projekty s `ACCEPTANCE.md` a `docs/ai-playbook.md` je pravidlo pro práci
s agenty. Checkpointy (18, 23, 31, 39, 50, 55, 60) jsou přirozené body hodnocení; když
kurz vedeš pro tým, mají se odevzdávat právě ony a projekty P01–P05.

Před publikací zkontroluj tři věci: řešení v `solutions/` možná chceš schovat do privátní
větve, `LICENSE` odpovídá tomu, jak chceš materiál šířit, a `make check` prochází na
čisté kopii repozitáře.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Rate limiter bez úklidu kbelíků | funguje v testu, mapa roste v produkci | `Cleanup` podle poslední aktivity, spouštěný tickerem |
| `time.Now()` uvnitř limiteru | „čas přece nejde injektovat" | `now func() time.Time` v konstruktoru, test je pak deterministický |
| Timeout jako nejvzdálenější middleware | zdá se, že má chránit všechno | timeout patří kolem handleru, recovery a hlavičky nad něj |
| Celé tělo požadavku do logu | zvyk z debugování v PHP | logovat jen ID, cestu, stav a délku; nikdy PII |
| `/healthz` kontrolující databázi | záměna liveness a readiness | liveness = žiju, readiness = mohu obsluhovat |
| Audit „myslím, že jsme ready" | checklist není v kódu | vyhodnoť ho funkcí a nech kritické položky blokovat |

## Úkol

Pracuj v `exercise/`. Cvičení je **kumulativní** — kombinuje souběžnost (fáze 5),
`net/http` a middleware (fáze 3), doménové chyby (fáze 1) a produkční disciplínu (fáze 6).

### A — token bucket (~20 min)

`NewRateLimiter(capacity int, refill time.Duration, now func() time.Time) *RateLimiter`.
Neplatná kapacita (`< 1`) se opraví na `1`, `refill <= 0` na jednu sekundu, `now == nil`
znamená `time.Now`.

- `Allow(key string) bool` — každý klíč má vlastní kbelík. Nový klíč začíná plný. Před
  odebráním tokenu doplň `elapsed / refill` tokenů (nikdy nad kapacitu) a posuň značku
  posledního doplnění **o vyčerpaný násobek** `refill`, ne na aktuální čas — jinak by se
  zbytek času ztrácel. Prázdný kbelík vrací `false`.
- `Cleanup(idle time.Duration) int` — zahodí kbelíky, které se nepoužily aspoň `idle`
  (podle času posledního `Allow`), a vrátí jejich počet.
- `Len() int` — počet sledovaných klíčů.

Test běží s `-race` a osm goroutin se pere o 500 tokenů; musí projít **přesně** 500.

např. `Allow("a")` ×4 při kapacitě 3 → `true, true, true, false`

### B — middleware chain (~35 min)

`Harden(h http.Handler, opts HardenOptions) http.Handler` složí (od vnějšku dovnitř):

1. **recovery** — `recover()`, `500` a tělo `internal server error\n`; panika nesmí uniknout,
2. **bezpečnostní hlavičky** — všechny z `SecurityHeaders`, i u chybových odpovědí,
3. **rate limiting** — jen když `opts.Limiter != nil`; klíč z `opts.KeyFunc`, a když je
   `nil`, z `r.RemoteAddr` bez portu; odmítnutí je `429` a tělo `too many requests\n`,
4. **limit těla** — jen když `opts.MaxBodyBytes > 0`; `ContentLength` nad limit je `413`
   a tělo `request body too large\n`, jinak se tělo obalí do `http.MaxBytesReader`,
   aby se ořízlo i při neznámé délce,
5. **timeout** — jen když `opts.Timeout > 0`; použij `http.TimeoutHandler` se zprávou
   `timeout` (odpoví `503`).

Každá vlastnost má vlastní test přes `httptest`; testy nesmí spoléhat na reálný čas
kromě samotného timeoutu.

např. `Harden` + limiter kapacita 2 → 3. požadavek `429`

### C — audit a pokrok (~25 min)

1. `AuditReport(checks []Check) (Report, error)`:
   - prázdný vstup → `ErrNoChecks`, kontrola bez ID → `ErrInvalidCheck`, duplicitní ID →
     `ErrDuplicateCheck`,
   - `Total`, `Passed`, `CriticalFailed` (nesplněné kritické), `Score = Passed / Total`,
   - `Missing` — ID nesplněných kontrol, seřazená,
   - `Ready` — `true`, jen když není žádný kritický nedodělek **a** `Score >= 0.8`.
2. `Coverage(lessons []LessonResult) Summary`:
   - `Total`, `Passed`, `Percent` (0–100; prázdný vstup dá 0),
   - `ByPhase` — vždy inicializovaná mapa fáze → `PhaseStat`,
   - `WeakestPhase` — fáze s nejnižším poměrem splněných; při shodě nižší číslo fáze,
     u prázdného vstupu `0`.

např. `AuditReport` (8/10, 1 kritický chybí) → `Ready=false`, `Score=0.8`

```bash
make lesson L=60
make race L=60
```

Až budeš hotový, porovnej se `solutions/` (spoiler). Pak zapoj `Harden` do
`projects/p05-capstone/` a projdi `ACCEPTANCE.md`.

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `60`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=60` i `make race L=60` prochází
- [ ] `go test -race ./projects/p05-capstone/...` prochází
- [ ] `ACCEPTANCE.md` projektu P05 je odškrtaný celý
- [ ] Umíš vysvětlit pořadí middleware v `Harden` a co se rozbije při prohození
- [ ] Umíš vysvětlit rozdíl mezi liveness a readiness na příkladu své služby
- [ ] Umíš říct, proč se čas do limiteru injektuje, a co by jinak bylo v testu flaky

## Sebehodnocení

Za každou položku 0–2 body (0 = neumím, 1 = s dokumentací, 2 = z hlavy a umím obhájit).

| # | Dovednost | Fáze | Body |
|---|-----------|------|------|
| 1 | Zero values, slices a aliasing, mapy | 1 | |
| 2 | Chyby jako hodnoty, `%w`, `Is`/`As` | 1 | |
| 3 | Interfacy u konzumenta, kompozice | 1–2 | |
| 4 | Idiomatické pojmenování a struktura balíčků | 2 | |
| 5 | `net/http`, ServeMux (vzory), middleware, `httptest` | 3 | |
| 6 | `context`, konfigurace, `slog`, graceful shutdown | 3 | |
| 7 | Porty a adaptéry, doména bez frameworku | 4 | |
| 8 | Validace na hranici a doménové chyby | 4 | |
| 9 | Goroutiny, kanály, `select`, lifetime | 5 | |
| 10 | Mutex vs kanál, `-race`, worker pool, `errgroup` | 5 | |
| 11 | Benchmarky, fuzz, pprof, alokace | 6 | |
| 12 | Spec-first, ADR, prompting pro Go | 7 | |
| 13 | Strukturované review a vlastní linter nad AST | 7 | |
| 14 | Hardening a provoz služby | 8 | |
| 15 | Capstone P05 kompletní a zelený | 8 | |

**Vyhodnocení (max 30):**

| Skóre | Co s tím |
|-------|----------|
| 26–30 | Umíš Go na úrovni, se kterou můžeš vést cizí Go kód. Jdi do OSS a do lekce „Co dál". |
| 20–25 | Solidní základ s dírami. Zopakuj lekce u položek s 0–1 bodem a znovu odevzdej P05. |
| 13–19 | Chybí ti jedna nebo dvě fáze celé. Podívej se na `Coverage` ze svého cvičení — nejslabší fáze projdi znovu, včetně cvičení. |
| 0–12 | Vracíš se na fázi 1. Není to prohra: kurz je 80 hodin práce a přeskakovat se nedá. Projdi znovu lekce 03–18 a udělej P01. |

## Co dál

1. **Dokonči a zveřejni P05.** Přidej README s příkazy, `Dockerfile` a jeden skutečný
   deployment. Služba, kterou nikdo nespustil, nic nedokazuje.
2. **Volitelný bonus P06** — nahraď in-memory store PostgreSQL (sqlc + pgx) a Redis
   cache: [`projects/p06-bookmarks-persist`](../../projects/p06-bookmarks-persist/README.md).
   Je to ADR z lekce 56 v praxi; Chi nepotřebuješ.
3. **Přepiš jeden vlastní PHP nástroj do Go.** Ideálně CLI nebo malý démon. Rozsah na
   jeden víkend, ne na půl roku.
4. **Přečti si `net/http` a `errors` celé.** Ne dokumentaci, zdrojáky. Po tomhle kurzu
   na to máš.
5. **Zavěs si vlastní linter z lekce 57 do CI** a nech ho hlídat i cizí PR.
6. **Přispěj do OSS.** Knihovna, kterou používáš, má issues označené `help wanted`.
7. **Veď kurz dál.** Repozitář jde exportovat podle
   [docs/course-export.md](../../docs/course-export.md) — nejrychleji se učí ten, kdo učí.

## AI režim

`TECH LEAD` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Od téhle chvíle je režim trvalý: spec a acceptance testy tvoje, implementaci smí psát
agent, review a architektonická rozhodnutí vlastníš ty a umíš je obhájit bez modelu.

## Další čtení

1. [net/http — dokumentace balíčku](https://pkg.go.dev/net/http)
2. [log/slog — dokumentace balíčku](https://pkg.go.dev/log/slog)
3. [Go blog — Contexts and structs](https://go.dev/blog/context-and-structs)
4. [Go — contribution guide](https://go.dev/doc/contribute)
