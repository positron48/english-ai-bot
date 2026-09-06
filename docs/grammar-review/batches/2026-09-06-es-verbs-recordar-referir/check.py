#!/usr/bin/env python3
"""Evidence checks for the recordar/recuperar/reducir/referir review batch."""

import csv
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
BASE = Path("/tmp/grammar-es-verbs-recordar-referir-baseline")
LEMMAS = {
    "recordar": "b7cca2e8368e8b2cb3e0b7ed",
    "recuperar": "fa283e5f149335317d797a67",
    "reducir": "98a646a484c52b3a46488571",
    "referir": "7e31848cd5d1a48aa674e426",
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

    # recuperar is regular and referir follows sentir. Both tables were checked
    # against their RAE DLE conjugation entries because Jehle lacks these lemmas.
    overrides = {
        "recuperar": [
            ["recupero", "recuperas", "recupera", "recuperamos", "recuperáis", "recuperan"],
            ["recuperaba", "recuperabas", "recuperaba", "recuperábamos", "recuperabais", "recuperaban"],
            ["recuperé", "recuperaste", "recuperó", "recuperamos", "recuperasteis", "recuperaron"],
            ["recuperaré", "recuperarás", "recuperará", "recuperaremos", "recuperaréis", "recuperarán"],
            ["recuperaría", "recuperarías", "recuperaría", "recuperaríamos", "recuperaríais", "recuperarían"],
            ["he recuperado", "has recuperado", "ha recuperado", "hemos recuperado", "habéis recuperado", "han recuperado"],
            ["había recuperado", "habías recuperado", "había recuperado", "habíamos recuperado", "habíais recuperado", "habían recuperado"],
            ["hube recuperado", "hubiste recuperado", "hubo recuperado", "hubimos recuperado", "hubisteis recuperado", "hubieron recuperado"],
            ["habré recuperado", "habrás recuperado", "habrá recuperado", "habremos recuperado", "habréis recuperado", "habrán recuperado"],
            ["habría recuperado", "habrías recuperado", "habría recuperado", "habríamos recuperado", "habríais recuperado", "habrían recuperado"],
            ["recupere", "recuperes", "recupere", "recuperemos", "recuperéis", "recuperen"],
            ["recuperara", "recuperaras", "recuperara", "recuperáramos", "recuperarais", "recuperaran"],
            ["recuperare", "recuperares", "recuperare", "recuperáremos", "recuperareis", "recuperaren"],
            ["haya recuperado", "hayas recuperado", "haya recuperado", "hayamos recuperado", "hayáis recuperado", "hayan recuperado"],
            ["hubiera recuperado", "hubieras recuperado", "hubiera recuperado", "hubiéramos recuperado", "hubierais recuperado", "hubieran recuperado"],
            ["hubiere recuperado", "hubieres recuperado", "hubiere recuperado", "hubiéremos recuperado", "hubiereis recuperado", "hubieren recuperado"],
        ],
        "referir": [
            ["refiero", "refieres", "refiere", "referimos", "referís", "refieren"],
            ["refería", "referías", "refería", "referíamos", "referíais", "referían"],
            ["referí", "referiste", "refirió", "referimos", "referisteis", "refirieron"],
            ["referiré", "referirás", "referirá", "referiremos", "referiréis", "referirán"],
            ["referiría", "referirías", "referiría", "referiríamos", "referiríais", "referirían"],
            ["he referido", "has referido", "ha referido", "hemos referido", "habéis referido", "han referido"],
            ["había referido", "habías referido", "había referido", "habíamos referido", "habíais referido", "habían referido"],
            ["hube referido", "hubiste referido", "hubo referido", "hubimos referido", "hubisteis referido", "hubieron referido"],
            ["habré referido", "habrás referido", "habrá referido", "habremos referido", "habréis referido", "habrán referido"],
            ["habría referido", "habrías referido", "habría referido", "habríamos referido", "habríais referido", "habrían referido"],
            ["refiera", "refieras", "refiera", "refiramos", "refiráis", "refieran"],
            ["refiriera", "refirieras", "refiriera", "refiriéramos", "refirierais", "refirieran"],
            ["refiriere", "refirieres", "refiriere", "refiriéremos", "refiriereis", "refirieren"],
            ["haya referido", "hayas referido", "haya referido", "hayamos referido", "hayáis referido", "hayan referido"],
            ["hubiera referido", "hubieras referido", "hubiera referido", "hubiéramos referido", "hubierais referido", "hubieran referido"],
            ["hubiere referido", "hubieres referido", "hubiere referido", "hubiéremos referido", "hubiereis referido", "hubieren referido"],
        ],
    }
    for lemma, rows in overrides.items():
        for (_, mood, tense), row in zip(SCOPES, rows):
            found[(lemma, mood, tense)] = row
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
        assert {k: v for k, v in data.items() if k != "cards"} == {
            k: v for k, v in old.items() if k != "cards"
        }, f"{lemma}: top-level contract changed"
        assert len(data["cards"]) == len(old["cards"]) == 96
        assert [(c["scope"], c["person"], c["number"]) for c in data["cards"]] == [
            (c["scope"], c["person"], c["number"]) for c in old["cards"]
        ], f"{lemma}: scope/person/number order changed"

        for scope, mood, tense in SCOPES:
            expected_forms = forms[(lemma, mood, tense)]
            scope_cards = [c for c in data["cards"] if c["scope"] == scope]
            assert [c["surface_form"] for c in scope_cards] == expected_forms, f"{lemma} {scope}: differs from reference"
            for slot_index, card in enumerate(scope_cards):
                assert (card["person"], card["number"]) == SLOTS[slot_index]
                assert card["mood"] == scope.split(".")[2]
                assert card["tense"] == scope.split(".")[1]
                assert card["question_es_with_blank"].count("_") == 1
                assert card["question_es_with_blank"].endswith(f"({lemma})")
                assert any("а" <= ch.lower() <= "я" or ch.lower() == "ё" for ch in card["translation_ru_full"])
                if card["question_es_with_blank"].startswith("De haberlo sabido"):
                    assert card["translation_ru_full"].startswith("Если бы ")
                    assert "Зная это заранее" not in card["translation_ru_full"]
                assert len(card["options"]) == len(set(card["options"])) == 4
                assert card["options"].count(card["surface_form"]) == 1
                assert set(card["options"]).issubset(set(expected_forms))
                signature = (card["question_es_with_blank"].casefold(), card["surface_form"].casefold())
                assert signature not in all_signatures, f"duplicate prompt+answer: {signature}"
                all_signatures.add(signature)
                total += 1

        report = read(ROOT / f"docs/grammar-review/reports/{report_id}.json")
        assert report["source"] == source.relative_to(ROOT).as_posix()
        assert report["source_sha256"] == fingerprint([source])
        assert report["context_sha256"] == context_sha
        assert report["phase"] == "done"
        assert report["editor"] and report["reviewed_at"]
        assert report["verifier"] and report["verified_at"]
        assert report["editor"] != report["verifier"]
        assert report["verification_note"].strip() and report["checks"]
        assert all(isinstance(check, dict) and check.get("command") and check.get("result") in {"pass", "known_baseline"} and check.get("evidence") for check in report["checks"])
        assert len(report["questions"]) == 96
        assert all(q["decision"] in {"fixed", "ok"} and
                   (q["decision"] == "ok" or q["note"].strip()) and
                   q["verification"] == "ok" for q in report["questions"])
        print(f"{lemma}: PASS 96 cards, exact reference forms, contracts and report fingerprint")

    checkpoints = [json.loads(line) for line in (
        ROOT / "docs/grammar-review/batches/2026-09-06-es-verbs-recordar-referir/checkpoints.jsonl"
    ).read_text(encoding="utf-8").splitlines()]
    assert len(checkpoints) == 24
    for lemma in LEMMAS:
        rows = [row for row in checkpoints if row["lemma"] == lemma]
        ranges = [row["range"] for row in rows]
        assert len(ranges) == 6 and ranges[0][0] == 0 and ranges[-1][1] == 96
        assert all(a[1] == b[0] and 10 <= a[1] - a[0] <= 20 for a, b in zip(ranges, ranges[1:]))
        assert 10 <= ranges[-1][1] - ranges[-1][0] <= 20
        assert rows[-1]["remaining_cards"] == 0
    print(f"PASS: {total} cards; 24 checkpoints; no duplicate prompt+answer signatures")


if __name__ == "__main__":
    main()
