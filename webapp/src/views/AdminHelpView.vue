<template>
  <div class="admin-help">
    <h1>Справка по системе</h1>
    <p class="intro">Документация по внутренним инструментам Linglow. Обновляется по мере добавления функций.</p>

    <nav class="toc">
      <strong>Разделы</strong>
      <ul>
        <li><a href="#lumi-facts">Lumi Facts — факты от Lumi</a></li>
      </ul>
    </nav>

    <!-- ===================== LUMI FACTS ===================== -->
    <section id="lumi-facts" class="help-section">
      <h2>Lumi Facts — факты от Lumi</h2>
      <p class="section-desc">
        Lumi Facts — небольшие познавательные карточки, которые Lumi показывает пользователю
        на разных экранах приложения. Каждый день пользователь видит новый факт.
      </p>

      <h3>Где хранятся</h3>
      <p>Таблица <code>lumi_facts</code> в PostgreSQL. Основные поля:</p>
      <table class="info-table">
        <thead>
          <tr><th>Поле</th><th>Тип</th><th>Смысл</th></tr>
        </thead>
        <tbody>
          <tr><td><code>course_code</code></td><td>text</td><td>Курс: <code>en_ru</code> или <code>es_ru</code>. Пустая строка — факт для всех курсов.</td></tr>
          <tr><td><code>context</code></td><td>text</td><td>Экран/тема, см. ниже.</td></tr>
          <tr><td><code>locale</code></td><td>text</td><td>Язык UI пользователя: <code>ru</code>, <code>en</code>, <code>es</code>.</td></tr>
          <tr><td><code>body</code></td><td>text</td><td>Текст факта.</td></tr>
          <tr><td><code>status</code></td><td>text</td><td><code>active</code> — показывается; <code>archived</code> — скрыт.</td></tr>
          <tr><td><code>last_shown_on</code></td><td>date</td><td>Дата последнего показа. Используется для ротации.</td></tr>
          <tr><td><code>shown_count</code></td><td>int</td><td>Сколько раз факт был показан суммарно.</td></tr>
        </tbody>
      </table>

      <h3>Контексты</h3>
      <p>Контекст определяет, на каком экране показывается факт:</p>
      <table class="info-table">
        <thead>
          <tr><th>Контекст</th><th>Где показывается</th><th>Тема фактов</th></tr>
        </thead>
        <tbody>
          <tr><td><code>general</code></td><td>Общие экраны, дашборд</td><td>Любопытные факты о языке: история, рекорды, этимология</td></tr>
          <tr><td><code>grammar</code></td><td>Экраны грамматики</td><td>Грамматические правила, сравнения с русским, исключения</td></tr>
          <tr><td><code>reading</code></td><td>Экраны чтения</td><td>Техники чтения, жанры, уровни текстов, авторы</td></tr>
          <tr><td><code>practice</code></td><td>Практические упражнения</td><td>Методы обучения, нейронаука, техники запоминания</td></tr>
          <tr><td><code>progress</code></td><td>Экран прогресса</td><td>Уровни CEFR, мотивация подкреплённая исследованиями, плато</td></tr>
          <tr><td><code>city</code></td><td>Карта города Linglow</td><td>География, история, архитектура городов носителей языка</td></tr>
        </tbody>
      </table>

      <h3>Как выбирается факт</h3>
      <p>
        Компонент <code>LgLumiFact.vue</code> вызывает <code>GET /api/lumi-facts/daily</code>
        с параметрами <code>context</code>, <code>course_code</code> и <code>locale</code>.
      </p>
      <p>Логика выбора на бэкенде:</p>
      <ol>
        <li>Берётся активный (<code>status = 'active'</code>) факт с совпадающим курсом и контекстом.</li>
        <li>Приоритет — те, что дольше всего не показывались (сортировка по <code>last_shown_on ASC NULLS FIRST</code>).</li>
        <li>После показа обновляются <code>last_shown_on = today</code> и <code>shown_count++</code>.</li>
        <li>Результат кэшируется на клиенте в <code>localStorage</code> до конца текущего дня (ключ включает дату).</li>
      </ol>
      <p class="hint">
        Пользователь видит один и тот же факт весь день. На следующий день — следующий по очереди.
        Это обеспечивает ротацию без случайных повторов.
      </p>

      <h3>Как управлять фактами</h3>
      <p>
        Страница управления:
        <router-link to="/lumi-facts"><strong>Lumi Facts</strong></router-link>
        в боковом меню.
      </p>
      <ul>
        <li><strong>Добавить факты</strong> — вставьте текст в поле (один абзац = один факт), выберите курс, контекст и локаль, нажмите «Добавить».</li>
        <li><strong>Редактировать</strong> — кнопка ✏️ в строке факта.</li>
        <li><strong>Архивировать</strong> — кнопка 🗑 скрывает факт, не удаляя из базы. Архивированные факты больше не показываются.</li>
        <li><strong>Восстановить</strong> — кнопка ↩️ возвращает архивированный факт в ротацию.</li>
        <li><strong>Фильтры</strong> — по курсу, контексту и статусу помогают найти нужные факты.</li>
      </ul>

      <h3>Как добавить факты через миграцию (при деплое)</h3>
      <p>
        Первоначальный набор фактов загружается SQL-миграциями. Каждая миграция защищена
        условием <code>WHERE NOT EXISTS</code>, чтобы не создавать дубли при повторном запуске.
      </p>
      <p>Текущие миграции с фактами:</p>
      <table class="info-table">
        <thead>
          <tr><th>Файл</th><th>Содержимое</th></tr>
        </thead>
        <tbody>
          <tr>
            <td><code>000026_lumi_facts.sql</code></td>
            <td>Создание таблицы + 10 general-фактов для <code>en_ru</code> и <code>es_ru</code></td>
          </tr>
          <tr>
            <td><code>000028_lumi_facts_grammar_seed.sql</code></td>
            <td>50 grammar-фактов для <code>en_ru</code> и 50 для <code>es_ru</code></td>
          </tr>
          <tr>
            <td><code>000029_lumi_facts_all_contexts_seed.sql</code></td>
            <td>15 фактов для каждого из контекстов reading, practice, progress, city + доп. general для обоих курсов</td>
          </tr>
        </tbody>
      </table>

      <h3>Как генерировать новые факты через Claude Code</h3>
      <p>
        В проекте есть slash-команда <code>/lumi-facts</code>, которую можно запустить
        прямо в сессии Claude Code (CLI):
      </p>
      <pre class="code-block">/lumi-facts</pre>
      <p>
        Команда создаст новую миграцию с 10 фактами для каждой комбинации курс × контекст
        (120 фактов суммарно). Файл появится в <code>internal/database/migrations/</code>
        и применится при следующем деплое вместе с остальными миграциями.
      </p>
      <p>
        Файл команды: <code>.claude/commands/lumi-facts.md</code>.
        Можно открыть и отредактировать инструкцию для Клода, если нужно изменить логику генерации.
      </p>

      <h3>Частые вопросы</h3>
      <div class="faq">
        <div class="faq-item">
          <div class="faq-q">Почему пользователь видит один и тот же факт несколько дней подряд?</div>
          <div class="faq-a">
            Скорее всего, в этом контексте и курсе мало активных фактов — они заканчиваются
            быстрее, чем успевает пройти цикл ротации. Добавьте больше фактов через админку
            или запустите <code>/lumi-facts</code>.
          </div>
        </div>
        <div class="faq-item">
          <div class="faq-q">Можно ли показывать один факт сразу во всех курсах?</div>
          <div class="faq-a">
            Да — оставьте <code>course_code</code> пустым при добавлении факта.
            Такой факт будет показываться независимо от активного курса пользователя.
          </div>
        </div>
        <div class="faq-item">
          <div class="faq-q">Как сбросить кэш факта у пользователя?</div>
          <div class="faq-a">
            Кэш хранится в <code>localStorage</code> браузера под ключом
            <code>lumi-fact:&lt;course&gt;:&lt;context&gt;:&lt;locale&gt;:&lt;YYYY-MM-DD&gt;</code>.
            Автоматически сбрасывается в полночь по местному времени пользователя.
            Вручную — через DevTools → Application → Local Storage.
          </div>
        </div>
        <div class="faq-item">
          <div class="faq-q">Где в коде находится компонент с фактами?</div>
          <div class="faq-a">
            <code>webapp/src/components/linglow/LgLumiFact.vue</code> — компонент отображения.<br>
            <code>webapp/src/api/factClient.ts</code> — клиент API.<br>
            Бэкенд: поиск по <code>/api/lumi-facts</code> в <code>internal/web/</code>.
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
// no logic needed — static documentation page
</script>

<style scoped>
.admin-help {
  max-width: 900px;
  margin: 0 auto;
  padding: 16px;
}
.intro {
  color: var(--text-secondary);
  margin-bottom: 24px;
}
.toc {
  background: var(--card-bg);
  border: 1px solid var(--border-primary, #ddd);
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 32px;
  display: inline-block;
  min-width: 220px;
}
.toc ul {
  margin: 6px 0 0;
  padding-left: 18px;
}
.toc a { color: var(--link, #2563eb); text-decoration: none; }
.toc a:hover { text-decoration: underline; }

.help-section {
  margin-bottom: 48px;
}
.help-section h2 {
  font-size: 1.4rem;
  border-bottom: 2px solid var(--border-primary, #e5e7eb);
  padding-bottom: 8px;
  margin-bottom: 16px;
}
.help-section h3 {
  font-size: 1.05rem;
  margin: 24px 0 8px;
  color: var(--text);
}
.section-desc {
  color: var(--text-secondary);
  margin-bottom: 16px;
}

.info-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
  margin-bottom: 8px;
}
.info-table th {
  text-align: left;
  padding: 6px 10px;
  background: var(--card-bg-secondary, var(--card-bg));
  border-bottom: 2px solid var(--border-primary, #ddd);
  font-weight: 600;
}
.info-table td {
  padding: 6px 10px;
  border-bottom: 1px solid var(--border-primary, #eee);
  vertical-align: top;
}
.info-table code {
  background: var(--code-bg, #f3f4f6);
  padding: 1px 5px;
  border-radius: 4px;
  font-size: 0.85em;
}

.hint {
  background: var(--card-bg);
  border-left: 3px solid var(--dorado, #f59e0b);
  padding: 8px 12px;
  border-radius: 0 6px 6px 0;
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 8px 0;
}

.code-block {
  background: var(--code-bg, #1e1e2e);
  color: var(--code-text, #cdd6f4);
  padding: 10px 14px;
  border-radius: 6px;
  font-family: monospace;
  font-size: 0.95rem;
  margin: 8px 0;
}

ol, ul { padding-left: 20px; line-height: 1.8; }

.faq { display: flex; flex-direction: column; gap: 12px; margin-top: 8px; }
.faq-item {
  background: var(--card-bg);
  border: 1px solid var(--border-primary, #ddd);
  border-radius: 8px;
  overflow: hidden;
}
.faq-q {
  padding: 10px 14px;
  font-weight: 600;
  font-size: 0.9rem;
  background: var(--card-bg-secondary, var(--card-bg));
  border-bottom: 1px solid var(--border-primary, #eee);
}
.faq-a {
  padding: 10px 14px;
  font-size: 0.9rem;
  line-height: 1.6;
  color: var(--text-secondary);
}
.faq-a code {
  background: var(--code-bg, #f3f4f6);
  padding: 1px 5px;
  border-radius: 4px;
  font-size: 0.85em;
  color: var(--text);
}
</style>
