# Lekce 37 — Autentizace a observabilita

> **Čas:** ~90 min · **Fáze:** 4 — Architektura v Go · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Rozdělit autentizaci a autorizaci a umístit každou tam, kam patří — jednu do middlewaru, druhou do domény.
- Naparsovat hlavičku `Authorization` přesně a porovnat tajemství bez časového úniku.
- Předat identitu requestem přes kontext s typově bezpečným klíčem.
- Pojmenovat metriku tak, aby přežila produkci, a vysvětlit, proč `user_id` jako label zabije Prometheus.
- Vybrat mezi čítačem, gauge a histogramem a popsat službu metodou RED.
- Napsat registr metrik bezpečný pro souběžné použití a instrumentovat HTTP vrstvu bez exploze kardinality.

## Teorie

### Autentizace, autorizace a kde která bydlí

**Autentizace** je „kdo jsi". Odpověď je identita, nebo 401.
**Autorizace** je „smíš tohle". Odpověď je ano/ne, nebo 403.

Autentizace patří do middlewaru: je to průřezová záležitost, stejná pro všechny
endpointy, a nepotřebuje znát doménu. Autorizace do middlewaru **nepatří**, jakmile
závisí na datech:

```go
// Tohle middleware rozhodnout nemůže — musel by načíst objednávku.
func (s *Service) Cancel(ctx context.Context, u User, id OrderID) error {
	o, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if o.CustomerID != u.ID && !u.HasScope("orders:admin") {
		return ErrForbidden
	}
	return s.cancel(ctx, o)
}
```

Middleware umí odmítnout „nemáš scope `orders:write`". Nikdy neumí odmítnout „tohle
není tvoje objednávka", protože k tomu potřebuje repozitář, a v ten moment už to není
middleware, ale doména schovaná na špatném místě.

### Parsování Bearer tokenu

Hlavička vypadá jednoduše, ale má překvapivě mnoho špatných tvarů: chybí, je prázdná,
obsahuje jen `Bearer`, má jiné schéma, má schéma malými písmeny, má tokenů dva, je
obalená mezerami. Celé to jde rozseknout jedním voláním:

```go
func ParseBearer(header string) (string, error) {
	fields := strings.Fields(header)   // sjednotí mezery i tabulátory
	switch {
	case len(fields) == 0:
		return "", ErrMissingAuthorization   // "" i "   "
	case !strings.EqualFold(fields[0], "bearer"):
		return "", ErrUnsupportedScheme      // "Basic …", "Token …"
	case len(fields) == 1:
		return "", ErrMissingToken           // "Bearer"
	case len(fields) > 2:
		return "", ErrMalformedAuthorization // "Bearer a b"
	}
	return fields[1], nil
}
```

`strings.Fields` sloučí libovolný počet bílých znaků, takže `Bearer  token` i
`\tbearer\ttoken\n` projdou jako `token`. To je vědomé rozhodnutí být k okrajovým
mezerám tolerantní — proxy a klienti je do hlavičky přidávají a odmítnout kvůli tomu
požadavek by nikomu nepomohlo. Tolerance má ale hranici: `Bearer a b` jsou dva tokeny
a takový požadavek je nejednoznačný, tedy chyba.

Schéma je podle RFC 6750 case-insensitive, proto `strings.EqualFold`, ne `==`. A všechny
chyby vracejí **prázdný token**, aby se volající nemohl splést a použít půlku vstupu.

### Konstantní porovnání

Tohle je řádek, který v code review musíš vidět:

```go
// ŠPATNĚ — doba běhu závisí na tom, kolik znaků sedí
if token == expected {

// SPRÁVNĚ
if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1 {
```

Porovnání řetězců přes `==` skončí na prvním rozdílném bajtu. Útočník, který umí měřit
dobu odpovědi, může token uhodnout znak po znaku — místo 64^32 pokusů jich potřebuje
64×32. V praxi to přes internet chce hodně vzorků, ale v lokální síti nebo mezi
kontejnery to je reálné a obrana stojí jeden řádek.

`ConstantTimeCompare` vrací 0, když se liší délky, takže délka tajemství se stále dá
změřit. Proto se porovnávají **otisky pevné délky**, ne tokeny.

Ještě jedna past je ve tvaru smyčky. Tohle vypadá nevinně:

```go
for token, name := range tokens {
	if subtle.ConstantTimeCompare([]byte(token), []byte(presented)) == 1 {
		return name    // ŠPATNĚ — dřívější návrat je taky časový únik
	}
}
```

Smyčka končí, jakmile najde shodu, takže doba odpovědi prozrazuje pozici tokenu. Řešení
je nudné: projdi vždycky všechny záznamy a výsledek si odlož do proměnné.

### Hashování hesel

Token je dlouhý náhodný řetězec, který nejde uhodnout slovníkem, takže se v úložišti
drží jako SHA-256 otisk. **Heslo je něco jiného.**

```go
// POZOR: SHA-256 je na hesla ŠPATNĚ. Je záměrně rychlá, takže se dá hrubou
// silou zkoušet miliardkrát za sekundu. V produkci patří bcrypt nebo argon2.
func HashPassword(password, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + password))
	return hex.EncodeToString(sum[:])
}

func VerifyPassword(hash, password, salt string) bool {
	return subtle.ConstantTimeCompare([]byte(HashPassword(password, salt)), []byte(hash)) == 1
}
```

`golang.org/x/crypto/bcrypt` **není** ve standardní knihovně, a protože kurz drží
nulové závislosti, cvičení používá SHA-256 se solí jen jako **demonstraci mechaniky**:
sůl, otisk, konstantní porovnání. Na skutečná hesla patří `bcrypt` nebo
`golang.org/x/crypto/argon2` — funkce, které jsou pomalé **schválně** a mají nastavitelný
faktor práce. Rychlá hašovací funkce je u hesel chyba, ne optimalizace.

Vyhledání podle otisku má bonus: mapa nebo `WHERE token_hash = $1` porovnává hodnoty
pevné délky, takže se řeší i timing.

### Identita v kontextu

```go
type contextKey int

const (
	userKey contextKey = iota
	routeKey
)

func UserFrom(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(userKey).(string)
	return user, ok
}
```

Klíč je hodnota neexportovaného typu. To je celý trik: cizí balíček nemůže vyrobit
stejný klíč, takže ti hodnotu nepřepíše ani omylem. `ctx.Value("user")` s řetězcovým
klíčem je kolize, která čeká na svou příležitost, a `go vet` na ni upozorní.

Kontext je na **request-scoped data**, ne na závislosti. Databázové spojení, logger
nebo konfigurace patří do struktury handleru, ne do `ctx`.

### Co nikdy nelogovat

| Nikdy | Proč |
|---|---|
| Celý `Authorization` header | token v logu = token v Elasticu, Slacku, ticketu |
| Heslo, i „jen při chybě" | tam se loguje nejčastěji |
| Číslo karty, rodné číslo | GDPR a PCI |
| Celé tělo požadavku „pro debug" | obsahuje všechno výše |

Logovat naopak chceš `user_id`, request ID, cestu, status a trvání. Když opravdu
potřebuješ token v logu, loguj jeho otisk nebo prvních osm znaků.

### Tři pilíře a jak metriky pojmenovat

**Logy** popisují událost, **metriky** agregují čísla, **trasování** spojuje jeden
požadavek napříč službami. Metriky jsou nejlevnější a nejrychleji odpoví na „hoří to?".

Konvence pojmenování z Prometheu, která platí i mimo něj:

- snake_case, prefix podle subsystému: `http_requests_total`
- **jednotka v názvu**: `_seconds`, `_bytes`, `_ratio`. Nikdy `_ms` — základní jednotky.
- čítače končí na `_total`

Typy:

| Typ | Kdy | Příklad |
|---|---|---|
| Counter | jen roste, restart na nulu | `http_requests_total` |
| Gauge | jde nahoru i dolů | `queue_depth`, `goroutines` |
| Histogram | rozdělení hodnot, umí percentily | `http_request_duration_seconds` |

Průměr trvání je k ničemu — schová přesně ten chvost, kvůli kterému lidé volají podpoře.
Chceš p95 a p99, a ty umí jen histogram.

### Kardinalita, aneb proč `user_id` zabije Prometheus

Každá kombinace hodnot labelů je **samostatná časová řada** v paměti. Metrika se třemi
labely, kde má každý deset hodnot, je tisíc řad. To je v pohodě.

```go
// ŠPATNĚ — jedna řada na uživatele a jedna na každou URL
m.Inc("http_requests_total", map[string]string{
	"user_id": u.ID,        // milion uživatelů = milion řad
	"path":    r.URL.Path,  // /orders/1a2b… = neomezeně řad
})

// SPRÁVNĚ — konečná, malá množina hodnot
m.Inc("http_requests_total", map[string]string{
	"method": r.Method,             // ~7 hodnot
	"route":  "/orders/{id}",       // VZOR cesty, ne konkrétní URL
	"status": strconv.Itoa(status), // několik desítek kódů
})
```

Vysoká kardinalita nezpůsobí chybu — způsobí, že Prometheus po týdnu sežere paměť
a spadne. A protože middleware `r.URL.Path` vidí, ale vzor trasy ne, musí mu ho někdo
podat. Nejjednodušší cesta je kontext:

```go
mux.Handle("GET /items/{id}", WithRoute("/items/{id}", Instrument(m)(itemHandler)))
```

Middleware pak čte `RouteFrom(r.Context())` a když vzor nikdo nenastavil, použije
náhradní `"unknown"`. To je lepší než spadnout — a v grafu je jedna řada `unknown`
okamžitě vidět jako chyba wiringu.

**RED metoda** říká, co u každé služby měřit: **R**ate (požadavky za sekundu),
**E**rrors (podíl chyb), **D**uration (rozdělení trvání). Všechny tři vyplynou
z jednoho čítače s labelem `status` a jednoho histogramu.

Ještě jedna drobnost, o kterou zakopne každý: `http.ResponseWriter` status kód po zápisu
nijak nezpřístupňuje. Middleware si ho proto musí zapamatovat obalením:

```go
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}
```

Vnořený interface znamená, že `statusRecorder` splňuje `http.ResponseWriter` a přepisuje
z něj jedinou metodu. Handler, který `WriteHeader` nezavolá vůbec, poslal implicitně 200 —
na to je potřeba pamatovat, jinak dostaneš do metriky `status="0"`.

Poslední dva kousky: **request ID** vygenerované na okraji a propsané do logu i do
odchozích volání spojí záznamy do jednoho příběhu. A `/healthz` (žiju, nerestartuj mě)
je něco jiného než `/readyz` (mám spojení do databáze, posílej provoz).

## Rozdíly proti PHP

V Symfony je autentizace konfigurace. `security.yaml`, firewall, authenticator, voter —
framework to poskládá a ty se dozvíš `$this->getUser()`:

```yaml
security:
    firewalls:
        api:
            pattern: ^/api
            custom_authenticators: [App\Security\TokenAuthenticator]
    access_control:
        - { path: ^/api/orders, roles: ROLE_API }
```

V Go je autentizace funkce, kterou vidíš:

```go
mux := http.NewServeMux()
mux.Handle("POST /orders", createOrder(svc))

var handler http.Handler = mux
handler = Authenticate(tokens)(handler)   // pořadí je vidět
handler = Instrument(metrics)(handler)
```

Co se mění v uvažování: v Symfony hledáš, **kde je to nakonfigurované**. V Go čteš
`main.go` shora dolů a víš, co se s požadavkem stane. Cenou je, že si musíš pamatovat
pořadí a nic ti nezakřičí, když middleware zapomeneš. Odměnou je, že žádný `access_control`
řádek ti tiše nepovolí endpoint, o kterém nevíš.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `token == expected` | v PHP se to taky píše | `subtle.ConstantTimeCompare` |
| Token uložený v plaintextu | „vždyť je to jen API klíč" | ukládej SHA-256 otisk |
| Autorizace na data v middlewaru | reflex z `access_control` | rozhodnutí patří do domény |
| Klíč do kontextu jako `string` | rychlé a čitelné | neexportovaný typ klíče |
| `r.URL.Path` jako label | je po ruce | vzor trasy z routeru |
| `return` uvnitř porovnávací smyčky | vypadá jako optimalizace | projdi vždy všechny záznamy |
| Metrika `latency_ms` | zvyk na milisekundy | základní jednotky: `_seconds` |
| 401 bez `WWW-Authenticate` | klient si poradí | RFC 7235 ji vyžaduje |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 37`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `ParseBearer`, `HashPassword`, `VerifyPassword`, `Authenticate`

```bash
make lesson L=37 PART=1
```

Pak **`/go-deep-review 37 easy`**.

### Střední

Funkce: `UserFrom`, `NewMetrics`, `SeriesKey`, `Inc`, `Observe`

```bash
make lesson L=37 PART=2
```

Pak **`/go-deep-review 37 medium`**.

### Obtížný

Funkce: `Snapshot`, `Text`, `WithRoute`, `RouteFrom`, `Instrument`

```bash
make lesson L=37 PART=3
```

Pak **`/go-deep-review 37 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 37 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=37` (+ `make race L=37`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč `==` na tajemství je zranitelnost, a spočítat, o kolik útok zjednoduší
- [ ] Umíš vysvětlit, proč je SHA-256 na hesla špatná volba a co se používá místo ní
- [ ] Umíš vysvětlit, které autorizační rozhodnutí middleware udělat nemůže
- [ ] Umíš vysvětlit, proč musí middleware znát vzor trasy, ne `r.URL.Path`
- [ ] Umíš vysvětlit, co se stane s Prometheem, když dáš `user_id` jako label
- [ ] Umíš popsat rozdíl mezi `/healthz` a `/readyz`

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

## Další čtení

1. [pkg.go.dev — crypto/subtle](https://pkg.go.dev/crypto/subtle)
2. [RFC 6750 — The OAuth 2.0 Authorization Framework: Bearer Token Usage](https://datatracker.ietf.org/doc/html/rfc6750)
3. [Prometheus — Metric and label naming](https://prometheus.io/docs/practices/naming/)
4. [Go blog — Contexts and structs](https://go.dev/blog/context-and-structs)
