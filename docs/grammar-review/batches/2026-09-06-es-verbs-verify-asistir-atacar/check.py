#!/usr/bin/env python3
"""Final evidence checks for the asistir/asociar/aspirar/atacar verification batch."""

import hashlib
import importlib.util
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
BATCH = Path(__file__).resolve().parent
BASE = Path("/tmp/grammar-es-verbs-verify-asistir-atacar-baseline")


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def fingerprint(path):
    digest = hashlib.sha256()
    digest.update(path.name.encode())
    digest.update(b"\0")
    digest.update(path.read_bytes())
    digest.update(b"\0")
    return digest.hexdigest()


def main():
    manifest = read(BATCH / "manifest.json")
    generator_path = ROOT / "courses/spanish-grammar/scripts/generate-verb-forms-training.py"
    spec = importlib.util.spec_from_file_location("verb_generator", generator_path)
    generator = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(generator)

    total = 0
    for lemma, item in manifest.items():
        source = ROOT / item["source"]
        embedded = ROOT / item["embedded"]
        report = read(ROOT / item["report"])
        baseline = read(BASE / item["source"])
        current = read(source)

        assert source.read_bytes() == embedded.read_bytes(), f"{lemma}: embedded drift"
        assert {key: value for key, value in current.items() if key != "cards"} == {
            key: value for key, value in baseline.items() if key != "cards"
        }, f"{lemma}: top-level contract changed"
        assert [
            (card["scope"], card["person"], card["number"])
            for card in current["cards"]
        ] == [
            (card["scope"], card["person"], card["number"])
            for card in baseline["cards"]
        ], f"{lemma}: identity/order changed"
        generator.validate_artifact(lemma, current["cards"])

        assert report["source"] == item["source"]
        assert report["source_sha256"] == item["source_sha256"] == fingerprint(source)
        assert report["context_sha256"] == item["context_sha256"]
        assert report["phase"] == "done"
        assert report["editor"] == item["editor"]
        assert report["verifier"] and report["verifier"] != report["editor"]
        assert report["verified_at"] and report["verification_note"].strip()
        assert len(report["questions"]) == len(current["cards"]) == 96
        assert all(question["verification"] == "ok" for question in report["questions"])
        assert all(
            check.get("command")
            and check.get("result") in {"pass", "known_baseline"}
            and check.get("evidence")
            for check in report["checks"]
        )
        print(f"{lemma}: PASS 96 independently verified cards")
        total += 96

    print(f"PASS: {total} cards; current fingerprints, contracts and reports")


if __name__ == "__main__":
    main()
