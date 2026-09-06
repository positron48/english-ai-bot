#!/usr/bin/env python3
"""Evidence checks for the luchar/mandar/manejar/manifestar review batch."""

import csv
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
BASE = Path("/tmp/grammar-es-verbs-luchar-manifestar-baseline")
LEMMAS = {
    "luchar": "f4f35b9830bb5a6c384beb2d",
    "mandar": "20cb3b7b6ffbcd7817fafe19",
    "manejar": "5e62f129bf55fc12befb2bcd",
    "manifestar": "ef7d6d9462d822e73cba6436",
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
SLOTS = [("1", "singular"), ("2", "singular"), ("3", "singular"),
         ("1", "plural"), ("2", "plural"), ("3", "plural")]
SUBJECTS = ["yo", "tú", "él", "nosotros", "vosotros", "ellos"]
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


def jehle_forms():
    found = {}
    paths = [
        ROOT / "resources/verbs/jehle_verb_database.csv",
        ROOT / "resources/verbs/jehle_supplement_aux_haber.csv",
    ]
    for path in paths:
        with path.open(encoding="utf-8-sig", newline="") as handle:
            for row in csv.DictReader(handle):
                if row["infinitive"] in LEMMAS:
                    found[(row["infinitive"], row["mood"], row["tense"])] = [
                        row[column].strip() for column in FORM_COLS
                    ]
    rae_manifestar = [
        ["manifiesto", "manifiestas", "manifiesta", "manifestamos", "manifestáis", "manifiestan"],
        ["manifestaba", "manifestabas", "manifestaba", "manifestábamos", "manifestabais", "manifestaban"],
        ["manifesté", "manifestaste", "manifestó", "manifestamos", "manifestasteis", "manifestaron"],
        ["manifestaré", "manifestarás", "manifestará", "manifestaremos", "manifestaréis", "manifestarán"],
        ["manifestaría", "manifestarías", "manifestaría", "manifestaríamos", "manifestaríais", "manifestarían"],
        ["he manifestado", "has manifestado", "ha manifestado", "hemos manifestado", "habéis manifestado", "han manifestado"],
        ["había manifestado", "habías manifestado", "había manifestado", "habíamos manifestado", "habíais manifestado", "habían manifestado"],
        ["hube manifestado", "hubiste manifestado", "hubo manifestado", "hubimos manifestado", "hubisteis manifestado", "hubieron manifestado"],
        ["habré manifestado", "habrás manifestado", "habrá manifestado", "habremos manifestado", "habréis manifestado", "habrán manifestado"],
        ["habría manifestado", "habrías manifestado", "habría manifestado", "habríamos manifestado", "habríais manifestado", "habrían manifestado"],
        ["manifieste", "manifiestes", "manifieste", "manifestemos", "manifestéis", "manifiesten"],
        ["manifestara", "manifestaras", "manifestara", "manifestáramos", "manifestarais", "manifestaran"],
        ["manifestare", "manifestares", "manifestare", "manifestáremos", "manifestareis", "manifestaren"],
        ["haya manifestado", "hayas manifestado", "haya manifestado", "hayamos manifestado", "hayáis manifestado", "hayan manifestado"],
        ["hubiera manifestado", "hubieras manifestado", "hubiera manifestado", "hubiéramos manifestado", "hubierais manifestado", "hubieran manifestado"],
        ["hubiere manifestado", "hubieres manifestado", "hubiere manifestado", "hubiéremos manifestado", "hubiereis manifestado", "hubieren manifestado"],
    ]
    for (_, mood, tense), values in zip(SCOPES, rae_manifestar):
        found[("manifestar", mood, tense)] = values
    return found


def main():
    forms = jehle_forms()
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
        assert {k: v for k, v in data.items() if k != "cards"} == {
            k: v for k, v in old.items() if k != "cards"
        }, f"{lemma}: top-level contract changed"
        assert len(data["cards"]) == len(old["cards"]) == 96

        old_ids = [(c["person"], c["number"]) for c in old["cards"]]
        new_ids = [(c["person"], c["number"]) for c in data["cards"]]
        assert new_ids == old_ids, f"{lemma}: person/number order changed"

        for scope_index, (scope, mood, tense) in enumerate(SCOPES):
            expected_forms = forms[(lemma, mood, tense)]
            scope_cards = data["cards"][scope_index * 6:(scope_index + 1) * 6]
            assert [c["surface_form"] for c in scope_cards] == expected_forms, (
                f"{lemma} {scope}: differs from Jehle/RAE"
            )
            for slot_index, card in enumerate(scope_cards):
                person, number = SLOTS[slot_index]
                assert card["scope"] == scope
                assert card["mood"] == scope.split(".")[2]
                assert card["tense"] == scope.split(".")[1]
                assert (card["person"], card["number"]) == (person, number)
                assert card["question_es_with_blank"].count("_") == 1
                assert card["question_es_with_blank"].endswith(f"({lemma})")
                assert SUBJECTS[slot_index] in card["question_es_with_blank"].lower().split()
                assert any("а" <= ch.lower() <= "я" or ch.lower() == "ё"
                           for ch in card["translation_ru_full"])
                assert len(card["options"]) == len(set(card["options"])) == 4
                assert card["options"].count(card["surface_form"]) == 1
                assert set(card["options"]).issubset(set(expected_forms))
                signature = (card["question_es_with_blank"].casefold(), card["surface_form"].casefold())
                assert signature not in all_signatures, f"duplicate prompt+answer: {signature}"
                all_signatures.add(signature)
                total += 1

        report_path = ROOT / f"docs/grammar-review/reports/{report_id}.json"
        report = read(report_path)
        assert report["source"] == source.relative_to(ROOT).as_posix()
        assert report["source_sha256"] == fingerprint([source])
        assert report["context_sha256"] == context_sha
        assert report["phase"] == "awaiting_verification"
        assert report["editor"] and report["reviewed_at"]
        assert not report["verifier"] and not report["verified_at"]
        assert len(report["questions"]) == 96
        assert all(q["decision"] == "fixed" and q["note"].strip() and
                   q["verification"] == "pending" for q in report["questions"])

        print(f"{lemma}: PASS 96 cards, exact reference forms, contracts and report fingerprint")

    checkpoints = [json.loads(line) for line in (
        ROOT / "docs/grammar-review/batches/2026-09-06-es-verbs-luchar-manifestar/checkpoints.jsonl"
    ).read_text(encoding="utf-8").splitlines()]
    assert len(checkpoints) == 24
    for lemma in LEMMAS:
        rows = [row for row in checkpoints if row["lemma"] == lemma]
        assert [row["range"] for row in rows] == [
            [0, 18], [18, 36], [36, 54], [54, 72], [72, 84], [84, 96]
        ]
        assert rows[-1]["remaining_cards"] == 0
    print(f"PASS: {total} cards; 24 checkpoints; no duplicate prompt+answer signatures")


if __name__ == "__main__":
    main()
