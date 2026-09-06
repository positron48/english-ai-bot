#!/usr/bin/env python3
"""Read-only validation of the ES 065-068 editorial batch against its baseline."""
import importlib.util
import json
import pathlib
import subprocess
from collections import Counter

ROOT = pathlib.Path(__file__).resolve().parents[4]
BATCH = pathlib.Path(__file__).resolve().parent
BASE = "501044644954d5aaa474f7d1eaee99518a4ebd9d"
spec = importlib.util.spec_from_file_location("grammar_review", ROOT / "scripts/grammar-review.py")
gr = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gr)
units = [u for u in gr.inventory() if u["language"] == "es" and u["order"].startswith(("065.", "066.", "067.", "068.")) and u["bank"] != "verbs"]
assert len(units) == 28 and sum(u["count"] for u in units) == 480

def baseline(path):
    rel = path.relative_to(ROOT / "courses/spanish-grammar").as_posix()
    data = json.loads(subprocess.check_output(["git", "-C", str(ROOT / "courses/spanish-grammar"), "show", f"{BASE}:{rel}"], text=True))
    return data

live = {u["source"]: u for u in gr.inventory()}
totals = Counter()
for i, unit in enumerate(units):
    path = ROOT / unit["source"]
    old = baseline(path)
    now = gr.read(path)
    report = gr.read(ROOT / unit["report"])
    assert report["source"] == unit["source"]
    assert report["source_sha256"] == unit["source_sha256"], (path, "stale source fingerprint")
    assert report["context_sha256"] == unit["context_sha256"], (path, "stale context fingerprint")
    assert {k:v for k,v in old.items() if k != "questions"} == {k:v for k,v in now.items() if k != "questions"}, path
    old_questions, questions = old["questions"], now["questions"]
    assert len(questions) == len(old_questions) == unit["count"]
    def identity(q):
        return (q["id"], q["type"], q.get("difficulty"), q.get("theory_block_id"), q.get("chapter_id"), q.get("concept_id"), [c["id"] for c in q.get("choices", [])])
    assert list(map(identity, old_questions)) == list(map(identity, questions)), path
    assert len({q["id"] for q in questions}) == len(questions)
    chapter_path = next(ROOT / x["source"] for x in units if x["chapter"] == unit["chapter"] and x["bank"] == "chapter")
    blocks = {b["id"] for b in gr.read(chapter_path.parent / "01-outline.json")["chapter_outline"]["theory_blocks"]}
    for before, question, row in zip(old_questions, questions, report["questions"]):
        assert row["id"] == question["id"]
        assert (before != question) == (row["decision"] == "fixed"), (path, question["id"], row["decision"])
        assert row["decision"] in ("ok", "fixed") and row["note"] and row["verification"] == "pending"
        without_signature_before = {k:v for k,v in before.items() if k != "signature"}
        without_signature_after = {k:v for k,v in question.items() if k != "signature"}
        totals["content_changed" if without_signature_before != without_signature_after else "signature_only" if before != question else "unchanged"] += 1
        assert question["theory_block_id"] in blocks
        if question.get("choices"):
            ids = [c["id"] for c in question["choices"]]
            assert question["correct_answer"] in ids
            assert len({c["text"].strip().lower() for c in question["choices"]}) == len(question["choices"]), (path, question["id"])
        if question["type"] == "reorder":
            assert question.get("translation_ru") and 2 <= len(question["correct_answer"].split()) <= 12
    if unit["bank"] == "training":
        gen_spec = importlib.util.spec_from_file_location("gen", path.parents[3] / "scripts/generate-training-pack.py")
        gen = importlib.util.module_from_spec(gen_spec)
        gen_spec.loader.exec_module(gen)
        for question in questions:
            errors = gen.validate_question(question, unit["chapter"], blocks)
            assert not errors, (path, question["id"], errors)
            assert question["signature"] == gen.question_signature(question), (path, question["id"], "stale signature")
        assert len({q["signature"] for q in questions}) == len(questions), path
    else:
        final = gr.read(path.parent / "05-final.json")
        assert final["question_bank"]["questions"] == questions, (path, "source-final mismatch")
        previous_final = baseline(path.parent / "05-final.json")
        for document in (final, previous_final):
            document["question_bank"].pop("questions")
            document["meta"].pop("updated_at")
        assert final == previous_final, (path, "non-question final content changed")
    assert gr.status(live[unit["source"]], report)[0] == "awaiting_verification"
    if unit["bank"] == "chapter":
        source = path.parent / "05-final.json"
        destination = ROOT / f'internal/grammarbundle/es/chapters/{unit["chapter"]}.json'
    else:
        source = path
        destination = ROOT / "internal/grammartrainingpack/es" / path.relative_to(path.parents[2])
    assert source.read_bytes() == destination.read_bytes(), (path, "embedded mismatch")
    print(i, unit["language"], unit["bank"], len(questions), dict(Counter(q["decision"] for q in report["questions"])))

gen_spec = importlib.util.spec_from_file_location("gen", ROOT / "courses/spanish-grammar/scripts/generate-training-pack.py")
gen = importlib.util.module_from_spec(gen_spec)
gen_spec.loader.exec_module(gen)
assigned = {u["source"] for u in units}
old_signatures, new_signatures = Counter(), Counter()
for path in (ROOT / "courses/spanish-grammar/training_pack/chapters").rglob("*.json"):
    rel = path.relative_to(ROOT).as_posix()
    previous = baseline(path) if rel in assigned else gr.read(path)
    old_signatures.update(gen.question_signature(q) for q in previous["questions"])
    new_signatures.update(gen.question_signature(q) for q in gr.read(path)["questions"])
assert all(n <= max(1, old_signatures[s]) for s,n in new_signatures.items()), [(s,n,old_signatures[s]) for s,n in new_signatures.items() if n > max(1,old_signatures[s])]
print("es no new course-wide duplicate signatures")
print("Changes:", dict(totals))
print("PASS: all 480 IDs/contracts, choices, theory bindings, report fingerprints, signatures and source-final equality")
