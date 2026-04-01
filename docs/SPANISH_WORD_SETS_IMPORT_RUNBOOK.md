# Spanish Word Sets: импорт наборов и догенерация POS-карточек

Документ описывает, как запускать:

1) импорт слов в `word_set_items` из CSV частотности;
2) догенерацию недостающих `training_cards` нужной части речи (`preferred_pos` набора).

## Что уже есть в коде

- Команда импорта наборов: `cmd/import_word_sets_from_csv`
- Команда догенерации POS-карточек: `cmd/fill_missing_set_pos_cards`

Обе команды читают конфиг через `config.Load()`, поэтому для Spanish явно задавай профиль:

- `LEARNING_PAIR=ru-es`
- `LEARNING_NATIVE_LANG=ru`
- `LEARNING_TARGET_LANG=es`
- `LEARNING_APP_CODE=spanish`
- `GRAMMAR_BUNDLE_ID=es`

## Локально

Рабочая директория: корень репозитория `english-ai-bot`.

### 1) Импорт слов в наборы

#### Dry run

```bash
LEARNING_PAIR=ru-es \
LEARNING_NATIVE_LANG=ru \
LEARNING_TARGET_LANG=es \
LEARNING_APP_CODE=spanish \
GRAMMAR_BUNDLE_ID=es \
go run ./cmd/import_word_sets_from_csv \
  --csv "data/spanish_word_freq_pos_ud_top6000.csv" \
  --only-set "Core Verbs — Top 50 (Ranks 1–50)"
```

#### Реальная запись в БД

```bash
LEARNING_PAIR=ru-es \
LEARNING_NATIVE_LANG=ru \
LEARNING_TARGET_LANG=es \
LEARNING_APP_CODE=spanish \
GRAMMAR_BUNDLE_ID=es \
go run ./cmd/import_word_sets_from_csv \
  --csv "data/spanish_word_freq_pos_ud_top6000.csv" \
  --only-set "Core Verbs — Top 50 (Ranks 1–50)" \
  --commit
```

Для полного импорта всех наборов убери `--only-set`.

### 2) Догенерация недостающих POS-карточек

Логика команды:

- обрабатывает только слова, у которых уже есть хотя бы одна обычная `training_card`;
- если карточка с нужным POS уже есть — пропускает;
- если нет — генерирует одну дополнительную карточку под `preferred_pos`;
- применяет те же валидации, что обычная генерация (`ValidateTrainingCardResponse`).

#### Dry run

```bash
LEARNING_PAIR=ru-es \
LEARNING_NATIVE_LANG=ru \
LEARNING_TARGET_LANG=es \
LEARNING_APP_CODE=spanish \
GRAMMAR_BUNDLE_ID=es \
go run ./cmd/fill_missing_set_pos_cards \
  --only-set "Core Verbs — Top 50 (Ranks 1–50)"
```

#### Реальная запись в БД

```bash
LEARNING_PAIR=ru-es \
LEARNING_NATIVE_LANG=ru \
LEARNING_TARGET_LANG=es \
LEARNING_APP_CODE=spanish \
GRAMMAR_BUNDLE_ID=es \
go run ./cmd/fill_missing_set_pos_cards \
  --only-set "Core Verbs — Top 50 (Ranks 1–50)" \
  --commit
```

Для полного прогона всех наборов убери `--only-set`.

## Prod (k3s)

GitOps-манифесты кластера лежат в `../devops-time-host`.  
Запуск команд — внутри pod `spanish` через `kubectl exec`.

### Важно про образ из CI

Сейчас runtime-образ (см. `Dockerfile`) копирует только бинарники:

- `/app/main`
- `/app/backfill_mastering`

Если хочешь запускать команды из этого документа через `kubectl exec`, в CI-образ должны быть добавлены:

- `/app/import_word_sets_from_csv`
- `/app/fill_missing_set_pos_cards`

После выката образа с этими бинарниками команды ниже заработают.

CSV вшивается в образ из основного репозитория:

- источник в репо: `resources/wordsets/spanish_word_freq_pos_ud_top6000.csv`
- путь в контейнере: `/app/data/spanish_word_freq_pos_ud_top6000.csv`

`courses/spanish-grammar` остаётся отдельным репозиторием (submodule), поэтому импорт наборов не зависит от наличия этого submodule в runtime.

### Preflight checks (prod)

Перед dry-run/commit проверь, что pod уже на новом образе и все артефакты на месте:

```bash
kubectl exec -it deployment/spanish -n spanish -- ls -l /app
kubectl exec -it deployment/spanish -n spanish -- ls -l /app/data/spanish_word_freq_pos_ud_top6000.csv
```

Ожидаем увидеть:

- `/app/import_word_sets_from_csv`
- `/app/fill_missing_set_pos_cards`
- `/app/data/spanish_word_freq_pos_ud_top6000.csv`

### 1) Dry run на проде

```bash
kubectl exec -it deployment/spanish -n spanish -- \
  env LEARNING_PAIR=ru-es LEARNING_NATIVE_LANG=ru LEARNING_TARGET_LANG=es LEARNING_APP_CODE=spanish GRAMMAR_BUNDLE_ID=es \
  /app/import_word_sets_from_csv \
    --csv "/app/data/spanish_word_freq_pos_ud_top6000.csv" \
    --only-set "Core Verbs — Top 50 (Ranks 1–50),Core Nouns — Top 50 (Ranks 1–50)"
```

```bash
kubectl exec -it deployment/spanish -n spanish -- \
  env LEARNING_PAIR=ru-es LEARNING_NATIVE_LANG=ru LEARNING_TARGET_LANG=es LEARNING_APP_CODE=spanish GRAMMAR_BUNDLE_ID=es \
  /app/fill_missing_set_pos_cards \
    --only-set "Core Verbs — Top 50 (Ranks 1–50),Core Nouns — Top 50 (Ranks 1–50)"
```

### 2) Реальный запуск на проде

```bash
kubectl exec -it deployment/spanish -n spanish -- \
  env LEARNING_PAIR=ru-es LEARNING_NATIVE_LANG=ru LEARNING_TARGET_LANG=es LEARNING_APP_CODE=spanish GRAMMAR_BUNDLE_ID=es \
  /app/import_word_sets_from_csv \
    --csv "/app/data/spanish_word_freq_pos_ud_top6000.csv" \
    --commit
```

```bash
kubectl exec -it deployment/spanish -n spanish -- \
  env LEARNING_PAIR=ru-es LEARNING_NATIVE_LANG=ru LEARNING_TARGET_LANG=es LEARNING_APP_CODE=spanish GRAMMAR_BUNDLE_ID=es \
  /app/fill_missing_set_pos_cards \
    --commit
```

## Рекомендуемый порядок

1. Создать категории/наборы миграциями.
2. Импортировать `word_set_items` (сначала dry run, потом commit).
3. Дать обычному воркеру сгенерировать базовые карточки.
4. Запустить `fill_missing_set_pos_cards` (сначала dry run, потом commit).
5. Точечно проверить в админке слова из `Core Verbs` и `Core Nouns`.

