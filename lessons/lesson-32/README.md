# Lekce 32 — Project layout a `internal/`

> **Čas:** ~90 min · **Fáze:** 4 — Architektura v Go · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Rozhodnout, kdy je jeden soubor dost a kdy je čas rozdělit kód do balíčků — a proč se
  začíná malým layoutem.
- Vysvětlit, co přesně vynucuje `internal/`, kdo to vynucuje a proč `pkg/` většinou nechceš.
- Pojmenovat balíčky podle domény místo podle vrstvy a obhájit to před kolegou, který
  chce `service/`, `repository/` a `entity/`.
- Použít zákaz cyklických importů jako nástroj návrhu, ne jako otravnou překážku.
- Napsat tenkou aplikační vrstvu, která jen skládá závislosti, a vědět, co do ní nepatří.

## PHP → Go most

Symfony ti dá layout zdarma a ty ho dodržuješ, protože ho dodržují všichni:

```
src/
  Controller/OrderController.php
  Entity/Product.php
  Repository/ProductRepository.php
  Service/PricingService.php
```

Adresáře jsou **vrstvy**. Doména je rozprostřená napříč nimi: produkt žije ve čtyřech
adresářích a ty musíš skákat, když chceš pochopit jednu věc. Funguje to, protože
autoloader ani kontejner hranice neřeší — všechno vidí na všechno.

Go protějšek vypadá jinak:

```
cmd/api/main.go
internal/catalog/catalog.go   // Product, Item, Catalog, pravidla
internal/pricing/pricing.go   // výpočty nad catalog
internal/postgres/products.go // adaptér
```

Adresáře jsou **domény**, protože adresář je v Go zároveň balíček a balíček je jednotka
zapouzdření. Malé písmeno u identifikátoru = neviditelný mimo balíček. To je silnější
hranice než `private` v PHP, protože platí najednou pro typy, funkce, konstanty i pole.

Návyk k opuštění: **nepřekládej `src/Service` na `service/`**. Balíček `service` neříká
nic o tom, co je uvnitř, a skončí jako smetiště. A hlavně: `service.PricingService` je
koktavé, zatímco `pricing.Total` se čte jako věta.

## Teorie

### Oficiální layout neexistuje

Go tým nikdy nevydal doporučenou strukturu projektu. Repozitář
`golang-standards/project-layout` má desítky tisíc hvězdiček a jméno, které svádí
k dojmu, že je oficiální — ale není, s Go týmem nemá nic společného a členové týmu
opakovaně řekli, že ho nedoporučují. Jeho hlavní hřích je, že začátečníka nutí vytvořit
deset prázdných adresářů (`api/`, `build/`, `deployments/`, `pkg/`, `third_party/`…)
dřív, než napíše první řádek.

Jediné, co Go skutečně definuje, jsou tři věci:

| Cesta | Kdo to vynucuje | Co to znamená |
|-------|-----------------|---------------|
| `cmd/neco/main.go` | konvence | binárka jménem `neco` |
| `internal/...` | **go tool** | importovat smí jen kód pod rodičem `internal/` |
| `testdata/` | go tool | adresář ignorovaný při buildu |

Všechno ostatní je zvyklost. Oficiální doporučení („Organizing a Go module") zní: **začni
jedním balíčkem v kořeni a rozděl ho, až tě k tomu něco donutí.** Typický spouštěč
dělení je, že soubor přesáhne velikost, kterou udržíš v hlavě, nebo že chceš zabránit
tomu, aby jedna část sahala do vnitřností druhé.

### `internal/` je jediné vynucené pravidlo

```
github.com/me/shop/
  internal/catalog/    // importovat smí jen kód pod github.com/me/shop/
  internal/pricing/
  money/               // importovat smí kdokoli na světě
```

Pokus o import `github.com/me/shop/internal/catalog` z cizího modulu skončí chybou
kompilace, ne jen zamračením v code review. Tohle je v Go náhrada za „package private"
v Javě a je to nejlevnější dostupná ochrana API.

Praktický důsledek: **dokud si nejsi jistý, dej všechno do `internal/`.** Co je uvnitř,
můžeš kdykoli přejmenovat, rozdělit nebo zahodit, aniž bys komukoli rozbil build. Co je
venku, jsi slíbil udržovat.

A `pkg/`? Je to jen adresář se jménem `pkg`, který nedělá vůbec nic. Když má tvůj modul
jednu binárku a nikdo tvoje balíčky neimportuje, přidá ti jednu úroveň zanoření a nulovou
hodnotu. Sáhni po něm jen tehdy, když skutečně publikuješ knihovnu vedle aplikace a chceš
tu hranici vidět přímo v cestě importu.

### Balíčky podle domény, ne podle vrstvy

Jméno balíčku je prefix všeho, co z něj vytáhneš. Proto se navrhuje *spolu* s API:

```go
// špatně — vrstva v názvu, koktavé volání
service.NewPricingService(repository.NewProductRepository(db))

// dobře — doména v názvu
pricing.Total(items)
```

Dobré jméno balíčku je jednoslovné podstatné jméno v jednotném čísle: `catalog`,
`pricing`, `order`, `billing`, `postgres`, `httpapi`. Zakázaná jména jsou `util`,
`common`, `helpers`, `models`, `base` — všechna znamenají „nevím, kam to dát", a to je
vždycky signál, že kód patří k něčemu konkrétnímu.

Kam patří typ, který používají dva balíčky? Tam, kde je doma. `Item` (produkt +
množství) je doménový pojem, takže bydlí v `catalog`, i když s ním počítá `pricing`.
Kdybys ho dal do `pricing`, začne `catalog` záviset na cenách — a to je obrácený směr,
než chceš.

### Cyklické importy jako nástroj

Go nedovolí, aby balíček A importoval B a zároveň B importoval A. Žádná výjimka, žádná
dopředná deklarace. Zpočátku to vypadá jako šikana, ale je to nejlevnější dostupná
kontrola architektury: **kompilátor za tebe hlídá, že závislosti mají směr.**

```go
// pricing zná catalog
package pricing

import ".../catalog"

func Total(items []catalog.Item) (int64, error) { /* ... */ }
```

Kdyby `catalog` chtěl zavolat `pricing.Total`, build spadne na `import cycle not
allowed`. To tě donutí zeptat se, co je uvnitř a co venku. Typická odpověď: doména je
uvnitř, výpočty, formátování, HTTP a SQL jsou venku. Když cyklus opravdu potřebuješ
rozbít, máš tři možnosti — přesunout sdílený typ do třetího balíčku, otočit směr přes
interface definovaný u konzumenta (o tom je lekce 33), nebo předat funkci jako parametr.

### Tenký `main` a kam s ostatními soubory

`main` není místo pro logiku. Je to místo, kde se čte konfigurace, otevírá spojení,
skládají závislosti a spouští server:

```go
func main() {
	cfg := config.Load()
	cat, err := catalog.New(loadProducts()...)
	if err != nil {
		log.Fatal(err)
	}
	srv := httpapi.New(cat)
	log.Fatal(http.ListenAndServe(cfg.Addr, srv))
}
```

Tohle je celý „DI kontejner" v Go — pár řádků, které přečteš shora dolů a hned víš, co na
čem visí. Žádné `services.yaml`, žádné autowiring, žádná runtime reflexe. Cenou je, že si
wiring píšeš ručně; výhodou je, že chybějící závislost je chyba kompilace, ne výjimka za
běhu.

Zbytek souborů:

- Testy jsou vedle kódu: `catalog/catalog_test.go`. Balíček `catalog` (bílá skříňka) nebo
  `catalog_test` (černá skříňka, jen veřejné API). Druhá varianta je lepší default.
- Fixtury patří do `testdata/` vedle testu — go tool ten adresář při buildu ignoruje.
- Migrace do `migrations/` v kořeni, číslované; nejsou to Go zdrojáky (lekce 35).
- Konfigurace se čte z prostředí (lekce 28), takže `config/packages/*.yaml` nepotřebuješ.
- Jeden modul (`go.mod` v kořeni) je správná odpověď skoro vždycky. Víc modulů znamená
  víc verzování a `replace` direktivy; sáhni po nich, až budeš části vydávat zvlášť.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `service/`, `repository/`, `entity/` | přímý překlad Symfony `src/` | balíčky podle domény: `catalog`, `pricing` |
| Deset prázdných adresářů na startu | „standard layout" z GitHubu | jeden balíček, dělit až při bolesti |
| `pkg/` u aplikace bez knihovny | zvyk na strukturu z tutoriálu | `internal/`, dokud někdo nepotřebuje opak |
| Balíček `utils` nebo `common` | reflex ze sdílených PHP tříd | přesuň funkci k typu, se kterým pracuje |
| Doména importuje konzumenta | anemické entity + fat services | směr importů má být jednosměrný dovnitř |
| Logika v kompozičním kořeni | žádný kontejner, tak to hodím sem | kořen jen skládá a spouští |

## Úkol

Pracuj v `exercise/`. Balíček není jeden — struktura je součástí zadání:

```
exercise/
  exercise.go            // tenká kompozice, obdoba main()
  catalog/               // doména: Product, Item, Catalog
  pricing/               // výpočty nad catalog
  internal/idgen/        // interní nástroj, ven se nedostane
```

Ve `exercise.go` jsou všechny veřejné funkce jen průchozí volání do podbalíčků —
doplň i je, ale žádné pravidlo tam nepiš. Postupuj A → B → C, po každé části spusť test.

### A — rozcvička (~10 min)

V `catalog/catalog.go`:

1. `Validate(p Product) error` — kontroluj v pořadí SKU, jméno, cenu. Prázdné nebo jen
   z bílých znaků složené SKU → `ErrEmptySKU`, totéž pro jméno → `ErrEmptyName`, záporná
   cena → `ErrNegativeCents`. Nula je platná cena. Chybu obal tak, aby její text
   obsahoval SKU a `errors.Is` dál fungovalo.
2. `New(products ...Product) (*Catalog, error)` — každý produkt projde `Validate`,
   duplicitní SKU je `ErrDuplicateSKU`. Prázdný vstup dává prázdný katalog, ne chybu.
3. `(*Catalog).BySKU(sku string) (Product, error)` — neznámé SKU vrací `ErrNotFound`.
   Musí přežít i volání nad `nil` katalogem (metoda na nil pointeru je legální).
4. `(*Catalog).All() []Product` — všechny produkty **seřazené vzestupně podle SKU**.
   Iterace mapy má náhodné pořadí, takže řazení je součást kontraktu.

např. `Validate({SKU:"A-1", Name:"Kniha", Cents:1999})` → `nil`

### B — jádro (~35 min)

V `pricing/pricing.go` implementuj `Total(items []catalog.Item) (int64, error)`:

- Prázdný nebo `nil` košík → `0, nil`.
- Každý produkt musí projít `catalog.Validate`; chybu propaguj beze změny sentinelu.
- Množství musí být kladné, jinak `ErrInvalidQty`.
- Cena řádku je `Cents * Qty`. Když násobení nebo součet přeteče `int64`, vrať
  `ErrOverflow` — nikdy tiše přetečenou hodnotu. (Test to zkouší s `math.MaxInt64`.)

Ve `exercise.go` doplň `TotalOf` a `PriceOf(c *Catalog, sku string, qty int)`, která
najde produkt v katalogu a nechá cenu spočítat `pricing`.

Všimni si, že `pricing` importuje `catalog`, ale ne naopak. Zkus si do `catalog` přidat
import `pricing` — build spadne na `import cycle not allowed`. Pak to vrať zpátky.

např. `TotalOf([{Kniha, 1999, Qty:3}])` → `5997, nil`

### C — rozšíření (~20 min)

V `internal/idgen/idgen.go` implementuj generátor:

- `New(prefix string) *Gen` — prázdný prefix se nahradí `"id"`.
- `(*Gen).NewID() string` — vrací `"<prefix>-000001"`, `"<prefix>-000002"`, … tedy
  pořadové číslo na šest míst doplněné nulami.
- Musí být **bezpečný pro souběžné volání** — test ho volá z osmi goroutin a hlídá, že
  se žádné ID neopakuje. Test běží i s `-race`.

Ve `exercise.go` doplň `NewIDGen`. Nakonec si vyzkoušej, co `internal/` znamená: zkus
`idgen` naimportovat z jiné lekce (třeba `lessons/lesson-03/exercise`) a přečti si, co
řekne kompilátor.

např. `NewIDGen("prod").NewID()` → `"prod-000001"`

```bash
make lesson L=32
make race L=32
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `32`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=32` prochází
- [ ] Umíš vysvětlit, co přesně `internal/` zakazuje a kdo to vynucuje
- [ ] Umíš říct, proč `golang-standards/project-layout` není standard
- [ ] Umíš zdůvodnit, proč `pricing.Total` bije `service.PricingService`
- [ ] Umíš uvést tři způsoby, jak rozbít cyklický import
- [ ] Umíš vyjmenovat, co smí a co nesmí být v kompozičním kořeni

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md). Wiring
a přeexportované aliasy si nech vygenerovat; hranice balíčků a směr závislostí navrhuj
sám, protože přesně tohle AI nejčastěji zkazí překladem ze Symfony.

## Další čtení

1. [Organizing a Go module](https://go.dev/doc/modules/layout) — oficiální dokumentace k layoutu
2. [Effective Go — Package names](https://go.dev/doc/effective_go#package-names)
3. [Go blog — Package names](https://go.dev/blog/package-names)
4. [cmd/go — Internal Directories](https://pkg.go.dev/cmd/go#hdr-Internal_Directories)
