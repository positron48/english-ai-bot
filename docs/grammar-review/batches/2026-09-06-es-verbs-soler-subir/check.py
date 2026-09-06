#!/usr/bin/env python3
"""Evidence checks for the soler/sonar/sorprender/subir review batch."""

import csv
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
BASE = Path("/tmp/grammar-es-verbs-soler-subir-baseline")
LEMMAS = {
    "soler": "1158494c987616cb829c6349",
    "sonar": "c93d4e52d86da061c77be5c0",
    "sorprender": "eaa2135b95b3478291c04e5c",
    "subir": "2f35ea3fd6c4d6b4bebb7652",
}
SCOPES = [
    ("es.presente.indicativo", "Indicativo", "Presente"),
    ("es.preterito_imperfecto.indicativo", "Indicativo", "Imperfecto"),
    ("es.preterito_indefinido.indicativo", "Indicativo", "Pretérito"),
    ("es.futuro_simple.indicativo", "Indicativo", "Futuro"),
    ("es.condicional_simple.indicativo", "Indicativo", "Condicional"),
    ("es.preterito_perfecto_compuesto.indicativo", "Indicativo", "Pretérito perfecto"),
    ("es.preterito_pluscuamperfecto.indicativo", "Indicativo", "Pluscuamperfecto"),
    ("es.preterito_anterior.indicativo", "Indicativo", "Pretérito anterior"),
    ("es.futuro_perfecto.indicativo", "Indicativo", "Futuro perfecto"),
    ("es.condicional_perfecto.indicativo", "Indicativo", "Condicional perfecto"),
    ("es.presente.subjuntivo", "Subjuntivo", "Presente"),
    ("es.preterito_imperfecto.subjuntivo", "Subjuntivo", "Imperfecto"),
    ("es.futuro_simple.subjuntivo", "Subjuntivo", "Futuro"),
    ("es.preterito_perfecto.subjuntivo", "Subjuntivo", "Pretérito perfecto"),
    ("es.preterito_pluscuamperfecto.subjuntivo", "Subjuntivo", "Pluscuamperfecto"),
    ("es.futuro_perfecto.subjuntivo", "Subjuntivo", "Futuro perfecto"),
]
SLOTS = [
    ("1", "singular"), ("2", "singular"), ("3", "singular"),
    ("1", "plural"), ("2", "plural"), ("3", "plural"),
]
FORM_COLS = ["form_1s", "form_2s", "form_3s", "form_1p", "form_2p", "form_3p"]


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


def reference_forms():
    found = {}
    for path in [
        ROOT / "resources/verbs/jehle_verb_database.csv",
        ROOT / "resources/verbs/jehle_supplement_aux_haber.csv",
    ]:
        with path.open(encoding="utf-8-sig", newline="") as handle:
            for row in csv.DictReader(handle):
                if row["infinitive"] in LEMMAS:
                    found[(row["infinitive"], row["mood"], row["tense"])] = [
                        row[column].strip() for column in FORM_COLS
                    ]
    return found


def main():
    forms = reference_forms()
    context = [
        ROOT / "courses/spanish-grammar/training_pack/verb_forms/index.json",
        ROOT / "courses/spanish-grammar/training_pack/verb_forms/unlock-gates.json",
    ]
    context_sha = fingerprint(context)
    all_signatures = set()
    total = 0
    for lemma, report_id in LEMMAS.items():
        source = ROOT / f"courses/spanish-grammar/training_pack/verb_forms/lemmas/{lemma}.json"
        embedded = ROOT / f"internal/grammartrainingpack/es/verb_forms/lemmas/{lemma}.json"
        baseline = BASE / f"courses/spanish-grammar/training_pack/verb_forms/lemmas/{lemma}.json"
        data, old = read(source), read(baseline)
        assert source.read_bytes() == embedded.read_bytes(), f"{lemma}: embedded drift"
        assert {k: v for k, v in data.items() if k != "cards"} == {k: v for k, v in old.items() if k != "cards"}, f"{lemma}: top-level contract changed"
        assert len(data["cards"]) == len(old["cards"]) == 96
        assert [(c["scope"], c["person"], c["number"]) for c in data["cards"]] == [(c["scope"], c["person"], c["number"]) for c in old["cards"]], f"{lemma}: scope/person/number order changed"
        for scope, mood, tense in SCOPES:
            expected = forms[(lemma, mood, tense)]
            cards = [card for card in data["cards"] if card["scope"] == scope]
            assert [card["surface_form"] for card in cards] == expected, f"{lemma} {scope}: differs from Jehle"
            for index, card in enumerate(cards):
                assert (card["person"], card["number"]) == SLOTS[index]
                assert card["mood"] == scope.split(".")[2] and card["tense"] == scope.split(".")[1]
                assert card["question_es_with_blank"].count("_") == 1
                assert card["question_es_with_blank"].endswith(f"({lemma})")
                assert any("а" <= char.lower() <= "я" or char.lower() == "ё" for char in card["translation_ru_full"])
                if card["question_es_with_blank"].startswith("De haberlo sabido"):
                    assert card["translation_ru_full"].startswith("Если бы ")
                    assert "Зная это заранее" not in card["translation_ru_full"]
                assert len(card["options"]) == len(set(card["options"])) == 4
                assert card["options"].count(card["surface_form"]) == 1
                assert set(card["options"]).issubset(set(expected))
                signature = (card["question_es_with_blank"].casefold(), card["surface_form"].casefold())
                assert signature not in all_signatures, f"duplicate prompt+answer: {signature}"
                all_signatures.add(signature)
                total += 1
        report = read(ROOT / f"docs/grammar-review/reports/{report_id}.json")
        assert report["source"] == source.relative_to(ROOT).as_posix()
        assert report["source_sha256"] == fingerprint([source])
        assert report["context_sha256"] == context_sha
        assert report["phase"] == "done" and report["editor"] and report["reviewed_at"]
        assert report["verifier"] and report["verified_at"] and report["editor"] != report["verifier"]
        assert report["verification_note"].strip() and report["checks"]
        assert all(
            isinstance(check, dict)
            and check.get("command")
            and check.get("result") in {"pass", "known_baseline"}
            and check.get("evidence")
            for check in report["checks"]
        )
        assert len(report["questions"]) == 96
        assert all(
            question["decision"] in {"fixed", "ok"}
            and (question["decision"] == "ok" or question["note"].strip())
            and question["verification"] == "ok"
            for question in report["questions"]
        )
        print(f"{lemma}: PASS 96 cards, exact Jehle forms, contracts and report fingerprint")
    checkpoints = [
        json.loads(line)
        for line in (ROOT / "docs/grammar-review/batches/2026-09-06-es-verbs-soler-subir/checkpoints.jsonl").read_text().splitlines()
    ]
    assert len(checkpoints) == 24
    for lemma in LEMMAS:
        rows = [row for row in checkpoints if row["lemma"] == lemma]
        ranges = [row["range"] for row in rows]
        assert len(ranges) == 6 and ranges[0][0] == 0 and ranges[-1][1] == 96
        assert all(a[1] == b[0] and 10 <= a[1] - a[0] <= 20 for a, b in zip(ranges, ranges[1:]))
        assert 10 <= ranges[-1][1] - ranges[-1][0] <= 20 and rows[-1]["remaining_cards"] == 0
    print(f"PASS: {total} cards; 24 checkpoints; no duplicate prompt+answer signatures")


if __name__ == "__main__":
    main()
