#!/usr/bin/env python3
"""Migrate all lessons to tiered flow: tiers.txt, stub markers, README structure."""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
LESSONS = ROOT / "lessons"
CHECKPOINTS = {18, 23, 31, 39, 50, 55, 60}

TIER_LABELS = ("jednoduchý", "střední", "obtížný")
TIER_MARKERS = tuple(f"// --- Stupeň: {lab} ---" for lab in TIER_LABELS)

FUNC_RE = re.compile(
    r"^func\s+(?:\([^)]*\)\s+)?([A-Z][A-Za-z0-9_]*)\s*\(",
    re.M,
)
TEST_RE = re.compile(r"^func\s+(Test[A-Za-z0-9_]*)\s*\(", re.M)
SECTION_RE = re.compile(r"(?m)^(## .+)$")


def split_thirds(items: list[str]) -> tuple[list[str], list[str], list[str]]:
    n = len(items)
    if n == 0:
        return [], [], []
    if n == 1:
        return items, [], []
    if n == 2:
        return items[:1], items[1:], []
    a = max(1, n // 3)
    b = max(1, (n - a) // 2)
    # ensure third gets remainder and at least 1 when n>=3
    c = n - a - b
    if c == 0:
        b -= 1
        c = 1
    return items[:a], items[a : a + b], items[a + b :]


def collect_tests(exercise_dir: Path) -> list[str]:
    names: list[str] = []
    for path in sorted(exercise_dir.rglob("*_test.go")):
        text = path.read_text(encoding="utf-8")
        for m in TEST_RE.finditer(text):
            name = m.group(1)
            if name not in names:
                names.append(name)
    return names


def collect_exported_funcs(go_file: Path) -> list[tuple[int, str]]:
    """Return (line_index_0based, func_name) for exported funcs in order."""
    text = go_file.read_text(encoding="utf-8")
    lines = text.splitlines(keepends=True)
    out: list[tuple[int, str]] = []
    # Walk with simple state: match func at line start
    for i, line in enumerate(lines):
        m = re.match(r"^func\s+(?:\([^)]*\)\s+)?([A-Z][A-Za-z0-9_]*)\s*\(", line)
        if m:
            out.append((i, m.group(1)))
    return out


def assign_tests_to_func_tiers(
    tests: list[str], f1: list[str], f2: list[str], f3: list[str]
) -> tuple[list[str], list[str], list[str]]:
    """Put each test into the tier of the best-matching exported func name."""
    ranked = sorted(set(f1 + f2 + f3), key=len, reverse=True)
    tier_of = {n: 0 for n in f1} | {n: 1 for n in f2} | {n: 2 for n in f3}
    buckets: list[list[str]] = [[], [], []]
    unmatched: list[str] = []
    for t in tests:
        rest = t[4:] if t.startswith("Test") else t  # strip Test
        hit = None
        for fn in ranked:
            # TestFoo, TestFooBar, TestTypeFoo… — longest func name that fits
            if rest == fn or rest.startswith(fn) or fn in rest:
                hit = fn
                break
        if hit is None:
            unmatched.append(t)
            continue
        buckets[tier_of[hit]].append(t)
    u1, u2, u3 = split_thirds(unmatched)
    for ti, group in enumerate((u1, u2, u3)):
        buckets[ti].extend(group)
    if tests and not buckets[0] and buckets[1]:
        buckets[0].append(buckets[1].pop(0))
    if tests and not buckets[2] and buckets[1]:
        buckets[2].append(buckets[1].pop())
    if tests and not buckets[1]:
        if len(buckets[0]) > 1:
            buckets[1].append(buckets[0].pop())
        elif len(buckets[2]) > 1:
            buckets[1].append(buckets[2].pop(0))
    return buckets[0], buckets[1], buckets[2]


def write_tiers_txt(lesson_dir: Path, t1: list[str], t2: list[str], t3: list[str]) -> None:
    lines = []
    for i, group in enumerate((t1, t2, t3), start=1):
        if group:
            lines.append(f"{i}:{'|'.join(group)}")
        else:
            lines.append(f"{i}:^$")
    (lesson_dir / "tiers.txt").write_text("\n".join(lines) + "\n", encoding="utf-8")


def insert_tier_markers(go_file: Path, tier_funcs: tuple[list[str], list[str], list[str]]) -> None:
    if not go_file.exists():
        return
    text = go_file.read_text(encoding="utf-8")
    # Remove old markers
    for m in TIER_MARKERS:
        text = text.replace(m + "\n", "")
        text = text.replace(m + "\r\n", "")

    lines = text.splitlines(keepends=True)
    funcs = []
    for i, line in enumerate(lines):
        m = re.match(r"^func\s+(?:\([^)]*\)\s+)?([A-Z][A-Za-z0-9_]*)\s*\(", line)
        if m:
            funcs.append((i, m.group(1)))

    if not funcs:
        return

    # Map func name -> tier index 0/1/2
    name_to_tier: dict[str, int] = {}
    for ti, names in enumerate(tier_funcs):
        for n in names:
            name_to_tier[n] = ti

    # If we only have test-based tiers, map exported funcs by thirds instead
    if not any(name_to_tier.get(fn) is not None for _, fn in funcs):
        f1, f2, f3 = split_thirds([fn for _, fn in funcs])
        for ti, names in enumerate((f1, f2, f3)):
            for n in names:
                name_to_tier[n] = ti

    # Also assign methods: heuristic — match test prefix without Test
    # For funcs not in map, assign by position thirds
    missing = [fn for _, fn in funcs if fn not in name_to_tier]
    if missing:
        m1, m2, m3 = split_thirds(missing)
        for ti, names in enumerate((m1, m2, m3)):
            for n in names:
                name_to_tier[n] = ti

    inserts: dict[int, str] = {}
    seen_tier = set()
    for i, fn in funcs:
        ti = name_to_tier.get(fn, 0)
        if ti not in seen_tier:
            seen_tier.add(ti)
            inserts[i] = TIER_MARKERS[ti] + "\n"

    # Apply inserts from bottom
    for i in sorted(inserts.keys(), reverse=True):
        # Find start of doc comment block above func
        j = i
        while j > 0 and lines[j - 1].startswith("//"):
            j -= 1
        lines.insert(j, inserts[i])

    go_file.write_text("".join(lines), encoding="utf-8")


def parse_sections(md: str) -> tuple[str, list[tuple[str, str]]]:
    """Return (preamble, [(header, body), ...])."""
    parts = SECTION_RE.split(md)
    preamble = parts[0]
    sections: list[tuple[str, str]] = []
    i = 1
    while i + 1 < len(parts):
        header = parts[i]
        body = parts[i + 1]
        sections.append((header, body))
        i += 2
    return preamble, sections


def section_title(header: str) -> str:
    return header[3:].strip()


def is_title(header: str, *prefixes: str) -> bool:
    t = section_title(header)
    return any(t == p or t.startswith(p) for p in prefixes)


def strip_make_checkboxes(body: str) -> str:
    lines = []
    for line in body.splitlines(keepends=True):
        if re.search(r"make (lesson|race)|go test .*procház|je zelené|jsou čisté", line):
            # drop checkbox lines about tests passing
            if line.lstrip().startswith("- ["):
                continue
        lines.append(line)
    return "".join(lines)


def build_quiz_section(nn: str) -> tuple[str, str]:
    body = (
        f"\nPo přečtení teorie spusť v Cursoru **`/go-deep-quiz {nn}`**. AI tě ~5 minut "
        f"prověří mentální model (ne hotové cvičení). Slabiny si uloží do "
        f"[`GAPS.md`](../../GAPS.md).\n\n"
    )
    return "## AI kvíz", body


def build_task_section(
    nn: str,
    f1: list[str],
    f2: list[str],
    f3: list[str],
    context: str,
    checkpoint: bool,
) -> tuple[str, str]:
    def fmt(names: list[str]) -> str:
        if not names:
            return "_(tento stupeň v lekci není — pokračuj dál)_"
        return ", ".join(f"`{n}`" for n in names)

    intro = context.strip() or "Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí."
    parts = [
        f"\n{intro}\n",
        "\nStupně jdou od jednodušších ke složitějším — po každém stupni spusť review, než jdeš dál.\n",
    ]
    tiers = [
        ("Jednoduchý", 1, "easy", f1),
        ("Střední", 2, "medium", f2),
        ("Obtížný", 3, "hard", f3),
    ]
    for title, part, mode, names in tiers:
        if checkpoint and not names:
            continue
        parts.append(f"\n### {title}\n\n")
        parts.append(f"Funkce: {fmt(names)}\n\n")
        if names:
            parts.append("```bash\n")
            parts.append(f"make lesson L={nn} PART={part}\n")
            parts.append("```\n\n")
            parts.append(f"Pak **`/go-deep-review {nn} {mode}`**.\n")
        else:
            parts.append(f"Přeskoč na další stupeň (nebo `/go-deep-review {nn} final`).\n")
    parts.append("\nAž budou stupně hotové, porovnej se `solutions/` (spoiler).\n\n")
    return "## Úkol", "".join(parts)


def build_final_section(nn: str, old_body: str) -> tuple[str, str]:
    cleaned = strip_make_checkboxes(old_body)
    # Drop old intros that pointed at /go-deep-review after tasks
    cleaned = re.sub(
        r"(?ms)^(?:\n)?Po dokončení úkolů.*?\n(?=- \[|###|\Z)",
        "\n",
        cleaned,
        count=1,
    )
    cleaned = re.sub(
        r"(?ms)^(?:\n)?Spusť v Cursoru.*?\n(?=- \[|###|\Z)",
        "\n",
        cleaned,
        count=1,
    )
    # Avoid duplicating final intro if re-run
    cleaned = re.sub(
        r"(?ms)^(?:\n)?Spusť \*\*`/go-deep-review.*?vyžaduje\).\n",
        "\n",
        cleaned,
        count=1,
    )
    intro = (
        f"\nSpusť **`/go-deep-review {nn} final`**. AI projde body níže, doptá se a ověří "
        f"pochopení. Celé cvičení ověří `make lesson L={nn}` "
        f"(+ `make race L={nn}`, pokud to lekce vyžaduje).\n"
    )
    # If cleaned already starts with checklist, fine
    if not cleaned.strip().startswith("- ["):
        # keep remaining content after intro
        body = intro + "\n" + cleaned.lstrip("\n")
    else:
        body = intro + "\n" + cleaned.lstrip("\n")
    if "- [" not in body:
        body += "\n- [ ] Umíš vysvětlit hlavní myšlenku lekce vlastními slovy\n"
    return "## Závěrečné otázky", body if body.endswith("\n") else body + "\n"


def extract_task_context(old_task_body: str) -> str:
    """Keep a short lesson context blurb; ignore already-migrated tier sections."""
    lines = []
    for line in old_task_body.splitlines():
        s = line.strip()
        if s.startswith("```") or s.startswith("###") or s.startswith("Funkce:"):
            break
        if s.startswith("Stupně jdou") or "make lesson" in line:
            break
        if s.startswith("Až budeš") or s.startswith("Až budou"):
            break
        if s.startswith("Pak **"):
            break
        lines.append(line)
    text = "\n".join(lines).strip()
    if not text:
        return "Pracuj v `exercise/`. Kontrakt je v komentáři nad funkcí."
    return text


def migrate_readme(
    path: Path,
    nn: str,
    f1: list[str],
    f2: list[str],
    f3: list[str],
    checkpoint: bool,
) -> None:
    md = path.read_text(encoding="utf-8")
    preamble, sections = parse_sections(md)

    by_key: dict[str, tuple[str, str]] = {}
    order_keys: list[str] = []
    extras_before_ai: list[tuple[str, str]] = []
    extras_after_task: list[tuple[str, str]] = []

    for header, body in sections:
        title = section_title(header)
        if title.startswith("Co budeš umět"):
            by_key["goals"] = (header, body)
        elif title.startswith("PHP"):
            by_key["php"] = ("## Rozdíly proti PHP", body)
        elif title.startswith("Rozdíly proti PHP"):
            by_key["php"] = (header, body)
        elif title.startswith("Teorie") or title.startswith("Recap"):
            by_key["theory"] = (header, body)
        elif title.startswith("Časté chyby"):
            by_key["errors"] = (header, body)
        elif title.startswith("AI kvíz"):
            by_key["quiz"] = (header, body)
        elif title.startswith("Úkol") or title.startswith("Kumulativní"):
            by_key["task"] = (header, body)
        elif title.startswith("Ověření") or title.startswith("Závěrečné"):
            by_key["final"] = (header, body)
        elif title.startswith("AI režim"):
            by_key["ai"] = (header, body)
        elif title.startswith("Další čtení"):
            by_key["reading"] = (header, body)
        elif title.startswith("Sebehodnocení"):
            by_key["self"] = (header, body)
        elif "Review" in title or title.startswith("Projekt"):
            extras_after_task.append((header, body))
        else:
            # Unknown mid sections (e.g. Deadlock tip) — keep near theory
            extras_before_ai.append((header, body))

    old_task_body = by_key.get("task", ("", ""))[1]
    context = extract_task_context(old_task_body)

    quiz = build_quiz_section(nn)
    task = build_task_section(nn, f1, f2, f3, context, checkpoint)
    old_final = by_key.get("final", ("## Ověření", "\n- [ ] Umíš vysvětlit hlavní myšlenku lekce\n"))
    final = build_final_section(nn, old_final[1])

    # AI režim: ensure note about mentor/quiz always OK
    if "ai" in by_key:
        h, b = by_key["ai"]
        if "Mentor, kvíz" not in b and "mentor/kvíz" not in b.lower():
            if "ZAKÁZÁNO" in b and "nesmí psát" not in b:
                b = b.rstrip() + (
                    "\n\nMentor, kvíz i review (dialog) jsou vždy OK; "
                    "v tomto režimu AI nesmí psát kód cvičení.\n"
                )
            by_key["ai"] = (h, b if b.endswith("\n") else b + "\n")

    out: list[str] = [preamble]
    def add(key: str | None = None, pair: tuple[str, str] | None = None) -> None:
        if key and key in by_key:
            h, b = by_key[key]
            out.append(h + b if b.startswith("\n") else h + "\n" + b)
        elif pair:
            h, b = pair
            out.append(h + b if b.startswith("\n") else h + "\n" + b)

    add("goals")
    add("theory")
    add("php")
    add("errors")
    for h, b in extras_before_ai:
        out.append(h + b if b.startswith("\n") else h + "\n" + b)
    add(pair=quiz)
    add(pair=task)
    for h, b in extras_after_task:
        out.append(h + b if b.startswith("\n") else h + "\n" + b)
    add("self")
    add(pair=final)
    add("ai")
    add("reading")

    path.write_text("".join(out), encoding="utf-8")


def primary_exercise_go(exercise_dir: Path) -> Path | None:
    candidates = [
        exercise_dir / "exercise.go",
        *sorted(exercise_dir.rglob("*.go")),
    ]
    for c in candidates:
        if c.name.endswith("_test.go"):
            continue
        if c.exists() and collect_exported_funcs(c):
            return c
    return None


def migrate_lesson(num: int) -> None:
    nn = f"{num:02d}"
    lesson_dir = LESSONS / f"lesson-{nn}"
    if not lesson_dir.is_dir():
        print(f"skip missing {lesson_dir}", file=sys.stderr)
        return
    exercise_dir = lesson_dir / "exercise"
    tests = collect_tests(exercise_dir)

    go = primary_exercise_go(exercise_dir)
    funcs: list[str] = []
    if go:
        funcs = [fn for _, fn in collect_exported_funcs(go)]
    if funcs:
        f1, f2, f3 = split_thirds(funcs)
        t1, t2, t3 = assign_tests_to_func_tiers(tests, f1, f2, f3)
    else:
        t1, t2, t3 = split_thirds(tests)
        f1, f2, f3 = [], [], []

    write_tiers_txt(lesson_dir, t1, t2, t3)

    if go:
        insert_tier_markers(go, (f1, f2, f3))
    solutions_dir = lesson_dir / "solutions"
    if solutions_dir.exists():
        for sol_go in sorted(solutions_dir.rglob("*.go")):
            if sol_go.name.endswith("_test.go"):
                continue
            if not collect_exported_funcs(sol_go):
                continue
            if sol_go.name == "exercise.go" and funcs:
                insert_tier_markers(sol_go, (f1, f2, f3))
            else:
                sf = [fn for _, fn in collect_exported_funcs(sol_go)]
                s1, s2, s3 = split_thirds(sf)
                insert_tier_markers(sol_go, (s1, s2, s3))

    readme = lesson_dir / "README.md"
    migrate_readme(readme, nn, f1, f2, f3, num in CHECKPOINTS)
    print(f"OK lesson-{nn} tests={len(tests)} funcs={len(funcs)}")



def main() -> None:
    nums = range(1, 61)
    if len(sys.argv) > 1:
        nums = [int(x) for x in sys.argv[1:]]
    for n in nums:
        migrate_lesson(n)


if __name__ == "__main__":
    main()
