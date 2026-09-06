#!/usr/bin/env python3
"""Read-only validation of the ES 137-140 editorial batch against its baseline."""
import importlib.util
import json
import pathlib
from collections import Counter

ROOT = pathlib.Path(__file__).resolve().parents[4]
BASE_DIR = pathlib.Path('/tmp/grammar-es137140-baseline')

spec = importlib.util.spec_from_file_location('grammar_review', ROOT / 'scripts/grammar-review.py')
gr = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gr)
units = [
    u for u in gr.inventory()
    if u['language'] == 'es'
    and u['order'].startswith(('137.', '138.', '139.', '140.'))
    and u['bank'] != 'verbs'
]
assert len(units) == 4 and sum(u['count'] for u in units) == 281

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

    by_id = {question['id']: question for question in questions}
    if unit['order'].startswith('137.'):
        assert by_id['q_b6_mcq_d5_1']['correct_answer'] in ('a', 'b', 'c', 'd')
        assert 'hubiéramos actuado' in json.dumps(by_id, ensure_ascii=False)
    if unit['order'].startswith('138.'):
        assert by_id['q138_b1_fill_1']['correct_answer'] == 'habría'
        assert by_id['q138_b2_fill_2']['correct_answer'] == 'habríamos'
        assert by_id['q138_b2_err_1']['choices'][0]['text'].startswith('*habría sabido*')
        assert 'устойчивая способность' in by_id['q138_b4_mcq_d3']['explanation'].lower()
        assert by_id['q138_b5_fill_1']['correct_answer'] == 'Ojalá'
        assert 'контрфактуальных моделях этой главы' in by_id['q138_b6_mcq_d5']['explanation']
    if unit['order'].startswith('139.'):
        assert by_id['q139_b1_fill_2']['correct_answer'] == 'habría'
        assert by_id['q139_b2_err_1']['prompt'].startswith('Найдите ошибку: «Me arrepiento no haber')
        assert by_id['q139_b3_mcq_d5']['choices'][3]['text'] == 'Si hubiera sabido, te lo diría ahora.'
        assert 'другое временное и аспектуальное значение' in by_id['q139_b3_tf_2']['explanation']
        assert by_id['q139_b4_fill_1']['correct_answer'] == 'tendría'
        assert 'Полное придаточное с *de que* возможно' in by_id['q139_b6_tf_1']['explanation']
    if unit['order'].startswith('140.'):
        assert by_id['q140_b1_m1']['correct_answer'] == 'b'
        assert 'не безусловная грамматическая ошибка' in by_id['q140_b6_m4']['explanation']

# No normalized prompt+answer duplicate was introduced across the four assigned banks.
def editorial_signature(question):
    normalize = lambda value: ' '.join(str(value).casefold().split())
    return normalize(question.get('prompt', '')), normalize(question.get('correct_answer', ''))

old_signatures = Counter()
new_signatures = Counter()
for unit in units:
    path = ROOT / unit['source']
    old_signatures.update(editorial_signature(q) for q in baseline(path)['questions'])
    new_signatures.update(editorial_signature(q) for q in gr.read(path)['questions'])
assert all(
    count <= max(1, old_signatures[signature])
    for signature, count in new_signatures.items()
), [
    (signature, count, old_signatures[signature])
    for signature, count in new_signatures.items()
    if count > max(1, old_signatures[signature])
]
print('es no new batch duplicate prompt+answer signatures')
assert totals == Counter({'content_changed': 25, 'unchanged': 256}), totals
print('Changes:', dict(totals))
print('PASS: all 281 IDs/contracts, choices, theory bindings, report fingerprints and source-final equality')
