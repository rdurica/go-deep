# Lekce 28 — Konfigurace z prostředí

> **Čas:** ~90 min · **Fáze:** 3 — net/http a tooling · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Vysvětlit, proč se konfigurace v Go čte z prostředí a ne z YAML souboru, a co na tom
  mění kontejnery.
- Rozhodnout mezi `os.Getenv` a `os.LookupEnv` podle toho, jestli musíš odlišit prázdnou
  hodnotu od nenastavené.
- Navrhnout `Load`, která spadne při startu se **seznamem všech** chyb, ne s tou první.
- Napsat konfigurační typ, který se nedá omylem vypsat do logu i s heslem.
- Testovat načítání konfigurace bez globálního stavu procesu.

## PHP → Go most

V Symfony máš na konfiguraci tři vrstvy: `.env` soubory, `config/services.yaml`
s `parameters` a v kódu injektované `%env(int:PORT)%`. Framework je poskládá,
zvaliduje a doručí do konstruktoru.

```php
# config/services.yaml
parameters:
    app.port: '%env(int:PORT)%'
    app.read_timeout: '%env(int:READ_TIMEOUT)%'

# src/Http/Server.php
public function __construct(
    private int $port,
    private int $readTimeout,
) {}
```

V Go žádná taková vrstva není. Konfiguraci si přečteš, převedeš a zvaliduješ **ty**,
obvykle jednou funkcí na začátku `main`:

```go
type Config struct {
	Port        int
	ReadTimeout time.Duration
}

cfg, err := Load(os.Getenv)
if err != nil {
	log.Fatalf("config: %v", err)
}
srv := NewServer(cfg)
```

Změna v uvažování: v Symfony je konfigurace **infrastruktura frameworku**, v Go je to
**obyčejný kód**, který se dá číst, testovat a krokovat. Zmizí ti magie `%env(int:...)%`,
ale taky zmizí situace, kdy aplikace nastartuje s poloviční konfigurací a rozbije se až
u prvního requestu. Druhý přenos návyku: `.env` soubor v Go nečteme. Proměnné do procesu
dostane Docker, systemd, Kubernetes nebo `direnv` — aplikace jen čte prostředí.

## Teorie

### Proč prostředí a ne konfigurační soubor

[12-factor app](https://12factor.net/config) definuje konfiguraci jako *všechno, co se
liší mezi nasazeními*: adresa databáze, klíče k API, feature flagy. Argument pro
proměnné prostředí je praktický:

- Soubor musíš dostat do image nebo do volume, proměnnou předáš přes `-e` nebo Secret.
- Soubor svádí k tomu, aby se commitnul — a s ním i heslo.
- Proměnné umí každý orchestrátor a CI, formát je vždycky stejný: `string`.

To poslední je zároveň jediná nevýhoda: prostředí neumí typy ani struktury. Všechno je
řetězec, takže parsování a validaci musíš udělat sám. Právě proto je zbytek téhle lekce
o převodech a chybách.

### `Getenv` vs `LookupEnv`

```go
port := os.Getenv("PORT")           // "" když není nastavená
port, ok := os.LookupEnv("PORT")    // ok == false když není nastavená
```

Rozdíl je jen v tom, jestli umíš odlišit `PORT=` (nastavená, prázdná) od nenastavené.
V drtivé většině případů to nepotřebuješ — prázdná hodnota znamená „použij default“.
Kdy to potřebuješ:

```go
// FEATURE_X="" má znamenat "vypnuto", ne "vezmi default true"
raw, ok := os.LookupEnv("FEATURE_X")
if !ok {
	raw = "true"
}
enabled, err := strconv.ParseBool(raw)
```

V tomhle cvičení používáme jednodušší pravidlo *prázdná = nenastavená*, protože v praxi
je nechtěně prázdná proměnná (`PORT=$UNDEFINED`) mnohem častější než záměrně prázdná.

### Parsování a validace typů

Standardní knihovna má na všechno funkci a všechny vracejí chybu:

```go
n, err := strconv.Atoi("8080")          // int
b, err := strconv.ParseBool("true")     // bool: 1, t, T, TRUE, true, 0, f, false…
f, err := strconv.ParseFloat("1.5", 64) // float64
d, err := time.ParseDuration("250ms")   // time.Duration: "1h30m", "5s", "250ms"
u, err := url.Parse("postgres://…")     // *url.URL
```

`time.ParseDuration` je důvod, proč se v Go timeouty nekonfigurují jako počet sekund.
`READ_TIMEOUT=1500ms` je čitelnější a nedá se splést jednotka.

Validace není totéž co parsování. `PORT=70000` se přes `Atoi` v pohodě převede — a pak
ti `net.Listen` spadne až za běhu. Rozsah musíš zkontrolovat sám:

```go
if port < 1 || port > 65535 {
	return fmt.Errorf("PORT=%d mimo rozsah 1-65535: %w", port, ErrInvalid)
}
```

### Fail-fast a sbírání chyb

Aplikace, která nastartuje se špatnou konfigurací, je horší než aplikace, která
nenastartuje. Kontejner se v prvním případě tváří zdravě a rozbije se pod provozem;
ve druhém se restart-loop objeví hned a nasazení se zastaví.

Naivní `Load` vrátí první chybu:

```go
// ŠPATNĚ — po opravě DATABASE_URL zjistíš, že chybí i API_KEY, a tak dokola
if cfg.DatabaseURL == "" {
	return Config{}, errors.New("DATABASE_URL is required")
}
if cfg.APIKey == "" {
	return Config{}, errors.New("API_KEY is required")
}
```

Od Go 1.20 na to máme `errors.Join`, který spojí několik chyb do jedné. `errors.Is`
i `errors.As` fungují napříč všemi spojenými chybami a `Error()` je vypíše po řádcích:

```go
var errs []error
if cfg.DatabaseURL == "" {
	errs = append(errs, fmt.Errorf("DATABASE_URL: %w", ErrMissing))
}
if cfg.APIKey == "" {
	errs = append(errs, fmt.Errorf("API_KEY: %w", ErrMissing))
}
if len(errs) > 0 {
	return Config{}, errors.Join(errs...)
}
```

`errors.Join` vrací `nil`, pokud jsou všechny argumenty `nil`, takže test na `len(errs)`
je jen kosmetika. Zpráva musí obsahovat **jméno klíče** — operátor, který ji uvidí
v logu kontejneru, nemá jak zjistit, kterou proměnnou nastavit.

### Konfigurace jako závislost, ne globál

Reflex ze Symfony je udělat `container.getParameter()` kdekoli. Reflex z jiných jazyků
je globální singleton. Obojí je v Go zbytečné: `Config` je obyčejný struct, který
předáš konstruktorem tam, kam patří.

```go
func main() {
	cfg, err := Load(os.Getenv)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	srv := NewServer(cfg.Addr, cfg.Port, cfg.ReadTimeout)
	// …
}
```

Stejná logika platí pro čtení prostředí uvnitř `Load`. Když funkce zavolá `os.Getenv`
přímo, testuješ ji jen přes `t.Setenv`, tedy přes globální stav procesu — nemůžeš pustit
testy paralelně a snadno si nastavíš proměnnou, kterou pak zapomeneš uklidit. Když si
`getenv func(string) string` vezme jako parametr, je to čistá funkce:

```go
cfg, err := Load(os.Getenv)                                   // produkce
cfg, err := Load(func(k string) string { return env[k] })     // test
```

Je to stejný trik jako „accept interfaces“ — jen s funkcí místo interface, protože
závislost má jedinou metodu.

### `flag` a kombinace flag + env

Pro CLI nástroje je přirozenější `flag`:

```go
fs := flag.NewFlagSet("api", flag.ContinueOnError)
addr := fs.String("addr", lookup("ADDR", "0.0.0.0"), "adresa pro poslech")
if err := fs.Parse(os.Args[1:]); err != nil {
	return err
}
```

Obvyklá priorita je **flag > env > default**. Dosáhneš jí tím, že výchozí hodnotu flagu
načteš z prostředí, jak je vidět výše. `flag.NewFlagSet` (místo globálního `flag.String`)
použij vždycky, když chceš věc testovat — globální `flag.CommandLine` v testech koliduje
s flagy balíčku `testing`.

### Tajemství v konfiguraci

Konfigurace skoro vždycky obsahuje heslo nebo token. Stačí jeden `log.Printf("%+v", cfg)`
při ladění a máš tajemství v centralizovaném logu, ze kterého se špatně maže.

Obrana je metoda `String()`. `fmt` ji použije pro `%v`, `%+v` i `%s`, takže se do výstupu
dostane jen to, co dovolíš:

```go
func (c Config) String() string {
	return fmt.Sprintf("Config{Addr:%s Port:%d APIKey:%s}", c.Addr, c.Port, "***")
}
```

Dvě upozornění. Za prvé: `%#v` a `json.Marshal` `String()` **ignorují**, takže disciplína
pořád platí. Za druhé: metoda musí být na hodnotovém receiveru (`func (c Config)`), jinak
by se neuplatnila při výpisu hodnoty, jen ukazatele.

Na hesla v URL má stdlib hotovou funkci:

```go
u, _ := url.Parse("postgres://app:s3cr3t@db:5432/app")
fmt.Println(u.Redacted()) // postgres://app:xxxxx@db:5432/app
```

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Čtení `.env` souboru v aplikaci | reflex ze Symfony Dotenv | proměnné dodá orchestrátor, aplikace čte jen prostředí |
| `Load` vrátí první chybu | zvyk na výjimky, které přeruší běh | posbírej chyby a spoj přes `errors.Join` |
| Validace až při použití hodnoty | konfigurace se bere jako „jen data“ | validuj při startu, fail-fast |
| `os.Getenv` uvnitř business kódu | zvyk na globální `getenv()` / DI kontejner | předej `Config` konstruktorem |
| `log.Printf("%+v", cfg)` | ladicí výpis se nikdy nesmaže | `String()` maskující tajemství |
| `PORT` jako `string` do `net.Listen` | prostředí je stringové, tak proč ne | převeď a zvaliduj hned v `Load` |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

### A — rozcvička (~10 min)

1. `LookupString(getenv func(string) string, key, def string) string` — vrátí hodnotu
   klíče, nebo `def`, pokud je hodnota prázdná.
2. `LookupInt(getenv func(string) string, key string, def int) (int, error)` — totéž pro
   `int`. Prázdná hodnota dá `def` a `nil`. Nečíselná hodnota vrátí `0` a chybu, která
   obaluje `ErrInvalid` a v textu obsahuje jméno klíče.

Obě funkce berou `getenv` jako parametr místo toho, aby volaly `os.Getenv`. Test díky
tomu běží paralelně a nedotýká se prostředí procesu.

### B — jádro (~35 min)

`Load(getenv func(string) string) (Config, error)` sestaví `Config` z těchto klíčů:

| Klíč | Pole | Výchozí | Pravidlo |
|------|------|---------|----------|
| `ADDR` | `Addr` | `0.0.0.0` | — |
| `PORT` | `Port` | `8080` | musí být 1–65535 |
| `READ_TIMEOUT` | `ReadTimeout` | `5s` | `time.ParseDuration`, musí být kladný |
| `DEBUG` | `Debug` | `false` | `strconv.ParseBool` |
| `DATABASE_URL` | `DatabaseURL` | — | povinné |
| `API_KEY` | `APIKey` | — | povinné |

Klíčová vlastnost: `Load` **nesmí skončit u první chyby**. Posbírej všechny problémy
a vrať je spojené přes `errors.Join`, aby `errors.Is(err, ErrMissing)` i
`errors.Is(err, ErrInvalid)` fungovalo a text obsahoval jména všech vadných klíčů.
Při chybě vrať nulový `Config`.

### C — rozšíření (~20 min)

1. `func (c Config) String() string` — vypíše konfiguraci tak, aby se do výstupu nedostal
   `APIKey` ani heslo z `DatabaseURL`. Test to kontroluje pro `String()`, `%v`, `%+v`
   i `%s`. Netajné hodnoty (port, host databáze) naopak ve výstupu zůstat musí. URL bez
   hesla se nemění.
2. `LoadFromEnviron(environ []string) (Config, error)` — dostane slice ve formátu
   `os.Environ()`, tedy `"KEY=VALUE"`. Rozděl na **prvním** `=` (hodnota může další `=`
   obsahovat), přeskoč položky bez `=` a s prázdným klíčem, a výsledek předej `Load`.

```bash
make lesson L=28
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `28`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=28` prochází
- [ ] Umíš vysvětlit, kdy potřebuješ `os.LookupEnv` místo `os.Getenv`
- [ ] Umíš vysvětlit, proč je fail-fast při startu lepší než chyba za běhu
- [ ] Umíš vysvětlit, co dělá `errors.Join` s `errors.Is`
- [ ] Umíš vysvětlit, proč `Load` bere `getenv` jako parametr
- [ ] Víš, které formátovací slovesa `String()` ignorují

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).
Mapování klíčů na pole si nech vygenerovat, ale error model, výchozí hodnoty
a maskování tajemství navrhni sám.

## Další čtení

1. [pkg.go.dev — os.LookupEnv](https://pkg.go.dev/os#LookupEnv)
2. [pkg.go.dev — errors.Join](https://pkg.go.dev/errors#Join)
3. [pkg.go.dev — flag](https://pkg.go.dev/flag)
4. [The Twelve-Factor App — Config](https://12factor.net/config)
