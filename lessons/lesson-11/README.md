# Lekce 11 — Balíčky, export a viditelnost

> **Čas:** ~85 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Vysvětlit, proč je jednotkou zapouzdření v Go balíček a ne typ, a co to mění v návrhu domény.
- Rozhodnout, jestli kód patří do veřejného balíčku, do `internal/`, nebo zůstane neexportovaný.
- Rozplést cyklický import a přeformulovat návrh tak, aby závislosti tekly jedním směrem.
- Napsat vícesouborový a víc-balíčkový program s doc comments a bez `init()`.

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

## Rozdíly proti PHP

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

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 11`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/money/` (balíček `money`). Kontrakt je v komentáři nad metodou.
Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `String` v `money/` (záměrně vadný formát u záporných a desetinných míst)

```bash
make lesson L=11 PART=1
```

Pak **`/go-deep-review 11 easy`**.

### Střední

Implementuj: `SumCents` v `money/` (sčítání přes neexportované pole `cents`)

```bash
make lesson L=11 PART=2
```

Pak **`/go-deep-review 11 medium`**.

### Obtížný

Doplň: `Split` v `money/` (dělení bez ztráty centů)

```bash
make lesson L=11 PART=3
```

Pak **`/go-deep-review 11 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 11 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=11` (+ `make race L=11`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč `SumCents` smí na `a.cents` a `TotalOf` ne
- [ ] Umíš vysvětlit, co přesně `internal/` zakazuje a komu
- [ ] Umíš popsat tři způsoby, jak rozplést cyklický import
- [ ] Umíš říct, proč je `init()` horší než explicitní konstruktor
- [ ] Umíš vysvětlit, k čemu je `_` před importem a proč je nepoužitý import chyba

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Mentor, kvíz i review (dialog) jsou vždy OK; v tomto režimu AI nesmí psát kód cvičení.
## Další čtení

1. [Effective Go — Names](https://go.dev/doc/effective_go#names)
2. [Go Modules Reference — internal packages](https://go.dev/ref/mod#internal-packages)
3. [Go Doc Comments](https://go.dev/doc/comment)
4. [Organizing a Go module](https://go.dev/doc/modules/layout)
