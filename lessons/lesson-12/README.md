# Lekce 12 — Interfaces I: implicitní implementace

> **Čas:** ~45 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Vysvětlit, proč v Go není `implements`, a co z toho plyne pro návrh závislostí.
- Rozhodnout, kde interface deklarovat — u konzumenta, ne u implementace.
- Odhalit past „non-nil interface s nil pointerem uvnitř" a vysvětlit, proč vzniká.
- Použít type assertion s comma-ok a type switch, aniž bys z nich udělal `instanceof` kaskádu.

## Teorie

### Interface je množina metod

Deklarace interfacu je jen výčet signatur. Žádná pole, žádné konstanty, žádná dědičnost
v Symfony smyslu:

```go
type Stringer interface {
	String() string
}
```

Typ splňuje interface, pokud má **všechny** jeho metody se shodnými signaturami. Kompilátor
to ověří v místě, kde hodnotu do interfacu přiřadíš — jinde nemá důvod.

Ideál v Go je interface o jedné metodě. Standardní knihovna to dodržuje důsledně:
`io.Reader`, `io.Writer`, `fmt.Stringer`, `sort.Interface` (výjimka se třemi), `error`.
Čím méně metod, tím víc typů ho splní a tím snáz ho v testu podvrhneš.

Když se přistihneš u interfacu s osmi metodami, který kopíruje jednu konkrétní třídu,
napsal jsi Symfony službu, ne Go interface. Rozděl ho podle toho, co jednotliví volající
skutečně potřebují.

### Deklaruj interface u konzumenta

Ze dvou balíčků — jeden umí posílat maily, druhý objednávky — patří interface do toho
druhého:

```go
package order

// Notifier je to, co balíček order potřebuje. Balíček mailer o něm neví.
type Notifier interface {
	Notify(msg string) error
}

type Service struct {
	notify Notifier
}

func NewService(n Notifier) *Service { return &Service{notify: n} }
```

Balíček `mailer` vrací konkrétní `*mailer.Client`. Ten `order.Notifier` splňuje sám od sebe.
Výhody: `order` nezávisí na `mailer` (žádný import, tedy ani riziko cyklu — viz lekce 11),
interface má přesně ty metody, které `order` volá, a v testu stačí desetiřádkový fake.

Odtud pochází pravidlo **„accept interfaces, return structs"**: funkce bere jako parametr
interface (aby přijala cokoli vhodného), ale vrací konkrétní typ (aby volající viděl
všechno, co typ umí). Detailně v lekci 20.

### Interface hodnota je dvojice (typ, hodnota)

Tohle je klíč k celé lekci. Proměnná typu interface není ukazatel na objekt. Je to
**dvojice**: dynamický typ a dynamická hodnota.

```go
var s Shape          // (nil, nil)          → s == nil
s = Rect{W: 2, H: 3} // (Rect, {2,3})       → s != nil
```

Interface se rovná `nil`, jen když jsou **obě** složky prázdné. A právě tady vzniká
nejslavnější past jazyka:

```go
type MyErr struct{}

func (e *MyErr) Area() float64 { return 0 }

func ReturnsNilPointer() Shape {
	var p *MyErr // p == nil
	return p     // ale výsledek je (*MyErr, nil) — a to není nil interface!
}

s := ReturnsNilPointer()
fmt.Println(s == nil) // false — překvapení
```

Ukazatel uvnitř je opravdu `nil`, ale dvojice má vyplněný typ, takže se `nil` nerovná.
Nejčastěji to bolí u `error`:

```go
// ŠPATNĚ — funkce nikdy nevrátí nil error
func do() error {
	var err *MyError // typovaný nil
	…
	return err // volající dostane non-nil error i při úspěchu
}

// SPRÁVNĚ — pracuj s proměnnou typu error
func do() error {
	if problem {
		return &MyError{}
	}
	return nil
}
```

Pravidlo: **nikdy nedeklaruj lokální proměnnou konkrétního pointer typu, kterou pak vrátíš
jako interface.** Vrať buď `nil` literál, nebo konkrétní nenulovou hodnotu.

Mimochodem, volat metodu na nil pointeru je legální, dokud metoda nesáhne na pole
receiveru. `s.Area()` výše proběhne a vrátí 0. Nil pointer receiver je dokonce občas
užitečný — `(*Tree).Insert` na prázdném stromu.

### Type assertion, comma-ok a type switch

Když potřebuješ z interfacu zpátky konkrétní typ, použij type assertion. Vždy ve verzi
s `ok`, pokud si nejsi jistý:

```go
s := shapes[0]

r := s.(Rect)     // panikuje, když s není Rect
r, ok := s.(Rect) // ok == false a r == Rect{}, žádná panika
```

Assertion funguje i na jiný interface — tak se ptáš, jestli hodnota umí něco navíc:

```go
if sc, ok := w.(io.StringWriter); ok {
	sc.WriteString("rychlejší cesta") // bez konverze na []byte
}
```

Když variant přibude, použij type switch:

```go
switch x := v.(type) {
case nil:
	return "nil"
case int:
	return fmt.Sprintf("int:%d", x)
case []int:
	return fmt.Sprintf("[]int:len=%d", len(x))
default:
	return fmt.Sprintf("other:%T", x)
}
```

Uvnitř každé větve má `x` ten konkrétní typ. Větev `case nil` chytá prázdnou interface
hodnotu — pozor, `ReturnsNilPointer()` se do ní **netrefí**, protože její typ je `*MyErr`.

Varování: type switch přes vlastní typy je v Go většinou signál, že ti chybí metoda.
Když píšeš `switch x := shape.(type) { case Rect: … case Circle: … }`, měl jsi místo toho
přidat metodu do interfacu. Legitimní použití je u `any` na hranici (JSON, logování,
formátování) a u `errors.As` (lekce 14).

### Prázdný interface a `any`

`interface{}` nemá žádnou metodu, takže ho splní úplně každý typ. Od Go 1.18 se píše
`any`, což je pouhý alias — jsou to dvě jména pro totéž.

`any` znamená „nevím, co to je", takže s tím nejde nic dělat, dokud to type switchem
nerozbalíš. Před generikami (lekce 15) to byla jediná cesta k univerzálním kontejnerům.
Dnes je `any` v API většinou známka, že jsi rezignoval na typový systém. Zůstává na místě
tam, kde skutečně nevíš: `fmt.Println(a ...any)`, `json.Unmarshal(data, v any)`.

### Method set a pointer receiver

Zádrhel, který stojí za zapamatování hned:

- Má-li typ metodu s **hodnotovým** receiverem `func (r Rect) Area()`, patří ta metoda
  do method setu `Rect` i `*Rect`.
- Má-li metodu s **pointer** receiverem `func (r *Recorder) Notify()`, patří jen do
  method setu `*Recorder`.

```go
var n Notifier = &Recorder{} // OK
var n Notifier = Recorder{}  // chyba: Recorder does not implement Notifier
                             // (method Notify has pointer receiver)
```

Důvod je prostý: `Recorder{}` v interfacu je kopie a metoda, která mění stav, by měnila
kopii. Go tenhle omyl nedovolí. Chybová hláška `method has pointer receiver` je jedna
z nejčastějších, co jako začátečník uvidíš — teď víš, že řešení je `&`.

## Rozdíly proti PHP

V PHP je implementace interfacu **deklarace**. Třída se k němu musí přihlásit jménem,
a když interface vlastní cizí knihovna, musíš ji do svého kódu natáhnout:

```php
interface Notifier { public function notify(string $msg): void; }

final class SlackNotifier implements Notifier   // explicitní přihlášení
{
    public function notify(string $msg): void { /* … */ }
}
```

V Go je implementace **pozorování**. Typ splňuje interface v okamžiku, kdy má správnou
sadu metod. Nikdo nikam nic nepíše:

```go
type Notifier interface {
	Notify(msg string) error
}

type SlackNotifier struct{ token string }

func (s SlackNotifier) Notify(msg string) error { return nil }
// hotovo — SlackNotifier splňuje Notifier, aniž by o něm věděl
```

Co se mění v uvažování: interface přestává být kontrakt, ke kterému se implementace
hlásí *shora*, a stává se **požadavkem, který si klade volající**. Proto se interface
v Go deklaruje v balíčku, který ho konzumuje, je co nejmenší a implementace o něm často
vůbec neví. Přesně naopak než Symfony, kde interface a jeho implementace bydlí vedle sebe.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Interface se sedmi metodami k jedné implementaci | reflex „interface ke každé službě" ze Symfony | interface o 1–2 metodách podle potřeby volajícího |
| Interface v balíčku implementace | PSR návyk „interface vedle třídy" | deklaruj ho v balíčku, který ho volá |
| `return err` s typovaným nil pointerem | interface hodnota vypadá jako pointer | vrať `nil` literál nebo konkrétní hodnotu |
| `x := v.(Foo)` bez `ok` | zvyk na `instanceof` a měkké přetypování | vždy comma-ok, pokud typ nemáš zaručený |
| Type switch místo metody | přenos `instanceof` kaskády z PHP | přidej metodu do interfacu |
| `any` v signatuře vlastní funkce | snaha o univerzálnost | konkrétní typ nebo generika (lekce 15) |
| `var n Notifier = Recorder{}` | přehlédnutý pointer receiver | `&Recorder{}`, nebo změň receiver na hodnotu |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 12`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od jednodušších
ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `TotalArea` (záměrně vadný — nezkontroluje nil prvky v `[]Shape`)

```bash
make lesson L=12 PART=1
```

Pak **`/go-deep-review 12 easy`**.

### Střední

Implementuj: `Describe`, `Notify`, `Messages`

```bash
make lesson L=12 PART=2
```

Pak **`/go-deep-review 12 medium`**.

### Obtížný

Doplň: `ReturnsNilPointer`, `IsNilInterface` (typed nil v interface)

```bash
make lesson L=12 PART=3
```

Pak **`/go-deep-review 12 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 12 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=12` (+ `make race L=12`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč `ReturnsNilPointer() != nil`
- [ ] Umíš vysvětlit, proč se interface deklaruje u konzumenta
- [ ] Umíš vysvětlit rozdíl mezi method setem `T` a `*T`
- [ ] Umíš říct, kdy je type switch v pořádku a kdy je to schovaná chybějící metoda
- [ ] Umíš napsat fake implementaci interfacu bez mockovací knihovny

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Mentor, kvíz i review (dialog) jsou vždy OK; v tomto režimu AI nesmí psát kód cvičení.
## Další čtení

1. [Effective Go — Interfaces and methods](https://go.dev/doc/effective_go#interfaces_and_types)
2. [Go blog — Go interfaces and the nil comparison](https://go.dev/doc/faq#nil_error)
3. [Go Code Review Comments — Interfaces](https://go.dev/wiki/CodeReviewComments#interfaces)
4. [Tour of Go — Type switches](https://go.dev/tour/methods/16)
