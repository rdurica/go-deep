#!/usr/bin/env python3
"""Vygeneruje BOOK.md a PROGRESS.md z hlaviček lekcí.

Jediný zdroj pravdy jsou soubory lessons/lesson-NN/README.md. Z každého se čte
nadpis první úrovně a meta řádek pod ním:

    # Lekce 03 — Typy, zero values a konstanty

    > **Čas:** ~90 min · **Fáze:** 1 — Jazyk a paměťový model · **AI režim:** `ZAKÁZÁNO`
"""

from __future__ import annotations

import re
import sys
from dataclasses import dataclass
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
LESSONS = ROOT / "lessons"

TITLE_RE = re.compile(r"^#\s*Lekce\s+(\d+)\s*[—–-]\s*(.+?)\s*$", re.MULTILINE)
META_RE = re.compile(
    r"^>\s*\*\*Čas:\*\*\s*(?P<time>[^·]+?)\s*·"
    r"\s*\*\*Fáze:\*\*\s*(?P<phase>[^·]+?)\s*·"
    r"\s*\*\*AI režim:\*\*\s*`(?P<ai>[^`]+)`",
    re.MULTILINE,
)

PROJECTS = [
    ("P01", "CSV CLI", "projects/p01-csv-cli", "lekce 17"),
    ("P02", "REST API", "projects/p02-http-api", "lekce 31"),
    ("P03", "Hexagonální služba", "projects/p03-hex-service", "lekce 38"),
    ("P04", "Worker pool", "projects/p04-worker-pool", "lekce 50"),
    ("P05", "Capstone", "projects/p05-capstone", "lekce 59–60"),
]

CHECKPOINTS = {18, 23, 31, 39, 50, 55, 60}


@dataclass
class Lesson:
    number: int
    slug: str
    title: str
    time: str
    phase: str
    ai: str

    @property
    def link(self) -> str:
        return f"lessons/{self.slug}/README.md"


def load() -> list[Lesson]:
    lessons: list[Lesson] = []
    for readme in sorted(LESSONS.glob("lesson-*/README.md")):
        text = readme.read_text(encoding="utf-8")
        title_m = TITLE_RE.search(text)
        meta_m = META_RE.search(text)
        if not title_m:
            print(f"varování: {readme} nemá rozpoznatelný nadpis", file=sys.stderr)
            continue
        if not meta_m:
            print(f"varování: {readme} nemá rozpoznatelný meta řádek", file=sys.stderr)
            continue
        lessons.append(
            Lesson(
                number=int(title_m.group(1)),
                slug=readme.parent.name,
                title=title_m.group(2).strip(),
                time=meta_m.group("time").strip(),
                phase=meta_m.group("phase").strip(),
                ai=meta_m.group("ai").strip(),
            )
        )
    lessons.sort(key=lambda x: x.number)
    return lessons


def group_by_phase(lessons: list[Lesson]) -> list[tuple[str, list[Lesson]]]:
    groups: list[tuple[str, list[Lesson]]] = []
    for lesson in lessons:
        if not groups or groups[-1][0] != lesson.phase:
            groups.append((lesson.phase, []))
        groups[-1][1].append(lesson)
    return groups


def label(lesson: Lesson) -> str:
    suffix = " **(checkpoint)**" if lesson.number in CHECKPOINTS else ""
    return f"Lekce {lesson.number:02d} — {lesson.title}{suffix}"


def render_book(lessons: list[Lesson]) -> str:
    out = [
        "# Go do hloubky — obsah",
        "",
        f"{len(lessons)} lekcí po 60–90 minutách. Každá lekce má teorii, most z PHP,",
        "tři stupňované úkoly a testy: [`lessons/lesson-NN/README.md`](lessons/lesson-01/README.md),",
        "cvičení v `exercise/`, referenční řešení v `solutions/`.",
        "",
        "Osnova lekcí: [course/lesson-map.md](course/lesson-map.md).",
        "",
    ]
    for phase, items in group_by_phase(lessons):
        out.append(f"## Fáze {phase}")
        out.append("")
        for lesson in items:
            out.append(f"- [{label(lesson)}]({lesson.link}) · {lesson.time}")
        out.append("")

    out.append("## Projekty")
    out.append("")
    out.append("| ID | Projekt | Cesta | Zadává |")
    out.append("|----|---------|-------|--------|")
    for pid, name, path, when in PROJECTS:
        out.append(f"| {pid} | {name} | [`{path}`]({path}/ACCEPTANCE.md) | {when} |")
    out.append("")

    out.append("## AI režim podle fází")
    out.append("")
    out.append("| Lekce | Režim |")
    out.append("|-------|-------|")
    seen: list[tuple[str, int, int]] = []
    for lesson in lessons:
        if seen and seen[-1][0] == lesson.ai:
            seen[-1] = (seen[-1][0], seen[-1][1], lesson.number)
        else:
            seen.append((lesson.ai, lesson.number, lesson.number))
    for ai, start, end in seen:
        span = f"{start:02d}" if start == end else f"{start:02d}–{end:02d}"
        out.append(f"| {span} | `{ai}` |")
    out.append("")
    out.append("Detail: [docs/ai-playbook.md](docs/ai-playbook.md).")
    out.append("")
    return "\n".join(out)


def render_progress(lessons: list[Lesson]) -> str:
    out = [
        "# Postup",
        "",
        "Odškrtávej až po úspěšném **`/go-deep-review NN final`** (skill odškrtne za tebe).",
        "Checkpointy neodškrtávej, dokud nemáš vyplněné sebehodnocení. Slabiny AI zapisuje",
        "do [`GAPS.md`](GAPS.md).",
        "",
    ]
    for phase, items in group_by_phase(lessons):
        out.append(f"## Fáze {phase}")
        out.append("")
        for lesson in items:
            out.append(f"- [ ] [{label(lesson)}]({lesson.link})")
        out.append("")

    out.append("## Projekty")
    out.append("")
    for pid, name, path, when in PROJECTS:
        out.append(f"- [ ] {pid} — {name} ([`{path}`]({path}/ACCEPTANCE.md), {when})")
    out.append("")
    return "\n".join(out)


def main() -> int:
    lessons = load()
    if not lessons:
        print("nenašel jsem žádné lekce", file=sys.stderr)
        return 1

    (ROOT / "BOOK.md").write_text(render_book(lessons), encoding="utf-8")
    (ROOT / "PROGRESS.md").write_text(render_progress(lessons), encoding="utf-8")

    missing = sorted(set(range(1, 61)) - {x.number for x in lessons})
    print(f"vygenerováno pro {len(lessons)} lekcí")
    if missing:
        print(f"chybí lekce: {', '.join(f'{n:02d}' for n in missing)}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
