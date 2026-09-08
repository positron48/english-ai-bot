#!/usr/bin/env python3
"""Read-only coverage/consistency audit. Does not certify linguistic quality."""
import argparse
import collections
import csv
import json
import pathlib
import re

ROOT = pathlib.Path(__file__).resolve().parents[2]
ALIASES = {'preterito_indefinido': 'pretérito', 'preterito_imperfecto': 'imperfecto',
           'futuro_simple': 'futuro', 'condicional_simple': 'condicional',
           'preterito_perfecto_compuesto': 'pretérito perfecto', 'preterito_perfecto': 'pretérito perfecto',
           'preterito_pluscuamperfecto': 'pluscuamperfecto', 'preterito_anterior': 'pretérito anterior',
           'futuro_perfecto': 'futuro perfecto', 'condicional_perfecto': 'condicional perfecto'}


def canonical(value):
    value = value.strip().lower()
    return ALIASES.get(value, value)


def audit():
    reference = {}
    for path in sorted((ROOT / 'resources/verbs').glob('*.csv')):
        with path.open(newline='') as file:
            for row in csv.DictReader(file):
                if 'infinitive' not in row:
                    continue
                key = (row['infinitive'].lower(), row['mood'].lower(), canonical(row['tense']))
                reference[key] = [row.get('form_' + slot, '').lower() for slot in ['1s', '2s', '3s', '1p', '2p', '3p']]
    core_src = (ROOT / 'internal/verbtraining/curriculum.go').read_text()
    core = set(re.findall(r'"([a-záéíóúñ]+)"', core_src.split('var CoreLemmas = []string{')[1].split('}')[0]))
    results = []
    totals = collections.Counter()
    for path in sorted((ROOT / 'courses/spanish-grammar/training_pack/verb_forms/lemmas').glob('*.json')):
        artifact = json.loads(path.read_text())
        lemma = artifact['lemma']
        groups = collections.defaultdict(list)
        for card in artifact['cards']:
            groups[(card['mood'], canonical(card['tense']))].append(card)
        for (mood, tense), cards in groups.items():
            issues = collections.Counter()
            for card in cards:
                totals['cards'] += 1
                slot = int(card['person']) - 1 + (3 if card['number'] == 'plural' else 0)
                ref = reference.get((lemma, mood, tense))
                if ref and ref[slot] and card['surface_form'].lower() != ref[slot]:
                    issues['dictionary_mismatch'] += 1
                if not re.search(r'_+', card['question_es_with_blank']):
                    issues['missing_blank'] += 1
                if not card['translation_ru_full'].strip():
                    issues['missing_translation'] += 1
                if len(set(card['options'])) != len(card['options']):
                    issues['duplicate_options'] += 1
                if card['surface_form'] not in card['options']:
                    issues['answer_missing_from_options'] += 1
                if ref and any(option.lower() not in ref for option in card['options']):
                    issues['options_outside_paradigm'] += 1
            results.append({'lemma': lemma, 'core': lemma in core, 'mood': mood, 'tense': tense,
                            'slots': len({(c['person'], c['number']) for c in cards}), 'issues': dict(issues)})
            totals.update(issues)
    return {'note': 'Structural checks only; dictionary differences require review, not automatic replacement.',
            'lemmas': len({r['lemma'] for r in results}), 'totals': dict(totals), 'coverage': results}


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--output', type=pathlib.Path, required=True)
    args = parser.parse_args()
    result = audit()
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps({'lemmas': result['lemmas'], **result['totals']}, ensure_ascii=False))
