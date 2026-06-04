#!/usr/bin/env python3
"""Create a dated triage journal under docs/complaints/ (versioned in git)."""
from __future__ import annotations

import datetime as dt
import re
import sys
from pathlib import Path


def main() -> int:
    root = Path(__file__).resolve().parents[2]
    docs = root / "docs" / "complaints"
    template = docs / "journal-TEMPLATE.md"
    if not template.is_file():
        print(f"Missing template: {template}", file=sys.stderr)
        return 1

    today = dt.datetime.now(dt.UTC).strftime("%Y-%m-%d")
    args = [a.strip() for a in sys.argv[1:] if a.strip()]
    if len(args) == 0:
        date, slug = today, "triage"
    elif len(args) == 1:
        if re.fullmatch(r"\d{4}-\d{2}-\d{2}", args[0]):
            date, slug = args[0], "triage"
        else:
            date, slug = today, args[0]
    else:
        date, slug = args[0], args[1]
    if not re.fullmatch(r"\d{4}-\d{2}-\d{2}", date):
        print("Usage: new_journal.py [YYYY-MM-DD] [slug]", file=sys.stderr)
        print("  make complaints-journal-new", file=sys.stderr)
        print("  make complaints-journal-new SLUG=en", file=sys.stderr)
        print("  make complaints-journal-new JOURNAL_DATE=2026-06-15 SLUG=hotfix", file=sys.stderr)
        return 1
    slug = slug.replace(" ", "-")
    slug = re.sub(r"[^a-zA-Z0-9_-]", "", slug) or "triage"

    out = docs / f"journal-{date}-{slug}.md"
    if out.exists():
        print(f"Already exists: {out}", file=sys.stderr)
        return 1

    text = template.read_text(encoding="utf-8")
    text = text.replace("YYYY-MM-DD", date)
    out.write_text(text, encoding="utf-8")
    print(out.relative_to(root))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
