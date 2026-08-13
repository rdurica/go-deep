# Jak z repozitáře udělat kurz

Po dokončení lekcí 01–60 je materiál v `course/` připravený k vedení nebo publikaci.

## Co už máš

| Artefakt | Účel |
|----------|------|
| [BOOK.md](../BOOK.md) | osnova 60 lekcí, generovaná z hlaviček |
| `lessons/*` | lekce, cvičení, testy a referenční řešení |
| `projects/*` | P01–P05 s akceptačními kritérii (+ volitelný bonus P06) |
| [course/syllabus.md](../course/syllabus.md) | kurzová forma a hodnocení |
| [course/modules.md](../course/modules.md) | moduly a learning outcomes |
| [course/lesson-map.md](../course/lesson-map.md) | rozpad lekcí a časový rozpočet |
| [course/facilitator-notes.md](../course/facilitator-notes.md) | pasty PHP→Go, kde se studenti sekají |
| [docs/ai-playbook.md](ai-playbook.md) | AI režimy a review checklist |
| [docs/authoring.md](authoring.md) | jak psát další lekce ve stejné kvalitě |

## Formáty kurzu

1. **Self-paced** — student si založí vlastní repo přes **Use this template**
   (GitHub: *Settings → Template repository* musí být zapnuté u upstreamu),
   odškrtává [`PROGRESS.md`](../PROGRESS.md) u sebe a commituje progress.
   Fork je varianta pro častý sync s upstreamem. Holý clone bez vlastní remote
   kopie progress neuchová.
2. **Cohort (8–12 týdnů)** — 4–5 lekcí týdně plus jedno live review; checkpointy
   jako společná sezení.
3. **Workshop track (3 dny)** — jen projekty P01–P05 a AI review lab (lekce 56–58);
   předpokládá, že účastníci Go už trochu umí.
4. **Firemní onboarding** — moduly M0–M3 jako povinný základ, zbytek volitelně.
5. **Bonus track po capstone** — P06 (Postgres + sqlc + Redis) pro ty, kdo chtějí
   produkční adaptéry nad porty z P05; viz
   [`projects/p06-bookmarks-persist`](../projects/p06-bookmarks-persist/README.md).

## Hodnocení

- Checkpointy 18, 23, 31, 39, 50, 55, 60 — každý má bodovanou rubriku přímo v lekci
- Projekty P01–P05 podle `ACCEPTANCE.md` (P06 volitelně)
- Závěrečný AI review: student odevzdá diff, vyplněný checklist a ADR se zdůvodněním

## Časová mapa

| | |
|---|---|
| 60 lekcí × 30–45 min | ~40 hodin |
| Projekty nad rámec lekcí | ~20 hodin |
| Při 5–7 h/týden | 8–12 týdnů |
| Při 4 h/týden | ~15 týdnů |

## Než to zveřejníš

- [ ] `make check` je zelené
- [ ] `python3 scripts/generate_index.py` neohlásí chybějící lekci
- [ ] Odkazy v README a BOOK.md sedí
- [ ] Repo je na GitHubu označené jako **Template repository**
- [ ] Rozhodl jsi, jestli `solutions/` zůstanou veřejné

Pokud vedeš placený kurz, referenční řešení můžeš držet v samostatné větvi nebo
privátním repozitáři a studentům je uvolňovat po odevzdání. `make solutions` v CI
pak poběží jen u tebe.

## Licence obsahu

Uprav si `LICENSE` v kořeni před publikací. Cvičení, testy a rubriky jsou tvoje učební IP.
