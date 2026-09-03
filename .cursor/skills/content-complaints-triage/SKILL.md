---
name: content-complaints-triage
description: Полный триаж content_reports с prod EN/ES: reading_text, grammar_training, grammar_chapter, grammar_test и word_training; исправление исходников, bundle, проверка production и точечный resolve.
---

# Content complaints triage

Локальный агент без llama.cpp. Runbook: [reference.md](reference.md), журналы: [docs/complaints/README.md](../../../docs/complaints/README.md).

## Запуск

- «Запусти content-complaints-triage»: dry-run обоих курсов, затем конкретный список необходимых исправлений.
- «Исправляй», «apply»: исправить выявленные ошибки и аналогичные случаи, проверить и пересобрать материалы.
- Секреты: `secrets/complaints-prod.env` из `env.example.complaints-prod`.
- Не выводить токены и не запускать локальный kubectl: production-команды выполняет оператор согласно AGENTS.md.

## A — Полный снимок

Последовательно выполнить:

```bash
make complaints-triage-dry-en
make complaints-triage-dry-es
```

`fetch_reports.py` получает все страницы **без серверного фильтра course**, затем определяет курс по метаданным и локальному каталогу. Старый серверный фильтр отбрасывает чтение с ID `free_es_*`. Та же проблема была в summary: сводку строим из самого снимка.

- Проверить `complete`, `api_mode`, `unfiltered_report_count` и `unknown_course_report_ids`.
- Неизвестный курс остаётся видимым; его нельзя молча отнести к выбранному курсу или пропустить.
- Оба URL могут вести в общую БД. Дедуплицировать одинаковые report ID из одного сервиса, а не считать их двумя жалобами.
- Legacy API означает неполную выборку. Не сообщать «жалоб нет» или «все исправлено» по такому снимку.

## B — Кластеры и контекст

`reading_text` группировать по `payload.text_id` (fallback `grammar_chapter_id`), грамматику — по главе/блоку/вопросу. Не объединять разные тексты в одну пустую word-группу.

Загрузить `GET /api/internal/content-reports/{id}` для каждой жалобы. `payload` — исторический снимок. Для reading поле `reading_text` содержит **текущий документ production-БД**, `reading_text_found=false` — отсутствие текста; старые серверы этих полей не возвращают.

## C — Исправления

| Тип | Что проверять и исправлять |
|-----|---------------------------|
| word_training | Карточку через `PUT /api/internal/training/card/{id}`, TTS через regenerate + ожидание ready |
| grammar_training | Канонический `courses/*-grammar/training_pack/`, затем embedded training pack |
| grammar_chapter / grammar_test | Теорию или тест в канонических исходниках главы и соответствующий bundle |
| reading_text | `courses/*-grammar/reading/texts/{id}.json`, `reading/index.json`, `assets/reading/{id}` |

Для reading проверить весь текст, русский перевод, вопросы/объяснения/ключи, голоса и соответствие изображения каждому визуальному факту. При изменении озвучиваемого текста обновить tokens и аудио. Для заменённых изображений/аудио использовать новые пути, чтобы клиентский кеш не сохранял старую версию. При удалении материала убрать запись из индекса и её assets; проверить удаление из production-каталога после импорта.

## D — Поиск аналогичных ошибок и проверки

- Проверить весь релевантный каталог по найденному паттерну, а не только ID жалобы.
- Reading-вопросы и объяснения — на естественном русском; без «План связан с вернуть книгу», «один/одна» и генераторных концовок в тексте.
- `make -C courses/spanish-grammar reading-validate` (или соответствующий курс).
- `python3 -m unittest discover -s tools-local/complaints-triage/tests -v` при изменениях инструментов.
- После исходников курса: `./scripts/generate-grammar-bundle.sh es` и `./scripts/generate-grammar-training-pack.sh es` (либо en).
- `make check` после изменений приложения; отдельно сообщать фактические результаты и блокеры.

## E — Журнал и релиз

В apply создать `make complaints-journal-new`. В `docs/complaints/journal-YYYY-MM-DD-*.md` записывать для каждого ID: дату, суть, изменение, проверку и необходимость импорта. Добавить строку в индекс журналов.

Коммитить только изменения этого прогона, сохраняя посторонние изменения. Релиз исходников и bundle — по reference.md. Для unified Linglow после здорового rollout оператор выполняет:

```bash
kubectl -n linglow exec deploy/linglow -- /app/import_learning_content --course-code es_ru --commit
```

## F — Проверка production и точечное закрытие

Закрывать только конкретные ID после проверки, что исправления действительно выдаются из production. Наличие правки в git или bundle и старый `question_snapshot` этого не доказывают.

Скрипт с историческим именем `resolve_all_active.py` теперь требует явные ID и причину, по умолчанию показывает план:

```bash
set -a; . ./secrets/complaints-prod.env; set +a
python3 tools-local/complaints-triage/resolve_all_active.py es --report-ids 17,18 --reason 'Проверены текущие тексты и обложка после импорта'
# Только после фактической проверки:
python3 tools-local/complaints-triage/resolve_all_active.py es --report-ids 17,18 --reason 'Проверены текущие тексты и обложка после импорта' --apply
```

Не использовать массовое закрытие всех активных жалоб или старый `apply_prod_word_fixes.py` как универсальное исправление. Повторить fetch, зафиксировать оставшиеся ID и причины; новые жалобы не закрывать заодно. JSONL хранится локально с текущими датой, причиной и выбранными ID.

Канонический production URL для обоих курсов: `https://linglow.qantrix.ru`. Старые домены перенаправляют GET, но POST resolve через urllib завершается HTTP 308. В `COMPLAINTS_SERVICE_URL_EN` и `COMPLAINTS_SERVICE_URL_ES` задавать канонический URL. Health endpoint: `/health`.
