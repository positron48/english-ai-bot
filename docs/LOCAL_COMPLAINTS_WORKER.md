# Local Complaints Worker

Этот документ описывает локальный hourly-процесс обработки grammar-жалоб, вынесенный из релизного Go-сервиса.

## Что находится в релизном сервисе

- `GET /api/internal/content-reports/grammar`
- `POST /api/internal/content-reports/grammar/resolve-bulk`
- Авторизация через `X-Service-Token`.
- Основной release-env: `COMPLAINTS_SERVICE_TOKEN`.
- Backward-compatible env: `WEBAPP_INTERNAL_SERVICE_TOKENS_JSON`.

В сервисе нет LLM-анализа, нет редактирования `courses/`, нет локального scheduler.

## Локальный инструмент

- Скрипт: `tools-local/complaints-worker/worker.py`
- Обертка: `tools-local/complaints-worker/run.sh`
- `dry-run` по умолчанию.
- Для реального удаления вопросов и resolve нужен флаг `--apply`.

## Переменные окружения

- `COMPLAINTS_SERVICE_URL` - базовый URL сервиса, например `http://127.0.0.1:8184`
- `COMPLAINTS_SERVICE_TOKEN` - service token для internal API
- `LLAMACPP_URL` - URL локального `llama.cpp` OpenAI-compatible API, например `http://127.0.0.1:8080`
- `LLAMACPP_MODEL` - имя модели
- `LLAMACPP_START_CMD` - опциональная команда автозапуска `llama.cpp`, если `LLAMACPP_URL` недоступен
- `LLAMACPP_START_MAX_WAIT_SEC` - сколько ждать готовности после автозапуска (по умолчанию `45`)
- `COURSE_SCOPE` - `english|spanish|both`
- `WORKSPACE_ROOT` - опционально, путь к workspace

## Ручной запуск

Dry-run:

```bash
python3 tools-local/complaints-worker/worker.py
```

Через `make` перед запуском воркера автоматически выполняется проверка `llama.cpp`:

- если `LLAMACPP_URL` отвечает, воркер стартует сразу;
- если не отвечает, будет выполнен `LLAMACPP_START_CMD` (если задан), затем ожидание готовности.

Боевой запуск:

```bash
python3 tools-local/complaints-worker/worker.py --apply
```

## JSONL журнал

Файл:

- `logs/complaints/complaints-YYYY-MM.jsonl`

Основные поля записи:

- `timestamp`
- `run_id`
- `course`
- `chapter_id`
- `theory_block_id`
- `question_ids`
- `report_ids`
- `llm_diagnosis`
- `action` (`dry_run|removed|noop|error`)
- `error`
- `resolve_status`
- `hash_before`
- `hash_after`
- `training_pack_relpath`
- `dry_run`

Список затронутых блоков:

- `logs/complaints/changed-theory-blocks-YYYYMMDDHH.json`

Этот файл используется как вход для последующей пакетной ревалидации/перегенерации блоков.

## launchd (каждый час)

Шаблон plist:

- `tools-local/complaints-worker/launchd/com.englishai.complaints-worker.plist`

Установка:

```bash
mkdir -p "$HOME/Library/LaunchAgents"
cp tools-local/complaints-worker/launchd/com.englishai.complaints-worker.plist "$HOME/Library/LaunchAgents/"
launchctl unload "$HOME/Library/LaunchAgents/com.englishai.complaints-worker.plist" 2>/dev/null || true
launchctl load "$HOME/Library/LaunchAgents/com.englishai.complaints-worker.plist"
```

Проверка:

```bash
launchctl list | rg complaints-worker
```

Логи:

- `logs/complaints/launchd.stdout.log`
- `logs/complaints/launchd.stderr.log`

## Smoke-check

1. Запустить dry-run и убедиться, что:
  - journal создан;
  - changed-blocks создан;
  - запросы в internal API проходят.
2. Запустить `--apply` на тестовом наборе жалоб.
3. Проверить:
  - вопросы удалены из `courses/*/training_pack/chapters/*.questions.json`;
  - API вернул успешный bulk resolve;
  - в `content_reports` статус сменился на `resolved`.

## Практический workflow (коротко)

1. Подготовить env:
  - `COMPLAINTS_SERVICE_URL` (или `COMPLAINTS_SERVICE_URL_EN/ES` для разных URL),
  - `COMPLAINTS_SERVICE_TOKEN`,
  - `LLAMACPP_URL`,
  - `LLAMACPP_START_CMD` (рекомендуется).
2. Запустить dry-run:
  - `make complaints-dry-both`
3. Проверить артефакты:
  - `logs/complaints/complaints-YYYY-MM.jsonl`,
  - `logs/complaints/changed-theory-blocks-YYYYMMDDHH.json`.
4. Принять решение по группам:
  - если в `llm_diagnosis` видно системную проблему блока -> переходить к `--apply`,
  - если diagnosis шумный/сомнительный -> сначала ручная проверка блока и повторный dry-run.
5. Применить изменения:
  - `make complaints-apply-both`
6. После apply обязательно:
  - пересобрать training-pack для затронутых блоков,
  - прогнать валидацию затронутых блоков целиком (не только удаленные вопросы),
  - при необходимости перегенерировать вопросы по блоку.
7. Повторить dry-run/apply цикл до стабилизации (пока не останутся `groups=0`).

## Как читать текущий результат dry-run

Если вывод вида:

- первый JSON: `"groups": 0` - для одного профиля жалоб нет;
- второй JSON: `"groups": 4` - для второго профиля найдено 4 группы (по theory block).

Это нормальное поведение для `complaints-dry-both`: команды выполняются последовательно для двух профилей.
В `dry_run`:

- вопросы в `courses/` не удаляются;
- `resolve-bulk` в API не вызывается;
- журнал и `changed-theory-blocks` уже пригодны для последующей ревалидации/перегенерации.

## Полный цикл до прод-обновления

Ниже рабочий цикл без пропусков для твоего сценария "жалобы -> чистка -> реген -> релиз".

### Шаг 1. Применить жалобы сразу (без dry-run) + автоанализ

```bash
make complaints-both
```

Эта команда делает:

- `complaints-apply-both` (удаляет проблемные вопросы и отправляет bulk resolve),
- `complaints-improve-both` (LLM-разбор журнала, группировка причин, рекомендации по улучшению).

Артефакты после шага:

- `logs/complaints/complaints-YYYY-MM.jsonl`
- `logs/complaints/changed-theory-blocks-YYYYMMDDHH.json`
- `logs/complaints/improvement-plan-*.json`
- `logs/complaints/improvement-plan-*.md`

Если нужен полный автоматический цикл до пересборки bundle:

```bash
make complaints-cycle-both
```

Команда выполнит:

1. `complaints-both` (apply + improvement plan),
2. `complaints-prompt-autofix-es` (автообновление prompt с ограниченным авто-блоком),
3. `complaints-prompt-regression` (регресс-проверка prompt + validator),
4. `courses/spanish-grammar: make training-pack-fill`,
5. `make grammar-bundle`.

### Шаг 2. Перегенерация вопросов

Минимум:

```bash
# в папке courses/spanish-grammar (или english-grammar для EN)
make training-pack-fill
```

Важно: `training-pack-fill` должен закрыть весь theory block, где были ошибки, а не только конкретный удаленный вопрос.

### Шаг 3. Обновить embedded bundle приложения

Из корня `english-ai-bot`:

```bash
make grammar-bundle
```

### Шаг 4. Финальная проверка перед релизом

```bash
make check
```

### Шаг 5. Ручной релиз

- `git add/commit`,
- `make tag`,
- push/tag (по вашему текущему release-flow).

### Что делает авторазбор улучшений

`complaints-improve-both` берет последний run из журнала и через LLM строит:

- `global_patterns` (повторяющиеся классы ошибок),
- `prompt_updates` (что менять в генераторном prompt),
- `validator_updates` (что усилить в автоматической валидации),
- `block_actions` (приоритетные блоки и действие),
- `execution_order` (порядок следующих шагов).

То есть разбор журнала и предложения по улучшению системы — автоматизированы.

### Как работает автообновление prompt

- Скрипт: `tools-local/complaints-worker/apply-prompt-improvements.py`
- Вход: последний `improvement-plan-*.json`
- Выход: обновление `courses/spanish-grammar/prompts/16-training-pack-generator-system.md`
- Механика защиты от бесконечного раздувания:
  - в prompt есть заменяемый блок между маркерами:
    - `<!-- AUTO_COMPLAINTS_GUARDRAILS:START -->`
    - `<!-- AUTO_COMPLAINTS_GUARDRAILS:END -->`
  - на каждом запуске блок **полностью перезаписывается**, а не дописывается.

### Интеграционный regression-тест prompt+validator

- Скрипт: `tools-local/complaints-worker/prompt-validator-regression.py`
- Проверяет:
  - что ключевые инварианты prompt не потеряны (JSON-only, `mcq_single`, маркеры авто-блока),
  - что авто-блок остается компактным,
  - что `validate_question()` из `generate-training-pack.py` пропускает валидный кейс и режет невалидный.

Это быстрый тест **без вызова LLM** (инварианты/контракт).

### Реальный LLM smoke тест генерации

Для проверки "prompt -> LLM generation -> validator ok" есть отдельная команда:

```bash
make complaints-prompt-integration-es
```

Она:

1. проверяет/поднимает `llama.cpp`,
2. запускает `generate-training-pack.py` на 1 block (append, `questions-per-block=1`),
3. проверяет, что `training_pack/reports/validation-report.json` имеет `ok=true`.