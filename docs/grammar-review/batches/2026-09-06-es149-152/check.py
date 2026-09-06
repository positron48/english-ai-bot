#!/usr/bin/env python3
"""Read-only validation of the ES 149-152 editorial batch against its baseline."""
import importlib.util
import json
import pathlib
from collections import Counter

ROOT = pathlib.Path(__file__).resolve().parents[4]
BASE_DIR = pathlib.Path('/tmp/grammar-es149152-baseline')

spec = importlib.util.spec_from_file_location('grammar_review', ROOT / 'scripts/grammar-review.py')
gr = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gr)
units = [
    u for u in gr.inventory()
    if u['language'] == 'es'
    and u['order'].startswith(('149.', '150.', '151.', '152.'))
    and u['bank'] != 'verbs'
]
assert len(units) == 4 and sum(u['count'] for u in units) == 112

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
            answer = question['correct_answer']
            if isinstance(answer, list):
                assert answer and set(answer).issubset(choice_ids)
            else:
                assert answer in choice_ids
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
    if unit['order'].startswith('149.'):
        assert by_id['q149_b1_m3']['choices'][0]['text'].startswith('Иногда яснее вернуть глагол')
        assert by_id['q149_b1_r4']['choices'][3]['text'].startswith('El análisis del impacto')
        assert by_id['q149_b3_m1']['choices'][0]['text'].startswith('Определить главное существительное')
        assert 'рекомендует' in by_id['q149_b6_m2']['explanation']
    if unit['order'].startswith('150.'):
        assert by_id['q150_b1_m1']['choices'][0]['text'].startswith('Según los datos preliminares')
        assert 'возможны оба наклонения' in by_id['q150_b2_m1']['explanation']
        assert by_id['q150_b2_m2']['choices'][1]['text'] == 'Quizás el acuerdo se haya firmado ayer.'
        assert 'la probabilidad de lluvia' in by_id['q150_b6_m1']['explanation']
    if unit['order'].startswith('151.'):
        assert by_id['q151_b1_m3']['choices'][0]['text'].startswith('Formal: Le escribo para solicitar una ampliación')
        assert 'cómo es el día a día' in by_id['q151_b4_m2']['prompt']
        assert 'este aspecto es central' in by_id['q151_b7_m3']['prompt']
        assert 'не переносит в речь' in by_id['q151_b7_r4']['explanation']
    if unit['order'].startswith('152.'):
        assert 'прямо и формально' in by_id['q152_b2_m2']['explanation']
        assert '*¿___ remitirme la factura' in by_id['q152_b2_m3']['prompt']
        assert by_id['q152_b3_m2']['choices'][0]['text'].endswith('Te avisaré en cuanto tenga una nueva fecha.')
        assert '___ parece bien?*' in by_id['q152_b3_m3']['prompt']
        assert by_id['q152_b4_m2']['choices'][0]['text'].startswith('Lamentablemente, no me será posible terminar')
        assert '*¿___ confirmarme' in by_id['q152_b5_m3']['prompt']

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
assert totals == Counter({'content_changed': 51, 'unchanged': 61}), totals
print('Changes:', dict(totals))
print('PASS: all 112 IDs/contracts, choices, theory bindings, report fingerprints and source-final equality')
