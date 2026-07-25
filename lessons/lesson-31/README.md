# Lekce 31 — Checkpoint fáze 3 + projekt P02

> **Čas:** ~90 min · **Fáze:** 3 — net/http a tooling · **AI režim:** `BOILERPLATE OK`

Checkpoint nemá novou látku. Místo teorie je **recap** lekcí 24–30, cvičení je
kumulativní a na konci je bodovaná rubrika. Lekce zároveň zadává projekt
**[P02 — REST API](../../projects/p02-http-api/ACCEPTANCE.md)**.

## Co budeš umět

- Poskládat z izolovaných dílů fáze 3 jednu službu: router, middleware, kontext,
  konfiguraci, logy a shutdown.
- Navrhnout konzistentní tvar chybové odpovědi a držet ho i tam, kde ho generuje stdlib.
- Přenášet hodnoty request scope kontextem tak, aby to nebyl skrytý globál.
- Ohodnotit vlastní službu podle produkčních kritérií, ne podle „vrací to 200“.

## PHP → Go most

V Symfony dostaneš službu poskládanou frameworkem: `#[Route]` atribut, `kernel.request`
listener, `LoggerInterface` autowiringem, `.env` a parametry. Kostra existuje dřív než
tvůj první řádek kódu.

```php
#[Route('/tasks/{id}', methods: ['GET'])]
public function show(string $id): JsonResponse
{
    return $this->json($this->repository->find($id) ?? throw new NotFoundHttpException());
}
```

V Go tu kostru píšeš ty — a je to asi 150 řádků, které se vejdou do hlavy:

```go
mux.HandleFunc("GET /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
	t, err := store.Get(r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newTaskResponse(t))
})
handler := Chain(mux, RequestID, Logging(logger), Recovery(logger))
```

Změna v uvažování: v Symfony hledáš **rozšiřovací bod** frameworku, v Go hledáš
**místo v `main`**, kde se to poskládá. Nic se neděje magicky před tvým kódem — což
znamená, že nic ani nemůže tiše zmizet při upgradu.

## Recap fáze 3

### Lekce 24–27 v tabulce

| Téma | Co si musíš pamatovat |
|------|------------------------|
| `http.Handler` | jediná metoda `ServeHTTP(w, r)`; `http.HandlerFunc` udělá handler z funkce |
| `httptest` | `NewRecorder()` pro handler bez sítě, `NewServer()` pro celou cestu včetně transportu |
| Routing ServeMux | vzor je `"METODA /cesta/{param}"`, hodnota přes `r.PathValue("param")` |
| Specifičnost | vzor s metodou je specifičtější než bez ní; `/` je nejobecnější |
| Middleware | `func(http.Handler) http.Handler`; skládá se od nejvíc vnější k nejvnitřnější |
| ResponseWriter | jakmile jednou zavoláš `WriteHeader`, status už nezměníš |
| `context` | první parametr funkce, nikdy pole ve structu; ruší se, nese deadline |
| Klíč v kontextu | vlastní neexportovaný typ, ne `string` — jinak si klíče přepíšete |

### Lekce 28–30 v tabulce

| Téma | Co si musíš pamatovat |
|------|------------------------|
| Konfigurace | čte se z prostředí, převádí se explicitně, validuje se při startu |
| Sbírání chyb | `errors.Join` místo návratu první chyby |
| Tajemství | `String()` maskuje; `%#v` a `json.Marshal` ji ignorují |
| `slog` | zpráva je konstanta, proměnné jsou atributy; `With` dělá odvozený logger |
| Vlastní handler | `Enabled`/`WithGroup` deleguj, `Handle`/`WithAttrs` uprav — a vracej nový handler |
| HTTP klient | `DefaultClient` nemá timeout; tělo dočíst a zavřít; kontrolovat status |
| Retry | exponenciální backoff + jitter, respektovat kontext, netočit neidempotentní operace |
| Shutdown | `signal.NotifyContext` → `srv.Shutdown(ctx)`; `ErrServerClosed` není chyba |

### Otázky

**Proč `http.Error` nestačí produkčnímu API?** Píše prostý text a `Content-Type: text/plain`.
Klient, který parsuje JSON, na tom spadne. Potřebuješ vlastní `writeError`, který drží
jeden tvar — a použít ho i tam, kde by chybu jinak generoval `ServeMux`.

**Jak vrátit 405 v JSONu?** Zaregistruj vedle `"GET /tasks"` ještě obecnější `"/tasks"`.
Vzor bez metody je méně specifický, takže se uplatní právě tehdy, když cesta sedí
a metoda ne. Nezapomeň na hlavičku `Allow`.

**Kam patří validace?** Na hranici. Handler zkontroluje tvar (Content-Type, JSON, velikost
těla), doména zkontroluje pravidla (prázdný název, neznámý stav). HTTP vrstva pak jen
překládá `errors.Is(err, task.ErrNotFound)` na `404`.

**Proč logger konstruktorem a ne globálně?** Protože v testu chceš buffer, v produkci
stdout a v jedné službě odvozený logger s `component`. Globál ti ani jedno nedovolí.

**Co se stane, když handler zapaníkuje?** `net/http` paniku zachytí, ukončí spojení
a vypíše stacktrace — klient dostane rozbitou odpověď bez status kódu. Recovery
middleware z toho udělá řízenou `500`.

**Proč `Shutdown` nesmí dostat zrušený kontext?** Protože kontext je pro `Shutdown` strop
grace periody. Zrušený kontext znamená nulovou grace periodu, tedy totéž co `Close()`.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| 404 a 405 vracejí text, zbytek JSON | vestavěné odpovědi `ServeMux` se přehlédnou | vlastní handler pro `/` a vzory bez metody |
| Middleware zaloguje 200 u každé odpovědi | `ResponseWriter` status neprozradí | obal ho a zachyť `WriteHeader` |
| `context.WithValue(ctx, "id", v)` | v PHP je klíč prostě string | neexportovaný typ klíče |
| Store bez zámku | jeden request = jeden proces jako v PHP-FPM | `sync.RWMutex`, ověřeno `-race` |
| Validace rozstrkaná po handlerech | copy-paste z prvního endpointu | jedno místo, které mapuje chyby domény |
| `main` dělá všechno | skript se rozroste | `main` volá `run(ctx, getenv, out)` |

## Úkol

Pracuj v `exercise/`. Kumulativní úloha: kostra REST API pro poznámky, na které si
vyzkoušíš všechno z fáze 3 v malém. Projekt P02 pak dělá totéž ve velkém.

### A — rozcvička (~10 min)

`LoadConfig(getenv func(string) string) (Config, error)`:

| Klíč | Pole | Výchozí | Pravidlo |
|------|------|---------|----------|
| `ADDR` | `Addr` | `127.0.0.1:8080` | — |
| `LOG_LEVEL` | `LogLevel` | `slog.LevelInfo` | parsuj přes `slog.Level.UnmarshalText` |
| `SHUTDOWN_TIMEOUT` | `ShutdownTimeout` | `5s` | `time.ParseDuration`, musí být kladný |

Chyby posbírej přes `errors.Join` a obal `ErrInvalid`; text musí obsahovat jména
vadných klíčů.

### B — jádro (~40 min)

1. `Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler` —
   složí middleware tak, že **první uvedená je nejvíc vně**. Bez middleware vrátí `h`.
2. `RequestIDMiddleware` — vezme `X-Request-ID` z požadavku, a když chybí, vygeneruje
   nové. Uloží ho do kontextu (neexportovaný typ klíče!) a nastaví do hlavičky odpovědi.
   `RequestIDFromContext(ctx) (string, bool)` ho vrátí; pro kontext bez ID vrátí `false`.
3. `NewServer(logger *slog.Logger) http.Handler` — router s in-memory úložištěm poznámek
   a middleware chainem:

| Metoda | Cesta | Chování |
|--------|-------|---------|
| `GET` | `/healthz` | `200 {"status":"ok"}` |
| `GET` | `/notes` | `200 {"notes":[…]}`, prázdné úložiště dá `[]`, ne `null` |
| `POST` | `/notes` | `201` + `Location`, tělo `{"text":"…"}` |
| `GET` | `/notes/{id}` | `200`, nebo `404` |
| `DELETE` | `/notes/{id}` | `204`, nebo `404` |

Chyby mají vždy tvar `{"error":{"code":"…","message":"…"}}` a `Content-Type`
`application/json`. Kódy: `not_found`, `method_not_allowed`, `unsupported_media_type`,
`bad_request`, `validation_failed`, `internal_error`. Neznámá cesta je `404`, špatná
metoda `405` s hlavičkou `Allow`, chybějící nebo jiný `Content-Type` u `POST` je `415`,
rozbitý JSON `400`, prázdný nebo delší než 500 znaků text `400`.

### C — rozšíření (~25 min)

1. `RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler` — zachytí
   paniku, zaloguje ji na úrovni Error s atributem `request_id` z kontextu a odpoví
   `500` v JSONu s kódem `internal_error`.
2. `Run(ctx context.Context, cfg Config, h http.Handler, ln net.Listener) error` — obsluhuje
   listener, na zrušení kontextu udělá `Shutdown` s `cfg.ShutdownTimeout` a vrátí `nil`
   při čistém ukončení.

```bash
make lesson L=31
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Projekt P02

Zadání a akceptační kritéria:
**[projects/p02-http-api/ACCEPTANCE.md](../../projects/p02-http-api/ACCEPTANCE.md)**.

Postav REST API pro správu úkolů — `GET`/`POST` `/tasks`, `GET`/`PUT`/`DELETE`
`/tasks/{id}` — s oddělenou doménou (`internal/task`), HTTP vrstvou (`internal/httpapi`)
a vstupním bodem (`cmd/api`). Proti cvičení přibývá `PUT`, časová razítka, DTO na hranici
a logovací middleware.

```bash
make project P=02
go test -race ./projects/p02-http-api/...
```

## Sebehodnocení

Bodovanou rubriku vyplň až po projektu P02. Za každý řádek 0–2 body
(0 = neumím, 1 = s dokumentací, 2 = z hlavy).

| # | Dovednost | Body |
|---|-----------|------|
| 1 | Napsat `http.Handler` a otestovat ho přes `httptest` | |
| 2 | Zaregistrovat vzor s metodou a wildcardem a přečíst `PathValue` | |
| 3 | Napsat middleware a vysvětlit pořadí v chainu | |
| 4 | Zachytit skutečný status kód obalením `ResponseWriter` | |
| 5 | Přenést hodnotu request scope kontextem s vlastním typem klíče | |
| 6 | Načíst a zvalidovat konfiguraci s fail-fast a `errors.Join` | |
| 7 | Nastavit `slog` a zalogovat strukturovaně bez tajemství | |
| 8 | Postavit HTTP klienta s timeoutem a správně uzavřít tělo | |
| 9 | Napsat retry s backoffem, který respektuje kontext | |
| 10 | Implementovat graceful shutdown a ošetřit `ErrServerClosed` | |

| Skóre | Co dál |
|-------|--------|
| 18–20 | pokračuj na lekci 32 |
| 13–17 | zopakuj lekce, kde máš 0–1 bodu, a přepiš odpovídající část P02 |
| 8–12 | projdi znovu lekce 26, 27 a 30 a napiš cvičení znovu od nuly |
| 0–7 | vrať se na lekci 24 a projdi fázi 3 celou; bez ní nedává fáze 4 smysl |

## Ověření

- [ ] `make lesson L=31` prochází
- [ ] `go test -race ./projects/p02-http-api/...` prochází
- [ ] Umíš vysvětlit, jak dostat 405 do JSONu bez vlastního routeru
- [ ] Umíš vysvětlit, proč se klíč v kontextu nedělá jako `string`
- [ ] Umíš vysvětlit, co se stane bez recovery middleware, když handler zapaníkuje
- [ ] Umíš vysvětlit, kde končí validace tvaru a začíná validace domény

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).
Handlery a DTO si nech vygenerovat, ale hranice balíčků, error model a middleware chain
vlastníš ty. Vygenerovaný kód projdi review checklistem z playbooku.

## Další čtení

1. [pkg.go.dev — net/http.ServeMux](https://pkg.go.dev/net/http#ServeMux)
2. [Go blog — Routing Enhancements for Go 1.22](https://go.dev/blog/routing-enhancements)
3. [Mat Ryer — How I write HTTP services in Go after 13 years](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/)
