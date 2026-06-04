#!/usr/bin/env python3
"""Fetch active content reports from prod/internal API (no LLM)."""
from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Any, Dict, List, Optional


def http_json(method: str, url: str, token: str) -> Any:
    headers = {"X-Service-Token": token, "Content-Type": "application/json"}
    req = urllib.request.Request(url=url, method=method, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            return json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        body = e.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"HTTP {e.code} {url}: {body}") from e
    except urllib.error.URLError as e:
        raise RuntimeError(f"Network error {url}: {e}") from e


def fetch_unified_reports(base_url: str, token: str, course: str, source_type: str) -> List[dict]:
    out: List[dict] = []
    cursor: Optional[int] = None
    while True:
        q: Dict[str, str] = {"limit": "200", "status": "active"}
        if course:
            q["course"] = course
        if source_type:
            q["source_type"] = source_type
        if cursor is not None:
            q["cursor"] = str(cursor)
        url = f"{base_url.rstrip('/')}/api/internal/content-reports?{urllib.parse.urlencode(q)}"
        data = http_json("GET", url, token)
        chunk = data.get("reports", [])
        if not chunk:
            break
        out.extend(chunk)
        cursor = data.get("next_cursor")
        if not cursor:
            break
    return out


def fetch_grammar_legacy(base_url: str, token: str, course: str) -> List[dict]:
    """Pre-unified prod: only grammar_training via /api/internal/content-reports/grammar."""
    out: List[dict] = []
    cursor: Optional[int] = None
    while True:
        q: Dict[str, str] = {"limit": "200", "course": course}
        if cursor is not None:
            q["cursor"] = str(cursor)
        url = f"{base_url.rstrip('/')}/api/internal/content-reports/grammar?{urllib.parse.urlencode(q)}"
        data = http_json("GET", url, token)
        chunk = data.get("reports", [])
        if not chunk:
            break
        for item in chunk:
            item.setdefault("source_type", "grammar_training")
            item.setdefault("status", "active")
            item.setdefault("report_category", "other")
        out.extend(chunk)
        cursor = data.get("next_cursor")
        if not cursor:
            break
    return out


def fetch_all_reports(base_url: str, token: str, course: str, source_type: str) -> tuple[List[dict], str]:
    st = source_type.strip()
    if st == "word_training":
        try:
            return fetch_unified_reports(base_url, token, course, st), "unified"
        except RuntimeError as e:
            if "HTTP 404" not in str(e):
                raise
            print("word_training requires unified /api/internal/content-reports (deploy new image)", file=sys.stderr)
            return [], "unified_missing"
    try:
        return fetch_unified_reports(base_url, token, course, st), "unified"
    except RuntimeError as e:
        if "HTTP 404" not in str(e):
            raise
    if st and st != "grammar_training":
        return [], "legacy_grammar_only"
    grammar = fetch_grammar_legacy(base_url, token, course)
    return grammar, "legacy_grammar"


def main() -> int:
    parser = argparse.ArgumentParser(description="Fetch content reports snapshot")
    parser.add_argument("--course", choices=["en", "es", "english", "spanish"], default="en")
    parser.add_argument("--source-type", default="", help="word_training|grammar_training|empty=all")
    parser.add_argument("--service-url", default=os.getenv("COMPLAINTS_SERVICE_URL", ""))
    parser.add_argument("--service-token", default=os.getenv("COMPLAINTS_SERVICE_TOKEN", ""))
    parser.add_argument("--out", default="")
    args = parser.parse_args()

    url = args.service_url.strip()
    token = args.service_token.strip()
    if not url or not token:
        print("Set COMPLAINTS_SERVICE_URL and COMPLAINTS_SERVICE_TOKEN", file=sys.stderr)
        return 1

    course = args.course
    if course == "english":
        course = "en"
    if course == "spanish":
        course = "es"

    reports, api_mode = fetch_all_reports(url, token, course, args.source_type.strip())
    summary: List[dict] = []
    if api_mode == "unified":
        try:
            summary_url = f"{url.rstrip('/')}/api/internal/content-reports/summary?course={course}"
            data = http_json("GET", summary_url, token)
            summary = data.get("summary", [])
        except RuntimeError:
            pass

    payload = {
        "fetched_at": dt.datetime.now(dt.UTC).replace(microsecond=0).isoformat().replace("+00:00", "Z"),
        "course": course,
        "service_url": url,
        "api_mode": api_mode,
        "report_count": len(reports),
        "summary": summary,
        "reports": reports,
    }

    root = Path(__file__).resolve().parents[2]
    out_dir = root / "logs" / "complaints"
    out_dir.mkdir(parents=True, exist_ok=True)
    ts = dt.datetime.now(dt.UTC).strftime("%Y%m%dT%H%M%SZ")
    out_path = Path(args.out) if args.out else out_dir / f"snapshot-{course}-{ts}.json"
    out_path.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"Wrote {len(reports)} reports -> {out_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
