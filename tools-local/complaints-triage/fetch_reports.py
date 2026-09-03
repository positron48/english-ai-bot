#!/usr/bin/env python3
"""Fetch active content reports from prod/internal API (no LLM)."""
from __future__ import annotations

import argparse
from collections import Counter
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
            return fetch_unified_reports(base_url, token, "", st), "unified"
        except RuntimeError as e:
            if "HTTP 404" not in str(e):
                raise
            print("word_training requires unified /api/internal/content-reports (deploy new image)", file=sys.stderr)
            return [], "unified_missing"
    try:
        # Older servers silently omit reading_text when course is supplied.
        # Always fetch every page without it and classify locally.
        return fetch_unified_reports(base_url, token, "", st), "unified"
    except RuntimeError as e:
        if "HTTP 404" not in str(e):
            raise
    if st and st != "grammar_training":
        return [], "legacy_grammar_only"
    grammar = fetch_grammar_legacy(base_url, token, course)
    return grammar, "legacy_grammar"


def report_course(report: dict) -> str:
    payload = report.get("payload") or {}
    snapshot = payload.get("content_snapshot") or {}
    candidates = [report.get("course_code"), payload.get("course_code"),
                  snapshot.get("target_language"), payload.get("target_language"),
                  report.get("grammar_chapter_id"), payload.get("text_id"),
                  payload.get("category_id"), snapshot.get("category_id")]
    for value in candidates:
        value = str(value or "").strip().lower()
        for course in ("en", "es"):
            if value == course or value.startswith((course + ".", course + "_", "free_" + course + "_")):
                return course
    # A report can outlive the catalog entry; unresolved IDs remain visible.
    text_id = payload.get("text_id") or report.get("grammar_chapter_id")
    if report.get("source_type") == "reading_text" and text_id:
        root = Path(__file__).resolve().parents[2]
        for course, directory in (("en", "english-grammar"), ("es", "spanish-grammar")):
            path = root / "courses" / directory / "reading" / "index.json"
            if path.exists() and text_id in json.loads(path.read_text(encoding="utf-8")).get("texts", {}):
                return course
    return ""


def select_course(reports: List[dict], course: str) -> List[dict]:
    selected = []
    for report in reports:
        detected = report_course(report)
        if course == "all" or not detected or detected == course:
            selected.append({**report, "triage_course": detected or "unknown"})
    return selected


def main() -> int:
    parser = argparse.ArgumentParser(description="Fetch content reports snapshot")
    parser.add_argument("--course", choices=["en", "es", "english", "spanish", "all"], default="en")
    parser.add_argument("--source-type", default="", help="word_training|grammar_training|grammar_chapter|grammar_test|reading_text|empty=all")
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

    all_reports, api_mode = fetch_all_reports(url, token, course, args.source_type.strip())
    reports = select_course(all_reports, course)
    counts = Counter((r.get("source_type"), r.get("report_category") or "other") for r in reports)
    summary = [{"source_type": st, "report_category": cat, "count": count}
               for (st, cat), count in sorted(counts.items())]
    unknown_ids = [r["id"] for r in reports if r["triage_course"] == "unknown"]
    if unknown_ids:
        print(f"Course unknown for report IDs {unknown_ids}; retained for manual triage (may appear in both courses)", file=sys.stderr)
    if api_mode != "unified":
        print(f"INCOMPLETE snapshot: {api_mode}; reading and other report types are not verified", file=sys.stderr)

    payload = {
        "fetched_at": dt.datetime.now(dt.UTC).replace(microsecond=0).isoformat().replace("+00:00", "Z"),
        "course": course,
        "service_url": url,
        "api_mode": api_mode,
        "complete": api_mode == "unified",
        "unfiltered_report_count": len(all_reports),
        "unknown_course_report_ids": unknown_ids,
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
    return 0 if api_mode == "unified" else 2


if __name__ == "__main__":
    raise SystemExit(main())
