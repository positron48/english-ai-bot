#!/usr/bin/env python3
"""Resolve explicitly reviewed report IDs; preview by default (legacy filename)."""
from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import sys
import urllib.error
import urllib.request
from pathlib import Path

from fetch_reports import fetch_all_reports, select_course


def http(method: str, url: str, token: str, body: dict | None = None) -> dict:
    data = None
    headers = {"X-Service-Token": token, "Content-Type": "application/json"}
    if body is not None:
        data = json.dumps(body).encode("utf-8")
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    with urllib.request.urlopen(req, timeout=90) as resp:
        raw = resp.read().decode("utf-8")
        return json.loads(raw) if raw else {}


def fetch_all(base: str, token: str, course: str) -> list[dict]:
    reports, mode = fetch_all_reports(base, token, course, "")
    if mode != "unified":
        raise RuntimeError("Cannot resolve reports using an incomplete legacy snapshot")
    return select_course(reports, course)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("course", choices=("en", "es"))
    parser.add_argument("--report-ids", required=True, help="Comma-separated IDs already verified on production")
    parser.add_argument("--reason", required=True, help="Actual correction and production verification")
    parser.add_argument("--apply", action="store_true", help="Send resolve-bulk; otherwise only preview")
    args = parser.parse_args()
    course = args.course
    try:
        ids = sorted({int(value.strip()) for value in args.report_ids.split(",")})
        if not ids or min(ids) <= 0 or not args.reason.strip():
            raise ValueError()
    except ValueError:
        parser.error("Provide positive report IDs and a nonempty reason")

    if course == "en":
        base = os.environ.get("COMPLAINTS_SERVICE_URL_EN", "").rstrip("/")
    else:
        base = os.environ.get("COMPLAINTS_SERVICE_URL_ES", "").rstrip("/")
    token = os.environ.get("COMPLAINTS_SERVICE_TOKEN_" + course.upper()) or os.environ.get("COMPLAINTS_SERVICE_TOKEN", "")
    if not base or not token:
        print("Load secrets/complaints-prod.env first", file=sys.stderr)
        return 1

    reports = fetch_all(base, token, course)
    active_ids = {int(r["id"]) for r in reports}
    missing = set(ids) - active_ids
    if missing:
        parser.error(f"IDs are not active in this course: {sorted(missing)}")
    reason = args.reason.strip()
    if not args.apply:
        print(json.dumps({"course": course, "report_ids": ids, "reason": reason, "apply": False}, ensure_ascii=False))
        return 0
    out = http(
        "POST",
        f"{base}/api/internal/content-reports/resolve-bulk",
        token,
        {"report_ids": ids, "reason": reason},
    )

    log_dir = Path("logs/complaints")
    log_dir.mkdir(parents=True, exist_ok=True)
    now = dt.datetime.now(dt.UTC)
    log_path = log_dir / f"triage-{now:%Y-%m}.jsonl"
    entry = {
        "run_id": f"triage-{now:%Y%m%dT%H%M%SZ}",
        "timestamp": now.isoformat(),
        "reason": reason,
        "course": course,
        "action": "resolve_bulk",
        "report_ids": ids,
        "resolve_status": out,
    }
    with log_path.open("a", encoding="utf-8") as f:
        f.write(json.dumps(entry, ensure_ascii=False) + "\n")
    if out.get("success") is not True or out.get("affected") != len(ids):
        print(f"Unexpected resolve result; inspect {log_path} and re-fetch before retrying", file=sys.stderr)
        return 1
    print(f"Resolved {len(ids)} {course} reports:", out)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
