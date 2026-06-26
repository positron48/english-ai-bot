#!/usr/bin/env python3
import argparse
import datetime as dt
import json
import os
import re
import subprocess
import time
from pathlib import Path
from typing import Dict, List


def utc_now() -> str:
    return dt.datetime.now(dt.UTC).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def read_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def write_json(path: Path, data: dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def discover_scenarios(course_dir: Path, limit: int) -> List[dict]:
    scenarios: List[dict] = []
    chapters_dir = course_dir / "chapters"
    chapter_dirs = sorted([p for p in chapters_dir.iterdir() if p.is_dir()])
    for idx, ch_dir in enumerate(chapter_dirs, start=1):
        final_path = ch_dir / "05-final.json"
        if not final_path.exists():
            final_path = ch_dir / "04-final.json"
            if not final_path.exists():
                continue
        try:
            ch = read_json(final_path)
        except Exception:
            continue
        blocks = [b for b in ch.get("blocks", []) if isinstance(b, dict) and b.get("type") == "theory"]
        if not blocks:
            continue
        scenarios.append(
            {
                "chapter_number": idx,
                "block_number": 1,
                "chapter_id": ch.get("id", ""),
                "theory_block_id": blocks[0].get("id", ""),
            }
        )
        if len(scenarios) >= limit:
            break
    return scenarios


def run_scenario(course_dir: Path, scenario: dict, qpb: int, llm_url: str, llm_model: str) -> dict:
    cmd = [
        "python3",
        "scripts/generate-training-pack.py",
        "--course-root",
        ".",
        "--append",
        "--chapter-number",
        str(scenario["chapter_number"]),
        "--block-number",
        str(scenario["block_number"]),
        "--questions-per-block",
        str(qpb),
        "--llm-base-url",
        llm_url,
        "--llm-model",
        llm_model,
    ]
    start = time.time()
    proc = subprocess.run(cmd, cwd=str(course_dir), text=True, capture_output=True)
    elapsed_ms = int((time.time() - start) * 1000)

    added, rejected = 0, 0
    for m in re.finditer(r"добавлено\s+(\d+),\s*отклонено\s+(\d+)", proc.stdout):
        added += int(m.group(1))
        rejected += int(m.group(2))

    validation_path = course_dir / "training_pack" / "reports" / "validation-report.json"
    validation_ok = False
    weak_blocks = 0
    chapter_errors = 0
    if validation_path.exists():
        try:
            vr = read_json(validation_path)
            validation_ok = bool(vr.get("ok", False))
            weak_blocks = len(vr.get("weak_blocks", []) or [])
            chapter_errors = sum(1 for _, ch in (vr.get("chapters", {}) or {}).items() if (ch.get("errors") or []))
        except Exception:
            pass

    return {
        "scenario": scenario,
        "ok": proc.returncode == 0 and validation_ok,
        "exit_code": proc.returncode,
        "elapsed_ms": elapsed_ms,
        "added": added,
        "rejected": rejected,
        "validation_ok": validation_ok,
        "weak_blocks": weak_blocks,
        "chapter_errors": chapter_errors,
        "stdout_tail": proc.stdout[-700:],
        "stderr_tail": proc.stderr[-700:],
    }


def aggregate(course_runs: List[dict]) -> dict:
    total_added = sum(r["added"] for r in course_runs)
    total_rejected = sum(r["rejected"] for r in course_runs)
    denom = max(1, total_added + total_rejected)
    rejected_rate = total_rejected / denom
    failed_runs = sum(1 for r in course_runs if not r["ok"])
    return {
        "runs": len(course_runs),
        "failed_runs": failed_runs,
        "total_added": total_added,
        "total_rejected": total_rejected,
        "rejected_rate": rejected_rate,
    }


def compare_with_baseline(current: dict, baseline: dict, max_rejected_rate_delta: float, max_failed_runs_delta: int) -> dict:
    cur_m = current["metrics"]
    base_m = baseline.get("metrics", {})
    cur_rate = float(cur_m.get("rejected_rate", 0.0))
    base_rate = float(base_m.get("rejected_rate", 0.0))
    cur_failed = int(cur_m.get("failed_runs", 0))
    base_failed = int(base_m.get("failed_runs", 0))
    degraded = False
    reasons: List[str] = []
    if (cur_rate - base_rate) > max_rejected_rate_delta:
        degraded = True
        reasons.append(f"rejected_rate increased from {base_rate:.3f} to {cur_rate:.3f}")
    if (cur_failed - base_failed) > max_failed_runs_delta:
        degraded = True
        reasons.append(f"failed_runs increased from {base_failed} to {cur_failed}")
    return {"degraded": degraded, "reasons": reasons, "base_metrics": base_m}


def main() -> int:
    parser = argparse.ArgumentParser(description="Quality regression gate for complaints prompt loop")
    parser.add_argument("--workspace", default=os.getenv("WORKSPACE_ROOT", str(Path(__file__).resolve().parents[2])))
    parser.add_argument("--scenarios-per-course", type=int, default=int(os.getenv("COMPLAINTS_QUALITY_SCENARIOS", "5")))
    parser.add_argument("--questions-per-block", type=int, default=int(os.getenv("COMPLAINTS_QUALITY_QPB", "3")))
    parser.add_argument("--llm-base-url", default=os.getenv("LLAMACPP_URL", "http://127.0.0.1:8090"))
    parser.add_argument("--llm-model", default=os.getenv("LLAMACPP_MODEL", "qwen3:30b"))
    parser.add_argument("--baseline-path", default="", help="Optional baseline json path")
    parser.add_argument("--set-baseline", action="store_true", help="Write current metrics as baseline")
    parser.add_argument("--max-rejected-rate-delta", type=float, default=float(os.getenv("COMPLAINTS_QUALITY_MAX_REJECTED_RATE_DELTA", "0.20")))
    parser.add_argument("--max-failed-runs-delta", type=int, default=int(os.getenv("COMPLAINTS_QUALITY_MAX_FAILED_RUNS_DELTA", "0")))
    args = parser.parse_args()

    workspace = Path(args.workspace)
    logs_dir = workspace / "logs" / "complaints"
    logs_dir.mkdir(parents=True, exist_ok=True)
    ts = dt.datetime.now(dt.UTC).strftime("%Y%m%d%H%M%S")
    out_path = logs_dir / f"quality-regression-{ts}.json"
    baseline_path = Path(args.baseline_path) if args.baseline_path else (logs_dir / "quality-baseline.json")

    all_runs: Dict[str, List[dict]] = {"english": [], "spanish": []}
    for course in ("english", "spanish"):
        course_dir = workspace / "courses" / f"{course}-grammar"
        scenarios = discover_scenarios(course_dir, args.scenarios_per_course)
        for sc in scenarios:
            all_runs[course].append(run_scenario(course_dir, sc, args.questions_per_block, args.llm_base_url, args.llm_model))

    flat_runs = all_runs["english"] + all_runs["spanish"]
    metrics = aggregate(flat_runs)
    report = {
        "generated_at": utc_now(),
        "llm_base_url": args.llm_base_url,
        "llm_model": args.llm_model,
        "scenarios_per_course": args.scenarios_per_course,
        "questions_per_block": args.questions_per_block,
        "runs": all_runs,
        "metrics": metrics,
    }
    write_json(out_path, report)

    if args.set_baseline:
        write_json(baseline_path, report)
        print(json.dumps({"status": "baseline_saved", "baseline": str(baseline_path), "report": str(out_path), "metrics": metrics}, ensure_ascii=False))
        return 0

    if baseline_path.exists():
        baseline = read_json(baseline_path)
        cmp = compare_with_baseline(report, baseline, args.max_rejected_rate_delta, args.max_failed_runs_delta)
        report["baseline_compare"] = cmp
        write_json(out_path, report)
        if cmp["degraded"]:
            print(json.dumps({"status": "degraded", "report": str(out_path), "baseline": str(baseline_path), "reasons": cmp["reasons"], "metrics": metrics}, ensure_ascii=False))
            return 2

    print(json.dumps({"status": "ok", "report": str(out_path), "baseline": str(baseline_path), "metrics": metrics}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

