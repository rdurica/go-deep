# Lekce 51 — Moduly, verzování a zranitelnosti

> **Čas:** ~90 min · **Fáze:** 6 — Production Go · **AI režim:** `JUNIOR POD REVIEW`

## Co budeš umět

- Přečíst `go.mod` řádek po řádku a vysvětlit, co dělá `require`, `replace`, `exclude`, `retract` i `toolchain`.
- Rozhodnout mezi `go get`, `go install` a `go mod tidy` podle toho, co skutečně chceš změnit.
- Vysvětlit, proč Go nemá dependency solver jako Composer, a odvodit výsledek minimal version selection na papíře.
- Rozpoznat pseudo-verzi, přečíst z ní čas commitu a revizi, a vysvětlit, proč má `v2+` sufix v cestě.
- Použít `govulncheck` a vysvětlit, čím se liší od `composer audit`.

## Teorie

### `go.mod` a jeho direktivy

```go.mod
module example.com/shop/v2      // cesta modulu VČETNĚ major sufixu

go 1.26                         // minimální verze jazyka a chování toolchainu
toolchain go1.26.5              // konkrétní toolchain, který se má stáhnout

require (
	example.com/pay v1.4.0
	example.com/log v0.3.1 // indirect
)

tool (
	golang.org/x/tools/cmd/stringer  // od Go 1.24: nástroje vázané na modul (go get -tool …)
)

replace example.com/pay => ../pay          // lokální vývoj, do mainu nepatří
exclude example.com/log v0.3.2             // tahle verze je rozbitá, přeskoč ji
retract [v1.0.0, v1.0.3]                   // MOJE verze, které nikdo nemá používat
```

- `require` je **minimální** požadavek, ne rozsah.
- `tool` (od Go 1.24) eviduje CLI nástroje u modulu — `go tool stringer` je spustí ve verzi z `go.mod`, bez globálního `go install`.
- `// indirect` znamená, že tvůj kód balíček neimportuje přímo; přidal ho `go mod tidy`,
  protože ho potřebuje některá závislost, jejíž vlastní `go.mod` ho neuvádí.
- `replace` platí **jen v hlavním modulu** — když tvůj modul někdo naimportuje, jeho
  `replace` se ignoruje. Proto je bezpečný pro lokální vývoj a nebezpečný v release.
- `exclude` vynechá konkrétní verzi z výběru; algoritmus pak zvolí nejbližší vyšší.
- `retract` je jediná direktiva, která mluví o **tvém vlastním** modulu. Vydáš-li rozbitou
  `v1.0.2`, nemůžeš ji z proxy smazat (je nesmazatelná), ale můžeš vydat `v1.0.3`
  s `retract v1.0.2`. `go get` ji pak přeskočí a `go list -m -versions` ji nenabídne.

### Tři příkazy, které se pletou

| Příkaz | Co dělá | Mění `go.mod`? |
|--------|---------|----------------|
| `go get example.com/pay@v1.5.0` | změní požadovanou verzi závislosti | ano |
| `go install example.com/tool@latest` | zkompiluje a nainstaluje **binárku** do `$GOBIN` | ne |
| `go mod tidy` | dopočítá `require` podle skutečných importů, doplní `go.sum` | ano |

`go install pkg@version` je nejbližší ekvivalent globálního `composer global require`.
Klíčové je, že běží **mimo tvůj modul** — nainstaluje nástroj a tvého `go.mod` se nedotkne.

`go mod tidy` je jediný příkaz, který umí i **ubrat**. Když smažeš import, `require` zůstane,
dokud `tidy` nespustíš. Proto patří do CI jako kontrola: `go mod tidy && git diff --exit-code`.

### Pseudo-verze

Když modul nemá tag, Go si verzi vyrobí samo:

```
v0.0.0-20230101120000-abcdef123456
│      │              └─ prvních 12 znaků hashe commitu
│      └─ čas commitu v UTC, formát yyyymmddhhmmss
└─ základ: nejbližší nižší tag (nebo v0.0.0, když žádný není)
```

Existují tři tvary a liší se jen tím, co je před razítkem:

| Tvar | Kdy vznikne |
|------|-------------|
| `v0.0.0-20230101120000-abcdef123456` | v historii commitu není žádný tag |
| `v1.2.4-0.20230101120000-abcdef123456` | poslední tag je `v1.2.3`, tohle je commit za ním |
| `v1.2.4-rc.1.0.20230101120000-abcdef123456` | poslední tag je `v1.2.4-rc.1` |

Trik je v tom, že pseudo-verze je platná semver **prerelease** verze, takže se korektně
řadí: `v1.2.3 < v1.2.4-0.2023… < v1.2.4`. Bez toho by minimal version selection nefungoval.

### Major verze v cestě (import compatibility rule)

Pravidlo zní: *pokud staré a nové balíčky mají stejnou import path, musí být kompatibilní.*
Go proto od `v2` výš vyžaduje sufix v cestě modulu:

```go
import "example.com/pay"      // v0 a v1
import "example.com/pay/v2"   // v2.x.y
```

`v0` a `v1` sufix **nemají** — `example.com/pay/v1` je chyba. `v0` navíc nemá záruku
zpětné kompatibility vůbec, což je způsob, jak říct „ještě to není hotové".

Praktický důsledek, který v Composeru nemá obdobu: `v1` a `v2` téhož balíčku mohou být
v jednom binárce **současně**, protože jde o dvě různé import paths. Migrace tedy nemusí
být big bang.

### Minimal version selection

Composer hledá nejvyšší verzi splňující všechny rozsahy. Go dělá pravý opak a je to
jednodušší, než to zní: **pro každý modul vezmi nejvyšší ze všech požadovaných minim.**

```
hlavní modul  → pay v1.2.0
example.com/a → pay v1.4.0
example.com/b → pay v1.3.0
────────────────────────────
výsledek      → pay v1.4.0
```

Žádný backtracking, žádné selhání. Algoritmus je deterministický a jeho výsledek se
nezmění, dokud někdo needituje `go.mod`. To je celé — proto Go nepotřebuje lock soubor
pro reprodukovatelnost verzí.

`go mod graph` vypíše celý graf hran „kdo koho požaduje", `go mod why -m example.com/log`
odpoví na otázku „proč to tu vůbec je" a ukáže nejkratší cestu z tvého kódu.

### `go.sum`, checksum databáze a privátní moduly

`go.sum` **není** lock soubor. Je to seznam očekávaných hashů: pro každý modul hash
obsahu a hash jeho `go.mod`. Verze určuje `go.mod`, integritu `go.sum`.

Navíc existuje veřejná, append-only checksum databáze (`sum.golang.org`). Při prvním
stažení modulu se hash ověří proti ní, takže nikdo nemůže přepsat obsah už vydaného tagu.
U firemních repozitářů to nechceš (a nefungovalo by to):

```bash
export GOPRIVATE=git.firma.cz/*        # nastaví GONOPROXY i GONOSUMDB
export GOFLAGS=-mod=readonly           # v CI: build nesmí sám měnit go.mod
```

Vendoring (`go mod vendor`) zkopíruje závislosti do `vendor/`. Dává smysl při auditu
dodavatelského řetězce nebo buildu bez sítě; jinak jen zvětšuje diffy.

### `govulncheck` versus `composer audit`

`composer audit` porovná verze v `composer.lock` s databází a nahlásí každou shodu.
Výsledek je typicky dlouhý seznam, z něhož tě reálně ohrožuje zlomek.

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

`govulncheck` udělá totéž, ale pak přidá krok navíc: postaví **call graph** a nechá
v reportu jen zranitelnosti, jejichž zranitelný symbol tvůj kód skutečně volá. Rozliší
tedy „máš tu verzi" od „ta díra je pro tebe dosažitelná". Nález, který projde tímhle
filtrem, znamená práci teď hned; ostatní jsou položka do plánovaného upgradu.

## Rozdíly proti PHP

V Composeru napíšeš rozsah a solver dopočítá, co se nainstaluje:

```json
{ "require": { "symfony/http-kernel": "^6.3", "psr/log": "^3.0" } }
```

Výsledek závisí na tom, **kdy** jsi spustil `composer update`. Proto existuje `composer.lock`:
bez něj by dva vývojáři dostali jiné verze. Solver navíc řeší backtracking — může selhat
hláškou o nekompatibilních rozsazích a hodinu ti hledat, kde je konflikt.

V Go napíšeš přesnou minimální verzi:

```go.mod
module example.com/shop

go 1.26

require (
	example.com/pay v1.4.0
	example.com/log v0.3.1
)
```

Žádné `^`, žádné `~`, žádný solver. `v1.4.0` neznamená „aspoň 1.4.0, klidně 1.9",
ale „tenhle build použije přesně 1.4.0, pokud někdo jiný nepožaduje víc".

Co se mění v uvažování: **build je deterministický z definice, ne z lock souboru.**
Ze stejného `go.mod` dostaneš dnes i za rok stejné verze. Upgrade není vedlejší efekt
instalace, ale samostatný, viditelný commit, který mění `go.mod`. Přestaň čekat, že
`go get` bez argumentu „aktualizuje závislosti" — takový příkaz v Go prostě neexistuje.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Očekávání, že `go get` upgraduje všechno | reflex z `composer update` | `go get -u ./...` explicitně, jako samostatný commit |
| `replace` v release verzi | pomohlo lokálně, zapomnělo se | `replace` jen dočasně, v CI kontroluj `go mod tidy && git diff --exit-code` |
| Čtení `go.sum` jako lock souboru | analogie s `composer.lock` | verze určuje `go.mod`, `go.sum` jsou jen hashe |
| Publikace v2 bez `/v2` v cestě | v Composeru se cesta nemění | uprav `module` řádek na `example.com/m/v2` |
| Ruční mazání „nepotřebných" `// indirect` | vypadá to jako smetí | vždy jen `go mod tidy` |
| Ignorování `govulncheck`, protože „to hlásí i Composer" | zvyk na šum z auditu | tady je nález dosažitelný voláním, ne jen shoda verzí |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 51`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `String`, `ParseSemver`

```bash
make lesson L=51 PART=1
```

Pak **`/go-deep-review 51 easy`**.

### Střední

Funkce: `Compare`, `ParsePseudoVersion`, `IsPseudo`

```bash
make lesson L=51 PART=2
```

Pak **`/go-deep-review 51 medium`**.

### Obtížný

Funkce: `MajorSuffix`, `SelectVersions`, `CheckCompat`

```bash
make lesson L=51 PART=3
```

Pak **`/go-deep-review 51 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 51 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=51` (+ `make race L=51`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč Go nepotřebuje lock soubor pro determinismus verzí
- [ ] Umíš na papíře spočítat výsledek minimal version selection pro tři moduly
- [ ] Umíš vysvětlit, proč má `v2` sufix v import path a `v1` ne
- [ ] Umíš říct, čím se `govulncheck` liší od `composer audit`
- [ ] Umíš vysvětlit rozdíl mezi `exclude` a `retract`

## AI režim

`JUNIOR POD REVIEW` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Agent ti semver parser napíše za pět vteřin a bude v něm chyba v řazení prerelease.
Napiš si nejdřív tabulku očekávaného pořadí, pak nech agenta implementovat, pak porovnej.

## Další čtení

1. [Go Modules Reference](https://go.dev/ref/mod) — normativní popis včetně MVS a pseudo-verzí
2. [Go blog — Go Modules: v2 and Beyond](https://go.dev/blog/v2-go-modules)
3. [Go blog — Govulncheck v1.0.0](https://go.dev/blog/govulncheck)
4. [Semantic Versioning 2.0.0](https://semver.org/)
