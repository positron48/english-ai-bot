# Журналы триажа content reports

Текстовые журналы **версионируются в git** по дате прогона. Рабочие снимки и JSONL остаются локально в `logs/complaints/` (в `.gitignore`).

## Именование

```
docs/complaints/journal-YYYY-MM-DD-<slug>.md
```

Примеры:

- `journal-2026-06-04-triage.md` — полный прогон EN+ES
- `journal-2026-06-15-en.md` — только English

`<slug>`: `triage`, `en`, `es`, `hotfix` — на усмотрение оператора.

## Как запустить триаж (после настройки)

### 1. Секреты (один раз)

```bash
cp env.example.complaints-prod secrets/complaints-prod.env
# заполнить COMPLAINTS_SERVICE_URL_EN/ES и COMPLAINTS_SERVICE_TOKEN_*
```

На prod ES часто принимается тот же токен, что и EN (см. `secrets/complaints-prod.env`).

### 2. Cursor (рекомендуется)

Команда в чате: **`/content-complaints-triage`**  
или явно: «запусти навык `content-complaints-triage`, course=en|es, apply».

Навык: `.cursor/skills/content-complaints-triage/SKILL.md`

### 3. Makefile — снимок и кластеры (без LLM)

```bash
# только загрузка + группировка (dry-run)
make complaints-triage-dry-en
make complaints-triage-dry-es

# по отдельности
make complaints-fetch-en
make complaints-fetch-es
```

Снимки: `logs/complaints/snapshot-{en|es}-*.json`

### 4. Новый журнал в git (в начале apply-прогона)

```bash
make complaints-journal-new
# или с slug:
make complaints-journal-new SLUG=en
```

Создаёт `docs/complaints/journal-YYYY-MM-DD-<slug>.md` из `journal-TEMPLATE.md`.

Дальше по ходу триажа дописывай блоки «дата → жалоба → изменение».

### 5. Apply на prod (слова / TTS / resolve)

```bash
set -a && . ./secrets/complaints-prod.env && set +a

# слова: дистракторы, TTS (пример)
python3 tools-local/complaints-triage/apply_prod_word_fixes.py

# закрыть все активные жалобы курса
python3 tools-local/complaints-triage/resolve_all_active.py en
python3 tools-local/complaints-triage/resolve_all_active.py es
```

Проверка, что активных нет:

```bash
set -a && . ./secrets/complaints-prod.env && set +a
curl -sS -H "X-Service-Token: $COMPLAINTS_SERVICE_TOKEN_EN" \
  "$COMPLAINTS_SERVICE_URL_EN/api/internal/content-reports?limit=1&status=active&course=en"
```

### 6. Грамматика → релиз

```bash
# правки в courses/*/training_pack, затем:
./scripts/generate-grammar-training-pack.sh en
./scripts/generate-grammar-training-pack.sh es
make check   # при изменениях в english-ai-bot
git commit + push + make tag
kubectl exec -it deployment/english -n english -- /app/import_learning_content --commit
kubectl exec -it deployment/spanish -n spanish -- /app/import_learning_content --commit
```

В журнале в шапке укажи **тег** и коммиты submodule.

### 7. Закоммитить журнал

```bash
git add docs/complaints/journal-YYYY-MM-DD-*.md
git commit -m "docs: content complaints triage journal YYYY-MM-DD"
```

## Локальные артефакты (не в git)

| Путь | Содержимое |
|------|------------|
| `logs/complaints/snapshot-*.json` | fetch с prod |
| `logs/complaints/clusters-*.json` | cluster_reports |
| `logs/complaints/triage-YYYY-MM.jsonl` | JSONL resolve |

Дублировать текст журнала в `logs/` не нужно — канонический файл в `docs/complaints/`.

## Индекс журналов

| Дата | Файл | Курс | Жалоб | Тег |
|------|------|------|-------|-----|
| 2026-06-04 | [journal-2026-06-04-triage.md](journal-2026-06-04-triage.md) | EN+ES | 33 | 0.11.13 |
| 2026-07-10 | [journal-2026-07-10-triage.md](journal-2026-07-10-triage.md) | EN+ES | 13 (3 EN + 10 ES, −3 дубль слов) | _(pending tag)_ |

При новом прогоне добавь строку в эту таблицу.
