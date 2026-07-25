# Lekce 39 — Checkpoint fáze 4

> **Čas:** ~90 min · **Fáze:** 4 — Architektura v Go · **AI režim:** `BOILERPLATE OK`

Tahle lekce nepřináší nic nového. Prochází fázi 4 (lekce 32–38) formou otázek
a odpovědí, jedním kumulativním cvičením a bodovanou rubrikou, ze které se pozná,
kterou lekci si zopakovat.

## Co budeš umět

- Vysvětlit každé architektonické rozhodnutí fáze 4 vlastními slovy, ne citací.
- Postavit malou službu, která kombinuje value objekt, port, adaptér, validaci na hranici a metriku.
- Poznat na cizím Go kódu čtyři typické zápachy „Symfony v Go".
- Změřit si vlastní stav a vybrat, co si zopakovat.
- Provést architektonické review vlastního projektu P03 podle checklistu.

## PHP → Go most

Shrnutí celé fáze do jedné tabulky. Levý sloupec je návyk, pravý to, co ho v Go nahradí.

| Symfony návyk | Go protějšek | Lekce |
|---|---|---|
| `src/Entity`, `src/Service`, `src/Repository` | balíčky podle domény, `internal/` na hranici modulu | 32 |
| `OrderRepositoryInterface` u implementace | malý interface u konzumenta | 33 |
| `float $price` nebo `int $priceInCents` volně | typ `Money` s konstruktorem a metodami | 34 |
| Doctrine entita = doména | agregát bez závislostí + port persistence | 35 |
| `#[Assert\Email]` na DTO | `ParseEmail(s) (Email, error)` na hranici | 36 |
| `security.yaml` + voters | middleware pro autentizaci, doména pro autorizaci | 37 |
| `bin/console` a autowiring | `cmd/x/main.go` s ručním wiringem | 38 |

## Recap

### 32 — Project layout a `internal/`

**Otázka:** Proč se v Go balíček nejmenuje `service` ani `utils`?
**Odpověď:** Jméno balíčku je předpona každého jeho identifikátoru. `order.Service` nese
informaci, `service.Service` je šum a `utils.Helper` nenese nic. Balíček je jednotka
viditelnosti a soudržnosti, ne složka podle role.

**Otázka:** Co přesně dělá `internal/`?
**Odpověď:** Balíček pod `internal/` smí importovat jen kód, který má společný rodičovský
adresář s tímhle `internal/`. Vynucuje to kompilátor, ne dohoda. Je to jediná skutečná
hranice modulu, kterou Go má.

| Chci | Kam to dát |
|---|---|
| Kód, který nemá být veřejné API | `internal/` |
| Spustitelný program | `cmd/<jméno>/main.go` |
| Doménu | `internal/<doména>/`, jméno podle domény |
| Sdílenou drobnost mezi dvěma balíčky | většinou nikam — duplikuj, dokud to nebolí |

### 33 — Porty a adaptéry

**Otázka:** Kde má být definovaný interface?
**Odpověď:** U toho, kdo ho volá. Konzument ví, co potřebuje; implementace se přizpůsobí.
Díky implicitní implementaci nemusí adaptér o interface vůbec vědět.

**Otázka:** Jak si ověřím, že adaptér port splňuje, aniž bych spustil program?
**Odpověď:** `var _ order.Repository = (*store.Memory)(nil)` — když přestane platit,
nepřeloží se to.

**Otázka:** Kolik metod má mít port?
**Odpověď:** Tolik, kolik jich konzument skutečně volá. „Kompletní CRUD pro budoucnost"
je závazek, který zaplatí každý budoucí adaptér.

### 34 — Doménové typy a value objekty

**Otázka:** Proč `Money int64` a ne `float64`?
**Odpověď:** `0.1 + 0.2 != 0.3`. Peníze se počítají v nejmenší jednotce jako celá čísla
a formátují se až na výstupu.

**Otázka:** Co dělá z pojmenovaného typu value objekt?
**Odpověď:** Konstruktor, který jediný umí vyrobit platnou hodnotu, neměnnost
(metody vracejí novou hodnotu) a porovnatelnost přes `==` po normalizaci.

| Vlastnost | Proč na ní záleží |
|---|---|
| Konstruktor `ParseX` | nevalidní hodnota nemůže vzniknout |
| Normalizace uvnitř | `==` a klíč v mapě fungují |
| Hodnotový receiver | kopie je levná a bezpečná |
| Použitelná nulová hodnota nebo poznatelná | zero value nesmí být tichá past |

### 35 — Persistence

**Otázka:** Kam patří `context.Context` v repozitáři?
**Odpověď:** První parametr každé metody. Nikdy do struktury.

**Otázka:** Proč in-memory fake ukládá kopii?
**Odpověď:** Databáze sdílenou paměť nenabízí. Kdyby fake ukládal ukazatel, testy by
prošly a produkce ne — a tenhle rozdíl se hledá nejhůř.

**Otázka:** Co repozitář vrací, když záznam není?
**Odpověď:** Doménovou sentinel chybu (`ErrNotFound`), ne `sql.ErrNoRows` a ne `nil, nil`.
Překlad chyb technologie na doménu je práce adaptéru.

### 36 — Validace na hranici

**Otázka:** Rozdíl mezi validací a invariantem?
**Odpověď:** Validace vstupu se dělá jednou, na hranici, na neznámých datech. Invariant
platí vždy a vynucuje ho konstruktor — nejde porušit ani z vlastního kódu.

**Otázka:** Proč `Parse` místo `IsValid`?
**Odpověď:** `IsValid` vrátí bool a informaci zahodí. `Parse` vrátí typ, který je důkazem.

**Otázka:** 400, nebo 422?
**Odpověď:** 400 = tělu nerozumím (syntaxe). 422 = rozumím a nedává smysl (sémantika).

| Chyba | Status |
|---|---|
| Nečitelné tělo, neznámé pole | 400 |
| Selhala validace polí | 422 |
| Konflikt se stavem, duplicita | 409 |
| Zdroj neexistuje | 404 |
| Cokoli neočekávaného | 500 + detail jen do logu |

### 37 — Autentizace a observabilita

**Otázka:** Proč `subtle.ConstantTimeCompare`?
**Odpověď:** `==` skončí na prvním rozdílném bajtu, takže doba odpovědi prozrazuje,
kolik znaků sedí.

**Otázka:** Které autorizační rozhodnutí middleware udělat nemůže?
**Odpověď:** Každé, které závisí na datech („je to tvoje objednávka?"). Middleware zvládne
jen scope; zbytek je doména.

**Otázka:** Proč `user_id` nesmí být label metriky?
**Odpověď:** Každá kombinace hodnot labelů je samostatná časová řada. Milion uživatelů
je milion řad a Prometheus po týdnu sežere paměť.

| Signál | Na co odpovídá |
|---|---|
| Counter `_total` | kolikrát se něco stalo |
| Gauge | jaká je hodnota teď |
| Histogram `_seconds` | jak je rozdělené trvání (p95, p99) |
| Request ID | který záznam patří k témuž požadavku |

### 38 — Hexagonální služba

**Otázka:** V jakém pořadí se staví?
**Odpověď:** Doména → porty → fake adaptér → testy případů užití → HTTP adaptér → wiring.
Prvních pět kroků nepotřebuje nic zvenku, takže zpětná vazba je v milisekundách.

**Otázka:** Jak poznám, že hranice drží?
**Odpověď:** Testem. `go/parser` nad soubory balíčku ověří, že doména neimportuje
`net/http` ani `encoding/json`.

**Otázka:** Jak poznám přepálenou abstrakci?
**Odpověď:** Interface s jedinou implementací a bez fake, vrstva, která jen přeposílá
volání, a cokoli, co nevysvětlíš juniorovi za třicet vteřin.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Balíčky podle vrstev | Symfony adresářová struktura | balíček podle domény |
| Interface u implementace | zvyk na `*Interface` | interface u konzumenta |
| `float64` na peníze | vypadá přirozeně | celé centy ve vlastním typu |
| Validace uvnitř domény | „ať je to na jednom místě" | hranice validuje, konstruktor vynucuje |
| Interní chyba v odpovědi | nejkratší cesta k debugu | detail do logu, ven neutrální titulek |
| `user_id` jako label metriky | „hodilo by se to" | whitelist labelů s nízkou kardinalitou |

## Úkol

Kumulativní cvičení: malá rezervační služba `booking`, která kombinuje value objekt
(34), port s in-memory adaptérem (33, 35), validaci na hranici a mapování chyb (36)
a jednoduchou metriku (37). Pracuj v `exercise/`.

V reálném projektu by to byly tři balíčky (`booking`, `store`, `httpapi`); tady jsou
kvůli rozsahu v jednom, ale hranice mezi nimi je v kódu vidět.

### A — hodnotové typy (~20 min)

1. `ParseRoomID(s string) (RoomID, error)` — ořízne bílé znaky, převede na velká
   písmena a ověří tvar „písmeno, pomlčka, tři číslice" (`A-101`). Prázdný vstup →
   `ErrEmptyRoomID`, cokoli jiného mimo tvar (včetně `A-000`, které neexistuje) →
   chyba obalující `ErrInvalidRoomID`. Při chybě vrať nulovou hodnotu.
2. `NewDateRange(from, to time.Time) (DateRange, error)` — zarovná oba časy na celý den
   v UTC. Konec musí být po začátku (jinak `ErrInvalidRange`), pobyt smí mít nejvýš
   `MaxNights` nocí (jinak `ErrRangeTooLong`).
3. `DateRange.Nights()` a `DateRange.Overlaps(other DateRange) bool`. Interval je
   **polootevřený**: odjezd 20. a příjezd dalšího hosta 20. se nepřekrývají.

např. `ParseRoomID("a-101")` → `"A-101"`

### B — port, adaptér, služba a metrika (~40 min)

1. `NewMemoryRepo`, `Save`, `ByRoom` — in-memory adaptér portu `Repository`. Duplicitní
   `Ref` → `ErrDuplicateRef`, zrušený kontext → chyba z `ctx.Err()`, `ByRoom` vrací
   rezervace seřazené podle začátku pobytu. Bezpečné pro souběžné použití.
2. `Metrics.Inc` a `Metrics.Snapshot` — čítače pod mutexem. **Nulová hodnota musí být
   použitelná** (`var m Metrics` a rovnou `Inc`), `Snapshot` vrací kopii.
3. `Service.Book(ctx, req)` — zvaliduje požadavek, převede vstup na hodnotové typy,
   načte rezervace pokoje, odmítne překryv (`ErrRoomTaken`), spočítá cenu
   (`nightly_rate × počet nocí`) a uloží. Metriky: `MetricCreated` po úspěchu,
   `MetricRejected` při validační chybě, překryvu i duplicitní referenci.

např. `Book(BK-1, A-101, 3 noci × 1500)` → `Total:4500`

### C — hranice systému (~30 min)

1. `ValidationErrors.Error()` a `Get(field)`.
2. `CreateBookingRequest.Validate()` — sbírá **všechny** chyby v pořadí `ref`, `room`,
   `guest`, `from`, `to`, `nightly_rate`. Kódy: `CodeRequired` pro chybějící hodnotu,
   `CodeFormat` pro špatný tvar (pokoj, datum, jméno mimo 2–40 znaků), `CodeRange` pro
   nesmyslný termín a nekladnou cenu. Datum má tvar `DateLayout`. Chyba termínu patří
   k poli `to`.
3. `Handler(svc *Service) http.Handler` — `POST /bookings` (přísné dekódování, 201
   s hlavičkou `Location` a JSON tělem včetně `nights` a `total`) a `GET /metrics`
   vracející snapshot. Mapování: 400 nečitelné tělo, 422 validace (s polem `errors`),
   409 `ErrRoomTaken` i `ErrDuplicateRef`, 500 cokoli jiného. Chyby jako
   `application/problem+json`.

např. `POST /bookings` → `201` + `Location:/bookings/BK-1`, `total:4500`

```bash
make lesson L=39
cd lessons/lesson-39/exercise && go test -race -count=1 .
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Review projektu P03

Druhá polovina checkpointu není psaní kódu, ale čtení vlastního. Otevři
[`projects/p03-hex-service/`](../../projects/p03-hex-service/ACCEPTANCE.md) a projdi ho
jako cizí pull request. U každého bodu si odpověz **ano / ne / nevím** — a „nevím" je
horší odpověď než „ne", protože znamená, že tomu kódu nevěříš.

### Šipky závislostí

- [ ] `internal/order` neimportuje `net/http`, `encoding/json`, `database/sql` ani `os`.
      Ověř to testem přes `go/parser`, ne pohledem.
- [ ] `internal/app` neimportuje `internal/httpapi` ani `internal/memstore`.
- [ ] Porty `Repository`, `Clock` a `IDGen` jsou v `app`, u konzumenta.
- [ ] Existuje `var _ app.Repository = (*memstore.Repository)(nil)`, takže splnění portu
      hlídá kompilátor.

### Doména

- [ ] Každý invariant vynucuje konstruktor, ne komentář ani volající.
- [ ] `Money` nejde vyrobit mimo `NewMoney`, sčítání různých měn je chyba a přetečení
      `int64` je ošetřené.
- [ ] Přechody stavů vracejí novou hodnotu; nepovolený přechod nemění nic.
- [ ] `order.New` si slice položek kopíruje.
- [ ] V doméně není `time.Now()` ani generování ID.

### Hranice systému

- [ ] HTTP adaptér má vlastní DTO; doménové typy nemají JSON tagy.
- [ ] Mapování chyb na statusy je na **jednom** místě, ne v každém handleru.
- [ ] Text interní chyby se do těla odpovědi nedostane. Zkus si to: dej fake repozitáři
      chybu s připojovacím řetězcem a podívej se na tělo pětistovky.
- [ ] Tělo požadavku je omezené velikostí a neznámá pole se odmítají.

### Testy a spuštění

- [ ] Doménové testy neznají `context`, `httptest` ani fake.
- [ ] Testy use-casů běží proti fake portu a ověřují i to, co se **nestalo**
      (že se `Save` nezavolal).
- [ ] Žádný test nezávisí na pevném portu, reálném čase ani náhodném ID.
- [ ] `main.go` obsahuje jen konfiguraci, wiring a životní cyklus.
- [ ] `SIGINT` i `SIGTERM` vedou na graceful shutdown s vlastním timeoutem.
- [ ] `go test -race ./...` a `go vet ./...` jsou čisté.

Najdi ve svém P03 **aspoň jednu věc, kterou bys smazal** — interface s jedinou
implementací, vrstvu, která jen přeposílá volání, konfigurační volbu, kterou nikdo
nepoužije. Smazat ji je součást cvičení. Architektura se nepoznává podle toho, co
v ní je, ale podle toho, co se z ní dá odebrat, aniž by se něco rozbilo.

## Sebehodnocení

Za každou položku si dej 1 bod, jen když je odpověď „ano, umím to bez hledání".

| # | Položka | Lekce |
|---|---|---|
| 1 | Vysvětlím, co vynucuje `internal/`, kdo to kontroluje, a pojmenuju balíček podle domény | 32 |
| 2 | Umístím interface ke konzumentovi a obhájím to | 33 |
| 3 | Ověřím splnění portu jedním řádkem, který kontroluje kompilátor | 33 |
| 4 | Napíšu value objekt s konstruktorem, normalizací a `==` sémantikou | 34 |
| 5 | Vysvětlím, proč peníze nejsou `float64`, na konkrétním příkladu | 34 |
| 6 | Napíšu in-memory adaptér, který ukládá kopie, a řeknu proč | 35 |
| 7 | Přeložím chybu úložiště na doménovou sentinel chybu | 35 |
| 8 | Rozliším validaci vstupu od invariantu a umístím obojí | 36 |
| 9 | Vrátím všechny validační chyby najednou bez pasti typovaného nil | 36 |
| 10 | Namapuju chyby na 400/404/409/422/500 a zdůvodním každý | 36 |
| 11 | Naparsuju Bearer token včetně chybných tvarů a porovnám tajemství v konstantním čase | 37 |
| 12 | Předám identitu kontextem s typově bezpečným klíčem | 37 |
| 13 | Pojmenuju metriku podle konvence a vyhnu se vysoké kardinalitě | 37 |
| 14 | Postavím službu v pořadí doména → port → adaptér → HTTP → wiring a hranici ohlídám testem | 38 |
| 15 | Kumulativní cvičení mám zelené včetně `-race` a v P03 jsem našel abstrakci ke smazání | 39 |

### Co s výsledkem

| Skóre | Doporučení |
|---|---|
| 14–15 | Fáze 4 je hotová. Dodělej projekt P03 podle `ACCEPTANCE.md` a jdi na fázi 5. |
| 11–13 | Dobré. Zopakuj lekce, ze kterých ti chybí body, a projekt P03 dopiš celý. |
| 8–10 | Vrať se k lekcím 33 a 36 — hranice a validace jsou jádro fáze. Cvičení 38 napiš znovu od nuly. |
| 4–7 | Projdi znovu lekce 32–35 včetně cvičení. Bez portů a value objektů nemá smysl jít dál. |
| 0–3 | Zopakuj celou fázi 4. Věnuj tomu radši týden navíc než měsíc ladění špatné architektury. |

Body 11–13 jsou nulové? Samo o sobě to fázi neblokuje, ale lekci 37 si projdi před tím,
než postavíš cokoli, co půjde na internet.

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `39`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=39` a `go test -race` prochází
- [ ] Vyplnil jsi rubriku poctivě, bez nahlížení do řešení
- [ ] Projekt P03 splňuje všechna kritéria v `ACCEPTANCE.md`
- [ ] Umíš z hlavy vyjmenovat pořadí prací při stavbě služby
- [ ] Umíš vysvětlit, kdy interface v Go nepsat vůbec

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).
Na checkpointu je poctivější AI nepoužít vůbec: měříš svůj stav, ne stav modelu.

## Další čtení

1. [Go blog — Organizing a Go module](https://go.dev/doc/modules/layout)
2. [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
3. [Effective Go — Interfaces and methods](https://go.dev/doc/effective_go#interfaces_and_types)
4. [Prometheus — Metric and label naming](https://prometheus.io/docs/practices/naming/)
