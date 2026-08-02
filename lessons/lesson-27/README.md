# Lekce 27 — context v request scope

> **Čas:** ~90 min · **Fáze:** 3 — net/http a tooling · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Rozhodnout, co do `context.Context` patří (rušení, deadline, request-scoped hodnoty) a co ne.
- Použít `WithCancel`, `WithTimeout` a `WithDeadline` bez toho, abys zapomněl na `cancel`.
- Uložit hodnotu do kontextu typově bezpečně přes neexportovaný typ klíče.
- Napsat handler, který skutečně skončí, když se klient odpojí.
- Vysvětlit, proč zrušení v Go není přerušení, ale dohoda.

## Teorie

### K čemu context je a k čemu není

`context.Context` je interface se čtyřmi metodami a jediným účelem: **nést informaci
o životnosti operace napříč hranicemi API a goroutin.**

Patří do něj:

- signál zrušení (`Done()`, `Err()`),
- deadline (`Deadline()`),
- hodnoty vázané na požadavek, které projdou celým stromem volání — ID požadavku,
  autentizovaný uživatel, trace ID.

Nepatří do něj:

- volitelné parametry funkce („když to nechci předávat v signatuře, dám to do contextu"),
- závislosti (logger, databáze, konfigurace) — ty patří do struktury nebo do konstruktoru,
- cokoli, bez čeho funkce **nemůže** fungovat. Hodnota v kontextu je vždy `any` a její
  chybějící přítomnost zjistíš až za běhu.

Test, jestli hodnota do kontextu patří: *fungovala by funkce správně, kdyby ta hodnota
chyběla?* U request ID ano (zaloguje se prázdné). U ID objednávky, který funkce zpracovává,
ne — ten patří do parametru.

### Kořeny a odvozování

Kořen stromu je `context.Background()` — prázdný, nikdy nezrušený. `context.TODO()` je
totéž, ale znamená „ještě nevím, odkud sem kontext povede"; použij ho při refaktoringu,
ať je v kódu vidět nedodělek.

Z kořene odvozuješ potomky. Každý potomek se zruší, když se zruší rodič — nikdy naopak.

```go
ctx, cancel := context.WithCancel(parent)
defer cancel() // POVINNÉ, i když se ctx zruší jinak

ctx, cancel := context.WithTimeout(parent, 2*time.Second)
defer cancel()

ctx, cancel := context.WithDeadline(parent, time.Now().Add(2*time.Second))
defer cancel()
```

`defer cancel()` **není volitelný**. Bez něj zůstane potomek zaregistrovaný u rodiče,
dokud rodič nezanikne — u dlouho žijícího kontextu je to učebnicový memory leak.
`go vet` tuhle chybu hlásí (`lostcancel`), tak ho poslouchej.

Zrušení se čte přes kanál:

```go
select {
case <-ctx.Done():
	return ctx.Err() // context.Canceled nebo context.DeadlineExceeded
case res := <-work:
	return res
}
```

`ctx.Err()` ti řekne *že* se zrušilo, ale ne *proč*. Od Go 1.20 máš
`context.WithCancelCause(parent)`, kde `cancel(err)` uloží důvod a `context.Cause(ctx)`
ho vrátí:

```go
ctx, cancel := context.WithCancelCause(parent)
cancel(errors.New("upstream vrátil 500"))

<-ctx.Done()
fmt.Println(ctx.Err())         // context.Canceled — obecné
fmt.Println(context.Cause(ctx)) // upstream vrátil 500 — konkrétní
```

### Context jako první parametr, nikdy ve struct fieldu

Konvence, kterou dodržuje celá stdlib:

```go
// SPRÁVNĚ
func (s *Service) Find(ctx context.Context, id string) (Order, error)

// ŠPATNĚ — kontext má životnost požadavku, struktura žije déle
type Service struct {
	ctx context.Context // ne
	db  *sql.DB
}
```

Kontext uložený ve struktuře znamená, že všechny požadavky sdílí jeden deadline a jedno
zrušení. Buď se zruší předčasně všem, nebo nikomu. Jediná výjimka, kterou stdlib připouští,
je `http.Request` — a ten je sám o sobě per-request hodnota.

### `WithValue` a proč klíč nesmí být string

```go
// ŠPATNĚ — klíč typu string se může srazit s klíčem cizího balíčku
ctx = context.WithValue(ctx, "user", u)

// SPRÁVNĚ — neexportovaný typ, který nikdo zvenčí nevyrobí
type userKey struct{}

func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userKey{}, u)
}

func UserFrom(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey{}).(User)
	return u, ok
}
```

`ctx.Value` hledá podle **rovnosti hodnoty klíče**. Když dva balíčky použijí řetězec
`"user"`, přepíšou si navzájem hodnotu a nikdo si toho nevšimne — kompilátor mlčí, protože
typy sedí. Prázdný struct jako klíč to řeší: `userKey{}` z tvého balíčku se nikdy nerovná
`userKey{}` z jiného, protože je to jiný typ. Navíc nezabírá ani bajt.

Druhé pravidlo: **hodnotu z kontextu vždy vystav přes dvojici funkcí**, ne přímým
`ctx.Value(...)` u volajícího. Typová aserce je pak na jednom místě a `UserFrom` má
poctivou signaturu `(User, bool)`.

Pozor, `WithValue` vytváří **nový** kontext — rodič zůstává beze změny. V handleru proto
musíš request nahradit:

```go
next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
```

### `r.Context()` v HTTP handleru

Každý request má kontext, který server zruší, když:

- klient zavře spojení nebo mu vyprší timeout,
- handler skončí (po návratu ze `ServeHTTP`),
- vyprší deadline nastavený `http.TimeoutHandler` nebo `Server.ReadTimeout`.

Tenhle kontext propaguj do všeho, co může trvat: databázových dotazů (`db.QueryContext`),
HTTP volání (`http.NewRequestWithContext`), interních goroutin. Odměna je konkrétní:
když uživatel zavře záložku, přestaneš platit za dotaz, jehož výsledek nikoho nezajímá.

Jenže — a to je jádro téhle lekce — **zrušení nikoho nezabije**. Go nemá způsob, jak
goroutinu přerušit zvenčí. Kontext je jen zavřený kanál; kód, který se na něj nedívá,
běží vesele dál:

```go
// ŠPATNĚ — ctx je k ničemu, tenhle handler poběží celé 3 vteřiny vždycky
func slow(w http.ResponseWriter, r *http.Request) {
	time.Sleep(3 * time.Second)
	fmt.Fprint(w, "hotovo")
}

// SPRÁVNĚ — práce po kouscích s kontrolou zrušení
func slow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	deadline := time.Now().Add(3 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return // klient je pryč, končíme
		case <-ticker.C:
			if !time.Now().Before(deadline) {
				fmt.Fprint(w, "hotovo")
				return
			}
		}
	}
}
```

Poslední detail: po zrušení už do `w` **nepiš**. Spojení je zavřené a zápis stejně nikam
nedojde; u `http.TimeoutHandler` bys navíc zapisoval do writeru, který už odeslal 503.

## Rozdíly proti PHP

V Symfony si „co se teď děje" vytáhneš odkudkoli — `RequestStack` je služba, kterou si
necháš injektovat i deset vrstev hluboko, a ona ti aktuální request najde.

```php
final class AuditLogger
{
    public function __construct(private RequestStack $requestStack) {}

    public function log(string $action): void
    {
        $request = $this->requestStack->getCurrentRequest(); // odkudkoli, kdykoli
        $user = $request?->attributes->get('user');
        // ...
    }
}
```

V Go žádný ambientní stav neexistuje. Request-scoped informace musí projít **explicitně
skrz všechny signatury** jako první parametr:

```go
func (a *AuditLogger) Log(ctx context.Context, action string) error {
	user, ok := UserFrom(ctx) // hodnota přišla s kontextem, ne z kontejneru
	// ...
}
```

Co se mění v uvažování: kontext se **předává**, nikdy se nehledá. Vypadá to jako otravný
boilerplate — dokud nezjistíš, že díky němu jde z libovolného místa v kódu zrušit celý
strom volání jedním `cancel()`, a že u každé funkce hned vidíš, jestli může trvat dlouho.
`RequestStack` ti nikdy neřekne, že se klient odpojil. Kontext ano.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Chybějící `defer cancel()` | „vždyť to vyprší samo" | vždy `defer cancel()` hned po vytvoření; `go vet` to hlídá |
| `context.WithValue(ctx, "user", u)` | řetězec jako klíč vypadá přirozeně | neexportovaný typ klíče a dvojice `WithX`/`XFrom` |
| Kontext ve struct fieldu | reflex „dám si to do služby jako RequestStack" | první parametr každé metody |
| Volitelné parametry v kontextu | vypadá to jako elegantní zkratka | co funkce potřebuje, patří do signatury |
| `time.Sleep` v handleru | zvyk z PHP, kde request nejde zrušit | `select` nad `ctx.Done()` a tickerem |
| Zápis do `w` po `ctx.Done()` | „ještě to zkusím doručit" | po zrušení jen `return` |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 27`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `WriteJSON`, `WithUser`

```bash
make lesson L=27 PART=1
```

Pak **`/go-deep-review 27 easy`**.

### Střední

Funkce: `UserFrom`, `Authenticate`, `WhoAmI`

```bash
make lesson L=27 PART=2
```

Pak **`/go-deep-review 27 medium`**.

### Obtížný

Funkce: `FetchWithTimeout`, `SlowHandler`, `SlowHandlerWithHook`

```bash
make lesson L=27 PART=3
```

Pak **`/go-deep-review 27 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 27 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=27` (+ `make race L=27`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, co do kontextu patří a co ne, a máš na to jednovětý test
- [ ] Umíš vysvětlit, proč klíč do `WithValue` nesmí být `string`
- [ ] Umíš vysvětlit, proč `defer cancel()` není volitelný
- [ ] Umíš vysvětlit rozdíl mezi `ctx.Err()` a `context.Cause(ctx)`
- [ ] Umíš vysvětlit, proč zrušený kontext sám o sobě žádnou goroutinu nezastaví

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Dvojice `WithX`/`XFrom` je ideální kandidát na vygenerování. Naopak všechno kolem rušení
si projdi ručně: nejčastější vada generovaného kódu je goroutina zapisující do
nebufferovaného kanálu, kterou po timeoutu nikdo nepřečte — leak, který se v testech
neprojeví a v provozu ti sežere paměť.

## Další čtení

1. [Go blog — Go Concurrency Patterns: Context](https://go.dev/blog/context)
2. [pkg.go.dev — context](https://pkg.go.dev/context)
3. [Go Wiki — Contexts and structs](https://go.dev/blog/context-and-structs)
4. [pkg.go.dev — http.Request.Context](https://pkg.go.dev/net/http#Request.Context)
