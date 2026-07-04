---
description: Ревью и рерайт тестовых вопросов курса грамматики (батч глав параллельными субагентами)
argument-hint: [course=spanish] [batch=3]
---

Обработай ОДИН батч глав курса грамматики: параллельные субагенты ревьюят вопросы по рубрике, ты валидируешь и коммитишь. Скилл рассчитан на запуск в цикле (`/loop /fix-course-questions`) — за вызов ровно один батч.

Аргументы: `$ARGUMENTS`. По умолчанию course=spanish (директория `courses/spanish-grammar`), batch=3.

## Шаги

1. Перейди в сабмодуль курса (`courses/<course>-grammar`). Если `review-questions/status.json` нет — запусти `./review-questions/init-status.sh`.
2. Прочитай `review-questions/status.json`, возьми первые `batch` глав со `status: "pending"` (в порядке ключей). Если есть застрявшие `in_progress` от упавших прогонов — сообщи о них пользователю, но не трогай.
   - **Если pending-глав нет** — сообщи «все главы обработаны, цикл можно останавливать» и заверши без ScheduleWakeup (если работаешь в /loop — не планируй следующую итерацию).
3. Пометь выбранные главы `"in_progress"` в status.json (через jq, сохраняя порядок ключей: `to_entries`-подход не нужен, обычный `.chapters["<dir>"].status = "in_progress"` порядок не меняет).
4. Спавни по одному субагенту (general-purpose) на главу **параллельно, в одном сообщении**. Промпт каждому:

   > Прочитай файл `<abs path>/courses/<course>-grammar/review-questions/RUBRIC.md` и выполни его инструкции для главы `chapters/<dirname>`. Работай из корня сабмодуля. НЕ трогай status.json и НЕ делай git-коммитов — это сделает оркестратор. В финальном ответе верни: rewritten=<N изменённых вопросов>, notes=<1 строка о типовых проблемах>, validation=<passed/failed>.

5. Когда все субагенты завершились, для каждой главы:
   - Проверь валидацию сам: `jq . chapters/<dir>/03-questions.json >/dev/null` и `./scripts/validate-chapter.sh <dir>` (агент уже пересобрал 05-final; если нет — сначала `./scripts/assemble-chapter.sh <dir>`).
   - **Успех:** обнови status.json: `status: "done"`, `model` (модель субагента), `reviewed_at` (ISO-дата), `rewritten`, `notes`. Затем отдельный коммит в сабмодуле: `git add chapters/<dir> review-questions/status.json && git commit -m "review(questions): <NNN> <slug> — <N> rewritten"`.
   - **Провал (невалидный JSON/схема, агент упал):** верни `status: "pending"` с `notes` о причине, `git checkout -- chapters/<dir>` чтобы откатить порченые правки, не коммить.
6. Итоговый отчёт пользователю: какие главы обработаны (номер, слог, сколько вопросов переписано, ключевые проблемы), какие упали и почему, сколько глав осталось pending.

## Ограничения

- Правится только `03-questions.json` (+ пересборка `05-final.json` скриптом). `training_pack/`, `reading/`, `speaking/`, theory blocks — не трогать.
- Коммиты только внутри сабмодуля; указатель сабмодуля в родительском репо не коммитить.
- Финальную регенерацию (`make final-all && make validate-all` в сабмодуле, `make grammar-bundle` в корне) НЕ запускать — это делает пользователь в конце.
