# Lekce 55 — Checkpoint fáze 6: kontejnery, health a production checklist

> **Čas:** ~90 min · **Fáze:** 6 — Production Go · **AI režim:** `JUNIOR POD REVIEW`

Checkpoint neopakuje teorii fáze 6 (lekce 51–54), jen ji shrne. Přidává poslední chybějící
kus provozu — kontejner, sondy a ukončování — a nechá tě z toho postavit jeden balíček,
který kombinuje čtyři témata fáze: build metadata, HTTP handlery, souběžnost s timeoutem
a `errors.Join`.

## Co budeš umět

- Napsat multi-stage Dockerfile pro Go a obhájit každý jeho řádek.
- Vysvětlit rozdíl mezi liveness, readiness a startup probe a co která z nich udělá v Kubernetes.
- Vložit verzi do binárky přes `-ldflags` a vystavit ji na `/version`.
- Postavit readiness endpoint, který se nezasekne na kontrole ignorující `context`.
- Ukončit službu v definovaném pořadí a v definovaném čase.

## Recap

### Otázky a odpovědi

**Proč Go nepotřebuje `composer.lock` pro reprodukovatelný build?** Protože `require`
v `go.mod` je konkrétní minimální verze a minimal version selection z ní deterministicky
odvodí výsledek. `go.sum` řeší integritu, ne výběr. (lekce 51)

**Čím se `govulncheck` liší od `composer audit`?** Postaví call graph a nahlásí jen ty
zranitelnosti, jejichž zranitelný symbol tvůj kód skutečně volá. Míň šumu, víc signálu.
(lekce 51)

**Proč má `v2` sufix v import path?** Kvůli import compatibility rule: stejná cesta musí
znamenat kompatibilní API. Vedlejší efekt je, že `v1` a `v2` můžou koexistovat v jednom
binárce. (lekce 51)

**Proč benchmark potřebuje package-level sink?** Bez použití výsledku smí kompilátor
volání odstranit a ty měříš prázdný cyklus. (lekce 52)

**Co je dobrý fuzz invariant?** Round-trip (`Decode(Encode(x)) == x`), „nepanikuje na
žádném vstupu" a idempotence. Špatný invariant je porovnání s konkrétní očekávanou
hodnotou. (lekce 52)

**Kdy má smysl golden file?** Když je výstup velký, strukturovaný a čitelný, takže jeho
diff dává v review smysl. Na tři pole napiš assert. (lekce 52)

**Kdy začít profilovat?** Až máš měřitelný problém a benchmark, který ho reprodukuje.
Optimalizace bez profilu je hádání. (lekce 53)

**Jaký je rozdíl mezi flat a cum?** Flat je čas v těle funkce, cum včetně volaných.
Vysoký cum a nízký flat znamená průchoďák — hledej hlouběji. (lekce 53)

**Proč `import _ "net/http/pprof"` může být díra?** Registruje endpointy na
`DefaultServeMux`. Když na něm běží veřejný server, vystavil jsi zásobníky goroutin
i možnost spustit profilování komukoli. (lekce 53)

**Kdy generika a kdy rozhraní?** Rozhraní, když potřebuješ chování. Typový parametr, když
potřebuješ znát konkrétní typ (vracíš ho, ukládáš do slice, porovnáváš). (lekce 54)

**Proč nejde `func (r Result[T]) Map[U any](...)`?** Metody nesmí mít vlastní typové
parametry, protože sada metod typu musí být konečná a známá při kompilaci. (lekce 54)

**Kam patří reflexe?** Do knihoven, které z principu neznají typ volajícího — serializace,
ORM, DI. V aplikačním kódu je ruční mapování rychlejší i čitelnější. (lekce 54)

### Co si musíš pamatovat

| Téma | Pravidlo | Nejčastější past |
|------|----------|------------------|
| Moduly | `require` je minimum, ne rozsah | čekání, že `go get` upgraduje vše |
| MVS | nejvyšší z požadovaných minim | hledání solveru, který tam není |
| `replace` | platí jen v hlavním modulu | zapomenutý `replace` v release |
| Benchmark | výsledek do package-level sinku | měření mrtvého kódu |
| Alokace | `AllocsPerRun` je deterministický | test na přesné číslo bez rezervy |
| Fuzz | invariant, ne očekávaná hodnota | fuzz, který jen kontroluje `err == nil` |
| Golden | `-update` a přečtený diff | zafixovaný rozbitý výstup |
| pprof | profil až po benchmarku | optimalizace podle intuice |
| Reflexe | knihovny ano, aplikace ne | vlastní mapper v service vrstvě |
| Build tagy | prázdný řádek před `package` | varianta bez testu tiše hnije |

## Rozdíly proti PHP

PHP aplikace v kontejneru je vždycky nejmíň dva procesy: php-fpm a nginx. Deploy znamená
image s celým interpretem, rozšířeními a `vendor/`:

```dockerfile
FROM php:8.3-fpm-alpine
COPY --from=composer /usr/bin/composer /usr/bin/composer
COPY . /app
RUN composer install --no-dev --optimize-autoloader
# výsledek: ~150 MB, dva procesy, supervisord navrch
```

Go služba je jeden statický soubor:

```dockerfile
FROM scratch
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
# výsledek: ~10 MB, jeden proces, PID 1 je tvůj kód
```

Co se mění v uvažování: **tvůj proces je PID 1 a nikdo za tebe neuklízí.** V php-fpm
worker po requestu umře a paměť se uvolní; graceful shutdown řeší master proces. V Go
běží tvůj proces měsíce, sám dostane `SIGTERM` a sám se musí korektně ukončit. Co
nedodeleguješ, to se neudělá.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Kontrola databáze v liveness | „health check je health check" | závislosti do readiness, liveness jen žije/nežije |
| `ENTRYPOINT /app` bez závorek | zvyk psát shell příkazy | exec forma `["/app"]`, jinak nedostaneš SIGTERM |
| Zavření serveru dřív, než pod zmizí z endpointů | přeskočená pauza na drain | nejdřív ready=false, pak počkat, pak Shutdown |
| Konfigurační soubor v image | reflex z `config/packages/prod` | všechno z env, image je jeden pro všechna prostředí |
| Logy do `/var/log/app.log` | zvyk na Monolog s rotací | stdout, sběr řeší platforma |
| `GOMAXPROCS` ponechaný na výchozí | v PHP se o CPU nestaráš | odvoď ho z CPU limitu podu |

## Nová látka: kontejner a sondy

### Statická binárka a multi-stage build

`CGO_ENABLED=0` vypne linkování proti libc. Výsledek běží i v `scratch` a nezávisí na
verzi glibc v base image. Cena je, že `os/user` a výchozí DNS resolver přejdou na čistě
Go implementaci — což je v kontejneru skoro vždy to, co chceš.

Kompletní Dockerfile je v [`exercise/Dockerfile`](exercise/Dockerfile). Řádek po řádku:

```dockerfile
FROM golang:1.26-alpine AS build     # build stage, do finálního image se nedostane
ARG VERSION=dev                       # metadata předaná z CI
ARG COMMIT=none
WORKDIR /src

COPY go.mod go.sum ./                 # nejdřív jen manifest…
RUN go mod download                   # …aby se vrstva s závislostmi cachovala

COPY . .                              # teprve teď zdrojáky (mění se často)

ENV CGO_ENABLED=0 GOOS=linux          # statická binárka
RUN go build -trimpath \              # -trimpath: bez absolutních cest, reprodukovatelný build
    -ldflags "-s -w \                 # -s -w: bez debug symbolů, o třetinu menší
      -X main.version=${VERSION} \    # -X: nastaví hodnotu string proměnné
      -X main.commit=${COMMIT} \
      -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /out/app ./cmd/app

FROM gcr.io/distroless/static-debian12:nonroot   # jen CA certifikáty, tzdata a nonroot uživatel
COPY --from=build /out/app /app                  # jediný soubor z build stage
USER nonroot:nonroot                             # nikdy ne root
EXPOSE 8080
ENTRYPOINT ["/app"]                              # exec forma — viz níže
```

`distroless/static` proti `scratch`: obsahuje CA certifikáty (bez nich ti selže každé
HTTPS volání), `/etc/passwd` s nonroot uživatelem a tzdata. Za pár megabajtů navíc ušetří
tři klasické produkční překvapení.

`-X` funguje **jen na `string` proměnné na úrovni balíčku**, a to na nesmíšené jméno
`import/path.jmenoPromenne`. Na konstantu ani na `int` ho nepoužiješ.

### `ENTRYPOINT` v exec formě a PID 1

```dockerfile
ENTRYPOINT ["/app"]        # exec forma: /app je PID 1 a dostane SIGTERM
ENTRYPOINT /app            # shell forma: PID 1 je /bin/sh, signál dostane on
```

Shell forma spustí `/bin/sh -c "/app"`. Shell signály nepředává, takže tvůj graceful
shutdown se nikdy nespustí a po `terminationGracePeriodSeconds` přijde `SIGKILL`. Je to
nejčastější chyba v Go Dockerfilech a projeví se jen tím, že se pody restartují „nějak
dlouho".

### Tři sondy a jejich chování

| Sonda | Otázka | Co udělá Kubernetes při selhání |
|-------|--------|---------------------------------|
| liveness | *žije proces?* | **restartuje kontejner** |
| readiness | *smí chodit provoz?* | vyřadí pod z Service endpoints, nerestartuje |
| startup | *už nastartoval?* | drží ostatní sondy vypnuté, po vypršení restartuje |

Nejdražší chyba je zapojit závislosti do **liveness**. Když spadne databáze, liveness
začne selhávat, Kubernetes restartuje všechny pody, ty po startu zjistí, že databáze
pořád nejede, a smyčka pokračuje. Liveness proto odpovídá 200, dokud proces žije —
nekontroluje nic. Závislosti patří do **readiness**: pod se jen přestane používat a
vrátí se, až se databáze zotaví.

Startup probe je pro pomalý start (načtení cache, migrace). Bez ní bys musel nastavit
`initialDelaySeconds` tak vysoko, že by liveness dlouho nic nehlídala.

### Graceful shutdown a signály

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
<-ctx.Done()
```

Správné pořadí ukončení má čtyři kroky a jejich záměna je vidět jako 502 v load balanceru:

1. **Přestat být ready** — readiness vrátí 503, Kubernetes vyřadí pod z endpointů.
2. **Počkat drain** — pár vteřin, než se změna endpointů propíše do všech kube-proxy.
   Bez téhle pauzy dostáváš requesty ještě po zavření serveru.
3. **Zavřít HTTP server** — `srv.Shutdown(ctx)` dokončí rozpracované requesty.
4. **Zavřít závislosti** — databáze, fronty, klienti.

Celý rozpočet musí být kratší než `terminationGracePeriodSeconds` (výchozích 30 s), jinak
přijde `SIGKILL` uprostřed.

### Konfigurace, logy a CPU

- **Konfigurace z env**, ne ze souboru v image. Stejný image musí jít nasadit do všech
  prostředí; soubor v image z něj dělá artefakt jednoho prostředí.
- **Logy na stdout** jako JSON přes `slog`. Zápis do souboru v kontejneru znamená rotaci,
  kterou nikdo neřeší, a logy, které zmizí s podem.
- **`GOMAXPROCS` ignoruje CPU limity cgroup.** Runtime vidí všechna jádra hosta, takže na
  16jádrovém uzlu s limitem `500m` běží 16 P a scheduler se zbytečně přepíná a naráží na
  throttling. Řeš to buď explicitním `runtime.GOMAXPROCS(n)` z limitu, nebo knihovnou
  `automaxprocs`. `GOMEMLIMIT` nastav zhruba na 90 % memory limitu.

## Production checklist

Odškrtávací seznam před nasazením Go služby. Používej ho jako šablonu PR description.

**Build a artefakt**

- [ ] `CGO_ENABLED=0`, binárka je statická
- [ ] Multi-stage build, finální image je `distroless` nebo `scratch`
- [ ] Verze, commit a čas buildu vložené přes `-ldflags -X`
- [ ] `-trimpath` a `-s -w` zapnuté
- [ ] Image běží pod nonroot uživatelem
- [ ] `ENTRYPOINT` v exec formě (`["/app"]`)
- [ ] Vrstvy seřazené tak, aby se `go mod download` cachoval

**Závislosti**

- [ ] `go mod tidy` je součástí CI a `git diff --exit-code` prochází
- [ ] Žádný `replace` v hlavní větvi
- [ ] `govulncheck ./...` běží v CI a nálezy jsou vyřešené nebo zdůvodněné
- [ ] `GOFLAGS=-mod=readonly` v CI

**Provoz**

- [ ] Liveness sonda nekontroluje žádné závislosti
- [ ] Readiness sonda kontroluje závislosti a má timeout
- [ ] Startup probe u služeb s pomalým startem
- [ ] Graceful shutdown: ready=false → drain → server → závislosti
- [ ] Rozpočet shutdownu je kratší než `terminationGracePeriodSeconds`
- [ ] Všechna konfigurace z env, žádný config soubor v image
- [ ] Logy jako JSON na stdout, žádný zápis do souboru
- [ ] `GOMAXPROCS` odvozený z CPU limitu podu
- [ ] `GOMEMLIMIT` nastavený na ~90 % memory limitu
- [ ] Resource requests i limits nastavené

**Kód a testy**

- [ ] Každý odchozí HTTP a DB klient má timeout
- [ ] `context.Context` protažený od requestu až k závislostem
- [ ] Testy běží v CI s `-race`
- [ ] Kritická cesta má benchmark a hlídané alokace
- [ ] Parsery vstupu mají fuzz test a commitnutý korpus
- [ ] pprof endpointy jen na interním portu, nikdy na `DefaultServeMux`
- [ ] Metriky nebo strukturované logy pokrývají chybové cesty

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 55`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `String`, `Current`

```bash
make lesson L=55 PART=1
```

Pak **`/go-deep-review 55 easy`**.

### Střední

Funkce: `VersionHandler`, `NewHealthChecker`, `Register`

```bash
make lesson L=55 PART=2
```

Pak **`/go-deep-review 55 medium`**.

### Obtížný

Funkce: `LiveHandler`, `ReadyHandler`, `ShutdownSequence`

```bash
make lesson L=55 PART=3
```

Pak **`/go-deep-review 55 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Sebehodnocení

Za každou položku, kterou zvládneš **bez nahlédnutí do lekce**, si dej 1 bod.

| # | Dovednost | Lekce |
|---|-----------|-------|
| 1 | Přečtu `go.mod` včetně `replace`, `exclude` a `retract` | 51 |
| 2 | Spočítám výsledek minimal version selection pro tři moduly | 51 |
| 3 | Vysvětlím pseudo-verzi a major sufix v import path | 51 |
| 4 | Řeknu, čím se `govulncheck` liší od auditu podle verzí | 51 |
| 5 | Napíšu benchmark, který neměří mrtvý kód | 52 |
| 6 | Přečtu ns/op, B/op a allocs/op a vím, které z nich věřit | 52 |
| 7 | Navrhnu fuzz invariant a založím seed korpus v `testdata/` | 52 |
| 8 | Napíšu golden test s `-update` a vysvětlím jeho riziko | 52 |
| 9 | Pořídím CPU i heap profil a najdu v `top`/`list` hrdlo | 53 |
| 10 | Odstraním alokace v horké cestě a doložím to číslem | 53 |
| 11 | Vystavím pprof bezpečně, mimo `DefaultServeMux` | 53 |
| 12 | Rozhodnu mezi generikou a rozhraním a zdůvodním to | 54 |
| 13 | Napíšu kód nad `reflect` včetně tagů a adresovatelnosti | 54 |
| 14 | Rozdělím implementaci build tagy a otestuju obě varianty | 54 |
| 15 | Napíšu multi-stage Dockerfile a obhájím každý řádek | 55 |
| 16 | Rozliším liveness, readiness a startup probe | 55 |
| 17 | Naimplementuju graceful shutdown ve správném pořadí | 55 |

| Skóre | Co s tím |
|-------|----------|
| 16–17 | Fáze 6 sedí, pokračuj na lekci 56. |
| 13–15 | Zopakuj lekce z řádků, kde jsi bod nedostal, a udělej jejich část C znovu. |
| 9–12 | Zopakuj lekce 52 a 53 — bez měření se produkční Go dělat nedá. |
| 5–8 | Projdi znovu celý blok 51–54, tentokrát s vlastními benchmarky a profily. |
| 0–4 | Vrať se na lekci 51 a jdi fází 6 znovu; fáze 7 předpokládá provozní úsudek. |

## Závěrečné otázky

Spusť **`/go-deep-review 55 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=55` (+ `make race L=55`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč se databáze nekontroluje v liveness sondě
- [ ] Umíš vysvětlit, co udělá shell forma `ENTRYPOINT` se signály
- [ ] Umíš vyjmenovat čtyři kroky graceful shutdownu ve správném pořadí
- [ ] Umíš říct, proč `CGO_ENABLED=0` a co za to platíš
- [ ] Umíš vysvětlit, proč `GOMAXPROCS` v kontejneru zlobí

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Dockerfile a health endpointy jsou přesně ten druh kódu, který ti agent napíše okamžitě a
skoro správně: shell forma ENTRYPOINTu, databáze v liveness sondě, `Shutdown` bez pauzy na
drain. Použij checklist výše jako review kritéria a projdi vygenerovaný kód řádek po řádku.

## Další čtení

1. [Kubernetes — Configure Liveness, Readiness and Startup Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
2. [Google — Distroless container images](https://github.com/GoogleContainerTools/distroless)
3. [pkg.go.dev — net/http.Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)
4. [Go — A Guide to the Go Garbage Collector (GOMEMLIMIT)](https://go.dev/doc/gc-guide#Memory_limit)
