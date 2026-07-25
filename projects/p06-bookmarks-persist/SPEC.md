# P06 — Bonus: záložky s Postgres a Redis

## Kontext

Volitelný bonus po capstone [P05](../p05-capstone/SPEC.md). Stejná doména a stejné HTTP API,
ale místo in-memory store jsou **Postgres (pgx + sqlc)** a **Redis (cache GET podle ID)**.
Router zůstává `http.ServeMux` — Chi ani jiný framework sem nepatří.

Projekt má **vlastní `go.mod`**, aby kořenový kurz zůstal bez externích závislostí.

## Akceptační kritéria (shrnutí)

- Stejné API jako P05 (`POST/GET/DELETE /bookmarks`, search, `/healthz`, `/readyz`).
- Doména v `internal/bookmark` neimportuje `net/http` ani `encoding/json`.
- Porty `app.BookmarkStore` a `app.Cache` u konzumenta; Postgres a Redis jsou adaptéry.
- Migrace + sqlc queries; generovaný kód jen v `internal/postgres/dbsqlc`.
- `pgx.ErrNoRows` / unique violation → doménové `ErrNotFound` / `ErrDuplicateURL`.
- Redis cache na `GET /bookmarks/{id}`; invalidace při delete; plnění při add/get miss.
- `/readyz` kontroluje Postgres i Redis; `/healthz` jen proces.
- Unit testy běží bez Dockeru (memstore); integrační test Postgres se skipne bez `DATABASE_URL`.

## Co záměrně neřešíme (non-goals)

- Chi, Gin, Echo, GORM, ent.
- Plné taktické DDD (domain events, event sourcing).
- Autentizace, OpenAPI, Kubernetes.
- Rate limiter v Redis (stačí cache hot path).

## Balíčky

| Balíček | Odpovědnost |
|---------|-------------|
| `internal/bookmark` | Doménový model (z P05). |
| `internal/app` | Porty `BookmarkStore`, `Cache` a dekorátor `CachedStore`. |
| `internal/memstore` | In-memory fake pro unit testy. |
| `internal/postgres` | pgx pool, migrace, sqlc adaptér. |
| `internal/postgres/dbsqlc` | Generovaný kód sqlc — neupravovat ručně. |
| `internal/rediscache` | Redis cache adaptér. |
| `internal/httpapi` | ServeMux, DTO, problem-details, middleware. |
| `internal/config` | Env včetně `DATABASE_URL`, `REDIS_URL`, `BOOKMARKS_CACHE_TTL`. |
| `cmd/bookmarks` | Wiring, migrace při startu, graceful shutdown. |

## API

Stejné jako P05. Rozdíl: `/readyz` selže, když Postgres nebo Redis neodpovídá.

## Konfigurace (env)

| Proměnná | Výchozí | Význam |
|----------|---------|--------|
| `DATABASE_URL` | *(povinná)* | Postgres DSN |
| `REDIS_URL` | `redis://127.0.0.1:6379/0` | Redis URL |
| `BOOKMARKS_CACHE_TTL` | `5m` | TTL cache záložky |
| `BOOKMARKS_ADDR` | `:8080` | listen adresa |
| `BOOKMARKS_READ_TIMEOUT` | `5s` | read timeout |
| `BOOKMARKS_WRITE_TIMEOUT` | `10s` | write timeout |
| `BOOKMARKS_SHUTDOWN_TIMEOUT` | `10s` | graceful shutdown |
| `BOOKMARKS_REQUEST_TIMEOUT` | `3s` | timeout handleru |
| `BOOKMARKS_MAX_BODY_BYTES` | `1048576` | strop těla |

## Ověření

```bash
cd projects/p06-bookmarks-persist
docker compose up -d
export DATABASE_URL='postgres://bookmarks:bookmarks@127.0.0.1:5433/bookmarks?sslmode=disable'
export REDIS_URL='redis://127.0.0.1:6380/0'
go test -count=1 -race ./...
go vet ./...
```
