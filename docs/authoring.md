# Jak se v tomhle repozitáři píše lekce

Závazná pravidla pro autory (včetně AI agentů). Etalon kvality:
[lessons/lesson-08](../lessons/lesson-08/README.md) (sekce Úkol + stuby v `exercise/`).

## Cílová skupina

Zkušený PHP/Symfony vývojář. Umí OOP, DI, testy, HTTP, SQL, deployment. **Neumí Go.**
Nevysvětluj mu, co je funkce nebo cyklus. Vysvětluj, čím se Go liší od toho, co má
zažité, a kde ho jeho reflexy zradí.

## Rozsah

| | |
|---|---|
| Čas na lekci | 30–45 minut včetně cvičení |
| Teorie | 1200–2000 slov |
| Cvičení | 3 stupně; cílově ~15–20 min práce studenta |
| Jednotek k úpravě | typicky **2–4** (funkce/metody, které student píše nebo opravuje) |

Cvičení trénuje **úsudek**, ne katalog API. Když zadání vypadá jako mini-projekt
(více balíčků, desítky TODO), patří do [`projects/`](../projects/), ne do lekce.

## Struktura souborů

```
lessons/lesson-NN/
  README.md
  tiers.txt              # 1:Test…|Test…  / 2:… / 3:…  pro make lesson PART=
  exercise/
    exercise.go          # package exercise, stuby s // TODO; // --- Stupeň: … ---
    exercise_test.go     # package exercise_test, JEDINÝ zdroj testů
  solutions/
    exercise.go          # package solutions, kompletní řešení
    exercise_test.go     # NEPIŠ RUČNĚ — generuje scripts/mirror_tests.sh
```

Modul je `github.com/rdurica/go-deep`, import cvičení tedy
`github.com/rdurica/go-deep/lessons/lesson-NN/exercise`.

Testy se píšou **jen jednou**, v `exercise/`, jako externí testovací balíček:

```go
package exercise_test

import (
	"testing"

	exercise "github.com/rdurica/go-deep/lessons/lesson-07/exercise"
)
```

Variantu pro `solutions/` vyrobí skript:

```bash
./scripts/mirror_tests.sh 07
```

Díky pojmenovanému importu `exercise` je tělo testu v obou balíčcích identické, takže
referenční řešení je ověřované přesně tím, co dostane student.

## Obsah README

Kostra je v [templates/lesson-README.md](../templates/lesson-README.md). Sekce jsou povinné
a v tomto pořadí.

**Co budeš umět** — 3–5 odrážek, každá je schopnost („vysvětlit, proč…"), ne téma.

**Teorie** — 2–5 podsekcí s vlastním nadpisem `###`. Každý blok kódu musí být skutečný Go,
který se zkompiluje. Ukazuj i chybné varianty, ale označ je komentářem.

**Rozdíly proti PHP** — konkrétní kód v PHP a Go protějšek vedle sebe, plus věta o tom,
který návyk je potřeba opustit. Ne tabulka pojmů. (Dřívější název: PHP → Go most.)

**Časté chyby** — tabulka `Chyba | Proč vzniká | Jak to udělat správně`, 4–6 řádků.
Aspoň dvě z nich musí vycházet z PHP reflexů.

**AI kvíz** — výzva spustit **`/go-deep-quiz NN`** (~5 min). Otázky generuje skill
z Teorie; v README nehardcoduj sadu otázek. Odkaz na [`GAPS.md`](../GAPS.md).

**Úkol** — tři podsekce **Jednoduchý / Střední / Obtížný**. U každé: krátká akce
(`Oprav …` / `Implementuj …` / `Doplň …`) + identifikátory pro navigaci,
`make lesson L=NN PART=1|2|3`, výzva `/go-deep-review NN easy|medium|hard`.
Kontrakt zůstává v komentáři nad metodou ve stubu.
Žádné labely A/B/C, žádné `např.` / `Příklady:`.

**Závěrečné otázky** — koncepční checklist (bez „make … prochází“ — to řeší final review
příkazem). Výzva **`/go-deep-review NN final`**.

**AI režim** — štítek podle fáze + odkaz na `docs/ai-playbook.md`. Mentor/kvíz/review
(dialog) vždy OK; `ZAKÁZÁNO` = zákaz codegen cvičení.

**Další čtení** — 2–4 skutečné odkazy na go.dev, pkg.go.dev nebo blog.golang.org.

## tiers.txt

Tři řádky, regex pro `go test -run` (jména `Test…` funkcí, oddělená `|`):

```
1:TestCloneMap|TestCloneMapIndependent
2:TestWordCount|TestNewSet|TestSetAdd|TestSetHas
3:TestAddStock
```

`make lesson L=NN PART=1` spustí jen odpovídající testy. Etalon: lekce 08.

## AI režimy podle lekcí

| Lekce | Štítek |
|-------|--------|
| 01–18 | `ZAKÁZÁNO` |
| 19–23 | `JEN VYSVĚTLENÍ` |
| 24–39 | `BOILERPLATE OK` |
| 40–55 | `JUNIOR POD REVIEW` |
| 56–60 | `TECH LEAD` |

## Pravidla pro cvičení

Cvičení musí **skutečně učit téma lekce**. Tohle je nejdůležitější pravidlo v celém
dokumentu, protože předchozí verze kurzu na něm ztroskotala.

### Typy podle stupně (pevný default)

| Stupeň | Default | Kdy jinak |
|--------|---------|-----------|
| 1 | Find-the-bug nebo complete-the-gap | L21/22/23/56/57 → review lab |
| 2 | Short greenfield 1–2 funkce | — |
| 3 | Jedna věc úsudku (edge / `map[K]*V` / race / leak) | — |

Žádný volný mix. Soft limit: studentská práce ve stubu typicky **< ~80 LOC** (ne 200+ TODO).

### Katalog typů úkolů

Slovník pojmů — typ patří na stupeň podle tabulky výše, ne libovolně.

1. **Find-the-bug** — záměrně vadný kód; testy před opravou padají (vzor L08 `CloneMap`, L44)
2. **Complete-the-gap** — většina kódu hotová, chybí kritický kus
3. **Short greenfield** — max 1–2 funkce, ne katalog CRUD
4. **Review lab** — hotový „PR“ / diff; student opraví nebo napíše failing test. Jen L21/22/23/56/57, stupeň 1.

Zakázané vzory:

```go
// ŠPATNĚ — lekce o kontejnerech a health endpointech
func HealthPath() string { return "/healthz" }

// ŠPATNĚ — lekce o pprof
func AllocSuspicious(allocsPerOp int) bool { return allocsPerOp > 1000 }

// ŠPATNĚ — lekce o goroutine leacích, ani jedna goroutina
func LeakRisk(hasWait, hasCancel bool) bool { return !hasWait || !hasCancel }

// ŠPATNĚ — test porovnává větu, ne chování
func VisibleAfterUnlock() string { return "unlock happens before lock" }
```

Správně vypadá cvičení tak, že student musí použít přesně tu konstrukci, o které je lekce:

- lekce o HTTP → student píše `http.Handler` a test ho volá přes `httptest`
- lekce o goroutinách → student spouští goroutiny, test běží s `-race` a hlídá leak
  přes `runtime.NumGoroutine()`
- lekce o pprof → student píše benchmark a optimalizuje alokace, test hlídá
  `testing.AllocsPerRun`
- lekce o fuzz → v lekci je skutečný `FuzzXxx` a korpus v `testdata/`

## Pravidla pro kód

- Pouze standardní knihovna. Žádné závislosti navíc, `go.mod` zůstává bez `require`.
- Cílová verze Go 1.26.
- Všechno projde `gofmt -l` (nic nevypíše) a `go vet ./...`.
- **Identifikátory v kódu jsou vždy anglicky** — funkce, typy, pole, proměnné,
  `func Test…`, názvy v `t.Run("…")`, klíče v `tiers.txt`. Čeština patří jen do
  výukového materiálu: README, stub doc komentáře, hlášky `t.Error`/`t.Fatal`.
- V české próze mluv o konceptu podle typu v kódu (`Node` → nod), ne o doslovném
  překladu parametru (`head` není „hlava“).
- Exportované identifikátory mají doc comment začínající jménem identifikátoru.
  Ve stubách je komentář nad metodou **zdroj pravdy** pro zadání: 2–5 řádků česky
  (chování + hraniční případy), bez hotového řešení. Pořadí funkcí od jednodušších
  ke složitějším; skupiny odděl komentáři `// --- Stupeň: jednoduchý|střední|obtížný ---`.
- Stuby v `exercise/` nepanikují — greenfield mají `// TODO` a vrací zero value
  (`""`, `0`, `nil`, `*new(T)` u pojmenovaných typů). Void funkce mají jen komentář;
  funkce s `t *testing.T` volají `t.Fatal("TODO")`. Typy, konstanty a signatury
  jsou předvyplněné, aby se balíček zkompiloval a `go test` padal přes `t.Error`/`t.Fatal`.
  Stuby vracející `<-chan` mají vracet zavřený prázdný kanál, ne `nil` — jinak testy
  visí na čtení z nil kanálu.
- **Find-the-bug** je výjimka: tělo je záměrně vadná (ale kompilující) implementace,
  ne `// TODO`. Nad funkcí uveď `// POZOR: kód níže je ZÁMĚRNĚ VADNÝ.` Testy musí
  před opravou padat a po opravě procházet. V `solutions/` je správná verze.
- Řešení v `solutions/` musí být idiomatické — je to zároveň ukázka stylu.
  V `solutions/` nepoužívej `// TODO`. Stejné `// --- Stupeň: … ---` komentáře jako ve stubu.

## Testy

- Table-driven, kde to dává smysl; podtesty přes `t.Run` s **anglickým** názvem.
- Jména `Test…` a položky v `tiers.txt` anglicky (CamelCase), bez českých kořenů.
- Hlášky česky, formát: `t.Errorf("Foo(%q) = %q, chci %q", in, got, want)`.
- Pokrytí hraničních případů: prázdný vstup, `nil`, nula, přetečení, duplicity.
- U souběžného kódu vždy varianta ověřitelná `-race` a kontrola, že goroutiny skončily.
- Test nesmí být splnitelný napevno zadrátovanou hodnotou — kde hrozí, přidej náhodná
  nebo generovaná data.
- Jména testů musí jít rozdělit do `tiers.txt` (prefixy `Test…` odpovídající stupňům).

### Izolace stubů

Test funkce/metody **X** smí volat jen **X** (+ stdlib / lokální helpery v `_test.go`).
Nedokončený stub B nesmí shodit test A.

- **Assert** nikdy nevolá jiný studentský stub (`Len`, `Has`, `Sorted`, `Cents`, …).
  Stav ověř přímo: `len(mapa)`, comma-ok, exportovaná pole, návratová hodnota FUT,
  HTTP body, `errors.As` na exportovaná pole chyby.
- **Fixture** preferuj literály / zero value / přímý zápis do exportovaného stavu
  (`Set{"a": {}}`, `&Node{Val: 1, Next: …}`), ne `NewX`/`Add`, pokud to typ dovolí.
- PART N nesmí v assertu volat stub z PART > N.
- **Výjimky:**
  - Oracle páry (Encode↔Decode, Slow↔Fast) — záměr lekce.
  - Lifecycle ADT ve **stejném PART** (např. Push+Pop), kde API nemá jiný
    pozorovatelný výstup — stále bez assertu přes metodu z pozdějšího PART.
- Opaque typ (unexported pole): assert jen přes návratovou hodnotu FUT / exportovaná
  pole výsledku, nebo whitebox `package exercise` (ne `exercise_test`). Whitebox
  jen když unexported stav jinak nejde izolovat; `mirror_tests.sh` přepíše
  `package exercise` → `package solutions`. **Ne** číst stav přes cizí TODO getter.

## Checkpointové lekce

Lekce 18, 23, 31, 39, 50, 55 a 60 jsou checkpointy. Liší se takto:

- Nemají novou teorii. Místo sekce **Teorie** mají **Recap** — hutné shrnutí fáze.
- Cvičení je **kumulativní**; stupně mohou být 2+3 nebo jen obtížný, pokud split nedává smysl.
- Mají **AI kvíz**, **Závěrečné otázky** a **Sebehodnocení** s bodovanou rubrikou.
- Checkpointy 31, 50 a 60 zároveň zadávají a ověřují projekt (P02, P04, P05).

## Než odevzdáš lekci

```bash
gofmt -l lessons/lesson-NN
./scripts/mirror_tests.sh NN
(cd lessons/lesson-NN/solutions && go test -count=1 .)   # musí projít
(cd lessons/lesson-NN/exercise  && go test -count=1 .)   # musí spadnout (neúplné stuby)
make lesson L=NN PART=1   # regex z tiers.txt musí sedět
go vet ./lessons/lesson-NN/...
```
