Ниже — полное ТЗ на SRS-тренировку: карточки по значениям, два направления (RU→EN и EN→RU), мультивыбор с задержкой вариантов, SM-2-подобное планирование с "скрытой шкалой" качества (авто-маппинг по времени/подсказкам/ошибкам), раз в сутки — одна тренировка/одно уведомление.

**Интеграция с существующей системой**: тренировочные карточки генерируются из существующих словарных карточек (`word_cards`) с помощью LLM.

------

## 1) Цели и ограничения

### Цель

Сделать тренировку слов для англ. карточек в боте по принципам интервального повторения:

- показывать пользователю “что пора”, а не перебирать всё подряд;
- минимизировать “муторность”: **нет явной шкалы 0–3**;
- использовать **retrieval + recognition**: сначала короткая попытка вспомнить, затем варианты ответа;
- подстроить интервалы под реальную сложность карточки на основе поведения.

### Ограничения

- Пользователь заходит редко ⇒ целевая частота: **1 тренировка в сутки**.
- Уведомления: бот может написать “повторите слова”; **если после уведомления пользователь не активен — повторно в этот день не пишем**.
- Слово может иметь несколько значений ⇒ это **несколько карточек** (единица SRS — “значение”, не “слово”).
- Тренировка в двух направлениях: **RU→EN** и **EN→RU** (и оба направления имеют собственный прогресс).

------

## 2) Интеграция с существующей системой

### 2.1. Существующие сущности

В системе уже есть:
- **`word_cards`** — таблица словарных карточек, созданных LLM
  - `id`, `word`, `definition`, `created_at`, `updated_at`
- **`word_request_history`** — история запросов пользователей

### 2.2. Источник данных для тренировок

Данные для тренировок создаются на основе `word_cards`:
- `word_card` → `training_cards` (1:N, на каждое значение слова — отдельная карточка)
- `training_card` → `user_cards` (1:2, карточки ru_en и en_ru)
- Генерация происходит через LLM отдельным воркером (см. раздел 12)

### 2.3. Принцип работы

1. Пользователь запрашивает слово → создаётся `word_card` (уже реализовано)
2. Воркер подхватывает новые `word_cards` без `training_cards`
3. LLM генерирует JSON с тренировочными данными
4. **Только после успешного ответа LLM** создаются:
   - `training_cards` — по одной записи на каждое значение слова (1-3 штуки)
   - `user_cards` — по 2 карточки на каждую `training_card` (ru_en + en_ru)
5. Если LLM вернул ошибку — ничего не создаём, воркер попробует позже

------

## 3) Сущности и термины

- **Словарная карточка** (`word_card`): исходная карточка, созданная LLM при запросе слова. Содержит `definition` в свободном формате.
- **Тренировочная карточка** (`training_card`): одно значение слова с данными для тренировки (перевод, примеры, дистракторы, подсказка). Связь N:1 с `word_card` (одно слово может иметь несколько значений).
- **Карточка пользователя** (`user_card`): единица SRS-тренировки = (training_card + направление).
   Пример: training_card "bank=берег" → отдельная карточка RU→EN и отдельная EN→RU.
- **Сессия тренировки** (`training_session`): один запуск тренировки пользователем (в идеале раз в сутки).
- **Попытка ответа** (`review_event`): один показ карточки и ответ пользователя.

**Связи между сущностями:**
```
word_cards (1) ←→ (N) training_cards (1) ←→ (2) user_cards
                                              ↑
                                         (ru_en, en_ru)
```

------

## 4) UX карточки: мультивыбор с задержкой вариантов

### 4.1. Общий поток одной карточки

1. **Шаг A — промпт без вариантов** (задержка):
   - показываем вопрос (например: “переведи на английский: *берег (реки)*”).
   - **варианты НЕ показываем** `T_delay` секунд.
   - на экране есть:
     - таймер/индикатор “подумайте…” (не обязателен),
     - кнопка **«Показать варианты»** (активна сразу) или “Я готов” (опционально).
2. **Шаг B — мультивыбор**:
   - показываем 4–6 вариантов (конфигурируемо), один правильный.
   - пользователь выбирает один вариант.
   - сразу показываем правильный ответ (и помечаем выбранный).
3. **Шаг C — (опционально) микро-объяснение**:
   - краткое пояснение/пример (если есть).

### 4.2. Параметры задержки

**Задержка перед показом вариантов (`T_delay`):**
- По умолчанию 5 секунд (настраивается через `TRAINING_OPTIONS_DELAY_MS` в миллисекундах).
- Логика:
  - если пользователь нажал "Показать варианты" до истечения задержки — считаем это сигналом "трудно".
  - если дождался автоматического появления вариантов — нейтрально.
  - если ответил быстро и правильно — "легко".

**Задержка после неправильного ответа:**
- По умолчанию 5 секунд (настраивается через `TRAINING_WRONG_ANSWER_DELAY_SECONDS`).
- Применяется только после неправильного ответа, чтобы дать время прочитать объяснение ошибки.
- После правильного ответа следующая карточка показывается сразу.

### 4.3. Количество вариантов

- Дефолт: 4 варианта.
- Если система видит частые ошибки/низкое качество — можно временно увеличить до 6 для снижения угадывания и лучшей дифференциации (опционально).

### 4.4. Формирование вариантов (генерируются LLM)

Для одной карточки нужны:

- `correct_answer`
- `distractors[]` (неверные варианты)
   Источники дистракторов (по приоритету):

1. **Слова из текущей сессии тренировки** (1-2 слова) — правильные ответы из других карточек в этой же сессии
   - Для RU→EN: английские слова (`WordEN`) из других карточек
   - Для EN→RU: русские слова (`WordRU`) из других карточек
   - **Исключение:** НЕ использовать слова из последних 1-2 правильно отвеченных карточек в текущей сессии (чтобы избежать "узнавания по свежести")
   - Цель: предотвратить угадывание по признаку "знакомости" слова (если пользователь видел слово недавно, он может угадать его даже не зная значения)
2. Дистракторы, сгенерированные LLM (`distractors_ru` / `distractors_en` из `training_cards`) — 1-2 слова
3. Недавно ошибочно выбранные варианты пользователя (`wrong_answers_json`) — персонализация

**Состав вариантов (для 4 вариантов):**
- 1 правильный ответ
- 1-2 слова из текущей сессии (другие карточки)
- 1-2 дистрактора от LLM

**Требования к дистракторам от LLM:**
- Дистракторы НЕ должны быть синонимами правильного ответа
  - Например, если правильный ответ "бежать", не использовать "мчаться", "нестись", "спешить"
- Дистракторы должны быть схожи по длине и стилю с правильным ответом
  - Поскольку `word_ru` всегда одно слово, дистракторы тоже должны быть однословными
- Фокус на словах, которые семантически отличаются, но могут быть перепутаны (ложные друзья, похожее написание, та же категория, но другое значение)

**Требования к русскому слову (`word_ru`):**
- Должно быть **ОДНИМ словом** — это не определение, а русское слово-перевод
- НЕ включать несколько синонимов через запятую (например, использовать "монах", а не "монах, духовен человек, живущий по монашескому обету")
- НЕ включать объяснения, описания или дополнительный контекст
- Выбирать наиболее распространенный однословный перевод (существительное, глагол, прилагательное и т.д.)
- Примеры: "монах" (не "монах, духовен человек"), "бежать" (не "бежать, бегать"), "банк" (не "банк, финансовое учреждение")

Хранить дистракторы можно в JSON прямо в карточке и/или вычислять на лету.

------

## 5) Алгоритм интервального повторения (SM-2-подобный, авто-качество)

### 5.1. Состояния карточки

У каждой карточки есть `state`:

- `new` — новая, ещё не тренировалась;
- `learning` — проходит короткие шаги внутри первых повторов/после провалов;
- `review` — обычный режим интервальных повторений.

### 5.2. Learning steps (учитывая "редко пользуются")

Чтобы не зависеть от внутридневных заходов, делаем шаги, которые могут пережить редкое использование:

**Рекомендуемые шаги по умолчанию (зависят от направления):**

- **RU→EN** (активное воспроизведение, сложнее):
  - Step 0: `+1 день`
  - Step 1: `+3 дня`
  - Step 2: `+7 дней`
  - Step 3: `+14 дней` (опционально, для более плавного перехода)

- **EN→RU** (пассивное узнавание, проще):
  - Step 0: `+1 день`
  - Step 1: `+3 дня`
  - Step 2: `+7 дней`

Пояснение: если пользователь заходит раз в пару дней, "10 минут" бессмысленно — карточка всё равно не будет повторена.

Настройка (конфиг):

- `LEARNING_STEPS_DAYS_RU_EN = [1, 3, 7, 14]` — для направления RU→EN
- `LEARNING_STEPS_DAYS_EN_RU = [1, 3, 7]` — для направления EN→RU

**Переход в review:**

После успешного прохождения последнего шага → `review` с начальным интервалом:
- Если прошло 2+ успешных learning-шага: `interval_days = 3-4` (более честный старт)
- Если прошло только 1 шаг: `interval_days = 1` (стандартный старт)

### 5.3. Параметры SM-2, которые храним на карточке

- `ef` (easiness factor), float, старт 2.5, минимум 1.3
- `reps` (число успешных повторений в review), int
- `interval_days` (текущий интервал), int
- `next_due_at` (когда снова показывать), datetime
- `learning_step` (индекс шага), int
- `lapse_count` (кол-во провалов), int
- `last_review_at`, datetime

### 5.4. Авто-маппинг "качества" (скрытая шкала)

После ответа рассчитываем внутренний `quality_bucket` (0..3), который пользователь не видит.

Входные сигналы на одну попытку:

- `correct` (bool)
- `delay_used_ms` — сколько прошло до показа вариантов (и было ли нажатие “показать варианты”)
- `answer_time_ms` — от появления вариантов до выбора ответа
- `early_reveal` — нажал ли “показать варианты” до истечения `T_delay`
- `option_count` — сколько вариантов показывали
- (опционально) `had_hint` / “показать перевод/пример” если добавите

Рекомендованный маппинг:

**Если неверно (`correct=false`) → quality=0.**

Если верно:

- quality=1 (“трудно”), если:
  - `early_reveal=true` ИЛИ
  - `answer_time_ms` > `SLOW_THRESHOLD_MS` (например 8000) ИЛИ
  - карточка уже была в lapse недавно (опционально)
- quality=3 (“легко”), если:
  - `early_reveal=false`
  - `answer_time_ms` < `FAST_THRESHOLD_MS` (например 2500)
  - и карточка не новая (или `reps>=1`)
- иначе quality=2 (“нормально”)

Пороговые значения (конфиг):

- `FAST_THRESHOLD_MS=2500`
- `SLOW_THRESHOLD_MS=8000`
- `T_delay=3000` (например)

### 5.5. Обновление карточки по quality (SM-2-подобно)

Дальше маппим `quality` в SM-2 качество `q` (0..5) внутренне:

- quality 0 → q=0
- quality 1 → q=3
- quality 2 → q=4
- quality 3 → q=5

И применяем стандартную логику:

**A) Если q < 3 (то есть quality=0):**

- `lapse_count += 1`
- `state = learning`
- `learning_step = 0`
- `reps = 0`
- `interval_days = 0`
- `ef = max(1.3, ef - 0.2)`
- `next_due_at = now + LEARNING_STEPS_DAYS[0]`

**B) Если state=learning и quality>=1:**

- Определяем массив шагов по направлению:
  - если `direction == 'ru_en'`: `steps = LEARNING_STEPS_DAYS_RU_EN`
  - если `direction == 'en_ru'`: `steps = LEARNING_STEPS_DAYS_EN_RU`

- если quality=1:
  - можно оставить тот же step (без продвижения): `learning_step = max(0, learning_step)`
  - `next_due_at = now + steps[learning_step]`
- если quality=2 или 3:
  - `learning_step += 1`
  - если вышли за пределы массива:
    - `state=review`
    - `reps=1`
    - `interval_days = 3-4` если `learning_step >= 2`, иначе `interval_days = 1`
    - `next_due_at=now + interval_days`
  - иначе `next_due_at = now + steps[learning_step]`

**C) Если state=review и quality>=1:**

1. Обновляем `ef` по SM-2 формуле:
   - `ef = max(1.3, ef + (0.1 - (5-q)*(0.08 + (5-q)*0.02)))`
2. Интервал:
   - если `reps==0`: `interval_days = 1`
   - если `reps==1`: `interval_days = 6`
   - если `reps>=2`: `interval_days = ceil(interval_days * ef)`
3. `reps += 1`
4. `next_due_at = now + interval_days`

> Примечание: `reps` тут — именно “кол-во успешных review-повторов”, в `learning` его можно держать 0.

------

## 6) Два направления (RU→EN и EN→RU)

### 6.1. Модель

Одна `training_card` порождает **две карточки пользователя**:

- `direction = "ru_en"`
- `direction = "en_ru"`

Каждая карточка пользователя имеет собственные:

- `state`, `ef`, `reps`, `interval`, `next_due_at`…
   Потому что человек может узнавать слово, но не уметь воспроизводить (и наоборот), а интервалы должны отличаться.

### 6.2. Баланс направлений в сессии

По умолчанию:

- если обе карточки due — показываем обе, но **не подряд**, а с разрывом 3–5 карточек (чтобы не было “подсказки контекстом”).
- если надо ограничить объём — используем квоты:
  - `max_due_per_session`
  - `direction_ratio` (например 50/50 или 60/40 в пользу RU→EN, если цель — активный словарь)

------

## 7) Формирование очереди в сессии

### 7.1. Входные данные для генерации

На старте сессии вычисляем:

- `learning_due`: карточки `state=learning` и `next_due_at<=now`
- `review_due`: карточки `state=review` и `next_due_at<=now`
- `new_candidates`: карточки `state=new` (если хотим добавлять новые)

Конфиги:

- `MAX_CARDS_PER_SESSION` (например 30)
- `MAX_NEW_PER_SESSION` (например 5)
- `OVERDUE_PRIORITY_MODE`: сортировка по “насколько просрочено”

### 7.2. Правила набора

1. Берём `learning_due` в приоритет.
2. Затем `review_due`.
3. Если места осталось — добавляем `new` (не больше `MAX_NEW_PER_SESSION`).

### 7.3. Сортировка и перемешивание

- Основной сортировщик: `overdue_seconds = now - next_due_at` по убыванию (самые просроченные раньше).
- После сортировки применяется двухэтапное перемешивание:
  1. **Полное случайное перемешивание** всех карточек для максимальной вариативности.
  2. **Алгоритм разнесения одинаковых слов** для предотвращения подряд идущих карточек одного слова.

### 7.4. Анти-повтор смысла/слова

Чтобы не выдать подряд карточки одного и того же слова с разными значениями/направлениями:

- Правило: карточки с одинаковым `word_card_id` разносятся на минимальное расстояние:
  - 3 позиции для маленьких очередей (< 10 карточек)
  - 4 позиции для средних (10-20 карточек)
  - 5 позиций для больших (> 20 карточек)
- Внутри группы одного слова карточки также перемешиваются случайным образом.
- Если разнести невозможно (слишком много карточек одного слова) — допускается, но стараемся максимально разнести.

------

## 8) Уведомления и расписание (раз в сутки)

### 8.1. Когда слать

**Окно отправки уведомления:**

Уведомление отправляется в период от `preferred_training_time` до конца дня (23:59:59) в локальном времени пользователя:

- если `due_count > 0`
- и **сегодня ещё не было отправлено уведомление**
- и пользователь **сегодня ещё не начинал тренировку**
- и текущее время >= `preferred_training_time` (дефолт 19:00)

**Логика "мягкого окна":**

- Если бот был недоступен в точное время `preferred_training_time`, уведомление отправляется при первом удобном тике крон-задачи в пределах окна (от `preferred_training_time` до конца дня)
- Если окно пропущено (например, бот был недоступен весь день) — уведомление не отправляется в этот день, на следующий день можно слать снова

### 8.2. Поведение "не проявил активность"

Если уведомление отправлено и в течение суток пользователь не начал сессию:

- **в этот день больше не слать**
- на следующий день можно слать снова (если due всё ещё есть)

### 8.3. Текст уведомления

Уведомление отправляется **только если пользователь сегодня ещё не тренировался** (по календарному дню в его timezone). Текст строится так, чтобы **приободрить** пользователя: признание усилий (серия дней, рост словаря), без вины при пропуске.

**Варианты текста (в зависимости от контекста):**

- **Есть серия** (пользователь тренировался вчера и подряд до этого): одобрение серии («Уже X дней подряд занимаешься — так держать!») + при наличии рост за неделю («За неделю +N слов в словаре — здорово!») + «К повторению: K карточек (~M мин). Продолжим?»
- **Вчера не занимался** (серия прервалась или нет серии): мягкое приглашение («Сегодня отличный день, чтобы вернуться») + при наличии «За неделю ты уже добавил N слов — давай не останавливаться» + «К повторению: K карточек (~M мин). Начать?»
- **Нет серии / новый пользователь**: при наличии роста за неделю — «За неделю +N слов в твоём словаре — каждый шаг считается. К повторению: …»; иначе нейтрально: «К повторению: K карточек (~M мин). Начать?»

Кнопки: **«Начать»** и **«Отписаться»**.

Оценка `M минут`:

- `M = ceil(N * avg_seconds_per_card / 60)`
- `avg_seconds_per_card` по умолчанию 15; можно держать по пользователю (скользящее среднее).

Данные для контекста:

- **Серия (streak)** — число подряд идущих дней с хотя бы одной тренировкой, считая по «вчера» и ранее (timezone пользователя).
- **Вчера тренировался** — во вчерашней дате (в timezone пользователя) есть хотя бы одна сессия.
- **За неделю +N слов** — количество записей в `user_cards` с `created_at` за последние 7 дней (в timezone пользователя).

------

## 9) Логи и аналитика (минимально необходимые)

На каждую попытку сохранять:

- что показали (карточка, направление, варианты),
- что выбрал пользователь,
- тайминги (до вариантов и до ответа),
- правильность,
- вычисленное `quality` и применённые изменения `ef/interval/next_due`.

Это нужно, чтобы:

- отлаживать алгоритм,
- подбирать пороги FAST/SLOW,
- строить метрики удержания.

------

## 10) Структура БД (адаптированная под существующую систему)

Ниже — рекомендуемая структура. Полной нормализации не требуется; я даю “скелет”, который удобно расширять.

### 10.1. word_cards (СУЩЕСТВУЮЩАЯ)

Таблица уже есть в системе, не меняем:

- `id` (PK)
- `word` (TEXT, UNIQUE)
- `definition` (TEXT) — полное описание от LLM
- `created_at`, `updated_at`

### 10.2. training_cards (НОВАЯ)

Тренировочные карточки — одна запись = одно значение слова. Создаётся только после успешного ответа LLM.

- `id` (PK)
- `word_card_id` (FK → word_cards.id)
- `word_en` (TEXT) — слово на английском (дублируем для удобства)
- `transcription` (TEXT) — IPA транскрипция
- `sense_index` (INT) — порядковый номер значения (0, 1, 2...)
- `word_ru` (TEXT) — русское слово (одно слово, не определение)
- `meaning_en` (TEXT) — определение на английском
- `example_en` (TEXT) — пример на английском
- `example_ru` (TEXT) — перевод примера
- `distractors_ru` (TEXT) — JSON-массив дистракторов на русском
- `distractors_en` (TEXT) — JSON-массив дистракторов на английском
- `hint` (TEXT) — подсказка для запоминания
- `created_at`

Индексы:
- уникальный `(word_card_id, sense_index)`
- `(word_card_id)` — для поиска всех значений слова

**Пример данных для слова "bank":**

| id | word_card_id | sense_index | word_en | word_ru | meaning_en |
|----|--------------|-------------|---------|------------|------------|
| 1 | 1 | 0 | bank | берег (реки) | the land along the side of a river |
| 2 | 1 | 1 | bank | банк (финансовая организация) | financial institution |

### 10.4. users (НОВАЯ)

- `id` (PK)
- `telegram_id` (BIGINT, UNIQUE)
- `timezone` (TEXT, default `Europe/Moscow`)
- `preferred_training_time` (TIME, default `19:00`)
- `settings_json` (JSON: лимиты, включён ли EN→RU, параметры)
- `created_at`, `updated_at`

### 10.5. user_cards

Главная таблица SRS. Одна запись = (user + training_card + direction).

- `id` (PK)
- `user_id` (FK → users.id)
- `training_card_id` (FK → training_cards.id)
- `direction` (enum: `ru_en`, `en_ru`)

SRS-поля:

- `state` (enum: `new`, `learning`, `review`)
- `ef` (float, default 2.5)
- `reps` (int, default 0)
- `interval_days` (int, default 0)
- `learning_step` (int, default 0)
- `lapse_count` (int, default 0)
- `next_due_at` (datetime, nullable; для new может быть null)
- `last_review_at` (datetime, nullable)
- `last_quality` (tinyint, nullable)

Контент/варианты/ошибки (денормализация разрешена):

- `last_options_json` (JSON: последний набор вариантов, чтобы потом “пояснить”)
- `wrong_answers_json` (JSON массив объектов вида `{option, ts, count}` — как ты и хотел)
- `stats_json` (JSON: avg_time_ms, correct_streak и т.п.)

Индексы:

- `(user_id, next_due_at)`
- `(user_id, state)`
- `(user_id, direction)`
- уникальный `(user_id, training_card_id, direction)`

### 10.6. training_sessions

- `id` (PK)
- `user_id` (FK)
- `started_at`
- `ended_at` (nullable)
- `source` (enum: `nudge`, `manual`)
- `planned_count` (сколько карточек было запланировано на старте)
- `done_count`
- `session_json` (JSON: параметры сессии, лимиты, версии алгоритма)

### 10.7. review_events (лог попыток)

Можно хранить детально, даже если часть данных дублируется в user_cards.

- `id` (PK)
- `session_id` (FK → training_sessions.id)
- `user_id` (FK → users.id)
- `user_card_id` (FK → user_cards.id)
- `direction`
- `shown_at`
- `options_shown_at` (nullable)
- `answered_at`
- `t_delay_ms` (сколько было задержки)
- `early_reveal` (bool)
- `option_count` (int)
- `options_json` (JSON: список вариантов)
- `chosen_option` (TEXT) — выбранный вариант ответа
- `is_correct` (bool)
- `quality` (0..3)
- `metrics_json` (JSON: `answer_time_ms`, `total_time_ms`, и т.п.)
- `srs_before_json` (JSON: снимок ef/reps/interval/state до)
- `srs_after_json` (JSON: снимок после)

### 10.8. training_nudges (уведомления)

- `id` (PK)
- `user_id` (FK)
- `local_date` (date в TZ пользователя)
- `sent_at`
- `consumed_at` (nullable; если пользователь начал сессию)
- `due_count_at_send`
- `message_id` (опционально)
- уникальный `(user_id, local_date)` чтобы гарантировать 1 раз в день

------

### 10.9. circuit_breaker_state (НОВАЯ)

Состояние circuit breaker для воркера генерации карточек:

- `id` (PK, всегда 1 — синглтон)
- `is_open` (BOOL, default false) — открыт ли breaker (заблокирована отправка)
- `failure_count` (INT, default 0) — счётчик последовательных ошибок
- `last_failure_at` (DATETIME, nullable)
- `last_reset_at` (DATETIME, nullable)
- `updated_at`

------

## 11) API/логика сервиса (как разбить по модулям)

### 11.1. Сервис "планировщик" (SRS)

- `getDueCards(user_id, now) -> list<CardDTO>`
- `gradeCard(user_id, card_id, attempt_data) -> updated_card`

Где `attempt_data` включает:

- `correct`, `early_reveal`, `answer_time_ms`, `t_delay_ms`, `option_count`, `chosen_option`, …

### 11.2. Сервис "генератор вариантов"

- `generateOptions(card, optionCount, sessionWords) -> {options[], correct}`
   
   Параметры:
   - `card` — карточка пользователя с данными тренировочной карточки
   - `optionCount` — количество вариантов ответа (обычно 4)
   - `sessionWords` — список правильных ответов из других карточек текущей сессии
   
   Источники вариантов (по приоритету):
   1. **Слова из текущей сессии** (`sessionWords`) — 1-2 слова
      - Смешиваются с дистракторами для предотвращения угадывания по признаку знакомости
      - **Исключение:** НЕ использовать слова из последних 1-2 правильно отвеченных карточек в текущей сессии (параметр `RECENT_CORRECT_EXCLUDE_COUNT=2`, настраивается)
   2. `distractors_ru` / `distractors_en` из `training_cards` — дистракторы, сгенерированные LLM (1-2 слова)
   3. `wrong_answers_json` пользователя — персонализированные ошибки
   4. fallback: случайные из других `training_cards` (если не хватает вариантов)
   
   Логика формирования:
   - Исключаются другие значения того же слова (чтобы не путать смыслы)
   - Исключается правильный ответ текущей карточки из `sessionWords`
   - Исключаются правильные ответы из последних `RECENT_CORRECT_EXCLUDE_COUNT` карточек сессии (чтобы избежать "узнавания по свежести")
   - Случайное перемешивание для рандомизации позиции правильного ответа

### 11.3. Сервис "сессия"

- `startSession(user_id, source) -> session_id + first_batch`
- `nextCard(session_id) -> card`
- `finishSession(session_id)`

### 11.4. Сервис "уведомления"

- ежедневный крон (запускается каждые 30-60 минут в течение дня):
  - для каждого пользователя определить `local_date` и `local_time` (в TZ пользователя)
  - если `local_time >= preferred_training_time` (начало окна)
  - и `local_time < 23:59:59` (конец окна)
  - и nudge на эту дату нет
  - и due_count>0
  - и пользователь сегодня ещё не начинал тренировку
  - → отправить уведомление
  - → сохранить запись в `training_nudges`

------

## 12) Воркер генерации тренировочных карточек (LLM)

### 12.1. Назначение

Отдельный фоновый процесс, который:
1. Находит `word_cards` без соответствующих `training_cards`
2. Отправляет запрос к LLM для генерации тренировочных данных
3. Сохраняет результат в `training_cards`
4. Создаёт `user_cards` для пользователей, запрашивавших это слово

### 12.2. Промпт для LLM

Файл: `prompts/training-card-generator.txt`

**Ключевые требования промпта:**

1. **Русское слово (`word_ru`):**
   - Должно быть **ОДНИМ словом** — это не определение, а русское слово-перевод
   - НЕ включать несколько синонимов через запятую
   - НЕ включать объяснения, описания или дополнительный контекст
   - Примеры: "монах" (не "монах, духовен человек"), "бежать" (не "бежать, бегать"), "банк" (не "банк, финансовое учреждение")

2. **Правила для дистракторов:**
   - **КРИТИЧНО:** Дистракторы НЕ должны быть синонимами правильного ответа
     - Если правильный ответ "бежать", не использовать "мчаться", "нестись", "спешить"
   - Фокус на словах, которые семантически отличаются, но могут быть перепутаны
   - Использовать слова с похожим написанием, ложные друзья, ту же категорию, но другое значение
   - Дистракторы должны быть схожи по длине и стилю с правильным ответом
     - Поскольку `word_ru` всегда одно слово, дистракторы тоже должны быть однословными
   - НЕ включать другие значения того же слова (например, для "bank" как "банк" не использовать "берег")

3. **Структура ответа:**
   - Извлекать 1-3 наиболее распространенных значения слова
   - Для каждого значения предоставлять: перевод, определение, примеры, дистракторы, подсказку

Полный текст промпта см. в файле `prompts/training-card-generator.txt`.

### 12.3. Формат ответа LLM

Ожидаемый JSON-ответ (пример для слова "run"):

```json
{
  "word_en": "run",
  "transcription": "/rʌn/",
  "senses": [
    {
      "index": 0,
      "word_ru": "бежать",
      "meaning_en": "to move quickly on foot",
      "example_en": "I run every morning in the park",
      "example_ru": "Я бегаю каждое утро в парке",
      "distractors_ru": ["идти", "прыгать", "ходить", "летать"],
      "distractors_en": ["walk", "jump", "skip", "fly"],
      "hint": "Представьте спринтера на старте — он готов RUN!"
    },
    {
      "index": 1,
      "word_ru": "управлять",
      "meaning_en": "to manage or be in charge of",
      "example_en": "She runs a successful business",
      "example_ru": "Она управляет успешным бизнесом",
      "distractors_ru": ["работать", "владеть", "создавать", "продавать"],
      "distractors_en": ["work", "own", "create", "sell"],
      "hint": "Директор RUNит компанию — руководит ею"
    },
    {
      "index": 2,
      "word_ru": "работать",
      "meaning_en": "to operate or function",
      "example_en": "The program runs on Windows",
      "example_ru": "Программа работает на Windows",
      "distractors_ru": ["включаться", "загружаться", "открываться", "запускаться"],
      "distractors_en": ["start", "load", "open", "execute"],
      "hint": "Компьютер RUNит программу — она работает"
    }
  ]
}
```

### 12.4. Логика воркера

```
every 30 seconds:
  if circuit_breaker.is_open:
    skip iteration
  
  pending_cards = SELECT word_cards WHERE NOT EXISTS training_cards
  
  for card in pending_cards (limit 5):
    try:
      response = LLM.generate(prompt + card.word)
      parsed = JSON.parse(response)
      validate(parsed)
      
      // Создаём training_cards только после успешного ответа LLM
      // По одной записи на каждое значение слова
      training_card_ids = []
      for sense in parsed.senses:
        training_card_id = INSERT INTO training_cards (
          word_card_id, word_en, transcription, sense_index,
          word_ru, meaning_en, example_en, example_ru,
          distractors_ru, distractors_en, hint
        )
        training_card_ids.append(training_card_id)
      
      // Создаём user_cards для всех пользователей, запрашивавших это слово
      users = SELECT DISTINCT user_id FROM word_request_history WHERE word = card.word
      
      for user in users:
        for training_card_id in training_card_ids:
          INSERT user_cards (user_id, training_card_id, direction='ru_en')
          INSERT user_cards (user_id, training_card_id, direction='en_ru')
      
      circuit_breaker.reset_failures()
    
    catch error:
      circuit_breaker.record_failure()
      // НЕ создаём training_cards при ошибке — воркер попробует позже
      
      if circuit_breaker.failure_count >= 5:
        circuit_breaker.open()
        notify_admin("Circuit breaker opened: 5 consecutive LLM failures")
```

### 12.5. Конфигурация воркера

Переменные окружения:

- `TRAINING_WORKER_INTERVAL=30s` — интервал между итерациями
- `TRAINING_WORKER_BATCH_SIZE=5` — сколько карточек обрабатывать за итерацию
- `TRAINING_WORKER_ENABLED=true` — включён ли воркер

------

## 13) Circuit Breaker и уведомления администратору

### 13.1. Circuit Breaker

Механизм защиты от каскадных ошибок LLM:

**Состояния:**
- `CLOSED` (is_open=false) — нормальная работа
- `OPEN` (is_open=true) — запросы заблокированы

**Переходы:**
- CLOSED → OPEN: 5 последовательных ошибок LLM
- OPEN → CLOSED: 
  - Автоматически раз в сутки (по крону в 00:00)
  - Вручную через команду `/reset_circuit`

**Конфигурация:**
- `CIRCUIT_BREAKER_THRESHOLD=5` — порог срабатывания
- `CIRCUIT_BREAKER_AUTO_RESET_HOURS=24` — автосброс через N часов

### 13.2. Уведомления администратору

При открытии circuit breaker отправляется уведомление в Telegram:

**Получатель:** `ADMIN_TELEGRAM_ID` из .env

**Формат сообщения:**
```
⚠️ Circuit Breaker ОТКРЫТ

Воркер генерации карточек остановлен.
Причина: 5 последовательных ошибок LLM.

Последняя ошибка: {error_message}
Время: {timestamp}

Для сброса используйте /reset_circuit
```

### 13.3. Переменные окружения

```env
# Telegram ID администратора (получить через /get_id)
ADMIN_TELEGRAM_ID=123456789
```

------

## 14) Команды бота

### 14.1. Публичные команды (отображаются в списке)

- `/start` — приветствие и краткая справка
- `/help` — помощь по боту
- `/train` — начать тренировку

### 14.2. Служебные команды (НЕ отображаются в списке)

Эти команды не регистрируются через `setMyCommands` и не видны пользователям:

- `/get_id` — получить свой Telegram ID
  - Доступна всем пользователям
  - Ответ: `Your Telegram ID: {user_id}`

- `/reset_circuit` — сбросить circuit breaker
  - **Доступна только администратору** (проверка `user_id == ADMIN_TELEGRAM_ID`)
  - При вызове не-админом: команда игнорируется
  - Успешный ответ: `✅ Circuit breaker сброшен. Воркер возобновит работу.`

- `/delete_train <word_en>` — удалить тренировочные карточки для указанного слова
  - **Доступна только администратору**
  - Удаляет все тренировочные карточки для слова (каскадно удаляются связанные `user_cards` и `review_events`)
  - Пример: `/delete_train monk`
  - Ответ: `✅ Удалено тренировочных карточек для слова 'monk': N`

- `/delete_train_all` — удалить все тренировочные карточки
  - **Доступна только администратору**
  - Удаляет все тренировочные карточки из системы (каскадно удаляются все связанные данные)
  - Использовать с осторожностью!
  - Ответ: `✅ Удалено всех тренировочных карточек: N` + информация о каскадном удалении

### 14.3. Регистрация команд

При запуске бота регистрируются только публичные команды:

```go
commands := []tgbotapi.BotCommand{
    {Command: "start", Description: "Начать работу с ботом"},
    {Command: "help", Description: "Помощь"},
    {Command: "train", Description: "Начать тренировку слов"},
}
bot.Request(tgbotapi.NewSetMyCommands(commands...))
```

Служебные команды (`/get_id`, `/reset_circuit`, `/delete_train`, `/delete_train_all`) обрабатываются в коде, но не регистрируются.

------

## 15) Версионирование алгоритма

Очень рекомендую хранить "версию алгоритма" в `training_sessions.session_json` и/или в `user_cards.stats_json`, чтобы потом можно было менять пороги и шаги без боли:

- `algo_version: "srs_v2_delayed_mcq_sm2_autoquality"`
- `config_snapshot: {...}`

------

## 16) Полный список переменных окружения

Добавить в `.env`:

```env
# === Training Worker ===
TRAINING_WORKER_ENABLED=true
TRAINING_WORKER_INTERVAL=30s
TRAINING_WORKER_BATCH_SIZE=5

# === Circuit Breaker ===
CIRCUIT_BREAKER_THRESHOLD=5
CIRCUIT_BREAKER_AUTO_RESET_HOURS=24

# === Admin Notifications ===
# Получить через команду /get_id в боте
ADMIN_TELEGRAM_ID=

# === Training Prompt ===
TRAINING_PROMPT_FILE=prompts/training-card-generator.txt

# === Training UI Configuration ===
# Задержка между показом вопроса и вариантов ответа (в миллисекундах)
TRAINING_OPTIONS_DELAY_MS=5000

# Задержка после неправильного ответа перед показом следующей карточки (в секундах)
TRAINING_WRONG_ANSWER_DELAY_SECONDS=5

# === Learning Steps Configuration ===
# Шаги для направления RU→EN (активное воспроизведение, сложнее)
LEARNING_STEPS_DAYS_RU_EN=1,3,7,14

# Шаги для направления EN→RU (пассивное узнавание, проще)
LEARNING_STEPS_DAYS_EN_RU=1,3,7

# === Options Generation Configuration ===
# Количество последних правильных ответов в сессии, которые исключаются из дистракторов
RECENT_CORRECT_EXCLUDE_COUNT=2
```

------

## 17) Схема миграции БД

### 17.1. Новые таблицы

```sql
-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    telegram_id INTEGER NOT NULL UNIQUE,
    timezone TEXT DEFAULT 'Europe/Moscow',
    preferred_training_time TEXT DEFAULT '19:00',
    settings_json TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Тренировочные карточки (одна запись = одно значение слова)
CREATE TABLE IF NOT EXISTS training_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    word_card_id INTEGER NOT NULL,
    word_en TEXT NOT NULL,
    transcription TEXT,
    sense_index INTEGER NOT NULL DEFAULT 0,
    word_ru TEXT NOT NULL,
    meaning_en TEXT NOT NULL,
    example_en TEXT,
    example_ru TEXT,
    distractors_ru TEXT,  -- JSON array
    distractors_en TEXT,  -- JSON array
    hint TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE,
    UNIQUE(word_card_id, sense_index)
);

-- Карточки пользователей (SRS)
CREATE TABLE IF NOT EXISTS user_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    training_card_id INTEGER NOT NULL,
    direction TEXT NOT NULL CHECK(direction IN ('ru_en', 'en_ru')),
    state TEXT DEFAULT 'new' CHECK(state IN ('new', 'learning', 'review')),
    ef REAL DEFAULT 2.5,
    reps INTEGER DEFAULT 0,
    interval_days INTEGER DEFAULT 0,
    learning_step INTEGER DEFAULT 0,
    lapse_count INTEGER DEFAULT 0,
    next_due_at DATETIME,
    last_review_at DATETIME,
    last_quality INTEGER,
    last_options_json TEXT,
    wrong_answers_json TEXT,
    stats_json TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (training_card_id) REFERENCES training_cards(id) ON DELETE CASCADE,
    UNIQUE(user_id, training_card_id, direction)
);

-- Сессии тренировок
CREATE TABLE IF NOT EXISTS training_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    ended_at DATETIME,
    source TEXT CHECK(source IN ('nudge', 'manual')),
    planned_count INTEGER,
    done_count INTEGER DEFAULT 0,
    session_json TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Лог попыток ответов
CREATE TABLE IF NOT EXISTS review_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER,
    user_id INTEGER NOT NULL,
    user_card_id INTEGER NOT NULL,
    direction TEXT NOT NULL,
    shown_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    options_shown_at DATETIME,
    answered_at DATETIME,
    t_delay_ms INTEGER,
    early_reveal INTEGER DEFAULT 0,
    option_count INTEGER,
    options_json TEXT,
    chosen_option TEXT,
    is_correct INTEGER,
    quality INTEGER,
    metrics_json TEXT,
    srs_before_json TEXT,
    srs_after_json TEXT,
    FOREIGN KEY (session_id) REFERENCES training_sessions(id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user_card_id) REFERENCES user_cards(id) ON DELETE CASCADE
);

-- Уведомления о тренировках
CREATE TABLE IF NOT EXISTS training_nudges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    local_date TEXT NOT NULL,
    sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    consumed_at DATETIME,
    due_count_at_send INTEGER,
    message_id INTEGER,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, local_date)
);

-- Состояние Circuit Breaker (синглтон)
CREATE TABLE IF NOT EXISTS circuit_breaker_state (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    is_open INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    last_failure_at DATETIME,
    last_reset_at DATETIME,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Инициализация circuit breaker
INSERT OR IGNORE INTO circuit_breaker_state (id) VALUES (1);
```

### 17.2. Индексы

```sql
CREATE INDEX IF NOT EXISTS idx_training_cards_word_card_id ON training_cards(word_card_id);
CREATE INDEX IF NOT EXISTS idx_user_cards_user_id ON user_cards(user_id);
CREATE INDEX IF NOT EXISTS idx_user_cards_training_card_id ON user_cards(training_card_id);
CREATE INDEX IF NOT EXISTS idx_user_cards_next_due_at ON user_cards(user_id, next_due_at);
CREATE INDEX IF NOT EXISTS idx_user_cards_state ON user_cards(user_id, state);
CREATE INDEX IF NOT EXISTS idx_training_sessions_user_id ON training_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_review_events_user_id ON review_events(user_id);
CREATE INDEX IF NOT EXISTS idx_training_nudges_user_date ON training_nudges(user_id, local_date);
```