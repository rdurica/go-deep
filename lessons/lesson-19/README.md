# Lekce 19 — Jména, balíčky a struktura kódu

> **Čas:** ~85 min · **Fáze:** 2 — Idiomatický Go · **AI režim:** `JEN VYSVĚTLENÍ`

## Co budeš umět

- Pojmenovat typ, funkci, proměnnou i balíček tak, aby to Go reviewer nekomentoval.
- Vysvětlit, proč `buf`, `i`, `r`, `w` nejsou lenost, a kdy naopak krátké jméno škodí.
- Rozpoznat a odstranit koktání (`http.HTTPServer`, `user.NewUser`, `utils.StringUtils`).
- Přepsat zanořený kód na plochý s early returnem, aniž změníš chování.
- Rozdělit dlouhou funkci podle jednoho kritéria: jedna funkce = jedna úroveň abstrakce.

## PHP → Go most

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

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

Tohle je refaktoringová lekce: v zadání dostaneš **funkční, ale ošklivý kód**. Tvoje verze
musí mít **stejné chování** a idiomatický tvar. Ošklivou verzi si můžeš zkopírovat do
`exercise.go` a přepsat, nebo ji rovnou napsat správně — testy hlídají chování, ne postup.

### A — rozcvička (~10 min)

`ParseUserID(raw string) (int, error)` — ořízne bílé znaky a převede text na kladné číslo.

- prázdný nebo jen bílé znaky → `0, ErrEmptyID`
- text, který není celé číslo (`"abc"`, `"4a2"`, `"4.2"`, přetečení) → `0, ErrInvalidID`
- číslo ≤ 0 → `0, ErrNonPositiveID`
- jinak číslo a `nil`

Napiš ji s **maximální hloubkou zanoření 1** (žádný `if` uvnitř `if`, žádné `else`).
Tohle je ta ošklivá verze, kterou přepisuješ:

```go
// ŠPATNĚ — přepiš to
func ParseUserID(rawUserIdString string) (int, error) {
	if rawUserIdString != "" {
		trimmedUserIdString := strings.TrimSpace(rawUserIdString)
		if trimmedUserIdString != "" {
			parsedUserIdValue, parseError := strconv.Atoi(trimmedUserIdString)
			if parseError == nil {
				if parsedUserIdValue > 0 {
					return parsedUserIdValue, nil
				} else {
					return 0, ErrNonPositiveID
				}
			} else {
				return 0, ErrInvalidID
			}
		} else {
			return 0, ErrEmptyID
		}
	}
	return 0, ErrEmptyID
}
```

`ParseUserIDs(raw string) ([]int, error)` — rozdělí vstup podle `,` a každý díl pošle do
`ParseUserID`. Prázdný nebo jen bílý vstup dá prázdný výsledek a `nil` chybu. Při chybě
vrať `nil` a chybu obalenou indexem (0-based) ve tvaru `fmt.Errorf("id at index %d: %w", i, err)`,
aby `errors.Is` dál našel sentinel.

### B — jádro (~35 min)

`ProcessOrders(orders []Order) (Summary, error)` — projde objednávky a spočítá souhrn.

Chování krok za krokem pro každou objednávku (v pořadí vstupu):

1. Prázdné `ID` → chyba `fmt.Errorf("order at index %d: %w", i, ErrMissingOrderID)`.
2. `Status == "cancelled"` → objednávka se **celá přeskočí**, včetně validace položek.
3. `Status` jiný než `"paid"`, `"pending"` nebo `"cancelled"` → chyba obalující
   `ErrUnknownStatus`, jejíž text obsahuje `ID` objednávky.
4. Položka s `Quantity <= 0` → chyba obalující `ErrInvalidQuantity`, text obsahuje `SKU`.
5. Položka s `UnitCents < 0` → chyba obalující `ErrInvalidPrice`, text obsahuje `ID`.
6. Jinak: `OrderCount++`, `ItemCount += Quantity` každé položky,
   `TotalCents += Quantity * UnitCents` a `Customer` se přidá do `Customers`, pokud tam
   ještě není a není prázdný.

Při jakékoli chybě vrať **nulový** `Summary{}`. `Customers` seřaď vzestupně.
`nil` i prázdný vstup dají nulový `Summary` a `nil` chybu.

Ošklivá verze, kterou nahrazuješ, vypadá takhle (jen výřez):

```go
// ŠPATNĚ — přepiš to
func ProcessOrders(orderList []Order) (Summary, error) {
	var summaryResult Summary
	var customerNamesMap map[string]bool = make(map[string]bool)
	for orderIndex := 0; orderIndex < len(orderList); orderIndex++ {
		if orderList[orderIndex].ID != "" {
			if orderList[orderIndex].Status != "cancelled" {
				if orderList[orderIndex].Status == "paid" || orderList[orderIndex].Status == "pending" {
					for itemIndex := 0; itemIndex < len(orderList[orderIndex].Items); itemIndex++ {
						// ... další tři úrovně zanoření
					}
				}
			}
		}
	}
	return summaryResult, nil
}
```

Cílem je plochá verze: validace přes early `continue` / `return` a agregace položek
vytažená do neexportované pomocné funkce.

### C — rozšíření (~25 min)

`RenderInvoice(inv Invoice) string` musíš **rozdělit** na tři neexportované funkce:
`renderHeader`, `renderLines` a `renderTotal`. Stuby pro ně už v `exercise.go` jsou.

Přesný formát výstupu (oddělovač je 32 pomlček, každý řádek včetně posledního končí `\n`):

```text
INVOICE 2024-001
CUSTOMER: Acme s.r.o.
--------------------------------
Widget | 2 x 19.99 = 39.98
Gadget | 1 x 5.00 = 5.00
--------------------------------
TOTAL: 44.98
```

- Částka se formátuje z centů na dvě desetinná místa: `1999` → `19.99`, `1` → `0.01`,
  `100000` → `1000.00`, `0` → `0.00`.
- Řádek položky: `<Description> | <Quantity> x <jednotková cena> = <cena za řádek>`.
- Faktura bez položek má oba oddělovače hned za sebou a `TOTAL: 0.00`.
- Součet počítej v centech, nikdy ne ve `float64`.

Na skládání textu použij `strings.Builder`, ne `+=` v cyklu.

```bash
make lesson L=19
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `19`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=19` prochází
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
