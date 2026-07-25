# Go do hloubky — obsah

60 lekcí po 60–90 minutách. Každá lekce má teorii, most z PHP,
tři stupňované úkoly a testy: [`lessons/lesson-NN/README.md`](lessons/lesson-01/README.md),
cvičení v `exercise/`, referenční řešení v `solutions/`.

Osnova lekcí: [course/lesson-map.md](course/lesson-map.md).

## Fáze 0 — Setup a mentální reset

- [Lekce 01 — Toolchain, moduly a workspace](lessons/lesson-01/README.md) · ~75 min
- [Lekce 02 — Mentální reset: PHP → Go](lessons/lesson-02/README.md) · ~75 min

## Fáze 1 — Jazyk a paměťový model

- [Lekce 03 — Typy, zero values a konstanty](lessons/lesson-03/README.md) · ~90 min
- [Lekce 04 — Funkce, multiple returns, closures](lessons/lesson-04/README.md) · ~90 min
- [Lekce 05 — Structs, metody a embedding](lessons/lesson-05/README.md) · ~90 min
- [Lekce 06 — Pointery: hodnota vs reference](lessons/lesson-06/README.md) · ~90 min
- [Lekce 07 — Slices: pole, append, internals, aliasing](lessons/lesson-07/README.md) · ~90 min
- [Lekce 08 — Mapy](lessons/lesson-08/README.md) · ~85 min
- [Lekce 09 — Stringy, runy a byty](lessons/lesson-09/README.md) · ~85 min
- [Lekce 10 — defer, panic, recover](lessons/lesson-10/README.md) · ~85 min
- [Lekce 11 — Balíčky, export a viditelnost](lessons/lesson-11/README.md) · ~85 min
- [Lekce 12 — Interfaces I: implicitní implementace](lessons/lesson-12/README.md) · ~85 min
- [Lekce 13 — Interfaces II: io.Reader, io.Writer, kompozice](lessons/lesson-13/README.md) · ~85 min
- [Lekce 14 — Errors: hodnoty, wrapping, Is a As](lessons/lesson-14/README.md) · ~90 min
- [Lekce 15 — Generics](lessons/lesson-15/README.md) · ~85 min
- [Lekce 16 — JSON: marshal, unmarshal, tagy](lessons/lesson-16/README.md) · ~90 min
- [Lekce 17 — Testování a projekt P01 (CSV CLI)](lessons/lesson-17/README.md) · ~90 min
- [Lekce 18 — Checkpoint fáze 1 **(checkpoint)**](lessons/lesson-18/README.md) · ~90 min

## Fáze 2 — Idiomatický Go

- [Lekce 19 — Jména, balíčky a struktura kódu](lessons/lesson-19/README.md) · ~85 min
- [Lekce 20 — Konstruktory, inicializace a design API](lessons/lesson-20/README.md) · ~90 min
- [Lekce 21 — Error handling v review](lessons/lesson-21/README.md) · ~85 min
- [Lekce 22 — Čtení stdlib a cizího kódu](lessons/lesson-22/README.md) · ~85 min
- [Lekce 23 — Checkpoint fáze 2: PHP zápachy a AI garbage **(checkpoint)**](lessons/lesson-23/README.md) · ~90 min

## Fáze 3 — net/http a tooling

- [Lekce 24 — net/http od nuly: handler, ServeMux, httptest](lessons/lesson-24/README.md) · ~90 min
- [Lekce 25 — Routing: metody, wildcardy, PathValue](lessons/lesson-25/README.md) · ~90 min
- [Lekce 26 — Middleware](lessons/lesson-26/README.md) · ~90 min
- [Lekce 27 — context v request scope](lessons/lesson-27/README.md) · ~90 min
- [Lekce 28 — Konfigurace z prostředí](lessons/lesson-28/README.md) · ~90 min
- [Lekce 29 — slog: strukturované logování](lessons/lesson-29/README.md) · ~90 min
- [Lekce 30 — HTTP klient, timeouty a graceful shutdown](lessons/lesson-30/README.md) · ~90 min
- [Lekce 31 — Checkpoint fáze 3 + projekt P02 **(checkpoint)**](lessons/lesson-31/README.md) · ~90 min

## Fáze 4 — Architektura v Go

- [Lekce 32 — Project layout a `internal/`](lessons/lesson-32/README.md) · ~90 min
- [Lekce 33 — Porty a adaptéry, interface u konzumenta](lessons/lesson-33/README.md) · ~90 min
- [Lekce 34 — Doménové typy: Money a value objekty](lessons/lesson-34/README.md) · ~90 min
- [Lekce 35 — Persistence: repozitář, in-memory fake, SQL mindset](lessons/lesson-35/README.md) · ~90 min
- [Lekce 36 — Validace na hranici](lessons/lesson-36/README.md) · ~90 min
- [Lekce 37 — Autentizace a observabilita](lessons/lesson-37/README.md) · ~90 min
- [Lekce 38 — Projekt P03: hexagonální služba](lessons/lesson-38/README.md) · ~90 min
- [Lekce 39 — Checkpoint fáze 4 **(checkpoint)**](lessons/lesson-39/README.md) · ~90 min

## Fáze 5 — Concurrency do hloubky

- [Lekce 40 — Goroutiny, WaitGroup a leaky](lessons/lesson-40/README.md) · ~90 min
- [Lekce 41 — Kanály a ownership](lessons/lesson-41/README.md) · ~90 min
- [Lekce 42 — `select` a timeouty](lessons/lesson-42/README.md) · ~90 min
- [Lekce 43 — Mutex vs kanál](lessons/lesson-43/README.md) · ~90 min
- [Lekce 44 — Race lab: detektor závodů](lessons/lesson-44/README.md) · ~90 min
- [Lekce 45 — Pipelines](lessons/lesson-45/README.md) · ~90 min
- [Lekce 46 — Worker pool](lessons/lesson-46/README.md) · ~90 min
- [Lekce 47 — errgroup a rušení přes context](lessons/lesson-47/README.md) · ~90 min
- [Lekce 48 — Paměťový model a happens-before](lessons/lesson-48/README.md) · ~90 min
- [Lekce 49 — Scheduler: mentální model G-M-P](lessons/lesson-49/README.md) · ~90 min
- [Lekce 50 — Checkpoint fáze 5 + projekt P04 **(checkpoint)**](lessons/lesson-50/README.md) · ~90 min

## Fáze 6 — Production Go

- [Lekce 51 — Moduly, verzování a zranitelnosti](lessons/lesson-51/README.md) · ~90 min
- [Lekce 52 — Testování do hloubky: benchmarky, fuzz, golden files](lessons/lesson-52/README.md) · ~90 min
- [Lekce 53 — pprof a profilování](lessons/lesson-53/README.md) · ~90 min
- [Lekce 54 — Generics v API, reflexe a build tagy](lessons/lesson-54/README.md) · ~90 min
- [Lekce 55 — Checkpoint fáze 6: kontejnery, health a production checklist **(checkpoint)**](lessons/lesson-55/README.md) · ~90 min

## Fáze 7 — Inženýrství v době AI

- [Lekce 56 — Spec-first, ADR a prompting pro Go](lessons/lesson-56/README.md) · ~90 min
- [Lekce 57 — Strukturované review AI kódu a diff lab](lessons/lesson-57/README.md) · ~90 min
- [Lekce 58 — Osobní checklist, pairing protokol a manual rewrite](lessons/lesson-58/README.md) · ~90 min

## Fáze 8 — Capstone

- [Lekce 59 — Capstone P05: spec a implementace](lessons/lesson-59/README.md) · ~90 min
- [Lekce 60 — Checkpoint závěrečný: hardening, retrospektiva a export kurzu **(checkpoint)**](lessons/lesson-60/README.md) · ~90 min

## Projekty

| ID | Projekt | Cesta | Zadává |
|----|---------|-------|--------|
| P01 | CSV CLI | [`projects/p01-csv-cli`](projects/p01-csv-cli/ACCEPTANCE.md) | lekce 17 |
| P02 | REST API | [`projects/p02-http-api`](projects/p02-http-api/ACCEPTANCE.md) | lekce 31 |
| P03 | Hexagonální služba | [`projects/p03-hex-service`](projects/p03-hex-service/ACCEPTANCE.md) | lekce 38 |
| P04 | Worker pool | [`projects/p04-worker-pool`](projects/p04-worker-pool/ACCEPTANCE.md) | lekce 50 |
| P05 | Capstone | [`projects/p05-capstone`](projects/p05-capstone/ACCEPTANCE.md) | lekce 59–60 |

## AI režim podle fází

| Lekce | Režim |
|-------|-------|
| 01–18 | `ZAKÁZÁNO` |
| 19–23 | `JEN VYSVĚTLENÍ` |
| 24–39 | `BOILERPLATE OK` |
| 40–55 | `JUNIOR POD REVIEW` |
| 56–60 | `TECH LEAD` |

Detail: [docs/ai-playbook.md](docs/ai-playbook.md).
