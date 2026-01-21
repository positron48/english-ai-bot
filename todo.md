Бери и делай задачи по порядку. Каждое сообщение - отдельная задача. 
После каждой задачи проверяй make check.
для нового функционала или вносимых изменений пиши тесты / корректируй существующие
правки по грамматике должны вноситься в рамках сервиса (vue-приложения и go). courses/ - сабмодуль, который ридонли и его править запрещено.
отмечай статус работы над задачей в этом файле ([done] ..., [in progress] ..., etc)


1. [done] В курс грамматики нужно добавить общий тест на определение текущего уровня пользователя, скажем, на 20-30 вопросов. В него должны включаться разные вопросы из всех опубликованных категорий курса, по результату должен присваиваться некий уровень и открываться все категории до этого уровня. Оперировать лучше не глобально уровнем B1 / A1, а конкретными категориями.
Как определять уровень клиента по результатам теста - предложи. Тест можно проходить сколько угодно раз, вопросы каждый раз рандомные. Но результат прохождения может перезаписываться только в большую сторону - если мы сначала определили что клиент знает 5 категорий, а в следующий раз - что знает 3 - считаем что он знает 5 все еще. В блоке статистики нужно считать все открытые по результатам теста категории пройденными несмотря на то, что пользователь не проходил в них тесты.

2. [done] В дашборд нужно добавить блок статистики из грамматики, над блоком Your Progress

3. [done] К вопросам внутри теста на определение текущего уровня нужно выводить название главы курса, из которого взят этот вопрос (именно главы, не теоретического блока внутри).

4. [done] Я прошел тест, мне показало такое:
Placement Test Results
14%
Your Level
You answered 3 out of 21 questions correctly.
Keep practicing to unlock more sections!

А как понять какой мой уровень по результатам теста? Какие категории/главы мне будут открыты? Если ты ориентируешься чисто на процент правильных ответов - это неверно. Нужно как то оценивать правильность ответов от первой катеории до последней и давать тот уровень, до которого пользователь уверенно отвечал на вопросы.

5. [done] Также в конце теста на определение уровня нужно выдавать все вопросы и все ответы - правильные и неправильные, если их дали мы - также как после теста по главе.

6. [done] [BACKEND] {"level":"ERROR","timestamp":"2026-01-21T11:46:33.395+0300","caller":"service/grammar_service.go:1158","msg":"failed to save placement test attempt","error":"failed to create attempt: constraint failed: CHECK constraint failed: scope_type IN ('chapter', 'category') (275)","stacktrace":"tgbot-skeleton/internal/service.(*GrammarService).SubmitPlacementTest\n\t/var/www/my/english-bot/internal/service/grammar_service.go:1158\ntgbot-skeleton/internal/web.(*Router).handleLearningGrammarSubmitPlacementTest\n\t/var/www/my/english-bot/internal/web/grammar.go:567\ntgbot-skeleton/internal/web.(*Router).setupProtectedRoutes.(*AuthMiddleware).RequireAuth.func25\n\t/var/www/my/english-bot/internal/web/auth.go:101\ntgbot-skeleton/internal/web.(*Router).setupProtectedRoutes.(*RateLimitMiddleware).Wrap.func26\n\t/var/www/my/english-bot/internal/web/rate_limit_middleware.go:82\nnet/http.HandlerFunc.ServeHTTP\n\t/home/linuxbrew/.linuxbrew/Cellar/go/1.24.6/libexec/src/net/http/server.go:2294\nnet/http.(*ServeMux).ServeHTTP\n\t/home/linuxbrew/.linuxbrew/Cellar/go/1.24.6/libexec/src/net/http/server.go:2822\ntgbot-skeleton/internal/web.(*Router).ServeHTTP.func1\n\t/var/www/my/english-bot/internal/web/router.go:882\ntgbot-skeleton/internal/web.(*Router).ServeHTTP.(*Router).corsMiddleware.func2\n\t/var/www/my/english-bot/internal/web/router.go:396\ntgbot-skeleton/internal/web.(*Router).ServeHTTP\n\t/var/www/my/english-bot/internal/web/router.go:883\nnet/http.(*ServeMux).ServeHTTP\n\t/home/linuxbrew/.linuxbrew/Cellar/go/1.24.6/libexec/src/net/http/server.go:2822\nnet/http.serverHandler.ServeHTTP\n\t/home/linuxbrew/.linuxbrew/Cellar/go/1.24.6/libexec/src/net/http/server.go:3301\nnet/http.(*conn).serve\n\t/home/linuxbrew/.linuxbrew/Cellar/go/1.24.6/libexec/src/net/http/server.go:2102"}
Исправлено: миграция пересоздаёт grammar_test_attempts с CHECK(scope_type IN ('chapter','category','placement')).