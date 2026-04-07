#!/usr/bin/env python3
"""Strip garbage from Spanish UD frequency CSVs (proper names, abbreviations, wrong POS)."""

from __future__ import annotations

import argparse
import csv
import re
import unicodedata
from pathlib import Path

SPANISH_LEMMA_RE = re.compile(r"^[a-záéíóúüñ]+(?:-[a-záéíóúüñ]+)*$")
# UD tags we actually import into training POS lists (matches import_word_sets_from_csv).
TRAINING_POS = frozenset({"NOUN", "VERB", "ADJ", "ADV", "AUX"})
SPANISH_VOWELS_RE = re.compile(r"[aeiouyáéíóúü]", re.IGNORECASE)
REPO_ROOT = Path(__file__).resolve().parent.parent
DEFAULT_BLOCKLIST = REPO_ROOT / "resources/wordsets/spanish_lemma_blocklist.txt"


def load_blocklist(path: Path) -> frozenset[str]:
    out: set[str] = set()
    if not path.is_file():
        raise FileNotFoundError(f"blocklist not found: {path}")
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        out.add(normalize_lemma(line))
    return frozenset(out)


def normalize_lemma(lemma: str) -> str:
    return unicodedata.normalize("NFC", lemma.strip().lower())


def is_vowelless_abbrev_ascii(lemma: str) -> bool:
    """2–5 plain Latin letters, no vowel (y counts as vowel). Units like cm/mm and acronyms like psc."""
    if len(lemma) < 2 or len(lemma) > 5:
        return False
    if not lemma.isascii() or not lemma.isalpha():
        return False
    return SPANISH_VOWELS_RE.search(lemma) is None


def is_valid_lemma(lemma: str, blocklist: frozenset[str]) -> bool:
    lemma_n = normalize_lemma(lemma)
    if not lemma_n:
        return False
    if len(lemma_n) == 1:
        return False
    if lemma_n in blocklist:
        return False
    if is_vowelless_abbrev_ascii(lemma_n):
        return False
    return bool(SPANISH_LEMMA_RE.match(lemma_n))


def clean_csv(path: Path, blocklist: frozenset[str]) -> tuple[list[dict[str, str]], int]:
    with path.open("r", encoding="utf-8", newline="") as f:
        reader = csv.DictReader(f)
        fieldnames = reader.fieldnames
        if not fieldnames:
            raise ValueError(f"{path}: missing header")
        rows = list(reader)

    kept: list[dict[str, str]] = []
    removed = 0
    for row in rows:
        lemma = row.get("lemma") or ""
        pos = (row.get("pos") or "").strip().upper()
        if pos not in TRAINING_POS:
            removed += 1
            continue
        if not is_valid_lemma(lemma, blocklist):
            removed += 1
            continue
        row["lemma"] = normalize_lemma(lemma)
        kept.append(row)

    with path.open("w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(kept)

    return kept, removed


def audit_csv(path: Path, blocklist: frozenset[str], limit: int) -> None:
    """Print suspicious rows that would remain after clean (manual review)."""
    with path.open("r", encoding="utf-8", newline="") as f:
        rows = list(csv.DictReader(f))
    ascii_word = re.compile(r"^[a-z]{2,8}$")
    hits: list[tuple[str, str, str]] = []
    for row in rows:
        pos = (row.get("pos") or "").strip().upper()
        if pos not in TRAINING_POS or pos == "PROPN":
            continue
        le = normalize_lemma(row.get("lemma") or "")
        if not le or not ascii_word.match(le):
            continue
        if le in blocklist or is_vowelless_abbrev_ascii(le):
            continue
        if not SPANISH_VOWELS_RE.search(le):
            continue
        hits.append((row.get("rank", ""), le, pos))
    print(f"{path}: ascii a–z lemmas (len 2–8) in training POS: {len(hits)} (showing up to {limit})")
    for t in hits[:limit]:
        print("  ", " ".join(t))


def main() -> int:
    parser = argparse.ArgumentParser(description="Clean Spanish frequency CSV from garbage/proper names")
    parser.add_argument("paths", nargs="+", help="CSV file paths to clean in-place")
    parser.add_argument(
        "--blocklist",
        type=Path,
        default=DEFAULT_BLOCKLIST,
        help="path to newline-separated lemma blocklist",
    )
    parser.add_argument(
        "--audit",
        action="store_true",
        help="only print ascii-token scan for manual review (does not write)",
    )
    parser.add_argument(
        "--audit-limit",
        type=int,
        default=300,
        help="max rows to print with --audit",
    )
    args = parser.parse_args()

    blocklist = load_blocklist(args.blocklist)

    if args.audit:
        for raw_path in args.paths:
            audit_csv(Path(raw_path), blocklist, args.audit_limit)
        return 0

    for raw_path in args.paths:
        path = Path(raw_path)
        kept, removed = clean_csv(path, blocklist)
        print(f"{path}: kept={len(kept)} removed={removed}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
