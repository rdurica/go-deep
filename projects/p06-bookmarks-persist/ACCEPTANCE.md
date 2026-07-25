# P06 — Bonus: záložky s Postgres a Redis

## Cíl

Po P05 vyměnit in-memory store za **Postgres (sqlc + pgx)** a přidat **Redis cache**
na `GET /bookmarks/{id}`, přičemž HTTP API a doména zůstanou stejné. Bez Chi.

## Akceptační kritéria

- [ ] Vlastní `go.mod` v `projects/p06-bookmarks-persist/` — kořenový modul kurzu
      zůstává bez `require`.
- [ ] `internal/bookmark` — doména z P05; **neimportuje** `net/http` ani `encoding/json`.
- [ ] `internal/app` — porty `BookmarkStore` a `Cache` + dekorátor `CachedStore`.
- [ ] `internal/memstore` — in-memory fake; unit testy HTTP běží bez Dockeru.
- [ ] `internal/postgres` — pgx pool, migrace, mapování `ErrNoRows` → `ErrNotFound`,
      unique violation → `ErrDuplicateURL` / `ErrDuplicateID`.
- [ ] `sql/` + `sqlc.yaml` — queries `Insert`/`Get`/`Delete`/`Search`; generovaný kód
      jen v `internal/postgres/dbsqlc`.
- [ ] `internal/rediscache` — cache `bookmark:{id}` s TTL; invalidace při delete.
- [ ] `internal/httpapi` — `ServeMux` se vzory, stejné routy jako P05, problem-details,
      middleware (request ID, recovery, body limit, timeout).
- [ ] `/healthz` nekontroluje závislosti; `/readyz` pingá Postgres i Redis.
- [ ] `cmd/bookmarks` — wiring, migrace při startu, graceful shutdown.
- [ ] `docker compose up -d` nastartuje Postgres (host `5433`) a Redis (host `6380`).
- [ ] `go test -count=1 -race ./...` a `go vet ./...` procházejí
      (s `DATABASE_URL` i bez ní — integrační test se skipne).

## Jak ověřit

```bash
cd projects/p06-bookmarks-persist
docker compose up -d
export DATABASE_URL='postgres://bookmarks:bookmarks@127.0.0.1:5433/bookmarks?sslmode=disable'
export REDIS_URL='redis://127.0.0.1:6380/0'
go test -count=1 -race ./...
go vet ./...
```

Ruční smoke:

```bash
go run ./cmd/bookmarks
curl -sS -X POST localhost:8080/bookmarks \
  -H 'content-type: application/json' \
  -d '{"url":"https://go.dev/blog/?utm_source=x","title":"Go Blog","tags":["go","blog"]}'
curl -sS localhost:8080/readyz
```
