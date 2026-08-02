# Lekce 20 — Konstruktory, inicializace a design API

> **Čas:** ~90 min · **Fáze:** 2 — Idiomatický Go · **AI režim:** `JEN VYSVĚTLENÍ`

## Co budeš umět

- Rozhodnout, jestli typ vůbec potřebuje konstruktor, nebo mu stačí užitečná zero value.
- Napsat konstruktor s validací, který vrací `(T, error)`, a jeho `Must` variantu.
- Rozdělit závislosti na povinné (parametr) a volitelné (functional options) a poznat,
  kdy jsou options overkill.
- Aplikovat „accept interfaces, return structs" a vysvětlit, proč to není libovůle.
- Nahradit Symfony DI kontejner ručním wiringem v `main`, aniž se ti to zvrhne v peklo.

## Teorie

### Nejlepší konstruktor je žádný konstruktor

Než napíšeš `New`, zeptej se, jestli zero value typu nedává smysl sama o sobě. Stdlib
tenhle přístup používá všude, kde může:

```go
var buf bytes.Buffer      // připravený
var mu sync.Mutex         // odemčený
var wg sync.WaitGroup     // nulový čítač
var mux http.ServeMux     // prázdný router
```

Trik, který to umožňuje u typů s mapou nebo slicem, je **líná inicializace** v mutující
metodě:

```go
type Registry struct {
	entries map[string]string
}

func (r *Registry) Set(key, value string) {
	if r.entries == nil {
		r.entries = make(map[string]string) // až tady, ne v konstruktoru
	}
	r.entries[key] = value
}

func (r *Registry) Lookup(key string) (string, bool) {
	v, ok := r.entries[key] // čtení z nil mapy je legální, nic řešit nemusíš
	return v, ok
}
```

Užitečná zero value má konkrétní přínos: typ jde vložit do jiného structu jako pole a nikdo
ho nemusí inicializovat. `var a app` prostě funguje.

Konstruktor napiš, když platí aspoň jedno: typ má **povinné závislosti**, které nejde
odvodit, nebo vstup potřebuje **validaci**, nebo vnitřní stav vyžaduje netriviální
přípravu (kanály, goroutiny, předalokované buffery).

### `New` vs `NewFoo` a co vracet

Uvnitř balíčku `user` se konstruktor hlavního typu jmenuje `New`, ne `NewUser` — volá se
`user.New()`. Když balíček exportuje víc typů, které se konstruují, dostane každý svoje
jméno: `slog.NewTextHandler`, `slog.NewJSONHandler`.

Konstruktor s validací vrací `(*T, error)`. Když se validace nepovede, vrať **`nil` a chybu**,
nikdy poloviční objekt:

```go
func NewServer(addr string, logger *slog.Logger) (*Server, error) {
	if addr == "" {
		return nil, ErrMissingAddr
	}
	if logger == nil {
		return nil, ErrMissingLogger
	}
	return &Server{addr: addr, logger: logger}, nil
}
```

Varianta `Must` je konvence pro případy, kdy vstup pochází z konstanty v kódu, ne od
uživatele — proto smí panikovat. Používá ji `regexp.MustCompile` i `template.Must`.
`Must` píšeš **navíc**, nikdy místo verze s chybou.

```go
func MustNewServer(addr string, logger *slog.Logger) *Server {
	s, err := NewServer(addr, logger)
	if err != nil {
		panic(err)
	}
	return s
}
```

### Povinné parametry vs functional options

Povinná závislost patří do signatury. Když ji tam dáš, kompilátor za tebe hlídá, že ji
nikdo nezapomene. Volitelná konfigurace do signatury nepatří, protože každý další parametr
rozbije všechny volající.

Vzor, který na to Go používá, je **functional options**: option je funkce, která mutuje
rozpracovaný objekt.

```go
type Option func(*Client)

func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

func NewClient(baseURL string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		return nil, ErrMissingBaseURL
	}

	c := &Client{baseURL: baseURL, timeout: DefaultTimeout, retries: DefaultRetries}
	for _, opt := range opts {
		opt(c)
	}

	if c.timeout <= 0 { // validuj až PO aplikaci options
		return nil, ErrInvalidTimeout
	}
	return c, nil
}
```

Pořadí kroků je podstatné: nejdřív povinné vstupy, pak výchozí hodnoty, pak options,
**pak teprve validace celku**. Kdo validuje před options, ověřuje hodnoty, které nikdy
nikdo neuvidí.

Kdy jsou options overkill? Když má typ jednu nebo dvě volitelné položky a nečekáš další.
Pak stačí exportovaný `Config` struct předaný hodnotou. Options se vyplatí od zhruba tří
voleb výš, u veřejného API knihovny a všude, kde potřebuješ rozlišit „nenastaveno" od
„nastaveno na zero value".

### Accept interfaces, return structs

Pravidlo zní: **na vstupu ber to nejobecnější, co ti stačí; na výstupu vracej to
nejkonkrétnější, co máš.**

```go
// Store definuje KONZUMENT, protože jen on ví, co potřebuje.
type Store interface {
	Save(Record) error
	All() []Record
}

type Service struct {
	store Store
}

func NewService(store Store) (*Service, error) { // přijímá interface
	if store == nil {
		return nil, ErrMissingStore
	}
	return &Service{store: store}, nil // vrací konkrétní typ
}
```

Proč na výstupu ne interface? Protože konkrétní typ může časem dostat další metodu, aniž
to kohokoli rozbije. Kdybys vracel interface, každá nová metoda by byla breaking change,
volající by ztratil přístup k ostatním metodám a `godoc` by ukazoval prázdno.

Z toho plyne i konkrétní anti-vzor, který AI generuje pořád: **neexportovaný typ vracený
jako exportovaný interface**.

```go
// ŠPATNĚ — uživatel nemá jak si typ pojmenovat, vložit do structu ani rozšířit
type Servicer interface{ Do() error }

type service struct{}

func New() Servicer { return &service{} }
```

Test na interface u konzumenta: **interface patří do balíčku, který ho volá, ne do toho,
který ho implementuje.** V Symfony je to naopak (interface leží u implementace v tomtéž
namespace) — tenhle reflex musí pryč. Implementace v Go interface „neví", že ho splňuje.

### Ruční wiring není utrpení

Symfony vývojář si po přečtení předchozích odstavců představí `main.go` na tisíc řádků.
V praxi se to nestane, protože kompozice je stromová a každá úroveň skládá jen svoje děti.
Pokud ti `main` přeroste přes zhruba sto řádků, extrahuj funkci
`func buildApp(cfg Config) (*App, error)` a hotovo.

Co ručním wiringem získáš: žádná reflexe za běhu, žádný cache warmup, chybějící závislost
je chyba kompilace místo výjimky při startu, a graf závislostí si přečteš shora dolů.

## Rozdíly proti PHP

V Symfony konstruktor typicky jen přiřazuje a o sestavení se stará kontejner. Autowiring
podle typu závislostí najde implementaci a předá ji:

```php
final class ApiClient
{
    public function __construct(
        private readonly HttpClientInterface $http,
        private readonly LoggerInterface $logger,
        private readonly int $timeout = 5,
    ) {}
}
// services.yaml se postará o zbytek, konstruktor nikdo ručně nevolá
```

V Go žádný kontejner není. Konstruktor je **obyčejná funkce** a wiring napíšeš ručně
v `main`. To zní jako krok zpět, dokud si neuvědomíš, co za to dostaneš: celý graf
závislostí je jeden čitelný blok kódu, který zkontroluje kompilátor.

```go
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	client, err := api.NewClient("https://api.example.com", api.WithTimeout(5*time.Second))
	if err != nil {
		logger.Error("client", "err", err)
		os.Exit(1)
	}

	srv, err := server.New(":8080", logger, client)
	if err != nil {
		logger.Error("server", "err", err)
		os.Exit(1)
	}
	log.Fatal(srv.ListenAndServe())
}
```

Návyk k opuštění: **nehledej framework, který wiring schová.** Pár desítek řádků v `main`
je cena, kterou platíš za to, že nikdy nebudeš ladit runtime chybu kontejneru.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Konstruktor pro typ, kde stačí zero value | „objekt bez `new` neexistuje" | líná inicializace v metodě |
| `New() ServiceInterface` | Symfony vrací abstrakci | vracej `*Service`, interface definuj u konzumenta |
| `NewUser()` v balíčku `user` | jméno má stát samo | `user.New()` |
| Options aplikované po validaci | validace „patří nahoru" | defaults → options → validace |
| Panika v konstruktoru místo `error` | zvyk na výjimky | `(T, error)`; panika jen ve `Must` |
| Interface v balíčku implementace | PHP drží interface u třídy | interface u toho, kdo ho volá |
| Volitelná závislost jako parametr `nil` | „ať to je v signatuře" | `Option`, nebo `Config` struct |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 20`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `NewServer`, `MustNewServer`, `Addr`, `Logger`, `WithTimeout`, `WithRetries`

```bash
make lesson L=20 PART=1
```

Pak **`/go-deep-review 20 easy`**.

### Střední

Funkce: `WithUserAgent`, `NewClient`, `BaseURL`, `Timeout`, `Retries`, `UserAgent`, `NewService`

```bash
make lesson L=20 PART=2
```

Pak **`/go-deep-review 20 medium`**.

### Obtížný

Funkce: `Add`, `Count`, `Values`, `Set`, `Lookup`, `Len`, `Keys`

```bash
make lesson L=20 PART=3
```

Pak **`/go-deep-review 20 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 20 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=20` (+ `make race L=20`, pokud to lekce vyžaduje).

- [ ] Žádná tvoje funkce nevrací interface
- [ ] `Registry` nemá konstruktor a testy na zero value prochází
- [ ] Umíš vysvětlit, proč se validace dělá až po aplikaci options
- [ ] Umíš vysvětlit, proč je `func New() Servicer` s neexportovaným typem anti-vzor
- [ ] Umíš říct, kdy jsou functional options overkill a co použít místo nich
- [ ] Umíš obhájit ruční wiring v `main` proti argumentu „ale Symfony to umí samo"

## AI režim

`JEN VYSVĚTLENÍ` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Zeptat se smíš: *„Jaké jsou nevýhody functional options oproti Config structu?"*
Nesmíš: *„Napiš mi options pattern pro tenhle typ."* Options jsou přesně ten kód, který
AI umí odsypat a ty se ho nikdy nenaučíš.

## Další čtení

1. [Go blog — Self-referential functions and the design of options](https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html) (Rob Pike)
2. [Dave Cheney — Functional options for friendly APIs](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)
3. [Go Code Review Comments — Interfaces](https://go.dev/wiki/CodeReviewComments#interfaces)
4. [Effective Go — Allocation with new and make](https://go.dev/doc/effective_go#allocation_new)
