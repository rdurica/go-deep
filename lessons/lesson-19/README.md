# Lekce 19 — Jména, balíčky a struktura kódu

> **Čas:** ~30 min · **Fáze:** 2 — Idiomatický Go · **AI režim:** `JEN VYSVĚTLENÍ`

## Co budeš umět

- Pojmenovat typ, funkci, proměnnou i balíček tak, aby to Go reviewer nekomentoval.
- Vysvětlit, proč `buf`, `i`, `r`, `w` nejsou lenost, a kdy naopak krátké jméno škodí.
- Rozpoznat a odstranit koktání (`http.HTTPServer`, `user.NewUser`, `utils.StringUtils`).
- Přepsat zanořený kód na plochý s early returnem, aniž změníš chování.
- Rozdělit dlouhou funkci podle jednoho kritéria: jedna funkce = jedna úroveň abstrakce.

## Teorie

### Konvence pojmenování

Go má pro jména jen pár tvrdých pravidel, ale drží se jich celá stdlib.

- **MixedCaps, nikdy podtržítka.** `maxRetries`, ne `max_retries`. Platí i pro konstanty:
  `MaxRetries`, ne `MAX_RETRIES`. Křičící konstanty jsou jazyk C, ne Go.
- **Zkratky se píšou celé stejným případem.** `URL`, `ID`, `HTTP`, `API`, `JSON`, `SQL`.
  Takže `userID`, `ServeHTTP`, `parseURL`, `HTTPClient`. Nikdy `userId` ani `Url`.
  Když zkratka začíná neexportované jméno, jde celá malým: `urlPath`, `httpClient`.
- **Export je velké písmeno.** Rozhodnutí exportovat je designové rozhodnutí, ne
  formalita — exportovaný identifikátor je závazek zpětné kompatibility.
- **Getter bez prefixu `Get`.** Pole `name` má getter `Name()`. Setter `Get` protějšek
  nemá, ten se jmenuje `SetName(...)`. Prefix `Get` má smysl jen tam, kde jde o skutečné
  načtení odjinud (`http.Get`).
- **Interface s jednou metodou nese jméno metody plus `-er`.** `Reader`, `Writer`,
  `Closer`, `Stringer`, `Formatter`. Nikdy `IReader` ani `ReaderInterface`.

### Délka jména je funkcí velikosti scope

Tohle je pravidlo, které PHP vývojáře nejvíc překvapí. V Go platí nepřímá úměra: **čím
menší rozsah platnosti, tím kratší jméno**.

```go
// Idiomatické. i žije tři řádky, delší jméno by nic nepřidalo.
for i, line := range lines {
	if strings.HasPrefix(line, "#") {
		continue
	}
	out[i] = strings.TrimSpace(line)
}

// Neidiomatické — jméno je delší než jeho život.
for currentLineIndex, currentLineContent := range lines { /* ... */ }
```

Konvence, které uvidíš v celé stdlib: `i`, `j` pro indexy, `r` pro `io.Reader`,
`w` pro `io.Writer`, `b` pro `[]byte` nebo `strings.Builder`, `buf` pro buffer,
`err` pro chybu, `ctx` pro context, `s` pro string, `n` pro počet. Nejsou to zkratky
z lenosti, je to slovník — když je vidíš, okamžitě víš, co drží.

Obrácená strana: **parametr balíčkové exportované funkce a pole struktury mají mít jméno
popisné**, protože jejich scope je celý zbytek programu. `func Copy(dst, src io.Writer)`
je v pořádku (dva řádky, doc comment vysvětluje), ale
`type Config struct { t int }` v pořádku není.

### Jména balíčků

Jméno balíčku je krátké, jednoslovné, malými písmeny, bez podtržítek a bez množného
čísla: `http`, `json`, `time`, `sort`, `user`, `billing`. Ne `userManagement`,
ne `user_service`, ne `models`.

Zápach číslo jedna je balíček bez jasné hranice:

```go
// ŠPATNĚ — utils, common, helpers, shared, misc, base
package utils

func StringUtilsTrim(s string) string { /* ... */ }
```

Takový balíček roste, dokud na něm nezávisí všechno a on nezávisí na všem. Test je
jednoduchý: **umíš jednou větou popsat, co balíček dělá, bez slova „a"?** Když ne, rozděl.
Funkci pro práci s textem zákazníků dej do `customer`, ne do `utils`.

### Vyhýbání se koktání

Jméno se čte s prefixem balíčku. Cokoli, co prefix zopakuje, je šum:

```go
// ŠPATNĚ                       // SPRÁVNĚ
http.HTTPServer                 http.Server
user.NewUser()                  user.New()
list.ListItem                   list.Item
billing.BillingConfig           billing.Config
utils.StringUtilsTrim()         text.Trim()
```

Jedna výjimka, kterou nesmíš „opravit": prefix `Err` u sentinelových chyb je konvence
stdlib (`os.ErrNotExist`, `sql.ErrNoRows`), takže `user.ErrNotFound` je správně — nečte se
to jako koktání, ale jako „chyba not found".

### Řazení importů

`gofmt` seřadí importy uvnitř skupiny, ale skupiny nevyrábí. Standard jsou dvě až tři,
oddělené prázdným řádkem: stdlib, externí závislosti, vlastní modul.

```go
import (
	"errors"
	"fmt"
	"strings"

	"github.com/rdurica/go-deep/internal/billing"
)
```

Nikdy nepiš alias jen proto, že se ti jméno nelíbí. Alias má tři legitimní důvody: kolize
jmen, verzované cesty (`v2`) a testovací zrcadlení, jaké používá tenhle kurz.

### Deklarace, zanoření a délka funkce

Deklaruj proměnnou co nejblíž prvnímu použití, ne na začátku funkce. To je návyk ze staré
C a z PHP, kde `$result = null;` nahoře „aby to bylo přehledné". V Go tím jen prodlužuješ
scope a zvyšuješ šanci, že hodnota bude použitá dřív, než dostane smysl.

Nejsilnější strukturální pravidlo Go je **early return**: chybový a okrajový případ se
řeší hned a odchází, šťastná cesta zůstává vlevo bez odsazení.

```go
// ŠPATNĚ — happy path je zavrtaná ve čtvrté úrovni
func parse(raw string) (int, error) {
	if raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			if n > 0 {
				return n, nil
			} else {
				return 0, errors.New("must be positive")
			}
		} else {
			return 0, err
		}
	} else {
		return 0, errors.New("empty")
	}
}

// SPRÁVNĚ — čteš to shora dolů jako seznam podmínek
func parse(raw string) (int, error) {
	if raw == "" {
		return 0, errors.New("empty")
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, errors.New("must be positive")
	}
	return n, nil
}
```

Go nemá limit na délku funkce a stdlib obsahuje funkce na sto řádků. Užitečné pravidlo
ale zní: **funkce má dělat věci na jedné úrovni abstrakce**. Jakmile v jedné funkci
současně skládáš text, počítáš součty a formátuješ čísla, rozděl ji — ne kvůli počtu
řádků, ale proto, že jméno té funkce už nejde napsat bez slova „a".

## Rozdíly proti PHP

V Symfony se jméno třídy typicky skládá z role, vrstvy a technologie, protože třída žije
v globálním jmenném prostoru a název je jediné, co ji odlišuje:

```php
namespace App\Service\User;

final class UserRegistrationService
{
    public function getUserById(int $id): ?User { /* ... */ }
}

// použití
$service = new UserRegistrationService($repo);
$user = $service->getUserById(42);
```

V Go je nositelem kontextu **balíček**, ne jméno typu. Jméno se čte vždy i s prefixem
balíčku, takže cokoli, co je už v prefixu, se ve jméně opakovat nesmí:

```go
package user

// New vrací registrační službu. Volá se user.New(...), ne user.NewUserService(...).
func New(store Store) *Registration { /* ... */ }

// ByID hledá uživatele. Volá se r.ByID(42), ne r.GetUserByID(42).
func (r *Registration) ByID(id int) (User, error) { /* ... */ }
```

Návyk k opuštění: **nepiš jméno tak, aby dávalo smysl samo o sobě**. Piš ho tak, aby
dávalo smysl na místě volání. `user.Registration` je lepší než `user.UserRegistrationService`,
protože volající vždycky vidí obojí.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `getUserId()`, `userId` | camelCase reflex z PHP/JS | `UserID`, `userID` — zkratka celá velká |
| `UserServiceInterface` | Symfony konvence pro DI | `user.Store`, `-er` jméno, žádný suffix `Interface` |
| Balíček `utils` / `helpers` | vrstvení podle typu kódu, ne podle domény | balíček podle domény, jméno bez „a" |
| `user.NewUser()` | jméno má dávat smysl samo o sobě | `user.New()` — čte se s prefixem |
| `MAX_RETRIES = 3` | konstanty se přece křičí | `MaxRetries = 3` (nebo `maxRetries`) |
| Všechny `var` na začátku funkce | návyk z C/PHP „přehledná hlavička" | deklaruj u prvního použití |
| `else` po `return` | symetrie větví je „hezčí" | early return, `else` zmiz |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 19`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od jednodušších
ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Oprav: `ParseUserID` (kód je záměrně vadný — else a hluboké zanoření)

```bash
make lesson L=19 PART=1
```

Pak **`/go-deep-review 19 easy`**.

### Střední

Implementuj: `ParseUserIDs`

```bash
make lesson L=19 PART=2
```

Pak **`/go-deep-review 19 medium`**.

### Obtížný

Doplň: `ProcessOrders` (early return, extrakce helperu)

```bash
make lesson L=19 PART=3
```

Pak **`/go-deep-review 19 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review 19 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=19` (+ `make race L=19`, pokud to lekce vyžaduje).

- [ ] Žádná funkce, kterou jsi napsal, nemá víc než 20 řádků těla
- [ ] `ParseUserID` nemá jediný `else` a maximální hloubku zanoření 1
- [ ] Umíš vysvětlit, proč je `user.New()` lepší než `user.NewUser()`
- [ ] Umíš vysvětlit, proč je `i` správné jméno a `currentLineIndex` špatné
- [ ] Umíš uvést tři důvody, proč je balíček `utils` problém
- [ ] Umíš zpaměti napsat `userID`, `ServeHTTP` a `parseURL` se správnými velikostmi písmen

## AI režim

`JEN VYSVĚTLENÍ` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Povolený prompt zní: *„Vysvětli, proč je tenhle Go kód neidiomatický z pohledu Code Review
Comments. Nenabízej přepsaný kód, jen body."* Refaktoring musíš napsat sám — o to v téhle
lekci jde.

## Další čtení

1. [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) — kanonický seznam, čti celý
2. [Effective Go — Names](https://go.dev/doc/effective_go#names)
3. [Go blog — Package names](https://go.dev/blog/package-names)
4. [Go Doc Comments](https://go.dev/doc/comment) — jak psát komentáře, které se čtou
