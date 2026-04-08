# Spanish Gender Lexicon (offline fallback)

Этот файл нужен как надежный офлайн-источник для:
- `gender` (`m|f|mf|n`)
- `article` (`el|la|el/la|lo`)
- `opposite_gender_word` (если есть)

Файл:
- `resources/wordsets/spanish_gender_lexicon.tsv`

Формат:
- `lemma\tgender\tarticle\topposite_gender_word\tsource\tnotes`

## Источник данных

- Датасет: `doozan/spanish_data` (`es-en.data`)
- URL: [raw es-en.data](https://raw.githubusercontent.com/doozan/spanish_data/master/es-en.data)
- Лицензия источника: см. `doozan/spanish_data` (CC-BY-SA / Wiktionary attribution)

Почему выбран этот источник:
- машиночитаемый, стабильный, открытый;
- содержит род (`g`) и `es-noun` шаблоны с `m=`/`f=` для части слов;
- позволяет получить большой покрывающий словарь для fallback без LLM.

## Как пересобрать файл

```bash
python3 scripts/build_spanish_gender_lexicon.py --download
```

или из уже скачанного локального файла:

```bash
python3 scripts/build_spanish_gender_lexicon.py \
  --input tmp/es-en.data \
  --output resources/wordsets/spanish_gender_lexicon.tsv
```

## Примеры

- `actor -> m, el, actriz`
- `actriz -> f, la, actor`
- `hermano -> m, el, hermana`
- `hermana -> f, la, hermano`
- `amor -> m, el, (empty opposite)`
- `problema -> m, el, (empty opposite)`

## Ограничения

- `opposite_gender_word` заполняется только когда:
  - в источнике есть явная пара (`m=...` / `f=...`), или
  - шаблонный маркер `+` можно безопасно развернуть в простую `-o <-> -a` пару.
- Нерегулярные пары, которых нет в источнике, останутся пустыми и должны идти в fallback (LLM/ручная правка).

## Рекомендованный порядок в рантайме

1. Lookup в `spanish_gender_lexicon.tsv`.
2. Если `lemma` не найден или `opposite` пустой и это важно — fallback в LLM.
3. Результат fallback можно дописывать в отдельный override-файл проекта.

Текущий runtime в проекте уже подключен именно так:
- `WordService` (новые слова): словарь -> fallback в LLM.
- `cmd/backfill_noun_gender`: словарь -> fallback в LLM.

Путь словаря по умолчанию:
- `resources/wordsets/spanish_gender_lexicon.tsv` (локально)
- `data/spanish_gender_lexicon.tsv` (в контейнере)

Можно переопределить через env:
- `SPANISH_GENDER_LEXICON_PATH=/absolute/or/relative/path.tsv`

