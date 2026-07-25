# Lekce 24 — net/http od nuly: handler, ServeMux, httptest

> **Čas:** ~90 min · **Fáze:** 3 — net/http a tooling · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Napsat `http.Handler` ručně i přes `http.HandlerFunc` a vysvětlit, proč jsou to dvě podoby téhož.
- Popsat závazné pořadí `Header().Set` → `WriteHeader` → `Write` a poznat, co se tiše rozbije při jeho prohození.
- Přijmout a zvalidovat JSON tělo tak, aby ti klient nemohl poslat gigabajt dat.
- Rozhodnout mezi `http.ListenAndServe` a vlastním `http.Server` a odůvodnit to timeouty.
- Otestovat handler bez jediného otevřeného portu přes `httptest.NewRecorder` a přes `httptest.NewServer`.

## PHP → Go most

V Symfony je controller metoda, která dostane `Request` a **vrátí** `Response`. Objekt
odpovědi existuje jako hodnota, můžeš ho odložit, upravit, zahodit a vrátit jiný.

```php
final class HealthController
{
    #[Route('/healthz', methods: ['GET'])]
    public function health(Request $request): JsonResponse
    {
        $response = new JsonResponse(['status' => 'ok']);
        // ještě tady můžu všechno změnit — nic se neodeslalo
        $response->setStatusCode(200);
        return $response;
    }
}
```

V Go žádný `Response` objekt neexistuje. Handler dostane **zapisovač** a rovnou do něj píše:

```go
func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// od téhle chvíle už status ani hlavičky nezměníš — jsou na drátě
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

Návyk, který musíš opustit: *„chybu vyřeším později, odpověď ještě upravím."* Jakmile
proběhl první `Write`, hlavičky i status kód jsou odeslané. Chyba, která nastane potom,
už se do HTTP odpovědi nevejde — jediné, co s ní můžeš udělat, je zalogovat ji. To zásadně
mění, kde v handleru validuješ: **všechno kontroluj dřív, než začneš psát tělo.**

## Teorie

### `http.Handler` a `http.HandlerFunc`

Celý `net/http` stojí na jednom interface s jedinou metodou:

```go
type Handler interface {
	ServeHTTP(w ResponseWriter, r *Request)
}
```

Cokoli, co má tuhle metodu, je handler. Můžeš ho napsat jako typ se stavem:

```go
type greeter struct{ name string }

func (g greeter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ahoj %s", g.name)
}
```

Nebo, což je mnohem častější, jako funkci obalenou adaptérem `http.HandlerFunc`:

```go
var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ahoj")
})
```

`http.HandlerFunc` je pojmenovaný funkční typ, který sám na sobě má metodu `ServeHTTP`
volající sebe sama. Není to magie ani makro — je to sedm řádků v stdlib a je to nejčistší
příklad toho, jak se v Go funkce povyšuje na interface.

Konvence, kterou používáme v celém kurzu: **handler se vyrábí konstruktorem, který vrací
`http.Handler`.** Závislosti (store, logger, konfiguraci) předáš konstruktoru a uzavřeš je
v closure. Tím nahradíš to, co v Symfony dělá autowiring přes konstruktor controlleru.

```go
func HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { /* ... */ })
}
```

### Pořadí operací a proč na něm záleží

`http.ResponseWriter` má přesně tři metody a jejich pořadí je závazné:

1. `Header()` — mapa hlaviček, měnitelná jen do prvního zápisu,
2. `WriteHeader(status int)` — odešle status řádek a hlavičky, volá se **maximálně jednou**,
3. `Write([]byte)` — tělo; pokud jsi `WriteHeader` nezavolal, provede se implicitně s 200.

```go
// ŠPATNĚ — hlavička přijde pozdě, na drátě už status i hlavičky jsou
w.WriteHeader(http.StatusOK)
w.Header().Set("Content-Type", "application/json") // tiše ignorováno

// ŠPATNĚ — dvojí WriteHeader; druhý zápis Go zaloguje jako
// "superfluous response.WriteHeader call" a ignoruje
w.WriteHeader(http.StatusOK)
w.WriteHeader(http.StatusInternalServerError)

// SPRÁVNĚ
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
_ = json.NewEncoder(w).Encode(v)
```

Druhá past je klasika s `http.Error`:

```go
// ŠPATNĚ — http.Error si nastaví Content-Type: text/plain sám,
// takže tvoje JSON hlavička je k ničemu a klient dostane mix
w.Header().Set("Content-Type", "application/json")
http.Error(w, `{"error":"nope"}`, http.StatusBadRequest)
```

`http.Error` je fajn pro rychlý plain-text výstup. Jakmile má tvoje API JSON kontrakt,
napiš si vlastní `writeError` a používej ho **všude**, ať klient nikdy nedostane jednou
JSON a podruhé holý text.

### Čtení JSON těla bezpečně

Naivní verze je jednořádková a má tři díry:

```go
// ŠPATNĚ
var req EchoRequest
json.NewDecoder(r.Body).Decode(&req)
```

Neověřuje Content-Type, nemá limit velikosti (klient ti může poslat nekonečný stream)
a tiše spolkne překlepy v názvech polí. Použitelná verze:

```go
mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
if err != nil || mediaType != "application/json" {
	writeError(w, http.StatusUnsupportedMediaType, "expected application/json")
	return
}

r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()

var req EchoRequest
if err := dec.Decode(&req); err != nil {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}
	writeError(w, http.StatusBadRequest, "invalid json body")
	return
}
```

`mime.ParseMediaType` je tam schválně: hlavička často přijde jako
`application/json; charset=utf-8` a prosté `== "application/json"` by ji odmítlo.

`http.MaxBytesReader` je jediná správná obrana proti velkému tělu. Vrací
`*http.MaxBytesError`, takže ho odlišíš od syntaktické chyby a můžeš vrátit 413 místo 400.

Pozor na sémantiku zero value: `{"message":"a"}` bez pole `repeat` ti do struktury dá
`Repeat == 0`. Go neumí rozlišit „nebylo poslané" od „poslali nulu" — tuhle hranici musíš
v návrhu API vědomě rozhodnout (buď je nula validní default, nebo pole uděláš `*int`).

### `http.ServeMux` a `http.Server`

`http.ServeMux` je router ze stdlib. Registruje se do něj vzory, `mux` sám implementuje
`http.Handler`, takže je z něj složitelná struktura:

```go
mux := http.NewServeMux()
mux.Handle("/healthz", HealthHandler())
mux.Handle("/echo", EchoHandler())
mux.Handle("/", NotFoundHandler()) // catch-all fallback
```

Vzor bez lomítka na konci (`/echo`) je **přesná shoda**. Vzor s lomítkem (`/static/`) je
prefix a chytí všechno pod ním. Vzor `/` je prefix nad vším, takže je to tvůj 404 fallback
— a je to jediný způsob, jak mít 404 ve stejném JSON tvaru jako ostatní chyby. Bez něj
dostane klient `404 page not found` jako `text/plain`.

V téhle lekci registrujeme vzory **bez HTTP metody**, takže si každý handler kontroluje
metodu sám a vrací 405 s hlavičkou `Allow`. Od Go 1.22 to jde delegovat na mux
(`"GET /echo"`) — to je téma příští lekce.

Server nikdy nespouštěj takhle:

```go
// ŠPATNĚ v produkci
http.ListenAndServe(":8080", mux)
```

Tahle zkratka použije `http.Server` s **nulovými timeouty**, což znamená „čekej navždy".
Pomalý nebo zlomyslný klient ti drží spojení donekonečna a snadno vyčerpá file
descriptory (útok Slowloris). Vždy si server postav ručně:

```go
srv := &http.Server{
	Addr:              ":8080",
	Handler:           mux,
	ReadHeaderTimeout: 5 * time.Second,  // obrana proti Slowloris
	ReadTimeout:       10 * time.Second, // celé čtení požadavku
	WriteTimeout:      15 * time.Second, // celý zápis odpovědi
	IdleTimeout:       60 * time.Second, // keep-alive spojení bez provozu
}
```

`http.ListenAndServe` je určený pro příklady v dokumentaci, ne pro službu v provozu.

### Testování přes `httptest`

Handler je jen funkce, takže ho lze zavolat přímo — nepotřebuješ port, síť ani `WebTestCase`.

```go
req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
rec := httptest.NewRecorder()

HealthHandler().ServeHTTP(rec, req)

if rec.Code != http.StatusOK { /* ... */ }
res := rec.Result()          // *http.Response, včetně hlaviček
body, _ := io.ReadAll(res.Body)
```

`httptest.NewRecorder()` je „fake" `ResponseWriter`, který si zápis pamatuje v paměti.
Je rychlý a stačí na drtivou většinu testů. Když potřebuješ ověřit chování skutečného
serveru (keep-alive, chování klienta, rušení při odpojení), použij `httptest.NewServer`:

```go
srv := httptest.NewServer(NewRouter())
defer srv.Close()

res, err := srv.Client().Get(srv.URL + "/healthz")
```

`httptest.NewServer` si sám vybere volný port a `srv.URL` ti ho řekne. Nikdy do testu
nedrátuj `:8080` — v CI běží testy paralelně a port ti někdo vezme.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `Header().Set` po `WriteHeader` | v Symfony jde `Response` upravovat do poslední chvíle | hlavičky nastav jako první, pak status, pak tělo |
| Chybějící `return` po chybové odpovědi | zvyk na `throw`, který ukončí tok sám | po každém `writeError` napiš `return` |
| `http.Error` v JSON API | vypadá jako standardní řešení | vlastní `writeError` s `ErrorResponse` |
| `Decode` bez `MaxBytesReader` | v PHP limity řeší `php.ini` a webserver | vždy `http.MaxBytesReader` a 413 přes `*http.MaxBytesError` |
| `http.ListenAndServe` v `main` | je to první řádek v každém tutoriálu | vlastní `http.Server` se čtyřmi timeouty |
| Test startuje server na `:8080` | reflex z funkčních testů | `httptest.NewRecorder` nebo `httptest.NewServer` |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

Všechny odpovědi celé služby jsou JSON s hlavičkou `Content-Type: application/json`,
chyby mají jednotný tvar `ErrorResponse` (`{"error":"..."}`).

### A — rozcvička (~10 min)

1. `WriteJSON(w http.ResponseWriter, status int, v any) error` — nastaví
   `Content-Type: application/json`, pošle `status` a zakóduje `v` přes
   `json.NewEncoder`. Vrací chybu enkodéru. Pořadí operací musí být správné, test to hlídá.
2. `HealthHandler() http.Handler` — na jakoukoli metodu odpoví 200 a tělem
   `{"status":"ok"}` (typ `HealthResponse`).

### B — jádro (~35 min)

`EchoHandler() http.Handler` přijímá `POST` s JSON tělem `EchoRequest`. Postup a hraniční
případy:

1. Jiná metoda než `POST` → **405**, navíc hlavička `Allow: POST`.
2. `Content-Type`, jehož media type není `application/json` (včetně chybějící hlavičky) →
   **415**. Parametr `; charset=utf-8` je platný a musí projít — použij `mime.ParseMediaType`.
3. Tělo omez přes `http.MaxBytesReader` na `MaxBodyBytes`. Překročení → **413**
   (rozliš přes `errors.As` a `*http.MaxBytesError`).
4. Dekóduj s `DisallowUnknownFields()`. Nevalidní JSON i neznámé pole → **400**.
5. Validace: `Message` po ořezání bílých znaků nesmí být prázdný → **400**.
   `Repeat == 0` znamená „neposláno", doplň `1`. Výsledná hodnota mimo rozsah 1–10 → **400**.
6. Úspěch → **200** a `EchoResponse{Echo: <message zopakovaná Repeat×>, Count: <Repeat>}`.

Každá chybová větev vrací `ErrorResponse` s neprázdným textem.

### C — rozšíření (~25 min)

1. `NotFoundHandler() http.Handler` — vždy **404** a `ErrorResponse`.
2. `NewRouter() http.Handler` — `http.ServeMux` s `/healthz`, `/echo` a `/` jako
   fallback na `NotFoundHandler`. Registruj bez HTTP metody ve vzoru; metodu si hlídá
   `EchoHandler` sám.
3. `NewServer(addr string, h http.Handler) *http.Server` — vyplněná `Addr`, `Handler`
   a všechny čtyři timeouty kladnou hodnotou, přičemž `ReadHeaderTimeout` nesmí být delší
   než `ReadTimeout`.

```bash
make lesson L=24
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `24`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=24` prochází
- [ ] Umíš vysvětlit, proč v Go neexistuje `Response` objekt a co to dělá s error handlingem
- [ ] Umíš vysvětlit, co se stane při `Header().Set` po `WriteHeader`
- [ ] Umíš vysvětlit rozdíl mezi 400, 413 a 415 a kdy který vrátit
- [ ] Umíš zpaměti vyjmenovat čtyři timeouty `http.Server` a k čemu každý slouží
- [ ] Umíš rozhodnout, kdy stačí `httptest.NewRecorder` a kdy potřebuješ `httptest.NewServer`

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Od téhle fáze smí AI psát DTO structy, JSON tagy a kostry table-driven testů. Kontrakt
handleru, status kódy a hranice validace vlastníš ty.

## Další čtení

1. [pkg.go.dev — net/http](https://pkg.go.dev/net/http)
2. [pkg.go.dev — net/http/httptest](https://pkg.go.dev/net/http/httptest)
3. [Go blog — Writing Web Applications](https://go.dev/doc/articles/wiki/)
4. [Cloudflare blog — The complete guide to Go net/http timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/)
