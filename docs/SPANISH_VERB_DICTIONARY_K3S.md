# Словарь спряжений испанских глаголов и выкат в k3s

**Единая страница для оператора:** образ → тег → миграции → `kubectl exec` (импорт Jehle) → переменные в `spanish-config` → проверка. Общий выкат инстанса Spanish: `docs/SPANISH_K3S_ROLLOUT_RUNBOOK.md`.

В образе приложения лежит CSV **Fred Jehle Spanish Verb Database** (репозиторий [ghidinelli/fred-jehle-spanish-verbs](https://github.com/ghidinelli/fred-jehle-spanish-verbs)), путь в контейнере:

- `/app/data/verbs/jehle_verb_database.csv`
- `/app/data/verbs/jehle_supplement_aux_haber.csv` — парадигма **haber** (не входит в Jehle; см. `SUPPLEMENT_HABER.txt`)
- `/app/data/verbs/ATTRIBUTION.txt` — атрибуция и SHA256; **лицензия CC BY-NC-SA 3.0** (есть ограничение NonCommercial).

Импорт в таблицы `verb_lemmas` / `verb_forms_dict` выполняется **отдельной командой** (не при старте пода): идемпотентный upsert, безопасно повторять. При upsert леммы **`metadata_json` не затирается пустым `{}`**: если импорт передаёт `{}`, сохраняется уже существующий JSON (чтобы не потерять `ru.gloss`, `verb_class`, шаблоны после бэкфиллов).

В Jehle CSV **нет** отдельной парадигмы для вспомогательного **haber**; в репозитории добавлен свой файл `resources/verbs/jehle_supplement_aux_haber.csv` (см. `resources/verbs/SUPPLEMENT_HABER.txt`). Цель `make import-spanish-verbs-jehle-bundled` импортирует Jehle и затем этот supplement.

## Локально (Postgres из `.env` / `.env.es`)

Цели `make import-spanish-verbs*`, `make backfill-word-verb-links`, `make backfill-verb-lemma-ru-glosses` подгружают **опционально** `.env`, затем **обязательно** `.env.es` (нужен `DATABASE_URL` на Spanish Postgres), как у `make backfill-noun-gender-es` и др.

- **`make build-verb-form-examples`** — сейчас **no-op** (совместимость Makefile): примеры для cloze-тренировки форм собираются **на лету** при выдаче карточки; таблица `verb_form_examples` для этого режима не обязательна.
- **`make backfill-verb-lemma-ru-glosses`** — **опционально**, отдельно: батч в LLM заполняет `verb_lemmas.metadata_json` → `ru.gloss` (русский глосс к инфинитиву) для литературной русской строки под примером. Нужны те же `DATABASE_URL` из `.env.es` и **`AI_URL` / `AI_API_KEY` / `AI_MODEL`** (как для остального AI). См. `make help` — передача флагов: `ARGS='--dry-run'`, `ARGS='--batch-size=35 --force'` и т.д. Флаг **`--fill-class`** — отдельный проход: только поле **`verb_class`** (`motion|speech|transfer|generic`), без глоссов.
- **`make backfill-verb-template-links`** — **опционально**, без LLM: офлайн-правила дописывают в `metadata_json` поля **`verb_class`** и **`allowed_template_ids`** для курируемых лемм (сейчас минимум `ir`), чтобы runtime cloze выбирал детерминированные шаблоны.
- Миграция **`000005_verb_example_catalog.sql`**: таблица **`verb_example_templates`** (опциональный каталог в БД; пустая таблица — только встроенный кодовый каталог). Кэш загрузки в процессе ~5 минут.
- **Топ-100 глаголов по частоте** (первые 100 лемм с `pos` **VERB** или **AUX** из `resources/wordsets/spanish_word_freq_pos_ud_top6000.csv`, список зашит в код как `Freq100VerbLemmas`): для **indicativo + presente** cloze использует компактные шаблоны с русской строкой вида `Ты: «говорить».`, если в `metadata_json` уже есть **`ru.gloss`** (или встроенный `DefaultRuGloss`). Лемма **`ir`** по-прежнему только «ходовые» шаблоны (домой, на работу и т.д.), без смешения с топ-100 паком.

После миграций (обычно уже применены при первом запуске бота):

```bash
make import-spanish-verbs-jehle-bundled
make backfill-word-verb-links
# опционально: русские глоссы к леммам (мало запросов к LLM, батчами)
make backfill-verb-lemma-ru-glosses ARGS='--dry-run'
make backfill-verb-lemma-ru-glosses
# опционально: офлайн привязка шаблонов к лемме ir (без LLM)
make backfill-verb-template-links
# опционально: no-op, можно не вызывать
make build-verb-form-examples
```

Либо вручную:

```bash
go run ./cmd/import_spanish_verbs \
  --input resources/verbs/jehle_verb_database.csv \
  --format jehle-csv \
  --source fred-jehle-ghidinelli \
  --source-version jehle-csv-sha256-f77f01d536cd351584051d76902ff8051ab1b945a38e69c7ed02da78ab082ea8
go run ./cmd/import_spanish_verbs \
  --input resources/verbs/jehle_supplement_aux_haber.csv \
  --format jehle-csv \
  --source project-supplement \
  --source-version haber-paradigm-v1
go run ./cmd/backfill_word_verb_links
# опционально (нужны AI_* в окружении):
go run ./cmd/backfill_verb_lemma_ru_glosses --batch-size=28 --sleep-ms=800
# опционально (no-op):
go run ./cmd/build_verb_form_examples
```

Включить фичу в окружении: `SPANISH_VERB_FORMS_ENABLED=true` (см. `internal/config/config.go`).

## k3s (namespace `spanish`)

GitOps-манифесты: репозиторий `devops-time-host`, каталог `apps/spanish/base/` (Deployment, ConfigMap, Postgres и т.д.).

### 1. Образ

Dockerfile собирает бинарники и кладёт данные в образ:

- `/app/import_spanish_verbs`
- `/app/backfill_word_verb_links`
- `/app/backfill_verb_lemma_ru_glosses`
- `/app/backfill_verb_template_links` — офлайн: `verb_class` + `allowed_template_ids` в `verb_lemmas.metadata_json` для курируемых лемм (см. `cmd/backfill_verb_template_links`)
- `/app/preview_verb_templates` — вывод всех пар **ES + RU** по текущим шаблонам для каждой строки `verb_forms_dict` выбранной леммы (см. `make preview-verb-templates`)
- `/app/build_verb_form_examples` (no-op)
- `/app/data/verbs/jehle_verb_database.csv`
- `/app/data/verbs/jehle_supplement_aux_haber.csv`

Релиз образа `ghcr.io/<owner>/spanish` — по **git tag** в `english-ai-bot` (см. `.github/workflows/ci.yml`, job `docker-image` с matrix `english` / `spanish`). Оба тега собираются **одним и тем же Dockerfile** (в образ `spanish` попадают те же утилиты и `/app/data/verbs/*`, что и для English).

### 2. Переменные в кластере (ConfigMap `spanish-config`)

Файл GitOps: `devops-time-host/apps/spanish/base/configmap.yaml` (имя объекта в кластере обычно `spanish-config`).

**Обязательно для режима «Формы» (тренировка спряжений + API):**

| Ключ | Рекомендуемое значение | Назначение |
|------|------------------------|------------|
| `SPANISH_VERB_FORMS_ENABLED` | `true` | Включает SRS/API для испанских форм глагола. |
| `VERB_FORMS_MAX_CARDS_PER_SESSION` | `30` (или опустить — дефолт в коде 30) | Максимум карточек за сессию. |
| `VERB_FORMS_MAX_NEW_PER_SESSION` | `30` | Максимум карточек в состоянии `new` за сессию. |
| `VERB_FORMS_TYPED_MIN_REPS` | `2` | После стольких успешных повторов без `learning` допускается ввод формы целиком. |
| `VERB_FORMS_TYPED_CHANCE_PERCENT` | `50` | Вероятность (0–100) показать ввод вместо вариантов, когда ввод уже допустим. |

Остальное (`LEARNING_TARGET_LANG=es`, `TRAINING_*`, TTS, `AI_URL` и т.д.) — как в общем runbook: `docs/SPANISH_K3S_ROLLOUT_RUNBOOK.md`, `devops-time-host/apps/spanish/RELEASE_K3S.md`. Секреты (`DATABASE_URL`, `AI_API_KEY`, …) — только в `spanish-secrets`, не в ConfigMap.

**Фрагмент для вставки в `data:` ConfigMap** (YAML, отступы как у соседних ключей):

```yaml
  SPANISH_VERB_FORMS_ENABLED: "true"
  VERB_FORMS_MAX_CARDS_PER_SESSION: "30"
  VERB_FORMS_MAX_NEW_PER_SESSION: "30"
  VERB_FORMS_TYPED_MIN_REPS: "2"
  VERB_FORMS_TYPED_CHANCE_PERCENT: "50"
```

После коммита в `devops-time-host` при необходимости: `flux reconcile kustomization …` (см. общий runbook §5).

### 2a. Проверка: что уже в образе из CI

Один Dockerfile для матрицы `english` / `spanish`; job только меняет имя образа в GHCR. В финальном образе есть:

- бинарники в `/app/`: `import_spanish_verbs`, `backfill_word_verb_links`, `backfill_verb_lemma_ru_glosses`, `backfill_verb_template_links`, `preview_verb_templates`, `build_verb_form_examples` (no-op);
- данные: `/app/data/verbs/jehle_verb_database.csv`, `jehle_supplement_aux_haber.csv`, `ATTRIBUTION.txt`, `SUPPLEMENT_HABER.txt`;
- миграции вшиты в бинарь бота и применяются при старте (`000002_verb_forms.sql`, `000005_verb_example_catalog.sql` и др.).

Отдельно собирать утилиты в CI не нужно — достаточно **tag push** и выката нового digest в `spanish` Deployment.

### 3. Заполнение БД одним блоком (рекомендуется после первого выката с миграциями)

Миграции `000002_verb_forms.sql` применяются при старте приложения. Затем один раз (или после обновления CSV в образе):

```bash
kubectl -n spanish exec -it deployment/spanish -- sh -c '
  /app/import_spanish_verbs \
    --input /app/data/verbs/jehle_verb_database.csv \
    --format jehle-csv \
    --source fred-jehle-ghidinelli \
    --source-version jehle-csv-sha256-f77f01d536cd351584051d76902ff8051ab1b945a38e69c7ed02da78ab082ea8 &&
  /app/import_spanish_verbs \
    --input /app/data/verbs/jehle_supplement_aux_haber.csv \
    --format jehle-csv \
    --source project-supplement \
    --source-version haber-paradigm-v1 &&
  /app/backfill_word_verb_links
'
```

#### 3a. Ошибка `sh: ... import_spanish_verbs: not found` (exit 127)

Чаще всего в кластере крутится **старый digest** образа `ghcr.io/.../spanish`: в GitOps у `Deployment` часто закреплён `image: ...@sha256:...` через Flux ImagePolicy. Если этот образ собран **до** появления в Dockerfile стадий `import_spanish_verbs` / `COPY` CSV, бинарника в слое просто нет — `not found` не из-за каталога, а из-за отсутствия файла.

Проверка в том же поде:

```bash
kubectl -n spanish exec deployment/spanish -- ls -la /app/import_spanish_verbs /app/data/verbs/jehle_verb_database.csv
```

Если `No such file` — нужен **новый релиз** `english-ai-bot` (git **tag** → CI публикует `spanish:latest` + digest), затем дождаться коммита Flux с новым digest в `devops-time-host` и `rollout status` для `deployment/spanish`. После этого снова блок из §3.

Опционально в том же поде (если в секрете/окружении пода доступны `AI_URL`, `AI_API_KEY`, `AI_MODEL` — как у основного бота):

```bash
kubectl -n spanish exec -it deployment/spanish -- /app/backfill_verb_lemma_ru_glosses --batch-size=28 --sleep-ms=800
```

Опционально **после** глоссов (или параллельно по смыслу данных):

```bash
kubectl -n spanish exec -it deployment/spanish -- /app/backfill_verb_template_links
# при необходимости LLM-классификация хвоста лемм (нужны AI_*):
kubectl -n spanish exec -it deployment/spanish -- /app/backfill_verb_lemma_ru_glosses --fill-class --batch-size=28 --sleep-ms=800
```

Просмотр **всех** сгенерированных примеров ES/RU по лемме (как в рантайме тренировки, столбец `source`: `catalog` или `generic`):

```bash
kubectl -n spanish exec -it deployment/spanish -- /app/preview_verb_templates -lemma=hablar
# первые 20 строк: -max=20 ; машиночитаемый вывод: -tsv
```

`/app/build_verb_form_examples` вызывать не обязательно (no-op).

Проверка логов: в stdout — счётчики импорта и бэкфилла ссылок; для `backfill_verb_lemma_ru_glosses` — число обновлённых строк `verb_lemmas`.

### 4. Порядок в жизненном цикле

1. Выкатить новый образ с миграциями и утилитами.
2. Дождаться готовности пода (`rollout status`).
3. Выполнить блок `kubectl exec` выше (при необходимости — `backfill_verb_lemma_ru_glosses`, затем при желании **`backfill_verb_template_links`** / **`--fill-class`**).
4. Убедиться, что `SPANISH_VERB_FORMS_ENABLED=true` (ConfigMap + reconcile Flux при необходимости).

### 5. Обновление только данных словаря

Если в следующем релизе обновили `resources/verbs/jehle_verb_database.csv` и `--source-version`, повторный запуск `import_spanish_verbs` обновит строки по upsert; затем снова `backfill_word_verb_links`. При появлении **новых** лемм в словаре имеет смысл снова запустить **`backfill_verb_lemma_ru_glosses`** (без `--force` он пропустит леммы, у которых уже есть `ru.gloss`). `build_verb_form_examples` для тренировки cloze не требуется.

### 6. Альтернатива без данных в образе

Можно смонтировать CSV через ConfigMap/Secret (до ~1 MiB ограничение ConfigMap) или отдельный volume; тогда путь передать в `--input`. Для полного Jehle CSV разумнее держать файл в образе (как сейчас).
