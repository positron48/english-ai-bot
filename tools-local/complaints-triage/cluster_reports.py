#!/usr/bin/env python3
"""Build triage clusters from fetch_reports.py snapshot JSON (dry-run aid)."""
from __future__ import annotations

import argparse
import json
import sys
from collections import defaultdict
from pathlib import Path
from typing import Any, DefaultDict, Dict, List, Tuple


def entity_key(report: dict) -> str:
    st = report.get("source_type", "")
    if st == "grammar_training":
        return "|".join(
            [
                st,
                report.get("grammar_chapter_id") or "",
                report.get("theory_block_id") or "",
                report.get("grammar_question_id") or "",
            ]
        )
    return "|".join(
        [
            st or "word_training",
            str(report.get("training_card_id") or ""),
            str(report.get("word_card_id") or ""),
            report.get("word") or "",
            report.get("translation_direction") or "",
        ]
    )


def cluster_key(report: dict) -> str:
    cat = report.get("report_category") or "other"
    return f"{cat}::{entity_key(report)}"


def main() -> int:
    parser = argparse.ArgumentParser(description="Cluster content report snapshot")
    parser.add_argument("snapshot", type=Path, help="snapshot-*.json from fetch_reports.py")
    parser.add_argument("--min-count", type=int, default=1)
    args = parser.parse_args()

    data = json.loads(args.snapshot.read_text(encoding="utf-8"))
    reports: List[dict] = data.get("reports", [])
    groups: DefaultDict[str, List[dict]] = defaultdict(list)
    for r in reports:
        groups[cluster_key(r)].append(r)

    ranked: List[Tuple[str, List[dict]]] = sorted(
        groups.items(), key=lambda x: len(x[1]), reverse=True
    )

    out: Dict[str, Any] = {
        "source_snapshot": str(args.snapshot),
        "report_count": len(reports),
        "cluster_count": len(ranked),
        "clusters": [],
    }
    for key, items in ranked:
        if len(items) < args.min_count:
            continue
        sample = items[0]
        out["clusters"].append(
            {
                "cluster_key": key,
                "count": len(items),
                "report_ids": [i.get("id") for i in items],
                "source_type": sample.get("source_type"),
                "report_category": sample.get("report_category"),
                "sample_comment": sample.get("comment_text"),
                "grammar_chapter_id": sample.get("grammar_chapter_id"),
                "theory_block_id": sample.get("theory_block_id"),
                "grammar_question_id": sample.get("grammar_question_id"),
                "word": sample.get("word"),
                "training_card_id": sample.get("training_card_id"),
            }
        )

    json.dump(out, sys.stdout, ensure_ascii=False, indent=2)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
