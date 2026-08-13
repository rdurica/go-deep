# Lekce 26 — Middleware

> **Čas:** ~35 min · **Fáze:** 3 — net/http a tooling · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Napsat middleware jako `func(http.Handler) http.Handler` a poskládat jich několik za sebou.
- Odvodit z pořadí v řetězu, který kód poběží první a který poslední.
- Obalit `http.ResponseWriter` tak, abys zjistil status a velikost odpovědi a nic přitom nerozbil.
- Zachytit paniku v handleru dřív, než shodí celý proces, a vědět, kterou paniku spolknout nesmíš.
- Vysvětlit, čím se explicitní kompozice liší od Symfony EventListenerů s prioritami.

## Teorie

### Typ middlewaru a řetězení

Podpis je vždy stejný, takže se vyplatí ho pojmenovat:

```go
type Middleware func(http.Handler) http.Handler
```

Vnořování `A(B(C(h)))` je čitelné pro tři middlewary a nesnesitelné pro osm. Pomocná
funkce z toho udělá seznam:

```go
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

handler := Chain(mux, RequestID(), Logging(logger), Recovery())
```

Iteruje se pozpátku schválně: chceme, aby **první uvedený middleware byl nejvíc vnější**,
protože tak se seznam čte stejně jako sekvence, kterou požadavek prochází.

Teď to nejdůležitější, na čem lidé pravidelně ujíždějí. Vnější middleware se spustí
**první**, ale jeho kód **za** `next.ServeHTTP` běží **poslední**:

```
A:před → B:před → handler → B:po → A:po
```

Praktické důsledky:

- **Recovery patří dovnitř, blízko k handleru.** Jinak zachytí paniku až za middlewary,
  které mezitím stihly zapsat do odpovědi.
- **Logging patří vně Recovery**, aby viděl status 500, který Recovery zapsala.
- **RequestID patří úplně navrch**, aby ID bylo v kontextu pro všechny ostatní.

### Obalení `http.ResponseWriter`

Logger potřebuje status kód, jenže `http.ResponseWriter` žádný getter nemá — je to
jednosměrný zapisovač. Řešení je obalit ho vlastním typem s vloženým interfacem:

```go
type statusRecorder struct {
	http.ResponseWriter // embedding: zdědíš Header() a zbytek zdarma

	status      int
	bytes       int
	wroteHeader bool
}

func (rec *statusRecorder) WriteHeader(status int) {
	if rec.wroteHeader {
		return // druhé volání je chyba handleru, nesmíme ho poslat dál
	}
	rec.wroteHeader = true
	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
}

func (rec *statusRecorder) Write(b []byte) (int, error) {
	if !rec.wroteHeader {
		rec.WriteHeader(http.StatusOK) // implicitní 200, stejně jako to dělá stdlib
	}
	n, err := rec.ResponseWriter.Write(b)
	rec.bytes += n
	return n, err
}
```

Tři pasti, které tady číhají:

1. **Výchozí hodnota musí být 200.** Handler, který jen zapíše tělo, `WriteHeader` nezavolá
   nikdy. Bez inicializace bys logoval status `0`.
2. **`WriteHeader` smíš přeposlat jen jednou.** Jinak ti server do logu vysype
   `superfluous response.WriteHeader call` u každého požadavku, kde handler chyboval.
3. **Embedding ti zahodí rozšiřující interfacy.** Původní writer může implementovat
   `http.Flusher`, `http.Hijacker` nebo `io.ReaderFrom`; tvůj obal je `*statusRecorder`,
   takže typová aserce v cizím kódu selže. Streaming (SSE) nebo WebSocket upgrade se tím
   rozbije. Když je potřebuješ, musíš je explicitně dopsat:

```go
func (rec *statusRecorder) Flush() {
	if f, ok := rec.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
```

Od Go 1.20 existuje `http.ResponseController`, který tenhle problém řeší elegantněji —
`http.NewResponseController(w).Flush()` si obalené writery rozbalí sám, pokud mají metodu
`Unwrap() http.ResponseWriter`. Přidat ji je jeden řádek a vyplatí se.

### Recovery a panika v handleru

`net/http` paniku v handleru zachytí a shodí jen dané spojení, ne celý proces. To zní
uklidňujícím dojmem, jenže klient dostane useknuté spojení bez odpovědi a ty ve svém logu
nic strukturovaného. Vlastní recovery middleware to napraví:

```go
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}
				if recovered == http.ErrAbortHandler {
					panic(recovered) // patří serveru, nesmíme ho spolknout
				}
				_ = WriteJSON(w, http.StatusInternalServerError,
					ErrorResponse{Error: "internal server error"})
			}()

			next.ServeHTTP(w, r)
		})
	}
}
```

Dvě věci, které se často zapomínají:

- `http.ErrAbortHandler` je **smluvená** panika, kterou handler říká „ukonči spojení bez
  logu". `httputil.ReverseProxy` ji používá běžně. Musíš ji poslat dál.
- Text paniky **nikdy** neposílej klientovi. Do odpovědi patří obecná hláška, do logu
  detail včetně `debug.Stack()`.

A jedno omezení, se kterým nic neuděláš: pokud handler stihl zapsat tělo a panikuje až
potom, status je dávno na drátě a recovery už odpověď nezmění. Znovu ten samý motiv —
validuj dřív, než začneš psát.

### CORS a Timeout stručně

CORS je v základní podobě jen několik hlaviček plus obsloužený preflight:

```go
func CORS(origin string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return // preflight nesmí jít do handleru
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

Nikdy nepiš `Access-Control-Allow-Origin: *` současně s
`Access-Control-Allow-Credentials: true` — prohlížeč to odmítne a je to bezpečnostní
antipattern.

Timeout nemusíš psát ručně, stdlib ho má:

```go
func Timeout(d time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, d, `{"error":"request timeout"}`)
	}
}
```

`http.TimeoutHandler` udělá dvě věci naráz: nastaví requestu kontext s deadline a po jeho
vypršení odešle **503** i v případě, že handler pořád běží. Handler tím ale nezabije —
goroutina běží dál, dokud sama nezareaguje na `ctx.Done()`. Zrušení v Go je vždycky
spolupráce, ne přerušení. To je téma příští lekce.

## Rozdíly proti PHP

V Symfony se do životního cyklu požadavku zapojuješ listenerem na `kernel.request`.
Kdo poběží dřív, určuje číslo v konfiguraci — a to číslo je jinde než kód.

```php
final class RequestIdListener
{
    #[AsEventListener(event: KernelEvents::REQUEST, priority: 250)]
    public function onKernelRequest(RequestEvent $event): void { /* ... */ }
}
```

V Go je middleware obyčejná funkce, která dostane handler a vrátí jiný handler:

```go
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ... něco před ...
		next.ServeHTTP(w, r)
		// ... něco po ...
	})
}

handler := Logging(Recovery(RequestID(mux))) // pořadí je vidět na jednom řádku
```

Co se mění v uvažování: **pořadí není konfigurace, je to kód.** Nemusíš hledat, jaká
priorita se kde nastavila; přečteš si jeden výraz. Zároveň tím ztrácíš možnost „přidat
listener zvenčí" — což je v praxi spíš úleva, protože skryté zásahy do request pipeline
patří mezi nejhůř laditelné věci ve velké Symfony aplikaci.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Recovery jako nejvnější middleware | „chci chytit úplně všechno" | Recovery patří blízko handleru, Logging nad ni |
| Zapomenuté `next.ServeHTTP` | middleware vypadá hotově i bez něj | bez zavolání se požadavek tiše ztratí a vrátí 200 s prázdným tělem |
| Obalený writer bez výchozího statusu 200 | handler `WriteHeader` volat nemusí | inicializuj `status: http.StatusOK` |
| Obalený writer bez `Flush`/`Unwrap` | embedding vypadá, že přenese všechno | doplň metody nebo `http.ResponseController` |
| Hledání priorit místo čtení kódu | reflex z `kernel.request` listenerů | pořadí je viditelné v `Chain(...)` |
| Text paniky v odpovědi klientovi | ladicí zvyk z `dev` prostředí | obecná hláška ven, `debug.Stack()` do logu |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 26`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. `WriteJSON` je hotové — soustřeď se na body níže. Stupně jdou od
jednodušších ke složitějším; po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `statusRecorder` (`WriteHeader` / `Write` — kód je záměrně vadný)

```bash
make lesson L=26 PART=1
```

Pak **`/go-deep-review 26 easy`**.

### Střední

Implementuj: `Chain`, `Logging`

```bash
make lesson L=26 PART=2
```

Pak **`/go-deep-review 26 medium`**.

### Obtížný

Doplň: `Recovery` (`http.ErrAbortHandler` nesmíš spolknout)

```bash
make lesson L=26 PART=3
```

Pak **`/go-deep-review 26 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 26 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=26` (+ `make race L=26`, pokud to lekce vyžaduje).

- [ ] Umíš z výrazu `Chain(mux, A, B, C)` odvodit pořadí „před" i „po" krocích
- [ ] Umíš vysvětlit, proč Recovery patří pod Logging a ne nad něj
- [ ] Umíš vyjmenovat tři věci, které se rozbijí při naivním obalení `ResponseWriter`
- [ ] Umíš vysvětlit, proč `http.ErrAbortHandler` nesmíš spolknout
- [ ] Umíš vysvětlit, proč `TimeoutHandler` handler nezabije, jen mu pošle signál

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Middleware je oblíbený generovaný boilerplate a taky místo, kde AI ochotně vyrobí obalený
writer bez `Flush`, bez ochrany proti dvojímu `WriteHeader` a s textem paniky v odpovědi.
Když si necháš `statusRecorder` napsat, projdi ho podle tabulky Časté chyby řádek po řádku.

## Další čtení

1. [pkg.go.dev — http.TimeoutHandler](https://pkg.go.dev/net/http#TimeoutHandler)
2. [pkg.go.dev — http.ResponseController](https://pkg.go.dev/net/http#ResponseController)
3. [Go blog — Structured Logging with slog](https://go.dev/blog/slog)
4. [pkg.go.dev — http.ErrAbortHandler](https://pkg.go.dev/net/http#pkg-variables)
