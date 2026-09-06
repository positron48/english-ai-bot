#!/usr/bin/env python3
"""Read-only validation of the ES 105-108 editorial batch against its baseline."""
import importlib.util
import json
import pathlib
from collections import Counter

ROOT = pathlib.Path(__file__).resolve().parents[4]
BASE_DIR = pathlib.Path('/tmp/grammar-es105108-baseline')

spec = importlib.util.spec_from_file_location('grammar_review', ROOT / 'scripts/grammar-review.py')
gr = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gr)
units = [
    u for u in gr.inventory()
    if u['language'] == 'es'
    and u['order'].startswith(('105.', '106.', '107.', '108.'))
    and u['bank'] != 'verbs'
]
assert len(units) == 4 and sum(u['count'] for u in units) == 270

def baseline(path):
    stored = BASE_DIR / path.relative_to(ROOT)
    assert stored.exists(), stored
    return json.loads(stored.read_text())

fill_spec = importlib.util.spec_from_file_location(
    'fill_training', ROOT / 'courses/spanish-grammar/scripts/fill-training-pack.py'
)
fill = importlib.util.module_from_spec(fill_spec)
fill_spec.loader.exec_module(fill)

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
    chapter_path = next(
        ROOT / x['source'] for x in units
        if x['chapter'] == unit['chapter'] and x['bank'] == 'chapter'
    )
    blocks = {
        block['id']
        for block in gr.read(chapter_path.parent / '01-outline.json')['chapter_outline']['theory_blocks']
    }
    for before, question, row in zip(old_questions, questions, report['questions']):
        assert row['id'] == question['id']
        assert (before != question) == (row['decision'] == 'fixed'), (path, question['id'], row['decision'])
        assert row['decision'] in ('ok', 'fixed')
        assert row['note'] and row['verification'] == 'pending'
        before_content = {k: v for k, v in before.items() if k != 'signature'}
        after_content = {k: v for k, v in question.items() if k != 'signature'}
        totals[
            'content_changed' if before_content != after_content
            else 'signature_only' if before != question
            else 'unchanged'
        ] += 1
        if question['theory_block_id'] not in blocks:
            assert before['theory_block_id'] == question['theory_block_id'], (path, question['id'], 'new theory binding defect')
        if question.get('choices'):
            choice_ids = [choice['id'] for choice in question['choices']]
            assert question['correct_answer'] in choice_ids
            normalized = [' '.join(choice['text'].strip().lower().split()) for choice in question['choices']]
            assert len(set(normalized)) == len(normalized), (path, question['id'], 'duplicate choice text')
        if question['type'] == 'reorder':
            assert question.get('translation_ru')
            assert 2 <= len(question['correct_answer'].split()) <= 20

    if unit['bank'] == 'training':
        for question in questions:
            assert question['signature'] == fill.question_signature(question), (path, question['id'], 'stale signature')
        assert len({q['signature'] for q in questions}) == len(questions), path
        destination = ROOT / 'internal/grammartrainingpack/es' / path.relative_to(ROOT / 'courses/spanish-grammar/training_pack')
        assert path.read_bytes() == destination.read_bytes(), (path, 'embedded mismatch')
    else:
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
        destination = ROOT / f'internal/grammarbundle/es/chapters/{unit["chapter"]}.json'
        assert (path.parent / '05-final.json').read_bytes() == destination.read_bytes(), (path, 'embedded mismatch')

    assert gr.status(live[unit['source']], report)[0] == 'awaiting_verification'
    print(i, unit['language'], unit['bank'], len(questions), dict(Counter(q['decision'] for q in report['questions'])))

    if unit['order'].startswith('108.'):
        corrected_keys = {
            'q001': 'a', 'q002': 'a', 'q003': 'a', 'q009': 'a',
            'q013': 'a', 'q014': 'a', 'q015': 'a', 'q021': 'a',
            'q023': 'a', 'q025': 'a', 'q034': 'a', 'q035': 'a',
            'q037': 'a', 'q042': 'a', 'q043': 'a', 'q045': 'a',
            'q046': 'a', 'q047': 'a', 'q053': 'a', 'q054': 'a',
            'q057': 'a', 'q058': 'a', 'q059': 'a', 'q065': 'a',
            'q066': 'Cuando acabó el discurso, aplaudieron.',
        }
        by_id = {question['id']: question for question in questions}
        assert {
            qid: by_id[qid]['correct_answer'] for qid in corrected_keys
        } == corrected_keys, 'ES 108 corrected answer-key regression'

# Ensure this batch did not introduce a new duplicate signature anywhere in the ES training pack.
assigned = {u['source'] for u in units if u['bank'] == 'training'}
old_signatures = Counter()
new_signatures = Counter()
for path in (ROOT / 'courses/spanish-grammar/training_pack/chapters').rglob('*.json'):
    rel = path.relative_to(ROOT).as_posix()
    previous = baseline(path) if rel in assigned else gr.read(path)
    old_signatures.update(fill.question_signature(q) for q in previous['questions'])
    new_signatures.update(fill.question_signature(q) for q in gr.read(path)['questions'])
assert all(
    count <= max(1, old_signatures[signature])
    for signature, count in new_signatures.items()
), [
    (signature, count, old_signatures[signature])
    for signature, count in new_signatures.items()
    if count > max(1, old_signatures[signature])
]
print('es no new course-wide duplicate signatures')
assert totals == Counter({'content_changed': 52, 'unchanged': 218}), totals
print('Changes:', dict(totals))
print('PASS: all 270 IDs/contracts, choices, theory bindings, report fingerprints, signatures and source-final equality')
