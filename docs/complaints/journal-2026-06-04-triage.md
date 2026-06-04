# Журнал триажа content reports

**Прогон:** 2026-06-04 (prod EN `qantrix.ru`, ES `es.qantrix.ru`)  
**Run ID:** `triage-2026-06-04`  
**Снимок:** `snapshot-en-20260604T142003Z.json` (7), `snapshot-es-20260604T142007Z.json` (26)  
**Закрыто на prod:** resolve-bulk 7 EN + 26 ES (причина: см. JSONL `triage-2026-06.jsonl`)  
**Тег:** `0.11.13` (english-ai-bot `25932484`; submodules `english-grammar@4e209a0`, `spanish-grammar@7f9fdd1`)  
**Деплой грамматики:** после Flux — `import_learning_content --commit` в english/spanish.

Формат блока: дата жалобы → суть → что изменено.

---

## English (grammar_training)

### #7 — 2026-05-02

**Жалоба:** «Что значит какое время описывает событие, непонятно» — вопрос про фон vs событие в предложении с Past Continuous / Past Simple (`b1_theory_background_vs_event::q7`).

**Изменено:** уточнён prompt: явно просим выбрать фрагмент **короткого события (Past Simple)**, а не фон. Файл: `courses/english-grammar/training_pack/chapters/043/en.01.b1_theory_background_vs_event.questions.json`.

---

### #6 — 2026-05-02

**Жалоба:** «I am тоже верный вариант» — план встречи в субботу (`b1_theory_scenario_weekend::q1`).

**Изменено:** удалён вариант (d) *I am meeting…* как дубликат (b) *I'm meeting…*; обновлено пояснение. Файл: `courses/english-grammar/training_pack/chapters/051/en.01.b1_theory_scenario_weekend.questions.json`.

---

### #5 — 2026-04-30

**Жалоба:** (без текста) — вопрос «в каком предложении пропущено than», но ключ указывал на вариант **с** than (`b4_than_the::q5`).

**Изменено:** `correct_answer` **a → b**; пояснение: ошибка — вариант без *than*. Файл: `courses/english-grammar/training_pack/chapters/069/en.04.b4_than_the.questions.json`.

---

### #4 — 2026-04-30

**Жалоба:** (без текста) — *present* после существительного (`b4_after_noun_fixed::q3`).

**Изменено:** контент оставлен (ключ b корректен); жалоба закрыта вместе с прогоном. При повторе — проверить формулировку prompt.

---

### #3 — 2026-04-30

**Жалоба:** (без текста) — *can't* для запрета (`b3_theory_possibility::q10`).

**Изменено:** без правки JSON (ключ b верен). Закрыто.

---

### #2 — 2026-04-30

**Жалоба:** (без текста) — базовая форма после *did* (`b5_theory_verb_form_after_did::q10`).

**Изменено:** без правки JSON. Закрыто.

---

### #1 — 2026-04-30

**Жалоба:** (без текста) — postpositive прилагательное (`b3_postpositive::q1`).

**Изменено:** уточнён prompt: «выберите английский вариант для пропуска». Файл: `courses/english-grammar/training_pack/chapters/067/en.03.b3_postpositive.questions.json`.

---

## Spanish — слова (word_training)

### #33 — 2026-05-27 · niño (card 1241)

**Жалоба:** «Произношение кривое».

**Изменено:** `POST /api/internal/tts/regenerate` для *niño* (prod ES).

---

### #31, #27 — 2026-05-22 / 2026-05-12 · pueblo (cards 698, 699)

**Жалоба:** «Озвучка кривая».

**Изменено:** TTS regenerate *pueblo*.

---

### #32, #26, #23 — 2026-05-26 / 2026-05-11 / 2026-05-07 · aunque (card 8641)

**Жалоба:** «Пустой вариант ответа» / «Пустой ответ».

**Изменено:** `PUT /api/internal/training/card/8641` — `distractors_ru` без пустых строк: `["потому","однако","если","тогда"]` (было несколько `""` в JSON-массиве).

---

### #25 — 2026-05-08 · perdonar (card 8603)

**Жалоба:** «Нет озвучки».

**Изменено:** TTS regenerate *perdonar*.

---

### #24 — 2026-05-07 · reír (card 1585)

**Жалоба:** «Нет озвучки».

**Изменено:** TTS regenerate *reír*.

---

### #22 — 2026-05-07 · producto (card 1579)

**Жалоба:** «Озвучка кривая».

**Изменено:** TTS regenerate *producto*.

---

## Spanish — грамматика (grammar_training)

### #30 — 2026-05-19 · ser после qué (`b5_theory_word_order_question_word_plus_ser::q5`)

**Жалоба:** «Несколько неправильных ответов» — и (c), и (d) с ошибкой согласования.

**Изменено:** (d) заменён на грамматически верный *¿Qué es la casa?*; ключ остаётся (c). Файл: `courses/spanish-grammar/training_pack/chapters/018/es.05.b5_theory_word_order_question_word_plus_ser.questions.json`.

---

### #29 — 2026-05-17 · agua + прилагательное (`b5_theory_high_frequency_exceptions::q4`)

**Жалоба:** «Какое прилагательное» — неясная формулировка.

**Изменено:** prompt: «прилагательное … с существительным *agua*». Файл: `chapters/007/es.05.b5_theory_high_frequency_exceptions.questions.json`.

---

### #28 — 2026-05-17 · bailar, лицо tú (`b2_theory_present_regular_ar_endings::q8`)

**Жалоба:** «Несколько неправильно» — bailo/bailamos тоже неверны для «ты».

**Изменено:** дистракторы (c)(d) заменены на формы других глаголов в tú (*comes*, *cantas*); единственная ошибка для *bailar* — (b) *baila*. Файл: `chapters/028/es.02.b2_theory_present_regular_ar_endings.questions.json`.

---

### #21 — 2026-05-04 · empezar, ortografía (`b4_theory_verbs_zar_c_before_e::q9`)

**Жалоба:** «Пример на русском в вопросе».

**Изменено:** prompt переведён на испанский: *Ayer yo ___ la clase…*. Файл: `chapters/033/es.04.b4_theory_verbs_zar_c_before_e.questions.json`.

---

### #20 — 2026-05-04 · dormir nosotros (`b4_theory_present_regular_ir_endings::q4`)

**Жалоба:** «Только 1 верная форма, а вопрос про неверную» — несколько неверных для «мы».

**Изменено:** (d) заменён на *duerme* (другое лицо); одна целевая ошибка — *dormemos*. Файл: `chapters/028/es.04.b4_theory_present_regular_ir_endings.questions.json`.

---

### #19, #14, #10 — 2026-05-04 / 2026-05-01 · estar/somos nerviosos (`b6_theory_building_everyday_lines_location_state::q8`)

**Жалоба:** «Два неправильных варианта» / «в пояснении estar, а ответ somos» / путаница ser/estar + инфинитив.

**Изменено:** (c)(d) — корректные фразы с *estar*; единственная ошибка — (b) *somos nerviosos*; уточнён prompt. Файл: `chapters/021/es.06.b6_theory_building_everyday_lines_location_state.questions.json`.

---

### #18 — 2026-05-02 · hay + артикль (`b3_theory_hay_with_articles_quantifiers::q7`)

**Жалоба:** непонятно, новый объект или известный.

**Изменено:** prompt: ошибка при **первом** упоминании (нужен неопределённый артикль). Файл: `chapters/022/es.03.b3_theory_hay_with_articles_quantifiers.questions.json`.

---

### #17, #11 — 2026-05-02 · «Где собака» (`b2_theory_estar_core_meanings_a1::q4`)

**Жалоба:** «Несколько правильных ответов» — все с *está* в разных местах.

**Изменено:** prompt → «ошибка ser/estar для местоположения»; `correct_answer` **a → b** (*es en*). Файл: `chapters/023/es.02.b2_theory_estar_core_meanings_a1.questions.json`.

---

### #16 — 2026-05-02 · звук z (`b7_theory_decoding_algorithm_and_regional_notes::q6`)

**Жалоба:** без страны несколько верных ([θ] vs [s]).

**Изменено:** в prompt указано **кастильское произношение (Испания)**. Файл: `chapters/004/es.07.b7_theory_decoding_algorithm_and_regional_notes.questions.json`.

---

### #15 — 2026-05-01 · месяцы (`b3_theory_weekdays_months_calendar::q13`)

**Жалоба:** в вариантах смешаны испанский и русский.

**Изменено:** (b)(c)(d) полностью на испанском (*vacaciones*, *mis amigos*, *llueve mucho*). Файл: `chapters/026/es.03.b3_theory_weekdays_months_calendar.questions.json`.

---

### #13 — 2026-05-01 · mucho перед глаголом (`b4_theory_present_adverbs_before_finite_verb_questions::q6`)

**Жалоба:** противоречие — в «правильном» (a) *mucho* после глагола.

**Изменено:** `correct_answer` **a → c** (*¿Mucho hablas…?*); пояснение. Файл: `chapters/030/es.04.b4_theory_present_adverbs_before_finite_verb_questions.questions.json`.

---

### #12 — 2026-05-01 · tampoco (`b3_theory_negation_tampoco_neither::q10`)

**Жалоба:** несколько правильных / неверное пояснение.

**Изменено:** prompt → «где *yo* не согласован с первой репликой»; ключ (b) *No bailan → Yo tampoco bailo*; краткое пояснение. Файл: `chapters/031/es.03.b3_theory_negation_tampoco_neither.questions.json`.

---

### #9 — 2026-04-30 · decode Me gusta (`b4_theory_decode_high_frequency_mini_patterns::q1`)

**Жалоба:** непонятно, почему одно слово чётко.

**Изменено:** уточнён prompt (служебное *me*, не проглатывать). Файл: `chapters/006/es.04.b4_theory_decode_high_frequency_mini_patterns.questions.json`.

---

### #8 — 2026-04-30 · Ella es médica (`b1_theory_basic_affirmative_statements::q19`)

**Жалоба:** неясно, почему *doctora* «неправильно».

**Изменено:** (d) заменён на явно неверный *Ella es doctores*; пояснение про *médica* / род. Файл: `chapters/017/es.01.b1_theory_basic_affirmative_statements.questions.json`.

---

## Правила на будущее

Добавлено: `.cursor/rules/content-quality-guardrails.mdc` (чеклист MCQ + word cards + релиз).

---

## Артефакты прогона

| Файл | Назначение |
|------|------------|
| `docs/complaints/journal-2026-06-04-triage.md` | этот текстовый журнал (в git) |
| `logs/complaints/triage-2026-06.jsonl` | машинный журнал resolve (локально) |
| `logs/complaints/clusters-*-latest.json` | кластеры на момент fetch |
| `tools-local/complaints-triage/apply_prod_word_fixes.py` | TTS + aunque |
| `tools-local/complaints-triage/resolve_all_active.py` | bulk resolve |
