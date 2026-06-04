# Content complaints triage (no LLM)

Cursor skill: `.cursor/skills/content-complaints-triage/`.  
Runbook + versioned journals: **`docs/complaints/README.md`**.

## Quick start

```bash
cp env.example.complaints-prod secrets/complaints-prod.env
# fill COMPLAINTS_SERVICE_TOKEN from k8s secrets

make complaints-journal-new          # docs/complaints/journal-YYYY-MM-DD-triage.md
make complaints-triage-dry-en
# apply: resolve_all_active.py, then git add docs/complaints/journal-*.md
```

## Scripts

| Script | Purpose |
|--------|---------|
| `fetch_reports.py` | Download active reports + summary to `logs/complaints/snapshot-{course}-*.json` |
| `cluster_reports.py` | Group snapshot into priority clusters (dry-run plan) |
| `new_journal.py` | Create dated journal under `docs/complaints/` (`make complaints-journal-new`) |
| `append_triage_log.py` | Append apply-mode line to `logs/complaints/triage-YYYY-MM.jsonl` |
| `resolve_all_active.py` | `resolve-bulk` all active reports for `en` or `es` |
| `apply_prod_word_fixes.py` | Example: aunque distractors + TTS regenerate on ES prod |

## Apply journal example

```bash
python3 tools-local/complaints-triage/append_triage_log.py \
  --course en --run-id 20260604-1 \
  --cluster-key "bad_audio::word_training|42||hello|" \
  --category bad_audio --action tts_regenerate \
  --report-ids 101,102 --resolve-status ok
```
