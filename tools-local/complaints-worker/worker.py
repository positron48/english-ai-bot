#!/usr/bin/env python3
import argparse
import base64
import datetime as dt
import hashlib
import json
import os
import re
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Dict, List, Tuple


def utc_now() -> str:
    return dt.datetime.now(dt.UTC).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def read_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def write_json(path: Path, data: dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def append_jsonl(path: Path, row: dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a", encoding="utf-8") as f:
        f.write(json.dumps(row, ensure_ascii=False) + "\n")


def sha256_text(s: str) -> str:
    return hashlib.sha256(s.encode("utf-8")).hexdigest()


def http_json(method: str, url: str, token: str, payload=None):
    body = None
    headers = {"X-Service-Token": token, "Content-Type": "application/json"}
    if payload is not None:
        body = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url=url, method=method, headers=headers, data=body)
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        msg = e.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"http error {e.code} {url}: {msg}") from e
    except urllib.error.URLError as e:
        raise RuntimeError(
            f"network error for {url}: {e}. "
            "Проверь COMPLAINTS_SERVICE_URL(_EN/_ES), доступность сервиса и порт."
        ) from e


def fetch_reports(service_url: str, token: str, course: str, limit: int = 200) -> List[dict]:
    out: List[dict] = []
    cursor = None
    while True:
        q = {"course": course, "limit": str(limit)}
        if cursor:
            q["cursor"] = str(cursor)
        url = f"{service_url.rstrip('/')}/api/internal/content-reports/grammar?{urllib.parse.urlencode(q)}"
        data = http_json("GET", url, token)
        chunk = data.get("reports", [])
        if not chunk:
            break
        out.extend(chunk)
        cursor = data.get("next_cursor")
        if not cursor:
            break
    return out


def extract_local_question_id(question_id: str, chapter_id: str, theory_block_id: str) -> str:
    # stable format: chapter::block::qid
    prefix = f"{chapter_id}::{theory_block_id}::"
    if question_id.startswith(prefix):
        return question_id[len(prefix):]
    return question_id


def load_training_pack_index(course_dir: Path) -> dict:
    idx_path = course_dir / "training_pack" / "index.json"
    return read_json(idx_path)


def block_to_relpath(index_data: dict, chapter_id: str, theory_block_id: str) -> str:
    blocks = index_data.get("blocks", {}) or {}
    key = f"{chapter_id}::{theory_block_id}"
    rel = blocks.get(key, "")
    return str(rel).strip()


def remove_questions_in_file(path: Path, question_ids: set) -> Tuple[int, str, str]:
    raw_before = path.read_text(encoding="utf-8")
    before_hash = sha256_text(raw_before)
    data = json.loads(raw_before)
    questions = data.get("questions", [])
    if not isinstance(questions, list):
        return 0, before_hash, before_hash
    kept = [q for q in questions if str(q.get("id", "")) not in question_ids]
    removed = len(questions) - len(kept)
    data["questions"] = kept
    raw_after = json.dumps(data, ensure_ascii=False, indent=2) + "\n"
    after_hash = sha256_text(raw_after)
    if removed > 0:
        path.write_text(raw_after, encoding="utf-8")
    return removed, before_hash, after_hash


def parse_llm_json(text: str) -> dict:
    text = text.strip()
    if text.startswith("```"):
        text = re.sub(r"^```[a-zA-Z0-9]*\n?", "", text)
        text = re.sub(r"\n?```$", "", text)
    try:
        return json.loads(text)
    except Exception:
        match = re.search(r"\{.*\}", text, flags=re.S)
        if not match:
            return {"summary": "failed to parse llm json", "raw": text}
        try:
            return json.loads(match.group(0))
        except Exception:
            return {"summary": "failed to parse llm json", "raw": text}


def analyze_block_with_llama(llama_url: str, llama_model: str, block_key: str, reports: List[dict]) -> dict:
    prompt = {
        "task": "Analyze complained grammar questions and explain likely issue patterns in one theory block.",
        "block_key": block_key,
        "reports": reports,
        "required_json_fields": ["summary", "root_causes", "risk_level", "recommended_fix"],
    }
    payload = {
        "model": llama_model,
        "messages": [
            {"role": "system", "content": "Return only JSON."},
            {"role": "user", "content": json.dumps(prompt, ensure_ascii=False)},
        ],
        "temperature": 0.1,
    }
    req = urllib.request.Request(
        url=f"{llama_url.rstrip('/')}/v1/chat/completions",
        method="POST",
        headers={"Content-Type": "application/json"},
        data=json.dumps(payload).encode("utf-8"),
    )
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            data = json.loads(resp.read().decode("utf-8"))
    except Exception as e:
        return {"summary": "llm call failed", "error": str(e)}
    content = ""
    try:
        content = data["choices"][0]["message"]["content"]
    except Exception:
        return {"summary": "llm response malformed", "raw": data}
    return parse_llm_json(content)


def detect_course(chapter_id: str) -> str:
    ch = (chapter_id or "").strip().lower()
    if ch.startswith("es.") or ch.startswith("spanish."):
        return "spanish"
    return "english"


def main() -> int:
    parser = argparse.ArgumentParser(description="Local grammar complaints worker")
    parser.add_argument("--service-url", default=os.getenv("COMPLAINTS_SERVICE_URL", "http://localhost:8184"))
    parser.add_argument("--service-token", default=os.getenv("COMPLAINTS_SERVICE_TOKEN", ""))
    parser.add_argument("--llama-url", default=os.getenv("LLAMACPP_URL", "http://127.0.0.1:8080"))
    parser.add_argument("--llama-model", default=os.getenv("LLAMACPP_MODEL", "local-model"))
    parser.add_argument("--course-scope", default=os.getenv("COURSE_SCOPE", "both"), choices=["english", "spanish", "both"])
    parser.add_argument("--workspace", default=os.getenv("WORKSPACE_ROOT", str(Path(__file__).resolve().parents[2])))
    parser.add_argument("--apply", action="store_true", help="Apply file changes and resolve reports. Default is dry-run.")
    args = parser.parse_args()

    if not args.service_token:
        print("COMPLAINTS_SERVICE_TOKEN is required", file=sys.stderr)
        return 2

    dry_run = not args.apply
    run_id = base64.urlsafe_b64encode(os.urandom(9)).decode("ascii").rstrip("=")
    now = dt.datetime.now(dt.UTC)
    logs_dir = Path(args.workspace) / "logs" / "complaints"
    journal_path = logs_dir / f"complaints-{now.strftime('%Y-%m')}.jsonl"
    changed_blocks_path = logs_dir / f"changed-theory-blocks-{now.strftime('%Y%m%d%H')}.json"

    course_dirs = {
        "english": Path(args.workspace) / "courses" / "english-grammar",
        "spanish": Path(args.workspace) / "courses" / "spanish-grammar",
    }
    selected = ["english", "spanish"] if args.course_scope == "both" else [args.course_scope]

    all_reports: List[dict] = []
    for course in selected:
        fetched = fetch_reports(args.service_url, args.service_token, course)
        all_reports.extend(fetched)

    grouped: Dict[Tuple[str, str, str], List[dict]] = {}
    for r in all_reports:
        chapter_id = str(r.get("grammar_chapter_id", "")).strip()
        theory_block_id = str(r.get("theory_block_id", "")).strip()
        if not chapter_id or not theory_block_id:
            continue
        course = detect_course(chapter_id)
        key = (course, chapter_id, theory_block_id)
        grouped.setdefault(key, []).append(r)

    changed_blocks: Dict[str, List[dict]] = {"blocks": []}
    resolved_ids: List[int] = []

    for (course, chapter_id, theory_block_id), reports in grouped.items():
        block_key = f"{course}:{chapter_id}:{theory_block_id}"
        llm = analyze_block_with_llama(args.llama_url, args.llama_model, block_key, reports)
        report_ids = [int(r["id"]) for r in reports if "id" in r]
        question_ids_local = set()
        for r in reports:
            qid = str(r.get("grammar_question_id", "")).strip()
            if not qid:
                continue
            question_ids_local.add(extract_local_question_id(qid, chapter_id, theory_block_id))

        rel = ""
        removed_count = 0
        before_hash = ""
        after_hash = ""
        action = "dry_run"
        err = ""
        try:
            index_data = load_training_pack_index(course_dirs[course])
            rel = block_to_relpath(index_data, chapter_id, theory_block_id)
            if not rel:
                raise RuntimeError("block not found in training pack index")
            qfile = course_dirs[course] / "training_pack" / "chapters" / rel
            if not qfile.exists():
                raise RuntimeError(f"questions file not found: {qfile}")
            if dry_run:
                raw = qfile.read_text(encoding="utf-8")
                before_hash = sha256_text(raw)
                after_hash = before_hash
                action = "dry_run"
            else:
                removed_count, before_hash, after_hash = remove_questions_in_file(qfile, question_ids_local)
                action = "removed" if removed_count > 0 else "noop"
                if removed_count > 0:
                    resolved_ids.extend(report_ids)
        except Exception as e:
            action = "error"
            err = str(e)

        row = {
            "timestamp": utc_now(),
            "run_id": run_id,
            "course": course,
            "chapter_id": chapter_id,
            "theory_block_id": theory_block_id,
            "question_ids": sorted(question_ids_local),
            "report_ids": report_ids,
            "llm_diagnosis": llm,
            "action": action,
            "error": err,
            "resolve_status": "pending",
            "hash_before": before_hash,
            "hash_after": after_hash,
            "training_pack_relpath": rel,
            "dry_run": dry_run,
        }
        append_jsonl(journal_path, row)

        if action in ("removed", "noop", "dry_run"):
            changed_blocks["blocks"].append({
                "course": course,
                "chapter_id": chapter_id,
                "theory_block_id": theory_block_id,
                "action": action,
                "report_ids": report_ids,
                "question_ids": sorted(question_ids_local),
            })

    # Resolve only when not dry-run and only for removed questions.
    if not dry_run and resolved_ids:
        payload = {"report_ids": sorted(set(resolved_ids)), "reason": "auto_cleanup_by_worker"}
        try:
            http_json("POST", f"{args.service_url.rstrip('/')}/api/internal/content-reports/grammar/resolve-bulk", args.service_token, payload)
            append_jsonl(journal_path, {
                "timestamp": utc_now(),
                "run_id": run_id,
                "event": "resolve_bulk",
                "report_ids": sorted(set(resolved_ids)),
                "status": "ok",
            })
        except Exception as e:
            append_jsonl(journal_path, {
                "timestamp": utc_now(),
                "run_id": run_id,
                "event": "resolve_bulk",
                "report_ids": sorted(set(resolved_ids)),
                "status": "error",
                "error": str(e),
            })

    write_json(changed_blocks_path, changed_blocks)
    print(
        json.dumps(
            {
                "run_id": run_id,
                "dry_run": dry_run,
                "groups": len(grouped),
                "journal": str(journal_path),
                "changed_blocks": str(changed_blocks_path),
            },
            ensure_ascii=False,
        )
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

