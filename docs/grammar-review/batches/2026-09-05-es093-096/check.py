#!/usr/bin/env python3
"""Read-only validation of the ES 093-096 editorial batch against its baseline."""
import importlib.util
import json
import pathlib
from collections import Counter

ROOT = pathlib.Path(__file__).resolve().parents[4]
BASE_DIR = pathlib.Path('/tmp/grammar-es9396-baseline')

spec = importlib.util.spec_from_file_location('grammar_review', ROOT / 'scripts/grammar-review.py')
gr = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gr)
units = [
    u for u in gr.inventory()
    if u['language'] == 'es'
    and u['order'].startswith(('093.', '094.', '095.', '096.'))
    and u['bank'] == 'chapter'
]
assert len(units) == 4 and sum(u['count'] for u in units) == 282

def baseline(path):
    stored = BASE_DIR / path.relative_to(ROOT)
    assert stored.exists(), stored
    return json.loads(stored.read_text())

live = {u['source']: u for u in gr.inventory()}
totals = Counter()
for i, unit in enumerate(units):
    path = ROOT / unit['source']
    old = baseline(path)
    now = gr.read(path)
    report = gr.read(ROOT / unit['report'])
    assert report['source'] == unit['source']
    assert report['source_sha256'] == unit['source_sha256'], (path, 'stale source fingerprint')
    assert report['context_sha256'] == unit['context_sha256'], (path, 'stale context fingerprint')
    assert {k: v for k, v in old.items() if k != 'questions'} == {k: v for k, v in now.items() if k != 'questions'}, path
    old_questions = old['questions']
    questions = now['questions']
    assert len(questions) == len(old_questions) == unit['count']

    def identity(question):
        return (
            question['id'], question['type'], question.get('difficulty'),
            question.get('theory_block_id'), question.get('chapter_id'),
            question.get('concept_id'),
            [choice['id'] for choice in question.get('choices', [])],
        )

    assert list(map(identity, old_questions)) == list(map(identity, questions)), path
    assert len({question['id'] for question in questions}) == len(questions)
    blocks = {
        block['id']
        for block in gr.read(path.parent / '01-outline.json')['chapter_outline']['theory_blocks']
    }
    for before, question, row in zip(old_questions, questions, report['questions']):
        assert row['id'] == question['id']
        assert (before != question) == (row['decision'] == 'fixed'), (path, question['id'], row['decision'])
        assert row['decision'] in ('ok', 'fixed')
        assert row['note'] and row['verification'] == 'pending'
        totals['content_changed' if before != question else 'unchanged'] += 1
        if question['theory_block_id'] not in blocks:
            assert before['theory_block_id'] == question['theory_block_id'], (path, question['id'], 'new theory binding defect')
        if question.get('choices'):
            choice_ids = [choice['id'] for choice in question['choices']]
            assert question['correct_answer'] in choice_ids
            assert len({choice['text'].strip().lower() for choice in question['choices']}) == len(question['choices']), (path, question['id'])
        if question['type'] == 'reorder':
            assert question.get('translation_ru')
            assert 2 <= len(question['correct_answer'].split()) <= 12

    validation = path.parent / '05-validation.json'
    baseline_validation = BASE_DIR / validation.relative_to(ROOT)
    assert validation.read_bytes() == baseline_validation.read_bytes(), (path, 'validation artifact changed')

    final = gr.read(path.parent / '05-final.json')
    assert final['question_bank']['questions'] == questions, (path, 'source-final mismatch')
    previous_final = baseline(path.parent / '05-final.json')
    for document in (final, previous_final):
        document['question_bank'].pop('questions')
        document['meta'].pop('updated_at')
    assert final == previous_final, (path, 'non-question final content changed')
    assert gr.status(live[unit['source']], report)[0] == 'awaiting_verification'
    destination = ROOT / f'internal/grammarbundle/es/chapters/{unit["chapter"]}.json'
    assert (path.parent / '05-final.json').read_bytes() == destination.read_bytes(), (path, 'embedded mismatch')
    print(i, unit['language'], unit['bank'], len(questions), dict(Counter(q['decision'] for q in report['questions'])))

assert totals == Counter({'unchanged': 233, 'content_changed': 49}), totals
print('Changes:', dict(totals))
print('PASS: all 282 IDs/contracts, choices, theory bindings, report fingerprints and source-final equality')
