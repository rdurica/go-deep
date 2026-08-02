# Lekce 29 — slog: strukturované logování

> **Čas:** ~90 min · **Fáze:** 3 — net/http a tooling · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Vysvětlit, proč se v produkci loguje strukturovaně a co to znamená pro dohledávání
  incidentů.
- Sestavit `slog.Logger` s vhodným handlerem a úrovní a předat ho jako závislost.
- Rozhodnout mezi `slog.String("k", v)` a variantou klíč-hodnota a vědět, co riskuješ.
- Napsat vlastní `slog.Handler`, který obaluje jiný handler a mění záznamy.
- Otestovat, co přesně kód zaloguje — a ověřit, že tam není heslo.

## Teorie

### Proč strukturovaně

Nestrukturovaný log je věta pro člověka:

```
2026-07-25 10:12:03 payment charged order 8123 in 145ms
```

Strukturovaný log je záznam pro stroj:

```json
{"time":"2026-07-25T10:12:03Z","level":"INFO","msg":"charged","component":"payment","order_id":"8123","duration":145000000}
```

Rozdíl se projeví ve chvíli, kdy máš deset instancí služby a někdo se ptá „proč byly
včera mezi 10:00 a 10:30 platby pomalé“. Nad prvním formátem píšeš `grep` a regulární
výrazy; nad druhým se ptáš `component:payment AND duration > 1s`. Log v kontejneru navíc
jde na stdout a sbírá ho platforma — soubory ani rotace tě nezajímají.

### `slog.Logger` a handlery

`slog` má dvě vrstvy. `Logger` je fasáda s metodami `Debug/Info/Warn/Error`, `Handler`
rozhoduje, co se se záznamem stane.

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))
logger.Info("server started", slog.String("addr", ":8080"))
```

Ve stdlib jsou dva handlery: `TextHandler` (`key=value`, čitelné při vývoji)
a `JSONHandler` (jeden JSON objekt na řádek, pro produkci). Od Go 1.26 je i
`slog.NewMultiHandler`, když chceš fan-out do více handlerů najednou (např. JSON
do souboru a text na stderr) bez vlastního wrapperu. Balíček má i globální funkce
`slog.Info(...)`, které píšou do výchozího loggeru. Používej je nanejvýš v `main`
a v jednorázových skriptech — jinak platí totéž co pro každý globál: nejde vyměnit
v testu a skryje závislost.

`HandlerOptions{Level}` je filtr. Když chceš úroveň měnit za běhu (třeba přes signál nebo
endpoint), použij `slog.LevelVar`:

```go
var lvl slog.LevelVar          // výchozí Info
lvl.Set(slog.LevelDebug)       // bezpečné i souběžně
h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: &lvl})
```

### Atributy: silné typy vs klíč-hodnota

`slog` přijímá atributy dvěma způsoby:

```go
logger.Info("processed", "id", 42, "ok", true)                 // klíč-hodnota
logger.Info("processed", slog.Int("id", 42), slog.Bool("ok", true)) // silné typy
```

První varianta je kratší, ale je to `...any` — když zapomeneš hodnotu, kompilátor mlčí
a v logu ti přistane `!BADKEY`:

```go
logger.Info("processed", "id")   // kompiluje se, ale záznam je rozbitý
```

Druhá varianta je typovaná, `slog.Int` chce dva argumenty. Navíc `slog.Duration`,
`slog.Time` a `slog.Any` řeknou handleru, jak má hodnotu serializovat. Pravidlo:
klíč-hodnota v aplikačním kódu, kde to čteš očima; `slog.Xxx` v knihovnách, sdílených
helperech a všude, kde hodnotu skládáš z proměnných.

Na horké cestě (middleware, smyčka nad tisíci záznamy) sáhni po `LogAttrs`. Je to jediná
varianta, která nealokuje `[]any` a nemusí atributy odvozovat za běhu:

```go
logger.LogAttrs(ctx, slog.LevelInfo, "http_request",
	slog.String("method", r.Method),
	slog.Int("status", status),
	slog.Duration("duration", time.Since(start)),
)
```

### Odvozený logger, skupiny a kontext

`With` vrátí nový logger, který ke každému záznamu přidá dané atributy. To je Go
protějšek Monolog kanálu a zároveň způsob, jak propagovat request ID:

```go
reqLog := logger.With("request_id", id, "path", r.URL.Path)
reqLog.Info("start")   // oba atributy jsou v záznamu
reqLog.Error("failed") // taky
```

`slog.Group` vytvoří vnořenou strukturu — v JSONu objekt, v textu prefix `auth.token`:

```go
logger.Info("call", slog.Group("auth",
	slog.String("user", "radek"),
	slog.String("scope", "read"),
))
```

Metody s příponou `Context` (`InfoContext`, `LogAttrs`) předají `context.Context`
handleru. Vlastní handler z něj může vytáhnout trace ID nebo request ID a přidat ho do
každého záznamu, aniž by to musel řešit volající.

### Vlastní `Handler`

Interface má čtyři metody:

```go
type Handler interface {
	Enabled(context.Context, Level) bool
	Handle(context.Context, Record) error
	WithAttrs(attrs []Attr) Handler
	WithGroup(name string) Handler
}
```

Nejužitečnější vzor je **obalení** existujícího handleru: `Enabled` a `WithGroup` deleguj,
v `Handle` a `WithAttrs` udělej svou práci. Pozor na dvě věci. Za prvé `Record` se
needituje na místě — postav nový přes `slog.NewRecord` a naplň ho. Za druhé
`WithAttrs`/`WithGroup` musí vracet **nový** handler, ne měnit ten svůj; jinak si dva
odvozené loggery přepíšou stav.

Pro jednoduché maskování existuje i levnější cesta bez vlastního handleru —
`HandlerOptions.ReplaceAttr`. Vlastní handler potřebuješ, když chceš měnit i strukturu
záznamu nebo sahat na kontext.

### Co se do logu nesmí dostat

Hesla, tokeny, API klíče, čísla karet, rodná čísla, e-maily a IP adresy podle kontextu.
Log jde do centrálního úložiště, replikuje se, zálohuje a čte ho víc lidí než databázi.
Když tam jednou tajemství spadne, prakticky ho nesmažeš — a token musíš rotovat.

Disciplína „nebudu to logovat“ nestačí, protože log píše i kód, který nepíšeš ty.
Proto se maskování dělá **na úrovni handleru**: co projde handlerem, je vždycky
vyčištěné, ať to zavolá kdokoli.

### Testování logů

Logger píše do `io.Writer`, takže test místo stdoutu podstrčí buffer, provede akci
a rozparsuje JSON:

```go
var buf bytes.Buffer
logger := slog.New(slog.NewJSONHandler(&buf, nil))
svc := NewService(logger)
svc.Process("42")

var rec map[string]any
_ = json.Unmarshal(buf.Bytes(), &rec)
if rec["level"] != "INFO" { … }
```

Když logger píše z jiné goroutiny (HTTP handler), obal buffer mutexem — jinak
ti `-race` právem vynadá.

## Rozdíly proti PHP

Monolog má kanály, handlery, procesory a formátory a konfiguruje se v YAML. Logger se
do služby dostane autowiringem, kanál se vybere jménem parametru:

```php
public function __construct(private LoggerInterface $paymentLogger) {}

public function charge(Order $o): void
{
    $this->paymentLogger->info('charged', ['order_id' => $o->getId()]);
}
```

Go má od verze 1.21 `log/slog` přímo ve stdlib. Konfigurace není v YAML, ale v kódu, a
„kanál“ nahradíš odvozeným loggerem s trvalými atributy:

```go
type PaymentService struct{ log *slog.Logger }

func NewPaymentService(logger *slog.Logger) *PaymentService {
	return &PaymentService{log: logger.With("component", "payment")}
}

func (s *PaymentService) Charge(o Order) {
	s.log.Info("charged", slog.String("order_id", o.ID))
}
```

Co se mění v uvažování: přestaň logger hledat („kde vezmu logger?“) a začni ho **přijímat**.
A přestaň skládat větu — `"Charged order ".$id` v Go neplatí. Zpráva je krátká konstanta,
proměnné jsou atributy. Jen tak je log strojově zpracovatelný.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `logger.Info(fmt.Sprintf("charged %s", id))` | zvyk skládat větu jako v Monologu | zpráva je konstanta, `id` je atribut |
| Globální `slog.Info` v celé aplikaci | připomíná statické `Log::info()` | logger přijmi konstruktorem |
| Lichý počet argumentů v klíč-hodnota | `...any` nic nekontroluje | `slog.String/Int/...` nebo `LogAttrs` |
| Logger uložený spolu s `context.Context` ve structu | reflex „mít všechno po ruce“ | logger do structu, context jen jako parametr |
| Heslo v atributu | logují se celé DTO / mapy | maskuj v handleru, ne v místě volání |
| Log a zároveň návrat chyby na každé úrovni | strach, že se chyba ztratí | zaloguj tam, kde chybu řešíš; jinak jen obal `%w` |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 29`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `NewLogger`, `LogRequest`, `NewService`

```bash
make lesson L=29 PART=1
```

Pak **`/go-deep-review 29 easy`**.

### Střední

Funkce: `Process`, `NewRedactingHandler`, `Enabled`

```bash
make lesson L=29 PART=2
```

Pak **`/go-deep-review 29 medium`**.

### Obtížný

Funkce: `Handle`, `WithAttrs`, `WithGroup`, `LoggingMiddleware`

```bash
make lesson L=29 PART=3
```

Pak **`/go-deep-review 29 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 29 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=29` (+ `make race L=29`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč `msg` nemá obsahovat proměnné hodnoty
- [ ] Umíš vysvětlit rozdíl mezi `logger.Info(...)` a `logger.LogAttrs(...)`
- [ ] Umíš vysvětlit, proč `WithAttrs` musí vracet nový handler
- [ ] Umíš vysvětlit, proč se maskování tajemství dělá v handleru
- [ ] Víš, jak zjistit status kód, který handler odeslal

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).
Boilerplate čtyř metod `slog.Handler` klidně nech vygenerovat, ale ověř si sám, že
`WithAttrs` nemutuje původní handler a že skupiny procházíš rekurzivně.

## Další čtení

1. [pkg.go.dev — log/slog](https://pkg.go.dev/log/slog)
2. [Go blog — Structured Logging with slog](https://go.dev/blog/slog)
3. [A Guide to Writing slog Handlers](https://github.com/golang/example/blob/master/slog-handler-guide/README.md)
