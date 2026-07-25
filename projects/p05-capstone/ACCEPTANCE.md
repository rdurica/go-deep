# P05 — Capstone: záložky URL

## Cíl

Dokončit produkčně vypadající mini-službu: čistá doména, souběžně bezpečný store,
HTTP API s problem-details a middleware, konfigurace z env a graceful shutdown.
Bez nových závislostí v `go.mod`.

## Akceptační kritéria

- [ ] `internal/bookmark` — model `Bookmark`, `NormalizeURL`, `NormalizeTags`, `New`,
      validace; **neimportuje** `net/http` ani `encoding/json` (test přes `go/parser`).
- [ ] `internal/store` — in-memory store s indexem podle tagu, `Add`/`Get`/`Delete`/`Search`,
      test souběhu s `-race`.
- [ ] `internal/httpapi` — `ServeMux` router (vzory metod/wildcardů), CRUD + search, `/healthz`, `/readyz`,
      problem-details, middleware (request ID, recovery, body limit, timeout).
- [ ] `internal/config` — načtení adresy a timeoutů z prostředí.
- [ ] `cmd/bookmarks` — wiring + `signal.NotifyContext` + `Shutdown`.
- [ ] `POST /bookmarks` — `201`; neplatná URL/JSON → `400`; duplicitní URL → `409`.
- [ ] `GET /bookmarks/{id}` — `200` nebo `404`.
- [ ] `DELETE /bookmarks/{id}` — `204` nebo `404`.
- [ ] `GET /bookmarks` — search (`q`, `tag`, `limit`, `cursor`).
- [ ] HTTP integrační testy přes `httptest` pokrývají happy path i 400/404/409.
- [ ] `SPEC.md` popisuje rozsah, API a non-goals.
- [ ] `go test -count=1 -race ./projects/p05-capstone/...` a `go vet` procházejí.

## Jak ověřit

```bash
gofmt -w projects/p05-capstone
go test -count=1 -race ./projects/p05-capstone/...
go vet ./projects/p05-capstone/...
```

Ruční smoke (volitelné):

```bash
go run ./projects/p05-capstone/cmd/bookmarks
curl -sS -X POST localhost:8080/bookmarks \
  -H 'content-type: application/json' \
  -d '{"url":"https://go.dev/blog/?utm_source=x","title":"Go Blog","tags":["go","blog"]}'
```
