#!/usr/bin/env python3
"""
Build Spanish noun gender lexicon from doozan/spanish_data (es-en.data).

Source:
  https://raw.githubusercontent.com/doozan/spanish_data/master/es-en.data

Output TSV columns:
  lemma\tgender\tarticle\topposite_gender_word\tsource\tnotes
"""

from __future__ import annotations

import argparse
import re
import urllib.request
from pathlib import Path

SRC_URL = "https://raw.githubusercontent.com/doozan/spanish_data/master/es-en.data"
DEFAULT_INPUT = Path("tmp/es-en.data")
DEFAULT_OUTPUT = Path("resources/wordsets/spanish_gender_lexicon.tsv")

META_M_RE = re.compile(r"\bm=([^|}]+)")
META_F_RE = re.compile(r"\bf=([^|}]+)")
WORD_RE = re.compile(r"^[a-záéíóúüñ]+(?:[-'][a-záéíóúüñ]+)*$", re.IGNORECASE)


def map_gender(raw: str) -> str:
    g = (raw or "").strip().lower()
    if not g:
        return ""
    if "mfbysense" in g or g == "mf":
        return "mf"
    if g.startswith("m"):
        return "m"
    if g.startswith("f"):
        return "f"
    if g.startswith("n"):
        return "n"
    return ""


def article_for(gender: str) -> str:
    return {"m": "el", "f": "la", "mf": "el/la", "n": "lo"}.get(gender, "")


def safe_token(value: str) -> str:
    v = (value or "").strip().lower()
    if not v or v == "+":
        return ""
    if " " in v:
        return ""
    if not WORD_RE.match(v):
        return ""
    return v


def is_clean_lemma(lemma: str) -> bool:
    return bool(WORD_RE.match(lemma))


def inferred_plus_opposite(word: str, gender: str, marker: str) -> str:
    # marker is "+" from m=+ / f=+ templates.
    if marker != "+":
        return ""
    w = word.lower().strip()
    if not w:
        return ""
    # Conservative: infer only simple -o/-a alternation.
    if gender == "m" and w.endswith("o") and len(w) > 1:
        return w[:-1] + "a"
    if gender == "f" and w.endswith("a") and len(w) > 1:
        return w[:-1] + "o"
    return ""


def parse_block(lines: list[str]) -> tuple[str, str, str]:
    """Return (gender, opposite, notes) for a word block."""
    best_gender = ""
    best_opp = ""
    best_notes = ""

    i = 1  # line 0 is word
    while i < len(lines):
        line = lines[i]
        if not line.startswith("pos: "):
            i += 1
            continue
        pos = line[5:].strip()
        i += 1
        if pos != "n":
            continue

        meta = ""
        g = ""
        while i < len(lines) and not lines[i].startswith("pos: "):
            cur = lines[i]
            if cur.startswith("  meta: "):
                meta = cur[len("  meta: ") :].strip()
            elif cur.startswith("  g: "):
                g = cur[len("  g: ") :].strip()
            elif cur == "_____":
                break
            i += 1

        gender = map_gender(g)
        if not gender:
            continue

        opposite = ""
        note = "wiktionary-gender"
        m_match = META_M_RE.search(meta)
        f_match = META_F_RE.search(meta)
        if gender == "m" and f_match:
            raw = f_match.group(1).strip().lower()
            opposite = safe_token(raw) or inferred_plus_opposite(lines[0], "m", raw)
            note = "meta-f=explicit" if safe_token(raw) else ("meta-f=+" if raw == "+" else note)
        elif gender == "f" and m_match:
            raw = m_match.group(1).strip().lower()
            opposite = safe_token(raw) or inferred_plus_opposite(lines[0], "f", raw)
            note = "meta-m=explicit" if safe_token(raw) else ("meta-m=+" if raw == "+" else note)

        # Prefer entries with explicit opposite; otherwise first valid gender.
        if opposite and not best_opp:
            best_gender, best_opp, best_notes = gender, opposite, note
        elif not best_gender:
            best_gender, best_opp, best_notes = gender, "", note

    return best_gender, best_opp, best_notes


def iter_blocks(path: Path):
    block: list[str] = []
    with path.open("r", encoding="utf-8", errors="ignore") as f:
        for raw in f:
            line = raw.rstrip("\n")
            if line == "_____":
                if block:
                    yield block
                    block = []
                continue
            block.append(line)
        if block:
            yield block


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", default=str(DEFAULT_INPUT), help="Path to es-en.data")
    parser.add_argument("--output", default=str(DEFAULT_OUTPUT), help="Output TSV path")
    parser.add_argument("--download", action="store_true", help="Download source to --input before parsing")
    args = parser.parse_args()

    input_path = Path(args.input)
    output_path = Path(args.output)

    if args.download:
        input_path.parent.mkdir(parents=True, exist_ok=True)
        print(f"Downloading {SRC_URL} -> {input_path}")
        urllib.request.urlretrieve(SRC_URL, input_path)

    if not input_path.exists():
        raise FileNotFoundError(f"Input file not found: {input_path}")

    rows = []
    seen = set()
    for block in iter_blocks(input_path):
        if not block:
            continue
        lemma = block[0].strip().lower()
        if not lemma or " " in lemma or not is_clean_lemma(lemma):
            continue
        gender, opposite, notes = parse_block(block)
        if not gender:
            continue
        key = (lemma, gender, opposite)
        if key in seen:
            continue
        seen.add(key)
        rows.append((lemma, gender, article_for(gender), opposite, "doozan/es-en.data", notes))

    rows.sort(key=lambda x: x[0])

    output_path.parent.mkdir(parents=True, exist_ok=True)
    with output_path.open("w", encoding="utf-8") as f:
        f.write("lemma\tgender\tarticle\topposite_gender_word\tsource\tnotes\n")
        for row in rows:
            f.write("\t".join(row) + "\n")

    print(f"Wrote {len(rows)} rows to {output_path}")


if __name__ == "__main__":
    main()

