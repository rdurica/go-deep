# Moduly kurzu

Mapování na fáze knížky ([BOOK.md](../BOOK.md)). Rozpad lekcí:
[lesson-map.md](lesson-map.md).

## M0 — Setup a mentální reset (lekce 01–02)

**Výstupy:** toolchain a moduly bez googlení, tři body mentálního resetu PHP → Go.  
**AI:** zakázán codegen.  
**Rozsah:** ~2,5 h.

## M1 — Jazyk a paměťový model (lekce 03–18)

**Výstupy:** zero values, hodnota vs pointer, slice internals a aliasing, mapy, runy,
defer, hranice balíčků, interfaces, error wrapping, generika, JSON, table-driven testy.  
**Projekt:** P01 (CSV CLI).  
**Checkpoint:** lekce 18.  
**AI:** zakázán codegen.  
**Rozsah:** ~22 h.

## M2 — Idiomatický Go (lekce 19–23)

**Výstupy:** pojmenování a struktura, konstruktory a design API, error handling v review,
čtení stdlib, rozpoznání PHP zápachů a AI garbage.  
**Checkpoint:** lekce 23.  
**AI:** jen vysvětlení.  
**Rozsah:** ~7 h.

## M3 — HTTP bez frameworku (lekce 24–31)

**Výstupy:** `http.Handler` a httptest, routing `ServeMux` (metody, wildcardy), middleware, context, konfigurace
z prostředí, `log/slog`, HTTP klient s timeouty, graceful shutdown.  
**Projekt:** P02 (REST API).  
**Checkpoint:** lekce 31.  
**AI:** boilerplate OK.  
**Rozsah:** ~11 h.

## M4 — Architektura (lekce 32–39)

**Výstupy:** layout a `internal/`, porty a adaptéry s interfacem u konzumenta, doménové
value objekty, repozitáře a SQL mindset, validace na hranici, auth a observabilita.  
**Projekt:** P03 (hexagonální služba).  
**Checkpoint:** lekce 39.  
**AI:** boilerplate OK.  
**Rozsah:** ~11 h.

## M5 — Concurrency (lekce 40–50)

**Výstupy:** goroutiny bez leaků, ownership kanálů, select a timeouty, mutex vs kanál,
race detektor, pipelines, worker pool, errgroup, paměťový model, scheduler G-M-P.  
**Projekt:** P04 (worker pool s backpressure).  
**Checkpoint:** lekce 50.  
**AI:** junior pod review.  
**Rozsah:** ~15 h.

## M6 — Production (lekce 51–55)

**Výstupy:** moduly a verzování, govulncheck, benchmarky a fuzz a golden files, pprof,
generika v API a reflexe a build tagy, kontejnery a health probes.  
**Checkpoint:** lekce 55.  
**AI:** junior pod review.  
**Rozsah:** ~7 h.

## M7 — Inženýrství v době AI (lekce 56–58)

**Výstupy:** spec-first a ADR, prompting pro Go, strukturované review AI diffu, osobní
checklist a pairing protokol.  
**AI:** tech lead.  
**Rozsah:** ~4,5 h.

## M8 — Capstone (lekce 59–60)

**Výstupy:** P05 postavený hybridně (spec člověk → implementace agent → review člověk),
hardening, retrospektiva, export vlastního kurzu.  
**Checkpoint:** lekce 60.  
**Rozsah:** ~3 h + samotný projekt.

## Bonus — P06 produkční backend (po lekci 60)

**Výstupy:** stejná doména a ServeMux API jako P05, ale Postgres (sqlc + pgx), Redis
cache, migrace a readiness na závislosti. Samostatný `go.mod`, bez Chi.  
**Projekt:** [`projects/p06-bookmarks-persist`](../projects/p06-bookmarks-persist/README.md).  
**AI:** tech lead.  
**Rozsah:** ~3–5 h *(volitelné)*.
