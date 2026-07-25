# Lekce 57 — Strukturované review AI kódu a diff lab

> **Čas:** ~90 min · **Fáze:** 7 — Inženýrství v době AI · **AI režim:** `TECH LEAD`

## Co budeš umět

- Vysvětlit, čím se review vygenerovaného kódu liší od review kolegy, a podle toho změnit postup.
- Číst diff od jádra domény ven místo shora dolů a najít kritickou cestu dřív, než ti dojde pozornost.
- Rozpoznat šest typických nálezů v Go od modelu a doložit každý konkrétním řádkem.
- Napsat vlastní kontrolu nad `go/parser` a `go/ast`, která ty nálezy hledá strojově.
- Rozhodnout, kdy má smysl diff opravit a kdy ho zahodit a napsat to sám.

## PHP → Go most

V Symfony týmu se review opírá o strukturu, kterou vidíš na první pohled: soubor
v `Controller/` dělá HTTP, soubor v `Entity/` je doména, a když v controlleru uvidíš SQL,
víš to hned. Statická analýza (PHPStan, Psalm) navíc chytí velkou část chyb typu
„zapomenutá návratová hodnota".

```php
// PHPStan level 8 tohle nepustí
$user = $this->repo->find($id);
echo $user->getName();   // možná null
```

Go ti nedá ani jedno zadarmo. `go vet` je záměrně konzervativní a nechytí `_ = err`,
context ve structu ani goroutinu bez ukončení. Zato ti dá něco, co v PHP nemáš:
**parser jazyka jako součást standardní knihovny**.

```go
fset := token.NewFileSet()
file, err := parser.ParseFile(fset, "src.go", src, parser.SkipObjectResolution)
// file je AST, po kterém můžeš chodit a hledat přesně své vzory
```

Změna v uvažování: pravidla, která v PHP kupuješ jako nástroj, si v Go za odpoledne
napíšeš sám — a přesně na míru chybám, které dělá tvůj agent. Review pak není jen čtení,
ale i pár desítek řádků kódu, které čtou za tebe.

## Teorie

### Proč AI kód potřebuje jiné review

Review kolegy stojí na tom, že **záměr odhadneš z kódu**. Když někdo napíše zvláštní
podmínku, obvykle za tím je důvod, který si vybavíš z ranního standupu. Model žádný záměr
nemá; má statistiku. Z toho plynou tři rozdíly:

1. **Kód vypadá sebejistě.** Jména jsou správná, komentáře souhlasí s názvem funkce,
   struktura odpovídá učebnici. Signál „tady si autor nebyl jistý" — nepřesné jméno,
   TODO, zakomentovaný pokus — chybí. Nemůžeš se řídit tím, kde kód *vypadá* podezřele.
2. **Chyby jsou v detailu, ne v záměru.** Celkový návrh bývá rozumný, protože je průměrem
   tisíců repozitářů. Chybí `rows.Err()`, `defer resp.Body.Close()`, wrapping s `%w`
   u jedné ze čtyř větví, `ctx` se předává všude kromě jednoho volání.
3. **Objem je levný.** Kolega ti nepošle 900 řádků. Agent ano — a když je čteš stejným
   tempem jako lidský diff, dojde ti pozornost přesně v místě, kde je chyba.

Praktický důsledek: review AI kódu musí být **checklistové a strojové**, ne intuitivní.
Intuici si šetři na otázku „patří tenhle kód vůbec do téhle vrstvy".

### Čtení diffu od jádra

Neproklikávej diff shora dolů. Soubory přijdou v abecedním pořadí, takže začneš u
`cmd/`, `config` a testů — a do domény dorazíš vyčerpaný.

Pořadí, které funguje:

1. **Podpis změn v doméně** — co se změnilo v `func` uvnitř doménových balíčků. Tady jsou
   chyby drahé.
2. **Error model** — nové sentinel chyby, nové `fmt.Errorf`, změněné návratové hodnoty.
3. **Souběžnost** — každé `go `, `chan`, `sync.`, `select`.
4. **Hranice** — nové importy v doménovém balíčku. `net/http` v doméně je stopka bez ohledu
   na zbytek diffu.
5. **Wiring a testy** — až nakonec, protože chyba tady je levná a viditelná.

Kroky 1 a 3 jde zautomatizovat: unified diff je textový formát a hunk hlavička obsahuje
kus kontextu (`@@ -24,9 +25,12 @@ func (s *Server) createBookmark(...)`). Filtr, který
z 900 řádků nechá 120 řádků měnících funkce, uděláš za třicet řádků kódu — a přesně to je
úkol C.

### Šest nálezů, které dělá model pořád dokola

```go
// 1. ignorovaná chyba — vypadá jako úklid, je to ztracený signál
_ = json.NewEncoder(w).Encode(resp)

// 2. context ve structu — lifetime requestu přilepený na objekt s jiným lifetimem
type Service struct {
	ctx  context.Context // ŠPATNĚ
	repo Repo
}

// 3. goroutina bez ukončení — nikdo nečeká, nikdo neruší
go func() { s.reindex() }()

// 4. obří interface „pro mockování" — Symfony service přepsaná do Go
type BookmarkServiceInterface interface {
	Create(...) ; Update(...) ; Delete(...) ; Find(...) ; FindAll(...)
}

// 5. panika místo chyby — v PHP by to byla výjimka, tady je to pád procesu
if url == "" {
	panic("url is required") // ŠPATNĚ, vrať error
}

// 6. chybějící defer Close a rows.Err()
rows, _ := db.Query(q)
for rows.Next() { /* ... */ }   // chybí defer rows.Close() i rows.Err()
```

Tři z nich (`_ =`, context ve structu, context ne jako první parametr) najdeš strojově
nad AST. Zbylé tři potřebují kontext, takže zůstávají na tobě — proto je checklist
rozdělený na „stroj" a „člověk".

### `go/parser` a `go/ast` v pěti minutách

```go
fset := token.NewFileSet()                 // mapuje pozice v AST na řádky
file, err := parser.ParseFile(fset, "src.go", src, parser.SkipObjectResolution)
if err != nil {
	return nil, fmt.Errorf("parse: %w", err)
}

ast.Inspect(file, func(n ast.Node) bool {
	assign, ok := n.(*ast.AssignStmt)   // hledáme přiřazení
	if !ok {
		return true                     // true = pokračuj do potomků
	}
	for _, lhs := range assign.Lhs {
		if id, ok := lhs.(*ast.Ident); ok && id.Name == "_" {
			fmt.Println("blank na řádku", fset.Position(id.Pos()).Line)
		}
	}
	return true
})
```

Tři věci, které je dobré vědět hned:

- **Parser netypuje.** `parser.ParseFile` neřeší, jestli `context` je opravdu balíček
  `context`; vidí jen `SelectorExpr{X: Ident("context"), Sel: Ident("Context")}`. Pro
  linter to skoro vždy stačí a je to o tři řády jednodušší než `go/types`.
- **Pozice se překládají přes `FileSet`.** Uzel nese `token.Pos` (offset), řádek dostaneš
  jen přes `fset.Position(pos).Line`.
- **`FieldList` sdružuje jména.** `func f(a, b int)` je jeden `Field` se dvěma `Names`.
  Kdo počítá `len(list.List)`, spočítá parametry špatně.

Heuristika linteru smí být hloupá, když je **předvídatelná**. `_ = doSomething()` označíme
vždycky, i když funkce nic nevrací — false positive, který stojí dvě sekundy, je lepší než
tichý průchod ignorované chyby.

### Jak měřit review a kdy diff zahodit

Založ si po review jednoduchou statistiku: kolik nálezů jsi našel ty, kolik jich našel
reviewer po tobě nebo produkce, a kolik jsi jich označil zbytečně. Poměr
*nalezené / celkem* je jediné číslo, které ti řekne, jestli se tvoje review zlepšuje.
V lekci 58 z něj uděláš `ScoreReview`.

A pravidlo, které šetří nejvíc času: **když diff opravuješ na víc než třech místech
zároveň, zahoď ho.** Tři vysvětlené opravy jsou konverzace; deset oprav znamená, že
implementaci vlastníš ty, jen ji čteš pozpátku. Napsat 80 řádků domény ručně trvá dvacet
minut. Přesvědčit model přes pět kol trvá hodinu a výsledek stejně nikdo nezná.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| Review shora dolů podle souborů | tak to zobrazuje UI | začni u `func` v doméně, wiring až nakonec |
| Důvěra v hezký vzhled kódu | model nikdy nepůsobí nejistě | řiď se checklistem, ne dojmem |
| Schválení diffu, který nikdo nespustil | v PHP CI odhalila většinu | `go build`, `go vet`, `go test -race` před čtením |
| Počítání parametrů jako `len(Params.List)` | reflex z PHP reflexe | `Field` sdružuje jména, sčítej `len(f.Names)` |
| Řádek z `token.Pos` bez `FileSet` | `Pos()` vypadá jako číslo řádku | vždy `fset.Position(pos).Line` |
| Nekonečné dokolečka opravování promptem | zdá se to levnější než psát | tři opravy = konverzace, deset = přepiš to sám |

## Úkol

Pracuj v `exercise/`. Postupuj A → B → C, po každé části spusť test.

Píšeš vlastní review nástroj: čtečku funkcí, tři kontroly nad AST a filtr diffu.

### A — rozcvička (~10 min)

1. `Severity.String()` → `"INFO"`, `"WARN"`, `"ERROR"`.
2. `ParseFuncs(src string) ([]FuncInfo, error)` — přes `go/parser` rozeber zdroják
   jednoho souboru a vrať pro každou funkci i metodu (v pořadí výskytu):
   - `Name` — jméno funkce (u metody bez přijímače),
   - `Params` / `Results` — počty **jednotlivých** parametrů a návratových hodnot
     (`func f(a, b int) (int, error)` → 2 a 2; přijímač se nepočítá),
   - `Lines` — počet řádků deklarace včetně hlavičky i uzavírací závorky,
   - `Exported` — začíná velkým písmenem.

   Nevalidní zdroják vrací chybu.

### B — jádro (~35 min)

Tři kontroly. Když se zdroják nepodaří rozebrat, každá vrací **jediný** nález
`{Rule: "parse-error", Severity: SeverityError, Line: 0}` s libovolnou neprázdnou zprávou.
Nálezy jsou vždy v pořadí výskytu ve zdrojáku a mají vyplněné `Line`, `Rule`, `Severity`
a `Message`.

1. `CheckIgnoredErrors(src string) []Finding` — najdi přiřazení, jehož pravá strana je
   **jediné volání funkce** a jehož levá strana obsahuje `_`. Zachytí tedy `_ = f()`
   i `v, _ := f()`, ale ne `_ = x` ani `v, ok := m[k]`. Řádek ber z pozice blank
   identifikátoru. `Rule: "ignored-error"`, `Severity: SeverityError`.
2. `CheckContextInStruct(src string) []Finding` — pole typu `context.Context`
   (i `*context.Context`) v jakémkoli struct typu, včetně anonymních structů uvnitř funkcí.
   Řádek ber z pozice pole. `Rule: "context-in-struct"`, `Severity: SeverityError`.
3. `CheckContextNotFirst(src string) []Finding` — funkce nebo metoda, která má parametr
   typu `context.Context` na jiné než první pozici (pozor na `func f(a, b int, ctx …)`,
   kde je `ctx` třetí parametr, ale druhý `Field`). Na jednu funkci nanejvýš jeden nález.
   `Rule: "context-not-first"`, `Severity: SeverityWarn`.

### C — rozšíření (~25 min)

1. `CriticalPath(diff string) []Hunk` — parser unified diffu. Vrať jen hunky, které
   **zároveň** obsahují změněný řádek (začíná `+` nebo `-`) a týkají se funkce (řetězec
   `func ` je v hlavičce hunku za `@@`, nebo v některém řádku těla). Vyplň:
   - `File` — cesta z `+++ b/...` bez prefixu `b/` (u `/dev/null` použij cestu z `--- a/...`),
   - `OldStart` / `NewStart` — čísla z `@@ -24,9 +25,12 @@`,
   - `Header` — text za druhým `@@`, ořezaný,
   - `Lines` — řádky těla hunku, jak jsou (včetně vedoucího znaku).

   Testovací diffy jsou v `testdata/`. Hunk bez změn a diff, který se netýká Go kódu,
   se do výsledku nedostanou.

2. `ReviewReport(findings []Finding) string` — přehled seskupený podle závažnosti
   v pořadí ERROR, WARN, INFO. Prázdné skupiny vynech, prázdný vstup dá `"Žádné nálezy.\n"`.
   Uvnitř skupiny řaď podle řádku, pak podle `Rule`, pak podle zprávy. Formát:

```
## ERROR (2)
- ř. 7 [context-in-struct] ctx ve structu
- ř. 30 [ignored-error] chyba do _

## WARN (1)
- ř. 12 [context-not-first] ctx není první
```

   Mezi skupinami je jeden prázdný řádek, výstup končí jediným `\n`.

```bash
make lesson L=57
```

Až budeš hotový, porovnej se `solutions/` (spoiler).

## Ověření

Po dokončení úkolů spusť v Cursoru **`/go-deep-review`** a zadej třeba jen `57`. AI tě postupně projde body níže, doptá se a ověří pochopení — nestačí jen zelené testy.

- [ ] `make lesson L=57` prochází
- [ ] Umíš vysvětlit tři rozdíly mezi review kolegy a review agenta
- [ ] Umíš popsat pořadí, ve kterém čteš velký diff, a proč zrovna takhle
- [ ] Umíš najít v cizím diffu ignorovanou chybu, context ve structu a goroutinu bez ukončení
- [ ] Umíš vysvětlit, proč `len(Params.List)` není počet parametrů
- [ ] Umíš říct, kdy diff přestaneš opravovat a napíšeš ho sám

## AI režim

`TECH LEAD` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Checklist z playbooku je vstupem téhle lekce. Agent smí generovat kód, který revidíruješ,
a smí navrhovat další pravidla lintera; samotný linter napiš sám, protože je to tvoje
záruka nezávislá na modelu.

## Další čtení

1. [go/ast — dokumentace balíčku](https://pkg.go.dev/go/ast)
2. [go/parser — dokumentace balíčku](https://pkg.go.dev/go/parser)
3. [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
4. [Go blog — Contexts and structs](https://go.dev/blog/context-and-structs)
