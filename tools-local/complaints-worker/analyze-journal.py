#!/usr/bin/env python3
import argparse
import datetime as dt
import json
import os
import re
import urllib.request
from pathlib import Path
from typing import Dict, List

ANSI_PROMPT = "\033[35m"  # magenta
ANSI_RESPONSE = "\033[32m"  # green
ANSI_RESET = "\033[0m"


def utc_now() -> str:
    return dt.datetime.now(dt.UTC).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def parse_llm_json(text: str) -> dict:
    text = text.strip()
    if text.startswith("```"):
        text = re.sub(r"^```[a-zA-Z0-9]*\n?", "", text)
        text = re.sub(r"\n?```$", "", text)
    try:
        return json.loads(text)
    except Exception:
        m = re.search(r"\{.*\}", text, flags=re.S)
        if not m:
            return {"summary": "failed to parse llm json", "raw": text}
        try:
            return json.loads(m.group(0))
        except Exception:
            return {"summary": "failed to parse llm json", "raw": text}


def read_jsonl(path: Path) -> List[dict]:
    rows: List[dict] = []
    if not path.exists():
        return rows
    with path.open("r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                rows.append(json.loads(line))
            except Exception:
                continue
    return rows


def latest_journal(workspace: Path) -> Path:
    logs_dir = workspace / "logs" / "complaints"
    candidates = sorted(logs_dir.glob("complaints-*.jsonl"))
    if not candidates:
        return logs_dir / f"complaints-{dt.datetime.now(dt.UTC).strftime('%Y-%m')}.jsonl"
    return candidates[-1]


def call_llm(llama_url: str, llama_model: str, payload_obj: dict) -> dict:
    payload = {
        "model": llama_model,
        "messages": [
            {"role": "system", "content": "Return only JSON."},
            {"role": "user", "content": json.dumps(payload_obj, ensure_ascii=False)},
        ],
        "temperature": 0.1,
    }
    print(f"{ANSI_PROMPT}[LLM REQUEST][analyze-journal] {json.dumps(payload, ensure_ascii=False)}{ANSI_RESET}")
    req = urllib.request.Request(
        url=f"{llama_url.rstrip('/')}/v1/chat/completions",
        method="POST",
        headers={"Content-Type": "application/json"},
        data=json.dumps(payload).encode("utf-8"),
    )
    with urllib.request.urlopen(req, timeout=120) as resp:
        data = json.loads(resp.read().decode("utf-8"))
    content = data["choices"][0]["message"]["content"]
    print(f"{ANSI_RESPONSE}[LLM RESPONSE][analyze-journal] {content}{ANSI_RESET}")
    return parse_llm_json(content)


def read_prompt_excerpt(workspace: Path, course: str, limit: int = 6000) -> str:
    prompt_path = workspace / "courses" / f"{course}-grammar" / "prompts" / "16-training-pack-generator-system.md"
    if not prompt_path.exists():
        return ""
    text = prompt_path.read_text(encoding="utf-8")
    if len(text) <= limit:
        return text
    return text[-limit:]


def build_markdown(plan: dict, run_id: str, source_journal: str) -> str:
    lines: List[str] = []
    lines.append(f"# Complaints Improvement Plan ({run_id})")
    lines.append("")
    lines.append(f"- Generated at: {utc_now()}")
    lines.append(f"- Source journal: `{source_journal}`")
    lines.append("")

    patterns = plan.get("global_patterns", []) or []
    lines.append("## Global Patterns")
    if patterns:
        for p in patterns:
            lines.append(f"- {p}")
    else:
        lines.append("- No stable patterns identified.")
    lines.append("")

    lines.append("## Prompt Updates")
    prompt_updates = plan.get("prompt_updates", []) or []
    if prompt_updates:
        for p in prompt_updates:
            lines.append(f"- {p}")
    else:
        lines.append("- No prompt updates proposed.")
    lines.append("")

    lines.append("## Validator Updates")
    validator_updates = plan.get("validator_updates", []) or []
    if validator_updates:
        for v in validator_updates:
            lines.append(f"- {v}")
    else:
        lines.append("- No validator updates proposed.")
    lines.append("")

    lines.append("## Block-Level Actions")
    block_actions = plan.get("block_actions", []) or []
    if block_actions:
        for a in block_actions:
            if isinstance(a, dict):
                block = a.get("block", "unknown")
                action = a.get("action", "")
                reason = a.get("reason", "")
                lines.append(f"- `{block}`: {action} ({reason})")
            else:
                lines.append(f"- {a}")
    else:
        lines.append("- No block-level actions proposed.")
    lines.append("")

    lines.append("## Proposed Execution Order")
    execution = plan.get("execution_order", []) or []
    if execution:
        for e in execution:
            lines.append(f"- {e}")
    else:
        lines.append("- 1) complaints-apply-both 2) training-pack-fill 3) grammar-bundle 4) make check")
    lines.append("")
    return "\n".join(lines) + "\n"


def main() -> int:
    parser = argparse.ArgumentParser(description="Analyze complaints journal and build improvement plan")
    parser.add_argument("--workspace", default=os.getenv("WORKSPACE_ROOT", str(Path(__file__).resolve().parents[2])))
    parser.add_argument("--journal", default="", help="Path to complaints jsonl. If empty, latest file is used.")
    parser.add_argument("--run-id", default="", help="Limit analysis to specific run_id (optional).")
    parser.add_argument("--llama-url", default=os.getenv("LLAMACPP_URL", "http://127.0.0.1:8080"))
    parser.add_argument("--llama-model", default=os.getenv("LLAMACPP_MODEL", "local-model"))
    args = parser.parse_args()

    workspace = Path(args.workspace)
    journal = Path(args.journal) if args.journal else latest_journal(workspace)
    rows = read_jsonl(journal)

    rows = [r for r in rows if isinstance(r.get("llm_diagnosis"), dict) and r.get("action") in ("removed", "noop", "dry_run")]
    if args.run_id:
        rows = [r for r in rows if r.get("run_id") == args.run_id]
    elif rows:
        # By default analyze latest run_id in file.
        last_run = rows[-1].get("run_id", "")
        if last_run:
            rows = [r for r in rows if r.get("run_id") == last_run]

    grouped: Dict[str, dict] = {}
    for r in rows:
        key = f"{r.get('course','')}::{r.get('chapter_id','')}::{r.get('theory_block_id','')}"
        bucket = grouped.setdefault(
            key,
            {
                "block": key,
                "count": 0,
                "report_ids": [],
                "question_ids": [],
                "summaries": [],
                "root_causes": [],
                "recommended_fixes": [],
            },
        )
        bucket["count"] += 1
        bucket["report_ids"].extend(r.get("report_ids", []) or [])
        bucket["question_ids"].extend(r.get("question_ids", []) or [])
        diag = r.get("llm_diagnosis", {}) or {}
        if diag.get("summary"):
            bucket["summaries"].append(diag["summary"])
        bucket["root_causes"].extend(diag.get("root_causes", []) or [])
        if diag.get("recommended_fix"):
            bucket["recommended_fixes"].append(diag["recommended_fix"])

    prompt_context = {
        "english": read_prompt_excerpt(workspace, "english"),
        "spanish": read_prompt_excerpt(workspace, "spanish"),
    }

    payload_obj = {
        "task": "Analyze grammar complaints and propose concrete system improvements.",
        "required_output_json_fields": [
            "global_patterns",
            "prompt_updates",
            "validator_updates",
            "block_actions",
            "execution_order",
        ],
        "constraints": [
            "Focus on recurring issues, not one-off noise.",
            "Propose changes that can be applied in prompts and validation logic.",
            "Keep recommendations concise and executable.",
        ],
        "current_generator_prompts_excerpt": prompt_context,
        "blocks": list(grouped.values()),
    }

    plan = call_llm(args.llama_url, args.llama_model, payload_obj)

    out_dir = workspace / "logs" / "complaints"
    out_dir.mkdir(parents=True, exist_ok=True)
    ts = dt.datetime.now(dt.UTC).strftime("%Y%m%d%H%M%S")
    rid = args.run_id or (rows[-1].get("run_id", "all") if rows else "empty")
    json_path = out_dir / f"improvement-plan-{ts}.json"
    md_path = out_dir / f"improvement-plan-{ts}.md"

    json_payload = {
        "generated_at": utc_now(),
        "journal": str(journal),
        "run_id": rid,
        "blocks_count": len(grouped),
        "rows_count": len(rows),
        "plan": plan,
    }
    json_path.write_text(json.dumps(json_payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    md_path.write_text(build_markdown(plan, rid, str(journal)), encoding="utf-8")

    print(
        json.dumps(
            {
                "journal": str(journal),
                "run_id": rid,
                "rows": len(rows),
                "blocks": len(grouped),
                "plan_json": str(json_path),
                "plan_md": str(md_path),
            },
            ensure_ascii=False,
        )
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

