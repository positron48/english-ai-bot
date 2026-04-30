# Complaints Improve Loop: Quickstart

## Основная команда

```bash
make complaints-improve-loop-both
```

## Что делает команда

Автономно запускает цикл улучшения prompt для EN+ES:

1. применяет жалобы (удаление проблемных вопросов + `resolve` в API),
2. строит improvement plan по журналу,
3. итеративно (до 3 раз) правит prompt и прогоняет strict-проверки,
4. при `OK` запускает догенерацию только затронутых глав,
5. делает финальный smoke,
6. пишет success/failure отчет.

## Составляющие шаги и где смотреть

- **Apply жалоб**
  - шаг: `make complaints-apply-both`
  - смотреть:
    - `logs/complaints/complaints-YYYY-MM.jsonl`
    - `logs/complaints/changed-theory-blocks-YYYYMMDDHH.json`
- **План улучшений**
  - шаг: `make complaints-plan-both`
  - смотреть:
    - `logs/complaints/improvement-plan-*.json`
    - `logs/complaints/improvement-plan-*.md`
- **Итеративный prompt update + strict gate**
  - шаги внутри loop:
    - `make complaints-prompt-autofix-both`
    - `make complaints-prompt-regression`
    - `make complaints-smoke-both`
  - смотреть:
    - `logs/complaints/iteration-feedback-*.json` (если итерация неуспешна)
- **Таргетированная догенерация (только при успехе итерации)**
  - шаг: `make complaints-regenerate-affected`
  - смотреть:
    - `courses/english-grammar/training_pack/reports/validation-report.json`
    - `courses/spanish-grammar/training_pack/reports/validation-report.json`
- **Итог цикла**
  - успех:
    - `logs/complaints/improve-loop-success-*.json`
    - `logs/complaints/improve-loop-success-*.md`
  - провал после 3 итераций:
    - `logs/complaints/improve-loop-failure-*.json`
    - `logs/complaints/improve-loop-failure-*.md`

## Как ловить деградацию prompt

- Один раз зафиксировать baseline:
  - `make complaints-quality-baseline-both`
  - файл baseline: `logs/complaints/quality-baseline.json`
- Регулярная проверка на деградацию:
  - `make complaints-quality-both`
  - отчет: `logs/complaints/quality-regression-*.json`
  - если метрики хуже baseline выше порога, команда завершится с ошибкой (`exit 2`).