#!/usr/bin/env python3
"""Opt-in lexical smoke: four bounded requests, no retries, no database writes."""
import argparse
import json
import os
from pathlib import Path
import urllib.request

CASES = [('apatecer', False), ('apetecer', True), ('xqzplm', False), ('software', True)]

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--live', action='store_true', help='Explicitly authorize four small paid API calls')
    parser.add_argument('--kind', choices=['training','dictionary'], default='training')
    args = parser.parse_args()
    if not args.live:
        print(json.dumps({'calls': 0, 'cases': CASES, 'note': 'Use --live to run; max 4 requests, 700 output tokens each, no retries.'}))
        return
    key = os.getenv('OPENROUTER_API_KEY') or os.getenv('TTS_API_KEY')
    if not key:
        raise SystemExit('OPENROUTER_API_KEY required')
    prompt = (Path(__file__).resolve().parents[1] / ('prompts/training-card-ru-es.txt' if args.kind == 'training' else 'prompts/teacher-ru-es.txt')).read_text()
    for k,v in [('target_lang','es'),('native_lang','ru'),('pair','ru→es')]:
        prompt = prompt.replace('{{'+k+'}}',v)
    total_cost = 0
    failures = []
    for word, accept in CASES:
        body = {'model':'openai/gpt-4o-mini','messages':[{'role':'user','content':prompt+'\n\n'+word}],
                'temperature':0,'max_tokens':700,'response_format':{'type':'json_object'}}
        req = urllib.request.Request('https://openrouter.ai/api/v1/chat/completions',data=json.dumps(body).encode(),
            headers={'Authorization':'Bearer '+key,'Content-Type':'application/json'})
        with urllib.request.urlopen(req,timeout=60) as response:
            result = json.load(response)
        card = json.loads(result['choices'][0]['message']['content'])
        senses = card.get('senses',[])
        ok = bool(senses) and not card.get('error') if accept else not senses and bool(card.get('error'))
        if args.kind == 'dictionary':
            ok = card.get('error') is (not accept)
            if not accept: ok = ok and not card.get('examples') and 'verb_forms' not in card
        if accept:
            ok = ok and card.get('lemma') == word and (card.get('word_target') == word if args.kind == 'training' else card.get('input_word') == word)
        if word == 'apetecer' and senses:
            ok = ok and any('хот' in s.get('word_native','').lower() or 'жел' in s.get('word_native','').lower() for s in senses)
            ok = ok and 'апатир' not in json.dumps(card,ensure_ascii=False).lower()
        cost = result.get('usage',{}).get('cost')
        if cost is not None: total_cost += cost
        print(json.dumps({'word':word,'ok':bool(ok),'error':card.get('error'),'cost_usd':cost},ensure_ascii=False),flush=True)
        if not ok: failures.append(word)
    print(json.dumps({'reported_cost_usd':total_cost,'failed':failures}))
    if failures: raise SystemExit(1)

if __name__ == '__main__':
    main()
