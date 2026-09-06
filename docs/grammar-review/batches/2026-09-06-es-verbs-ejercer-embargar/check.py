#!/usr/bin/env python3
"""Evidence checks for the ejercer/elegir/eliminar/embargar review batch."""

import csv
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
BASE = Path("/tmp/grammar-es-verbs-ejercer-embargar-baseline")
LEMMAS = {'ejercer': 'ce2d16335ee76178734b56bc',
 'elegir': '4004f5e871e21d49d1eb8b20',
 'eliminar': 'a112983f6c94449a53d94adb',
 'embargar': 'a9e6610cff985520ee4c2ea0'}
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
AUTHORITATIVE_FORMS = {('embargar', 'Indicativo', 'Presente'): ['embargo',
                                          'embargas',
                                          'embarga',
                                          'embargamos',
                                          'embargáis',
                                          'embargan'],
 ('embargar', 'Indicativo', 'Imperfecto'): ['embargaba',
                                            'embargabas',
                                            'embargaba',
                                            'embargábamos',
                                            'embargabais',
                                            'embargaban'],
 ('embargar', 'Indicativo', 'Pretérito'): ['embargué',
                                           'embargaste',
                                           'embargó',
                                           'embargamos',
                                           'embargasteis',
                                           'embargaron'],
 ('embargar', 'Indicativo', 'Futuro'): ['embargaré',
                                        'embargarás',
                                        'embargará',
                                        'embargaremos',
                                        'embargaréis',
                                        'embargarán'],
 ('embargar', 'Indicativo', 'Condicional'): ['embargaría',
                                             'embargarías',
                                             'embargaría',
                                             'embargaríamos',
                                             'embargaríais',
                                             'embargarían'],
 ('embargar', 'Indicativo', 'Pretérito perfecto'): ['he embargado',
                                                    'has embargado',
                                                    'ha embargado',
                                                    'hemos embargado',
                                                    'habéis embargado',
                                                    'han embargado'],
 ('embargar', 'Indicativo', 'Pluscuamperfecto'): ['había embargado',
                                                  'habías embargado',
                                                  'había embargado',
                                                  'habíamos embargado',
                                                  'habíais embargado',
                                                  'habían embargado'],
 ('embargar', 'Indicativo', 'Pretérito anterior'): ['hube embargado',
                                                    'hubiste embargado',
                                                    'hubo embargado',
                                                    'hubimos embargado',
                                                    'hubisteis embargado',
                                                    'hubieron embargado'],
 ('embargar', 'Indicativo', 'Futuro perfecto'): ['habré embargado',
                                                 'habrás embargado',
                                                 'habrá embargado',
                                                 'habremos embargado',
                                                 'habréis embargado',
                                                 'habrán embargado'],
 ('embargar', 'Indicativo', 'Condicional perfecto'): ['habría embargado',
                                                      'habrías embargado',
                                                      'habría embargado',
                                                      'habríamos embargado',
                                                      'habríais embargado',
                                                      'habrían embargado'],
 ('embargar', 'Subjuntivo', 'Presente'): ['embargue',
                                          'embargues',
                                          'embargue',
                                          'embarguemos',
                                          'embarguéis',
                                          'embarguen'],
 ('embargar', 'Subjuntivo', 'Imperfecto'): ['embargara',
                                            'embargaras',
                                            'embargara',
                                            'embargáramos',
                                            'embargarais',
                                            'embargaran'],
 ('embargar', 'Subjuntivo', 'Futuro'): ['embargare',
                                        'embargares',
                                        'embargare',
                                        'embargáremos',
                                        'embargareis',
                                        'embargaren'],
 ('embargar', 'Subjuntivo', 'Pretérito perfecto'): ['haya embargado',
                                                    'hayas embargado',
                                                    'haya embargado',
                                                    'hayamos embargado',
                                                    'hayáis embargado',
                                                    'hayan embargado'],
 ('embargar', 'Subjuntivo', 'Pluscuamperfecto'): ['hubiera embargado',
                                                  'hubieras embargado',
                                                  'hubiera embargado',
                                                  'hubiéramos embargado',
                                                  'hubierais embargado',
                                                  'hubieran embargado'],
 ('embargar', 'Subjuntivo', 'Futuro perfecto'): ['hubiere embargado',
                                                 'hubieres embargado',
                                                 'hubiere embargado',
                                                 'hubiéremos embargado',
                                                 'hubiereis embargado',
                                                 'hubieren embargado']}


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
    path = ROOT / "resources/verbs/jehle_verb_database.csv"
    with path.open(encoding="utf-8-sig", newline="") as handle:
        for row in csv.DictReader(handle):
            if row["infinitive"] in LEMMAS:
                found[(row["infinitive"], row["mood"], row["tense"])] = [
                    row[column].strip() for column in FORM_COLS
                ]
    found.update(AUTHORITATIVE_FORMS)
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

        # Person/number slots and their order are the stable identity here.
        old_ids = [(c["person"], c["number"]) for c in old["cards"]]
        new_ids = [(c["person"], c["number"]) for c in data["cards"]]
        assert new_ids == old_ids, f"{lemma}: person/number order changed"

        for scope_index, (scope, mood, tense) in enumerate(SCOPES):
            expected_forms = forms[(lemma, mood, tense)]
            scope_cards = data["cards"][scope_index * 6:(scope_index + 1) * 6]
            assert [c["surface_form"] for c in scope_cards] == expected_forms, (
                f"{lemma} {scope}: differs from authoritative forms"
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

        print(f"{lemma}: PASS 96 cards, exact Jehle/RAE forms, contracts and report fingerprint")

    checkpoints = [json.loads(line) for line in (
        ROOT / "docs/grammar-review/batches/2026-09-06-es-verbs-ejercer-embargar/checkpoints.jsonl"
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
