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
| `resolve_all_active.py` | Preview/resolve explicit verified IDs with `--report-ids`, `--reason`, `--apply` |
| `apply_prod_word_fixes.py` | Example: aunque distractors + TTS regenerate on ES prod |

## Apply journal example

```bash
python3 tools-local/complaints-triage/append_triage_log.py \
  --course en --run-id 20260604-1 \
  --cluster-key "bad_audio::word_training|42||hello|" \
  --category bad_audio --action tts_regenerate \
  --report-ids 101,102 --resolve-status ok
```

## Полнота выборки и чтение

Сборщик всегда читает все страницы без серверного `course`: старые версии API отбрасывают `reading_text` с ID `free_es_*`. Курс определяется локально; неизвестные остаются видимыми в `unknown_course_report_ids`. Сводка строится из снимка. Проверять `complete`; legacy API не подтверждает отсутствие жалоб. Оба URL могут возвращать общую БД — не считать одинаковые ID дважды.

Чтение: канонические `courses/*-grammar/reading/texts`, индекс `reading/index.json` и `assets/reading`. Проверять текст, перевод, вопросы, ключи, аудио и соответствие обложки описанию; выполнить поиск аналогичных ошибок. `GET .../{id}` на обновлённом API отдаёт `reading_text` из текущей БД отдельно от исторического payload.

Для unified Linglow после здорового rollout оператор запускает `kubectl -n linglow exec deploy/linglow -- /app/import_learning_content --course-code es_ru --commit`. Локальный kubectl не использовать (AGENTS.md). Закрывать только явно указанные ID после проверки production; `resolve_all_active.py` без `--apply` ничего не изменяет.
