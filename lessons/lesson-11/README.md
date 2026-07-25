# Lekce 11 — Balíčky, export a viditelnost

> **Čas:** ~85 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Vysvětlit, proč je jednotkou zapouzdření v Go balíček a ne typ, a co to mění v návrhu domény.
- Rozhodnout, jestli kód patří do veřejného balíčku, do `internal/`, nebo zůstane neexportovaný.
- Rozplést cyklický import a přeformulovat návrh tak, aby závislosti tekly jedním směrem.
- Napsat vícesouborový a víc-balíčkový program s doc comments a bez `init()`.

## PHP → Go most

V PHP je jednotkou zapouzdření **třída**. Modifikátor visibility platí na úrovni objektu,
takže ani jiná instance téže třídy… vlastně ano, může:

```php
final class Amount
{
    public function __construct(private readonly int $cents) {}

    public function add(Amount $other): Amount
    {
        return new self($this->cents + $other->cents); // cizí instance, ale stejná třída → OK
    }

    public function cents(): int { return $this->cents; }
}
```

Go tenhle princip rozšiřuje o řád: hranice není třída, ale **balíček**.

```go
package money

type Amount struct {
	cents int64 // malé písmeno = neexportované
}

func SumCents(amounts []Amount) int64 {
	var total int64
	for _, a := range amounts {
		total += a.cents // jiná funkce, jiná instance — pořád stejný balíček, takže OK
	}
	return total
}
```

Mimo balíček `money` pole `cents` neexistuje. Ani reflexí ho slušně nepřečteš, ani ho
nenastavíš v composite literalu. Návyk, který je potřeba opustit: **přestaň dělit kód podle
tříd a vrstev a začni ho dělit podle hranic, kde chceš mít kontrolu nad tím, co je vidět.**
Balíček je tvůj `private`, jeden soubor není nic.

## Teorie

### Velké písmeno místo klíčových slov

Go nemá `public`, `private`, `protected` ani `friend`. Má jediné pravidlo: identifikátor
začínající **velkým písmenem** je exportovaný a je vidět z jiných balíčků, cokoli jiného
je vidět jen uvnitř balíčku, kde je deklarované.

To platí uniformně pro všechno — typy, funkce, metody, pole struktur, konstanty i
proměnné na úrovni balíčku:

```go
package catalog

type Product struct {   // exportovaný typ
	Name  string        // exportované pole
	price int64         // neexportované pole
}

func New(name string) Product { … }   // exportovaná funkce
func normalize(s string) string { … } // jen pro tenhle balíček
```

Typ `Product` je zvenku použitelný, ale `price` z něj nikdo cizí nepřečte. Zároveň si
všimni, že `catalog.Product{Name: "kniha"}` se zvenku vytvořit dá, protože `Name` je
exportované. Když chceš vynutit konstruktor, nesmí být exportované **žádné** pole —
přesně to dělá `money.Amount` v cvičení.

Pravidlo je čistě lexikální, takže se řídí prvním písmenem, ne celým jménem. `URL` je
exportované, `urlPath` ne, `_secret` taky ne (podtržítko není velké písmeno).

### Balíček je jednotka zapouzdření

Tohle je nejdůležitější věta lekce. Uvnitř jednoho balíčku **neexistuje soukromí**.
Když má balíček deset souborů, vidí každý z nich všechny neexportované identifikátory
ostatních. Není potřeba nic importovat, žádné `use`, žádný autoloader — soubory se stejným
`package foo` na začátku tvoří jedno jméno prostoru.

Praktický důsledek: rozdělení kódu do souborů je čistě organizační pohodlí pro čtenáře.
Chceš-li skutečnou hranici, musíš udělat nový **balíček**, tedy nový adresář.

Druhý důsledek, který PHP vývojáře překvapí: uvnitř balíčku můžeš sáhnout na privátní
pole **cizího typu**, ne jen na `$this`. Právě proto se v Go dá napsat efektivní
`func SumCents(amounts []Amount) int64`, která nemusí pro každou položku volat getter.
Testy ze stejného balíčku (`package money`) tuhle výhodu mají taky — proto existuje
konvence psát „bílé skříňky" jako `package money` a „černé skříňky" jako `package money_test`.
V tomhle kurzu píšeme testy jako externí balíček právě proto, aby tě kompilátor donutil
používat jen veřejné API.

### `internal/` — vynutitelná hranice modulu

Někdy chceš kód sdílet mezi několika svými balíčky, ale nechceš ho vystavit světu.
Na to je adresář `internal/`, na který kompilátor uplatňuje speciální pravidlo:

> Balíček uvnitř `internal/` smí importovat jen kód, jehož cesta začíná rodičem toho
> `internal/` adresáře.

```
myapp/
  cmd/api/main.go          // smí importovat myapp/internal/store
  internal/
    store/store.go         // vidí ho jen myapp/...
  pkg/client/client.go      // veřejné API modulu
```

Pokusí-li se cizí modul napsat `import "myapp/internal/store"`, kompilace skončí chybou
`use of internal package not allowed`. Není to konvence ani linter — je to jazykové
pravidlo. V Symfony bys tohle řešil dohodou, dokumentací nebo statickou analýzou.
Go to prostě neumožní.

Pravidlo: **výchozí volba pro aplikační kód je `internal/`.** Do veřejného balíčku dej
jen to, co opravdu chceš, aby na tobě mohl někdo postavit závislost.

### Cyklické importy jsou zakázané

Když balíček `order` importuje `customer` a `customer` importuje `order`, kompilace
skončí chybou `import cycle not allowed`. Žádné lazy loading, žádné forward deklarace.

Tohle je zpočátku frustrující, protože v PHP je vzájemná závislost tříd běžná. Ale je to
funkce, ne chyba: nutí tě to udělat směr závislostí explicitní. Tři standardní východiska:

1. **Sluč balíčky.** Když se dva balíčky nutně potřebují navzájem, je to často jeden
   koncept rozřezaný na dva adresáře. Nejčastější správná odpověď.
2. **Vytáhni sdílené jádro dolů.** Doménové typy (`money`, `user`) dej do balíčku, který
   neimportuje nic z tvé aplikace, a nech oba původní balíčky záviset na něm.
3. **Otoč závislost přes interface.** Balíček, který službu *používá*, si definuje malý
   interface u sebe; implementace zůstane nahoře. Detail v lekcích 12 a 33.

Co **není** řešení: vytvořit balíček `common`, `utils` nebo `helpers` a nasypat do něj
všechno, co se nikam nevešlo. Cyklus tím zmizí, ale hranice taky.

### Jméno balíčku vs jméno adresáře, importy

Jméno balíčku je to, co je za `package`, a používá se při volání. Cesta v importu je
adresář. Obojí se **má** shodovat, ale kompilátor to nekontroluje:

```go
import "github.com/rdurica/go-deep/lessons/lesson-11/exercise/money"

a := money.New(1999) // jméno z klauzule package, ne z cesty
```

Konvence pro jména balíčků: krátké, malými písmeny, jednoslovné, bez podtržítek a bez
velbloudů — `http`, `strconv`, `money`. Jméno se opakuje při každém použití, takže
`money.New` je lepší než `money.NewMoney` a `moneyutils.NewMoneyAmount` je katastrofa.
Balíček se nejmenuje podle vrstvy (`services`, `models`), ale podle toho, co poskytuje.

Když se dva importy jmenují stejně nebo je jméno matoucí, dá se import pojmenovat:

```go
import (
	crand "crypto/rand"
	"math/rand"
)
```

Zvláštní případ je **blank import** — importuješ balíček jen kvůli jeho vedlejšímu efektu
(registraci v nějakém globálním registru), sám ho nepoužíváš:

```go
import _ "image/png" // zaregistruje PNG dekodér do image.Decode
```

Bez podtržítka by kompilátor nepřeložený import odmítl: nepoužitý import je v Go **chyba**,
ne varování.

### `init()` a proč se mu vyhýbat

Každý soubor může mít funkci `func init()`, která se spustí po inicializaci proměnných
balíčku a před `main`. Balíček jich může mít víc a pořadí mezi soubory není nic, na co
by ses měl spoléhat.

```go
var registry = map[string]Handler{}

func init() {
	registry["json"] = jsonHandler{} // funguje, ale je to skrytá magie
}
```

Problémy: `init()` nejde zavolat ručně, nejde mu předat parametry, nemůže vrátit chybu
(jen panikovat), spustí se i v testech a znemožňuje řízené sestavení závislostí. Je to
Go ekvivalent statické inicializace v PHP bootstrapu — a stejně jako tam platí, že
explicitní `New…()` volané z `main` je čitelnější a testovatelnější.

Legitimní použití jsou vzácná: registrace driverů (`database/sql`), kompilace regulárních
výrazů do balíčkové proměnné (i to jde přes `var re = regexp.MustCompile(...)`).

### Doc comments a `package main`

Komentář těsně nad klauzulí `package` je dokumentace balíčku a zobrazí ho `go doc`
i pkg.go.dev. Píše se jen v jednom souboru balíčku (typicky `doc.go` nebo v tom hlavním)
a začíná slovem `Package`:

```go
// Package money je doménový balíček pro peněžní částky v celých centech.
package money
```

Doc comment exportovaného identifikátoru začíná jeho jménem: `// New vytvoří…`,
`// Amount je…`. Není to estetika — nástroje z toho skládají větu v dokumentaci
a `golint`/revive to hlídají.

`package main` je speciální: označuje spustitelný program, musí obsahovat
`func main()`, a nikdo ho nemůže importovat. Jméno binárky se bere z názvu adresáře,
ne z `main`. Proto se spustitelné programy dávají do `cmd/nazev/main.go` a veškerá
skutečná logika žije v importovatelných balíčcích — jinak by nešla otestovat ani znovu použít.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Balíček na každou třídu | reflex „jeden soubor = jedna třída" z PSR-4 | balíček je hranice viditelnosti, ne obal na typ |
| Balíčky `service`, `model`, `repository` | vrstvený layout ze Symfony | pojmenuj podle domény: `money`, `order`, `store` |
| Getter ke každému neexportovanému poli | zvyk na PHP encapsulation | uvnitř balíčku sáhni na pole přímo, ven vystav jen to potřebné |
| Cyklus se „vyřeší" balíčkem `utils` | snaha zachovat původní rozdělení | sluč balíčky nebo vytáhni doménové jádro dolů |
| `init()` pro sestavení závislostí | zvyk na bootstrap/DI kontejner | explicitní konstruktor volaný z `main` |
| Exportováno „pro jistotu" | v PHP se `public` píše automaticky | začni neexportovaně, exportuj až když je potřeba |
| `NewMoneyAmount` v balíčku `money` | jméno balíčku se nezapočítává | `money.New`, jméno balíčku je součást volání |

## Úkol

Pracuj v `exercise/`. Cvičení je **víc-balíčkové**: kromě `exercise/exercise.go` upravuješ
i `exercise/money/money.go`. Postupuj A → B → C, po každé části spusť test.

Testy jsou v `package exercise_test`, tedy mimo oba balíčky. Volají jen `exercise.*`,
protože hlavní balíček má nad `money` tenkou fasádu (`NewAmount`, `SumCents`, `Split`) —
ty jsou hotové, neimplementuj je. Zkus si po dokončení do testu napsat
`a.cents` a podívej se, co řekne kompilátor.

### A — rozcvička (~10 min)

V `exercise/money/money.go` implementuj:

- `New(cents int64) Amount` — konstruktor, jediná cesta, jak částku vyrobit zvenku.
- `func (a Amount) Cents() int64` — vrací hodnotu v centech.
- `func (a Amount) String() string` — formát s právě dvěma desetinnými místy a znaménkem:
  `0` → `"0.00"`, `5` → `"0.05"`, `1999` → `"19.99"`, `-1` → `"-0.01"`, `-250` → `"-2.50"`.
  Tím zároveň splníš `fmt.Stringer`.

Zero value `Amount{}` musí dál fungovat jako nula — nespoléhej na to, že konstruktor
proběhl vždy.

např. `New(1999).String()` → `"19.99"`

### B — jádro (~35 min)

V `exercise/exercise.go`, tedy **zvenku** balíčku `money`:

1. `TotalOf(amounts []Amount) Amount` — součet částek. Prázdný nebo `nil` vstup dá nulovou
   částku. Smíš použít jen veřejné API `money` — na `cents` se odsud nedostaneš.
2. `MustParse(s string) Amount` — převede textový zápis na částku:
   - volitelné znaménko `+` nebo `-` na začátku,
   - povinná celá část z jedné a více číslic,
   - volitelná desetinná část: tečka a jedna nebo dvě číslice (`"2.5"` je 250 centů).

   Platné: `"0"`, `"7"`, `"19.99"`, `"-2.5"`, `"+3.05"`, `"-0.01"`.
   Neplatné (funkce **panikuje**): `""`, `"abc"`, `"1.234"`, `"1."`, `".5"`, `"1.2.3"`,
   `"--1"`, `"1,5"`, `" 1"`, `"1 "`. Bílé znaky se neořezávají.

   Prefix `Must` je konvence: „při špatném vstupu panikuj místo vracení chyby".
   Používá se jen tam, kde je vstup konstanta v kódu (viz `regexp.MustCompile`).
   Pomocné funkce si klidně přidej — nech je neexportované.

např. `MustParse("19.99").Cents()` → `1999`

### C — rozšíření (~25 min)

Zpět v `exercise/money/money.go`, tedy **uvnitř** balíčku:

1. `SumCents(amounts []Amount) int64` — součet v centech. Napiš ho tak, aby sahal
   **přímo na `a.cents`**, ne přes metodu `Cents()`. Tohle je ta ukázka zapouzdření:
   uvnitř balíčku je pole cizí instance dostupné, mimo něj neexistuje. `nil` vstup dá 0.
2. `Split(a Amount, n int) ([]Amount, bool)` — rozdělí částku na `n` dílů tak, aby se
   **žádný cent neztratil**. Zbytek po celočíselném dělení rozdej po jednom centu od
   prvního dílu: `1000` na 3 díly → `334, 333, 333`. U záporných částek jde zbytek taky
   „ven od nuly": `-250` na 3 díly → `-84, -83, -83`. Pro `n <= 0` vrať `(nil, false)`.

Bonus bez testu: vytvoř v repozitáři vedle sebe adresář `internal/` s balíčkem a zkus ho
importovat z jiného modulu. Chyba `use of internal package ... not allowed` je to,
co chceš vidět.

např. `Split(1000, 3)` → `[334, 333, 333], true`

```bash
make lesson L=11
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `11`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=11` prochází
- [ ] Umíš vysvětlit, proč `SumCents` smí na `a.cents` a `TotalOf` ne
- [ ] Umíš vysvětlit, co přesně `internal/` zakazuje a komu
- [ ] Umíš popsat tři způsoby, jak rozplést cyklický import
- [ ] Umíš říct, proč je `init()` horší než explicitní konstruktor
- [ ] Umíš vysvětlit, k čemu je `_` před importem a proč je nepoužitý import chyba

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

## Další čtení

1. [Effective Go — Names](https://go.dev/doc/effective_go#names)
2. [Go Modules Reference — internal packages](https://go.dev/ref/mod#internal-packages)
3. [Go Doc Comments](https://go.dev/doc/comment)
4. [Organizing a Go module](https://go.dev/doc/modules/layout)
