# Lekce 02 — Mentální reset: PHP → Go

> **Čas:** ~75 min · **Fáze:** 0 — Setup a mentální reset · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Vysvětlit, co konkrétně pro tvůj denní workflow znamená, že Go je kompilovaný a staticky typovaný.
- Rozhodnout, kdy v Go potřebuješ typ s metodami a kdy stačí obyčejná funkce.
- Nahradit reflex „vyhoď výjimku" a „zdědit z base třídy" tím, co Go skutečně nabízí.
- Vysvětlit, proč se v Go místo DI kontejneru píše wiring ručně v `main` a proč to není krok zpět.
- Přečíst signaturu Go funkce a poznat z ní, jestli se předává hodnota nebo se něco sdílí.

## PHP → Go most

Typický Symfony návrh: služba, konstruktor se závislostmi, výjimka při chybě.

```php
final class PriceCalculator
{
    public function __construct(
        private readonly DiscountPolicy $policy,
        private readonly LoggerInterface $logger,
    ) {}

    public function total(array $items): int
    {
        if ($items === []) {
            throw new EmptyCartException();
        }
        // ...
    }
}
```

Stejná úloha v Go:

```go
// TotalCents je jen funkce — nemá stav, nepotřebuje objekt.
func TotalCents(items []Item) int {
	total := 0
	for _, it := range items {
		total += it.PriceCents * it.Qty
	}
	return total
}

// Kde je „nic k vrácení", vrátí se druhá hodnota bool místo výjimky.
func Cheapest(items []Item) (Item, bool) {
	if len(items) == 0 {
		return Item{}, false
	}
	// ...
}
```

Co se mění v uvažování: v PHP je třída výchozí jednotka kódu a všechno ostatní je odchylka.
V Go je výchozí jednotka **balíček** a v něm **funkce**. Typ se strukturou a metodami zavádíš
až ve chvíli, kdy máš skutečný stav, který k sobě patří. `PriceCalculator` bez stavu je v Go
zbytečná vrstva — je to jen jmenný prostor navíc.

## Teorie

### Kompilovaný a staticky typovaný — co to mění v praxi

V PHP je jednotkou nasazení soubor. Změníš ho, refreshneš prohlížeč, hotovo. Typová chyba
v cestě, kterou zrovna neprocházíš, počká, až ji najde uživatel. Statická analýza (PHPStan,
Psalm) je nadstavba, kterou si musíš přidat a vyladit.

Go se kompiluje celé, do jednoho statického binárního souboru. Z toho plyne pár praktických
věcí, na které se v prvních týdnech naráží denně:

- **Nepoužitá proměnná a nepoužitý import jsou chyby kompilace**, ne varování. Není to
  buzerace — je to způsob, jak udržet zdrojáky bez mrtvého kódu bez lintru.
- **Kompilátor je součást vývojového cyklu.** `go build` na běžném projektu trvá jednotky
  sekund a nahrazuje velkou část toho, co v PHP dělá suite jednotkových testů typu
  „zavolá se to vůbec správně?".
- **Nasazuješ jeden soubor.** Žádné PHP-FPM, žádné `vendor/`, žádná verze runtime na serveru.
  Binárka obsahuje i runtime a garbage collector.
- **Neexistuje `mixed` ani pole jako univerzální kontejner.** Když v PHP píšeš
  `array{name: string, price: int}` do PHPDocu, v Go napíšeš `struct` a kompilátor to vynutí.

```go
func main() {
	items := []Item{{Name: "kava", PriceCents: 4500, Qty: 2}}
	total := TotalCents(items)
	unused := 42 // chyba kompilace: declared and not used
	fmt.Println(total)
}
```

### Chyby jsou hodnoty, ne výjimky

Go nemá `try`/`catch`. Funkce, která může selhat, vrací chybu jako **poslední návratovou
hodnotu**:

```go
n, err := strconv.Atoi("12x")
if err != nil {
	// tady se rozhodneš, co s tím
	return err
}
fmt.Println(n)
```

Detailní práce s chybami — vlastní typy, `errors.Is`, `errors.As`, obalování přes `%w` — je
lekce 14. Teď stačí přijmout tři důsledky:

1. **Chyba je vidět v signatuře.** Nemusíš číst dokumentaci ani `@throws`, abys věděl, že
   funkce může selhat. `func Parse(s string) (Config, error)` to říká rovnou.
2. **Chyba nepropadne nahoru sama.** Když ji ignoruješ, program pokračuje s neplatnou
   hodnotou. Proto se v Go chyba řeší hned na místě vzniku, ne o pět vrstev výš v
   `try { $kernel->handle() }`.
3. **Ne každé „nepovedlo se" je chyba.** Když funkce jen nemá co vrátit — prázdný košík,
   klíč není v mapě — použije se **comma-ok** idiom: druhá návratová hodnota `bool`.

```go
value, ok := m["key"]      // stdlib to dělá takhle
item, ok := Cheapest(items) // a ty to děláš stejně
if !ok {
	// prázdný vstup, není to chyba, jen prázdno
}
```

Tohle je Go náhrada za `null`. Rozdíl proti PHP `?Item` je, že `ok` **musíš** převzít nebo
explicitně zahodit — nemůžeš na něj omylem zapomenout tak, jako se zapomíná na `!== null`.

### Žádná dědičnost, jen kompozice

V Go není `extends`, není abstraktní třída, není `parent::`. Neexistuje způsob, jak vytvořit
hierarchii typů. Existují dva nástroje, které tu potřebu pokrývají:

- **Embedding** — vložíš jeden struct do druhého a jeho metody se „promotují" navenek.
  Vypadá to jako dědičnost, ale je to delegování; podrobně v [lekci 05](../lesson-05/README.md).
- **Interfaces** — malé sady metod, které typ splňuje **implicitně**, bez `implements`.
  Podrobně v lekci 12.

Praktický dopad na návrh: v Symfony je běžné mít `AbstractController`, `AbstractRepository`
nebo `AbstractHandler` s pěti šablonovými metodami. Tenhle vzor v Go nejde přepsat 1:1 a
nemá se o to zkoušet. Místo sdílené základní třídy se sdílená logika dá do funkce a zavolá
se z obou míst. Duplicita tří řádků je v Go levnější než abstrakce, kterou musí čtenář
rozplétat přes tři soubory.

### Balíček je jednotka zapouzdření

V PHP je hranice viditelnosti třída: `private` znamená „jen tato třída". V Go je hranice
**balíček** a řídí se velikostí prvního písmene:

```go
package pricing

type Catalog struct {
	items []Item // malé i — mimo balíček pricing neexistuje
}

func NewCatalog(items []Item) *Catalog { /* ... */ } // velké N — exportováno
func (c *Catalog) Price(name string) (int, bool)     // velké P — exportováno
```

Dva typy ve stejném balíčku na sebe vidí úplně, včetně neexportovaných polí. To zní jako
slabší zapouzdření, ale v praxi vede k jinému, zdravějšímu členění: balíček odpovídá doméně
(`pricing`, `order`, `auth`), ne vrstvě (`service`, `repository`, `dto`). Vrstvové balíčky
jsou nejčastější zápach „Symfony napsaného v Go" — vedou k tomu, že jeden případ užití je
rozsypaný přes pět balíčků, které na sebe kruhově ukazují. A kruhový import je v Go tvrdá
chyba kompilace, takže se to hned pozná.

### Žádný DI kontejner a stdlib-first

Symfony ti závislosti sestaví autowiringem: přidáš typ do konstruktoru a kontejner ho najde.
Go nic takového nemá a mít nechce — sestavení závislostí je **obyčejný kód v `main`**:

```go
func main() {
	catalog := pricing.NewCatalog(loadItems())
	handler := api.NewHandler(catalog, slog.Default())

	http.ListenAndServe(":8080", handler)
}
```

Výhoda: graf závislostí je vidět v jedné funkci, dá se do něj vstoupit debuggerem a
kompilátor ho kontroluje. Nevýhoda: u velké aplikace je `main` dlouhý. To je přijatelná cena
a řeší se rozdělením na několik konstrukčních funkcí, ne reflexí.

S tím souvisí druhý návyk: **stdlib first**. Ve světě Symfony je první krok u nového problému
`composer require`. V Go se první ptáš, jestli to neumí standardní knihovna — a odpověď je
překvapivě často ano: HTTP server i klient, JSON, šablony, testovací framework, strukturované
logování (`log/slog`), profiler. Celý tenhle kurz vystačí se standardní knihovnou.

### Hodnoty vs reference

V PHP je objekt vždycky handle. Předáš ho do metody, ta ho změní, a změna se projeví u tebe.
Skalární hodnoty a pole se naopak kopírují (pole líně, přes copy-on-write).

V Go platí jednoduché pravidlo: **všechno se předává hodnotou, vždy**. Když předáš struct,
funkce dostane jeho kopii:

```go
func rename(it Item) {
	it.Name = "jiny" // mění kopii
}

item := Item{Name: "kava"}
rename(item)
fmt.Println(item.Name) // "kava" — beze změny
```

Když chceš změnu vidět, musíš předat pointer (`*Item`) — a ten se také předává hodnotou,
jen ta hodnota je adresa. Detaily jsou [lekce 06](../lesson-06/README.md).

Důsledek pro návrh je zásadní: v Go je normální vracet struct **hodnotou**, protože z toho
plyne, že příjemce dostal svoji kopii a nikdo mu ji zvenčí nezmění. Ve světě PHP by ses na
to spolehnout nemohl, a proto se tam píšou obranné klony a `readonly`. V Go to dostaneš zdarma.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Struct s metodami tam, kde stačí funkce | reflex „všechno je třída" | typ zaveď, až když má stav |
| `panic()` místo návratové chyby | náhrada za `throw` | vrať `error` nebo `(T, bool)` |
| Pointer všude „pro jistotu" | objekty v PHP jsou reference | výchozí je hodnota, pointer až s důvodem |
| Balíčky `service/`, `repository/`, `dto/` | přenesená struktura Symfony | balíček podle domény, ne podle vrstvy |
| Hledání DI kontejneru pro Go | autowiring jako samozřejmost | ruční wiring v `main` |
| Interface „pro každou službu" | zvyk na mockování všeho | interface až u konzumenta, když ho potřebuješ |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

Doména je záměrně ta, kterou znáš: ceník a objednávka. Cena je vždy v **celých centech**
(`int`), nikdy ve `float64` — peníze se v žádném jazyce nepočítají jako desetinné číslo.

### A — rozcvička (~10 min)

`ApplyDiscount(priceCents int, percent int) int` vrátí cenu po slevě.

- Sleva `percent` je v procentech. Hodnoty mimo rozsah `0–100` **ořízni** (clamp): záporné
  na `0`, nad sto na `100`. Nepanikař a nevracej chybu.
- Výsledek zaokrouhli na celé centy, půlku nahoru (`5` centů se slevou 50 % je `3`).
- `priceCents` menší nebo rovno nule vrací `0`.

např. `ApplyDiscount(5, 50)` → `3`

### B — jádro (~35 min)

Struktura `Item` je předpřipravená: `Name string`, `PriceCents int`, `Qty int`.

1. `TotalCents(items []Item) int` — součet `PriceCents * Qty` přes všechny položky. Položky
   s `Qty` menším nebo rovným nule se přeskočí. `nil` i prázdný slice dají `0`. Funkce
   **nesmí měnit vstup**.
2. `Cheapest(items []Item) (Item, bool)` — vrátí celou položku s nejnižší **jednotkovou**
   cenou (`PriceCents`, na `Qty` se nehledí) a `true`. Při shodě ceny vyhrává první výskyt.
   Pro prázdný nebo `nil` vstup vrať `Item{}` a `false` — tohle je comma-ok idiom, tvoje
   náhrada za `null`.

Všimni si, že vracíš `Item` hodnotou. Test ověřuje, že když volající vrácenou položku změní,
původní slice se nezmění.

např. `TotalCents([{kava, 4500, 2}])` → `9000`

### C — rozšíření (~25 min)

Postav malý „service", který má skutečný stav — ceník. Typ `Catalog` je předpřipravený a má
**neexportované** pole `items`.

- `NewCatalog(items []Item) *Catalog` — konstruktor. Vstupní slice si **okopíruj**, aby
  pozdější změna slice u volajícího ceník neovlivnila. (Slice je jediná věc, která se sice
  předává hodnotou, ale ukazuje na sdílené pole — to je téma lekce 07.)
- `func (c *Catalog) Price(name string) (int, bool)` — jednotková cena podle jména, comma-ok.
  Nenalezeno → `0, false`.
- `func (c *Catalog) Checkout(names []string, percent int) (int, bool)` — sečti ceny položek
  podle jmen (opakované jméno se počítá vícekrát), na součet aplikuj `ApplyDiscount` a vrať
  `true`. Pokud kterékoli jméno v ceníku není, vrať `0, false`. Prázdný seznam jmen je
  platná objednávka za `0`.

Až budeš hotový, podívej se na `Checkout`: složil jsi ji z funkce a metody, které už
existovaly, bez kontejneru, bez interface a bez dědičnosti. Tohle je v Go běžný způsob, jak
vzniká „služba".

např. ceník s `kava=4500`, `Checkout(["kava", "kava"], 0)` → `9000, true`

```bash
make lesson L=02
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `02`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=02` prochází
- [ ] Umíš vysvětlit, proč `Cheapest` vrací `(Item, bool)` a ne `*Item`
- [ ] Umíš vysvětlit, co je v Go jednotka zapouzdření a jak se pozná exportovaný identifikátor
- [ ] Umíš vysvětlit, čím nahradíš `AbstractHandler` z Symfony
- [ ] Umíš popsat, co se stane, když do funkce předáš struct a funkce ho změní
- [ ] Umíš zdůvodnit, proč `main` sestavuje závislosti ručně

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

## Další čtení

1. [Effective Go — Introduction a Names](https://go.dev/doc/effective_go)
2. [Go FAQ — Is Go an object-oriented language?](https://go.dev/doc/faq#Is_Go_an_object-oriented_language)
3. [Go blog — Errors are values](https://go.dev/blog/errors-are-values)
4. [Go Proverbs](https://go-proverbs.github.io/)
