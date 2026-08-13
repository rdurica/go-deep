# Lekce NN — Název

> **Čas:** ~90 min · **Fáze:** N — Název fáze · **AI režim:** `ZAKÁZÁNO`

## Co budeš umět

- …
- …
- …

## Teorie

### Podsekce 1

Vysvětlení s funkčním příkladem. Každý blok kódu musí jít zkopírovat a spustit.

### Podsekce 2

…

## Rozdíly proti PHP

Krátké srovnání konkrétního PHP/Symfony vzoru s Go protějškem. Ne tabulka pojmů,
ale ukázka kódu vedle sebe a věta o tom, co se mění v uvažování.

## Časté chyby

| Chyba | Proč vzniká | Jak to udělat správně |
|-------|-------------|------------------------|
| … | … | … |

## AI kvíz

Po přečtení teorie spusť v Cursoru **`/go-deep-quiz NN`**. AI tě ~5 minut prověří
mentální model (ne hotové cvičení). Slabiny si uloží do [`GAPS.md`](../../GAPS.md).

## Úkol

Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí. Stupně jdou od jednodušších
ke složitějším — po každém stupni spusť review, než jdeš dál.

Pevný default: 1 = find-the-bug / complete-the-gap, 2 = greenfield 1–2 funkce,
3 = jedna věc úsudku. Etalon: [lekce 08](../lessons/lesson-08/README.md).

### Jednoduchý

Oprav: `…` (find-the-bug — kód je záměrně vadný)

```bash
make lesson L=NN PART=1
```

Pak **`/go-deep-review NN easy`**.

### Střední

Implementuj: `…` (krátký greenfield, 1–2 funkce)

```bash
make lesson L=NN PART=2
```

Pak **`/go-deep-review NN medium`**.

### Obtížný

Doplň: `…` (jedna věc úsudku: edge / `map[K]*V` / race / leak)

```bash
make lesson L=NN PART=3
```

Pak **`/go-deep-review NN hard`** (nebo rovnou `final`, pokud checklist říká jinak).

Až budou všechny stupně hotové, porovnej se `solutions/` (spoiler).

## Závěrečné otázky

Spusť **`/go-deep-review NN final`**. AI projde body níže, doptá se a ověří pochopení.
Celé cvičení musí projít (`make lesson L=NN`); review mezitím kontrolovalo jen stupně.

- [ ] Umíš vysvětlit: …
- [ ] Umíš vysvětlit: …
- [ ] Umíš vysvětlit: …

## AI režim

`ZAKÁZÁNO` — viz [docs/ai-playbook.md](../../docs/ai-playbook.md).

Mentor, kvíz i review (dialog) jsou vždy OK; v tomto režimu AI nesmí psát kód cvičení.

## Další čtení

1. …
2. …
