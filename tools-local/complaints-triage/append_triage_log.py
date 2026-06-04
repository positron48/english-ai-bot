#!/usr/bin/env python3
"""Append one line to logs/complaints/triage-YYYY-MM.jsonl (apply mode journal)."""
from __future__ import annotations

import argparse
import datetime as dt
import json
import sys
from pathlib import Path


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--course", required=True, choices=["en", "es"])
    parser.add_argument("--run-id", required=True)
    parser.add_argument("--cluster-key", required=True)
    parser.add_argument("--category", default="")
    parser.add_argument("--action", required=True)
    parser.add_argument("--report-ids", default="", help="comma-separated ids")
    parser.add_argument("--files-changed", default="", help="comma-separated paths")
    parser.add_argument("--resolve-status", default="pending")
    parser.add_argument("--note", default="")
    args = parser.parse_args()

    root = Path(__file__).resolve().parents[2]
    log_dir = root / "logs" / "complaints"
    log_dir.mkdir(parents=True, exist_ok=True)
    month = dt.datetime.now(dt.UTC).strftime("%Y-%m")
    log_path = log_dir / f"triage-{month}.jsonl"

    report_ids = []
    if args.report_ids.strip():
        report_ids = [int(x.strip()) for x in args.report_ids.split(",") if x.strip()]

    row = {
        "timestamp": dt.datetime.now(dt.UTC).replace(microsecond=0).isoformat().replace("+00:00", "Z"),
        "run_id": args.run_id,
        "course": args.course,
        "cluster_key": args.cluster_key,
        "category": args.category,
        "action": args.action,
        "report_ids": report_ids,
        "files_changed": [p.strip() for p in args.files_changed.split(",") if p.strip()],
        "resolve_status": args.resolve_status,
        "note": args.note,
    }
    with log_path.open("a", encoding="utf-8") as f:
        f.write(json.dumps(row, ensure_ascii=False) + "\n")
    print(log_path)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
