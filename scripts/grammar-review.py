#!/usr/bin/env python3
"""Read-only inventory and evidence checks for the full grammar editorial review."""

import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parents[1]
REPORTS = Path("docs/grammar-review/reports")


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def fingerprint(paths):
    digest = hashlib.sha256()
    for path in sorted(paths):
        digest.update(path.name.encode())
        digest.update(b"\0")
        digest.update(path.read_bytes())
        digest.update(b"\0")
    return digest.hexdigest()


def inventory(root=ROOT):
    units = []
    for language, course in (("es", "spanish-grammar"), ("en", "english-grammar")):
        base = root / "courses" / course
        contexts = {}
        for path in sorted((base / "chapters").glob("*/03-questions.json")):
            outline = path.parent / "01-outline.json"
            context = [outline, *sorted((path.parent / "02-theory-blocks").glob("*.json"))]
            data = read(outline)["chapter_outline"]
            contexts[data["chapter_id"]] = (context, path.parent.name, data.get("level", ""))
            units.append(make_unit(root, path, language, "chapter", data["chapter_id"],
                                   [str(q["id"]) for q in read(path)["questions"]], context,
                                   path.parent.name, data.get("level", "")))
        pack = base / "training_pack"
        index = read(pack / "index.json")
        refs = list(index.get("blocks", {}).values())
        if not refs:
            for value in index.get("chapters", {}).values():
                refs.extend(value if isinstance(value, list) else [value])
        indexed = {(pack / "chapters" / p).resolve() for p in refs}
        for path in indexed:
            if not path.is_relative_to((pack / "chapters").resolve()) or not path.is_file():
                raise ValueError(f"Invalid training index reference: {path}")
        for path in sorted((pack / "chapters").rglob("*.json")):
            data = read(path)
            chapter = data["chapter_id"]
            context, order, level = contexts[chapter]
            units.append(make_unit(root, path, language, "training", chapter,
                                   [str(q["id"]) for q in data["questions"]],
                                   [*context, pack / "index.json"], order, level,
                                   path.resolve() in indexed))
        verbs = pack / "verb_forms"
        if (verbs / "index.json").is_file():
            indexed = {(verbs / p).resolve() for p in read(verbs / "index.json")["lemmas"].values()}
            for path in indexed:
                if not path.is_relative_to((verbs / "lemmas").resolve()) or not path.is_file():
                    raise ValueError(f"Invalid verb index reference: {path}")
            for path in sorted((verbs / "lemmas").glob("*.json")):
                data = read(path)
                ids = [f'{c["scope"]}:{c["person"]}:{c["number"]}' for c in data["cards"]]
                context = [verbs / "index.json"]
                if (verbs / "unlock-gates.json").is_file():
                    context.append(verbs / "unlock-gates.json")
                units.append(make_unit(root, path, language, "verbs", data["lemma"], ids,
                                       context, data["lemma"], "", path.resolve() in indexed))
    return sorted(units, key=lambda u: (u["bank"] == "verbs", u["language"] != "es",
                                       u["order"], u["bank"] != "chapter", u["source"]))


def make_unit(root, path, language, bank, chapter, ids, context, order, level, indexed=True):
    if len(ids) != len(set(ids)):
        raise ValueError(f"Duplicate question IDs: {path}")
    source = path.relative_to(root).as_posix()
    return dict(source=source, language=language, bank=bank, chapter=chapter, level=level,
                order=order, indexed=indexed, question_ids=ids, count=len(ids),
                source_sha256=fingerprint([path]), context_sha256=fingerprint(context),
                report=(REPORTS / (hashlib.sha256(source.encode()).hexdigest()[:24] + ".json")).as_posix())


def template(unit):
    return dict(version=1, source=unit["source"], source_sha256=unit["source_sha256"],
                context_sha256=unit["context_sha256"], editor="", verifier="",
                phase="in_review", reviewed_at="", verified_at="",
                questions=[dict(id=qid, decision="pending", note="", verification="pending")
                           for qid in unit["question_ids"]],
                checks=[], verification_note="")


def status(unit, report):
    if report is None:
        return "pending", "No report for this review cycle"
    if not isinstance(report, dict):
        return "invalid", "Report must be a JSON object"
    if report.get("source") != unit["source"] or report.get("version") != 1:
        return "invalid", "Report source/version mismatch"
    if any(report.get(k) != unit[k] for k in ("source_sha256", "context_sha256")):
        return "stale", "Source or theory changed; review must be reconfirmed"
    rows = report.get("questions", [])
    if not isinstance(rows, list) or any(not isinstance(row, dict) for row in rows):
        return "invalid", "Questions must be a list of review records"
    ids = [row.get("id") for row in rows]
    if len(ids) != len(set(ids)) or set(ids) != set(unit["question_ids"]):
        return "invalid", "Question coverage differs from the current source"
    if any(row.get("decision") not in ("pending", "ok", "fixed", "blocked") or
           row.get("verification") not in ("pending", "ok", "needs_fix") for row in rows):
        return "invalid", "Unknown question decision or verification"
    if any(row["decision"] in ("fixed", "blocked") and not row.get("note", "").strip() for row in rows):
        return "invalid", "Changed or blocked questions need an explanation"
    if any(row["decision"] == "blocked" or row["verification"] == "needs_fix" for row in rows):
        return "needs_fix", "Unresolved editorial findings"
    if any(row["decision"] == "pending" for row in rows):
        return "in_review", "Editorial review is incomplete"
    if not report.get("editor", "").strip() or not report.get("reviewed_at", "").strip():
        return "invalid", "Missing editor or review date"
    if (any(row["verification"] != "ok" for row in rows) or
            (not rows and report.get("phase") == "awaiting_verification")):
        return "awaiting_verification", "Independent review is incomplete"
    if (not report.get("verifier", "").strip() or report["editor"].strip() == report["verifier"].strip()
            or not report.get("verified_at", "").strip() or not report.get("verification_note", "").strip()):
        return "invalid", "Independent verifier, date and conclusion are required"
    checks = report.get("checks", [])
    if not isinstance(checks, list) or not checks or any(not isinstance(c, dict) or
                         c.get("result") not in ("pass", "known_baseline") or
                         not c.get("command", "").strip() or not c.get("evidence", "").strip() for c in checks):
        return "invalid", "Validation commands and evidence are required"
    if report.get("phase") != "done":
        return "awaiting_verification", "Report is not signed off"
    return "done", "Full recorded coverage; matching source and theory fingerprints"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("status", "next", "inventory", "template", "check"))
    parser.add_argument("--language", choices=("es", "en"))
    parser.add_argument("--bank", choices=("chapter", "training", "verbs"))
    parser.add_argument("--limit", type=int, default=3)
    parser.add_argument("--source", help="Exact repository-relative source from the inventory")
    args = parser.parse_args()
    units = [u for u in inventory() if (not args.language or u["language"] == args.language)
             and (not args.bank or u["bank"] == args.bank)
             and (not args.source or u["source"] == args.source)]
    if not units:
        parser.error("No matching sources")
    if args.limit < 1:
        parser.error("--limit must be positive")
    if args.command == "template":
        if not args.source or len(units) != 1:
            parser.error("template requires --source for exactly one file")
        print(json.dumps(template(units[0]), ensure_ascii=False, indent=2))
        return 0
    states = []
    for unit in units:
        path = ROOT / unit["report"]
        state, reason = status(unit, read(path) if path.is_file() else None)
        states.append(dict(**unit, status=state, reason=reason))
    if args.command in ("inventory", "next"):
        if args.command == "next":
            priority = {"invalid": 0, "needs_fix": 1, "stale": 2, "in_review": 3,
                        "awaiting_verification": 4, "pending": 5}
            states = sorted((u for u in states if u["status"] != "done"),
                            key=lambda u: priority[u["status"]])[:args.limit]
        print(json.dumps(states, ensure_ascii=False, indent=2))
    else:
        totals = Counter()
        for u in states:
            totals[(u["language"], u["bank"], u["status"])] += u["count"]
        for key, count in sorted(totals.items()):
            print(" / ".join(key), ":", count, "questions")
        print(f'Files: {len(states)}; questions: {sum(u["count"] for u in states)}; '
              f'done files: {sum(u["status"] == "done" for u in states)}')
        for u in states:
            if u["status"] in ("stale", "invalid", "needs_fix"):
                print(u["status"], u["source"], u["reason"])
        if args.command == "check" and any(u["status"] != "done" for u in states):
            return 1
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (ValueError, KeyError, OSError, TypeError) as error:
        print(f"Grammar review error: {error}", file=sys.stderr)
        sys.exit(2)
