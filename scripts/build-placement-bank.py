#!/usr/bin/env python3
"""Validate canonical standalone questions and produce deterministic runtime banks.

No LLM, course-question extraction or template expansion is performed here.
"""
import argparse
import collections
import hashlib
import json
from pathlib import Path
import re
import sys

ROOT = Path(__file__).resolve().parents[1]
LEVELS = ("A1", "A2", "B1", "B2", "C1")

def read(path):
    return json.loads(path.read_text())

def validate(language, check):
    course = ROOT / "courses" / ("english-grammar" if language == "en" else "spanish-grammar")
    source = course / "placement"
    meta = read(source / "skills.json")
    sections = read(course / "config/generation-status.json")["sections"]
    chapters = {c: s for s in sections for c in s["chapter_ids"]}
    skills = {s["id"]: s for s in meta["skills"]}
    assert len(skills) == len(meta["skills"]), "duplicate skills"
    assert meta["language"] == language and meta["course_code"] == language + "_ru"
    for s in skills.values():
        assert s["level"] in LEVELS and s["title"] and s["description"] and s["chapter_ids"], s
        for ch in s["chapter_ids"]:
            assert ch in chapters, f"unknown chapter {ch}"
            assert chapters[ch]["level"] == s["level"], f"course level mismatch {s['id']}: {ch}"
        assert s["section_id"] in {chapters[ch]["section_id"] for ch in s["chapter_ids"]}, s
    items = []
    for path in sorted((source / "items").glob("*/*.json")):
        entries = read(path)
        assert isinstance(entries, list), f"expected array: {path}"
        items.extend(entries)
    ids, prompts = set(), {}
    counts = collections.Counter()
    families = collections.defaultdict(set)
    skill_counts = collections.Counter()
    required = {"id","revision","skill_id","family_id","level","difficulty","context","instruction","prompt","choices","correct_answer","explanation","status"}
    for q in items:
        assert set(q) == required, f"fields {q.get('id')}: missing={required-set(q)} extra={set(q)-required}"
        assert q["id"] not in ids, f"duplicate {q['id']}"
        ids.add(q["id"])
        assert q["id"].startswith(language+".") and q["family_id"].startswith(language+"."), q["id"]
        assert q["revision"] >= 1 and q["status"] == "approved", q["id"]
        assert q["skill_id"] in skills and skills[q["skill_id"]]["level"] == q["level"], q["id"]
        assert q["difficulty"] in (1,2,3), q["id"]
        assert all(isinstance(q[k],str) and q[k].strip() for k in ("instruction","prompt","explanation","family_id")), q["id"]
        assert 3 <= len(q["choices"]) <= 4, q["id"]
        keys = [c["id"] for c in q["choices"]]
        texts = [c["text"].strip().casefold() for c in q["choices"]]
        assert len(set(keys)) == len(keys) and len(set(texts)) == len(texts) and all(texts), q["id"]
        assert q["correct_answer"] in keys, q["id"]
        visible = " ".join(q[k] for k in ("context","instruction","prompt"))
        assert not re.search(r"как (?:в|из) (?:уроке|главе)|согласно (?:теории|уроку)|теоретическ[а-я]+ блок|V[123]",visible,re.I), q["id"]
        normalized = re.sub(r"\s+", " ", (q["context"] + " " + q["prompt"]).strip().casefold())
        if normalized in prompts:
            assert prompts[normalized] == q["family_id"], f"same prompt, different family: {q['id']}"
        prompts[normalized] = q["family_id"]
        counts[q["level"],q["difficulty"]] += 1
        families[q["level"],q["difficulty"]].add(q["family_id"])
        skill_counts[q["skill_id"]] += 1
    assert len(items) >= 400, f"incomplete bank: {len(items)}/400"
    for level in LEVELS:
        assert sum(counts[level,d] for d in (1,2,3)) >= 80, f"incomplete {level}"
        assert len({q['skill_id'] for q in items if q['level']==level}) >= 6, f"too few skills {level}"
        for d in (1,2,3):
            assert len(families[level,d]) >= 16, f"insufficient family reserve {level}/{d}"
    bank = {"course_code": meta["course_code"], "language": language, "skills": sorted(skills.values(),key=lambda s:s['id']), "items": sorted(items,key=lambda q:q['id'])}
    digest = hashlib.sha256(json.dumps(bank,ensure_ascii=False,sort_keys=True,separators=(',',':')).encode()).hexdigest()
    bank["version"] = digest
    payload = json.dumps(bank,ensure_ascii=False,indent=2) + "\n"
    dest = ROOT / "internal/placementbundle" / language / "bank.json"
    if check:
        assert dest.exists() and dest.read_text() == payload, f"stale bundle: {language}"
    else:
        dest.parent.mkdir(parents=True,exist_ok=True)
        dest.write_text(payload)
        report = {"version":digest,"items":len(items),"skills":len(skills),"levels":{l:{"items":sum(counts[l,d] for d in (1,2,3)),"families_by_difficulty":{str(d):len(families[l,d]) for d in (1,2,3)}} for l in LEVELS},"pilot":"pending","note":"Structural validation; language review is recorded separately."}
        (source/"reports").mkdir(exist_ok=True)
        (source/"reports/coverage.json").write_text(json.dumps(report,ensure_ascii=False,indent=2)+"\n")
    print(f"{language}: {len(items)} items, {len(skills)} skills, version {digest[:12]}")

if __name__ == '__main__':
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('language',choices=['en','es','all'])
    p.add_argument('--check',action='store_true',help='validate and require up-to-date generated banks')
    args=p.parse_args()
    try:
        for lang in (('en','es') if args.language=='all' else (args.language,)):
            validate(lang,args.check)
    except (AssertionError,KeyError,ValueError,OSError) as e:
        print(f'Placement validation failed: {e}',file=sys.stderr)
        sys.exit(1)
