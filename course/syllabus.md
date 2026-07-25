# Syllabus — Go do hloubky

## Pro koho

Vývojáři se silným PHP/Symfony (nebo obdobným) zázemím, kteří potřebují pevný mentální
model Go a chtějí umět řídit AI při psaní Go kódu. Předpoklad: OOP, testy, HTTP, SQL,
základy provozu. Předchozí zkušenost s Go se nevyžaduje.

## Výstupy absolventa

1. Píše idiomatický Go bez framework-first reflexu a bez layer-cake balíčků.
2. Rozumí paměťovému modelu natolik, že u slice, mapy a pointeru předem ví, co se stane.
3. Postaví HTTP službu jen na `net/http` včetně middlewaru, kontextu a graceful shutdownu.
4. Navrhne hranice balíčků a ports/adapters bez zbytečné abstrakce.
5. Odhalí race, goroutine leak a špatný error handling v code review — včetně AI výstupu.
6. Má osobní AI playbook: kdy codegen ano, kdy ne, a jak výsledek reviewovat.
7. Dokáže z materiálu vést nebo publikovat vlastní kurz.

## Forma

- Self-paced repozitář: [BOOK.md](../BOOK.md) + `lessons/`
- **60 lekcí × 60–90 min ≈ 80 hodin** plus čas na projekty
- Cohort varianta: 12–16 týdnů při 6–8 h/týden
- Sedm checkpointů se sebehodnotící rubrikou

## Moduly

Viz [modules.md](modules.md), rozpad na lekce viz [lesson-map.md](lesson-map.md).

## Hodnocení

| Artefakt | Váha (doporučení) |
|----------|-------------------|
| Checkpointy 18 / 23 / 31 / 39 / 50 / 55 | 20 % |
| Projekty P01–P04 | 40 % |
| P05 capstone + AI review report | 30 % |
| Závěrečný checkpoint 60 a retrospektiva | 10 % |

Každý projekt má `ACCEPTANCE.md` s odškrtávacími kritérii. Zelené testy jsou nutná,
ne postačující podmínka — hodnotí se i návrh hranic a čitelnost.

## Pravidla AI během kurzu

Závazná politika: [docs/ai-playbook.md](../docs/ai-playbook.md).

| Lekce | Režim |
|-------|-------|
| 01–18 | `ZAKÁZÁNO` |
| 19–23 | `JEN VYSVĚTLENÍ` |
| 24–39 | `BOILERPLATE OK` |
| 40–55 | `JUNIOR POD REVIEW` |
| 56–60 | `TECH LEAD` |

Porušení no-codegen v modulu M1 znamená lekci zopakovat. Není to trest — bez vlastního
pokusu se mentální model nevytvoří a checkpoint to stejně odhalí.
