//nolint:unused // These schemas and annotation functions are consumed by the Swagger generator.
package web

import "tgbot-skeleton/internal/service"

// Documentation-only request/response types and functions keep Swagger attached
// to functions while the runtime routes share handlePlacementSession.
type placementStartDocumentation struct {
	// Defaults to the requested or currently selected course when omitted.
	CourseCode string `json:"course_code" enums:"en_ru,es_ru" example:"es_ru"`
	// Reuse the same key when retrying the same start/resume request.
	IdempotencyKey string `json:"idempotency_key" validate:"required" minLength:"8" maxLength:"100" example:"d83b3477-1a78-414e-92c4-f1287a5c58d5"`
	// True abandons an active attempt and creates another selection.
	NewAttempt bool `json:"new_attempt" default:"false"`
}

type placementAnswerDocumentation struct {
	QuestionID string `json:"question_id" validate:"required" example:"es.placement.a1.location.01"`
	// A choice ID, or the empty string for an explicit "I don't know". Missing/null is invalid.
	Answer *string `json:"answer" validate:"required" example:"a"`
}

type placementErrorDocumentation struct {
	Code    string `json:"code" example:"placement_expired"`
	Message string `json:"message,omitempty" example:"Срок этой попытки истёк. Начните новый тест."`
}

type placementSessionDocumentation struct {
	service.PlacementSessionView
	// Present only for a completed attempt with usable result links. This is a
	// selected subset of currently accessible chapter IDs from result recommendations
	// and review, not the complete course catalog. Link only to chapter IDs present
	// here; a recommended topic may remain visible even when all its lessons are
	// locked by the course sequence. Omitted when no accessible result links exist.
	AvailableChapterIDs []string `json:"available_chapter_ids,omitempty"`
}

type placementHistoryDocumentation struct {
	// The latest ten completed attempts for this course. Questions and answers are null;
	// the diagnostic profile, answer review and published recommendations are in result.
	Sessions []*placementSessionDocumentation `json:"sessions"`
}

type placementRetiredDocumentation struct {
	Code    string `json:"code" example:"placement_replaced"`
	Message string `json:"message" example:"Откройте новый тест определения уровня."`
	Path    string `json:"path" example:"/learning/placement-test"`
}

// documentPlacementStart documents the explicit creation/resume action.
// @Summary      Начать или восстановить диагностику языка
// @Description  Создаёт самостоятельный тест из 30 вопросов A1–C1 либо восстанавливает активную попытку пользователя в указанном курсе. Повтор одного idempotency_key возвращает ту же попытку; для пересдачи отправьте новый ключ и new_attempt=true. Вопросы и резерв до шести уточнений закреплены в сессии. Правильные ответы и объяснения отсутствуют до завершения. Все ответы имеют Cache-Control: no-store.
// @Tags         Learning Placement
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        course_code query string false "Курс, если course_code отсутствует в теле; иначе используется текущий курс" Enums(en_ru,es_ru)
// @Param        request body placementStartDocumentation true "Курс и ключ идемпотентности"
// @Success      200 {object} placementSessionDocumentation "Новая или восстановленная попытка; available_chapter_ids только для завершённой попытки с доступными переходами из результата"
// @Failure      400 {object} placementErrorDocumentation "placement_invalid_request / placement_invalid_answer"
// @Failure      401 {object} placementErrorDocumentation "unauthorized"
// @Failure      404 {object} placementErrorDocumentation "placement_course_not_found / placement_not_found"
// @Failure      410 {object} placementErrorDocumentation "placement_expired: попытка завершена отказом, истекла или её политика больше не поддерживается"
// @Failure      503 {object} placementErrorDocumentation "placement_unavailable"
// @Router       /api/learning/placement/sessions [post]
func documentPlacementStart() {}

// documentPlacementGet documents loading an owned, course-bound snapshot.
// @Summary      Получить сохранённую попытку диагностики
// @Description  Возвращает закреплённые публичные вопросы и сохранённые ответы пользователя. Курс определяется сессией, а доступ — её владельцем. После завершения содержит result с профилем уровней, разбором и опубликованными рекомендациями. available_chapter_ids перечисляет выбранные доступные переходы из результата, а не весь каталог; ссылка на урок допустима только для ID из этого списка. Закрытая последовательностью тема остаётся видимой без ссылки. При отсутствии доступных переходов поле опущено. Не создаёт новую попытку. Cache-Control: no-store.
// @Tags         Learning Placement
// @Produce      json
// @Security     ApiKeyAuth
// @Param        session_id path string true "Идентификатор сессии: 32 шестнадцатеричных символа" minlength(32) maxlength(32)
// @Success      200 {object} placementSessionDocumentation
// @Failure      401 {object} placementErrorDocumentation "unauthorized"
// @Failure      404 {object} placementErrorDocumentation "placement_not_found: не найдена либо принадлежит другому пользователю"
// @Failure      410 {object} placementErrorDocumentation "placement_expired"
// @Failure      503 {object} placementErrorDocumentation "placement_unavailable"
// @Router       /api/learning/placement/sessions/{session_id} [get]
func documentPlacementGet() {}

// documentPlacementAnswer documents the shared handler's answers action.
// @Summary      Сохранить ответ диагностики
// @Description  Принимает один идентификатор варианта или пустую строку для «Не знаю»; отсутствующий/null answer недопустим. После 30 ответов base_closed=true и основной блок больше не изменяется. При необходимости сервер добавляет ровно шесть уточнений: clarifying=true, questions содержит 36 вопросов. Повтор того же сохранённого ответа безопасен; ключи ответов и объяснения до finish не раскрываются. Cache-Control: no-store.
// @Tags         Learning Placement
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        session_id path string true "Идентификатор сессии" minlength(32) maxlength(32)
// @Param        request body placementAnswerDocumentation true "Один ответ на выданный вопрос"
// @Success      200 {object} placementSessionDocumentation "Состояние после сохранения, включая уточнения при закрытии основного блока"
// @Failure      400 {object} placementErrorDocumentation "placement_invalid_request / placement_invalid_answer"
// @Failure      401 {object} placementErrorDocumentation "unauthorized"
// @Failure      404 {object} placementErrorDocumentation "placement_not_found"
// @Failure      409 {object} placementErrorDocumentation "placement_conflict: попытка или основной блок закрыты для изменений"
// @Failure      410 {object} placementErrorDocumentation "placement_expired"
// @Failure      503 {object} placementErrorDocumentation "placement_unavailable"
// @Router       /api/learning/placement/sessions/{session_id}/answers [post]
func documentPlacementAnswer() {}

// documentPlacementFinish documents the shared handler's finish action.
// @Summary      Завершить диагностику и получить рекомендации
// @Description  Требует ответы на все выданные вопросы, включая уточнения, если они появились. Атомарно сохраняет результат и расширяет доступ только в курсе этой сессии. Слабая пересдача не уменьшает ранее открытый доступ; уроки не помечаются пройденными. Повтор finish возвращает сохранённый результат с актуальными доступными переходами. available_chapter_ids содержит только выбранные доступные уроки из рекомендаций/разбора и не является каталогом курса; при отсутствии доступных переходов поле опущено. Закрытые последовательностью темы остаются видимыми без ссылки. Оценка относится к грамматике в письменном контексте и остаётся ориентировочной. Тело запроса не требуется. Cache-Control: no-store.
// @Tags         Learning Placement
// @Produce      json
// @Security     ApiKeyAuth
// @Param        session_id path string true "Идентификатор сессии" minlength(32) maxlength(32)
// @Success      200 {object} placementSessionDocumentation "status=completed, result с profile, review и recommended_skills; необязательный available_chapter_ids для доступных переходов"
// @Failure      401 {object} placementErrorDocumentation "unauthorized"
// @Failure      404 {object} placementErrorDocumentation "placement_not_found"
// @Failure      409 {object} placementErrorDocumentation "placement_conflict: остались вопросы без ответа или попытка недоступна для завершения"
// @Failure      410 {object} placementErrorDocumentation "placement_expired"
// @Failure      503 {object} placementErrorDocumentation "placement_unavailable"
// @Router       /api/learning/placement/sessions/{session_id}/finish [post]
func documentPlacementFinish() {}

// documentPlacementHistory documents the course-specific completed history.
// @Summary      Получить историю диагностики языка
// @Description  Возвращает до десяти последних завершённых попыток текущего пользователя только для выбранного курса, от новых к старым. В элементах sessions поля questions и answers равны null; результат и разбор находятся в result. Снятые с публикации главы/разделы исключены из chapter_ids. Необязательный available_chapter_ids содержит выбранные доступные переходы из результата, а не все доступные уроки курса; темы, закрытые последовательностью, остаются видимыми без ссылки. Cache-Control: no-store.
// @Tags         Learning Placement
// @Produce      json
// @Security     ApiKeyAuth
// @Param        course_code query string false "Курс; при отсутствии используется текущий выбранный курс" Enums(en_ru,es_ru)
// @Success      200 {object} placementHistoryDocumentation
// @Failure      401 {object} placementErrorDocumentation "unauthorized"
// @Failure      404 {object} placementErrorDocumentation "placement_course_not_found"
// @Failure      503 {object} placementErrorDocumentation "placement_unavailable"
// @Router       /api/learning/placement/results [get]
func documentPlacementHistory() {}

// documentRetiredPlacementGet documents the registered retirement handler.
// @Summary      Устаревший тест из вопросов курса — заменён
// @Description  Всегда возвращает 410 placement_replaced. Для диагностики используйте POST /api/learning/placement/sessions, для интерфейса — /learning/placement-test.
// @Tags         Learning Placement
// @Produce      json
// @Security     ApiKeyAuth
// @Deprecated
// @Failure      401 {object} placementErrorDocumentation "unauthorized"
// @Failure      410 {object} placementRetiredDocumentation "placement_replaced"
// @Router       /api/learning/grammar/placement-test [get]
func documentRetiredPlacementGet() {}

// documentRetiredPlacementSubmit documents the registered retirement handler.
// @Summary      Устаревшая отправка ответов теста курса — заменена
// @Description  Всегда возвращает 410 placement_replaced; переданные ответы не оцениваются. Используйте answers и finish новой серверной сессии диагностики.
// @Tags         Learning Placement
// @Produce      json
// @Security     ApiKeyAuth
// @Deprecated
// @Failure      401 {object} placementErrorDocumentation "unauthorized"
// @Failure      410 {object} placementRetiredDocumentation "placement_replaced"
// @Router       /api/learning/grammar/placement-test/submit [post]
func documentRetiredPlacementSubmit() {}
