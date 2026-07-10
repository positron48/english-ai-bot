# Журнал триажа content reports

**Прогон:** 2026-07-10 (prod EN `qantrix.ru`, ES `es.qantrix.ru`)  
**Run ID:** `triage-2026-07-10`  
**Снимок:** `logs/complaints/snapshot-en-20260710T125421Z.json` (3), `snapshot-es-20260710T125422Z.json` (10)  
**Закрыто на prod:** resolve-bulk 3 EN + 7 ES (причина: RU hints на карточках + правки training_pack; грамматика — после tag/import)  
**Тег:** _(заполнить после `make tag` + Flux rollout)_  
**Деплой грамматики:** `kubectl -n spanish exec deploy/spanish -- /app/import_learning_content --commit` после выката образа.

Формат блока: **дата жалобы** → **суть** → **что изменено**.

---

## English (word_training)

### #5 — 2026-06-25 · preferir (card 12430)

**Жалоба:** «Нет подсказки».

**Изменено:** `PUT /api/internal/training/card/12430` на EN+ES — `hint`: «Выбирать одно вместо другого.» (ранее подсказка была только на испанском). Карточка `course_code=es_ru` в EN-инстансе — известное загрязнение unified merge; отдельный backfill не в этом прогоне.

---

### #4 — 2026-06-25 · parar (card 12557)

**Жалоба:** «Нет подсказки».

**Изменено:** `PUT` card 12557 — `hint`: «Остановить движение или действие.» (EN+ES prod).

---

### #3 — 2026-06-25 · parecer (card 6661)

**Жалоба:** «Нет подсказки».

**Изменено:** `PUT` card 6661 — `hint`: «Выражать мнение: «кажется, что…».»; `word_ru`: **казаться** (было «считать»). EN+ES prod.

---

## Spanish (grammar_training)

### #12 — 2026-07-09 · numbers b6 q9

**Жалоба:** «Libro же мужского, объяснение кривое» — *pocas libros* как ключ для «мало книг».

**Изменено:** `correct_answer` **d → a** (*pocos libros*); пояснение про муж. род *libro/libros*. Файл: `courses/spanish-grammar/training_pack/chapters/012/es.06.b6_theory_strategy_quantity_questions_traps.questions.json`. Pattern sweep: **q8** в том же файле — prompt/варианты для *treinta y un libro* (муж. род), было «женский род» + *libros*.

---

### #11 — 2026-07-09 · habits b1 q6

**Жалоба:** «Непонятно почему cada día или todos не подходит».

**Изменено:** prompt уточнён: «использует маркер **de lunes a viernes**»; пояснение без отклонения валидных маркеров *todos los días* / *cada día* как «ошибочных». Файл: `.../034/es.01.b1_theory_present_for_routines_and_habits.questions.json`.

---

### #9 — 2026-07-05 · -ar endings b2 q8

**Жалоба:** «2 варианта вообще к вопросу не относится» (*comes*, *cantas*).

**Изменено:** дистракторы заменены на *bailamos* / *bailan* (другие лица *bailar*). Файл: `.../028/es.02.b2_theory_present_regular_ar_endings.questions.json`.

---

### #8 — 2026-07-05 · pretérito perfecto b4 q10

**Жалоба:** «Несколько неправильных».

**Изменено:** исправлено пояснение: *ya no* = «больше не», ключ (d) остаётся ошибочной конструкцией для рамки *este mes*. Файл: `.../046/es.04.b4_theory_pp_unfinished_periods_with_ya_micro_combo_a2.questions.json`.

---

### #2 — 2026-06-19 · sequencing b3 q4

**Жалоба:** «Тут русское слово в ответе» — *домой* в варианте (c).

**Изменено:** `voy домой` → `voy a casa`. Pattern sweep в том же блоке: q5 — `на работу` → `al trabajo`. Файл: `.../034/es.03.b3_theory_sequencing_daily_actions.questions.json`.

---

### #1 — 2026-06-19 · participles b4 q7

**Жалоба:** «Несколько правильных вариантов» (*roto*, *vuelto* как слова).

**Изменено:** prompt: «причастие от глагола **morir**»; дистракторы — только неверные формы *morir* (*morido*, *moriendo*, *morído*). Файл: `.../044/es.04.b4_theory_participio_irregular_roto_muerto_vuelto.questions.json`.

---

## Spanish (grammar_chapter)

### #6 — 2026-06-27 · alphabet chapter

**Жалоба:** «тест тг уведомлений».

**Изменено:** без правки контента; закрыто как тестовая жалоба.

---

## Spanish (word_training)

Жалобы #3–#5 — те же карточки, что EN (см. выше); исправления применены на ES prod.

---

## Правила / pattern sweep

- `libro`/`libros` — муж. род: исправлены q8–q9 в `es.06.b6_theory_strategy_quantity_questions_traps.questions.json`.
- Русский в `choices` grammar ES: q4–q5 `es.03.b3_theory_sequencing_daily_actions.questions.json`.
- Bundle: `./scripts/generate-grammar-training-pack.sh es` → `internal/grammartrainingpack/es/...`.

---

## Артефакты прогона

| Файл | Назначение |
|------|------------|
| `docs/complaints/journal-2026-07-10-triage.md` | этот журнал (в git) |
| `logs/complaints/triage-2026-07.jsonl` | машинный журнал resolve (локально) |
| `logs/complaints/snapshot-*-20260710T125421Z.json` | снимок жалоб с prod |
