#!/usr/bin/env python3
import hashlib,importlib.util,json
from pathlib import Path
ROOT=Path(__file__).resolve().parents[4]
BASE=Path('/tmp/grammar-es-counterfactual-baseline')
BATCH=Path(__file__).resolve().parent
ITEMS=json.loads((BATCH/'manifest.json').read_text())
EXPECTED_COMMAND='python3 docs/grammar-review/batches/2026-09-06-es-verbs-counterfactual-fix/check.py'

def read(p): return json.loads(p.read_text())
def fingerprint(p):
 h=hashlib.sha256(); h.update(p.name.encode()); h.update(b'\0'); h.update(p.read_bytes()); h.update(b'\0'); return h.hexdigest()
assert len(ITEMS)==266 and len({x['lemma'] for x in ITEMS})==266
fixed=0
for item in ITEMS:
 source=ROOT/item['source']; embedded=ROOT/item['embedded']; report_path=ROOT/item['report']; base_source=BASE/item['source']; base_report=BASE/item['report']
 data,old=read(source),read(base_source); report,old_report=read(report_path),read(base_report)
 assert source.read_bytes()==embedded.read_bytes(),item['lemma']
 assert [(c['scope'],c['person'],c['number']) for c in data['cards']]==[(c['scope'],c['person'],c['number']) for c in old['cards']]
 changed=[]
 for before,after in zip(old['cards'],data['cards']):
  if before==after: continue
  b=dict(before); a=dict(after); old_ru=b.pop('translation_ru_full'); new_ru=a.pop('translation_ru_full'); assert a==b
  qid=f"{after['scope']}:{after['person']}:{after['number']}"; assert qid in item['ids']; assert old_ru.startswith('Зная это заранее, '); assert new_ru.startswith('Если бы '); assert 'знал об этом заранее' in new_ru or 'знали об этом заранее' in new_ru; assert not any(x in new_ru for x in ['долженл','должныли']); changed.append(qid)
 assert changed==item['ids'] and len(changed)==6,(item['lemma'],changed)
 assert report['source_sha256']==fingerprint(source)
 for key in ['version','source','context_sha256','editor','verifier','phase','reviewed_at','verified_at','verification_note']:
  assert report[key]==old_report[key],(item['lemma'],key)
 old_rows={q['id']:q for q in old_report['questions']}; rows={q['id']:q for q in report['questions']}; assert rows.keys()==old_rows.keys()
 for qid,row in rows.items():
  before=old_rows[qid]
  for key in ['id','decision','verification']:
   assert row[key]==before[key]
  if qid in item['ids']:
   assert row['note'].startswith(before['note']) and 'контрфактического условия' in row['note']
  else: assert row==before
 assert report['checks'][-1]['command']==EXPECTED_COMMAND and report['checks'][-1]['result']=='pass'
 fixed+=len(changed)
for name,expected_min in [('shard1',67),('shard2',67),('shard3',66),('shard4',66)]:
 rows=[json.loads(x) for x in (BATCH/f'checkpoints-{name}.jsonl').read_text().splitlines()]
 assert len(rows)>=expected_min
print('PASS:',len(ITEMS),'files;',fixed,'counterfactual translations; reports and embedded copies synchronized')
