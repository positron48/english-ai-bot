#!/usr/bin/env python3
import importlib.util
import json
import re
from pathlib import Path


START_MARKER = "<!-- AUTO_COMPLAINTS_GUARDRAILS:START -->"
END_MARKER = "<!-- AUTO_COMPLAINTS_GUARDRAILS:END -->"


def load_module(path: Path):
    spec = importlib.util.spec_from_file_location("gen_pack", str(path))
    mod = importlib.util.module_from_spec(spec)
    assert spec and spec.loader
    spec.loader.exec_module(mod)
    return mod


def assert_true(cond: bool, msg: str):
    if not cond:
        raise AssertionError(msg)


def main() -> int:
    root = Path(__file__).resolve().parents[2]
    prompt_path = root / "courses" / "spanish-grammar" / "prompts" / "16-training-pack-generator-system.md"
    gen_path = root / "courses" / "spanish-grammar" / "scripts" / "generate-training-pack.py"

    prompt = prompt_path.read_text(encoding="utf-8")
    assert_true("только JSON-массив" in prompt, "prompt must keep JSON-only requirement")
    assert_true("type=mcq_single" in prompt, "prompt must keep mcq_single requirement")
    assert_true(START_MARKER in prompt and END_MARKER in prompt, "auto guardrails markers must exist")
    assert_true(prompt.count(START_MARKER) == 1 and prompt.count(END_MARKER) == 1, "markers must be unique")

    auto_block = re.search(re.escape(START_MARKER) + r".*?" + re.escape(END_MARKER), prompt, flags=re.S)
    assert_true(auto_block is not None, "auto block must be present")
    auto_lines = [ln for ln in auto_block.group(0).splitlines() if ln.strip().startswith("- ")]
    assert_true(len(auto_lines) <= 20, "auto block must stay compact and bounded")

    mod = load_module(gen_path)
    validate_question = getattr(mod, "validate_question")

    ok_q = {
        "id": "q1",
        "type": "mcq_single",
        "prompt": "Выбери правильный вариант: Сегодня понедельник.",
        "choices": [{"id": "a", "text": "Hoy es lunes"}, {"id": "b", "text": "Hoy está lunes"}],
        "correct_answer": "a",
        "explanation": "С днями недели используется ser.",
        "theory_block_id": "b1",
        "chapter_id": "es.test.chapter",
        "concept_id": "ser_estar",
        "difficulty": 2,
    }
    bad_q = {
        "id": "q2",
        "type": "mcq_single",
        "prompt": "Choose the answer",  # no cyrillic
        "choices": [{"id": "a", "text": "A"}, {"id": "a", "text": "A"}],  # duplicate ids and text
        "correct_answer": "c",
        "explanation": "",
        "theory_block_id": "missing_block",
        "chapter_id": "other.chapter",
        "concept_id": "",
        "difficulty": 2,
    }

    ok_errs = validate_question(ok_q, "es.test.chapter", {"b1"})
    bad_errs = validate_question(bad_q, "es.test.chapter", {"b1"})
    assert_true(len(ok_errs) == 0, f"valid question should pass, got: {json.dumps(ok_errs, ensure_ascii=False)}")
    assert_true(len(bad_errs) >= 3, "invalid question should produce multiple validation errors")

    print(json.dumps({"status": "ok", "checks": 8}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

