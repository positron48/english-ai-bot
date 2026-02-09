Бери и делай задачи по порядку. Каждое сообщение - отдельная задача. 
После каждой задачи проверяй make check.
для нового функционала или вносимых изменений пиши тесты / корректируй существующие
правки по грамматике должны вноситься в рамках сервиса (vue-приложения и go). courses/ - сабмодуль, который ридонли и его править запрещено.
отмечай статус работы над задачей в этом файле ([done] ..., [in progress] ..., etc)

[done] Обрабатывать ошибки `bot was blocked`/`chat not found`: ставить периодичность уведомлений в `never`; не отправлять уведомления при <10 карточках на повторение.
[done] Грамматика (адаптив): переносы строк заголовков глав (в т.ч. со слешами/спецсимволами).
[done] Грамматика: скрыть заголовок «Проверка знаний», если квиза нет.
[done] Для ошибок `bot was blocked`/`chat not found` логировать warning (не error) и писать info о результате отписки пользователя; прочие ошибки оставить как есть.

2026-02-09 19:20:15.109	
{"level":"ERROR","timestamp":"2026-02-09T16:20:15.108Z","caller":"service/notification_service.go:131","msg":"failed to send notification","user_id":14,"error":"failed to send message: Forbidden: bot was blocked by the user"}
2026-02-09 19:20:15.045	
{"level":"ERROR","timestamp":"2026-02-09T16:20:15.045Z","caller":"service/notification_service.go:131","msg":"failed to send notification","user_id":13,"error":"failed to send message: Bad Request: chat not found"}
2026-02-09 19:20:14.724	
{"level":"ERROR","timestamp":"2026-02-09T16:20:14.724Z","caller":"service/notification_service.go:131","msg":"failed to send notification","user_id":10,"error":"failed to send message: Forbidden: bot was blocked by the user"}
2026-02-08 23:20:13.688	
{"level":"ERROR","timestamp":"2026-02-08T20:20:13.688Z","caller":"service/notification_service.go:131","msg":"failed to send notification","user_id":14,"error":"failed to send message: Forbidden: bot was blocked by the user"}
2026-02-08 23:20:13.624	
{"level":"ERROR","timestamp":"2026-02-08T20:20:13.624Z","caller":"service/notification_service.go:131","msg":"failed to send notification","user_id":13,"error":"failed to send message: Bad Request: chat not found"}
Если видишь такую ошибку - переводи настройку периодичности подписки пользователя на уведомления как "никогда", чтобы не отправлять уведомления повторно. Также, если у пользователя нет хотя бы 10 карточек для повторения - не шли уведомление.

В грамматике нужно сделать переносы строк заголовков глав, если не влезают (слеши, спец символы) в адаптиве.

Убрать заголовок "Проверка знаний" в главах грамматики, если нет квиза
