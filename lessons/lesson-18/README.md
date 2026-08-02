# Lekce 18 — Checkpoint fáze 1

> **Čas:** ~90 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`

Checkpoint nepřináší novou látku. Projdeš si celou fázi 1 (lekce 03–17), postavíš jeden
malý balíček, který kombinuje šest jejích témat, a bodovou rubrikou zjistíš, kam se máš
vrátit, než se pustíš do fáze 2.

## Co budeš umět

- Zpaměti vyjmenovat pravidla, která tě v Go nejčastěji pálí: zero values, aliasing slice, nil mapa, hodnotový vs pointer receiver.
- Poskládat malý balíček z doménového typu, vlastní chyby, mapové agregace a generické pomocné funkce.
- Přijmout `io.Reader` na vstupu a `fmt.Stringer` na výstupu, místo abys pracoval s řetězci a `os.Stdout`.
- Odhadnout podle vlastního skóre, které lekce fáze 1 si máš zopakovat.

## Recap

### Otázky a odpovědi

**Co se stane, když zapíšeš do nil mapy?** Panika. Čtení a `len` jsou v pořádku, zápis ne.
Mapu musíš vytvořit přes `make` nebo literál. (lekce 03, 08)

**Proč `append` někdy změní i cizí slice?** Slice je hlavička (pointer, len, cap) nad
polem. Když se do kapacity vejde nový prvek, `append` píše do sdíleného pole a změna je
vidět všude. Když ne, alokuje nové pole a vazba se přeruší. Proto nikdy nesmíš spoléhat
na to, které z toho nastane — a proto existuje `slices.Clone`. (lekce 07)

**Kdy použít pointer receiver?** Když metoda mění stav, když je struktura velká, nebo když
jich část metod pointer potřebuje (pak je mají mít všechny, kvůli konzistenci sady metod).
Ne „pro jistotu". (lekce 06)

**Proč `len("žluťoučký")` nevrací počet písmen?** Protože vrací počet **bajtů**. String je
neměnná sekvence bajtů v UTF-8; iterace přes `range` dává runy, `[]byte` bajty a
`utf8.RuneCountInString` počet run. (lekce 09)

**Kdy se vyhodnotí argumenty `defer`?** Při jeho zápisu, ne při běhu. `defer fmt.Println(i)`
vytiskne hodnotu `i` z okamžiku deklarace; když chceš aktuální, zabal to do uzávěru.
Pojmenovaná návratová hodnota jde v deferu ještě změnit. (lekce 10)

**Jak se v Go implementuje rozhraní?** Nijak — implicitně tím, že typ má potřebné metody.
Rozhraní definuje **konzument**, ne ten, kdo ho splňuje. Proto jsou malá: `io.Reader` má
jednu metodu. (lekce 12, 13)

**Proč je `err != nil` lepší než výjimka?** Není „lepší", je **viditelná**. Chyba je
hodnota, takže se dá obalit (`%w`), porovnat (`errors.Is`), přetypovat (`errors.As`) a
vrátit dál. Panika je pro chyby programátora, ne pro chyby uživatele. (lekce 14)

**Kdy sáhnout po generice?** Když píšeš stejný algoritmus nad více typy a jediné, co je
odlišuje, je typ prvku (`Map`, `Filter`, `GroupBy`). Ne když se dá použít rozhraní. (lekce 15)

**Proč se neexportované pole neserializuje do JSON?** Protože `encoding/json` pracuje přes
reflexi a ta k neexportovaným polím nemá přístup. Tag na tom nic nezmění. (lekce 16)

**Proč testy nepotřebují assertion knihovnu?** Protože `if got != want { t.Errorf(...) }`
řekne přesně to, co chceš, a tabulka případů nahradí data providery. (lekce 17)

### Co si musíš pamatovat

| Téma | Pravidlo | Nejčastější past |
|------|----------|------------------|
| Zero values | každá proměnná má platnou hodnotu | zápis do nil mapy panikuje |
| Konstanty | netypovaná konstanta se přizpůsobí kontextu | `float64(1999/100)` je 19 |
| Funkce | více návratových hodnot, poslední `error` | ignorovaná chyba přes `_` |
| Structs | hodnotová sémantika, kopíruje se | metoda s hodnotovým receiverem nic nezmění |
| Pointery | `nil` dereference = panika | pointer „pro jistotu" všude |
| Slices | hlavička nad sdíleným polem | aliasing po `append` a `s[:2]` |
| Mapy | pořadí iterace je náhodné | test spoléhající na pořadí |
| Stringy | neměnné bajty v UTF-8 | `s[i]` je bajt, ne znak |
| defer | LIFO, argumenty se vyhodnotí hned | `defer` v cyklu drží zdroje |
| Balíčky | velké písmeno = veřejné, hranice je balíček | `utils` balíček |
| Interfaces | implicitní, definuje je konzument | obří rozhraní podle Symfony service |
| Errors | hodnota, obaluj `%w`, čti `Is`/`As` | `fmt.Errorf("%v")` zahodí řetěz |
| Generika | typový parametr, ne `any` | generika tam, kde stačí rozhraní |
| JSON | tagy + exportovaná pole | `omitempty` na `0`/`false` |
| Testy | table-driven, externí balíček | zapomenuté `-count=1` |

### Idiomy, které bys teď měl psát automaticky

```go
// 1. Chyba hned, šťastná cesta bez odsazení.
f, err := os.Open(path)
if err != nil {
	return fmt.Errorf("open %s: %w", path, err)
}
defer f.Close()

// 2. Přijmi rozhraní, vrať konkrétní typ.
func ParseRecords(r io.Reader) ([]Record, error)

// 3. Nulová hodnota, která je rovnou použitelná.
var buf bytes.Buffer

// 4. Kopie místo sdílení, když si výsledek nechceš rozbít.
sorted := slices.Clone(in)
```

## Rozdíly proti PHP

Za fází 1 stojí jedna jediná myšlenka, kterou stojí za to shrnout do jedné dvojice ukázek.
V PHP je objekt reference, chyba je výjimka a viditelnost je vlastnost třídy:

```php
final class Ledger {
    /** @throws ValidationException */
    public function add(Transaction $t): void { /* ... */ }
}
```

V Go je hodnota hodnota, chyba je návratová hodnota a hranicí viditelnosti je balíček:

```go
type Ledger struct{ txs []Transaction }

func (l *Ledger) Add(t Transaction) error { /* ... */ }
```

Co se mění v uvažování: **v Go je všechno, co PHP schovává, součástí signatury.** Kopíruje
se? Pozná se z toho, jestli je receiver pointer. Může to selhat? Pozná se z `error`. Je to
veřejné? Pozná se z prvního písmene. Fáze 1 byla o tom naučit se tyhle signály číst — a
psát je tak, aby je četl i někdo jiný.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Funkce bere `string` s obsahem souboru | v PHP se předává `file_get_contents()` | ber `io.Reader`, ať funguje soubor i stdin |
| Funkce si sama tiskne na `os.Stdout` | zvyk na `echo` v service vrstvě | vrať hodnotu nebo piš do `io.Writer` |
| Formátování ve volajícím kódu | v PHP `__toString()` nikdo nečeká | implementuj `String()` na typu |
| Chyba jako `errors.New(fmt.Sprintf(...))` | mix zvyků | vlastní typ chyby, nebo `fmt.Errorf` s `%w` |
| Agregace přes slice a lineární hledání | pole v PHP dělá obojí | mapa `map[K]V` a jeden průchod |
| Zkopírované `Map`/`Filter` pro každý typ | před generikami to jinak nešlo | typový parametr |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz 18`**. AI tě ~5 minut prověří mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Po doplnění spouštěj testy:

Stupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.

### Jednoduchý

Funkce: `String`, `Error`

```bash
make lesson L=18 PART=1
```

Pak **`/go-deep-review 18 easy`**.

### Střední

Funkce: `ParseTransactions`, `TotalsByCategory`

```bash
make lesson L=18 PART=2
```

Pak **`/go-deep-review 18 medium`**.

### Obtížný

Funkce: `String`, `BuildReport`

```bash
make lesson L=18 PART=3
```

Pak **`/go-deep-review 18 hard`**.

Až budou stupně hotové, porovnej se `solutions/` (spoiler).

## Sebehodnocení

Za každou položku, kterou zvládneš **bez nahlédnutí do lekce**, si dej 1 bod.

| # | Dovednost | Lekce |
|---|-----------|-------|
| 1 | Vyjmenuju zero value slice, mapy, pointeru, rozhraní a structu | 03 |
| 2 | Vysvětlím netypovanou konstantu a past celočíselného dělení | 03 |
| 3 | Napíšu funkci s více návratovými hodnotami a uzávěrem nad proměnnou cyklu | 04 |
| 4 | Rozhodnu mezi hodnotovým a pointer receiverem a zdůvodním to | 05, 06 |
| 5 | Popíšu, kdy `append` alokuje nové pole a co to udělá s aliasingem | 07 |
| 6 | Použiju `v, ok := m[k]`, smažu klíč a vím, proč je pořadí iterace náhodné | 08 |
| 7 | Rozliším bajt, runu a znak a projdu string obojím způsobem | 09 |
| 8 | Vysvětlím pořadí `defer`, vyhodnocení argumentů a `recover` na hranici | 10 |
| 9 | Rozvrhnu kód do balíčků podle domény a zdůvodním, co je exportované | 11 |
| 10 | Navrhnu malé rozhraní u konzumenta a napíšu k němu fake | 12, 13 |
| 11 | Obalím chybu přes `%w` a rozliším `errors.Is` od `errors.As` | 14 |
| 12 | Napíšu generickou funkci s omezením a poznám, kdy je zbytečná | 15 |
| 13 | Namapuju struct na JSON tagy a vysvětlím past `omitempty` | 16 |
| 14 | Napíšu table-driven test s podtesty, `t.TempDir` a `testdata/` | 17 |

| Skóre | Co s tím |
|-------|----------|
| 13–14 | Fáze 1 sedí, pokračuj na lekci 19. |
| 10–12 | Zopakuj konkrétní lekce z řádků, kde jsi bod nedostal, a udělej jejich část C znovu. |
| 7–9 | Zopakuj lekce 06, 07 a 14 — pointery, slices a errory jsou základ všeho dalšího. |
| 4–6 | Projdi znovu celý blok 03–10, ideálně tak, že cvičení napíšeš od nuly bez nahlížení. |
| 0–3 | Vrať se na lekci 03 a jdi fází 1 znovu; na fázi 2 zatím nemá smysl přecházet. |

Nezapomeň, že součástí fáze 1 je i projekt
[P01 — csvstats](../../projects/p01-csv-cli/ACCEPTANCE.md). Pokud jsi ho nedokončil,
udělej ho teď: je to nejlepší kontrola, jestli témata fáze umíš složit dohromady.

## Závěrečné otázky

Spusť **`/go-deep-review 18 final`**. AI projde body níže, doptá se a ověří pochopení. Celé cvičení ověří `make lesson L=18` (+ `make race L=18`, pokud to lekce vyžaduje).

- [ ] Umíš vysvětlit, proč `Money` drží centy a ne `float64`
- [ ] Umíš vysvětlit, proč `ParseTransactions` bere `io.Reader` a ne `string`
- [ ] Umíš vysvětlit rozdíl mezi `errors.Is` a `errors.As` na vlastním kódu
- [ ] Umíš vysvětlit, proč je `GroupBy` generická, ale `TotalsByCategory` ne

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Mentor, kvíz i review (dialog) jsou vždy OK; v tomto režimu AI nesmí psát kód cvičení.
## Další čtení

1. [Effective Go](https://go.dev/doc/effective_go)
2. [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
3. [Go blog — Errors are values](https://go.dev/blog/errors-are-values)
4. [Go Proverbs](https://go-proverbs.github.io/)
