# Lekce 33 — Porty a adaptéry, interface u konzumenta

> **Čas:** ~35 min · **Fáze:** 4 — Architektura v Go · **AI režim:** `BOILERPLATE OK`

## Co budeš umět

- Navrhnout port jako malý interface definovaný tam, kde se používá, a poznat, kdy je
  příliš velký.
- Rozlišit driving a driven porty a určit, kterým směrem mají mířit závislosti.
- Nahradit mockovací framework fake adaptérem a otestovat i chybovou cestu.
- Vysvětlit, proč „accept interfaces, return structs" a „interface u konzumenta" jsou
  totéž pravidlo z jiné strany.
- Sestavit graf závislostí v jedné funkci a obhájit, proč Go nepotřebuje DI kontejner.

## Teorie

### Port je interface u konzumenta

Hexagonální architektura (porty a adaptéry) říká jedinou podstatnou věc: aplikační
jádro definuje rozhraní se světem a světy se k nim připojují. V Go to není framework,
je to důsledek toho, jak fungují interfacy.

```go
package ordering

// port — malý, doménovým jazykem, u konzumenta
type Clock interface{ Now() time.Time }
```

Balíček `ordering` teď nezná `time.Now`, ale ví, že čas odněkud dostane. Adaptér vedle:

```go
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }
```

Nikde není `implements`. Kompilátor to ověří až v místě, kde `SystemClock` dosadíš do
parametru typu `Clock`. Právě proto může adaptér vzniknout v úplně jiném balíčku, dokonce
v cizím modulu, který o tvém portu nikdy neslyšel.

### Driving vs driven

- **Driving (primární)** port pohání aplikaci zvenku: HTTP handler, CLI, konzument fronty.
  Zpravidla to není interface, ale prostě veřejné API tvé doménové služby.
- **Driven (sekundární)** port aplikace používá směrem ven: úložiště, hodiny, notifikace,
  platební brána. Tohle jsou ty interfacy, o kterých je celá lekce.

Závislost vždy míří **dovnitř**: adaptér zná doménu, doména nezná adaptér. Test toho, že
to máš správně, je jednoduchý — ve svém doménovém balíčku nesmíš najít import
`net/http`, `database/sql` ani `encoding/json`.

### Jak velký má port být

Ideál je jedna až tři metody. Čím menší port, tím víc typů ho splní a tím snáz se
testuje.

```go
// ŠPATNĚ — port zrcadlí implementaci, konzument volá dvě metody ze čtrnácti
type UserRepositoryInterface interface {
	Save(u User) error
	Delete(id string) error
	FindAll() ([]User, error)
	FindByEmail(e string) (User, error)
	CountActive() (int, error)
	// …
}

// DOBŘE — dva porty, každý pro jednoho konzumenta
type UserSaver interface{ Save(u User) error }
type UserFinder interface{ FindByEmail(e string) (User, error) }
```

Tlustý klient obě rozhraní splní najednou, aniž bys psal jediný řádek adaptéru navíc.
To je praktický důsledek implicitní implementace: **malé porty nic nestojí.**

Stejné pravidlo se v Code Review Comments říká jinak: *accept interfaces, return
structs*. Přijímáš interface, protože chceš mluvit o schopnosti; vracíš struct, protože
konzument si svůj interface vyrobí sám, až ho bude potřebovat.

### Fake místo mocku

Pro port se dvěma metodami je fake kratší než konfigurace mocku:

```go
type FailingStore struct {
	Err    error
	Orders map[string]Order
}

func (s FailingStore) Save(o Order) error { return s.Err }
func (s FailingStore) Get(id string) (Order, bool) {
	o, ok := s.Orders[id]
	return o, ok
}
```

Fake je skutečná implementace: nemá „expectations", nekontroluje pořadí volání
a nerozbije se, když službu refaktoruješ. Precedens najdeš přímo ve standardní
knihovně — `testing/fstest.MapFS` je fake souborového systému, `httptest.Server` je fake
protistrany.

Deterministický `Clock` a `IDGen` mají ještě jeden efekt: test může porovnat celou
strukturu přes `==`, protože v ní není nic náhodného.

### Kde se sestavuje graf závislostí

Na jediném místě, co nejblíž vstupnímu bodu:

```go
func Wire() (*OrderService, error) {
	return NewOrderService(NewMemoryStore(), SystemClock{}, RandomIDGen{})
}
```

Tohle je celý DI kontejner. Žádné `services.yaml`, žádné autowiring, žádná reflexe za
běhu. Když závislost chybí, program se nezkompiluje — v Symfony bys to zjistil až při
`cache:warmup`, nebo hůř, v produkci. Cena je, že wiring píšeš ručně; u desítek služeb
to je pár desítek řádků a pořád je to čitelnější než graf, který nikdo nevidí.

### Chyba adaptéru na hranici domény

Doména nemá pouštět chybu adaptéru ven tak, jak přišla — konzument by pak musel znát
`pq.Error`. Obal ji vlastním sentinelem a příčinu zachovej:

```go
if err := s.store.Save(o); err != nil {
	return Order{}, fmt.Errorf("place %s: %w: %w", o.ID, ErrStore, err)
}
```

Dvojité `%w` (od Go 1.20) dá volajícímu obojí: `errors.Is(err, ErrStore)` pro rozhodnutí,
co dělat, i původní chybu pro log.

## Rozdíly proti PHP

V Symfony je interface deklarace u implementace a autowiring je slepí dohromady:

```php
interface OrderRepositoryInterface {
    public function save(Order $o): void;
    public function find(string $id): ?Order;
    public function findAll(): array;
    public function findByCustomer(string $c): array;
    public function countPending(): int;
    // …a dalších devět metod, protože „interface má popsat repozitář"
}

final class DoctrineOrderRepository implements OrderRepositoryInterface { /* … */ }
```

Interface je tu **popis implementace**. Vzniká zrcadlením třídy, roste s ní a každý
konzument dostane všech čtrnáct metod, i když volá jednu.

V Go je to obráceně. Interface deklaruje **konzument** a popisuje jen to, co sám volá:

```go
// v balíčku ordering — u konzumenta
type OrderStore interface {
	Save(o Order) error
	Get(id string) (Order, bool)
}
```

Implementace o tom interface nikdy nemusí vědět; splní ho tím, že má správné metody.
Návyk k opuštění: **nepiš `XxxInterface` vedle `Xxx`.** Když má interface stejné metody
jako jediná implementace, není to abstrakce, jen zdvojený kód.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `UserRepositoryInterface` se 14 metodami | zrcadlení Symfony service | port podle toho, co konzument volá |
| Interface v balíčku implementace | zvyk deklarovat `implements` | interface patří ke konzumentovi |
| Mockovací framework na port se dvěma metodami | reflex z PHPUnit | ruční fake, deset řádků |
| `time.Now()` uvnitř domény | „to přece není závislost" | port `Clock` a fake v testu |
| Doména importuje `database/sql` | vrstvení podle Doctrine | závislost míří dovnitř, ne ven |
| Vracení interfacu z konstruktoru | „ať je to abstraktní" | accept interfaces, return structs |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 33`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od
jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `NewOrder` (volá `time.Now()` místo portu `Clock`)

```bash
make lesson L=33 PART=1
```

Pak **`/go-deep-review 33 easy`**.

### Střední

Implementuj: `NewMemoryStore`, `Save`, `Get`

```bash
make lesson L=33 PART=2
```

Pak **`/go-deep-review 33 medium`**.

### Obtížný

Doplň: `Place` (obal chyby adaptéru přes `ErrStore`)

```bash
make lesson L=33 PART=3
```

Pak **`/go-deep-review 33 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 33 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=33` (+ `make race L=33`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč port patří ke konzumentovi, ne k implementaci
- [ ] Umíš rozlišit driving a driven port a určit směr závislostí
- [ ] Umíš říct, proč je `UserRepositoryInterface` se 14 metodami anti-vzor
- [ ] Umíš napsat fake adaptér rychleji, než bys nakonfiguroval mock
- [ ] Umíš vysvětlit, proč Go nepotřebuje DI kontejner a co je jeho náhrada

## AI režim

`BOILERPLATE OK` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md). Fake adaptéry
a wiring si nech vygenerovat. Tvar portů navrhuj sám: AI má silný sklon vyrobit
`XxxRepositoryInterface` se všemi metodami implementace, což je přesně to, čemu se
tahle lekce vyhýbá.

## Další čtení

1. [Go Code Review Comments — Interfaces](https://go.dev/wiki/CodeReviewComments#interfaces)
2. [Effective Go — Interfaces and methods](https://go.dev/doc/effective_go#interfaces_and_types)
3. [Go blog — Errors are values](https://go.dev/blog/errors-are-values)
4. [pkg.go.dev — testing/fstest](https://pkg.go.dev/testing/fstest) — fake ve standardní knihovně
