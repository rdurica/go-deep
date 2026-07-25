# P03 — Hexagonální služba objednávek

Projekt fáze 4 (lekce 32–39). Cílem není „REST API", to už umíš z P02. Cílem je
**služba, jejíž závislosti míří dovnitř**: doména nezná okolí, adaptéry jsou vyměnitelné
a všechno se spojuje ručně v `main`.

## Zadání

Služba spravuje objednávky. Objednávka má zákazníka, jednu nebo víc položek a stav, který
se mění jen po povolených přechodech.

```
                 ┌──────── cancel ────────┐
                 ↓                        │
  new ──pay──> paid ──ship──> shipped     │
   │             │                        │
   └─── cancel ──┴────────────────────────┘   (shipped a cancelled jsou koncové)
```

## Struktura

```
projects/p03-hex-service/
  ACCEPTANCE.md
  cmd/orders/main.go        # wiring, konfigurace, graceful shutdown
  internal/order/           # DOMÉNA: Money, Line, Order, stavy, invarianty, chyby
  internal/app/             # use-casy + porty Repository, Clock, IDGen
  internal/memstore/        # adaptér portu Repository (in-memory)
  internal/httpapi/         # vstupní adaptér: router, DTO, problem+json
```

Závislosti míří jedním směrem: `httpapi → app → order` a `memstore → order`.
Balíček `order` neimportuje nic z `app`, `httpapi` ani ze standardní knihovny pro I/O.

## Akceptační kritéria

### Doména (`internal/order`)

- [ ] `Money` má neexportovaná pole, drží celé centy a kód měny, a nejde vyrobit
      jinak než přes `NewMoney`. Záporná částka a neplatný kód měny jsou chyby.
- [ ] `Money.Add` odmítne sčítat různé měny (`ErrCurrencyMismatch`) a hlídá přetečení
      `int64`. Nulová hodnota se chová jako neutrální prvek.
- [ ] `Money.Mul` odmítne záporný násobek a hlídá přetečení.
- [ ] `Status` je `iota` typ se `String()`; nulová hodnota je `unknown`.
- [ ] `NewLine` odmítne prázdné SKU, nekladné množství, množství nad `MaxQuantity`
      a nekladnou cenu.
- [ ] `order.New` odmítne prázdné ID, prázdného zákazníka, nulový čas, objednávku
      bez položek, položku porušující invariant a míchání měn.
- [ ] `order.New` kopíruje slice položek — změna vstupu zvenčí objednávku neovlivní.
- [ ] `Pay`, `Ship`, `Cancel` vracejí **novou** hodnotu a nepovolený přechod hlásí
      chybou obalující `ErrInvalidTransition`. Odeslanou objednávku nelze zrušit.

### Aplikační vrstva (`internal/app`)

- [ ] Porty `Repository`, `Clock` a `IDGen` jsou definované tady, u konzumenta.
- [ ] `Service` má use-casy `Place`, `Get`, `List`, `Pay`, `Ship`, `Cancel`.
- [ ] Žádný use-case nerozhoduje o doménovém pravidle — jen načte, nechá rozhodnout
      doménu a uloží.
- [ ] Když doména vstup zamítne, do úložiště se **nic** nezapíše.
- [ ] Chyby portů se obalují přes `%w` s kontextem; doménové chyby se vracejí beze změny.

### Adaptér úložiště (`internal/memstore`)

- [ ] Splňuje port `app.Repository` (ověřeno `var _ app.Repository = …`).
- [ ] Je bezpečný pro souběžné použití a projde `-race`.
- [ ] Kopíruje položky při zápisu i čtení, aby se chovala jako skutečná databáze.
- [ ] `List` vrací objednávky seřazené podle ID (jinak by testy byly flaky).
- [ ] Respektuje zrušený `context.Context`.
- [ ] Chybějící objednávka → chyba obalující `order.ErrNotFound`.

### HTTP adaptér (`internal/httpapi`)

| Metoda a trasa | Význam | Úspěch |
|----------------|--------|--------|
| `GET /healthz` | liveness, bez závislostí | 200 |
| `GET /readyz` | readiness, kontroluje úložiště | 200 / 503 |
| `POST /orders` | založí objednávku | 201 |
| `GET /orders` | výpis seřazený podle ID | 200 |
| `GET /orders/{id}` | detail | 200 |
| `POST /orders/{id}/pay` | zaplacení | 200 |
| `POST /orders/{id}/ship` | odeslání | 200 |
| `POST /orders/{id}/cancel` | zrušení | 200 |

- [ ] Trasy používají vzory `http.ServeMux` (metoda + `{id}`).
- [ ] Adaptér má **vlastní DTO** s JSON tagy; doménové typy tagy nemají.
- [ ] Chybová odpověď je problem+json s `Content-Type: application/problem+json`
      a poli `type`, `title`, `status`, `detail`.
- [ ] Mapování chyb je na jednom místě: 400 nepoužitelné JSON, 404 `ErrNotFound`,
      409 `ErrInvalidTransition`, 422 ostatní doménové chyby, 500 zbytek.
- [ ] Text interní chyby se **nikdy** neobjeví v těle odpovědi.
- [ ] Tělo požadavku je omezené velikostí a odmítá neznámá pole.

### Spuštění (`cmd/orders`)

- [ ] `main.go` nedělá nic než konfiguraci, wiring a životní cyklus.
- [ ] Konfigurace se čte z prostředí: `ORDERS_ADDR`, `ORDERS_READ_TIMEOUT`,
      `ORDERS_WRITE_TIMEOUT`, `ORDERS_SHUTDOWN_TIMEOUT`. Neplatná hodnota = chyba při startu.
- [ ] `http.Server` má nastavené read, write a idle timeouty.
- [ ] `SIGINT` i `SIGTERM` spustí graceful shutdown s vlastním timeoutem.

### Testy

- [ ] Unit testy domény bez fakes a bez kontextu.
- [ ] Testy use-casů proti ručně psanému fake portu, který počítá volání.
- [ ] HTTP integrační testy přes `httptest` proti skutečné doméně a in-memory adaptéru.
- [ ] Test hranice balíčků přes `go/parser` hlídá, že `order` ani `app` neimportují
      `net/http`, `encoding/json`, `database/sql` ani `os`.
- [ ] Žádný test nezávisí na síti, pevném portu, reálném čase ani na náhodných ID.
- [ ] `go test -race ./...` a `go vet ./...` jsou čisté.

## Jak ověřit

```bash
cd projects/p03-hex-service
go vet ./...
go test -race ./...
go run ./cmd/orders     # ORDERS_ADDR=:9090 go run ./cmd/orders
```

Ruční zkouška:

```bash
curl -s -X POST localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"customer":"radek@example.com","currency":"CZK",
       "lines":[{"sku":"kniha-go","quantity":2,"unit_price_cents":49900}]}'

curl -s localhost:8080/orders
curl -s -X POST localhost:8080/orders/<id>/pay
curl -i -X POST localhost:8080/orders/<id>/cancel   # po ship vrátí 409 problem+json
```

## Na co si dát pozor

1. **Doména nesmí importovat okolí.** Jakmile v `internal/order` uvidíš `encoding/json`
   nebo `net/http`, něco se rozlilo.
2. **Port patří ke konzumentovi.** `Repository` je v `app`, ne v `memstore`.
3. **Fake musí kopírovat data.** Jinak testy ukážou úspěch tam, kde produkce selže.
4. **`time.Now()` a generování ID jsou závislosti.** Bez portů `Clock` a `IDGen` je
   každý assert hádanka.
5. **Interní chyba do logu, ne do odpovědi.** `err.Error()` v těle 500 je bezpečnostní
   incident, ne pomoc klientovi.

Checklist pro architektonické review vlastního řešení je v
[lekci 39](../../lessons/lesson-39/README.md#review-projektu-p03).
