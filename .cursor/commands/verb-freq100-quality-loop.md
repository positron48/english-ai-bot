# verb-freq100-quality-loop

Итеративная проверка **качества** строк `GenerateVerbExamplePair` / каталога freq100 для **случайной леммы из топ‑100** (`Freq100VerbLemmas`): сэмпл → тесты → ревью → при необходимости правки → повтор до **PASS** или лимита итераций.

## Использование

`/verb-freq100-quality-loop`

Опционально в сообщении пользователя: `maxIterations=3`, «включить `ir`», «только presente+futuro».

Параметры по умолчанию: `.cursor/config.json` → секция **`verbQuality`**.

## Шаги (оркестратор выполняет сценарий; роли — только через mcp_task)

Инициализировать `iteration = 0`, прочитать `verbQuality.maxIterations` из `config.json` (если нет — **5**).

### Цикл (пока `iteration < maxIterations`)

1. **Сэмпл.** Вызвать **mcp_task** с `subagent_type="verb-freq100-sampler"`, `prompt`: кратко — «следующая случайная лемма (без `ir`, если не просили иначе); матрица indicativo: presente, imperfecto, pretérito, futuro; передай DefaultRuGloss; добавь seed для воспроизводимости». `description`: "Sample freq100 verb plan".

2. **Тесты.** Вызвать **mcp_task** с `subagent_type="verb-example-tester"`, `prompt`: вставить **полный** вывод шага 1 + «собери ES/RU для каждого слота». `description`: "Run verb example tests".

3. **Ревью.** Вызвать **mcp_task** с `subagent_type="verb-quality-reviewer"`, `prompt`: план сэмплера + артефакты тестера. `description`: "Review ES/RU quality".

4. Если в ответе ревьюера есть **`VERDICT: PASS`** — выйти из цикла, перейти к **Финальный отчёт**.

5. Если **`VERDICT: FAIL`** — вызвать **mcp_task** с `subagent_type="verb-quality-implementer"`, `prompt`: вставить вердикт и **нумерованные** замечания ревьюера. `description`: "Fix verb quality". Затем `iteration++` и **повторить с шага 2** (тот же план сэмпла и те же слоты, пока не потребуется новый сэмпл — оркестратор решает: обычно **один** лемма на весь цикл до PASS).

6. Если после шага 5 снова FAIL и `iteration >= maxIterations` — выйти из цикла с отчётом «лимит исчерпан».

### Финальный отчёт

Сообщить пользователю: лемма (и seed), число итераций, **PASS** или **лимит**, список коммитов/файлов (если был implementer).

## Результат

- Зафиксированное качество по чеклисту для одной случайной леммы или явный список оставшихся FAIL-пунктов.
- При успехе — подтверждение прохождения ревью и зелёного `make check-quick` после правок (если правки были).

## Заметки

- Не вызывать sampler и implementer параллельно.
- Таймаут **mcp_task** для шагов с `go test` / `make check-quick` — не меньше **10–15 минут** при полном прогоне; для быстрой итерации в промпте тестеру можно указать только `go test ./internal/spanishverbs/...`.
- Скилл с критериями: `@verb-freq100-quality` (или путь `.cursor/skills/verb-freq100-quality/SKILL.md`).
