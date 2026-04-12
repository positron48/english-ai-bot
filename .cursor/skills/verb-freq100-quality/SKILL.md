---
name: verb-freq100-quality
description: Критерии качества ES/RU для топ-100 испанских глаголов и ссылки на код
---

# Качество примеров: freq100 + русское спряжение

Используй при ревью и доработках (`verb-quality-reviewer`, `verb-quality-implementer`, оркестратор `/verb-freq100-quality-loop`).

## Источники истины в коде

- Список лемм: `internal/spanishverbs/core100_verbs_list.go` — `Freq100VerbLemmas`.
- Глоссы по умолчанию: `internal/spanishverbs/lemma_ru_defaults.go` — `DefaultRuGloss`.
- Каталог шаблонов freq100: `internal/spanishverbs/template_catalog_freq100.go` (32 шаблона: 8 рамок × 4 времени indicativo).
- RU спряжение (OpenCorpora через gomorphy): `internal/spanishverbs/russian_gomorphy_conjugation.go`.
- Сборка пары: `internal/spanishverbs/template_catalog.go` — `TryGenerateCatalogPair`; fallback — `internal/spanishverbs/dynamic_example_pair.go` — `GenerateVerbExamplePair`.
- Отдельно **движение `ir`**: `internal/spanishverbs/template_catalog_ir.go`, `russian_idti_conjugation.go` — не смешивать с fp100.

## Что считать «хорошим»

1. Для каталога с **`RuSecond: gloss`**: в русской строке во втором слоте — **форма глагола** (не сырой глосс-инфинитив), согласованная с лицом/числом испанского субъекта и выбранным **tense** (где морфология это позволяет).
2. **Будущее** русского: аналитическое `буду/будешь/… + инфинитив` — ожидаемо.
3. **Прошедшее** ед.ч.: по умолчанию мужской род в тегах OpenCorpora — приемлемо для учебной строки «Он/Ты …».
4. Глосс с пояснением в скобках или запятой — инфинитив извлекается предсказуемо (`RussianInfinitiveFromRuGloss`).

## Известные ограничения (не путать с регрессией)

- **Один** русский инфинитив не различает испанский imperfecto и pretérito по аспекту: обе временные формы могут дать **одну** impf past форму леммы.
- Лемма **`быть`**: полноценное настоящее в словаре в основном **3 ед.**; остальные лица — откат к инфинитиву в слоте каталога, пока не добавлен отдельный хелпер.
- Лицензия данных gomorphy/OpenCorpora — см. комментарий в `russian_gomorphy_conjugation.go`.

## Проверки

- `go test ./internal/spanishverbs/... -count=1`
- После изменений в пакете: **`make check-quick`** в корне репозитория (минимум для итерации роя).

## Превью в k8s / CLI

- `make preview-verb-templates` и `docs/SPANISH_VERB_DICTIONARY_K3S.md` — по необходимости для оператора, не обязательно в каждом цикле сэмплера.
