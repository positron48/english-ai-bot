#!/usr/bin/env python3
import argparse
import datetime as dt
import json
import os
import re
import urllib.request
from pathlib import Path
from typing import List

START_MARKER = "<!-- AUTO_COMPLAINTS_GUARDRAILS:START -->"
END_MARKER = "<!-- AUTO_COMPLAINTS_GUARDRAILS:END -->"


def read_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def latest_plan(workspace: Path) -> Path:
    logs_dir = workspace / "logs" / "complaints"
    plans = sorted(logs_dir.glob("improvement-plan-*.json"))
    if not plans:
        raise FileNotFoundError("No improvement plan found in logs/complaints")
    return plans[-1]


def call_llm(llama_url: str, llama_model: str, payload_obj: dict) -> dict:
    payload = {
        "model": llama_model,
        "messages": [
            {"role": "system", "content": "Return only JSON."},
            {"role": "user", "content": json.dumps(payload_obj, ensure_ascii=False)},
        ],
        "temperature": 0.1,
    }
    req = urllib.request.Request(
        url=f"{llama_url.rstrip('/')}/v1/chat/completions",
        method="POST",
        headers={"Content-Type": "application/json"},
        data=json.dumps(payload).encode("utf-8"),
    )
    with urllib.request.urlopen(req, timeout=120) as resp:
        data = json.loads(resp.read().decode("utf-8"))
    text = data["choices"][0]["message"]["content"].strip()
    if text.startswith("```"):
        text = re.sub(r"^```[a-zA-Z0-9]*\n?", "", text)
        text = re.sub(r"\n?```$", "", text)
    return json.loads(text)


def bounded(items: List[str], limit: int) -> List[str]:
    out: List[str] = []
    seen = set()
    for item in items:
        s = " ".join(str(item).split())
        if not s or s in seen:
            continue
        seen.add(s)
        out.append(s)
        if len(out) >= limit:
            break
    return out


def render_auto_block(plan: dict, compact: dict) -> str:
    generated_at = dt.datetime.now(dt.UTC).strftime("%Y-%m-%dT%H:%M:%SZ")
    lines: List[str] = []
    lines.append(START_MARKER)
    lines.append("## Auto Complaints Guardrails")
    lines.append(f"- Обновлено автоматически: {generated_at}")
    lines.append("- Этот блок полностью перезаписывается скриптом; не расширяй его вручную.")
    lines.append("")
    lines.append("### Частые паттерны ошибок")
    for row in bounded(compact.get("patterns", []), 4):
        lines.append(f"- {row}")
    lines.append("")
    lines.append("### Дополнительные правила генерации")
    for row in bounded(compact.get("prompt_rules", []), 6):
        lines.append(f"- {row}")
    lines.append("")
    lines.append("### Проверки перед финальным JSON")
    for row in bounded(compact.get("validation_rules", []), 6):
        lines.append(f"- {row}")
    lines.append(END_MARKER)
    return "\n".join(lines) + "\n"


def upsert_auto_block(prompt_text: str, block: str) -> str:
    pattern = re.compile(re.escape(START_MARKER) + r".*?" + re.escape(END_MARKER) + r"\n?", re.S)
    if pattern.search(prompt_text):
        return pattern.sub(block, prompt_text)
    tail = "" if prompt_text.endswith("\n") else "\n"
    return prompt_text + tail + "\n" + block


def main() -> int:
    parser = argparse.ArgumentParser(description="Apply automated complaints improvements to generator prompt")
    parser.add_argument("--workspace", default=os.getenv("WORKSPACE_ROOT", str(Path(__file__).resolve().parents[2])))
    parser.add_argument("--course", choices=["spanish", "english"], default="spanish")
    parser.add_argument("--plan-json", default="", help="Path to improvement-plan json (optional)")
    parser.add_argument("--llama-url", default=os.getenv("LLAMACPP_URL", "http://127.0.0.1:8080"))
    parser.add_argument("--llama-model", default=os.getenv("LLAMACPP_MODEL", "local-model"))
    args = parser.parse_args()

    workspace = Path(args.workspace)
    plan_path = Path(args.plan_json) if args.plan_json else latest_plan(workspace)
    data = read_json(plan_path)
    plan = data.get("plan", {})
    course_dir = workspace / "courses" / f"{args.course}-grammar"
    prompt_path = course_dir / "prompts" / "16-training-pack-generator-system.md"
    old_text = prompt_path.read_text(encoding="utf-8")

    llm_payload = {
        "task": "Compress improvement plan into short prompt+validation rules for Spanish grammar question generation.",
        "constraints": [
            "Output JSON only with fields: patterns, prompt_rules, validation_rules",
            "Each list should have max 6 items",
            "Each item must be short (<= 140 chars), actionable, and non-duplicative",
            "Do not include markdown fencing",
        ],
        "improvement_plan": plan,
    }
    compact = call_llm(args.llama_url, args.llama_model, llm_payload)
    auto_block = render_auto_block(plan, compact)
    new_text = upsert_auto_block(old_text, auto_block)
    prompt_path.write_text(new_text, encoding="utf-8")

    print(
        json.dumps(
            {
                "course": args.course,
                "prompt_path": str(prompt_path),
                "plan_path": str(plan_path),
                "updated": old_text != new_text,
                "auto_block_lines": len(auto_block.splitlines()),
            },
            ensure_ascii=False,
        )
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

