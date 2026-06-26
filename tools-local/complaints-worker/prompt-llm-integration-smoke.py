#!/usr/bin/env python3
import argparse
import json
import os
import subprocess
from pathlib import Path


def run(cmd: list[str], cwd: Path) -> None:
    subprocess.run(cmd, cwd=str(cwd), check=True)


def main() -> int:
    parser = argparse.ArgumentParser(description="LLM integration smoke for Spanish prompt+validator")
    parser.add_argument("--workspace", default=os.getenv("WORKSPACE_ROOT", str(Path(__file__).resolve().parents[2])))
    parser.add_argument("--course", choices=["spanish", "english"], default=os.getenv("COMPLAINTS_SMOKE_COURSE", "spanish"))
    parser.add_argument("--chapter-number", type=int, default=int(os.getenv("COMPLAINTS_SMOKE_CHAPTER", "1")))
    parser.add_argument("--block-number", type=int, default=int(os.getenv("COMPLAINTS_SMOKE_BLOCK", "1")))
    parser.add_argument("--questions-per-block", type=int, default=int(os.getenv("COMPLAINTS_SMOKE_QPB", "1")))
    parser.add_argument("--llm-base-url", default=os.getenv("LLAMACPP_URL", "http://127.0.0.1:8090"))
    parser.add_argument("--llm-model", default=os.getenv("LLAMACPP_MODEL", "qwen3:30b"))
    args = parser.parse_args()

    workspace = Path(args.workspace)
    course_root = workspace / "courses" / f"{args.course}-grammar"
    gen_script = course_root / "scripts" / "generate-training-pack.py"
    report_path = course_root / "training_pack" / "reports" / "validation-report.json"

    run(
        [
            "python3",
            str(gen_script),
            "--course-root",
            ".",
            "--append",
            "--chapter-number",
            str(args.chapter_number),
            "--block-number",
            str(args.block_number),
            "--questions-per-block",
            str(args.questions_per_block),
            "--llm-base-url",
            args.llm_base_url,
            "--llm-model",
            args.llm_model,
        ],
        cwd=course_root,
    )

    if not report_path.exists():
        raise SystemExit("validation-report.json not found after generation")
    report = json.loads(report_path.read_text(encoding="utf-8"))
    if not report.get("ok", False):
        raise SystemExit("validation-report indicates failure")

    print(
        json.dumps(
            {
                "status": "ok",
                "course": args.course,
                "chapter_number": args.chapter_number,
                "block_number": args.block_number,
                "questions_per_block": args.questions_per_block,
                "llm_base_url": args.llm_base_url,
                "llm_model": args.llm_model,
                "validation_ok": True,
                "validation_report_path": str(report_path),
            },
            ensure_ascii=False,
        )
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

