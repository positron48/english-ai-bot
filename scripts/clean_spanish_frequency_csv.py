#!/usr/bin/env python3
import argparse
import csv
import re
from pathlib import Path


SPANISH_LEMMA_RE = re.compile(r"^[a-záéíóúüñ]+(?:-[a-záéíóúüñ]+)*$")
BLOCKED_LEMMAS = {"&", "a", "an", "the", "and", "km", "george", "ugt", "wahid"}


def is_valid_lemma(lemma: str) -> bool:
    lemma = lemma.strip().lower()
    if not lemma:
        return False
    if lemma in BLOCKED_LEMMAS:
        return False
    return bool(SPANISH_LEMMA_RE.match(lemma))


def clean_csv(path: Path) -> tuple[int, int]:
    with path.open("r", encoding="utf-8", newline="") as f:
        reader = csv.DictReader(f)
        fieldnames = reader.fieldnames
        if not fieldnames:
            raise ValueError(f"{path}: missing header")
        rows = list(reader)

    kept = []
    removed = 0
    for row in rows:
        lemma = (row.get("lemma") or "").strip().lower()
        pos = (row.get("pos") or "").strip().upper()
        if pos == "PROPN":
            removed += 1
            continue
        if not is_valid_lemma(lemma):
            removed += 1
            continue
        row["lemma"] = lemma
        kept.append(row)

    with path.open("w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(kept)

    return len(kept), removed


def main() -> int:
    parser = argparse.ArgumentParser(description="Clean Spanish frequency CSV from garbage/proper names")
    parser.add_argument("paths", nargs="+", help="CSV file paths to clean in-place")
    args = parser.parse_args()

    for raw_path in args.paths:
        path = Path(raw_path)
        kept, removed = clean_csv(path)
        print(f"{path}: kept={kept} removed={removed}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
