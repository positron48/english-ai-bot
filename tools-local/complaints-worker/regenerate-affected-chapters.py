#!/usr/bin/env python3
import argparse
import json
import os
import subprocess
from pathlib import Path
from typing import Dict, List


def read_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def latest_changed_blocks(logs_dir: Path) -> Path:
    files = sorted(logs_dir.glob("changed-theory-blocks-*.json"))
    if not files:
        raise FileNotFoundError("No changed-theory-blocks-*.json found")
    return files[-1]


def chapter_number_map(course_dir: Path) -> Dict[str, int]:
    chapters_dir = course_dir / "chapters"
    chapter_dirs = [p for p in chapters_dir.iterdir() if p.is_dir()]
    chapter_dirs.sort()
    mapping: Dict[str, int] = {}
    idx = 1
    for ch in chapter_dirs:
        name = ch.name
        cid = name[4:] if len(name) > 4 and name[:3].isdigit() and name[3] == "." else name
        if cid not in mapping:
            mapping[cid] = idx
            idx += 1
    return mapping


def collect_affected(data: dict) -> Dict[str, List[str]]:
    out = {"english": [], "spanish": []}
    seen = {"english": set(), "spanish": set()}
    for b in data.get("blocks", []):
        course = b.get("course", "")
        cid = b.get("chapter_id", "")
        if course not in out or not cid:
            continue
        if cid in seen[course]:
            continue
        seen[course].add(cid)
        out[course].append(cid)
    return out


def run(cmd: List[str], cwd: Path) -> subprocess.CompletedProcess:
    return subprocess.run(cmd, cwd=str(cwd), text=True, capture_output=True)


def main() -> int:
    parser = argparse.ArgumentParser(description="Regenerate training-pack only for chapters affected by complaints")
    parser.add_argument("--workspace", default=os.getenv("WORKSPACE_ROOT", str(Path(__file__).resolve().parents[2])))
    parser.add_argument("--changed-blocks-json", default="", help="Optional explicit changed-theory-blocks json path")
    parser.add_argument("--batch-size", type=int, default=10)
    parser.add_argument("--target-valid", type=int, default=1)
    args = parser.parse_args()

    workspace = Path(args.workspace)
    logs_dir = workspace / "logs" / "complaints"
    changed_path = Path(args.changed_blocks_json) if args.changed_blocks_json else latest_changed_blocks(logs_dir)
    changed = read_json(changed_path)
    affected = collect_affected(changed)

    results = {"changed_blocks": str(changed_path), "affected": affected, "runs": []}

    for course in ("english", "spanish"):
        course_dir = workspace / "courses" / f"{course}-grammar"
        mapping = chapter_number_map(course_dir)
        for cid in affected.get(course, []):
            ch_num = mapping.get(cid)
            if not ch_num:
                results["runs"].append({"course": course, "chapter_id": cid, "ok": False, "error": "chapter not found in mapping"})
                continue
            cmd = [
                "python3",
                "scripts/fill-training-pack.py",
                "--course-root",
                ".",
                "--chapter-number",
                str(ch_num),
                "--batch-size",
                str(args.batch_size),
                "--target-valid",
                str(args.target_valid),
            ]
            proc = run(cmd, cwd=course_dir)
            results["runs"].append(
                {
                    "course": course,
                    "chapter_id": cid,
                    "chapter_number": ch_num,
                    "ok": proc.returncode == 0,
                    "cmd": " ".join(cmd),
                    "stdout_tail": proc.stdout[-500:],
                    "stderr_tail": proc.stderr[-500:],
                }
            )
            if proc.returncode != 0:
                print(json.dumps(results, ensure_ascii=False))
                return 2

    print(json.dumps(results, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

