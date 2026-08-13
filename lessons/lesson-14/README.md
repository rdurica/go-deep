# Lekce 14 — Errors: hodnoty, wrapping, Is a As

> **Čas:** ~45 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Rozhodnout mezi sentinel chybou, vlastním typem chyby a prostým `errors.New`.
- Obalit chybu tak, aby si zachovala kontext i strojovou rozpoznatelnost, a vědět, kdy neobalovat.
- Použít `errors.Is`, `errors.As` a `errors.Join` a vysvětlit, čím se liší.
- Napsat texty chyb podle konvence a poskládat je do věty, která se dá číst v logu.

## Teorie

### `error` je obyčejný interface

```go
type error interface {
	Error() string
}
```

Nic víc. Cokoli s metodou `Error() string` je chyba — a díky implicitní implementaci
(lekce 12) se k tomu nikdo nemusí hlásit. Dva nejjednodušší způsoby, jak chybu vyrobit:

```go
errors.New("division by zero")                 // statický text
fmt.Errorf("load user %s: %w", id, err)        // text s kontextem a obalením
```

Protože je chyba hodnota, dá se uložit do proměnné, poslat kanálem, porovnat, seřadit
nebo posbírat do slice. Sekce „Errors are values" z Go blogu na tom staví celý vzor,
kdy si typ pamatuje první chybu a další zápisy tiše ignoruje.

### Konvence textů chyb

Text chyby je **fragment věty**, ne věta. Chová se jako článek řetězu, na který se
navěsí další kontext:

- malé počáteční písmeno (`"division by zero"`, ne `"Division by zero"`),
- bez tečky na konci,
- bez prefixu `"error:"`, `"failed to"` a podobných vycpávek,
- bez velkých písmen s výjimkou vlastních jmen a zkratek (`"parse URL: ..."`).

Důvod je praktický. Když se chyby obalují, vznikne z nich souvětí oddělené dvojtečkami:

```
load user u9: query users: dial tcp 127.0.0.1:5432: connection refused
```

Kdyby každá vrstva začínala velkým písmenem a končila tečkou, byla by to nečitelná kaše.
Kontext přidávej **shora dolů** a nikdy neopakuj to, co už řekla vrstva pod tebou —
`fmt.Errorf("failed to load user: %w", err)` je horší než
`fmt.Errorf("load user %s: %w", id, err)`, protože „failed" je z existence chyby zřejmé,
zatímco `id` je informace navíc.

### Sentinel chyby a `errors.Is`

Sentinel je předem vytvořená chybová hodnota na úrovni balíčku, na kterou se dá ptát
podle identity:

```go
var ErrDivideByZero = errors.New("division by zero")

func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("divide %d: %w", a, ErrDivideByZero)
	}
	return a / b, nil
}
```

Volající se ptá `errors.Is`, ne `==`:

```go
if errors.Is(err, ErrDivideByZero) { … }
```

Rozdíl je zásadní. `==` porovná jen tu jednu hodnotu; `errors.Is` prochází **celý řetěz**
obalených chyb a porovnává každý článek. Jakmile chybu kdekoli obalíš, `==` přestane
fungovat — proto se `==` na chyby v moderním Go nepoužívá vůbec.

Sentinely znáš ze stdlib: `io.EOF`, `sql.ErrNoRows`, `os.ErrNotExist`, `context.Canceled`.
Konvence: jméno začíná `Err`, proměnná je `var` (ne `const`, `errors.New` není konstantní
výraz) a je exportovaná, jen když se na ni má někdo ptát.

Sentinel se hodí, když chyba nenese žádná data. Jakmile chceš předat detail — které pole,
které ID, kolikátý řádek — potřebuješ typ.

### Vlastní typy chyb a `errors.As`

```go
type NotFoundError struct {
	ID string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("id %s not found", e.ID)
}
```

Metoda `Error()` má skoro vždycky **pointer receiver**. Důvody: chyba se pak porovnává
podle identity, ne podle obsahu, a `errors.As` bude hledat `*NotFoundError`. Vracej ji
tedy jako `&NotFoundError{ID: id}`.

Ptát se na typ se dá přes `errors.As`, což je type assertion procházející řetěz:

```go
var nf *NotFoundError
if errors.As(err, &nf) {
	log.Printf("chybí záznam %s", nf.ID) // máš přístup k datům
}
```

Druhý parametr musí být **ukazatel na** proměnnou cílového typu — proto `&nf`, kde `nf`
je už samo `*NotFoundError`. Nejčastější začátečnická chyba je předat `nf` místo `&nf`;
`errors.As` v tom případě panikuje.

Pravidlo: `errors.Is` pro „je to tahle konkrétní chyba", `errors.As` pro „je to chyba
tohohle druhu a chci z ní data".

Od Go 1.26 můžeš místo `errors.As` + deklarace proměnné použít generické
`errors.AsType[*NotFoundError](err)`, které vrátí `(T, bool)`. Stejná sémantika,
méně boilerplate — v novém kódu je to často čitelnější; `errors.As` zůstává
všude ve starším kódu a v API, které bere `any`.

Pozor, tady se vyplatí zapomenout na PHP: **nedělej hierarchii**. V Go nepotřebuješ
společného předka, protože se neptáš „je to potomek DomainException", ale „dá se z toho
vytáhnout `*ValidationError`". Když opravdu potřebuješ kategorii, přidej typu metodu
(`func (e *NotFoundError) Temporary() bool`) a ptej se na malý interface.

### Wrapping přes `%w`

`fmt.Errorf` s `%w` vyrobí chybu, která si pamatuje tu původní:

```go
err := fmt.Errorf("load user %s: %w", id, cause)

err.Error()          // "load user u9: id u9 not found"
errors.Unwrap(err)   // cause
errors.Is(err, …)    // projde i cause
errors.As(err, …)    // projde i cause
```

Rozdíl proti `%v` je klíčový: `%v` vloží jen **text** a řetěz přeruší. Chyba pak vypadá
stejně, ale `errors.Is` i `errors.As` na ní selžou. Je to nejčastější tichá chyba
v Go kódu — všechno se tváří v pořádku, dokud někdo nepotřebuje chybu rozpoznat.

Od Go 1.20 smí jeden `fmt.Errorf` obsahovat `%w` **vícekrát**:

```go
err := fmt.Errorf("import selhal: %w, %w", errParse, errStore)
```

Řetěz se tím větví do stromu a `errors.Is` prohledá obě větve.

**Kdy neobalovat.** Wrapping vystaví vnitřní chybu jako součást veřejného API tvého
balíčku. Jakmile ji někdo začne testovat přes `errors.Is`, nemůžeš implementaci vyměnit,
aniž bys mu to rozbil. Nechceš-li prozradit, že uvnitř je `sql.ErrNoRows` nebo konkrétní
HTTP klient, přelož chybu na vlastní a řetěz **schválně** přeruš:

```go
if errors.Is(err, sql.ErrNoRows) {
	return &NotFoundError{ID: id} // bez %w — detail úložiště ven neuteče
}
return fmt.Errorf("query user %s: %w", id, err)
```

Stejné pravidlo platí na hranici, kde by vnitřní text mohl uniknout uživateli
(cesty na disku, jména tabulek, connection stringy).

### `errors.Join` a `Unwrap`

Chceš-li vrátit **víc chyb najednou** — typicky validaci, kde chceš uživateli ukázat
všechna špatná pole naráz — je od Go 1.20 k dispozici `errors.Join`:

```go
var errs []error
if name == "" {
	errs = append(errs, &ValidationError{Field: "name", Reason: "must not be empty"})
}
if !strings.Contains(email, "@") {
	errs = append(errs, &ValidationError{Field: "email", Reason: "must contain @"})
}
return errors.Join(errs...) // nad prázdným seznamem vrací nil
```

Výsledek má `Error()` složené z jednotlivých textů oddělených `\n` a implementuje
`Unwrap() []error` — všimni si množného čísla. Řetěz chyb tedy obecně **není řetěz,
ale strom**:

| Metoda | Kdo ji má | Co znamená |
|--------|-----------|------------|
| `Unwrap() error` | `fmt.Errorf` s jedním `%w` | jeden rodič |
| `Unwrap() []error` | `errors.Join`, `fmt.Errorf` s více `%w` | více větví |

`errors.Is` i `errors.As` obě varianty prohledávají samy. Ale pozor: `errors.As` vrátí
jen **první** nález. Chceš-li posbírat všechny (třeba všechna vadná pole), musíš strom
projít ručně přes type assertion na `interface{ Unwrap() []error }`. Přesně to dělá `ValidateUser`.

## Rozdíly proti PHP

V PHP je chyba **výjimka** — ovládací tok, který se sám probublá nahoru, dokud ho někdo
nechytí. Kdo ji nechytí, o ní nemusí vědět:

```php
final class UserNotFound extends \RuntimeException {}

function loadUser(string $id): User
{
    $row = $this->store->find($id);
    if ($row === null) {
        throw new UserNotFound("User $id not found");   // odsud dál nic
    }
    return User::fromRow($row);
}
```

V Go je chyba **hodnota**, kterou funkce vrací jako druhý výsledek. Nikam se sama
nepropaguje — buď ji zpracuješ, nebo ji explicitně pošleš dál:

```go
func LoadUser(id string) (User, error) {
	row, err := store.Find(id)
	if err != nil {
		return User{}, fmt.Errorf("load user %s: %w", id, err)
	}
	return userFromRow(row), nil
}
```

Změna v uvažování: `if err != nil { return ... }` **není boilerplate**. Je to
dokumentace toku chyb přímo v kódu. V PHP musíš pro každou funkci hledat v docblocku
nebo ve zdrojáku, co může vyhodit. V Go to vidíš v signatuře, a když chybu ignoruješ,
je to vidět v diffu. Zvyk, který je potřeba opustit: **nekopíruj hierarchii výjimek.**
`AppException → DomainException → UserException → UserNotFoundException` je v Go
zbytečná — chyby se nerozlišují podle rodokmenu, ale podle toho, na co se dá zeptat.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `%v` místo `%w` při obalování | vypadají stejně ve výstupu | `%w` všude, kde má chyba zůstat rozpoznatelná |
| `err == ErrFoo` | zvyk na porovnání instancí | `errors.Is(err, ErrFoo)` |
| Hierarchie typů chyb | kopie `Exception` stromu z PHP | ploché typy + `errors.As` |
| `errors.As(err, nf)` bez `&` | nezvyklé dvojité odkazování | `errors.As(err, &nf)`, druhý parametr je `**T` |
| `fmt.Errorf("failed to load: %w", err)` | zvyk psát chybu jako větu | fragment s daty: `"load user u9: %w"` |
| `panic()` na doménovou chybu | zvyk na `throw` | vrať `error`, panika je jen pro chyby programátora |
| Obalování cizí chyby na hranici balíčku | wrapping jako automatismus | přelož na vlastní typ, řetěz schválně přeruš |
| `_ = err` nebo ignorovaný druhý návrat | „to nemůže selhat" | zpracuj, nebo alespoň zaloguj s kontextem |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 14`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od jednodušších
ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `Divide` (záměrně vadný — vrací holý sentinel bez `%w`)

```bash
make lesson L=14 PART=1
```

Pak **`/go-deep-review 14 easy`**.

### Střední

Implementuj: `Error` na `ValidationError`, `ValidateUser`

```bash
make lesson L=14 PART=2
```

Pak **`/go-deep-review 14 medium`**.

### Obtížný

Doplň: `LoadUser`, `IsNotFound` (obalení chyby a `errors.As`)

```bash
make lesson L=14 PART=3
```

Pak **`/go-deep-review 14 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 14 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=14` (+ `make race L=14`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit rozdíl mezi `errors.Is` a `errors.As` a kdy použít který
- [ ] Umíš vysvětlit, co přesně rozbije `%v` místo `%w`
- [ ] Umíš uvést situaci, kdy se chyba obalovat **nemá**
- [ ] Umíš vyjmenovat čtyři pravidla pro text chyby
- [ ] Umíš vysvětlit, proč je `if err != nil` lepší dokumentace než `@throws`

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Mentor, kvíz i review (dialog) jsou vždy OK; v tomto režimu AI nesmí psát kód cvičení.
## Další čtení

1. [Go blog — Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
2. [Go blog — Errors are values](https://go.dev/blog/errors-are-values)
3. [pkg.go.dev — errors](https://pkg.go.dev/errors)
4. [Go Code Review Comments — Error Strings](https://go.dev/wiki/CodeReviewComments#error-strings)
