# Mapa lekcí

Kurz má **60 lekcí**, z nichž každá je plnohodnotná **60–90 minut**: teorie s funkčními
příklady, PHP → Go most, tři stupňované úkoly (rozcvička → jádro → rozšíření) a otázky
k sebehodnocení.

Sedm lekcí jsou **checkpointy**: nemají novou látku, ale kumulativní cvičení přes celou
fázi a sebehodnotící rubriku.

| Lekce | Téma |
|-------|------|
| **Fáze 0 — Setup a mentální reset** | |
| 01 | Toolchain, moduly a workspace |
| 02 | Mentální reset: PHP → Go |
| **Fáze 1 — Jazyk a paměťový model** | |
| 03 | Typy, zero values a konstanty |
| 04 | Funkce, multiple returns, closures |
| 05 | Structs, metody a embedding |
| 06 | Pointery: hodnota vs reference |
| 07 | Slices: pole, append, internals, aliasing |
| 08 | Mapy |
| 09 | Stringy, runy a byty |
| 10 | defer, panic, recover |
| 11 | Balíčky, export a viditelnost |
| 12 | Interfaces I — implicitní implementace |
| 13 | Interfaces II — io.Reader, io.Writer, kompozice |
| 14 | Errors — hodnoty, wrapping, Is/As |
| 15 | Generics |
| 16 | JSON — marshal, unmarshal, tagy |
| 17 | Testování a projekt P01 (CSV CLI) |
| 18 | **Checkpoint fáze 1** |
| **Fáze 2 — Idiomatický Go** | |
| 19 | Jména, balíčky a struktura kódu |
| 20 | Konstruktory, inicializace a nulová hodnota API |
| 21 | Error handling v review — texty, obalování, ignorované chyby |
| 22 | Čtení stdlib a cizího kódu |
| 23 | **Checkpoint fáze 2** — PHP zápachy a AI garbage |
| **Fáze 3 — net/http a tooling** | |
| 24 | net/http od nuly — handler, ServeMux, httptest |
| 25 | Routing: metody, wildcardy, PathValue |
| 26 | Middleware |
| 27 | context v request scope |
| 28 | Konfigurace z prostředí |
| 29 | slog — strukturované logování |
| 30 | HTTP klient, timeouty a graceful shutdown |
| 31 | **Checkpoint fáze 3** + projekt P02 (REST API) |
| **Fáze 4 — Architektura v Go** | |
| 32 | Project layout a `internal/` |
| 33 | Porty a adaptéry, interface u konzumenta |
| 34 | Doménové typy — Money a value objekty |
| 35 | Persistence — repozitář, in-memory fake, SQL mindset |
| 36 | Validace na hranici |
| 37 | Autentizace a observabilita |
| 38 | Projekt P03 — hexagonální služba |
| 39 | **Checkpoint fáze 4** |
| **Fáze 5 — Concurrency do hloubky** | |
| 40 | Goroutiny, WaitGroup a leaky |
| 41 | Kanály a ownership |
| 42 | select a timeouty |
| 43 | Mutex vs kanál |
| 44 | Race lab — detektor závodů |
| 45 | Pipelines |
| 46 | Worker pool |
| 47 | errgroup a rušení přes context |
| 48 | Paměťový model a happens-before |
| 49 | Scheduler — mentální model G-M-P |
| 50 | **Checkpoint fáze 5** + projekt P04 (worker pool, backpressure) |
| **Fáze 6 — Production Go** | |
| 51 | Moduly, verzování a zranitelnosti |
| 52 | Testování do hloubky — benchmarky, fuzz, golden files |
| 53 | pprof a profilování |
| 54 | Generics v API, reflexe a build tagy |
| 55 | **Checkpoint fáze 6** — kontejnery, health, production checklist |
| **Fáze 7 — Inženýrství v době AI** | |
| 56 | Spec-first, ADR a prompting pro Go |
| 57 | Strukturované review AI kódu a diff lab |
| 58 | Osobní checklist, pairing protokol, manual rewrite |
| **Fáze 8 — Capstone** | |
| 59 | Capstone P05 — spec a implementace |
| 60 | **Checkpoint závěrečný** — hardening, retrospektiva, export kurzu |

## Časový rozpočet

| | |
|---|---|
| Lekcí | 60 |
| Délka lekce | 60–90 min (medián ~80) |
| Celkem | ~80 hodin |
| Tempo 3 lekce/týden | ~20 týdnů |
| Tempo 5 lekcí/týden | ~12 týdnů |

## Struktura lekce

Každá lekce má stejnou kostru — viz [templates/lesson-README.md](../templates/lesson-README.md):

1. **Co budeš umět** — 3–4 konkrétní schopnosti, ne témata
2. **PHP → Go most** — stejný problém v obou jazycích vedle sebe
3. **Teorie** — 2–5 podsekcí, každý příklad spustitelný
4. **Časté chyby** — tabulka chyba / proč / správně
5. **Úkol A/B/C** — rozcvička ~10 min, jádro ~35 min, rozšíření ~20–25 min
6. **Ověření** — checklist a otázky k sebehodnocení
7. **AI režim** a **Další čtení**
