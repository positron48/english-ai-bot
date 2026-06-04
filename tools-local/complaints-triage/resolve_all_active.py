#!/usr/bin/env python3
"""Resolve all active content reports for a course on prod."""
from __future__ import annotations

import json
import os
import sys
import urllib.error
import urllib.request
from pathlib import Path


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
    reports: list[dict] = []
    cursor = ""
    while True:
        url = f"{base}/api/internal/content-reports?limit=200&status=active&course={course}"
        if cursor:
            url += f"&cursor={cursor}"
        data = http("GET", url, token)
        chunk = data.get("reports") or []
        reports.extend(chunk)
        cursor = data.get("next_cursor")
        if not cursor or not chunk:
            break
    return reports


def main() -> int:
    course = (sys.argv[1] if len(sys.argv) > 1 else "en").strip().lower()
    if course not in ("en", "es"):
        print("Usage: resolve_all_active.py [en|es]", file=sys.stderr)
        return 1

    if course == "en":
        base = os.environ.get("COMPLAINTS_SERVICE_URL_EN", "").rstrip("/")
    else:
        base = os.environ.get("COMPLAINTS_SERVICE_URL_ES", "").rstrip("/")
    token = os.environ.get("COMPLAINTS_SERVICE_TOKEN_ES") or os.environ.get(
        "COMPLAINTS_SERVICE_TOKEN_EN", ""
    )
    if not base or not token:
        print("Load secrets/complaints-prod.env first", file=sys.stderr)
        return 1

    reports = fetch_all(base, token, course)
    ids = [int(r["id"]) for r in reports if r.get("id")]
    if not ids:
        print(f"No active reports for {course}")
        return 0

    reason = (
        "triage-2026-06-04: content fixes in courses/training_pack + word/TTS API; "
        "grammar bundle regen pending deploy/import"
    )
    out = http(
        "POST",
        f"{base}/api/internal/content-reports/resolve-bulk",
        token,
        {"report_ids": ids, "reason": reason},
    )
    print(f"✓ resolved {len(ids)} {course} reports:", out)

    log_dir = Path("logs/complaints")
    log_dir.mkdir(parents=True, exist_ok=True)
    log_path = log_dir / "triage-2026-06.jsonl"
    entry = {
        "run_id": "triage-2026-06-04",
        "course": course,
        "action": "resolve_bulk",
        "report_ids": ids,
        "resolve_status": out,
    }
    with log_path.open("a", encoding="utf-8") as f:
        f.write(json.dumps(entry, ensure_ascii=False) + "\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
