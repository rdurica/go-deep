# Lekce 01 — Toolchain, moduly a workspace

> **Čas:** ~75 min · **Fáze:** 0 — Setup a mentální reset · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- Vysvětlit, co je Go modul, k čemu je `go.mod` a jak se liší od `composer.json`.
- Používat `go build`, `go run`, `go test`, `go vet`, `gofmt` a `go doc` bez googlení.
- Založit vlastní modul od nuly a pochopit, proč se importuje cestou, ne názvem balíčku.
- Číst chybové hlášky kompilátoru, které v PHP nemají obdobu (nepoužitá proměnná, nepoužitý import).

## PHP → Go most

V PHP je jednotka distribuce **balíček na Packagistu** a jednotka běhu **soubor, který
načte autoloader**. Kód spustíš tak, že na něj ukážeš webserverem nebo `php script.php`.
Není žádný build krok, chyby typu překlepu ve jménu funkce najdeš až za běhu.

V Go je jednotka distribuce **modul** (jeden `go.mod` = jeden modul, klidně se stovkami
balíčků) a jednotka kompilace **balíček** (adresář se soubory `package foo`). Než něco
poběží, projde to kompilátorem, který je nepříjemně přísný.

```php
// composer.json
{
  "name": "acme/billing",
  "autoload": { "psr-4": { "Acme\\Billing\\": "src/" } }
}
```

```go
// go.mod
module github.com/acme/billing

go 1.26
```

Rozdíl, který tě bude první týden štvát nejvíc: **cesta importu je odvozená z cesty
v modulu, ne z názvu balíčku**. `import "github.com/acme/billing/invoice"` naimportuje
adresář `invoice/`, ve kterém je `package invoice`. Žádný autoloader, žádné mapování
namespace → adresář v konfiguraci. Adresář *je* mapování.

## Teorie

### Modul, balíček, soubor

Tři úrovně, které se často pletou:

| Úroveň | Co to je | Kde se definuje |
|--------|----------|-----------------|
| Modul | jednotka verzování a závislostí | `go.mod` v kořeni |
| Balíček | jednotka kompilace a viditelnosti | adresář + `package X` v souborech |
| Soubor | jednotka organizace | `.go` soubor |

Balíček nemůže být rozprostřený přes dva adresáře a jeden adresář nemůže obsahovat dva
balíčky (výjimka: testovací balíček `foo_test`, k tomu se dostaneme v lekci 21).

Tenhle repozitář je **jeden modul** — podívej se do `go.mod` v kořeni:

```
module github.com/rdurica/go-deep

go 1.26
```

Každý adresář `lessons/lesson-NN/exercise` je samostatný balíček uvnitř toho modulu.
Proto se v testech importuje `github.com/rdurica/go-deep/lessons/lesson-01/exercise`.

### Jak se importují balíčky

V PHP napíšeš `use App\Service\Foo;` a autoloader soubor najde. V Go **importuješ cestu**
a kompilátor podle ní najde adresář. Standardní knihovna má krátké cesty — jen název
balíčku. Cizí kód a tvoje balíčky v modulu mají cestu začínající názvem modulu.

```go
package exercise

import (
	"fmt"     // standardní knihovna
	"strings" // taky stdlib
)

func demo() string {
	return fmt.Sprintf("ahoj, %s", strings.TrimSpace("  Go  "))
}
```

Pravidla, která tě budou štípat hned:

- Import patří **nad** funkce, typicky hned po `package …`. Jeden import můžeš napsat
  jako `import "fmt"`, víc jich dej do závorek — `gofmt` je stejně seřadí.
- Používáš je jako `balíček.Symbol`: `fmt.Sprintf`, `strings.TrimSpace`. Ne `use` alias
  ani globální import všech jmen (to v Go skoro nikdo nedělá).
- Nepoužitý import je **chyba kompilace**, ne varování. Přidáš `fmt`, nepoužiješ ho →
  program se nezkompiluje. Odebereš volání a import necháš → stejný osud.

Dva balíčky, které v téhle lekci skoro jistě otevřeš:

| Balíček | K čemu je | Typické volání |
|---------|-----------|----------------|
| `strings` | práce s textem (ořez, hledání, dělení…) | `strings.TrimSpace(s)` |
| `fmt` | formátovaný výstup a skládání řetězců | `fmt.Sprintf("x=%d", n)` |

`fmt` je zkratka z *format*. Umí tisknout (`Println`, `Printf`) i **vracet** hotový
string (`Sprintf`) — něco jako `sprintf` v C nebo skládání přes placeholdery místo
konkatenace `"a" + $x + "b"`. Placeholdery poznáš podle `%`: `%s` string, `%d` celé
číslo, `%v` „nějak rozumně vypiš hodnotu". Detailněji až později; pro úkol A a B ti
stačí `Sprintf` a tyhle tři.

Když nevíš, co funkce dělá: `go doc strings.TrimSpace` nebo `go doc fmt.Sprintf`.

### Příkazy, které budeš psát každý den

```bash
go run ./cmd/api        # zkompiluj a rovnou spusť (binárku zahodí)
go build ./...          # zkompiluj vše, binárky nech ležet
go test ./...           # spusť všechny testy v modulu
go test -run TestGreet . # jen testy, jejichž jméno matchuje regex
go test -v .            # ukecaně, včetně t.Log
go vet ./...            # statická analýza nad rámec kompilátoru
gofmt -l .              # vypiš soubory, které nejsou naformátované
gofmt -w .              # naformátuj je
go doc strings.Builder  # dokumentace do terminálu, bez prohlížeče
go env GOMODCACHE       # kde leží stažené závislosti
```

`./...` znamená „tento adresář a všechno pod ním“. Je to nejčastější argument, jaký
budeš psát.

Za povšimnutí stojí `go vet`. Není to linter v Symfony smyslu (styl), ale detektor
konstrukcí, které se kompilují, ale skoro jistě jsou chyba — třeba `Printf` se špatným
počtem argumentů nebo zapomenutý `sync.Mutex` kopírovaný hodnotou.

### Formátování není otázka názoru

V PHP se týmy dohadují o PSR-12 a nastavují si `php-cs-fixer`. V Go tahle debata
neexistuje: existuje `gofmt` a jeho výstup je definice správného formátování. Tabulátory,
umístění závorek, řazení importů — všechno je dané. Nikdo si nenastavuje dvě mezery.

Nastav si v editoru „format on save" **hned teď**. Ušetříš si stovky zbytečných řádků
v diffech.

### Kompilátor jako první reviewer

Dva případy, které PHP vývojáře překvapí, protože to nejsou varování, ale **chyby**:

```go
func broken() {
	x := 42          // declared and not used: x
	fmt.Println("a") // imported and not used: "fmt" — pokud fmt nepoužiješ jinde
}
```

Nepoužitá lokální proměnná a nepoužitý import znamenají, že se program nezkompiluje.
Zní to buzerantsky, ale v praxi to znamená, že v Go kódu prakticky neuvidíš mrtvé
importy a zapomenuté proměnné. Když potřebuješ hodnotu vědomě zahodit, použij `_`:

```go
_, err := fmt.Println("a") // počet zapsaných bajtů mě nezajímá
```

Pozor: nepoužitý **parametr** funkce ani nepoužitá **globální** proměnná chyba nejsou.
Pravidlo se týká jen lokálních proměnných a importů.

### Kde končí `go.mod` a začíná `go.sum`

`go.mod` říká, co chceš. `go.sum` obsahuje kryptografické hashe toho, co se skutečně
stáhlo — obdoba `composer.lock`, ale nezaznamenává jen verze, nýbrž otisky obsahu.
Oba soubory patří do gitu. Modul bez závislostí (jako tenhle kurz) `go.sum` mít nemusí.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| `go run main.go` u víc souborů | reflex z `php script.php` | `go run .` nebo `go run ./cmd/api` — pracuje se s balíčkem, ne souborem |
| Název balíčku ≠ název adresáře | zvyk na namespace | drž je stejné; kompilátor to dovolí, lidi ne |
| Ruční editace `go.sum` | zvyk na `composer.lock` merge | `go mod tidy` |
| Balíček `utils` / `helpers` | PHP `App\Util` | pojmenuj podle domény (lekce 19) |
| Ignorovaný `gofmt` | „naformátuju to potom" | format on save |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

### A — rozcvička (~10 min)

Implementuj `Greet(name string) string`. Pro prázdné jméno (i po odstranění bílých znaků)
vrať `"Hello, Go!"`, jinak `"Hello, <name>!"` s ořezanými bílými znaky.
Hodí se `strings` (ořez) a `fmt` (složení výsledku) — viz sekce o importech výše.

### B — jádro (~25 min)

Doplň dvě funkce v `exercise/exercise.go` (místo `panic("TODO: úkol B")`).
#### `SumAll(nums ...int) int`
Sečti libovolný počet celých čísel předaných jako argumenty.
- `...int` znamená variadickou funkci: voláš ji jako `SumAll(1, 2, 3)`, ne jako slice.
- Bez argumentů vrať `0`.
- Záporná čísla sečti normálně (např. `SumAll(-4, 2)` → `-2`).
Příklady:
- `SumAll()` → `0`
- `SumAll(5)` → `5`
- `SumAll(1, 2, 3)` → `6`
#### `Describe(vals []int) string`
Vrať textový popis slice čísel.
- Pokud je `vals` `nil` nebo prázdný (`len == 0`), vrať přesně `"empty"`.
- Jinak vrať řetězec **přesně** ve tvaru:
  `"count=<počet> sum=<součet> max=<maximum>"`
  (mezery a pořadí polí musí sedět — testy porovnávají celý string).
Příklady:
- `Describe(nil)` → `"empty"`
- `Describe([]int{})` → `"empty"`
- `Describe([]int{1, 2, 3})` → `"count=3 sum=6 max=3"`
- `Describe([]int{9, 2, 3})` → `"count=3 sum=14 max=9"`
- `Describe([]int{-5, -2})` → `"count=2 sum=-7 max=-2"`
Cíl části B: potkat se s variadickou funkcí, `range` cyklem a `fmt.Sprintf`
(import `"fmt"` — stejný blok jako u úkolu A).
Schválně si zkus nechat nepoužitou proměnnou a přečti si hlášku kompilátoru.

### C — rozšíření (~25 min, ověřuje se checklistem)

Tohle je hlavní část lekce. Bez testu, ale nepřeskakuj ji.

1. Mimo tento repozitář si založ vlastní modul:

```bash
mkdir -p ~/scratch/hello && cd ~/scratch/hello
go mod init example.com/hello
```

2. Vytvoř `main.go` s `package main` a funkcí `main`, která něco vypíše. Spusť `go run .`.
3. Přidej podadresář `greet/` s `package greet` a exportovanou funkcí. Zavolej ji z `main.go`
   plným importem `example.com/hello/greet`. Všimni si, že cesta začíná názvem modulu.
4. Zkus záměrně: přejmenovat funkci na malé písmeno a importovat ji. Přečti si chybu.
5. Spusť `go build .` a podívej se, jaká binárka vznikla a jak je velká.
6. Zavolej `go doc strings.TrimSpace` a `go doc -all strings | head -40`.

```bash
make lesson L=01
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

- [ ] `make lesson L=01` prochází
- [ ] Umíš vysvětlit rozdíl mezi modulem, balíčkem a souborem
- [ ] Umíš vysvětlit, proč se importuje cestou a ne názvem balíčku
- [ ] Umíš napsat `import` blok a zavolat `fmt.Sprintf` / `strings.TrimSpace`
- [ ] Víš, co udělá `go vet` navíc oproti kompilátoru
- [ ] Máš v editoru zapnutý `gofmt` on save
- [ ] Založil jsi vlastní modul mimo tento repozitář a spustil ho

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Ve fázi 0 a 1 si cvičení píšeš sám. AI smí odpovídat na koncepční otázky
(„proč je nepoužitý import chyba?"), ale nesmí produkovat kód cvičení.

## Další čtení

1. [How to Write Go Code](https://go.dev/doc/code) — oficiální úvod do modulů
2. [Go Modules Reference](https://go.dev/ref/mod) — referenčně, ne k přečtení najednou
3. [Command go](https://pkg.go.dev/cmd/go) — kompletní seznam příkazů
