# P02 — REST API pro správu úkolů

Projekt uzavírá fázi 3. Postavíš službu, která umí všechno, co se v lekcích 24–30
probíralo: `http.Handler`, routing `ServeMux` (metody, wildcardy), middleware chain, `context`, konfiguraci
z prostředí, `log/slog` a graceful shutdown. Bez frameworku, bez databáze, bez závislostí.

## Cíl

REST API nad in-memory úložištěm úkolů, které se chová jako produkční služba:
konzistentní chybové odpovědi, validace na hranici, strukturované logy s request ID
a čisté ukončení na SIGTERM.

## Kontrakt API

Základ: `application/json`. Chyby mají vždy tvar
`{"error":{"code":"...","message":"..."}}`.

| Metoda | Cesta | Tělo | Úspěch | Chyby |
|--------|-------|------|--------|-------|
| `GET` | `/healthz` | — | `200 {"status":"ok"}` | `405` |
| `GET` | `/tasks` | — | `200 {"tasks":[…]}` | `405` |
| `POST` | `/tasks` | `{"title","status?"}` | `201` + `Location` | `400`, `405`, `415` |
| `GET` | `/tasks/{id}` | — | `200` | `404`, `405` |
| `PUT` | `/tasks/{id}` | `{"title","status?"}` | `200` | `400`, `404`, `405`, `415` |
| `DELETE` | `/tasks/{id}` | — | `204` bez těla | `404`, `405` |

Úkol v odpovědi: `{"id","title","status","created_at","updated_at"}`, časy v RFC3339.
Povolené stavy: `todo`, `doing`, `done`; při vytvoření je `status` volitelný a výchozí
hodnota je `todo`.

Kódy chyb: `bad_request`, `validation_failed`, `not_found`, `method_not_allowed`,
`unsupported_media_type`, `internal_error`.

## Akceptační kritéria

**Doména (`internal/task`)**

- [ ] Typ `Task` s `ID`, `Title`, `Status`, `CreatedAt`, `UpdatedAt`; balíček nezná HTTP ani JSON.
- [ ] `Status` je vlastní typ s metodou `Valid()`.
- [ ] Chyby domény jsou hodnoty (`ErrNotFound`, `ErrEmptyTitle`, `ErrTitleTooLong`, `ErrInvalidStatus`) a jdou rozlišit přes `errors.Is`.
- [ ] `Store` je in-memory úložiště chráněné `sync.RWMutex` s operacemi `Create`, `Get`, `List`, `Update`, `Delete`.
- [ ] `List` vrací úkoly v pořadí vzniku a nikdy `nil`.
- [ ] Zdroj času je vyměnitelný (`NewStoreWithClock`), aby testy nezávisely na hodinách.

**HTTP vrstva (`internal/httpapi`)**

- [ ] Router používá vzory `ServeMux` (`"GET /tasks/{id}"`) a `r.PathValue("id")`.
- [ ] Chybný způsob přístupu vrací `405` s hlavičkou `Allow` a **JSON** tělem, ne prostý text.
- [ ] `POST` a `PUT` bez `Content-Type: application/json` vrací `415`.
- [ ] Tělo požadavku je omezené `http.MaxBytesReader`, dekodér má `DisallowUnknownFields`.
- [ ] Doménové chyby se mapují na status kódy v jednom místě, ne v každém handleru.
- [ ] Middleware chain: request ID → logování → recovery; skládá se funkcí `Chain`.
- [ ] Request ID se bere z hlavičky `X-Request-ID`, jinak se vygeneruje, a vrací se v odpovědi.
- [ ] Logy jsou JSON přes `slog` a obsahují `request_id`, `method`, `path`, `status`, `duration`.
- [ ] Panika v handleru skončí jako `500` v JSONu, ne jako spadlý proces.

**Vstupní bod (`cmd/api`)**

- [ ] Konfigurace z prostředí: `ADDR`/`PORT`, `LOG_LEVEL`, `READ_TIMEOUT`, `SHUTDOWN_TIMEOUT`.
- [ ] Chybná konfigurace zastaví start a vypíše **všechny** problémy najednou.
- [ ] `signal.NotifyContext` + `srv.Shutdown` s grace periodou; `http.ErrServerClosed` není chyba.
- [ ] `main` jen volá `run(ctx, getenv, out)`, takže je aplikace složitelná i v testu.

**Testy**

- [ ] Unit testy `Store` včetně validace, pořadí, časových razítek a souběhu.
- [ ] HTTP integrační testy přes `httptest.NewServer` pokrývají happy path i `400`, `404`, `405`, `415`.
- [ ] Žádný test nezávisí na pevném portu, reálné síti ani na `time.Sleep` jako synchronizaci.
- [ ] `go test -race ./...` prochází, `go vet ./...` je čistý.

## Jak ověřit

```bash
cd /home/rdurica/Projects/go-deep
go vet ./projects/p02-http-api/...
go test -count=1 -race ./projects/p02-http-api/...
```

Ruční zkouška:

```bash
cd projects/p02-http-api
PORT=8080 LOG_LEVEL=debug go run ./cmd/api

curl -i localhost:8080/healthz
curl -i -X POST localhost:8080/tasks -H 'Content-Type: application/json' -d '{"title":"napsat testy"}'
curl -i localhost:8080/tasks
curl -i -X DELETE localhost:8080/tasks/1
curl -i -X PATCH localhost:8080/tasks   # 405 + Allow
```

Nakonec pošli procesu `Ctrl+C` uprostřed požadavku a zkontroluj, že požadavek doběhl
a v logu je `server stopped`.

## Kam to dál posunout

Tohle nejsou akceptační kritéria, ale rozšíření, na kterých se dá stavět v dalších fázích:

- filtrování `GET /tasks?status=todo` a stránkování,
- `PATCH` s částečnou aktualizací (pozor na rozdíl mezi „nezasláno“ a „nastav na prázdné“),
- ETag a `If-Match` pro optimistické zamykání,
- výměna `Store` za port a SQL adaptér (lekce 33 a 35).
