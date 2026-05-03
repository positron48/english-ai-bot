# Verb Forms LLM Training: Runbook

Этот документ описывает:

- как обновляются данные тренировки форм глаголов в проде автоматически;
- как запускать и проверять процесс локально;
- какие smoke-проверки сделать после релиза.

## Что является источником данных

- Источник правды: `courses/spanish-grammar/training_pack/verb_forms/`
  - `index.json`
  - `unlock-gates.json`
  - `lemmas/*.json`
- Рантайм API не читает эти JSON напрямую.
- На релизе JSON синхронизируются в БД через `sync_verb_training_json`:
  - upsert существующих/новых;
  - удаление записей, которых нет в JSON (delete-missing).

## Прод: как это запускается автоматически

## 1) Генерация артефактов в CI

В CI должен быть job, который запускает генерацию:

- `make -C courses/spanish-grammar verb-training-pack-fill`

Скрипт генерации: `courses/spanish-grammar/scripts/generate-verb-forms-training.py`.

Требуемые env для генерации:

- `VERB_TRAINING_INTERNAL_API`
- `WEBAPP_INTERNAL_SERVICE_TOKEN` (или fallback `COMPLAINTS_SERVICE_TOKEN`)
- `AI_URL`
- `AI_MODEL`
- `AI_API_KEY`

## 2) Сборка образа

`Dockerfile` уже включает:

- бинарник `/app/sync_verb_training_json`;
- артефакты `courses/spanish-grammar/training_pack/verb_forms` внутрь образа.

## 3) Автосинк БД в k3s при rollout

В `devops-time-host/apps/spanish/base/deployment.yaml` добавлен initContainer:

- `sync-verb-training-json`
- запускает:
  - `/app/sync_verb_training_json --course-root /app/courses/spanish-grammar`

Это означает:

- при каждом новом rollout синк выполняется автоматически;
- основной контейнер стартует после успешного sync;
- ручные команды на проде не нужны.

## 4) Что проверить после выката

На кластере:

- pod стартует без ошибок initContainer;
- `/health` отвечает `200`;
- `GET /api/verb-training/upcoming` для тестового пользователя возвращает данные;
- в карточке тренировки есть `(i)` и открывается таблица спряжений.

Полезные команды:

```bash
kubectl get pods -n spanish -l app=spanish
kubectl logs -n spanish deploy/spanish -c sync-verb-training-json --tail=200
kubectl logs -n spanish deploy/spanish -c spanish --tail=200
```

## Локальный запуск

## 1) Подготовка env

Убедиться, что есть:

- корневые `.env` и `.env.es` (или эквивалентные переменные в shell);
- при необходимости `courses/spanish-grammar/.env.local`.

Ключевое:

- `AI_URL`, `AI_MODEL`, `AI_API_KEY`
- `VERB_TRAINING_INTERNAL_API` (обычно `http://127.0.0.1:8184`)
- `WEBAPP_INTERNAL_SERVICE_TOKEN` (или `COMPLAINTS_SERVICE_TOKEN`)

## 2) Генерация JSON

```bash
make -C courses/spanish-grammar verb-training-pack-fill
```

## 3) Синк JSON в БД

```bash
make verb-training-sync-db
```

Или напрямую:

```bash
go run ./cmd/sync_verb_training_json --course-root courses/spanish-grammar
```

Dry-run:

```bash
go run ./cmd/sync_verb_training_json --course-root courses/spanish-grammar --dry-run
```

## 4) Обновление embedded артефактов

```bash
make grammar-bundle
```

## 5) Полная проверка

```bash
make check
```

## Локальное тестирование UI/API

## 1) Включение фичи

Нужны Spanish-конфиги, в том числе:

- `LEARNING_TARGET_LANG=es`
- `GRAMMAR_BUNDLE_ID=es`
- `SPANISH_VERB_FORMS_ENABLED=true`

## 2) Smoke API

- Internal pending:

```bash
curl -H "X-Service-Token: $WEBAPP_INTERNAL_SERVICE_TOKEN" \
  "$VERB_TRAINING_INTERNAL_API/api/internal/verb-training/pending?limit=20&cursor=0"
```

- Обычные endpoints:
  - `POST /api/verb-training/start`
  - `GET /api/verb-training/current`
  - `POST /api/verb-training/answer`
  - `GET /api/verb-training/upcoming`

## 3) Smoke UI

Открыть страницу тренировки форм:

- должна приходить карточка с cloze-предложением и RU переводом;
- в choice-режиме варианты берутся из БД (без runtime эвристик);
- `(i)` в правом нижнем углу открывает popup с полной таблицей спряжений (`/api/vocab/{word_card_id}/verb-forms`).

## Админка Verb Forms (визуальная проверка)

Для ручной проверки сгенерированных `verb_forms`-артефактов есть lightweight админка курса.

Запуск:

```bash
make -C courses/spanish-grammar training-pack-admin
```

или алиас:

```bash
make -C courses/spanish-grammar verb-forms-admin
```

URL:

- `http://127.0.0.1:8010/`

В админке доступны 2 режима:

- `Training Pack` — существующий просмотр вопросов по theory blocks;
- `Verb Forms` — просмотр глагольных артефактов.

Что есть в режиме `Verb Forms`:

- слева:
  - поиск по lemma,
  - список глаголов из `training_pack/verb_forms/index.json`;
- справа:
  - мета выбранной lemma (`word_card_id`, `cards_count`, `scopes_count`, `generated_at`),
  - список scopes (времена/наклонения),
  - таблица по выбранному scope:
    - `person`, `number`,
    - `surface_form`,
    - `question_es_with_blank`,
    - `translation_ru_full`,
    - `options`.

Backend endpoints админки для verb forms:

- `GET /api/verb-forms/lemmas?q=...`
- `GET /api/verb-forms/lemma?lemma=...`

Эти endpoints читают данные из:

- `courses/spanish-grammar/training_pack/verb_forms/index.json`
- `courses/spanish-grammar/training_pack/verb_forms/lemmas/*.json`

## Частые проблемы и диагностика

- `401` на `/api/internal/verb-training/pending`:
  - неверный `X-Service-Token`.
- генератор пропускает леммы:
  - проверяйте ответ pending API и логи `generate-verb-forms-training.py`.
- sync падает на валидации:
  - в конкретном `lemmas/<lemma>.json` нет полного покрытия `16 scopes x 6 slots`.
- в UI нет карточек:
  - проверьте unlock-gates, доступ к главам и что `SPANISH_VERB_FORMS_ENABLED=true`.

