---
name: multilang-spanish-plan
description: Этапы A–F плана multilang/Spanish, English-first чеклист, ссылки на документы
---

# Multilang / Spanish — контекст этапов

Источник истины по реализации: `docs/MULTILANG_SPANISH_LAUNCH_PLAN.md`. Инфраструктура k3s после кода: `docs/SPANISH_K3S_ROLLOUT_CHECKLIST.md`.

## Приоритет №1

Не ломать текущий **RU→EN** и прод English: дефолтная конфигурация, SRS, существующие API и поведение должны сохраняться, если план не требует иного явно.

## Этапы A–F (кратко)

| Этап | Фокус |
|------|--------|
| **A** | `LearningConfig` в config, env, валидация в startup, тесты, `make check` |
| **B** | Нейтральные alias в моделях/DTO/репозиториях, SQL, тесты репозиториев |
| **C** | `LearningConfig` в сервисах (Word, Training, WordSet, Grammar, Pronunciation), SRS без смены алгоритма |
| **D** | Prompt placeholders (`{{native_lang}}`, `{{target_lang}}`, `{{pair}}`), файлы `prompts/*-ru-en.txt` / `*-ru-es.txt`, тесты парсеров LLM |
| **E** | Grammar bundle `en/` vs `es/`, `GrammarContentRepository`, тесты выбора bundle |
| **F** | API/UI нейтральные имена + backward compatibility, UI без hardcoded «English» |

## Проверки

Перед завершением этапа в этом репозитории обязателен полный **`make check`** (см. `AGENTS.md` в родительском workspace).

## Покрытие

Процент из `make check` и `COVER_PKGS`; исключения `internal/integration/`, `internal/testutil` — как в Makefile. Цель не «100% везде», а отсутствие регрессии и адекватное покрытие новых зон.
