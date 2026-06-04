---
name: content-complaints-triage
description: Триаж жалоб content_reports с prod (EN/ES), исправление слов/карточек/TTS/грамматики, resolve через internal API. Используй при жалобах пользователей, content reports, complaints triage.
---

# Триаж content reports (Cursor skill)

Локальный агент **без llama.cpp**. Источник жалоб по умолчанию: **prod k3s** (`qantrix.ru` / `es.qantrix.ru`).

Подробный runbook: [reference.md](reference.md).

## Перед стартом

1. Скопируй `env.example.complaints-prod` → `secrets/complaints-prod.env` и заполни URL + `COMPLAINTS_SERVICE_TOKEN`.
2. Уточни у пользователя режим: `dry-run` (по умолчанию) или `apply`.
3. Один прогон = один курс: `course=en` **или** `course=es` (не смешивать URL/токены).

Снимок жалоб:

```bash
make complaints-triage-dry-en   # fetch + cluster → logs/complaints/
# или вручную:
set -a && . ./secrets/complaints-prod.env && set +a
COMPLAINTS_SERVICE_URL="$COMPLAINTS_SERVICE_URL_EN" \
  python3 tools-local/complaints-triage/fetch_reports.py --course en
python3 tools-local/complaints-triage/cluster_reports.py logs/complaints/snapshot-en-*.json
```

## Фазы

### A — Fetch

- `GET /api/internal/content-reports?course=&source_type=&cursor=&limit=`
- `GET /api/internal/content-reports/summary?course=`
- Для одной жалобы: `GET /api/internal/content-reports/{id}` (training_card, word_card, tts_status, training_pack_relpath)

### B — Кластеризация

Группируй по `report_category` + entity key:

| source_type | entity_key |
|-------------|------------|
| word_training | training_card_id / word_card_id / word |
| grammar_training | chapter_id + theory_block_id + question_id |

Сортируй кластеры по `summary` count.

### C — Решение (playbook)

| category | действие |
|----------|----------|
| wrong_translation, wrong_example, wrong_distractors, typo | `PUT /api/internal/training/card/{id}` JSON |
| bad_audio | `POST /api/internal/tts/regenerate` + `GET /api/internal/tts/status?word=` |
| wrong_answer, ambiguous, wrong_explanation, theory_mismatch, too_hard | правка JSON в `courses/*-grammar/training_pack/` → bundle → import |
| other | полный `payload` + ручной разбор |

### D — Pattern sweep (обязательно)

После фикса кластера ищи аналоги (grep по courses, тот же theory_block_id, word_card_id, pos).

### E — Resolve (только apply)

`POST /api/internal/content-reports/resolve-bulk` с `report_ids` и `reason`.

Журнал: `logs/complaints/triage-YYYY-MM.jsonl`.

### F — Релиз грамматики (если менялись courses)

См. [reference.md](reference.md) § Grammar sync.

## Проверки

- После Go/web изменений в репо: `make check`
- После правки грамматики: `make grammar-bundle` + smoke API

## Устаревшее

Не используй `tools-local/complaints-worker/worker.py` (llama) для триажа — только fetch из `complaints-triage/`.
