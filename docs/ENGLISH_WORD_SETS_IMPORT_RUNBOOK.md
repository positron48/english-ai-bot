# English Word Sets: zero-touch bootstrap from frequency dictionary

Этот runbook описывает, как English-наборы слов формируются из `wordFrequency.ods` без ручных команд на сервере.

## Что делает пайплайн

1. `scripts/prepare_english_frequency_csv.py`:
   - конвертирует `wordFrequency.ods` (лист `1 lemmas`) в CSV;
   - мапит POS в UD (`NOUN/VERB/ADJ/ADV`);
   - применяет rule-based очистку;
   - применяет LLM-фильтрацию подозрительных лемм;
   - сохраняет финальный CSV и JSON-отчёт.
2. При старте English-инстанса (`LEARNING_APP_CODE=english`, `LEARNING_TARGET_LANG=en`) выполняется `AutoSyncEnglishWordSets`:
   - проверяет checksum CSV;
   - при изменении checksum синхронизирует `word_set_items` через `import_word_sets_from_csv`-логику;
   - сохраняет checksum/summary в `app_settings`.

## Источники и артефакты

- Исходник: `wordFrequency.ods`
- Финальный CSV для импорта: `resources/wordsets/english_word_freq_pos_ud_top6000.filtered.csv`
- Отчёт очистки: `resources/wordsets/english_word_freq_pos_ud_top6000.report.json`
- Блоклист: `resources/wordsets/english_lemma_blocklist.txt`

## Сетка English наборов (fixed blueprint)

- `Core Frequency`
  - `Top 500 Verbs` -> `Core Verbs — Top 50 (Ranks 1–50 ... 451–500)`
  - `Top 500 Nouns` -> `Core Nouns — Top 50 (Ranks 1–50 ... 451–500)`
  - `Top 500 Adjectives` -> `Core Adjectives — Top 50 (Ranks 1–50 ... 451–500)`
  - `Top Adverbs` -> `Core Adverbs — Top 50 (Ranks 1–50 ... 251–300)`
  - `Top 500–1000 Nouns` -> `Core Nouns — Top 50 (Ranks 501–550 ... 951–1000)`
- `Extended Frequency`
  - `Top 1001–1500 Nouns` -> `Extended Nouns — Top 50 (Ranks 1001–1050 ... 1451–1500)`
  - `Top 1501–2000 Nouns` -> `Extended Nouns — Top 50 (Ranks 1501–1550 ... 1951–2000)`

## Локальная подготовка данных (перед коммитом)

```bash
make prepare-english-csv
```

Если нужно пропустить LLM-шаг (например, офлайн), можно запустить напрямую:

```bash
python3 scripts/prepare_english_frequency_csv.py --skip-llm
```

## Безопасная ручная проверка (локально)

Хотя prod использует zero-touch bootstrap, для отладки можно сделать preflight/dry-run/manual commit:

1. **Preflight**: проверь профиль English (`.env.en`) и что CSV путь English.
2. **Dry-run**:

```bash
set -a; [ -f .env ] && . ./.env; set +a
set -a && . ./.env.en && set +a
LEARNING_PAIR=ru-en LEARNING_NATIVE_LANG=ru LEARNING_TARGET_LANG=en LEARNING_APP_CODE=english GRAMMAR_BUNDLE_ID=en \
go run ./cmd/import_word_sets_from_csv --lang en --csv resources/wordsets/english_word_freq_pos_ud_top6000.filtered.csv
```

3. **Commit** (только после dry-run):

```bash
make sync-english-word-sets
```

## Zero-touch deploy

После деплоя нового образа в English k3s-под:

- ручные `kubectl exec` для импорта наборов не нужны;
- bootstrap сам синхронизирует наборы один раз на новый checksum CSV;
- при неизменном CSV bootstrap логирует skip.

## Логи и проверки

Ищи в логах English-пода:

- `english word sets bootstrap completed`
- или `english word sets bootstrap skipped: csv checksum unchanged`

Также можно проверить `app_settings` ключи:

- `word_sets.english.csv_sha256`
- `word_sets.english.last_import_summary`

## Guardrails (защита от смешивания языков)

`cmd/import_word_sets_from_csv` теперь:

- требует согласованный профиль:
  - `--lang en` -> `LEARNING_TARGET_LANG=en` и `LEARNING_APP_CODE=english`;
  - `--lang es` -> `LEARNING_TARGET_LANG=es` и `LEARNING_APP_CODE=spanish`;
- отказывает в запуске при несоответствующем CSV пути (english/spanish).

