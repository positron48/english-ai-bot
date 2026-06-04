---
name: content-complaints-triage
description: Триаж жалоб content_reports с prod (EN/ES), исправление слов/карточек/TTS/грамматики, resolve через internal API. Используй при жалобах пользователей, content reports, complaints triage.
---

# Триаж content reports (Cursor skill)

Локальный агент **без llama.cpp**. Источник жалоб по умолчанию: **prod k3s** (`qantrix.ru` / `es.qantrix.ru`).

Подробный runbook: [reference.md](reference.md).  
Пошаговый запуск и журналы в git: [docs/complaints/README.md](../../../docs/complaints/README.md).

## Как запустить потом

### Быстро (Cursor)

1. Команда **`/content-complaints-triage`** или текст: «триаж жалоб, course=en, apply».
2. Навык сам пройдёт фазы A–F ниже.

### Вручную (терминал)

```bash
# 1) секреты
cp env.example.complaints-prod secrets/complaints-prod.env   # один раз

# 2) новый журнал в git (в начале apply-прогона)
make complaints-journal-new              # docs/complaints/journal-YYYY-MM-DD-triage.md
make complaints-journal-new SLUG=en      # только EN в имени файла

# 3) снимок + кластеры (dry-run)
make complaints-triage-dry-en
make complaints-triage-dry-es

# 4) правки контента (агент / редактор) + prod API для слов/TTS
set -a && . ./secrets/complaints-prod.env && set +a
python3 tools-local/complaints-triage/apply_prod_word_fixes.py   # при необходимости
python3 tools-local/complaints-triage/resolve_all_active.py en
python3 tools-local/complaints-triage/resolve_all_active.py es

# 5) грамматика: courses → generate-grammar-training-pack → commit → make tag → import в pod

# 6) закоммитить журнал
git add docs/complaints/journal-*.md docs/complaints/README.md
```

## Перед стартом

1. `secrets/complaints-prod.env` из `env.example.complaints-prod` (URL + токены).
2. Режим: `dry-run` или `apply`.
3. Один прогон за раз по курсу: `en` **или** `es` (для fetch/resolve); журнал может описывать оба.

## Фазы

### A — Fetch

```bash
make complaints-fetch-en   # или complaints-fetch-es
```

Снимки: `logs/complaints/snapshot-*.json` (локально, не в git).

### B — Кластеризация

```bash
make complaints-cluster-en   # после fetch EN
```

### C — Решение (playbook)

| category | действие |
|----------|----------|
| wrong_translation, wrong_example, wrong_distractors, typo | `PUT /api/internal/training/card/{id}` |
| bad_audio | `POST /api/internal/tts/regenerate` |
| wrong_answer, ambiguous, … | `courses/*-grammar/training_pack/` → bundle → import |
| other | ручной разбор `GET .../content-reports/{id}` |

### D — Pattern sweep

grep по `theory_block_id`, pos, word; см. `.cursor/rules/content-quality-guardrails.mdc`.

### E — Resolve + журнал (apply)

1. `resolve_all_active.py en|es` или `resolve-bulk`.
2. **Текстовый журнал в git:** дописать `docs/complaints/journal-YYYY-MM-DD-<slug>.md` — на каждую жалобу: дата → суть → что изменено.
3. **JSONL локально:** `logs/complaints/triage-YYYY-MM.jsonl` (`append_triage_log.py`).

Шаблон нового журнала: `make complaints-journal-new` → `docs/complaints/journal-TEMPLATE.md`.

### F — Релиз грамматики

См. [reference.md](reference.md) § Grammar sync; в шапке журнала указать **тег** и коммиты submodule.

## Проверки

- `make check` после изменений в english-ai-bot
- `make grammar-bundle` / `generate-grammar-training-pack.sh` после courses
- активных жалоб 0: `GET /api/internal/content-reports?status=active&course=`

## Устаревшее

Не используй `tools-local/complaints-worker/worker.py` (llama).
