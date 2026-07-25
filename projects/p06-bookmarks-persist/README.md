# P06 — Bonus: produkční backend nad P05

Volitelný projekt po capstone. Stejná služba záložek jako [P05](../p05-capstone/),
ale s **PostgreSQL (sqlc + pgx)** a **Redis cache**. Router zůstává `http.ServeMux`
(viz lekce 25 — Chi tu nepotřebuješ).

**Čas:** ~3–5 h · **AI režim:** `TECH LEAD` (viz [docs/ai-playbook.md](../../docs/ai-playbook.md))

Spec a checklist: [SPEC.md](SPEC.md), [ACCEPTANCE.md](ACCEPTANCE.md).

## Proč samostatný modul

Kořenový kurz drží `go.mod` bez externích závislostí. P06 má vlastní modul
`github.com/rdurica/go-deep/projects/p06-bookmarks-persist`, aby se pgx/redis/sqlc
nedostaly do lekcí 01–60.

## Rychlý start

```bash
cd projects/p06-bookmarks-persist
docker compose up -d

export DATABASE_URL='postgres://bookmarks:bookmarks@127.0.0.1:5433/bookmarks?sslmode=disable'
export REDIS_URL='redis://127.0.0.1:6380/0'

go test -count=1 -race ./...
go run ./cmd/bookmarks
```

Host porty `5433` / `6380` jsou záměrně mimo výchozí 5432/6379, ať se nesrazíš
s lokálním Postgres/Redis.

Po změně SQL:

```bash
sqlc generate   # vyžaduje sqlc CLI (go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest)
```

---

## Průvodce ve 4 krocích

### 1. Porty (hranice z P05)

V P05 HTTP sahalo přímo na konkrétní `*store.Store`. Tady je interface u konzumenta:

```go
// internal/app/ports.go
type BookmarkStore interface {
    Add(ctx context.Context, b bookmark.Bookmark) error
    Get(ctx context.Context, id string) (bookmark.Bookmark, error)
    Delete(ctx context.Context, id string) error
    Search(ctx context.Context, q Query) (Page, error)
    Ready(ctx context.Context) error
}
```

`internal/memstore` implementuje port pro unit testy — stejně rychlé a bez Dockeru
jako P05. Doména (`internal/bookmark`) se nemění.

**Ověření:** `go test ./internal/memstore/... ./internal/httpapi/...`

### 2. Migrace + sqlc

Schéma žije vedle adaptéru:

- `internal/postgres/migrations/001_bookmarks.sql`
- queries v `sql/queries/bookmarks.sql`
- `sqlc.yaml` generuje `internal/postgres/dbsqlc` (balíček `dbsqlc`)

sqlc ti dá typované metody (`InsertBookmark`, `GetBookmark`, …). Doména o tom neví —
mapování `dbsqlc.Bookmark` → `bookmark.Bookmark` je v adaptéru.

**Ověření:** po `sqlc generate` se `go build ./internal/postgres/...` zkompiluje.

### 3. Postgres adaptér

`internal/postgres.Store`:

- pool přes `pgxpool`
- `pgx.ErrNoRows` → `bookmark.ErrNotFound`
- unique violation (`23505`) → `ErrDuplicateURL` / `ErrDuplicateID`
- migrace při startu (`ApplyMigrations`)

Integrační test `store_integration_test.go` se **skipne**, když chybí `DATABASE_URL`.
S Dockerem musí projít CRUD + search + duplicitní URL.

**Ověření:**

```bash
export DATABASE_URL='postgres://bookmarks:bookmarks@127.0.0.1:5433/bookmarks?sslmode=disable'
go test -count=1 -race ./internal/postgres/...
```

### 4. Redis + HTTP wiring

- `internal/rediscache` — JSON pod klíčem `bookmark:{id}`, TTL z `BOOKMARKS_CACHE_TTL`
- `app.CachedStore` — Get nejdřív cache, Add plní, Delete invaliduje; Search jde do DB
- `/readyz` volá `Ready` na store (Postgres + Redis); `/healthz` ne
- `cmd/bookmarks` sestaví pg → cache → `CachedStore` → `httpapi.NewServer`

**Ověření:** `go test -count=1 -race ./...` a ruční `curl` proti běžící službě.
Druhý `GET /bookmarks/{id}` má jít z Redis (můžeš ověřit `redis-cli -p 6380 KEYS 'bookmark:*'`).

---

## ADR (co bys měl umět obhájit)

1. Proč zůstává `ServeMux` a ne Chi?
2. Proč sqlc místo ORM?
3. Proč cache jen na Get-by-id, ne na search?
4. Proč jsou Postgres/Redis mimo doménový balíček?

Nápověda: lekce 25, 33, 35 a ADR příklad z lekce 56.

## AI režim

`TECH LEAD` — SPEC a acceptance vlastníš ty. Agent smí generovat sqlc boilerplate
a wiring, ale mapování chyb, porty a readiness reviewuješ ty.
