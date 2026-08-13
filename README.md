# Go do hloubky

Kurz Go (česky) pro vývojáře se silným PHP/Symfony zázemím. Cíl není „umět syntaxi",
ale mít pevný mentální model: paměť, hranice balíčků, chyby, souběžnost, architektura —
a umět řídit AI místo slepého přijímání jejího kódu.

Očekává se solidní zkušenost s programováním — není to intro do kódování. Detaily
(API stdlib, přesná syntaxe) si dohledáš (`go doc`, [go.dev](https://go.dev)); lekce učí
mentální model, ne kompletní referenci. Při učení použij AI skilly v repu
(`[go-deep-mentor](.cursor/skills/go-deep-mentor/)`,
`[go-deep-quiz](.cursor/skills/go-deep-quiz/)`,
`[go-deep-review](.cursor/skills/go-deep-review/)`) — viz [Cursor skilly](#cursor-skilly).
AI pomáhá s pochopením; v režimu `ZAKÁZÁNO` nepíše cvičení za tebe (dialog/kvíz/review ano).

Repozitář a Go modul: `[github.com/rdurica/go-deep](https://github.com/rdurica/go-deep)`
(anglický slug, český obsah).

**60 lekcí × 60–90 minut ≈ 80 hodin.** Každá lekce má teorii s funkčními příklady,
most z PHP, cvičení v `exercise/` a testy, které je ověří.
Adresář `solutions/` je referenční řešení — spoiler, ne začínej u něj.

Obsah: [BOOK.md](BOOK.md) · Postup: [PROGRESS.md](PROGRESS.md) · Osnova lekcí:
[course/lesson-map.md](course/lesson-map.md)

## Jak začít

1. Na GitHubu použij **Use this template** a vytvoř si vlastní repo
  (nebo forkni, pokud chceš často tahat updaty z upstreamu).
2. Naklonuj **svou** kopii a ověř toolchain:

```bash
go version          # potřebuješ Go 1.26+
make lesson L=01    # spustí testy první lekce (mají padat, než něco napíšeš)
```

1. Otevři [lessons/lesson-01/README.md](lessons/lesson-01/README.md): teorie → PHP rozdíly →
  `/go-deep-quiz` → stupně úkolu s `/go-deep-review` mezi nimi → final.
   Až final projde, skill odškrtne lekci v `[PROGRESS.md](PROGRESS.md)`; commitni u sebe.

Upstream drží `PROGRESS.md` a `[GAPS.md](GAPS.md)` jako šablony (CI regeneruje index).
Checkboxy a slabiny patří jen do tvé kopie — nespouštěj u sebe
`python3 scripts/generate_index.py`, jinak si progress přepíšeš. Updaty kurzu: remote
`upstream` a občas `git fetch upstream` + merge/rebase, nebo nová kopie z template.

## Cursor skilly

Tři skilly. Mentor při zaseknutí; kvíz po teorii; review po každém stupni a na závěr.
Všechny smí dialog i v `ZAKÁZÁNO` — nepíšou ale řešení cvičení za tebe. Slabiny si
ukládají do `[GAPS.md](GAPS.md)` a cíleně je přetrénovávají.


| Skill                                              | Kdy                      | Co dělá                                       |
| -------------------------------------------------- | ------------------------ | --------------------------------------------- |
| `[go-deep-mentor](.cursor/skills/go-deep-mentor/)` | při zaseknutí            | koncepty, Sokratovské nápovědy, GAPS          |
| `[go-deep-quiz](.cursor/skills/go-deep-quiz/)`     | po teorii, před cvičením | ~5 min mentální model                         |
| `[go-deep-review](.cursor/skills/go-deep-review/)` | po stupni / final        | testy PART nebo celku, dril, PROGRESS u final |


Příklady:

> Zasekl jsem se na lekci 7 u append — vysvětli backing array

> `/go-deep-quiz 01`

> `/go-deep-review 08 easy`



## Struktura lekce

```
lessons/lesson-07/
  README.md              # teorie → PHP → kvíz → 3 stupně úkolu → závěr
  tiers.txt              # které testy patří ke stupni 1/2/3
  exercise/
    exercise.go          # tvůj kód — zadání v komentářích; stupně oddělené
    exercise_test.go     # testy
  solutions/             # referenční řešení (spoiler)
```

Zadání je ve stubách. Stupně jsou krátké: oprava nebo mezera, 1–2 funkce, jedna věc
úsudku — ne mini-projekt. Po stupni: `make lesson L=NN PART=1` (nebo 2/3). Celé: `make lesson L=NN`.

## Checkpointy

Sedm lekcí je kontrolních. Nemají novou látku — mají shrnutí fáze, jednu velkou
kumulativní úlohu a bodovanou sebehodnotící rubriku, která ti řekne, co zopakovat.


| Lekce | Uzavírá                |
| ----- | ---------------------- |
| 18    | jazyk a paměťový model |
| 23    | idiomatický Go         |
| 31    | net/http a tooling     |
| 39    | architektura           |
| 50    | concurrency            |
| 55    | production             |
| 60    | celý kurz + capstone   |




## Projekty


| ID  | Projekt                               | Cesta                                                                        | Zadává      |
| --- | ------------------------------------- | ---------------------------------------------------------------------------- | ----------- |
| P01 | CSV CLI                               | `[projects/p01-csv-cli](projects/p01-csv-cli/ACCEPTANCE.md)`                 | lekce 17    |
| P02 | REST API                              | `[projects/p02-http-api](projects/p02-http-api/ACCEPTANCE.md)`               | lekce 31    |
| P03 | Hexagonální služba                    | `[projects/p03-hex-service](projects/p03-hex-service/ACCEPTANCE.md)`         | lekce 38    |
| P04 | Worker pool                           | `[projects/p04-worker-pool](projects/p04-worker-pool/ACCEPTANCE.md)`         | lekce 50    |
| P05 | Capstone                              | `[projects/p05-capstone](projects/p05-capstone/ACCEPTANCE.md)`               | lekce 59–60 |
| P06 | Bonus: Postgres + Redis *(volitelné)* | `[projects/p06-bookmarks-persist](projects/p06-bookmarks-persist/README.md)` | po lekci 60 |




## AI politika

Kurz nezakazuje AI. Nejdřív si vybuduješ úsudek, pak ji řídíš jako juniora.


| Lekce | Režim               | Co to znamená                                           |
| ----- | ------------------- | ------------------------------------------------------- |
| 01–18 | `ZAKÁZÁNO`          | AI nesmí psát cvičení, smí vysvětlovat koncepty         |
| 19–23 | `JEN VYSVĚTLENÍ`    | AI komentuje tvůj kód, nepřepisuje ho                   |
| 24–39 | `BOILERPLATE OK`    | tagy, DTO, wiring ano; jádro a hranice vlastníš ty      |
| 40–55 | `JUNIOR POD REVIEW` | agent navrhne, ty diffuješ a přepisuješ kritickou cestu |
| 56–60 | `TECH LEAD`         | spec → agent → strukturované review → capstone          |


Detail a review checklist: [docs/ai-playbook.md](docs/ai-playbook.md)

## Příkazy

```bash
make lesson L=07          # celé cvičení
make lesson L=07 PART=1   # jen jednoduchý stupeň
make race L=44            # s detektorem závodů (volitelně PART=)
make solutions            # všechna referenční řešení musí projít
make project P=02         # testy projektu
make check                # gofmt + vet + solutions + projekty (CI)
make mirror               # přegeneruje testy v solutions/ z exercise/
make help
```



## Dokumentace

- [docs/php-to-go.md](docs/php-to-go.md) — mosty ze Symfony
- [docs/best-practices.md](docs/best-practices.md) — konsolidované idiomy
- [docs/ai-playbook.md](docs/ai-playbook.md) — AI režimy a review checklist
- [docs/sources.md](docs/sources.md) — kanonické zdroje k dalšímu čtení
- [docs/tooling.md](docs/tooling.md) — Go, make, lint, CI
- [docs/authoring.md](docs/authoring.md) — jak se píše lekce (když chceš přispět)
- [docs/course-export.md](docs/course-export.md) — jak z repa udělat vedený kurz
- [course/](course/) — syllabus, moduly, poznámky pro lektora



## Tempo

Realisticky **5–7 h týdně** ≈ 14 týdnů, při 3 lekcích týdně ≈ 20 týdnů. Nesnaž se to
proklikat. Cvičení trénují úsudek, ne objem kódu: rané lekce píšeš ty krátké kusy
sám (AI vysvětluje), později AI navrhuje a ty držíš review. Cíl je mentální model,
ať AI neřídí tebe.

## Licence a přispívání

MIT — viz [LICENSE](LICENSE). Chyby a nápady: GitHub Issues. Jak psát další lekci:
[docs/authoring.md](docs/authoring.md).