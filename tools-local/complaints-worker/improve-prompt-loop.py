#!/usr/bin/env python3
import argparse
import datetime as dt
import json
import os
import subprocess
import sys
from pathlib import Path
from typing import Dict, List, Tuple


def utc_now() -> str:
    return dt.datetime.now(dt.UTC).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def run_cmd(command: List[str], cwd: Path) -> Tuple[int, str, str]:
    proc = subprocess.Popen(
        command,
        cwd=str(cwd),
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        bufsize=1,
    )
    stdout_parts: List[str] = []
    stderr_parts: List[str] = []
    assert proc.stdout is not None
    assert proc.stderr is not None
    for line in proc.stdout:
        stdout_parts.append(line)
        sys.stdout.write(line)
        sys.stdout.flush()
    for line in proc.stderr:
        stderr_parts.append(line)
        sys.stderr.write(line)
        sys.stderr.flush()
    proc.wait()
    return proc.returncode, "".join(stdout_parts), "".join(stderr_parts)


def latest_matching(path: Path, pattern: str) -> Path:
    items = sorted(path.glob(pattern))
    if not items:
        raise FileNotFoundError(f"no files matching {pattern} in {path}")
    return items[-1]


def read_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def write_json(path: Path, data: dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def write_markdown_report(path: Path, payload: dict) -> None:
    lines = [
        f"# Improve Prompt Loop Report ({payload.get('status','unknown')})",
        "",
        f"- Generated at: {payload.get('generated_at','')}",
        f"- Loop id: {payload.get('loop_id','')}",
        f"- Status: {payload.get('status','')}",
        "",
        "## Affected Chapters",
    ]
    for course, chapters in (payload.get("affected_chapters", {}) or {}).items():
        lines.append(f"- {course}: {', '.join(chapters) if chapters else 'none'}")
    lines.append("")
    lines.append("## Iterations")
    for it in payload.get("iterations", []):
        lines.append(f"- iteration {it.get('iteration')}: strict_ok={it.get('strict_ok')} next_action={it.get('next_action','')}")
    if payload.get("manual_actions_required"):
        lines.append("")
        lines.append("## Manual Actions Required")
        lines.append("- Автоулучшение не стабилизировало качество за лимит итераций.")
        lines.append("- Проверь iteration-feedback и корректировки prompt/validator вручную.")
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def chapter_number_map(course_dir: Path) -> Dict[str, int]:
    chapters_dir = course_dir / "chapters"
    chapter_dirs = [p for p in chapters_dir.iterdir() if p.is_dir()]
    chapter_dirs.sort()
    mapping: Dict[str, int] = {}
    idx = 1
    for ch in chapter_dirs:
        name = ch.name
        if len(name) > 4 and name[3] == "." and name[:3].isdigit():
            cid = name[4:]
        else:
            cid = name
        if cid not in mapping:
            mapping[cid] = idx
            idx += 1
    return mapping


def collect_affected_chapters(changed_blocks: dict) -> Dict[str, List[str]]:
    out: Dict[str, List[str]] = {"english": [], "spanish": []}
    seen = {"english": set(), "spanish": set()}
    for b in changed_blocks.get("blocks", []):
        course = b.get("course", "")
        cid = b.get("chapter_id", "")
        if course not in out or not cid:
            continue
        if cid in seen[course]:
            continue
        seen[course].add(cid)
        out[course].append(cid)
    return out


def strict_iteration_ok(regression_ok: bool, smoke_ok: Dict[str, bool], validation_ok: Dict[str, bool]) -> bool:
    return regression_ok and smoke_ok.get("english", False) and smoke_ok.get("spanish", False) and validation_ok.get("english", False) and validation_ok.get("spanish", False)


def next_action(iteration: int, max_iterations: int, is_ok: bool) -> str:
    if is_ok:
        return "accept"
    if iteration >= max_iterations:
        return "stop_failed"
    return "retry"


def build_feedback(iteration: int, regression: dict, smoke: dict, validation: dict) -> dict:
    return {
        "timestamp": utc_now(),
        "iteration": iteration,
        "regression": regression,
        "smoke": smoke,
        "validation": validation,
    }


def run_regression(workspace: Path) -> dict:
    code, out, err = run_cmd(["python3", "tools-local/complaints-worker/prompt-validator-regression.py"], cwd=workspace)
    parsed = {}
    if out.strip():
        try:
            parsed = json.loads(out.strip().splitlines()[-1])
        except Exception:
            parsed = {"raw_stdout": out.strip()}
    return {"ok": code == 0, "exit_code": code, "stdout": out, "stderr": err, "parsed": parsed}


def run_smoke_for_course(workspace: Path, course: str) -> dict:
    code, out, err = run_cmd(
        ["python3", "tools-local/complaints-worker/prompt-llm-integration-smoke.py", "--course", course],
        cwd=workspace,
    )
    parsed = {}
    validation_ok = False
    if out.strip():
        try:
            parsed = json.loads(out.strip().splitlines()[-1])
            validation_ok = bool(parsed.get("validation_ok", False))
        except Exception:
            parsed = {"raw_stdout": out.strip()}
    return {"ok": code == 0 and validation_ok, "exit_code": code, "stdout": out, "stderr": err, "parsed": parsed, "validation_ok": validation_ok}


def run_targeted_fill(workspace: Path, affected: Dict[str, List[str]]) -> dict:
    result: Dict[str, dict] = {}
    for course in ("english", "spanish"):
        course_dir = workspace / "courses" / f"{course}-grammar"
        mapping = chapter_number_map(course_dir)
        result[course] = {"chapters": [], "commands": []}
        for cid in affected.get(course, []):
            ch_num = mapping.get(cid)
            if not ch_num:
                result[course]["chapters"].append({"chapter_id": cid, "status": "missing_in_mapping"})
                continue
            cmd = [
                "python3",
                "scripts/fill-training-pack.py",
                "--course-root",
                ".",
                "--chapter-number",
                str(ch_num),
                "--batch-size",
                "10",
                "--target-valid",
                "1",
            ]
            code, out, err = run_cmd(cmd, cwd=course_dir)
            result[course]["commands"].append({"chapter_id": cid, "chapter_number": ch_num, "cmd": " ".join(cmd), "exit_code": code})
            result[course]["chapters"].append({"chapter_id": cid, "chapter_number": ch_num, "ok": code == 0, "stderr_tail": err[-500:], "stdout_tail": out[-500:]})
            if code != 0:
                return {"ok": False, "details": result}
    return {"ok": True, "details": result}


def main() -> int:
    parser = argparse.ArgumentParser(description="Iterative prompt improvement loop for EN+ES")
    parser.add_argument("--workspace", default=os.getenv("WORKSPACE_ROOT", str(Path(__file__).resolve().parents[2])))
    parser.add_argument("--max-iterations", type=int, default=3)
    args = parser.parse_args()

    workspace = Path(args.workspace)
    logs_dir = workspace / "logs" / "complaints"
    logs_dir.mkdir(parents=True, exist_ok=True)
    loop_id = dt.datetime.now(dt.UTC).strftime("%Y%m%d%H%M%S")

    # 1) Apply complaints for both profiles.
    code_apply, out_apply, err_apply = run_cmd(["make", "complaints-apply-both"], cwd=workspace)
    if code_apply != 0:
        write_json(
            logs_dir / f"improve-loop-failure-{loop_id}.json",
            {"status": "failed", "stage": "complaints-apply-both", "exit_code": code_apply, "stdout": out_apply, "stderr": err_apply, "generated_at": utc_now()},
        )
        return 2

    # 2) Initial improvement plan from latest journal/run.
    code_plan, out_plan, err_plan = run_cmd(["python3", "tools-local/complaints-worker/analyze-journal.py"], cwd=workspace)
    if code_plan != 0:
        write_json(
            logs_dir / f"improve-loop-failure-{loop_id}.json",
            {"status": "failed", "stage": "analyze-journal", "exit_code": code_plan, "stdout": out_plan, "stderr": err_plan, "generated_at": utc_now()},
        )
        return 2

    plan_meta = json.loads(out_plan.strip().splitlines()[-1])
    plan_json = Path(plan_meta["plan_json"])
    changed_blocks = read_json(latest_matching(logs_dir, "changed-theory-blocks-*.json"))
    affected = collect_affected_chapters(changed_blocks)

    iterations: List[dict] = []
    feedback_path = ""

    for i in range(1, args.max_iterations + 1):
        # Apply prompt changes for both courses.
        course_updates = {}
        for course in ("english", "spanish"):
            cmd = [
                "python3",
                "tools-local/complaints-worker/apply-prompt-improvements.py",
                "--course",
                course,
                "--plan-json",
                str(plan_json),
            ]
            if feedback_path:
                cmd.extend(["--feedback-json", feedback_path])
            c, out, err = run_cmd(cmd, cwd=workspace)
            course_updates[course] = {"ok": c == 0, "exit_code": c, "stdout": out, "stderr": err}
            if c != 0:
                fb = build_feedback(i, {"ok": False, "reason": "prompt update failed"}, {"english": {"ok": False}, "spanish": {"ok": False}}, {"english": {"ok": False}, "spanish": {"ok": False}})
                feedback_path = str(logs_dir / f"iteration-feedback-{loop_id}-{i}.json")
                write_json(Path(feedback_path), fb)
                iterations.append({"iteration": i, "course_updates": course_updates, "feedback_path": feedback_path, "ok": False})
                continue

        regression = run_regression(workspace)
        smoke_en = run_smoke_for_course(workspace, "english")
        smoke_es = run_smoke_for_course(workspace, "spanish")
        smoke_ok = {"english": smoke_en["ok"], "spanish": smoke_es["ok"]}
        validation_ok = {"english": smoke_en["validation_ok"], "spanish": smoke_es["validation_ok"]}
        iter_ok = strict_iteration_ok(regression["ok"], smoke_ok, validation_ok)

        iter_info = {
            "iteration": i,
            "course_updates": course_updates,
            "regression": regression,
            "smoke": {"english": smoke_en, "spanish": smoke_es},
            "strict_ok": iter_ok,
        }

        action = next_action(i, args.max_iterations, iter_ok)
        iter_info["next_action"] = action
        if action == "accept":
            regen = run_targeted_fill(workspace, affected)
            iter_info["targeted_regeneration"] = regen
            if regen["ok"]:
                final_smoke_en = run_smoke_for_course(workspace, "english")
                final_smoke_es = run_smoke_for_course(workspace, "spanish")
                final_ok = final_smoke_en["ok"] and final_smoke_es["ok"]
                iter_info["final_smoke"] = {"english": final_smoke_en, "spanish": final_smoke_es, "ok": final_ok}
                if final_ok:
                    success = {
                        "status": "success",
                        "generated_at": utc_now(),
                        "loop_id": loop_id,
                        "iterations_used": i,
                        "affected_chapters": affected,
                        "plan_json": str(plan_json),
                        "iterations": iterations + [iter_info],
                    }
                    success_json = logs_dir / f"improve-loop-success-{loop_id}.json"
                    success_md = logs_dir / f"improve-loop-success-{loop_id}.md"
                    write_json(success_json, success)
                    write_markdown_report(success_md, success)
                    print(json.dumps({"status": "success", "loop_id": loop_id, "iterations": i, "report_json": str(success_json), "report_md": str(success_md)}, ensure_ascii=False))
                    return 0

        feedback = build_feedback(i, regression, {"english": smoke_en, "spanish": smoke_es}, {"english": {"ok": smoke_en["validation_ok"]}, "spanish": {"ok": smoke_es["validation_ok"]}})
        feedback_path = str(logs_dir / f"iteration-feedback-{loop_id}-{i}.json")
        write_json(Path(feedback_path), feedback)
        iter_info["feedback_path"] = feedback_path
        iterations.append(iter_info)

    failure = {
        "status": "failed",
        "generated_at": utc_now(),
        "loop_id": loop_id,
        "reason": "3 iterations failed to reach strict OK",
        "manual_actions_required": True,
        "affected_chapters": affected,
        "plan_json": str(plan_json),
        "iterations": iterations,
    }
    fail_json = logs_dir / f"improve-loop-failure-{loop_id}.json"
    fail_md = logs_dir / f"improve-loop-failure-{loop_id}.md"
    write_json(fail_json, failure)
    write_markdown_report(fail_md, failure)
    print(json.dumps({"status": "failed", "loop_id": loop_id, "iterations": args.max_iterations, "report_json": str(fail_json), "report_md": str(fail_md)}, ensure_ascii=False))
    return 2


if __name__ == "__main__":
    raise SystemExit(main())

