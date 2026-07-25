# Jak se v tomhle repozitáři píše lekce

Závazná pravidla pro autory (včetně AI agentů). Etalon kvality:
[lessons/lesson-03](../lessons/lesson-03/README.md).

## Cílová skupina

Zkušený PHP/Symfony vývojář. Umí OOP, DI, testy, HTTP, SQL, deployment. **Neumí Go.**
Nevysvětluj mu, co je funkce nebo cyklus. Vysvětluj, čím se Go liší od toho, co má
zažité, a kde ho jeho reflexy zradí.

## Rozsah

| | |
|---|---|
| Čas na lekci | 60–90 minut včetně cvičení |
| Teorie | 1200–2000 slov |
| Cvičení | 3 části (A rozcvička ~10 min, B jádro ~35 min, C rozšíření ~20–25 min) |
| Funkcí k implementaci | 5–9 napříč A/B/C |

## Struktura souborů

```
lessons/lesson-NN/
  README.md
  exercise/
    exercise.go          # package exercise, stuby s panic("TODO: úkol X")
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

**Co budeš umět** — 3–5 odrážek, každá je schopnost („vysvětlit, proč…", „rozhodnout
mezi…"), ne téma („slices").

**PHP → Go most** — konkrétní kód v PHP a jeho Go protějšek vedle sebe, plus věta o tom,
který návyk je potřeba opustit. Ne tabulka pojmů.

**Teorie** — 2–5 podsekcí s vlastním nadpisem `###`. Každý blok kódu musí být skutečný Go,
který se zkompiluje. Ukazuj i chybné varianty, ale označ je komentářem.

**Časté chyby** — tabulka `Chyba | Proč vzniká | Jak to udělat správně`, 4–6 řádků.
Aspoň dvě z nich musí vycházet z PHP reflexů.

**Úkol** — části A, B, C s odhadem času. Popiš, co má funkce dělat, včetně hraničních
případů. Nepíšeš řešení, ale zadání musí být jednoznačné, aby ho testy nemohly překvapit.
Každá část A/B/C končí **jedním** krátkým řádkem `např. \`…\` → \`…\`` (vstup → výstup,
hodnoty ze testů). Ne blok s více odrážkami. U checklistových částí stačí jedna konkrétní
očekávaná akce nebo výsledek.

**Ověření** — checklist: `make lesson L=NN` plus 3–5 otázek k sebehodnocení
(„umíš vysvětlit…"). Hned pod nadpis `## Ověření` patří výzva spustit
Cursor skill **`/go-deep-review`** s číslem lekce (viz šablona) — AI postupně
projde body a ověří pochopení.

**AI režim** — jeden ze štítků podle fáze (viz níže) a odkaz na `docs/ai-playbook.md`.

**Další čtení** — 2–4 skutečné odkazy na go.dev, pkg.go.dev nebo blog.golang.org.
Nevymýšlej URL.

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
- Exportované identifikátory mají doc comment začínající jménem identifikátoru.
- Stuby v `exercise/` panikují: `panic("TODO: úkol B")`. Typy, konstanty a signatury
  jsou předvyplněné, aby se balíček zkompiloval.
- Řešení v `solutions/` musí být idiomatické — je to zároveň ukázka stylu.

## Testy

- Table-driven, kde to dává smysl; podtesty přes `t.Run`.
- Hlášky česky, formát: `t.Errorf("Foo(%q) = %q, chci %q", in, got, want)`.
- Pokrytí hraničních případů: prázdný vstup, `nil`, nula, přetečení, duplicity.
- U souběžného kódu vždy varianta ověřitelná `-race` a kontrola, že goroutiny skončily.
- Test nesmí být splnitelný napevno zadrátovanou hodnotou — kde hrozí, přidej náhodná
  nebo generovaná data.

## Checkpointové lekce

Lekce 18, 23, 31, 39, 50, 55 a 60 jsou checkpointy. Liší se takto:

- Nemají novou teorii. Místo sekce **Teorie** mají **Recap** — hutné shrnutí fáze
  formou otázek a odpovědí a tabulky „co si musíš pamatovat".
- Cvičení je **kumulativní**: jedna větší úloha, která kombinuje aspoň čtyři témata
  z celé fáze.
- Navíc mají sekci **Sebehodnocení** s bodovanou rubrikou a doporučením, které lekce
  zopakovat při nízkém skóre.
- Checkpointy 31, 50 a 60 zároveň zadávají a ověřují projekt (P02, P04, P05).

## Než odevzdáš lekci

```bash
gofmt -l lessons/lesson-NN
./scripts/mirror_tests.sh NN
(cd lessons/lesson-NN/solutions && go test -count=1 .)   # musí projít
(cd lessons/lesson-NN/exercise  && go test -count=1 .)   # musí spadnout na TODO
go vet ./lessons/lesson-NN/...
```
