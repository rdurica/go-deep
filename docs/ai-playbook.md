# AI playbook — Go v době agentů

Cíl není zakázat AI navždy. Cíl je **nejdřív vybudovat úsudek**, pak AI řídit jako juniora.

## Režimy podle fází

### ZAKÁZÁNO (lekce 01–18)

- AI nesmí generovat ani doplňovat kód cvičení.
- Po vlastním pokusu smí vysvětlit koncept („proč `append` někdy alokuje znovu?“).
- Zakázané: „napiš mi řešení lekce 7“, Copilot autocomplete na celou funkci.

### JEN VYSVĚTLENÍ (lekce 19–23)

Povolené promptování:

> Vysvětli, proč je tento Go kód neidiomatický z pohledu Code Review Comments. Nenabízej přepsaný kód, jen body.

Zakázané: „přepiš to idiomaticky za mě“ (dokud neuděláš vlastní verzi).

### BOILERPLATE OK (lekce 24–39)

AI smí:

- JSON tagy, DTO structy (až po tom, co jsi structy psal ručně ≥3×),
- repetitivní wiring, test table scaffolding,
- `go generate` / drobnosti.

Ty vlastníš:

- hranice balíčků, error model, concurrency, API design, acceptance testy.

### JUNIOR POD REVIEW (lekce 40–55)

Workflow:

1. Napiš spec + acceptance testy.
2. Nech agenta navrhnout implementaci.
3. Projdi checklist níže.
4. Přepiš kritickou cestu, pokud „funguje, ale smrdí“.

### TECH LEAD (lekce 56–60)

Spec → agent → structured review → ADR rozhodnutí → capstone. Ty jsi architekt a reviewer.

## Review checklist (Go od AI)

Označ u každého PR / diffu:

- [ ] Interfacy jsou malé a definované u **konzumenta**
- [ ] Žádné `_ = err` / ignorované chyby
- [ ] Wrapping s `%w` a užitečným kontextem
- [ ] Value vs pointer receiver má důvod
- [ ] `context.Context` je první parametr; není uložený ve structu
- [ ] Goroutines mají jasný lifetime (cancel / WaitGroup / errgroup)
- [ ] Žádný unbounded fan-out / neomezený buffer bez důvodu
- [ ] Package ≠ PHP layer cake (`service/`, `repository/` jako dogma)
- [ ] Žádný panický control-flow místo `error`
- [ ] Testy table-driven; concurrent kód ověřený `-race`
- [ ] Exportované identifikátory mají doc comment
- [ ] Žádný zbytečný DI framework / reflection magie

## Dobré promptovací vzorce pro Go

```
Piš idiomatický Go 1.26+, stdlib first, bez web frameworku.
Accept interfaces, return structs.
Errors wrap with %w. No panic for business errors.
Packages named by domain, not by layer.
Include table-driven tests.
Explain package boundaries in 3 bullets before coding.
```

## Anti-patterny „AI Go“

1. Obří interfacy 1:1 se Symfony service.
2. `utils` / `common` / `helpers` balíčky.
3. Pointer všude „pro jistotu“.
4. Context ve struct fieldu.
5. Goroutine bez cancelace.
6. Framework dřív, než umíš `net/http`.
