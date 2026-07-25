# P04 — Worker pool s backpressure

Projekt fáze 5. Zadává ho a ověřuje [lekce 50](../../lessons/lesson-50/README.md).

## Cíl

Postavit **znovupoužitelnou komponentu** pro dávkové zpracování, ne jednorázový skript.
Tedy generický worker pool, který má omezenou souběžnost, tlačí zpět na producenta,
umí se korektně ukončit dvěma různými způsoby, sbírá výsledky i chyby a měří sám sebe.

Rozdíl proti cvičením v lekcích 46–49 je v tom, že tady musí všechno fungovat **naráz**.
Většina chyb v souběžném kódu vzniká právě na styku dvou správně napsaných částí.

## Struktura

```
projects/p04-worker-pool/
  ACCEPTANCE.md
  pool/                 # generický worker pool + metriky
    pool.go
    pool_test.go
  ratelimit/            # token bucket s injektovanými hodinami
    ratelimit.go
    ratelimit_test.go
  cmd/crawl/            # ukázkové použití nad stdin
    main.go
    main_test.go
```

Import: `github.com/rdurica/go-deep/projects/p04-worker-pool/pool`

## Akceptační kritéria

### Balíček `pool`

- [ ] `New(ctx, cfg)` validuje konfiguraci a vrací `ErrInvalidWorkers`,
      `ErrInvalidQueueSize` a `ErrNilHandler` — ne paniku.
- [ ] Pool je generický: `Config[T, U]` a `Result[T, U]`, žádné `any` v API.
- [ ] Souběžnost je omezená na `cfg.Workers`. Metrika `MaxInFlight` to prokazuje,
      test to měří atomickým maximem.
- [ ] `Submit(ctx, v)` **blokuje**, když je vstupní fronta plná (backpressure), a vrací
      `ctx.Err()`, když se během čekání zruší kontext.
- [ ] `Submit` po `Close` vrací `ErrClosed` a nepanikuje (žádný zápis do zavřeného kanálu).
- [ ] `Close` je idempotentní a znamená **graceful drain**: zařazené úlohy se dokončí.
- [ ] Zrušení kontextu znamená **okamžité zastavení**: zbytek fronty se zahodí a kanál
      výsledků se zavře, i když je fronta plná.
- [ ] Kanál výsledků zavírá vlastník až poté, co do něj přestali psát všichni workeři.
- [ ] Chyba jedné úlohy je součástí `Result`, ne návratová hodnota poolu.
- [ ] `Stats()` vrací přijato, zpracováno, chyb, maximální souběh a dobu běhu.
- [ ] `Collect(ctx, cfg, inputs)` vrací výsledky ve **stejném pořadí** jako vstup.
- [ ] Po dočtení kanálu výsledků nezůstane žádná goroutina navíc.

### Balíček `ratelimit`

- [ ] Token bucket bez závislostí, čas injektovaný přes `Clock`, takže testy nespí.
- [ ] `Wait(ctx)` je zrušitelný a nikdy nevisí po zrušení kontextu.

### `cmd/crawl`

- [ ] Přečte úlohy ze stdin, přeskočí prázdné řádky a komentáře začínající `#`.
- [ ] Zpracuje je s omezenou souběžností a omezenou rychlostí.
- [ ] Vypíše výsledky **v pořadí vstupu** a na konci report s metrikami.
- [ ] Nic nestahuje ze sítě — testy nepotřebují internet.
- [ ] `SIGINT`/`SIGTERM` přes `signal.NotifyContext` běh korektně ukončí.
- [ ] Nenulový exit code, když některá úloha selhala.

### Testy

- [ ] Všechno prochází s `go test -race -count=3 ./...`.
- [ ] Testy pokrývají: dodržení limitu souběžnosti, zrušení kontextem, backpressure při
      zaplněné frontě, nulový leak goroutin, korektnost a pořadí výsledků, sběr chyb.
- [ ] Žádný test není závislý na `time.Sleep` jako synchronizaci — čeká se na kanálech,
      časové meze jsou velkorysé.

## Jak ověřit

```bash
cd projects/p04-worker-pool
gofmt -l .
go vet ./...
go test -race -count=3 ./...

# ruční zkouška
printf 'jedna\ndve tri\n# komentar\n!chybna\n' | go run ./cmd/crawl -workers=2 -queue=4
```

## Na co si dát pozor

| Past | Jak se projeví |
|------|----------------|
| Kanál výsledků zavírá worker | `panic: send on closed channel` pod zatížením |
| `Submit` bez ochrany proti `Close` | `panic: send on closed channel`, jednou z tisíce |
| Velký buffer „ať to nebrzdí" | backpressure zmizí, paměť roste s velikostí dávky |
| Zrušení jen ve workerovi, ne v `Submit` | producent visí na plné frontě navždy |
| Metriky přes obyčejné `int` | `-race` to nahlásí, protože do nich píše N workerů |
| Pořadí výsledků z kanálu | výsledky lezou v pořadí dokončení; potřebuješ index |
