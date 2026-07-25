# Lekce 05 — Structs, metody a embedding

> **Čas:** ~90 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Vytvořit struct, zvolit mezi pojmenovaným a pozičním literálem a vědět, proč se ten druhý
  skoro nepoužívá.
- Rozhodnout mezi value a pointer receiverem podle pravidla, které obhájíš v code review.
- Vysvětlit, co je method set a proč hodnota typu nesplní interface, jehož metody mají
  pointer receiver.
- Použít embedding pro kompozici a popsat, čím se liší od dědičnosti.

## PHP → Go most

V PHP je třída zároveň datová struktura, jmenný prostor pro metody i uzel v hierarchii.

```php
class Base
{
    public function __construct(protected string $id) {}
    public function describe(): string { return "base:{$this->id}"; }
}

class User extends Base
{
    public function __construct(string $id, private string $name) { parent::__construct($id); }
    public function describe(): string { return "user:{$this->name} (" . parent::describe() . ")"; }
}
```

V Go se data (`struct`) a chování (metody) definují zvlášť a místo dědičnosti se vkládá:

```go
type Base struct{ ID string }

func (b Base) Describe() string { return "base:" + b.ID }

type User struct {
	Base        // embedding, ne extends
	Name string
}

func (u User) Describe() string {
	return "user:" + u.Name + " (" + u.Base.Describe() + ")"
}
```

Co se mění v uvažování: `u.Base.Describe()` **není** `parent::describe()`. Není tu žádný
rodič — `User` prostě obsahuje hodnotu typu `Base` jako pole a ty na ni voláš metodu.
Rozdíl je vidět, jakmile do hry vstoupí interfaces: v PHP `Base` zavolá přepsanou metodu
potomka (pozdní vazba), v Go `Base.Describe()` o existenci `User` nikdy nezjistí nic.

## Teorie

### Struct literály

Struct je pojmenovaná kolekce polí. Zero value je struct, kde má každé pole svoji zero value,
takže `var p Point` je rovnou použitelný bod `(0,0)`.

```go
type Point struct {
	X, Y int
}

p1 := Point{X: 1, Y: 2} // pojmenovaná pole — používej tohle
p2 := Point{1, 2}       // poziční — jen u malých, ustálených typů
p3 := Point{X: 1}       // Y zůstane 0
var p4 Point            // {0 0}
```

Poziční literál musí vyjmenovat **všechna** pole ve správném pořadí. To znamená, že přidání
pole do structu rozbije každý poziční literál v celém programu — a u exportovaného typu
i v cizím kódu. Proto je pojmenovaná varianta výchozí volba a poziční si nech pro
`Point`-like typy, kde je pořadí samozřejmé.

Anonymní struct je typ bez jména, definovaný na místě použití. Nejčastěji ho uvidíš
v tabulkových testech, přesně jako v testech k téhle lekci:

```go
tests := []struct {
	name string
	in   int
	want string
}{
	{"nula", 0, "zero"},
	{"kladné", 1, "positive"},
}
```

### Porovnatelnost

Struct se dá porovnat operátorem `==`, pokud jsou porovnatelná všechna jeho pole. Porovnává
se pole po poli, ne adresa:

```go
a := Point{1, 2}
b := Point{1, 2}
fmt.Println(a == b) // true
```

Porovnatelné jsou čísla, řetězce, booly, pointery, kanály, pole a structy z nich. Naopak
**slice, mapa a funkce porovnatelné nejsou** — struct, který je obsahuje, tím porovnatelnost
ztratí a `==` na něm je chyba kompilace. Na hloubkové porovnání takových typů je
`reflect.DeepEqual` (v testech), pro produkční kód se píše metoda `Equal`.

Porovnatelnost má jeden praktický důsledek: jen porovnatelný typ může být klíčem mapy.
`map[Point]string` funguje, `map[struct{ tags []string }]int` ne.

### Metody a receiver

Metoda je funkce s receiverem uvedeným před jménem. Může být na jakémkoli pojmenovaném typu
definovaném ve stejném balíčku — tedy i na `type Celsius float64`, nejen na structu.

```go
func (p Point) Add(q Point) Point { // value receiver
	return Point{p.X + q.X, p.Y + q.Y}
}

func (c *Counter) Inc() { // pointer receiver
	c.n++
}
```

Value receiver dostane **kopii**. Cokoli v ní změníš, zahodí se při návratu. Pointer receiver
dostane adresu, takže vidí a mění originál.

Rozhodovací pravidlo, které obstojí v review:

1. **Musí metoda měnit stav?** → pointer receiver. Jiná možnost není.
2. **Je struct velký, nebo obsahuje `sync.Mutex`?** → pointer receiver. Kopírovat mutex je
   chyba (a `go vet` na ni upozorní).
3. **Jinak** value receiver, hlavně u malých neměnných hodnotových typů (`time.Time`,
   `Point`, `Money`).
4. **Konzistence vyhrává.** Pokud jedna metoda typu potřebuje pointer receiver, dej ho
   všem. Míchání je matoucí a plodí subtilní chyby s method setem.

Na volání to většinou nemá vliv, protože Go za tebe doplní `&` a `*`:

```go
var c Counter
c.Inc()   // ve skutečnosti (&c).Inc() — c je adresovatelné
p := &Point{1, 2}
p.Add(q)  // ve skutečnosti (*p).Add(q)
```

Podmínka je **adresovatelnost**. Proměnná ji má, výsledek volání funkce nebo prvek mapy ne:
`counters["a"].Inc()` je chyba kompilace.

### Method set a interfaces

Method set je seznam metod, které typ „umí" z pohledu interface. Pravidlo je asymetrické:

| Typ | Method set |
|-----|------------|
| `T` | metody s receiverem `T` |
| `*T` | metody s receiverem `T` **i** `*T` |

Proto tohle nejde:

```go
type Incrementer interface{ Inc() }

var i Incrementer = Counter{}  // chyba kompilace
var j Incrementer = &Counter{} // OK
```

Důvod je logický: kdyby hodnota `Counter{}` uložená v interface směla mít `Inc()`, volala by
se metoda na kopii uvnitř interface a změna by se nikam neprojevila. Kompilátor tomu radši
zabrání. Když ti tedy vyskočí `X does not implement Y (method Z has pointer receiver)`, víš,
že máš předat `&x`.

### Embedding a promotion

Pole bez jména (jen typ) je *embedded field*. Jeho pole i metody se promotují na vnější typ:

```go
type Admin struct {
	User      // embedded
	Level int
}

a := Admin{User: User{Base: Base{ID: "a1"}, Name: "Root"}, Level: 9}
a.ID          // promotováno přes dvě úrovně: a.User.Base.ID
a.Describe()  // promotováno z User
```

Promotovaná metoda se počítá i do method setu, takže `Admin` splní interface `Describer`,
aniž bys `Describe` napsal. Přesně tohle se používá ve stdlib — `bufio.ReadWriter` vkládá
`*Reader` a `*Writer` a tím splní `io.ReadWriter`.

**Shadowing:** když vnější typ definuje metodu se stejným jménem, promotovaná se překryje.
Původní zůstane dostupná přes explicitní cestu `u.Base.Describe()`. Nejde ale o override —
volání uvnitř `Base.Describe()` půjde vždycky na `Base`, nikdy na `User`. Kdo z Symfony čeká
šablonovou metodu, tady narazí.

Pokud jsou promotovaná jména na stejné hloubce dvě, nepromotuje se ani jedno a musíš psát
explicitní cestu. Chyba to je jen ve chvíli, kdy takové jméno použiješ.

**Embedding vs dědičnost** shrnuto:

| | PHP `extends` | Go embedding |
|---|---|---|
| Vztah | `User` *je* `Base` | `User` *má* `Base` |
| Přetížení metody | pozdní vazba, rodič volá potomka | žádná, `Base` o `User` neví |
| Přístup k rodiči | `parent::` | `u.Base` — obyčejné pole |
| Víc předků | ne | ano, vlož jich kolik chceš |
| Přístup k privátnímu | `protected` | řídí balíček, ne vztah typů |

Embedding používej, když chceš skutečně **znovupoužít implementaci** a rozšířit typ o pár
metod. Když jen sdílíš data, dej `Base` jako pojmenované pole (`base Base`) — je to
čitelnější a nevzniknou nechtěné promotované metody ve veřejném API.

### Embedding interface

Vložit jde i interface, a to jak do interface, tak do structu.

```go
type ReadCloser interface {
	io.Reader // interface v interface = skládání kontraktu
	io.Closer
}

type loggingStore struct {
	Store        // interface ve structu
	log *slog.Logger
}

func (s loggingStore) Save(u User) error { // přepíšeme jen tuhle metodu
	s.log.Info("saving", "id", u.ID)
	return s.Store.Save(u)
}
```

Druhý vzor je v Go idiomatický dekorátor: vložený interface zajistí, že typ splní celý
kontrakt, a ty přepíšeš jen metody, které chceš. Pozor na to, že nepřepsané metody spadnou
na `nil` panic, pokud vložený interface nenaplníš.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Metoda mění stav a má value receiver | vypadá to, že to funguje | pointer receiver, jinak měníš kopii |
| Míchání value a pointer receiverů | každou metodu psal někdo jinak | jeden typ = jeden styl receiveru |
| `var i Iface = T{}` u pointer metod | method set se neintuituje | předej `&T{}` |
| `parent::` reflex u embeddingu | `Base` vypadá jako rodič | `Base` o vnějším typu neví, není override |
| Poziční literál u velkého structu | méně psaní | pojmenovaná pole, jinak rozbiješ kód přidáním pole |
| `==` na structu se slicem | struct vypadá porovnatelně | metoda `Equal`, nebo v testu `reflect.DeepEqual` |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.
Typy jsou předpřipravené, dopisuješ jen metody.

### A — rozcvička (~10 min)

Typ `Point struct{ X, Y int }`:

- `func (p Point) Add(q Point) Point` — vrátí **nový** bod se součtem složek. Přijímače se
  nesmí dotknout.
- `func (p Point) String() string` — formát přesně `"(1,2)"`, bez mezer. Pro
  `Point{-3, 40}` tedy `"(-3,40)"`. Tím implementuješ `fmt.Stringer`, test to ověří přes
  `fmt.Sprintf("%v", p)`.

Test taky porovnává body přes `==`. Rozmysli si, proč to jde.

### B — jádro (~35 min)

Typ `Counter struct{ n int }` má **neexportované** pole, takže se dá měnit jen metodami:

- `func (c *Counter) Inc()` — o jedna nahoru.
- `func (c *Counter) Add(n int)` — přičte `n`, záporné `n` počítadlo snižuje.
- `func (c Counter) Value() int` — aktuální hodnota, **value receiver**.

Zero value musí být použitelná: `var c Counter; c.Inc()` dá `1`.

Testy ověřují dvě věci navíc, které jsou pointou úkolu. Za prvé, že předání `Counter`
do funkce hodnotou vytvoří kopii a `Inc()` na ní originál nezmění. Za druhé, přes
type assertion na `any`, že `Counter` **nemá** `Inc` v method setu, zatímco `*Counter` ano.
Kdybys `Inc` napsal s value receiverem, obojí selže.

### C — rozšíření (~25 min)

Embedding ve třech úrovních. Typy jsou hotové: `Base{ID string}`, `User{Base; Name string}`,
`Admin{User; Level int}`.

- `func (b Base) Describe() string` → `"base:<ID>"`, pro `Base{ID: "b1"}` tedy `"base:b1"`.
- `func (u User) Describe() string` → `"user:<Name> (<popis base>)"`, tedy pro
  `User{Base{"u1"}, "Radek"}` přesně `"user:Radek (base:u1)"`. Popis základu **musíš získat
  voláním** `u.Base.Describe()`, ne slepením řetězce ručně.
- `func (a Admin) Tag() string` → `"admin:<ID>/<Level>"`, tedy `"admin:a1/9"`. `ID` ber
  přes promotované pole (`a.ID`), ne přes `a.User.Base.ID`. Pro `Admin{}` vyjde
  `"admin:/0"`. Na převod čísla použij `strconv.Itoa`.

`Admin` schválně `Describe` nedefinuje — test ověří, že se promotuje z `User` a že `Admin`
díky tomu splní interface s metodou `Describe()`.

```bash
make lesson L=05
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `05`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=05` prochází
- [ ] Umíš vyjmenovat čtyři kritéria pro volbu pointer receiveru
- [ ] Umíš vysvětlit, proč `var i Incrementer = Counter{}` neprojde kompilací
- [ ] Umíš vysvětlit, proč embedding není dědičnost, na příkladu shadowingu
- [ ] Umíš říct, které typy dělají struct neporovnatelným
- [ ] Umíš najít ve stdlib typ, který používá embedding (nápověda: `bufio`)

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

## Další čtení

1. [Effective Go — Embedding](https://go.dev/doc/effective_go#embedding)
2. [Go Code Review Comments — Receiver Type](https://go.dev/wiki/CodeReviewComments#receiver-type)
3. [Go spec — Method sets](https://go.dev/ref/spec#Method_sets)
4. [Tour of Go — Methods and pointer indirection](https://go.dev/tour/methods/6)
