# P05 — Capstone: služba záložek URL

## Kontext

Závěrečný projekt kurzu *Go do hloubky*. Skládá dohromady doménu, souběžně bezpečné
úložiště, HTTP API (`ServeMux` se vzory), konfiguraci z prostředí a provozní návyky
(health, middleware, graceful shutdown). Inspirace je v lekcích 59–60, ale kód musí stát
sám o sobě v `projects/p05-capstone/`.

Cíl není „největší možná aplikace“, ale **čitelná, testovatelná služba se správnými
hranicemi balíčků**. Persistence mimo proces záměrně chybí — úložiště je in-memory.

## Akceptační kritéria (shrnutí)

- Vytvoření záložky z JSON těla (`url`, volitelně `title`, `tags`); server přidělí `id`.
- Normalizace URL (schéma `http`/`https`, malý host, bez `utm_*`, bez trailing slash).
- Načtení a smazání podle `id`.
- Vyhledávání přes `q`, filtr `tag`, limit a cursor.
- Liveness (`/healthz`) a readiness (`/readyz`).
- Doménový balíček neimportuje `net/http` ani `encoding/json` (hlídá test přes `go/parser`).
- Store je bezpečný pro souběh (`go test -race`).
- HTTP integrační testy přes `httptest` (happy path + 400/404/409).
- Pouze standardní knihovna.

## Co záměrně neřešíme (non-goals)

- Autentizace, autorizace, multi-tenant.
- Trvalé úložiště (PostgreSQL, Redis, soubor).
- Crawl metadat stránky, zkracování URL, redirecty.
- OpenAPI, Docker, Kubernetes manifests, CI vlastní pro projekt.
- Plný rate limiter z lekce 60 (stačí body limit, timeout, recovery, request ID).

## Balíčky

| Balíček | Odpovědnost |
|---------|-------------|
| `internal/bookmark` | Doménový model, validace, normalizace URL a tagů. Bez I/O a bez HTTP/JSON. |
| `internal/store` | In-memory úložiště s indexem podle tagu, mutex, search + cursor. |
| `internal/httpapi` | Router, DTO, problem-details, middleware chain. |
| `internal/config` | Konfigurace z prostředí (`ADDR`, timeouty, limity). |
| `cmd/bookmarks` | Wiring, `ListenAndServe`, graceful shutdown přes `signal.NotifyContext`. |

## API

| Metoda a cesta | Chování |
|----------------|---------|
| `POST /bookmarks` | Tělo `{"url","title?","tags?"}` → `201` + záložka. Neplatná data → `400`. Duplicitní normalizovaná URL → `409`. |
| `GET /bookmarks/{id}` | `200` + záložka, nebo `404`. |
| `DELETE /bookmarks/{id}` | `204` při úspěchu, nebo `404`. |
| `GET /bookmarks?q=&tag=&limit=&cursor=` | Stránka výsledků (`items`, `next_cursor`, `total`). Neplatný cursor/limit → `400`. |
| `GET /healthz` | `200` `{"status":"ok"}` — proces žije. |
| `GET /readyz` | `200` `{"status":"ready"}` — služba může přijímat provoz. |

Chybové odpovědi mají `Content-Type: application/problem+json` (RFC 7807).

## Doménová pravidla

- URL musí mít schéma `http` nebo `https` a neprázdný host.
- Normalizace: malé schéma a host, drop výchozího portu, drop fragmentu, drop `utm_*`
  query parametrů, trim trailing `/` na path, seřazený query.
- Titulek: po ořezání max 200 run; prázdný titulek se nahradí hostem z URL.
- Tagy: malá písmena, číslice, pomlčky; max 10; duplicity se sloučí; seřadí se.
- Unikátnost v úložišti je podle normalizované URL (ne podle titulu).

## Middleware (pořadí odvnějšku dovnitř)

1. Recovery — panika → `500` problem-details, proces nespadne.
2. Request ID — respektuj/přidej `X-Request-ID`.
3. Body limit — `ContentLength` + `http.MaxBytesReader`.
4. Timeout — `http.TimeoutHandler` kolem routeru.

## Konfigurace (env)

| Proměnná | Výchozí | Význam |
|----------|---------|--------|
| `BOOKMARKS_ADDR` | `:8080` | listen adresa |
| `BOOKMARKS_READ_TIMEOUT` | `5s` | `ReadHeaderTimeout` / `ReadTimeout` |
| `BOOKMARKS_WRITE_TIMEOUT` | `10s` | `WriteTimeout` |
| `BOOKMARKS_SHUTDOWN_TIMEOUT` | `10s` | graceful shutdown |
| `BOOKMARKS_REQUEST_TIMEOUT` | `3s` | timeout handleru |
| `BOOKMARKS_MAX_BODY_BYTES` | `1048576` | strop těla požadavku |

## Ověření

```bash
gofmt -w projects/p05-capstone
go test -count=1 -race ./projects/p05-capstone/...
go vet ./projects/p05-capstone/...
```
