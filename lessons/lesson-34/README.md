# Lekce 34 — Doménové typy: Money a value objekty

> **Čas:** ~90 min · **Fáze:** 4 — Architektura v Go · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Vysvětlit na konkrétním čísle, proč `float64` na peníze nikdy nepatří, a spočítat,
  kde se ztráta projeví na faktuře.
- Navrhnout value objekt v Go: neměnný, porovnatelný přes `==`, použitelný jako
  klíč mapy a bez cizích závislostí.
- Rozhodnout mezi konstruktorem s validací a užitečnou nulovou hodnotou.
- Použít pojmenované typy proti primitive obsession a vysvětlit, proč
  `type UserID string` není kosmetika.
- Rozdělit částku mezi n příjemců beze ztráty jediného haléře a doložit to testem.

## Teorie

### Proč `float64` na peníze nikdy

`float64` je binární plovoucí čárka podle IEEE 754. Desetinná čísla jako 0.1 v ní
nemají přesnou reprezentaci, stejně jako 1/3 nemá přesný zápis v desítkové soustavě.

```go
fmt.Println(0.1 + 0.2)          // 0.30000000000000004
fmt.Println(0.1+0.2 == 0.3)     // false
```

Na jedné položce je odchylka neviditelná. Problém nastane, když ji vynásobíš:

```go
var total float64
for i := 0; i < 10_000; i++ {
	total += 0.07      // deset tisíc položek po sedmi centech
}
fmt.Println(total)     // 700.0000000000217, ne 700
```

Faktura pak nesedí o setiny, kontrolní součet v účetnictví neprojde a ty hledáš
chybu tam, kde žádná není. Řešení je triviální a používá ho každý platební systém:
**drž částku v celých minoritních jednotkách.** 19,99 EUR je `int64(1999)`. Sčítání,
odečítání i násobení celým číslem jsou pak přesné a `int64` ti stačí zhruba do
92 biliard centů.

Kde `float64` naopak smysl má: kurzy, procenta, statistika. Jakmile ale výsledek
vrací do peněz, musíš explicitně zaokrouhlit a vrátit se k celým číslům.

### Value objekt je pojmenovaný typ s metodami

Value objekt je hodnota definovaná svým obsahem, ne identitou. Dvě stokoruny jsou
zaměnitelné. V Go z toho plynou čtyři konkrétní designová rozhodnutí:

```go
type Money struct {
	cents    int64
	currency Currency
}
```

1. **Neexportovaná pole.** Nikdo zvenčí nesestaví `Money{cents: 100}` s nesmyslnou
   měnou. Jediná cesta dovnitř je konstruktor.
2. **Hodnotový receiver.** `func (m Money) Add(...)` dostane kopii, takže metoda
   nemůže zmutovat příjemce, ani kdyby chtěla. Metody vracejí **novou hodnotu**:

```go
price, _ := NewMoney(1999, "EUR")
double := price.Mul(2)      // price je pořád 19.99
```

3. **Jen porovnatelné typy uvnitř.** `int64` a `string` jsou porovnatelné, takže
   funguje `==` a `Money` jde použít jako klíč mapy. Kdybys dovnitř dal slice nebo
   mapu, `==` by přestalo kompilovat a klíčem už by to nebylo:

```go
a, _ := NewMoney(1999, "EUR")
b, _ := NewMoney(1999, "EUR")
fmt.Println(a == b)              // true

counts := map[Money]int{}
counts[a]++                      // legální, protože Money je comparable
```

4. **`fmt.Stringer`.** Jakmile má typ `String() string`, vypíše ho `fmt.Println`,
   `%v` i `slog` čitelně. Je to jednořádková investice s obrovskou návratností
   v logu a v testovacích hláškách.

Když typ potřebuje jít do JSON, přidej `MarshalJSON`. Bez ní by se serializoval
prázdný objekt, protože pole jsou neexportovaná:

```go
func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Cents    int64    `json:"cents"`
		Currency Currency `json:"currency"`
	}{m.cents, m.currency})
}
```

### Konstruktor s validací vs užitečná nulová hodnota

Lekce 03 tvrdila, že užitečná nulová hodnota je designový cíl. Tady narazíš na její
mez: `Money{}` je nula bez měny a nikdo nedokáže rozhodnout, jestli je to
„nula eur", nebo „nevyplněno". Máš dvě poctivé možnosti:

- **Konstruktor s validací** — `NewMoney(cents int64, c Currency) (Money, error)`.
  Nulová hodnota zůstane technicky použitelná, ale její měna je prázdná, takže
  jakákoli operace s reálnou částkou skončí na `ErrCurrencyMismatch`. Tohle je
  přesně to chování, které chceš: chyba se ozve hned, ne až na faktuře.
- **Zvolit výchozí měnu** — funguje jen v aplikaci s jedinou měnou a zákeřně selže
  v den, kdy přibude druhá.

Neexistuje třetí varianta „konstruktor, který paniká". Neplatný vstup od uživatele
je běžný stav, ne programátorská chyba, takže patří do `error`.

### Primitive obsession a pojmenované typy

V PHP máš `string $userId` a `string $orderId` a jediné, co ti brání je prohodit,
je pozornost. Go dává obranu zdarma:

```go
type UserID string
type OrderID string

func Load(u UserID) {}

var o OrderID = "ord-1"
Load(o)          // chyba kompilace: cannot use o (OrderID) as UserID
Load(UserID(o))  // projde, ale musels to napsat — a to je ta pointa
```

Pojmenovaný typ nad `string` nestojí ani bajt paměti navíc, nemá režii za běhu
a přesune celou třídu chyb z produkce do kompilace. Stejně tak `type Currency string`
v téhle lekci: nejde omylem předat jméno produktu jako měnu.

### Dělení peněz beze ztráty haléřů

Rozděl 100 centů mezi tři příjemce. Naivní `100/3` dá 33 a jeden cent se ztratí.
V účetnictví je to nepřijatelné — součet dílů se musí rovnat originálu na jednotku
přesně. Standardní postup (popsal ho Martin Fowler v *Patterns of Enterprise
Application Architecture*) má dva kroky:

1. Spočítej základní díl celočíselným dělením: `base := cents / n`.
2. Zbytek `cents - base*n` rozdej po jedné jednotce od začátku.

```go
// 100 centů na tři díly
base := int64(100 / 3)        // 33
rem := 100 - base*3           // 1
// výsledek: 34, 33, 33 — součet přesně 100
```

Pro záporné částky jde zbytek stejným směrem, takže −100 na tři díly dá
−34, −33, −33. Verze s poměry (`AllocateRatio`) je stejná myšlenka: každý díl je
`cents * ratio_i / total` a zbytek se rozdá od začátku. Klasický příklad je 5 centů
v poměru 3 : 7 — naivní zaokrouhlení dá 1 a 3, tedy 4 centy; správná alokace dá 2 a 3.

Kdo dostane cent navíc, je věcné rozhodnutí, ne technické. Důležité je, že je
deterministické a otestované. Právě proto test v tomhle cvičení nekontroluje jen
pár konkrétních čísel, ale na dvou tisících generovaných vstupech ověřuje invariant
„součet dílů je přesně originál".

## Rozdíly proti PHP

V Symfony aplikaci vypadá cena typicky takhle:

```php
final class Order
{
    public float $total = 0.0;          // nebo string kvůli Doctrine DECIMAL

    public function addItem(float $price): void
    {
        $this->total += $price;         // mutace, žádná kontrola měny
    }
}

$order->addItem(0.1);
$order->addItem(0.2);
var_dump($order->total === 0.3);        // bool(false)
```

Kdo to má rozmyšlené, sáhne po `moneyphp/money` — knihovně, která dělá přesně to,
co za chvíli napíšeš sám. Její jádro je celočíselná částka plus měna a metody, které
vracejí nový objekt.

Go protějšek:

```go
type Money struct {
	cents    int64
	currency Currency
}

func (m Money) Add(o Money) (Money, error) { /* vrací novou hodnotu */ }
```

Dvě věci se v uvažování mění. Za prvé: `Money` je **hodnota**, ne objekt. Kopíruje
se při přiřazení i při předání do funkce, takže „sdílená mutovatelná cena" ani
neexistuje. Za druhé: nepotřebuješ knihovnu. Struct se dvěma poli a devět metod je
padesát řádků, které si celé přečteš.

Návyk k opuštění: **přestaň hledat balíček.** V PHP je `composer require` levnější
než napsat vlastní typ, protože každá třída s magií kolem sebe je práce. V Go je
vlastní doménový typ nejlevnější a nejčitelnější řešení a přidaná závislost je
naopak drahá.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `float64` nebo `string` na částku | zvyk z PHP a Doctrine `DECIMAL` | `int64` v minoritních jednotkách |
| Metoda mutuje příjemce | reflex z mutovatelných PHP objektů | hodnotový receiver, vracej novou hodnotu |
| Sčítání částek bez kontroly měny | v PHP je částka jen číslo | porovnej měnu a vrať `ErrCurrencyMismatch` |
| `Allocate` přes `math.Round` | dělení chápané jako desetinné | celočíselné dělení plus rozdaný zbytek |
| Slice nebo mapa uvnitř value objektu | „ať to unese víc" | jen porovnatelná pole, jinak přijdeš o `==` |
| `string` místo `type UserID string` | primitive obsession z PHP | pojmenovaný typ, záměna je chyba kompilace |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 34`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `NewMoney`, `Cents`, `Currency`, `String`

```bash
make lesson L=34 PART=1
```

Pak **`/go-deep-review 34 easy`**.

### Střední

Funkce: `Add`, `Sub`, `Mul`, `IsZero`

```bash
make lesson L=34 PART=2
```

Pak **`/go-deep-review 34 medium`**.

### Obtížný

Funkce: `Neg`, `Compare`, `Allocate`, `AllocateRatio`, `ParseMoney`

```bash
make lesson L=34 PART=3
```

Pak **`/go-deep-review 34 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 34 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=34` (+ `make race L=34`, pokud to lekce vyžaduje).

- [ ] Umíš na příkladu ukázat, kde `float64` u peněz selže, a odhadnout velikost chyby
- [ ] Umíš vyjmenovat čtyři vlastnosti, které dělají typ value objektem v Go
- [ ] Umíš vysvětlit, proč `Money` funguje jako klíč mapy a co by to rozbilo
- [ ] Umíš popsat algoritmus alokace zbytku a doložit, proč nic neztrácí
- [ ] Umíš zdůvodnit, kdy sáhnout po konstruktoru s chybou a kdy po nulové hodnotě

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md). Tabulkové
testy, `MarshalJSON` a formátovací pomocníky si nech vygenerovat. Alokační algoritmus
napiš sám a ověř ho invariantem, ne příklady — AI tady s oblibou navrhne
`math.Round(float64(cents)/float64(n))`, což je přesně ta chyba, kvůli které lekce
existuje.

## Další čtení

1. [Go blog — Constants](https://go.dev/blog/constants) — proč je celočíselná aritmetika v Go předvídatelná
2. [pkg.go.dev — fmt.Stringer](https://pkg.go.dev/fmt#Stringer)
3. [Go spec — Comparison operators](https://go.dev/ref/spec#Comparison_operators) — které typy jsou porovnatelné
4. [Effective Go — Methods on values](https://go.dev/doc/effective_go#pointers_vs_values)
