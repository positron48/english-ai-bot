#!/usr/bin/env python3
"""Read-only validation of the ES chapters 004 editorial batch against its baseline."""
import hashlib
import json
import pathlib
import importlib.util
import subprocess
from collections import Counter

ROOT = pathlib.Path(__file__).resolve().parents[4]
spec = importlib.util.spec_from_file_location("grammar_review", ROOT / "scripts/grammar-review.py")
gr = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gr)
units = [u for u in gr.inventory() if u["language"] == "es" and u["order"].startswith(("004.",)) and u["bank"] != "verbs"]
BATCH = pathlib.Path(__file__).resolve().parent

manifest = gr.read(BATCH / "baseline-manifest.json")
for relative, expected in manifest["files"].items():
    assert hashlib.sha256((BATCH / "baseline" / relative).read_bytes()).hexdigest() == expected, relative

def baseline(path):
    return gr.read(BATCH / 'baseline' / path.relative_to(ROOT / 'courses/spanish-grammar'))

live={u['source']:u for u in gr.inventory()}
totals = Counter()
for i,u in enumerate(units):
 p=ROOT/u['source'];old=baseline(p);now=gr.read(p);r=gr.read(ROOT/u['report'])
 assert {k:v for k,v in old.items() if k!='questions'} == {k:v for k,v in now.items() if k!='questions'},p
 oldqs=old['questions'];qs=now['questions'];assert len(qs)==len(oldqs)==u['count']
 def identity(q):return (q['id'],q['type'],q.get('difficulty'),q.get('theory_block_id'),q.get('chapter_id'),q.get('concept_id'),[x['id'] for x in q.get('choices',[])])
 assert list(map(identity,oldqs))==list(map(identity,qs)),p
 assert len(set(q['id'] for q in qs))==len(qs)
 chapterpath=next(ROOT/x['source'] for x in units if x['chapter']==u['chapter'] and x['bank']=='chapter')
 blocks={b['id'] for b in gr.read(chapterpath.parent/'01-outline.json')['chapter_outline']['theory_blocks']}
 for before,q,row in zip(oldqs,qs,r['questions']):
  assert (before!=q)==(row['decision']=='fixed'),(p,q['id'],row['decision'])
  totals['content_changed' if {k:v for k,v in before.items() if k!='signature'} != {k:v for k,v in q.items() if k!='signature'} else 'signature_only' if before!=q else 'unchanged'] += 1
  assert q['theory_block_id'] in blocks
  if q.get('choices'):
   assert q['correct_answer'] in [c['id'] for c in q['choices']]
   assert len({c['text'].strip().lower() for c in q['choices']})==len(q['choices']),(p,q['id'])
  if q['type']=='reorder':assert q.get('translation_ru') and 2<=len(q['correct_answer'].split())<=10
 if u['bank']=='training':
  ss=importlib.util.spec_from_file_location('gen',p.parents[3]/'scripts/generate-training-pack.py');gen=importlib.util.module_from_spec(ss);ss.loader.exec_module(gen)
  assert all(q['signature']==gen.question_signature(q) for q in qs),p
  assert len({q['signature'] for q in qs})==len(qs),p
 else:
  final=gr.read(p.parent/'05-final.json');assert final['question_bank']['questions']==qs,(p,'source-final mismatch')
  previous_final=baseline(p.parent/'05-final.json')
  for document in (final,previous_final):
   document['question_bank'].pop('questions')
   document['meta'].pop('updated_at')
  assert final==previous_final,(p,'non-question content changed')
 assert gr.status(live[u['source']],r)[0] in ('awaiting_verification', 'done')
 if u['bank']=='chapter':
  source=p.parent/'05-final.json';dest=ROOT/f'internal/grammarbundle/{u["language"]}/chapters/{u["chapter"]}.json'
 else:
  source=p;dest=ROOT/'internal/grammartrainingpack'/u['language']/p.relative_to(p.parents[2])
 assert source.read_bytes()==dest.read_bytes(),(p,'embedded mismatch')
 print(i,u['language'],u['bank'],len(qs),dict(Counter(q['decision'] for q in r['questions'])))
# Ensure edited signatures introduce no new duplicates anywhere in the same course training pack.
for lang,course in [('es','spanish-grammar')]:
 ss=importlib.util.spec_from_file_location('gen',ROOT/'courses'/course/'scripts/generate-training-pack.py');gen=importlib.util.module_from_spec(ss);ss.loader.exec_module(gen)
 oldsigns=Counter();newsigns=Counter()
 for p in (ROOT/'courses'/course/'training_pack/chapters').rglob('*.json'):
  previous = baseline(p) if any(x['source'] == p.relative_to(ROOT).as_posix() for x in units) else gr.read(p)
  oldsigns.update(gen.question_signature(q) for q in previous['questions'])
  newsigns.update(gen.question_signature(q) for q in gr.read(p)['questions'])
 assert all(n<=max(1,oldsigns[s]) for s,n in newsigns.items()),[(s,n,oldsigns[s]) for s,n in newsigns.items() if n>max(1,oldsigns[s])]
 print(lang,'no new course-wide duplicate signatures')
print('Changes:',dict(totals))
print('PASS: all 108 IDs/contracts, choices, theory bindings, report fingerprints, signatures and source-final equality')
