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

# Предпросмотр закрытия конкретных проверенных ID; добавить --apply для выполнения
python3 tools-local/complaints-triage/resolve_all_active.py en --report-ids "$VERIFIED_REPORT_IDS" --reason "Проверено после импорта"
python3 tools-local/complaints-triage/resolve_all_active.py es --report-ids "$VERIFIED_REPORT_IDS" --reason "Проверено после импорта"
```

Проверка, что активных нет:

```bash
set -a && . ./secrets/complaints-prod.env && set +a
curl -sS -H "X-Service-Token: $COMPLAINTS_SERVICE_TOKEN_EN" \
  "$COMPLAINTS_SERVICE_URL_EN/api/internal/content-reports?limit=200&status=active"
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
| 2026-07-10 | [journal-2026-07-10-triage.md](journal-2026-07-10-triage.md) | EN+ES | 13 (3 EN + 10 ES, −3 дубль слов) | 0.12.183 |
| 2026-08-18 | [journal-2026-08-18-triage.md](journal-2026-08-18-triage.md) | EN+ES | 2 (0 EN + 2 ES) | ожидает релиза |
| 2026-09-03 | [journal-2026-09-03-triage.md](journal-2026-09-03-triage.md) | EN+ES | 7 (6 reading + 1 grammar) | 0.12.189 |

При новом прогоне добавь строку в эту таблицу.

## Полнота выборки и чтение

Сборщик всегда читает все страницы без серверного `course`: старые версии API отбрасывают `reading_text` с ID `free_es_*`. Курс определяется локально; неизвестные остаются видимыми в `unknown_course_report_ids`. Сводка строится из снимка. Проверять `complete`; legacy API не подтверждает отсутствие жалоб. Оба URL могут возвращать общую БД — не считать одинаковые ID дважды.

Чтение: канонические `courses/*-grammar/reading/texts`, индекс `reading/index.json` и `assets/reading`. Проверять текст, перевод, вопросы, ключи, аудио и соответствие обложки описанию; выполнить поиск аналогичных ошибок. `GET .../{id}` на обновлённом API отдаёт `reading_text` из текущей БД отдельно от исторического payload.

Для unified Linglow после здорового rollout оператор запускает `kubectl -n linglow exec deploy/linglow -- /app/import_learning_content --course-code es_ru --commit`. Локальный kubectl не использовать (AGENTS.md). Закрывать только явно указанные ID после проверки production; `resolve_all_active.py` без `--apply` ничего не изменяет.

Канонический production URL для обоих курсов: `https://linglow.qantrix.ru`. Старые домены перенаправляют GET, но POST resolve через urllib завершается HTTP 308. В `COMPLAINTS_SERVICE_URL_EN` и `COMPLAINTS_SERVICE_URL_ES` задавать канонический URL. Health endpoint: `/health`.
