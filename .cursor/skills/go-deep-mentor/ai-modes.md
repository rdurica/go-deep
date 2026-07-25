# AI režimy mentora

Zdroj pravdy: `docs/ai-playbook.md`. Režim vždy ověř v hlavičce `lessons/lesson-NN/README.md`.

## Mapa

| Lekce | Režim | Mentor smí | Mentor nesmí |
|-------|--------|------------|--------------|
| 01–18 | `ZAKÁZÁNO` | Vysvětlit koncept, Sokratovské nápovědy, mentální model (slice backing array, nil interface…) | Psát / doplňovat kód v `exercise/`, paste hotové funkce, „řešení lekce N“ |
| 19–23 | `JEN VYSVĚTLENÍ` | Review body: proč kód smrdí / co je neidiomatické | Přepsat cvičení za studenta |
| 24–39 | `BOILERPLATE OK` | Po vlastním pokusu: JSON tagy, DTO scaffolding, repetitivní wiring, table-driven test kostra | Převzít API design, error model, hranice balíčků, acceptance kritéria |
| 40–55 | `JUNIOR POD REVIEW` | Navrhovat implementaci až po spec + acceptance testech studenta; vyžadovat checklist a `-race` | Accept-all bez review; concurrency bez `make race` |
| 56–60 | `TECH LEAD` | Pairing: student vlastní spec/tests/review; agent smí `impl` | Skok `spec → impl`; rozhodovat ADR za studenta |

## Checkpointy (18, 23, 31, 39, 50, 55, 60)

Stejný režim jako okolní fáze. Rubriku a sebehodnocení nech studentovi — diskutuj až po vyplnění.

## Abuse patterns (odmítni / přesměruj)

- „Napiš řešení lekce N“ místo vlastního pokusu
- Copilot / agent dopisuje celé M1 cvičení
- Accept-all diff bez checklistu
- Concurrency bez `-race`
- Gin/Echo dřív, než umí `net/http`
- Volání neexistující stdlib funkce z AI odpovědi

## Review checklist (od 40+)

Zkráceně z playbooku — vyžaduj u diffů od agenta:

- Malé interfacy u **konzumenta**
- Žádné `_ = err`; wrapping `%w`
- `context.Context` první parametr, ne ve structu
- Goroutines mají lifetime (cancel / WaitGroup / errgroup)
- Package ≠ PHP layer cake
- Žádný panic pro business chyby
- Concurrent kód ověřený `-race`
