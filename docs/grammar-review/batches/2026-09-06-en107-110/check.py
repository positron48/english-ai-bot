#!/usr/bin/env python3
"""Read-only validation of the EN 107-110 editorial batch against its baseline."""
import importlib.util
import json
import pathlib
import subprocess
from collections import Counter

ROOT = pathlib.Path(__file__).resolve().parents[4]
BATCH = pathlib.Path(__file__).resolve().parent
BASE_SHA = '1ec076ada14b88fba983a482f3cfcdc6d37ee7ca'
BASE_DIR = pathlib.Path('/tmp/grammar-en107110-baseline')

spec = importlib.util.spec_from_file_location('grammar_review', ROOT / 'scripts/grammar-review.py')
gr = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gr)
units = [
    u for u in gr.inventory()
    if u['language'] == 'en'
    and u['order'].startswith(('107.', '108.', '109.', '110.'))
    and u['bank'] != 'verbs'
]
assert len(units) == 24 and sum(u['count'] for u in units) == 427

def baseline(path):
    stored = BASE_DIR / path.relative_to(ROOT)
    if stored.exists():
        return json.loads(stored.read_text())
    rel = path.relative_to(ROOT / 'courses/english-grammar').as_posix()
    return json.loads(subprocess.check_output([
        'git', '-C', str(ROOT / 'courses/english-grammar'), 'show', f'{BASE_SHA}:{rel}'
    ], text=True))

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
            assert len({choice['text'].strip().lower() for choice in question['choices']}) == len(question['choices']), (path, question['id'])
        if question['type'] == 'reorder':
            assert question.get('translation_ru')
            assert 2 <= len(question['correct_answer'].split()) <= 12

    if unit['bank'] == 'training':
        gen_spec = importlib.util.spec_from_file_location('gen', path.parents[3] / 'scripts/generate-training-pack.py')
        gen = importlib.util.module_from_spec(gen_spec)
        gen_spec.loader.exec_module(gen)
        baseline_by_id = {q['id']: q for q in old_questions}
        for question in questions:
            errors = gen.validate_question(question, unit['chapter'], blocks)
            baseline_errors = gen.validate_question(baseline_by_id[question['id']], unit['chapter'], blocks)
            assert errors == baseline_errors, (path, question['id'], errors, baseline_errors)
            assert question['signature'] == gen.question_signature(question), (path, question['id'], 'stale signature')
        assert len({q['signature'] for q in questions}) == len(questions), path
    else:
        final = gr.read(path.parent / '05-final.json')
        assert final['question_bank']['questions'] == questions, (path, 'source-final mismatch')
        previous_final = baseline(path.parent / '05-final.json')
        for document in (final, previous_final):
            document['question_bank'].pop('questions')
            document['meta'].pop('updated_at')
        assert final == previous_final, (path, 'non-question final content changed')
        validation = path.parent / '05-validation.json'
        stored_validation = BASE_DIR / validation.relative_to(ROOT)
        assert validation.read_bytes() == stored_validation.read_bytes(), (path, 'validation file was not restored')

    assert gr.status(live[unit['source']], report)[0] == 'awaiting_verification'
    if unit['bank'] == 'chapter':
        source = path.parent / '05-final.json'
        destination = ROOT / f'internal/grammarbundle/en/chapters/{unit["chapter"]}.json'
    else:
        source = path
        destination = ROOT / 'internal/grammartrainingpack/en' / path.relative_to(path.parents[2])
    assert source.read_bytes() == destination.read_bytes(), (path, 'embedded mismatch')
    print(i, unit['language'], unit['bank'], len(questions), dict(Counter(q['decision'] for q in report['questions'])))

gen_spec = importlib.util.spec_from_file_location(
    'gen', ROOT / 'courses/english-grammar/scripts/generate-training-pack.py'
)
gen = importlib.util.module_from_spec(gen_spec)
gen_spec.loader.exec_module(gen)
assigned = {u['source'] for u in units}
old_signatures = Counter()
new_signatures = Counter()
for path in (ROOT / 'courses/english-grammar/training_pack/chapters').rglob('*.json'):
    rel = path.relative_to(ROOT).as_posix()
    previous = baseline(path) if rel in assigned else gr.read(path)
    old_signatures.update(gen.question_signature(q) for q in previous['questions'])
    new_signatures.update(gen.question_signature(q) for q in gr.read(path)['questions'])
assert all(
    count <= max(1, old_signatures[signature])
    for signature, count in new_signatures.items()
), [
    (signature, count, old_signatures[signature])
    for signature, count in new_signatures.items()
    if count > max(1, old_signatures[signature])
]
print('en no new course-wide duplicate signatures')
assert totals == Counter({'unchanged': 275, 'content_changed': 152}), totals
print('Changes:', dict(totals))
print('PASS: all 427 IDs/contracts, choices, theory bindings, report fingerprints, signatures and source-final equality')
