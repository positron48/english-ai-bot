# Локальный backfill `noun_gender` и `opposite_gender_word` для Spanish (`.env.es`)

Этот документ описывает локальный прогон скрипта `cmd/backfill_noun_gender` для уже существующих слов в БД Spanish.

## Что делает скрипт

- Обрабатывает карточки, где:
  - `noun_gender` пустой/невалидный, или
  - `opposite_gender_word` ещё не заполнен.
- Для карточек с POS `noun/sustantivo`:
  - сначала берёт данные из `spanish_gender_lexicon.tsv`,
  - если записи нет в словаре — идёт в AI fallback.
- Для карточек с POS НЕ `noun/sustantivo`:
  - берёт данные **только из `spanish_gender_lexicon.tsv`**,
  - в AI **не идёт**.
- Для `target_lang = es` в noun-ветке AI определяет:
  - `noun_gender` (`m|f|mf|n`)
  - `opposite_gender_word` (если естественная парная форма есть, иначе пусто).
- Для `target_lang = en` ставит `n` (без AI-вызова).

## Подготовка

1. Убедиться, что есть `.env.es`:

```bash
cp env.example.es .env.es
```

2. Заполнить в `.env.es`/`.env` минимум:
- `DATABASE_URL`
- `AI_URL`
- `AI_API_KEY`
- `AI_MODEL` (или модель из дефолта, если у вас так принято)
- `LEARNING_TARGET_LANG=es`

Важно: локальные `make`-таргеты ниже автоматически импортируют:
- сначала опциональный `.env`,
- затем обязательный `.env.es` (перекрывает значения из `.env`).

## Рекомендуемый запуск (через Makefile)

### 0) (Рекомендуется) сначала нормализовать POS у старых карточек

Это нужно для кейсов вроде:
- `sustantivo femenino`
- `sustantivo masculino diminutivo / apelativo afectivo`

Также этот шаг очищает явные ложные `opposite_gender_word` (например `amor -> amora`),
оставляя только безопасные пары вида `-o <-> -a` (например `hermano <-> hermana`).

Dry-run:

```bash
make normalize-word-pos-es-dry
```

Боевой запуск:

```bash
make normalize-word-pos-es
```

С ограничениями:

```bash
BATCH=200 LIMIT=5000 make normalize-word-pos-es
```

### 1) Dry-run (без записи в БД)

```bash
make backfill-noun-gender-es-dry
```

С ограничениями:

```bash
BATCH=50 LIMIT=200 make backfill-noun-gender-es-dry
```

### 2) Боевой запуск (с записью в БД)

```bash
make backfill-noun-gender-es
```

С ограничениями:

```bash
BATCH=50 LIMIT=2000 make backfill-noun-gender-es
```

## Альтернатива: прямой запуск бинарника с ручным импортом `.env.es`

```bash
set -a; [ -f .env ] && . ./.env; set +a
set -a && . ./.env.es && set +a
go run ./cmd/backfill_noun_gender -dry-run=true -batch 100 -limit 200
```

Боевой вариант:

```bash
set -a; [ -f .env ] && . ./.env; set +a
set -a && . ./.env.es && set +a
go run ./cmd/backfill_noun_gender -dry-run=false -batch 100 -limit 0
```

## K3s: запуск внутри pod (после выката нового образа)

После выката образа команды можно запускать прямо в pod Spanish-инстанса.

1) Dry-run:

```bash
kubectl exec -it deployment/spanish -n spanish -- /app/normalize_word_pos -dry-run=true -batch 200 -limit 0
kubectl exec -it deployment/spanish -n spanish -- /app/backfill_noun_gender -dry-run=true -batch 100 -limit 0
```

2) Боевой запуск:

```bash
kubectl exec -it deployment/spanish -n spanish -- /app/normalize_word_pos -dry-run=false -batch 200 -limit 0
kubectl exec -it deployment/spanish -n spanish -- /app/backfill_noun_gender -dry-run=false -batch 100 -limit 0
```

Примечания:
- `DATABASE_URL`, `LEARNING_TARGET_LANG`, `AI_*` берутся из env деплоймента.
- Словарь в образе доступен по пути `/app/data/spanish_gender_lexicon.tsv`.

## Проверка результата (SQL)

Осталось существительных без валидного `noun_gender`:

```sql
SELECT COUNT(*)
FROM word_cards
WHERE (
    LOWER(COALESCE(pos, '')) = 'noun'
    OR LOWER(COALESCE(pos, '')) LIKE 'noun %'
    OR LOWER(COALESCE(pos, '')) = 'sustantivo'
    OR LOWER(COALESCE(pos, '')) LIKE 'sustantivo %'
  )
  AND LOWER(COALESCE(noun_gender, '')) NOT IN ('m','f','mf','n');
```

Распределение по значениям `noun_gender`:

```sql
SELECT LOWER(COALESCE(noun_gender, '')) AS noun_gender, COUNT(*)
FROM word_cards
WHERE (
    LOWER(COALESCE(pos, '')) = 'noun'
    OR LOWER(COALESCE(pos, '')) LIKE 'noun %'
    OR LOWER(COALESCE(pos, '')) = 'sustantivo'
    OR LOWER(COALESCE(pos, '')) LIKE 'sustantivo %'
  )
GROUP BY 1
ORDER BY 2 DESC;
```

Сколько осталось "нечистых" POS (не канонические токены):

```sql
SELECT COUNT(*)
FROM word_cards
WHERE COALESCE(TRIM(pos), '') <> ''
  AND LOWER(TRIM(pos)) NOT IN (
    'noun','verb','adjective','adverb','pronoun','preposition','conjunction','interjection','article'
  );
```

Сколько существительных без заполненного `opposite_gender_word`:

```sql
SELECT COUNT(*)
FROM word_cards
WHERE (
    LOWER(COALESCE(pos, '')) = 'noun'
    OR LOWER(COALESCE(pos, '')) LIKE 'noun %'
    OR LOWER(COALESCE(pos, '')) = 'sustantivo'
    OR LOWER(COALESCE(pos, '')) LIKE 'sustantivo %'
  )
  AND COALESCE(TRIM(opposite_gender_word), '') = '';
```

