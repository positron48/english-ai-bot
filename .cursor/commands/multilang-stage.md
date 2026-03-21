# multilang-stage

Оркестрация этапа **A–F** по `docs/MULTILANG_SPANISH_LAUNCH_PLAN.md`: план → код → тесты → **`make check`** → покрытие → ревью. **Приоритет №1:** не сломать текущий работающий сервис для английского (RU→EN).

## Использование

`/multilang-stage A` … `/multilang-stage F`

Опционально в том же сообщении: уточнения (ветка, ограничение файлов, что уже сделано).

Оркестратор (чат-агент пользователя) выполняет шаги ниже и **не подменяет** роли субагентов — только вызывает **mcp_task**.

## Шаги

### 1. Планирование

Вызвать **mcp_task** с `subagent_type="multilang-planner"`, `prompt`: «Этап: [A|B|C|D|E|F]. Прочитай `docs/MULTILANG_SPANISH_LAUNCH_PLAN.md` (раздел про этот этап). Дополнительный контекст от пользователя: [вставить]. Верни Tasks, Acceptance, EnglishRegressionChecks, OutOfScope, KeyFiles.», `description`: "Plan multilang stage".

### 2. Реализация

Вызвать **mcp_task** с `subagent_type="multilang-implementer"`, `prompt`: «План этапа (вставить вывод planner целиком). Реализуй только задачи этапа; English-first.», `description`: "Implement multilang stage".

### 3. Тесты

Вызвать **mcp_task** с `subagent_type="multilang-tester"`, `prompt`: «Этап [X]. Реализовано: [кратко / diff]. Критерии и чеклист English из плана: [вставить из planner]. Добавь/обнови тесты.», `description`: "Tests for multilang stage".

### 4. Проверка (цикл)

**Максимум 5 итераций** исправлений по ошибкам `make check`:

- Вызвать **mcp_task** с `subagent_type="test-runner"`, `prompt`: «Запусти `make check` в корне проекта english-ai-bot. Верни: passed или failed, точное coverage_percent из строки Total test coverage, при failed — excerpt ошибок.», `description`: "Run make check". **Таймаут вызова — не менее 15 минут.**

- Если **passed** — перейти к шагу 5.

- Если **failed** — вызвать **mcp_task** с `subagent_type="multilang-implementer"`, `prompt`: «make check failed. Лог: [excerpt]. Исправь код в рамках этапа [X].», затем при необходимости **multilang-tester** (если правки требуют тестов), затем снова **test-runner**. Если после 5 итераций всё ещё failed — остановиться и отчитаться пользователю с логом.

### 5. Покрытие

После **passed** вызвать **mcp_task** с `subagent_type="multilang-coverage"`, `prompt`: «Этап [X]. coverage_percent: [из test-runner]. Кратко изменённые зоны: [из контекста]. Сверь с .cursor/config.json testing.coverageMinPercent.», `description`: "Multilang coverage report".

### 6. Ревью

Вызвать **mcp_task** с `subagent_type="multilang-reviewer"`, `prompt`: «Этап [X]. Список затронутых файлов/изменений: [кратко]. Результат make check: passed. Покрытие: [coverage_percent]. Оцени соответствие плану и риски English.», `description`: "Review multilang stage".

### 7. Коммит (опционально)

Только если пользователь явно просит закоммитить итог этапа: **mcp_task** с `subagent_type="committer"`, `prompt`: «Закоммить изменения с сообщением по смыслу этапа multilang [X], без шаблона improve coverage.», `description`: "Commit multilang stage".

## Результат

- Краткий отчёт: этап, passed/failed `make check`, процент покрытия, вердикт reviewer (`ready` / `needs_followup`), список follow-up.

## Заметки

- Роли **test-runner** и при необходимости **committer** уже существуют в `.cursor/agents/`.
- Для доведения покрытия до 100% по репозиторию используй `/improve-coverage` — это отдельный процесс.
- Правило контекста: при работе по multilang подключай **@multilang-english-first** или правило `multilang-english-first.mdc` в Cursor.
