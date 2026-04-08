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

### 0) Очистка CSV от мусора (обязательно перед импортом)

Чистка удаляет из источника:

- имена собственные (`PROPN`);
- не-слова (`1993`, `&`, токены с цифрами/знаками);
- явный англоязычный мусор (`the`, `a`, `an`).

```bash
make clean-spanish-csv
```

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

`import_word_sets_from_csv` перезаписывает `word_set_items` существующих наборов (наборы/категории не пересоздаёт), поэтому после удаления мусорных слов из CSV этот шаг «схлопывает» пробелы и снова заполняет наборы подряд по популярности.

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

Runtime-образ (см. `Dockerfile`) содержит:

- `/app/main`
- `/app/backfill_mastering`
- `/app/import_word_sets_from_csv`
- `/app/fill_missing_set_pos_cards`
- `/app/revalidate_training_cards`
- `/app/data/spanish_word_freq_pos_ud_top6000.csv`

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
- `/app/revalidate_training_cards`
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

## End-to-end: что делать сейчас (prod)

Ниже порядок, который учитывает все последние изменения: очистка словарного источника, локализация названий на испанский, перевалидация/перегенерация карточек и пересборка слов в наборах.

1. **Выкатить свежий образ `spanish`** (в нём уже есть):
   - миграция `000004_localize_spanish_word_sets_to_es`;
   - обновлённый `import_word_sets_from_csv` с фильтрацией мусора/`PROPN`;
   - бинарник `/app/revalidate_training_cards`.

2. **Убедиться, что миграция локализации применилась**:
   ```bash
   kubectl logs -n spanish deployment/spanish --tail=200 | rg "applied sql migration|000004_localize_spanish_word_sets_to_es"
   ```

3. **Перевалидировать текущие карточки (сначала dry-run, потом commit)**:
   ```bash
   kubectl exec -it deployment/spanish -n spanish -- \
     env LEARNING_PAIR=ru-es LEARNING_NATIVE_LANG=ru LEARNING_TARGET_LANG=es LEARNING_APP_CODE=spanish GRAMMAR_BUNDLE_ID=es \
     /app/revalidate_training_cards
   ```
   ```bash
   kubectl exec -it deployment/spanish -n spanish -- \
     env LEARNING_PAIR=ru-es LEARNING_NATIVE_LANG=ru LEARNING_TARGET_LANG=es LEARNING_APP_CODE=spanish GRAMMAR_BUNDLE_ID=es \
     /app/revalidate_training_cards --commit
   ```

4. **Пересобрать слова в существующих наборах (схлопнуть пробелы, без пересоздания наборов)**:
   ```bash
   kubectl exec -it deployment/spanish -n spanish -- \
     env LEARNING_PAIR=ru-es LEARNING_NATIVE_LANG=ru LEARNING_TARGET_LANG=es LEARNING_APP_CODE=spanish GRAMMAR_BUNDLE_ID=es \
     /app/import_word_sets_from_csv \
       --csv "/app/data/spanish_word_freq_pos_ud_top6000.csv" \
       --commit
   ```

5. **Дождаться фоновой догенерации карточек воркером** и проверить логи:
   ```bash
   kubectl logs -n spanish deployment/spanish --tail=200
   ```

6. **Опционально**: заполнить недостающие карточки под `preferred_pos`:
   ```bash
   kubectl exec -it deployment/spanish -n spanish -- \
     env LEARNING_PAIR=ru-es LEARNING_NATIVE_LANG=ru LEARNING_TARGET_LANG=es LEARNING_APP_CODE=spanish GRAMMAR_BUNDLE_ID=es \
     /app/fill_missing_set_pos_cards --commit
   ```

7. **Проверка в админке**:
   - категории/наборы/описания — на испанском;
   - в наборах снова плотные списки слов (без «дыр» от удалённого мусора);
   - для проблемных слов `definition_ru` на русском.

### Важный ответ на вопрос про переименование наборов

Да, **слова в наборах будут корректно пересобраны и после переименования на испанский**, потому что импорт берёт:

- `preferred_pos` из `word_sets`;
- диапазон рангов из `title` (например, `rangos 1–50`, `rangos 501–550` и т.д.).

Формат испанских названий в миграции сохранён с тем же шаблоном диапазонов (`rangos X–Y`), поэтому «схлопывание» и повторный импорт продолжают работать корректно.

## Быстрый восстановительный прогон (локально)

После ручной чистки БД или удаления «кривых» слов:

```bash
make clean-spanish-csv
make sync-spanish-word-sets
```

Это обновит списки слов в уже существующих наборах без удаления/создания самих наборов.

## Быстрый восстановительный прогон (prod, k3s)

Команда ниже перезаписывает `word_set_items` в существующих Spanish-наборах по очищенному CSV, без пересоздания наборов/категорий:

```bash
kubectl exec -it deployment/spanish -n spanish -- \
  env LEARNING_PAIR=ru-es LEARNING_NATIVE_LANG=ru LEARNING_TARGET_LANG=es LEARNING_APP_CODE=spanish GRAMMAR_BUNDLE_ID=es \
  /app/import_word_sets_from_csv \
    --csv "/app/data/spanish_word_freq_pos_ud_top6000.csv" \
    --commit
```

После выполнения проверь выборочно 2-3 набора в админке: должно снова быть по 50 слов подряд по популярности (в рамках `preferred_pos`).

## Перевалидация и перегенерация текущих карточек (prod, без пересоздания word_cards)

Команда `revalidate_training_cards`:

- прогоняет существующие `training_cards` через текущую `ValidateTrainingCardResponse`;
- для `ru-es` дополнительно проверяет, что `definition_ru` в `word_cards` содержит кириллицу;
- ищет дубли `training_cards` внутри одного `word_card` (по `pos/display_word/word_ru/meaning_en`);
- проверяет TTS-статусы: `failed_*` и `ready` без реального аудиофайла;
- в `--commit` режиме для невалидных слов:
  - удаляет их `training_cards`,
  - сбрасывает у `word_cards` `processed_at/processing_error` (ставит обратно в очередь воркера),
  - если `definition_ru` не на русском — обнуляет `definition_ru`, чтобы воркер запросил поле заново.
- в `--commit` режиме для дублей:
  - удаляет только лишние `training_cards` (без reset `word_cards`, без перегенерации).
- в `--commit` режиме для невалидных TTS:
  - сбрасывает только проблемные записи `tts_generation_status` в `pending` (валидные TTS не трогаются).

Dry run:

```bash
kubectl exec -it deployment/spanish -n spanish -- \
  env LEARNING_PAIR=ru-es LEARNING_NATIVE_LANG=ru LEARNING_TARGET_LANG=es LEARNING_APP_CODE=spanish GRAMMAR_BUNDLE_ID=es \
  /app/revalidate_training_cards
```

Реальное применение:

```bash
kubectl exec -it deployment/spanish -n spanish -- \
  env LEARNING_PAIR=ru-es LEARNING_NATIVE_LANG=ru LEARNING_TARGET_LANG=es LEARNING_APP_CODE=spanish GRAMMAR_BUNDLE_ID=es \
  /app/revalidate_training_cards --commit
```

Точечный запуск:

```bash
kubectl exec -it deployment/spanish -n spanish -- \
  env LEARNING_PAIR=ru-es LEARNING_NATIVE_LANG=ru LEARNING_TARGET_LANG=es LEARNING_APP_CODE=spanish GRAMMAR_BUNDLE_ID=es \
  /app/revalidate_training_cards --only-word "puerto" --commit
```

Запуск без проверки TTS (только word/training):

```bash
kubectl exec -it deployment/spanish -n spanish -- \
  env LEARNING_PAIR=ru-es LEARNING_NATIVE_LANG=ru LEARNING_TARGET_LANG=es LEARNING_APP_CODE=spanish GRAMMAR_BUNDLE_ID=es \
  /app/revalidate_training_cards --check-tts=false --commit
```

## Локализация названий наборов/категорий на испанский (без пересоздания)

В код добавлена миграция `000004_localize_spanish_word_sets_to_es.sql`, которая обновляет существующие записи `word_set_categories` и `word_sets` in-place (ID и связи сохраняются).

Применение на prod:

1. Выкатить новый образ `spanish` с этой миграцией.
2. Перезапустить pod (или дождаться rollout) — миграции применяются при старте приложения автоматически.

Проверка:

```bash
kubectl logs -n spanish deployment/spanish --tail=200 | rg "applied sql migration"
```

Ожидается запись про версию:

- `000004_localize_spanish_word_sets_to_es`

