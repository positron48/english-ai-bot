# Content complaints triage (no LLM)

Cursor skill: `.cursor/skills/content-complaints-triage/`.

## Quick start

```bash
cp env.example.complaints-prod secrets/complaints-prod.env
# fill COMPLAINTS_SERVICE_TOKEN from k8s secrets

make complaints-fetch-en
python3 tools-local/complaints-triage/cluster_reports.py logs/complaints/snapshot-en-*.json | less
```

## Scripts

| Script | Purpose |
|--------|---------|
| `fetch_reports.py` | Download active reports + summary to `logs/complaints/snapshot-{course}-*.json` |
| `cluster_reports.py` | Group snapshot into priority clusters (dry-run plan) |
| `append_triage_log.py` | Append apply-mode line to `logs/complaints/triage-YYYY-MM.jsonl` |

## Apply journal example

```bash
python3 tools-local/complaints-triage/append_triage_log.py \
  --course en --run-id 20260604-1 \
  --cluster-key "bad_audio::word_training|42||hello|" \
  --category bad_audio --action tts_regenerate \
  --report-ids 101,102 --resolve-status ok
```
