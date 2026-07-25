# P02 — REST API pro správu úkolů

Projekt k lekci 31 (checkpoint fáze 3). Akceptační kritéria a kontrakt API jsou
v [ACCEPTANCE.md](ACCEPTANCE.md).

## Struktura

```
cmd/api/main.go          konfigurace z prostředí, sestavení aplikace, graceful shutdown
internal/task/           doména: typ Task, validace, in-memory Store s mutexem
internal/httpapi/        router (ServeMux vzory), handlery, chybové odpovědi, middleware
```

Hranice je záměrná: `internal/task` neimportuje `net/http` ani `encoding/json`,
`internal/httpapi` překládá doménové chyby na status kódy a doménový typ na DTO.
`internal/` znamená, že si tyhle balíčky nikdo mimo tento projekt nenaimportuje.

## Spuštění

```bash
go run ./cmd/api
```

| Proměnná | Výchozí | Význam |
|----------|---------|--------|
| `ADDR` | `0.0.0.0:8080` | adresa pro poslech |
| `PORT` | — | zkratka pro `0.0.0.0:PORT`, přebíjí `ADDR` |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `READ_TIMEOUT` | `5s` | limit pro čtení hlaviček požadavku |
| `SHUTDOWN_TIMEOUT` | `10s` | grace perioda pro dokončení požadavků |

## Testy

```bash
go test -race ./...
```
